package planar

import (
	"math"

	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/kernel"
)

// Kernel is the planar implementation of kernel.Kernel.
// It is a stateless value type; callers may use the package-level Default
// rather than instantiating their own.
type Kernel struct{}

// Default is the canonical planar kernel instance.
var Default kernel.Kernel = Kernel{}

func (Kernel) Name() string { return "planar" }

func (Kernel) Distance(a, b geom.XY) float64 {
	return math.Hypot(a.X-b.X, a.Y-b.Y)
}

func (Kernel) DistanceSquared(a, b geom.XY) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	return dx*dx + dy*dy
}

// SegmentIntersection returns the intersection of segment [a1,a2] with
// [b1,b2]. Collinear-overlap segments return ok=false; consumers needing
// to detect the shared sub-segment should call SegmentIntersect (the
// structured form) instead.
func (Kernel) SegmentIntersection(a1, a2, b1, b2 geom.XY) (geom.XY, bool) {
	rx := a2.X - a1.X
	ry := a2.Y - a1.Y
	sx := b2.X - b1.X
	sy := b2.Y - b1.Y

	denom := rx*sy - ry*sx
	if denom == 0 {
		// Parallel or collinear; no unique intersection point.
		return geom.XY{}, false
	}
	tNum := (b1.X-a1.X)*sy - (b1.Y-a1.Y)*sx
	uNum := (b1.X-a1.X)*ry - (b1.Y-a1.Y)*rx
	t := tNum / denom
	u := uNum / denom
	if t < 0 || t > 1 || u < 0 || u > 1 {
		return geom.XY{}, false
	}
	return geom.XY{X: a1.X + t*rx, Y: a1.Y + t*ry}, true
}

// SegmentIntersect is the structured form of SegmentIntersection. Unlike
// the single-point predicate, it distinguishes:
//
//   - NoIntersection — the segments are disjoint.
//   - PointIntersection — the segments meet at exactly one point (P).
//   - CollinearOverlap — the segments are collinear and share the
//     sub-segment [P, Q]. P != Q.
//
// The CollinearOverlap branch is the load-bearing one for noding shared
// boundary edges between adjacent polygons; SegmentIntersection returns
// false for those cases, leaving topology graphs disconnected.
//
// Note: this is not on the Kernel interface. Spherical/geodesic noding
// is not yet implemented; once it is, those kernels will grow their own
// SegmentIntersect with the same return shape.
func (k Kernel) SegmentIntersect(a1, a2, b1, b2 geom.XY) kernel.SegmentIntersectionResult {
	rx := a2.X - a1.X
	ry := a2.Y - a1.Y
	sx := b2.X - b1.X
	sy := b2.Y - b1.Y

	denom := rx*sy - ry*sx
	if denom != 0 {
		// Single-point intersection candidate.
		tNum := (b1.X-a1.X)*sy - (b1.Y-a1.Y)*sx
		uNum := (b1.X-a1.X)*ry - (b1.Y-a1.Y)*rx
		t := tNum / denom
		u := uNum / denom
		if t < 0 || t > 1 || u < 0 || u > 1 {
			return kernel.SegmentIntersectionResult{Kind: kernel.NoIntersection}
		}
		// Endpoint snapping: when an exact orient check makes one of the
		// parameters vanish, the intersection coincides with that
		// endpoint. Returning the parametric reconstruction would carry
		// rounding error and break downstream noding (the same vertex
		// re-emerges as two near-duplicate nodes, disconnecting the
		// topology graph). Snap to the exact endpoint instead.
		switch {
		case tNum == 0:
			return kernel.SegmentIntersectionResult{Kind: kernel.PointIntersection, P: a1}
		case tNum == denom:
			return kernel.SegmentIntersectionResult{Kind: kernel.PointIntersection, P: a2}
		case uNum == 0:
			return kernel.SegmentIntersectionResult{Kind: kernel.PointIntersection, P: b1}
		case uNum == denom:
			return kernel.SegmentIntersectionResult{Kind: kernel.PointIntersection, P: b2}
		}
		return kernel.SegmentIntersectionResult{
			Kind: kernel.PointIntersection,
			P:    geom.XY{X: a1.X + t*rx, Y: a1.Y + t*ry},
		}
	}

	// Parallel: collinear iff b1 lies on line(a1,a2). Test via the
	// adaptive Orient predicate so this is sign-correct for all float64
	// inputs that don't overflow.
	if adaptiveOrient(a1, a2, b1) != kernel.Collinear {
		return kernel.SegmentIntersectionResult{Kind: kernel.NoIntersection}
	}

	// Collinear: parameterise b1, b2 along (a1, a2) and intersect the
	// 1D intervals [0, 1] and [t1, t2]. Use the dominant axis to avoid
	// dividing by a near-zero |a2-a1|^2 when the segment is degenerate.
	if rx == 0 && ry == 0 {
		// Degenerate a-segment: a is a point. Check whether it lies on b.
		if pointOnSegment(a1, b1, b2) {
			return kernel.SegmentIntersectionResult{Kind: kernel.PointIntersection, P: a1}
		}
		return kernel.SegmentIntersectionResult{Kind: kernel.NoIntersection}
	}

	// t parameter of b1 and b2 on segment [a1, a2].
	var tB1, tB2 float64
	if math.Abs(rx) >= math.Abs(ry) {
		tB1 = (b1.X - a1.X) / rx
		tB2 = (b2.X - a1.X) / rx
	} else {
		tB1 = (b1.Y - a1.Y) / ry
		tB2 = (b2.Y - a1.Y) / ry
	}
	tMin, tMax := tB1, tB2
	if tMin > tMax {
		tMin, tMax = tMax, tMin
	}
	// Intersect [tMin, tMax] with [0, 1].
	lo := math.Max(0, tMin)
	hi := math.Min(1, tMax)
	if lo > hi {
		return kernel.SegmentIntersectionResult{Kind: kernel.NoIntersection}
	}
	if lo == hi {
		return kernel.SegmentIntersectionResult{
			Kind: kernel.PointIntersection,
			P:    geom.XY{X: a1.X + lo*rx, Y: a1.Y + lo*ry},
		}
	}
	return kernel.SegmentIntersectionResult{
		Kind: kernel.CollinearOverlap,
		P:    geom.XY{X: a1.X + lo*rx, Y: a1.Y + lo*ry},
		Q:    geom.XY{X: a1.X + hi*rx, Y: a1.Y + hi*ry},
	}
}

