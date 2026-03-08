package algorithm

import (
	"math"

	"github.com/robert-malhotra/go-topology-suite/geom"
)

// LineIntersectionResult represents the result of a line intersection computation.
type LineIntersectionResult struct {
	// HasIntersection is true if the lines intersect.
	HasIntersection bool

	// IsProper is true if the intersection is in the interior of both segments.
	IsProper bool

	// IsCollinear is true if the segments are collinear.
	IsCollinear bool

	// Intersection is the intersection point (if any).
	Intersection geom.Coordinate

	// Second intersection point for collinear overlapping segments.
	Intersection2 geom.Coordinate
}

// LineIntersection computes the intersection point of two line segments.
// Returns the intersection point and whether the intersection is proper.
func LineIntersection(p1, p2, p3, p4 geom.Coordinate) LineIntersectionResult {
	result := LineIntersectionResult{
		Intersection2: geom.NewCoordinateNaN(), // Initialize to NaN so IsNaN() check works
	}

	// Compute determinants
	dx1 := p2.X - p1.X
	dy1 := p2.Y - p1.Y
	dx2 := p4.X - p3.X
	dy2 := p4.Y - p3.Y

	denom := dx1*dy2 - dy1*dx2

	// Check for parallel lines
	if math.Abs(denom) < geom.DefaultEpsilon {
		// Lines are parallel - check for collinearity
		if OrientationIndex(p1, p2, p3) == Collinear && OrientationIndex(p1, p2, p4) == Collinear {
			result.IsCollinear = true
			// Check for overlap
			result = computeCollinearIntersection(p1, p2, p3, p4)
		}
		return result
	}

	dx3 := p1.X - p3.X
	dy3 := p1.Y - p3.Y

	// Compute parameters
	t := (dx2*dy3 - dy2*dx3) / denom
	s := (dx1*dy3 - dy1*dx3) / denom

	// Check if intersection is within both segments
	if t >= -geom.DefaultEpsilon && t <= 1+geom.DefaultEpsilon &&
		s >= -geom.DefaultEpsilon && s <= 1+geom.DefaultEpsilon {
		result.HasIntersection = true
		result.Intersection = geom.NewCoordinate(p1.X+t*dx1, p1.Y+t*dy1)

		// Check if proper (not at endpoints)
		result.IsProper = t > geom.DefaultEpsilon && t < 1-geom.DefaultEpsilon &&
			s > geom.DefaultEpsilon && s < 1-geom.DefaultEpsilon
	}

	return result
}

func computeCollinearIntersection(p1, p2, p3, p4 geom.Coordinate) LineIntersectionResult {
	result := LineIntersectionResult{
		IsCollinear:   true,
		Intersection2: geom.NewCoordinateNaN(), // Initialize to NaN so IsNaN() check works
	}

	// Project all points onto the same axis
	// Use the axis with larger extent
	dx := math.Max(math.Abs(p2.X-p1.X), math.Abs(p4.X-p3.X))
	dy := math.Max(math.Abs(p2.Y-p1.Y), math.Abs(p4.Y-p3.Y))

	var t1, t2, t3, t4 float64

	if dx > dy {
		// Project onto X axis
		t1 = 0
		t2 = 1
		t3 = (p3.X - p1.X) / (p2.X - p1.X)
		t4 = (p4.X - p1.X) / (p2.X - p1.X)
	} else if dy > 0 {
		// Project onto Y axis
		t1 = 0
		t2 = 1
		t3 = (p3.Y - p1.Y) / (p2.Y - p1.Y)
		t4 = (p4.Y - p1.Y) / (p2.Y - p1.Y)
	} else {
		// Both segments are points
		if p1.Equals2D(p3, geom.DefaultEpsilon) {
			result.HasIntersection = true
			result.Intersection = p1
		}
		return result
	}

	// Ensure t3 <= t4
	if t3 > t4 {
		t3, t4 = t4, t3
	}

	// Find overlap
	tMin := math.Max(t1, t3)
	tMax := math.Min(t2, t4)

	if tMin > tMax+geom.DefaultEpsilon {
		// No overlap
		return result
	}

	result.HasIntersection = true

	// Compute intersection point(s)
	result.Intersection = geom.NewCoordinate(
		p1.X+tMin*(p2.X-p1.X),
		p1.Y+tMin*(p2.Y-p1.Y),
	)

	if tMin < tMax-geom.DefaultEpsilon {
		// There's a second intersection point (overlap)
		result.Intersection2 = geom.NewCoordinate(
			p1.X+tMax*(p2.X-p1.X),
			p1.Y+tMax*(p2.Y-p1.Y),
		)
	}

	return result
}

