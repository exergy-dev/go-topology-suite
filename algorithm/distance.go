package algorithm

import (
	"math"

	"github.com/go-topology-suite/gts/geom"
)

// Distance computes the minimum distance between two geometries.
func Distance(g1, g2 geom.Geometry) float64 {
	// Quick rejection using envelopes
	env1 := g1.Envelope()
	env2 := g2.Envelope()
	if env1.Distance(env2) > 0 {
		// Envelopes don't touch - can use envelope distance as lower bound
		// but need to compute actual distance
	}

	return computeDistance(g1, g2)
}

func computeDistance(g1, g2 geom.Geometry) float64 {
	// Handle empty geometries
	if g1.IsEmpty() || g2.IsEmpty() {
		return math.Inf(1)
	}

	// Get coordinates of both geometries
	coords1 := g1.Coordinates()
	coords2 := g2.Coordinates()

	// For points
	if len(coords1) == 1 && len(coords2) == 1 {
		return coords1[0].Distance(coords2[0])
	}

	// For point to other geometry
	if len(coords1) == 1 {
		return DistancePointToGeometry(coords1[0], g2)
	}
	if len(coords2) == 1 {
		return DistancePointToGeometry(coords2[0], g1)
	}

	// For line/polygon to line/polygon
	return DistanceGeometryToGeometry(g1, g2)
}

// DistancePointToPoint computes the distance between two points.
func DistancePointToPoint(p1, p2 geom.Coordinate) float64 {
	return p1.Distance(p2)
}

// DistancePointToSegment computes the distance from a point to a line segment.
func DistancePointToSegment(p, a, b geom.Coordinate) float64 {
	if a.Equals2D(b, geom.DefaultEpsilon) {
		return p.Distance(a)
	}

	// Vector from a to b
	dx := b.X - a.X
	dy := b.Y - a.Y
	lenSq := dx*dx + dy*dy

	// Parameter t for the projection of p onto the line
	t := ((p.X-a.X)*dx + (p.Y-a.Y)*dy) / lenSq

	if t < 0 {
		// Closest to point a
		return p.Distance(a)
	}
	if t > 1 {
		// Closest to point b
		return p.Distance(b)
	}

	// Closest to projection on segment
	proj := geom.NewCoordinate(a.X+t*dx, a.Y+t*dy)
	return p.Distance(proj)
}

// DistancePointToLine computes the perpendicular distance from a point to an infinite line.
func DistancePointToLine(p, a, b geom.Coordinate) float64 {
	if a.Equals2D(b, geom.DefaultEpsilon) {
		return p.Distance(a)
	}

	// Using the formula: |cross product| / |line vector|
	dx := b.X - a.X
	dy := b.Y - a.Y
	cross := math.Abs((p.X-a.X)*dy - (p.Y-a.Y)*dx)
	lineLen := math.Sqrt(dx*dx + dy*dy)

	return cross / lineLen
}

// DistancePointToLineString computes the distance from a point to a linestring.
func DistancePointToLineString(p geom.Coordinate, ls *geom.LineString) float64 {
	coords := ls.Coordinates()
	if len(coords) == 0 {
		return math.Inf(1)
	}
	if len(coords) == 1 {
		return p.Distance(coords[0])
	}

	minDist := math.Inf(1)
	for i := 1; i < len(coords); i++ {
		dist := DistancePointToSegment(p, coords[i-1], coords[i])
		if dist < minDist {
			minDist = dist
		}
	}
	return minDist
}

// DistancePointToPolygon computes the distance from a point to a polygon.
// Returns 0 if the point is inside the polygon.
func DistancePointToPolygon(p geom.Coordinate, poly *geom.Polygon) float64 {
	if poly.IsEmpty() {
		return math.Inf(1)
	}

	// Check if point is inside
	if poly.ContainsPoint(p) {
		return 0
	}

	// Distance to exterior ring
	minDist := distancePointToRing(p, poly.ExteriorRing())

	// Distance to holes
	for i := 0; i < poly.NumInteriorRings(); i++ {
		dist := distancePointToRing(p, poly.InteriorRingN(i))
		if dist < minDist {
			minDist = dist
		}
	}

	return minDist
}

