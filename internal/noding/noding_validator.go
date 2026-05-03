package noding

import (
	"fmt"

	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/kernel"
	"github.com/terra-geo/terra/kernel/planar"
)

// NodingValidator is a strict O(n^2) port of
// org.locationtech.jts.noding.NodingValidator.
//
// Whereas FastNodingValidator uses a MonotoneChain index and may rely
// on chain-envelope filtering to reject pairs early, NodingValidator
// performs a brute-force pairwise comparison of every (segment-string,
// segment) pair against every other. It is intended as a CI rail and
// reference oracle: it must catch every interior intersection that
// FastNodingValidator could potentially miss due to index pruning or
// floating-point envelope inflation.
//
// Three failure modes are detected (matching JTS):
//
//   - Endpoint/interior-vertex intersection: an endpoint of one
//     SegmentString equals an interior vertex of another (an
//     unrepresented topology node).
//   - Interior intersection: two segments cross at a point that is in
//     the interior of at least one of them (proper or improper).
//   - Collapse: a SegmentString contains a triplet a-b-a (a self-
//     intersecting cycle through a single edge).
//
// On detection Validate returns a descriptive error. CheckValid is the
// JTS-style alias.
type NodingValidator struct {
	segStrings []*SegmentString
}

// NewNodingValidator creates a validator over the given segment strings.
// The slice is retained by reference; the caller must not mutate it
// during validation.
func NewNodingValidator(segStrings []*SegmentString) *NodingValidator {
	return &NodingValidator{segStrings: segStrings}
}

// CheckValid is the JTS-named entry point. Returns nil if input is
// correctly noded, or an error describing the first problem found.
//
// Order of checks: collapse first (a triplet a-b-a manifests as a
// collinear-overlap in the interior-intersection scan, so we want the
// more specific diagnostic first), then endpoint/interior-vertex,
// then strict pairwise interior intersection.
func (v *NodingValidator) CheckValid() error {
	if err := v.checkCollapses(); err != nil {
		return err
	}
	if err := v.checkEndPtVertexIntersections(); err != nil {
		return err
	}
	if err := v.checkInteriorIntersections(); err != nil {
		return err
	}
	return nil
}

// checkCollapses finds any segment string with three consecutive
// vertices p0-p1-p0, which implies an instantaneous reversal at p1
// (a non-noded self-intersection).
func (v *NodingValidator) checkCollapses() error {
	for _, ss := range v.segStrings {
		if err := checkCollapsesForString(ss); err != nil {
			return err
		}
	}
	return nil
}

func checkCollapsesForString(ss *SegmentString) error {
	pts := ss.Coords
	for i := 0; i+2 < len(pts); i++ {
		if pts[i] == pts[i+2] {
			return fmt.Errorf("found non-noded collapse at LINESTRING(%v %v %v)",
				pts[i], pts[i+1], pts[i+2])
		}
	}
	return nil
}

// checkInteriorIntersections does the strict O(n^2) per-segment check.
// JTS iterates the Cartesian product (i,j) with i,j ranging over every
// SegmentString — and within those, every (segIndex0, segIndex1) edge
// pair. The same SegmentString is compared against itself, EXCEPT when
// the two segment indices are identical (which would compare an edge
// to itself). That is, adjacent edges within a single SegmentString
// ARE compared, mirroring the JTS check.
func (v *NodingValidator) checkInteriorIntersections() error {
	for _, ss0 := range v.segStrings {
		for _, ss1 := range v.segStrings {
			if err := checkInteriorIntersectionsPair(ss0, ss1); err != nil {
				return err
			}
		}
	}
	return nil
}

func checkInteriorIntersectionsPair(e0, e1 *SegmentString) error {
	pts0 := e0.Coords
	pts1 := e1.Coords
	for i0 := 0; i0+1 < len(pts0); i0++ {
		for i1 := 0; i1+1 < len(pts1); i1++ {
			if e0 == e1 && i0 == i1 {
				continue
			}
			if err := checkInteriorIntersectionAt(e0, i0, e1, i1); err != nil {
				return err
			}
		}
	}
	return nil
}

// checkInteriorIntersectionAt tests whether segment (e0,i0) and segment
// (e1,i1) intersect with the intersection point lying interior to at
// least one of them (i.e. NOT only at shared endpoints). Mirrors JTS
// NodingValidator.checkInteriorIntersections(LineSegment, LineSegment).
func checkInteriorIntersectionAt(e0 *SegmentString, i0 int, e1 *SegmentString, i1 int) error {
	p00, p01 := e0.Segment(i0)
	p10, p11 := e1.Segment(i1)
	res := planar.SegmentIntersect(p00, p01, p10, p11)
	if res.Kind == kernel.NoIntersection {
		return nil
	}

	switch res.Kind {
	case kernel.PointIntersection:
		// "Proper" iff the intersection point is in the interior of
		// both segments (matches no endpoint of either).
		ip := res.P
		isProper := ip != p00 && ip != p01 && ip != p10 && ip != p11
		if isProper || hasInteriorIntersection(ip, p00, p01) || hasInteriorIntersection(ip, p10, p11) {
			return fmt.Errorf("found non-noded intersection at %v-%v and %v-%v",
				p00, p01, p10, p11)
		}
	case kernel.CollinearOverlap:
		// Collinear-overlap is by definition an interior intersection
		// of at least one of the two edges (each shared sub-segment
		// endpoint lies interior to at least one input segment).
		return fmt.Errorf("found non-noded collinear overlap between %v-%v and %v-%v",
			p00, p01, p10, p11)
	}
	return nil
}

// hasInteriorIntersection returns true if intPt lies on segment p0-p1
// at a point that is NOT one of the segment's endpoints.
func hasInteriorIntersection(intPt, p0, p1 geom.XY) bool {
	return intPt != p0 && intPt != p1
}

// checkEndPtVertexIntersections finds endpoint vertices of one segment
// string that coincide with interior vertices (index 1..n-2) of
// another. This is JTS's "endpt/interior pt intersection" check.
func (v *NodingValidator) checkEndPtVertexIntersections() error {
	for _, ss := range v.segStrings {
		pts := ss.Coords
		if len(pts) == 0 {
			continue
		}
		if err := v.checkEndPtVertexHit(pts[0]); err != nil {
			return err
		}
		if err := v.checkEndPtVertexHit(pts[len(pts)-1]); err != nil {
			return err
		}
	}
	return nil
}

func (v *NodingValidator) checkEndPtVertexHit(testPt geom.XY) error {
	for _, ss := range v.segStrings {
		pts := ss.Coords
		// JTS scans indices 1..len-2 (inclusive) — the strict interior
		// vertices, excluding the segment string's own endpoints.
		for j := 1; j+1 < len(pts); j++ {
			if pts[j] == testPt {
				return fmt.Errorf("found endpt/interior pt intersection at index %d :pt %v",
					j, testPt)
			}
		}
	}
	return nil
}
