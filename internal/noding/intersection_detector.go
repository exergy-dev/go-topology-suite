package noding

import (
	"github.com/exergy-dev/go-topology-suite/geom"
	"github.com/exergy-dev/go-topology-suite/kernel"
	"github.com/exergy-dev/go-topology-suite/kernel/planar"
)

// SegmentIntersectionDetector is a Boolean "any-intersection?"
// detector across a set of SegmentStrings. Mirrors
// org.locationtech.jts.noding.SegmentIntersectionDetector.
//
// In contrast to a noder, the detector returns as soon as any
// intersection is found — making it substantially faster than running
// a full noder when the only required output is a yes/no flag plus
// (optionally) a witness intersection point.
//
// Configuration mirrors JTS:
//
//   - FindProper: report only proper intersections (interior of both
//     segments). Endpoint-only hits are ignored.
//   - FindAllTypes: report any intersection kind, including collinear
//     overlap and endpoint touches. (Default: true.)
//
// After Detect returns true, Witness exposes the first hit's segment
// endpoints and intersection point.
type SegmentIntersectionDetector struct {
	FindProper   bool
	FindAllTypes bool

	hasIntersection       bool
	hasProperIntersection bool
	intersectionPoint     geom.XY
	witness               [4]geom.XY
}

// NewSegmentIntersectionDetector returns a detector with JTS defaults
// (FindAllTypes=true, FindProper=false).
func NewSegmentIntersectionDetector() *SegmentIntersectionDetector {
	return &SegmentIntersectionDetector{FindAllTypes: true}
}

// HasAnyIntersection is the convenience entry point: returns true iff
// any pair of segments in input intersects (interior or endpoint).
//
// O(n^2) brute force; for large inputs use Detect with a custom
// configuration if early termination is required.
func HasAnyIntersection(input []*SegmentString) bool {
	d := NewSegmentIntersectionDetector()
	d.Detect(input)
	return d.HasIntersection()
}

// Detect runs the detector across input. Returns once the first
// matching intersection is found (subject to FindProper).
func (d *SegmentIntersectionDetector) Detect(input []*SegmentString) {
	d.hasIntersection = false
	d.hasProperIntersection = false
	for i, ss0 := range input {
		n0 := ss0.NumSegments()
		for j0 := 0; j0 < n0; j0++ {
			a0, a1 := ss0.Segment(j0)
			for k := i; k < len(input); k++ {
				ss1 := input[k]
				n1 := ss1.NumSegments()
				j1Start := 0
				if k == i {
					j1Start = j0 + 1
				}
				for j1 := j1Start; j1 < n1; j1++ {
					if k == i && (j1 == j0+1 || j0 == j1+1) {
						// Adjacent edges in the same string share a
						// vertex by construction — skip.
						continue
					}
					b0, b1 := ss1.Segment(j1)
					res := planar.SegmentIntersect(a0, a1, b0, b1)
					if res.Kind == kernel.NoIntersection {
						continue
					}
					proper := false
					if res.Kind == kernel.PointIntersection {
						p := res.P
						proper = p != a0 && p != a1 && p != b0 && p != b1
					}
					if d.FindProper && !proper {
						continue
					}
					if !d.FindAllTypes && res.Kind == kernel.CollinearOverlap {
						continue
					}
					d.hasIntersection = true
					if proper {
						d.hasProperIntersection = true
					}
					d.intersectionPoint = res.P
					d.witness = [4]geom.XY{a0, a1, b0, b1}
					return
				}
			}
		}
	}
}

// HasIntersection reports whether at least one matching intersection
// was found.
func (d *SegmentIntersectionDetector) HasIntersection() bool { return d.hasIntersection }

// HasProperIntersection reports whether the first matching
// intersection was proper (interior of both segments).
func (d *SegmentIntersectionDetector) HasProperIntersection() bool {
	return d.hasProperIntersection
}

// IntersectionPoint returns the first witness intersection point.
// Undefined if HasIntersection is false.
func (d *SegmentIntersectionDetector) IntersectionPoint() geom.XY {
	return d.intersectionPoint
}

// IntersectionSegments returns (a0, a1, b0, b1) for the first
// witness pair. Undefined if HasIntersection is false.
func (d *SegmentIntersectionDetector) IntersectionSegments() (geom.XY, geom.XY, geom.XY, geom.XY) {
	return d.witness[0], d.witness[1], d.witness[2], d.witness[3]
}