func distancePointToRing(p geom.Coordinate, ring *geom.LinearRing) float64 {
	coords := ring.Coordinates()
	if len(coords) < 2 {
		return math.Inf(1)
	}

	minDist := math.Inf(1)
	for i := 1; i < len(coords); i++ {
		dist := DistancePointToSegment(p, coords[i-1], coords[i])
		if dist < minDist {
			minDist = dist
		}
	}
	return minDist
}

// DistancePointToGeometry computes the distance from a point to any geometry.
func DistancePointToGeometry(p geom.Coordinate, g geom.Geometry) float64 {
	switch v := g.(type) {
	case *geom.Point:
		if v.IsEmpty() {
			return math.Inf(1)
		}
		return p.Distance(v.Coordinate())
	case *geom.LineString:
		return DistancePointToLineString(p, v)
	case *geom.LinearRing:
		return distancePointToRing(p, v)
	case *geom.Polygon:
		return DistancePointToPolygon(p, v)
	case *geom.MultiPoint:
		return distancePointToMultiPoint(p, v)
	case *geom.MultiLineString:
		return distancePointToMultiLineString(p, v)
	case *geom.MultiPolygon:
		return distancePointToMultiPolygon(p, v)
	case *geom.GeometryCollection:
		return distancePointToCollection(p, v)
	default:
		return math.Inf(1)
	}
}

func distancePointToMultiPoint(p geom.Coordinate, mp *geom.MultiPoint) float64 {
	if mp.IsEmpty() {
		return math.Inf(1)
	}
	minDist := math.Inf(1)
	for i := 0; i < mp.NumGeometries(); i++ {
		pt := mp.GeometryN(i).(*geom.Point)
		dist := p.Distance(pt.Coordinate())
		if dist < minDist {
			minDist = dist
		}
	}
	return minDist
}

func distancePointToMultiLineString(p geom.Coordinate, mls *geom.MultiLineString) float64 {
	if mls.IsEmpty() {
		return math.Inf(1)
	}
	minDist := math.Inf(1)
	for i := 0; i < mls.NumGeometries(); i++ {
		ls := mls.GeometryN(i).(*geom.LineString)
		dist := DistancePointToLineString(p, ls)
		if dist < minDist {
			minDist = dist
		}
	}
	return minDist
}

func distancePointToMultiPolygon(p geom.Coordinate, mp *geom.MultiPolygon) float64 {
	if mp.IsEmpty() {
		return math.Inf(1)
	}
	minDist := math.Inf(1)
	for i := 0; i < mp.NumGeometries(); i++ {
		poly := mp.GeometryN(i).(*geom.Polygon)
		dist := DistancePointToPolygon(p, poly)
		if dist < minDist {
			minDist = dist
		}
		if dist == 0 {
			return 0
		}
	}
	return minDist
}

func distancePointToCollection(p geom.Coordinate, gc *geom.GeometryCollection) float64 {
	if gc.IsEmpty() {
		return math.Inf(1)
	}
	minDist := math.Inf(1)
	for i := 0; i < gc.NumGeometries(); i++ {
		dist := DistancePointToGeometry(p, gc.GeometryN(i))
		if dist < minDist {
			minDist = dist
		}
		if dist == 0 {
			return 0
		}
	}
	return minDist
}

// DistanceSegmentToSegment computes the distance between two line segments.
func DistanceSegmentToSegment(a1, a2, b1, b2 geom.Coordinate) float64 {
	// Check if segments intersect
	if SegmentsIntersect(a1, a2, b1, b2) {
		return 0
	}

	// Distance from endpoints of each segment to the other segment
	d1 := DistancePointToSegment(a1, b1, b2)
	d2 := DistancePointToSegment(a2, b1, b2)
	d3 := DistancePointToSegment(b1, a1, a2)
	d4 := DistancePointToSegment(b2, a1, a2)

	return math.Min(math.Min(d1, d2), math.Min(d3, d4))
}