// pointOnSegment reports whether p lies on the closed segment [a, b].
// Used by the degenerate-segment branch of SegmentIntersect.
func pointOnSegment(p, a, b geom.XY) bool { return onSegment(p, a, b) }

// SegmentIntersect is the package-level convenience that calls
// Kernel{}.SegmentIntersect; useful for the noders, which already wire
// to planar primitives directly rather than through the Kernel interface.
func SegmentIntersect(a1, a2, b1, b2 geom.XY) kernel.SegmentIntersectionResult {
	return Kernel{}.SegmentIntersect(a1, a2, b1, b2)
}

// SegmentDistance returns the shortest distance from p to segment [a,b].
// For a degenerate segment (a == b) it falls back to the point-distance.
func (k Kernel) SegmentDistance(p, a, b geom.XY) float64 {
	dx := b.X - a.X
	dy := b.Y - a.Y
	lenSq := dx*dx + dy*dy
	if lenSq == 0 {
		return k.Distance(p, a)
	}
	t := ((p.X-a.X)*dx + (p.Y-a.Y)*dy) / lenSq
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	closest := geom.XY{X: a.X + t*dx, Y: a.Y + t*dy}
	return k.Distance(p, closest)
}

// PointToSegmentSq returns the squared shortest distance from p to segment
// [a,b]. Equivalent to SegmentDistance(p,a,b)² but avoids the sqrt — useful
// for inner-loop comparisons where only relative ordering is needed.
//
// Mirrors JTS Distance.pointToSegmentSq (org.locationtech.jts.algorithm.Distance).
func (k Kernel) PointToSegmentSq(p, a, b geom.XY) float64 {
	if a.X == b.X && a.Y == b.Y {
		return k.DistanceSquared(p, a)
	}
	dx := b.X - a.X
	dy := b.Y - a.Y
	lenSq := dx*dx + dy*dy
	r := ((p.X-a.X)*dx + (p.Y-a.Y)*dy) / lenSq
	if r <= 0 {
		return k.DistanceSquared(p, a)
	}
	if r >= 1 {
		return k.DistanceSquared(p, b)
	}
	// Perpendicular component (signed cross / |AB|²) times |AB| gives the
	// perpendicular distance; squared form is s²·|AB|².
	s := ((a.Y-p.Y)*dx - (a.X-p.X)*dy) / lenSq
	return s * s * lenSq
}

// PointToLinePerpendicular returns the perpendicular distance from p to the
// INFINITE line through a and b (NOT clamped to the segment). a and b must
// be distinct.
//
// Mirrors JTS Distance.pointToLinePerpendicular (org.locationtech.jts.algorithm.Distance).
func (Kernel) PointToLinePerpendicular(p, a, b geom.XY) float64 {
	dx := b.X - a.X
	dy := b.Y - a.Y
	lenSq := dx*dx + dy*dy
	s := ((a.Y-p.Y)*dx - (a.X-p.X)*dy) / lenSq
	return math.Abs(s) * math.Sqrt(lenSq)
}

// Orient classifies the turn (a, b, c) using a Shewchuk-style adaptive
// 2D orientation predicate (see robust.go).
//
// Common case: native float64 cost. Near-collinear inputs (where the
// sign of the determinant cannot be safely decided in double precision)
// are recomputed at 113-bit precision via math/big and the exact sign is
// returned. The result is therefore correct for ALL float64 inputs that
// don't overflow.
func (Kernel) Orient(a, b, c geom.XY) kernel.Orientation {
	return adaptiveOrient(a, b, c)
}

