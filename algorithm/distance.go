package algorithm

import (
	"math"

	"github.com/robert-malhotra/go-topology-suite/geom"
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
//
// Deprecated: Use p1.Distance(p2) instead.
func DistancePointToPoint(p1, p2 geom.Coordinate) float64 {
	return p1.Distance(p2)
}

// DistancePointToSegment computes the distance from a point to a line segment.
func DistancePointToSegment(p, a, b geom.Coordinate) float64 {
	closest := closestPointOnSegmentCoord(p, a, b)
	return p.Distance(closest)
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
	return geom.SegmentsIntersect(a1, a2, b1, b2)
}

// Segment represents a line segment defined by two coordinates.
type Segment = geom.Segment

// getGeometryPoints returns all vertex coordinates from a geometry.
func getGeometryPoints(g geom.Geometry) []geom.Coordinate {
	return g.Coordinates()
}

// DistanceGeometryToGeometry computes the distance between two geometries.
// This function properly handles polygon rings and multi-geometries by
// extracting actual segments without creating phantom segments between rings.
func DistanceGeometryToGeometry(g1, g2 geom.Geometry) float64 {
	if g1.IsEmpty() || g2.IsEmpty() {
		return math.Inf(1)
	}

	// Get all vertex points
	points1 := getGeometryPoints(g1)
	points2 := getGeometryPoints(g2)

	if len(points1) == 0 || len(points2) == 0 {
		return math.Inf(1)
	}

	// Get all real segments (respecting geometry boundaries)
	segments1 := geom.GeometrySegments(g1)
	segments2 := geom.GeometrySegments(g2)

	minDist := math.Inf(1)

	// Point to point distances
	for _, p1 := range points1 {
		for _, p2 := range points2 {
			dist := p1.Distance(p2)
			if dist < minDist {
				minDist = dist
			}
		}
	}

	// Point to segment distances (points from g1 to segments of g2)
	for _, p := range points1 {
		for _, seg := range segments2 {
			dist := DistancePointToSegment(p, seg.P0, seg.P1)
			if dist < minDist {
				minDist = dist
			}
		}
	}

	// Point to segment distances (points from g2 to segments of g1)
	for _, p := range points2 {
		for _, seg := range segments1 {
			dist := DistancePointToSegment(p, seg.P0, seg.P1)
			if dist < minDist {
				minDist = dist
			}
		}
	}

	// Segment to segment distances
	for _, seg1 := range segments1 {
		for _, seg2 := range segments2 {
			dist := DistanceSegmentToSegment(seg1.P0, seg1.P1, seg2.P0, seg2.P1)
			if dist < minDist {
				minDist = dist
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
// This function properly handles polygon rings and multi-geometries by
// extracting actual segments without creating phantom segments between rings.
// It finds the closest point on each geometry, including points on segment interiors.
func NearestPoints(g1, g2 geom.Geometry) (geom.Coordinate, geom.Coordinate) {
	if g1.IsEmpty() || g2.IsEmpty() {
		return geom.Coordinate{X: math.NaN(), Y: math.NaN()},
			geom.Coordinate{X: math.NaN(), Y: math.NaN()}
	}

	points1 := getGeometryPoints(g1)
	points2 := getGeometryPoints(g2)

	if len(points1) == 0 || len(points2) == 0 {
		return geom.Coordinate{X: math.NaN(), Y: math.NaN()},
			geom.Coordinate{X: math.NaN(), Y: math.NaN()}
	}

	segments1 := geom.GeometrySegments(g1)
	segments2 := geom.GeometrySegments(g2)

	minDist := math.Inf(1)
	var nearest1, nearest2 geom.Coordinate

	// Point to point distances
	for _, p1 := range points1 {
		for _, p2 := range points2 {
			dist := p1.Distance(p2)
			if dist < minDist {
				minDist = dist
				nearest1 = p1
				nearest2 = p2
			}
		}
	}

	// Point from g1 to segments of g2: find closest point on segment
	for _, p := range points1 {
		for _, seg := range segments2 {
			closestOnSeg := closestPointOnSegmentCoord(p, seg.P0, seg.P1)
			dist := p.Distance(closestOnSeg)
			if dist < minDist {
				minDist = dist
				nearest1 = p
				nearest2 = closestOnSeg
			}
		}
	}

	// Point from g2 to segments of g1: find closest point on segment
	for _, p := range points2 {
		for _, seg := range segments1 {
			closestOnSeg := closestPointOnSegmentCoord(p, seg.P0, seg.P1)
			dist := p.Distance(closestOnSeg)
			if dist < minDist {
				minDist = dist
				nearest1 = closestOnSeg
				nearest2 = p
			}
		}
	}

	// Segment to segment: find the closest pair of points
	for _, seg1 := range segments1 {
		for _, seg2 := range segments2 {
			p1, p2 := closestPointsOnSegments(seg1.P0, seg1.P1, seg2.P0, seg2.P1)
			dist := p1.Distance(p2)
			if dist < minDist {
				minDist = dist
				nearest1 = p1
				nearest2 = p2
			}
		}
	}

	return nearest1, nearest2
}

// closestPointOnSegmentCoord returns the closest point on segment (a,b) to point p.
func closestPointOnSegmentCoord(p, a, b geom.Coordinate) geom.Coordinate {
	dx := b.X - a.X
	dy := b.Y - a.Y

	if a.Equals2D(b, geom.DefaultEpsilon) {
		return a
	}

	t := ((p.X-a.X)*dx + (p.Y-a.Y)*dy) / (dx*dx + dy*dy)
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}

	return geom.NewCoordinate(a.X+t*dx, a.Y+t*dy)
}

// closestPointsOnSegments returns the closest pair of points between two segments.
func closestPointsOnSegments(a1, a2, b1, b2 geom.Coordinate) (geom.Coordinate, geom.Coordinate) {
	// Check all combinations and return the pair with minimum distance
	minDist := math.Inf(1)
	var best1, best2 geom.Coordinate

	// Endpoints of segment 1 to segment 2
	c := closestPointOnSegmentCoord(a1, b1, b2)
	if d := a1.Distance(c); d < minDist {
		minDist = d
		best1, best2 = a1, c
	}
	c = closestPointOnSegmentCoord(a2, b1, b2)
	if d := a2.Distance(c); d < minDist {
		minDist = d
		best1, best2 = a2, c
	}

	// Endpoints of segment 2 to segment 1
	c = closestPointOnSegmentCoord(b1, a1, a2)
	if d := b1.Distance(c); d < minDist {
		minDist = d
		best1, best2 = c, b1
	}
	c = closestPointOnSegmentCoord(b2, a1, a2)
	if d := b2.Distance(c); d < minDist {
		minDist = d
		best1, best2 = c, b2
	}

	return best1, best2
}
