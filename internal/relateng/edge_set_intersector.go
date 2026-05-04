package relateng

import (
	"github.com/exergy-dev/go-topology-suite/geom"
	"github.com/exergy-dev/go-topology-suite/index"
	"github.com/exergy-dev/go-topology-suite/internal/noding"
)

// EdgeSetIntersector drives the segment-pair intersection pass for
// RelateNG. For each input edge it builds monotone chains, indexes
// them, and asks the chain-pair binary subdivision (ComputeOverlaps)
// to surface candidate segment pairs. Each candidate is dispatched to
// the EdgeSegmentIntersector, which records intersections on the
// TopologyComputer.
//
// Port of org.locationtech.jts.operation.relateng.EdgeSetIntersector.
//
// The Go port reuses the existing internal/noding monotone-chain types
// (MonotoneChain / BuildMonotoneChains) which match JTS's index/chain
// package one-for-one.
//
// The chain index is partitioned by side: chainsA / idxA hold A's
// monotone chains, chainsB / idxB hold B's. When the active predicate
// does not require self-noding (the common Intersects/Disjoint/
// Contains/Covers/Touches case), Process queries A-chains against the
// B-only index, which avoids descending the R-tree into same-side
// candidate pairs only to reject them in the visitor callback. When
// self-noding is required the scan also runs A-vs-A and B-vs-B over
// each side's own index.
type EdgeSetIntersector struct {
	envelope geom.Envelope
	chainsA  []*relateChain
	chainsB  []*relateChain
	idxA     *index.RTree[*relateChain]
	idxB     *index.RTree[*relateChain]
}

type relateChain struct {
	id int
	ss *RelateSegmentString
	mc *noding.MonotoneChain
}

// NewEdgeSetIntersector indexes edgesA and edgesB. envelope is an
// optional clip envelope; chains whose envelopes don't intersect it
// are dropped (matches JTS).
func NewEdgeSetIntersector(edgesA, edgesB []*RelateSegmentString, env geom.Envelope) *EdgeSetIntersector {
	es := &EdgeSetIntersector{
		envelope: env,
		idxA:     index.New[*relateChain](),
		idxB:     index.New[*relateChain](),
	}
	// Chain ids are unique across A and B so the existing
	// `tc.id <= qc.id` ordering guard still works for the
	// self-noding=true intra-side queries.
	nextID := 0
	for _, ss := range edgesA {
		es.chainsA = append(es.chainsA, es.buildChains(ss, &nextID)...)
	}
	for _, ss := range edgesB {
		es.chainsB = append(es.chainsB, es.buildChains(ss, &nextID)...)
	}
	es.idxA.Bulk(toRTreeItems(es.chainsA))
	es.idxB.Bulk(toRTreeItems(es.chainsB))
	return es
}

func toRTreeItems(chs []*relateChain) []index.Item[*relateChain] {
	items := make([]index.Item[*relateChain], 0, len(chs))
	for _, c := range chs {
		items = append(items, index.Item[*relateChain]{
			Env:   c.mc.Envelope(),
			Value: c,
		})
	}
	return items
}

func (es *EdgeSetIntersector) buildChains(ss *RelateSegmentString, nextID *int) []*relateChain {
	// BuildMonotoneChains expects a noding.SegmentString. We adapt by
	// constructing one with the underlying coords.
	tmp := &noding.SegmentString{Coords: ss.Coords}
	mcs := noding.BuildMonotoneChains(tmp)
	out := make([]*relateChain, 0, len(mcs))
	for _, mc := range mcs {
		if !es.envelope.IsEmpty() && !es.envelope.Intersects(mc.Envelope()) {
			continue
		}
		mc.ID = *nextID
		*nextID++
		out = append(out, &relateChain{
			id: mc.ID,
			ss: ss,
			mc: mc,
		})
	}
	return out
}

