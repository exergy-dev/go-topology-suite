package relateng

import (
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/index"
	"github.com/terra-geo/terra/internal/noding"
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
type EdgeSetIntersector struct {
	envelope geom.Envelope
	chains   []*relateChain
	idx      *index.RTree[*relateChain]
}

type relateChain struct {
	id  int
	ss  *RelateSegmentString
	mc  *noding.MonotoneChain
}

// NewEdgeSetIntersector indexes edgesA and edgesB. envelope is an
// optional clip envelope; chains whose envelopes don't intersect it
// are dropped (matches JTS).
func NewEdgeSetIntersector(edgesA, edgesB []*RelateSegmentString, env geom.Envelope) *EdgeSetIntersector {
	es := &EdgeSetIntersector{
		envelope: env,
		idx:      index.New[*relateChain](),
	}
	es.addEdges(edgesA)
	es.addEdges(edgesB)
	es.idx.Bulk(toRTreeItems(es.chains, env))
	return es
}

func toRTreeItems(chs []*relateChain, env geom.Envelope) []index.Item[*relateChain] {
	items := make([]index.Item[*relateChain], 0, len(chs))
	for _, c := range chs {
		items = append(items, index.Item[*relateChain]{
			Env:   c.mc.Envelope(),
			Value: c,
		})
	}
	_ = env
	return items
}

func (es *EdgeSetIntersector) addEdges(edges []*RelateSegmentString) {
	for _, ss := range edges {
		es.addToIndex(ss)
	}
}

func (es *EdgeSetIntersector) addToIndex(ss *RelateSegmentString) {
	// BuildMonotoneChains expects a noding.SegmentString. We adapt by
	// constructing one with the underlying coords.
	tmp := &noding.SegmentString{Coords: ss.Coords}
	mcs := noding.BuildMonotoneChains(tmp)
	for _, mc := range mcs {
		if !es.envelope.IsEmpty() && !es.envelope.Intersects(mc.Envelope()) {
			continue
		}
		mc.ID = len(es.chains)
		es.chains = append(es.chains, &relateChain{
			id: mc.ID,
			ss: ss,
			mc: mc,
		})
	}
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
// requireSelfNoding controls whether intra-input pairs (A-vs-A and
// B-vs-B) participate. When false (the common case for predicates like
// Intersects / Disjoint / Contains / Covers / Touches whose answer
// does not depend on self-intersections of an individual operand) the
// scan visits only A-vs-B candidate pairs, skipping work that the
// predicate would discard anyway. When true (Relate matrix predicates,
// or any predicate whose JTS counterpart returns
// requireSelfNoding=true) every chain pair is tested as before.
//
// Mirrors the corresponding short-circuit in JTS RelateNG, where the
// EdgeSetIntersector consults the active TopologyPredicate before
// dispatching each chain pair.
func (es *EdgeSetIntersector) Process(intersector SegmentPairProcessor, requireSelfNoding bool) {
	for _, qc := range es.chains {
		queryEnv := qc.mc.Envelope()
		stop := false
		es.idx.Search(queryEnv, func(it index.Item[*relateChain]) bool {
			tc := it.Value
			// Pair-ordering: only test chains where target.id > query.id
			// (so each unordered chain pair is visited once and a chain
			// is never tested against itself).
			if tc.id <= qc.id {
				return true
			}
			// Same-side guard: when self-noding isn't required, skip
			// A-vs-A and B-vs-B chain pairs. The predicate only needs
			// the AB interaction.
			if !requireSelfNoding && tc.ss.IsA == qc.ss.IsA {
				return true
			}
			tc.mc.ComputeOverlaps(qc.mc, 0, func(mc1 *noding.MonotoneChain, s1 int, mc2 *noding.MonotoneChain, s2 int) {
				// mc1 is tc.mc; mc2 is qc.mc (matches the JTS callback
				// orientation). The intersector handles isA-ordering
				// internally.
				intersector.ProcessIntersections(tc.ss, s1, qc.ss, s2)
			})
			if intersector.IsDone() {
				stop = true
				return false
			}
			return true
		})
		if stop {
			return
		}
	}
}
