package kernel

import "github.com/terra-geo/terra/geom"

// Orientation classifies the turn formed by three points a, b, c when
// traversed in that order.
type Orientation int8

const (
	Collinear        Orientation = 0
	CounterClockwise Orientation = 1
	Clockwise        Orientation = -1
)

// String returns "CCW", "CW", or "Collinear".
func (o Orientation) String() string {
	switch o {
	case CounterClockwise:
		return "CCW"
	case Clockwise:
		return "CW"
	default:
		return "Collinear"
	}
}

// Containment is the result of a point-in-region test.
type Containment uint8

const (
	Outside Containment = iota
	OnBoundary
	Inside
)

// Kernel is the strategy interface implemented by planar, spherical, and
// geodesic kernels. Every higher-level Terra operation that needs a
// geometric primitive — overlay, predicates, measurements — calls through
// this interface (with the option of generic specialisation for hot paths).
//
// All twelve primitives below are required. Implementations panic only on
// truly programmer-error inputs (NaN coordinates outside an explicitly
// NaN-tolerant primitive); domain edge cases (antipodal points, degenerate
// triangles) return well-defined values documented per primitive.
type Kernel interface {
	// Name returns a stable identifier ("planar", "spherical", "geodesic").
	Name() string

	// Distance returns the kernel-appropriate distance between a and b.
	// Planar: Euclidean. Spherical/Geodesic: great-circle / geodesic in
	// metres on the WGS84 ellipsoid.
	Distance(a, b geom.XY) float64

	// DistanceSquared returns Distance squared, allowing callers to skip
	// a sqrt in tight loops where only ordering matters.
	DistanceSquared(a, b geom.XY) float64

	// SegmentIntersection returns the intersection point of segment [a1,a2]
	// with [b1,b2] and a flag indicating whether they intersect at all.
	// Collinear-overlap segments return ok=false (callers needing the full
	// relate-9IM should use the relate package).
	SegmentIntersection(a1, a2, b1, b2 geom.XY) (geom.XY, bool)

	// SegmentDistance returns the shortest distance from p to segment [a,b].
	SegmentDistance(p, a, b geom.XY) float64

	// Orient classifies the turn (a,b,c).
	Orient(a, b, c geom.XY) Orientation

	// PointInRing tests p against a closed ring (first vertex == last).
	// The ring orientation does not affect the result.
	PointInRing(p geom.XY, ring []geom.XY) Containment

	// InitialBearing returns the bearing from a to b in degrees, measured
	// clockwise from true north. For the planar kernel, "north" means +Y.
	InitialBearing(a, b geom.XY) float64

	// Destination returns the point reached by travelling distance metres
	// (or planar units, for the planar kernel) from from at the given
	// bearing in degrees.
	Destination(from geom.XY, bearing, distance float64) geom.XY

	// RingArea returns the signed area of a closed ring. The sign
	// distinguishes orientation: positive for CCW, negative for CW under
	// each kernel's natural convention. (Documented per impl.)
	RingArea(ring []geom.XY) float64

	// Midpoint returns the kernel-appropriate midpoint between a and b.
	// Planar: arithmetic mean. Spherical/Geodesic: midpoint along the
	// shortest path.
	Midpoint(a, b geom.XY) geom.XY

	// AngleBetween returns the interior angle at b formed by points a-b-c,
	// in radians, in [0, pi].
	AngleBetween(a, b, c geom.XY) float64
}
