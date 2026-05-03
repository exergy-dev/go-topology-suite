package snaprounding

import (
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/internal/noding"
	"github.com/terra-geo/terra/kernel"
	"github.com/terra-geo/terra/kernel/planar"
)

var planarK = planar.Kernel{}

// IntersectionAdder is a port of
// org.locationtech.jts.noding.snapround.SnapRoundingIntersectionAdder.
//
// Given a set of input segment strings it computes — at full float
// precision — every segment-segment intersection AND every "near-vertex"
// adjacency (a segment endpoint that lies within nearnessTol of the
// interior of another segment), recording the resulting points so they
// can be seeded into the snap-rounding hot-pixel set BEFORE the snap-
// rounding pass starts. This is JTS's mechanism for converging the
// snap-rounding fix-point in a single pass: every robustness-sensitive
// near-touch is realised as a hot pixel up-front, so the post-snap
// segment-splitting phase has no further intersections to discover.
//
// The class is purely a planning step — it does NOT itself snap, split,
// or modify segments. Callers feed Points() into a HotPixelIndex (or a
// HotPixelSet) and proceed with their normal snap-round pipeline.
type IntersectionAdder struct {
	nearnessTol   float64
	intersections []geom.XY
}

// NewIntersectionAdder returns an empty adder configured with the given
// tolerance distance for vertex-near-segment heuristic adjacencies.
//
// nearnessTol should be SIGNIFICANTLY below the snap-rounding grid
// spacing — JTS uses precisionScale/100 — so the heuristic tightens
// rather than relaxes the geometric intersection set.
func NewIntersectionAdder(nearnessTol float64) *IntersectionAdder {
	return &IntersectionAdder{nearnessTol: nearnessTol}
}

// Process exhaustively pairs every segment of every input string with
// every segment of every other (and itself, excluding self-pairs) and
// records the resulting intersection points and near-vertex adjacencies.
//
// O(N²) in the segment count — fine for the post-noded inputs of the
// snap-rounding pipeline (which already culled most pairs in the prior
// IndexedNoder pass) but not suitable as a general-purpose noder.
func (a *IntersectionAdder) Process(strs []*noding.SegmentString) {
	for i, e0 := range strs {
		for j, e1 := range strs {
			// Allow same-string crossings (j == i) so chains can detect
			// their own self-intersections, but skip the pair where both
			// indices refer to the SAME segment.
			_ = i
			_ = j
			if e0.NumSegments() == 0 || e1.NumSegments() == 0 {
				continue
			}
			for s0 := 0; s0 < e0.NumSegments(); s0++ {
				for s1 := 0; s1 < e1.NumSegments(); s1++ {
					if e0 == e1 && s0 == s1 {
						continue
					}
					a.processSegmentPair(e0, s0, e1, s1)
				}
			}
		}
	}
}

// Points returns the recorded intersection / near-vertex points.
// The slice may contain duplicates; callers feeding it into a
// HotPixelIndex / HotPixelSet rely on the index's own deduplication.
func (a *IntersectionAdder) Points() []geom.XY { return a.intersections }

// processSegmentPair mirrors JTS processIntersections: it computes the
// LineIntersector result for the two segments and records the proper
// (interior) intersection point(s); failing that, it falls back to the
// vertex-near-segment heuristic.
func (a *IntersectionAdder) processSegmentPair(
	e0 *noding.SegmentString, s0 int,
	e1 *noding.SegmentString, s1 int,
) {
	p00, p01 := e0.Segment(s0)
	p10, p11 := e1.Segment(s1)

	res := planarK.SegmentIntersect(p00, p01, p10, p11)

	if res.Kind == kernel.PointIntersection {
		// JTS only records "interior" intersections — points that are
		// not coincident with any of the four segment endpoints.
		if isInteriorIntersection(res.P, p00, p01, p10, p11) {
			a.intersections = append(a.intersections, res.P)
			return
		}
	}
	if res.Kind == kernel.CollinearOverlap {
		// Both endpoints of the overlap can be hot pixels: any segment
		// boundary along the collinear sub-segment must be a node.
		if isInteriorIntersection(res.P, p00, p01, p10, p11) {
			a.intersections = append(a.intersections, res.P)
		}
		if isInteriorIntersection(res.Q, p00, p01, p10, p11) {
			a.intersections = append(a.intersections, res.Q)
		}
		return
	}

	// No proper intersection. Fall back to the four near-vertex tests.
	a.processNearVertex(p00, p10, p11)
	a.processNearVertex(p01, p10, p11)
	a.processNearVertex(p10, p00, p01)
	a.processNearVertex(p11, p00, p01)
}

// isInteriorIntersection reports whether p is NOT coincident with any
// of the four segment endpoints. JTS LineIntersector.isInteriorIntersection
// implements the same predicate.
func isInteriorIntersection(p, p00, p01, p10, p11 geom.XY) bool {
	if p.EqualBitwise(p00) || p.EqualBitwise(p01) ||
		p.EqualBitwise(p10) || p.EqualBitwise(p11) {
		return false
	}
	return true
}

// processNearVertex records p as an intersection iff p lies within
// nearnessTol of the INTERIOR of segment [p0, p1]. p is excluded if it
// is itself near either endpoint of the segment (this avoids zig-zag
// linework as documented in JTS).
func (a *IntersectionAdder) processNearVertex(p, p0, p1 geom.XY) {
	if a.nearnessTol <= 0 {
		return
	}
	tolSq := a.nearnessTol * a.nearnessTol
	if planarK.DistanceSquared(p, p0) < tolSq {
		return
	}
	if planarK.DistanceSquared(p, p1) < tolSq {
		return
	}
	if planarK.PointToSegmentSq(p, p0, p1) < tolSq {
		a.intersections = append(a.intersections, p)
	}
}