// SegmentsIntersect returns true if two line segments intersect.
func SegmentsIntersect(a1, a2, b1, b2 geom.Coordinate) bool {
	o1 := OrientationIndex(a1, a2, b1)
	o2 := OrientationIndex(a1, a2, b2)
	o3 := OrientationIndex(b1, b2, a1)
	o4 := OrientationIndex(b1, b2, a2)

	// General case
	if o1 != o2 && o3 != o4 {
		return true
	}

	// Collinear cases
	if o1 == 0 && onSegment(a1, b1, a2) {
		return true
	}
	if o2 == 0 && onSegment(a1, b2, a2) {
		return true
	}
	if o3 == 0 && onSegment(b1, a1, b2) {
		return true
	}
	if o4 == 0 && onSegment(b1, a2, b2) {
		return true
	}

	return false
}

func onSegment(p, q, r geom.Coordinate) bool {
	return q.X <= math.Max(p.X, r.X) && q.X >= math.Min(p.X, r.X) &&
		q.Y <= math.Max(p.Y, r.Y) && q.Y >= math.Min(p.Y, r.Y)
}

// DistanceGeometryToGeometry computes the distance between two geometries.
func DistanceGeometryToGeometry(g1, g2 geom.Geometry) float64 {
	coords1 := g1.Coordinates()
	coords2 := g2.Coordinates()

	if len(coords1) == 0 || len(coords2) == 0 {
		return math.Inf(1)
	}

	minDist := math.Inf(1)

	// Compare all segments
	for i := 0; i < len(coords1); i++ {
		for j := 0; j < len(coords2); j++ {
			// Point to point
			dist := coords1[i].Distance(coords2[j])
			if dist < minDist {
				minDist = dist
			}

			// Point to segment (if j has next)
			if j+1 < len(coords2) {
				dist = DistancePointToSegment(coords1[i], coords2[j], coords2[j+1])
				if dist < minDist {
					minDist = dist
				}
			}

			// Segment to point (if i has next)
			if i+1 < len(coords1) {
				dist = DistancePointToSegment(coords2[j], coords1[i], coords1[i+1])
				if dist < minDist {
					minDist = dist
				}
			}

			// Segment to segment
			if i+1 < len(coords1) && j+1 < len(coords2) {
				dist = DistanceSegmentToSegment(coords1[i], coords1[i+1], coords2[j], coords2[j+1])
				if dist < minDist {
					minDist = dist
				}
			}
		}
	}

	return minDist
}

// IsWithinDistance returns true if two geometries are within the specified distance.
func IsWithinDistance(g1, g2 geom.Geometry, distance float64) bool {
	// Quick check using envelopes
	env1 := g1.Envelope()
	env2 := g2.Envelope()
	if env1.Distance(env2) > distance {
		return false
	}

	return Distance(g1, g2) <= distance
}

// NearestPoints returns the nearest points on two geometries.
func NearestPoints(g1, g2 geom.Geometry) (geom.Coordinate, geom.Coordinate) {
	coords1 := g1.Coordinates()
	coords2 := g2.Coordinates()

	if len(coords1) == 0 || len(coords2) == 0 {
		return geom.Coordinate{X: math.NaN(), Y: math.NaN()},
			geom.Coordinate{X: math.NaN(), Y: math.NaN()}
	}

	minDist := math.Inf(1)
	var nearest1, nearest2 geom.Coordinate

	for i := 0; i < len(coords1); i++ {
		for j := 0; j < len(coords2); j++ {
			dist := coords1[i].Distance(coords2[j])
			if dist < minDist {
				minDist = dist
				nearest1 = coords1[i]
				nearest2 = coords2[j]
			}
		}
	}

	return nearest1, nearest2
}