// SegmentPairProcessor is the abstract sink for chain-pair dispatch
// inside the EdgeSetIntersector. Production wiring uses
// *EdgeSegmentIntersector; tests can supply a counter / fake.
type SegmentPairProcessor interface {
	ProcessIntersections(ss0 *RelateSegmentString, seg0 int, ss1 *RelateSegmentString, seg1 int)
	IsDone() bool
}

// Process runs the segment-pair scan, calling intersector.ProcessIntersections
// for every candidate pair. The scan stops early when intersector.IsDone.
//
// requireSelfNoding=false skips intra-input (A-vs-A, B-vs-B) chain
// pairs, which the predicate would discard anyway.
func (es *EdgeSetIntersector) Process(intersector SegmentPairProcessor, requireSelfNoding bool) {
	// Cross-index scan: every (A-chain, B-chain) pair is unique by
	// construction, so no id-ordering guard is needed here. Iterate
	// the smaller side as the query set so the visitor descends the
	// larger tree fewer times.
	if len(es.chainsA) <= len(es.chainsB) {
		if es.processCross(es.chainsA, es.idxB, intersector, false) {
			return
		}
	} else {
		if es.processCross(es.chainsB, es.idxA, intersector, true) {
			return
		}
	}
	if !requireSelfNoding {
		return
	}
	// Self-noding: also test A-vs-A and B-vs-B pairs. The id-ordering
	// guard ensures each unordered same-side pair is visited once.
	if es.processSelf(es.chainsA, es.idxA, intersector) {
		return
	}
	es.processSelf(es.chainsB, es.idxB, intersector)
}

// processCross dispatches every candidate pair where one chain comes
// from queries and the other from idx. queriesAreB tells the
// intersector how to orient the (ss0, ss1) callback so it always sees
// the A chain first (matching the legacy callback orientation, where
// `tc` came from the index visitor and `qc` from the iteration loop).
func (es *EdgeSetIntersector) processCross(queries []*relateChain, idx *index.RTree[*relateChain], intersector SegmentPairProcessor, queriesAreB bool) bool {
	for _, qc := range queries {
		stop := false
		idx.Search(qc.mc.Envelope(), func(it index.Item[*relateChain]) bool {
			tc := it.Value
			if queriesAreB {
				// queries are B chains, idx holds A chains: tc is the
				// A chain. Match the legacy (tc, qc) → (ss0, ss1)
				// callback orientation.
				tc.mc.ComputeOverlaps(qc.mc, 0, func(_ *noding.MonotoneChain, s1 int, _ *noding.MonotoneChain, s2 int) {
					intersector.ProcessIntersections(tc.ss, s1, qc.ss, s2)
				})
			} else {
				// queries are A chains, idx holds B chains: qc is the
				// A chain. Swap so the A side is reported first.
				qc.mc.ComputeOverlaps(tc.mc, 0, func(_ *noding.MonotoneChain, s1 int, _ *noding.MonotoneChain, s2 int) {
					intersector.ProcessIntersections(qc.ss, s1, tc.ss, s2)
				})
			}
			if intersector.IsDone() {
				stop = true
				return false
			}
			return true
		})
		if stop {
			return true
		}
	}
	return false
}

// processSelf dispatches same-side candidate pairs (A-vs-A or
// B-vs-B). The id-ordering guard avoids visiting each unordered pair
// twice and avoids self-pairs.
func (es *EdgeSetIntersector) processSelf(chains []*relateChain, idx *index.RTree[*relateChain], intersector SegmentPairProcessor) bool {
	for _, qc := range chains {
		stop := false
		idx.Search(qc.mc.Envelope(), func(it index.Item[*relateChain]) bool {
			tc := it.Value
			if tc.id <= qc.id {
				return true
			}
			tc.mc.ComputeOverlaps(qc.mc, 0, func(_ *noding.MonotoneChain, s1 int, _ *noding.MonotoneChain, s2 int) {
				intersector.ProcessIntersections(tc.ss, s1, qc.ss, s2)
			})
			if intersector.IsDone() {
				stop = true
				return false
			}
			return true
		})
		if stop {
			return true
		}
	}
	return false
}