// PointInRing implements the standard ray-cast (crossing-number) test.
// The ring is assumed closed (first vertex == last). Boundary points are
// detected by an explicit segment-distance check with epsilon zero — they
// must lie exactly on the segment to be reported OnBoundary.
func (k Kernel) PointInRing(p geom.XY, ring []geom.XY) kernel.Containment {
	if len(ring) < 4 {
		return kernel.Outside
	}
	inside := false
	for i := 0; i < len(ring)-1; i++ {
		a := ring[i]
		b := ring[i+1]
		// Boundary check first.
		if onSegment(p, a, b) {
			return kernel.OnBoundary
		}
		if (a.Y > p.Y) != (b.Y > p.Y) {
			xCross := a.X + (p.Y-a.Y)*(b.X-a.X)/(b.Y-a.Y)
			if p.X < xCross {
				inside = !inside
			}
		}
	}
	if inside {
		return kernel.Inside
	}
	return kernel.Outside
}

func onSegment(p, a, b geom.XY) bool {
	if (b.X-a.X)*(p.Y-a.Y)-(b.Y-a.Y)*(p.X-a.X) != 0 {
		return false
	}
	if p.X < math.Min(a.X, b.X) || p.X > math.Max(a.X, b.X) {
		return false
	}
	if p.Y < math.Min(a.Y, b.Y) || p.Y > math.Max(a.Y, b.Y) {
		return false
	}
	return true
}

// InitialBearing returns the bearing from a to b in degrees clockwise from
// +Y (planar "north"). Result is in [0, 360).
func (Kernel) InitialBearing(a, b geom.XY) float64 {
	dx := b.X - a.X
	dy := b.Y - a.Y
	deg := math.Atan2(dx, dy) * 180 / math.Pi
	if deg < 0 {
		deg += 360
	}
	return deg
}

// Destination returns the point at distance and bearing from from.
// Bearing is measured clockwise from +Y.
func (Kernel) Destination(from geom.XY, bearingDeg, distance float64) geom.XY {
	rad := bearingDeg * math.Pi / 180
	return geom.XY{
		X: from.X + distance*math.Sin(rad),
		Y: from.Y + distance*math.Cos(rad),
	}
}

// RingArea returns the signed shoelace area. Positive for CCW rings;
// negative for CW rings.
func (Kernel) RingArea(ring []geom.XY) float64 {
	if len(ring) < 3 {
		return 0
	}
	var sum float64
	for i := 0; i < len(ring)-1; i++ {
		sum += ring[i].X*ring[i+1].Y - ring[i+1].X*ring[i].Y
	}
	return sum / 2
}

func (Kernel) Midpoint(a, b geom.XY) geom.XY {
	return geom.XY{X: (a.X + b.X) / 2, Y: (a.Y + b.Y) / 2}
}

// AngleBetween returns the interior angle at b formed by a-b-c, in radians,
// in [0, pi].
func (Kernel) AngleBetween(a, b, c geom.XY) float64 {
	v1x, v1y := a.X-b.X, a.Y-b.Y
	v2x, v2y := c.X-b.X, c.Y-b.Y
	dot := v1x*v2x + v1y*v2y
	mag := math.Hypot(v1x, v1y) * math.Hypot(v2x, v2y)
	if mag == 0 {
		return 0
	}
	cos := dot / mag
	if cos > 1 {
		cos = 1
	} else if cos < -1 {
		cos = -1
	}
	return math.Acos(cos)
}

// SinSnap returns sin(a) with values whose magnitude is below 5e-16 snapped
// to exactly 0. At multiples of π this lets callers obtain a clean 0
// instead of ~1e-16 noise — useful for buffer/offset construction where the
// noise propagates into geometry coordinates.
//
// Mirrors JTS Angle.sinSnap (org.locationtech.jts.algorithm.Angle).
func (Kernel) SinSnap(a float64) float64 {
	res := math.Sin(a)
	if math.Abs(res) < 5e-16 {
		return 0
	}
	return res
}

// CosSnap returns cos(a) with values whose magnitude is below 5e-16 snapped
// to exactly 0. At odd multiples of π/2 this lets callers obtain a clean 0
// instead of ~6e-17 noise.
//
// Mirrors JTS Angle.cosSnap (org.locationtech.jts.algorithm.Angle).
func (Kernel) CosSnap(a float64) float64 {
	res := math.Cos(a)
	if math.Abs(res) < 5e-16 {
		return 0
	}
	return res
}

// AngleBetweenOriented returns the oriented (signed) smallest angle between
// the two vectors (vertex -> tip0) and (vertex -> tip1), in radians, in the
// range (-π, π].
//
// A positive result corresponds to a counterclockwise rotation from v0 to v1;
// a negative result to a clockwise rotation; zero means no rotation.
//
// Mirrors JTS Angle.angleBetweenOriented (org.locationtech.jts.algorithm.Angle).
func (Kernel) AngleBetweenOriented(tip0, vertex, tip1 geom.XY) float64 {
	a1 := math.Atan2(tip0.Y-vertex.Y, tip0.X-vertex.X)
	a2 := math.Atan2(tip1.Y-vertex.Y, tip1.X-vertex.X)
	angDel := a2 - a1
	if angDel <= -math.Pi {
		return angDel + 2*math.Pi
	}
	if angDel > math.Pi {
		return angDel - 2*math.Pi
	}
	return angDel
}