// LineLineIntersection computes the intersection of two infinite lines.
// Returns the intersection point and whether the lines are parallel.
func LineLineIntersection(p1, p2, p3, p4 geom.Coordinate) (geom.Coordinate, bool) {
	dx1 := p2.X - p1.X
	dy1 := p2.Y - p1.Y
	dx2 := p4.X - p3.X
	dy2 := p4.Y - p3.Y

	denom := dx1*dy2 - dy1*dx2

	if math.Abs(denom) < geom.DefaultEpsilon {
		// Lines are parallel
		return geom.Coordinate{}, false
	}

	dx3 := p1.X - p3.X
	dy3 := p1.Y - p3.Y

	t := (dx2*dy3 - dy2*dx3) / denom

	return geom.NewCoordinate(p1.X+t*dx1, p1.Y+t*dy1), true
}

// RaySegmentIntersection computes where a ray intersects a segment.
// The ray starts at origin and goes in direction dir.
// Returns the intersection point, parameter t (distance along ray), and whether there's an intersection.
func RaySegmentIntersection(origin, dir, segStart, segEnd geom.Coordinate) (geom.Coordinate, float64, bool) {
	dx := segEnd.X - segStart.X
	dy := segEnd.Y - segStart.Y

	denom := dir.X*dy - dir.Y*dx

	if math.Abs(denom) < geom.DefaultEpsilon {
		// Ray and segment are parallel
		return geom.Coordinate{}, 0, false
	}

	dx2 := origin.X - segStart.X
	dy2 := origin.Y - segStart.Y

	t := (dx*dy2 - dy*dx2) / denom
	s := (dir.X*dy2 - dir.Y*dx2) / denom

	if t < -geom.DefaultEpsilon || s < -geom.DefaultEpsilon || s > 1+geom.DefaultEpsilon {
		return geom.Coordinate{}, 0, false
	}

	intersection := geom.NewCoordinate(origin.X+t*dir.X, origin.Y+t*dir.Y)
	return intersection, t, true
}

// ProjectPointOntoLine projects a point onto an infinite line.
func ProjectPointOntoLine(p, lineStart, lineEnd geom.Coordinate) geom.Coordinate {
	dx := lineEnd.X - lineStart.X
	dy := lineEnd.Y - lineStart.Y
	lenSq := dx*dx + dy*dy

	if lenSq < geom.DefaultEpsilon*geom.DefaultEpsilon {
		return lineStart
	}

	t := ((p.X-lineStart.X)*dx + (p.Y-lineStart.Y)*dy) / lenSq

	return geom.NewCoordinate(lineStart.X+t*dx, lineStart.Y+t*dy)
}

// ProjectPointOntoSegment projects a point onto a line segment.
func ProjectPointOntoSegment(p, segStart, segEnd geom.Coordinate) geom.Coordinate {
	return geom.ClosestPointOnSegment(p, segStart, segEnd)
}

// ReflectPointOverLine reflects a point over a line.
func ReflectPointOverLine(p, lineStart, lineEnd geom.Coordinate) geom.Coordinate {
	proj := ProjectPointOntoLine(p, lineStart, lineEnd)
	return geom.NewCoordinate(2*proj.X-p.X, 2*proj.Y-p.Y)
}
