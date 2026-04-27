package topology

import "github.com/robert-malhotra/go-topology-suite/geom"

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

// PointLocationInPolygon determines the location of a point in a polygon.
func PointLocationInPolygon(p geom.Coordinate, poly *geom.Polygon) geom.Location {
	if poly == nil || poly.IsEmpty() {
		return geom.LocationExterior
	}

	if geom.PointOnRing(p, poly.ExteriorRing()) {
		return geom.LocationBoundary
	}

	for i := 0; i < poly.NumInteriorRings(); i++ {
		if geom.PointOnRing(p, poly.InteriorRingN(i)) {
			return geom.LocationBoundary
		}
	}

	if !geom.PointInRing(p, poly.ExteriorRing()) {
		return geom.LocationExterior
	}

	for i := 0; i < poly.NumInteriorRings(); i++ {
		if geom.PointInRing(p, poly.InteriorRingN(i)) {
			return geom.LocationExterior
		}
	}

	return geom.LocationInterior
}

func pointLocationInPoint(p geom.Coordinate, pt *geom.Point) geom.Location {
	if pt == nil || pt.IsEmpty() {
		return geom.LocationExterior
	}
	if p.Equals2D(pt.Coordinate(), geom.DefaultEpsilon) {
		return geom.LocationInterior
	}
	return geom.LocationExterior
}

func pointLocationInLineString(p geom.Coordinate, ls *geom.LineString) geom.Location {
	if ls == nil || ls.IsEmpty() {
		return geom.LocationExterior
	}

	coords := ls.Coordinates()
	for i := 1; i < len(coords); i++ {
		if geom.PointOnSegment(p, coords[i-1], coords[i]) {
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
	if ring == nil || ring.IsEmpty() {
		return geom.LocationExterior
	}
	if geom.PointOnRing(p, ring) {
		return geom.LocationBoundary
	}
	if geom.PointInRing(p, ring) {
		return geom.LocationInterior
	}
	return geom.LocationExterior
}

func pointLocationInMultiPoint(p geom.Coordinate, mp *geom.MultiPoint) geom.Location {
	if mp == nil || mp.IsEmpty() {
		return geom.LocationExterior
	}
	for i := 0; i < mp.NumGeometries(); i++ {
		pt := mp.GeometryN(i).(*geom.Point)
		if !pt.IsEmpty() && p.Equals2D(pt.Coordinate(), geom.DefaultEpsilon) {
			return geom.LocationInterior
		}
	}
	return geom.LocationExterior
}

func pointLocationInMultiLineString(p geom.Coordinate, mls *geom.MultiLineString) geom.Location {
	if mls == nil || mls.IsEmpty() {
		return geom.LocationExterior
	}
	onBoundary := false
	for i := 0; i < mls.NumGeometries(); i++ {
		loc := pointLocationInLineString(p, mls.GeometryN(i).(*geom.LineString))
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
	if mp == nil || mp.IsEmpty() {
		return geom.LocationExterior
	}
	polygons := make([]*geom.Polygon, 0, mp.NumGeometries())
	for i := 0; i < mp.NumGeometries(); i++ {
		polygons = append(polygons, mp.GeometryN(i).(*geom.Polygon))
	}
	return PointLocationInPolygonSet(p, polygons)
}

// PointLocationInPolygonSet determines the location of a point relative to a
// polygonal set. Interior wins over boundary, matching OGC union semantics for
// adjacent polygons.
func PointLocationInPolygonSet(p geom.Coordinate, polygons []*geom.Polygon) geom.Location {
	onBoundary := false
	for _, polygon := range polygons {
		loc := PointLocationInPolygon(p, polygon)
		if loc == geom.LocationInterior {
			return geom.LocationInterior
		}
		if loc == geom.LocationBoundary {
			onBoundary = true
		}
	}
	if onBoundary {
		if boundaryPointIsPolygonSetInterior(p, polygons) {
			return geom.LocationInterior
		}
		return geom.LocationBoundary
	}
	return geom.LocationExterior
}

func boundaryPointIsPolygonSetInterior(p geom.Coordinate, polygons []*geom.Polygon) bool {
	offset := geom.DefaultEpsilon * 1000
	if offset == 0 {
		offset = 1e-9
	}
	samples := [][2]geom.Coordinate{
		{
			geom.NewCoordinate(p.X-offset, p.Y),
			geom.NewCoordinate(p.X+offset, p.Y),
		},
		{
			geom.NewCoordinate(p.X, p.Y-offset),
			geom.NewCoordinate(p.X, p.Y+offset),
		},
	}
	for _, pair := range samples {
		if pointInAnyPolygonInterior(pair[0], polygons) && pointInAnyPolygonInterior(pair[1], polygons) {
			return true
		}
	}
	return false
}

func pointInAnyPolygonInterior(p geom.Coordinate, polygons []*geom.Polygon) bool {
	for _, polygon := range polygons {
		if PointLocationInPolygon(p, polygon) == geom.LocationInterior {
			return true
		}
	}
	return false
}

func pointLocationInCollection(p geom.Coordinate, gc *geom.GeometryCollection) geom.Location {
	if gc == nil || gc.IsEmpty() {
		return geom.LocationExterior
	}

	var polygons []*geom.Polygon
	geom.ForEachPolygon(gc, func(poly *geom.Polygon) bool {
		polygons = append(polygons, poly)
		return false
	})
	if len(polygons) > 0 {
		loc := PointLocationInPolygonSet(p, polygons)
		if loc == geom.LocationInterior {
			return geom.LocationInterior
		}
		if loc == geom.LocationBoundary {
			for i := 0; i < gc.NumGeometries(); i++ {
				if _, ok := gc.GeometryN(i).(*geom.Polygon); ok {
					continue
				}
				if PointLocation(p, gc.GeometryN(i)) == geom.LocationInterior {
					return geom.LocationInterior
				}
			}
			return geom.LocationBoundary
		}
	}

	onBoundary := false
	for i := 0; i < gc.NumGeometries(); i++ {
		loc := PointLocation(p, gc.GeometryN(i))
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
