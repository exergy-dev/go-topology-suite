package algorithm

import (
	"github.com/robert-malhotra/go-topology-suite/geom"
)

// PointLocation determines the location of a point relative to a geometry.
func PointLocation(p geom.Coordinate, g geom.Geometry) geom.Location {
	switch v := g.(type) {
	case *geom.Point:
		return pointLocationInPoint(p, v)
	case *geom.LineString:
		return pointLocationInLineString(p, v)
	case *geom.LinearRing:
		return pointLocationInRing(p, v)
	case *geom.Polygon:
		return PointLocationInPolygon(p, v)
	case *geom.MultiPoint:
		return pointLocationInMultiPoint(p, v)
	case *geom.MultiLineString:
		return pointLocationInMultiLineString(p, v)
	case *geom.MultiPolygon:
		return pointLocationInMultiPolygon(p, v)
	case *geom.GeometryCollection:
		return pointLocationInCollection(p, v)
	default:
		return geom.LocationExterior
	}
}

func pointLocationInPoint(p geom.Coordinate, pt *geom.Point) geom.Location {
	if pt.IsEmpty() {
		return geom.LocationExterior
	}
	if p.Equals2D(pt.Coordinate(), geom.DefaultEpsilon) {
		return geom.LocationInterior
	}
	return geom.LocationExterior
}

func pointLocationInLineString(p geom.Coordinate, ls *geom.LineString) geom.Location {
	if ls.IsEmpty() {
		return geom.LocationExterior
	}

	coords := ls.Coordinates()

	// Check if on any segment
	for i := 1; i < len(coords); i++ {
		if geom.PointOnSegment(p, coords[i-1], coords[i]) {
			// Check if at endpoint
			if !ls.IsClosed() && (p.Equals2D(coords[0], geom.DefaultEpsilon) ||
				p.Equals2D(coords[len(coords)-1], geom.DefaultEpsilon)) {
				return geom.LocationBoundary
			}
			return geom.LocationInterior
		}
	}

	return geom.LocationExterior
}

func pointLocationInRing(p geom.Coordinate, ring *geom.LinearRing) geom.Location {
	if ring.IsEmpty() {
		return geom.LocationExterior
	}

	// Check if on boundary
	if geom.PointOnRing(p, ring) {
		return geom.LocationBoundary
	}

	// Check if inside
	if IsPointInRing(p, ring) {
		return geom.LocationInterior
	}

	return geom.LocationExterior
}

// PointLocationInPolygon determines the location of a point in a polygon.
func PointLocationInPolygon(p geom.Coordinate, poly *geom.Polygon) geom.Location {
	if poly.IsEmpty() {
		return geom.LocationExterior
	}

	// Check if on shell boundary
	if geom.PointOnRing(p, poly.ExteriorRing()) {
		return geom.LocationBoundary
	}

	// Check if on hole boundaries
	for i := 0; i < poly.NumInteriorRings(); i++ {
		if geom.PointOnRing(p, poly.InteriorRingN(i)) {
			return geom.LocationBoundary
		}
	}

	// Check if inside shell
	if !IsPointInRing(p, poly.ExteriorRing()) {
		return geom.LocationExterior
	}

	// Check if inside any hole
	for i := 0; i < poly.NumInteriorRings(); i++ {
		if IsPointInRing(p, poly.InteriorRingN(i)) {
			return geom.LocationExterior
		}
	}

	return geom.LocationInterior
}

func pointLocationInMultiPoint(p geom.Coordinate, mp *geom.MultiPoint) geom.Location {
	for i := 0; i < mp.NumGeometries(); i++ {
		pt := mp.GeometryN(i).(*geom.Point)
		if p.Equals2D(pt.Coordinate(), geom.DefaultEpsilon) {
			return geom.LocationInterior
		}
	}
	return geom.LocationExterior
}

func pointLocationInMultiLineString(p geom.Coordinate, mls *geom.MultiLineString) geom.Location {
	onBoundary := false
	for i := 0; i < mls.NumGeometries(); i++ {
		ls := mls.GeometryN(i).(*geom.LineString)
		loc := pointLocationInLineString(p, ls)
		if loc == geom.LocationInterior {
			return geom.LocationInterior
		}
		if loc == geom.LocationBoundary {
			onBoundary = true
		}
	}
	if onBoundary {
		return geom.LocationBoundary
	}
	return geom.LocationExterior
}

func pointLocationInMultiPolygon(p geom.Coordinate, mp *geom.MultiPolygon) geom.Location {
	for i := 0; i < mp.NumGeometries(); i++ {
		poly := mp.GeometryN(i).(*geom.Polygon)
		loc := PointLocationInPolygon(p, poly)
		if loc == geom.LocationInterior {
			return geom.LocationInterior
		}
		if loc == geom.LocationBoundary {
			return geom.LocationBoundary
		}
	}
	return geom.LocationExterior
}

func pointLocationInCollection(p geom.Coordinate, gc *geom.GeometryCollection) geom.Location {
	for i := 0; i < gc.NumGeometries(); i++ {
		loc := PointLocation(p, gc.GeometryN(i))
		if loc == geom.LocationInterior {
			return geom.LocationInterior
		}
		if loc == geom.LocationBoundary {
			return geom.LocationBoundary
		}
	}
	return geom.LocationExterior
}

// IsPointInRing determines if a point is inside a ring using the ray casting algorithm.
func IsPointInRing(p geom.Coordinate, ring *geom.LinearRing) bool {
	return geom.PointInRing(p, ring)
}

// IsPointInEnvelope returns true if a point is within an envelope.
//
// Deprecated: Use env.Contains(p) instead.
func IsPointInEnvelope(p geom.Coordinate, env *geom.Envelope) bool {
	return env.Contains(p)
}

// LocatePointInTriangle determines the location of a point relative to a triangle.
func LocatePointInTriangle(p, t0, t1, t2 geom.Coordinate) geom.Location {
	// Check if on any edge
	if geom.PointOnSegment(p, t0, t1) || geom.PointOnSegment(p, t1, t2) || geom.PointOnSegment(p, t2, t0) {
		return geom.LocationBoundary
	}

	// Use barycentric coordinates
	if isPointInTriangle(p, t0, t1, t2) {
		return geom.LocationInterior
	}

	return geom.LocationExterior
}

func isPointInTriangle(p, t0, t1, t2 geom.Coordinate) bool {
	o1 := OrientationIndex(t0, t1, p)
	o2 := OrientationIndex(t1, t2, p)
	o3 := OrientationIndex(t2, t0, p)

	// All same orientation means inside
	return (o1 >= 0 && o2 >= 0 && o3 >= 0) || (o1 <= 0 && o2 <= 0 && o3 <= 0)
}

// IndexOfPointInRing returns the index of a point in a ring, or -1 if not found.
func IndexOfPointInRing(p geom.Coordinate, ring *geom.LinearRing) int {
	coords := ring.Coordinates()
	for i, c := range coords {
		if p.Equals2D(c, geom.DefaultEpsilon) {
			return i
		}
	}
	return -1
}

// IndexOfClosestPointInSequence returns the index of the closest coordinate to a point.
func IndexOfClosestPointInSequence(p geom.Coordinate, coords geom.CoordinateSequence) int {
	if len(coords) == 0 {
		return -1
	}

	minDist := p.Distance(coords[0])
	minIdx := 0

	for i := 1; i < len(coords); i++ {
		dist := p.Distance(coords[i])
		if dist < minDist {
			minDist = dist
			minIdx = i
		}
	}

	return minIdx
}
