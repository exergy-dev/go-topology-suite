package geom

import "math"

// Segment represents a line segment defined by two coordinates.
type Segment struct {
	P0, P1 Coordinate
}

// SegmentsIntersect returns true if two line segments intersect.
func SegmentsIntersect(a1, a2, b1, b2 Coordinate) bool {
	o1 := OrientationIndex(a1, a2, b1)
	o2 := OrientationIndex(a1, a2, b2)
	o3 := OrientationIndex(b1, b2, a1)
	o4 := OrientationIndex(b1, b2, a2)

	// General case
	if o1 != o2 && o3 != o4 {
		return true
	}

	// Collinear cases
	if o1 == 0 && onSegmentBounds(a1, b1, a2) {
		return true
	}
	if o2 == 0 && onSegmentBounds(a1, b2, a2) {
		return true
	}
	if o3 == 0 && onSegmentBounds(b1, a1, b2) {
		return true
	}
	if o4 == 0 && onSegmentBounds(b1, a2, b2) {
		return true
	}

	return false
}

// GeometrySegments returns all line segments from a geometry.
// For polygons, includes shell and hole segments. Linear rings are included.
func GeometrySegments(g Geometry) []Segment {
	var segments []Segment
	appendGeometrySegments(&segments, g, true)
	return segments
}

// BoundarySegments returns all line segments for geometry boundaries.
// Linear rings are excluded to match their empty boundary semantics.
func BoundarySegments(g Geometry) []Segment {
	var segments []Segment
	appendGeometrySegments(&segments, g, false)
	return segments
}

func appendGeometrySegments(segments *[]Segment, g Geometry, includeLinearRing bool) {
	switch v := g.(type) {
	case *LineString:
		*segments = appendSegmentsFromCoords(*segments, v.Coordinates())
	case *LinearRing:
		if includeLinearRing {
			*segments = appendSegmentsFromCoords(*segments, v.Coordinates())
		}
	case *Polygon:
		*segments = appendSegmentsFromCoords(*segments, v.shell.Coordinates())
		for _, hole := range v.holes {
			*segments = appendSegmentsFromCoords(*segments, hole.Coordinates())
		}
	case *MultiLineString:
		for i := 0; i < v.NumGeometries(); i++ {
			appendGeometrySegments(segments, v.GeometryN(i), includeLinearRing)
		}
	case *MultiPolygon:
		for i := 0; i < v.NumGeometries(); i++ {
			appendGeometrySegments(segments, v.GeometryN(i), includeLinearRing)
		}
	case *GeometryCollection:
		for i := 0; i < v.NumGeometries(); i++ {
			appendGeometrySegments(segments, v.GeometryN(i), includeLinearRing)
		}
	}
}

func appendSegmentsFromCoords(segments []Segment, coords CoordinateSequence) []Segment {
	for i := 1; i < len(coords); i++ {
		segments = append(segments, Segment{P0: coords[i-1], P1: coords[i]})
	}
	return segments
}

func onSegmentBounds(p, q, r Coordinate) bool {
	return q.X <= math.Max(p.X, r.X) && q.X >= math.Min(p.X, r.X) &&
		q.Y <= math.Max(p.Y, r.Y) && q.Y >= math.Min(p.Y, r.Y)
}
