// Port of org.locationtech.jts.noding.SegmentSetMutualIntersector and its
// MCIndex-based implementation. Used for the "red/blue" intersection
// problem: compute every candidate intersecting segment pair where one
// segment is drawn from a fixed BASE set and the other from a QUERY set.
//
// JTS uses this for prepared-geometry-vs-other-geometry workflows: the
// base SegmentStrings are indexed once, then any number of query
// geometries can be tested against them. Within either set, segments
// are assumed to only meet at endpoints, so intra-set pairs are NOT
// tested.
//
// SegmentIntersector is the JTS callback invoked for each candidate
// pair. Implementations may either record hits (for predicate logic)
// or inject intersection nodes back into the SegmentStrings (for
// noding). Since JTS's MCIndex-based intersector hands the actual
// planar intersection to the caller's SegmentIntersector, the
// intersector itself just reports candidate (string, segment-index)
// pairs whose envelopes overlap.

package noding

import (
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/index"
)

// SegmentIntersector is the callback invoked for every candidate pair
// of segments produced by a SegmentSetMutualIntersector.
//
// The receiver may record observations (e.g. predicate hits) or
// imperatively node the input strings. When the callback no longer
// needs further pairs it may return true from IsDone (mirroring JTS's
// SegmentIntersector.isDone()): the host intersector will then short-
// circuit the rest of the traversal.
type SegmentIntersector interface {
	// ProcessIntersections reports a candidate pair: segment s1[i1] and
	// segment s2[i2]. The (i1, i2) indices are the start vertices of
	// the segments within their owning SegmentStrings.
	ProcessIntersections(s1 *SegmentString, i1 int, s2 *SegmentString, i2 int)
	// IsDone allows early termination. Implementations that always
	// process every pair should return false.
	IsDone() bool
}

// SegmentSetMutualIntersector is the red/blue intersection detector.
// Implementations hold a fixed BASE set of segments (provided at
// construction) and process zero or more query sets against it.
type SegmentSetMutualIntersector interface {
	// Process visits every candidate (base, query) segment pair using
	// the given SegmentIntersector.
	Process(query []*SegmentString, si SegmentIntersector)
}

// SimpleSegmentSetMutualIntersector is a brute-force implementation:
// every base segment is tested against every query segment. Useful as
// a reference and on small inputs.
type SimpleSegmentSetMutualIntersector struct {
	base []*SegmentString
}

// NewSimpleSegmentSetMutualIntersector returns an intersector with the
// given base set. The base set is captured by reference; do not mutate
// it while the intersector is in use.
func NewSimpleSegmentSetMutualIntersector(base []*SegmentString) *SimpleSegmentSetMutualIntersector {
	return &SimpleSegmentSetMutualIntersector{base: base}
}

// Process implements SegmentSetMutualIntersector.
func (s *SimpleSegmentSetMutualIntersector) Process(query []*SegmentString, si SegmentIntersector) {
	for _, qs := range query {
		nq := qs.NumSegments()
		for j := 0; j < nq; j++ {
			qa, qb := qs.Segment(j)
			qenv := geom.SegmentEnvelope(qa, qb)
			for _, bs := range s.base {
				nb := bs.NumSegments()
				for i := 0; i < nb; i++ {
					ba, bb := bs.Segment(i)
					if !qenv.Intersects(geom.SegmentEnvelope(ba, bb)) {
						continue
					}
					si.ProcessIntersections(bs, i, qs, j)
					if si.IsDone() {
						return
					}
				}
			}
		}
	}
}

// chainEntry pairs a MonotoneChain with the SegmentString it was built
// from, so the SegmentIntersector callback can be invoked with the
// owning string (rather than the chain) — matching JTS's
// MCIndexSegmentSetMutualIntersector contract.
type chainEntry struct {
	owner *SegmentString
	chain *MonotoneChain
}

// MCIndexSegmentSetMutualIntersector indexes the base set's monotone
// chains in an R-tree once, then for each query-side chain queries the
// tree and uses the chain-pair binary subdivision (MonotoneChain.
// ComputeOverlaps) to find candidate segment pairs.
//
// This is a Go port of org.locationtech.jts.noding.MCIndexSegmentSetMutualIntersector.
// The implementation reuses our existing MonotoneChain helpers for
// chain construction and the index package for the spatial index.
type MCIndexSegmentSetMutualIntersector struct {
	OverlapTolerance float64
	tree             *index.RTree[chainEntry]
}

// NewMCIndexSegmentSetMutualIntersector builds the MC index over base.
// The base set is captured by reference; do not mutate it after this
// call.
func NewMCIndexSegmentSetMutualIntersector(base []*SegmentString) *MCIndexSegmentSetMutualIntersector {
	return NewMCIndexSegmentSetMutualIntersectorTol(base, 0)
}

// NewMCIndexSegmentSetMutualIntersectorTol builds the MC index over
// base with the given overlap tolerance. The tolerance inflates chain
// envelopes during the index lookup so segment pairs separated by less
// than tol metres also surface (used by snapping/rounding workflows).
func NewMCIndexSegmentSetMutualIntersectorTol(base []*SegmentString, overlapTolerance float64) *MCIndexSegmentSetMutualIntersector {
	m := &MCIndexSegmentSetMutualIntersector{OverlapTolerance: overlapTolerance}
	var items []index.Item[chainEntry]
	for _, ss := range base {
		if ss.NumSegments() == 0 {
			continue
		}
		for _, mc := range BuildMonotoneChains(ss) {
			items = append(items, index.Item[chainEntry]{
				Env:   mc.EnvelopeExpanded(overlapTolerance),
				Value: chainEntry{owner: ss, chain: mc},
			})
		}
	}
	m.tree = index.New[chainEntry]()
	if len(items) > 0 {
		m.tree.Bulk(items)
	}
	return m
}

// Process implements SegmentSetMutualIntersector.
func (m *MCIndexSegmentSetMutualIntersector) Process(query []*SegmentString, si SegmentIntersector) {
	if m.tree == nil {
		return
	}
	for _, qs := range query {
		if qs.NumSegments() == 0 {
			continue
		}
		for _, qChain := range BuildMonotoneChains(qs) {
			queryEnv := qChain.EnvelopeExpanded(m.OverlapTolerance)
			done := false
			m.tree.Search(queryEnv, func(it index.Item[chainEntry]) bool {
				if done {
					return false
				}
				baseEntry := it.Value
				qChain.ComputeOverlaps(baseEntry.chain, m.OverlapTolerance, func(_ *MonotoneChain, qStart int, _ *MonotoneChain, bStart int) {
					if done {
						return
					}
					si.ProcessIntersections(baseEntry.owner, bStart, qs, qStart)
					if si.IsDone() {
						done = true
					}
				})
				return !done
			})
			if done {
				return
			}
		}
	}
}
