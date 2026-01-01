package transform

import (
	"fmt"

	"github.com/robert-malhotra/go-topology-suite/geom"
)

// TransformGeometry applies a transformation to all coordinates in a geometry.
// Returns a new transformed geometry with the same type and structure as the input.
// Handles all geometry types including collections.
func TransformGeometry(t Transform, g geom.Geometry) (geom.Geometry, error) {
	if g == nil || g.IsEmpty() {
		return g, nil
	}

	switch geomType := g.(type) {
	case *geom.Point:
		return transformPoint(t, geomType)

	case *geom.LineString:
		return transformLineString(t, geomType)

	case *geom.LinearRing:
		return transformLinearRing(t, geomType)

	case *geom.Polygon:
		return transformPolygon(t, geomType)

	case *geom.MultiPoint:
		return transformMultiPoint(t, geomType)

	case *geom.MultiLineString:
		return transformMultiLineString(t, geomType)

	case *geom.MultiPolygon:
		return transformMultiPolygon(t, geomType)

	case *geom.GeometryCollection:
		return transformGeometryCollection(t, geomType)

	default:
		return nil, fmt.Errorf("unsupported geometry type: %T", g)
	}
}

// transformPoint transforms a Point geometry.
func transformPoint(t Transform, p *geom.Point) (*geom.Point, error) {
	if p.IsEmpty() {
		result := geom.NewPointEmpty()
		result.SetSRID(p.SRID())
		return result, nil
	}

	coord, err := TransformCoordinate(t, p.Coordinates()[0])
	if err != nil {
		return nil, fmt.Errorf("transforming point: %w", err)
	}

	result := geom.NewPointFromCoordinate(coord)
	result.SetSRID(p.SRID())
	return result, nil
}

// transformLineString transforms a LineString geometry.
func transformLineString(t Transform, ls *geom.LineString) (*geom.LineString, error) {
	if ls.IsEmpty() {
		result := geom.NewLineStringEmpty()
		result.SetSRID(ls.SRID())
		return result, nil
	}

	coords, err := TransformCoordinates(t, ls.Coordinates())
	if err != nil {
		return nil, fmt.Errorf("transforming linestring: %w", err)
	}

	result := geom.NewLineString(coords)
	result.SetSRID(ls.SRID())
	return result, nil
}

// transformLinearRing transforms a LinearRing geometry.
func transformLinearRing(t Transform, lr *geom.LinearRing) (*geom.LinearRing, error) {
	if lr.IsEmpty() {
		result := geom.NewLinearRingEmpty()
		result.SetSRID(lr.SRID())
		return result, nil
	}

	coords, err := TransformCoordinates(t, lr.Coordinates())
	if err != nil {
		return nil, fmt.Errorf("transforming linear ring: %w", err)
	}

	result := geom.NewLinearRing(coords)
	result.SetSRID(lr.SRID())
	return result, nil
}

// transformPolygon transforms a Polygon geometry including its shell and holes.
func transformPolygon(t Transform, poly *geom.Polygon) (*geom.Polygon, error) {
	if poly.IsEmpty() {
		result := geom.NewPolygonEmpty()
		result.SetSRID(poly.SRID())
		return result, nil
	}

	// Transform exterior ring
	shell, err := transformLinearRing(t, poly.ExteriorRing())
	if err != nil {
		return nil, fmt.Errorf("transforming polygon shell: %w", err)
	}

	// Transform holes
	holes := make([]*geom.LinearRing, poly.NumInteriorRings())
	for i := 0; i < poly.NumInteriorRings(); i++ {
		hole, err := transformLinearRing(t, poly.InteriorRingN(i))
		if err != nil {
			return nil, fmt.Errorf("transforming polygon hole %d: %w", i, err)
		}
		holes[i] = hole
	}

	result := geom.NewPolygon(shell, holes)
	result.SetSRID(poly.SRID())
	return result, nil
}

// transformMultiPoint transforms a MultiPoint geometry.
func transformMultiPoint(t Transform, mp *geom.MultiPoint) (*geom.MultiPoint, error) {
	if mp.IsEmpty() {
		result := geom.NewMultiPointEmpty()
		result.SetSRID(mp.SRID())
		return result, nil
	}

	points := make([]*geom.Point, mp.NumGeometries())
	for i := 0; i < mp.NumGeometries(); i++ {
		point, err := transformPoint(t, mp.GeometryN(i).(*geom.Point))
		if err != nil {
			return nil, fmt.Errorf("transforming multipoint component %d: %w", i, err)
		}
		points[i] = point
	}

	result := geom.NewMultiPoint(points)
	result.SetSRID(mp.SRID())
	return result, nil
}

// transformMultiLineString transforms a MultiLineString geometry.
func transformMultiLineString(t Transform, mls *geom.MultiLineString) (*geom.MultiLineString, error) {
	if mls.IsEmpty() {
		result := geom.NewMultiLineStringEmpty()
		result.SetSRID(mls.SRID())
		return result, nil
	}

	linestrings := make([]*geom.LineString, mls.NumGeometries())
	for i := 0; i < mls.NumGeometries(); i++ {
		ls, err := transformLineString(t, mls.GeometryN(i).(*geom.LineString))
		if err != nil {
			return nil, fmt.Errorf("transforming multilinestring component %d: %w", i, err)
		}
		linestrings[i] = ls
	}

	result := geom.NewMultiLineString(linestrings)
	result.SetSRID(mls.SRID())
	return result, nil
}

// transformMultiPolygon transforms a MultiPolygon geometry.
func transformMultiPolygon(t Transform, mpoly *geom.MultiPolygon) (*geom.MultiPolygon, error) {
	if mpoly.IsEmpty() {
		result := geom.NewMultiPolygonEmpty()
		result.SetSRID(mpoly.SRID())
		return result, nil
	}

	polygons := make([]*geom.Polygon, mpoly.NumGeometries())
	for i := 0; i < mpoly.NumGeometries(); i++ {
		poly, err := transformPolygon(t, mpoly.GeometryN(i).(*geom.Polygon))
		if err != nil {
			return nil, fmt.Errorf("transforming multipolygon component %d: %w", i, err)
		}
		polygons[i] = poly
	}

	result := geom.NewMultiPolygon(polygons)
	result.SetSRID(mpoly.SRID())
	return result, nil
}

// transformGeometryCollection transforms a GeometryCollection recursively.
func transformGeometryCollection(t Transform, gc *geom.GeometryCollection) (*geom.GeometryCollection, error) {
	if gc.IsEmpty() {
		result := geom.NewGeometryCollectionEmpty()
		result.SetSRID(gc.SRID())
		return result, nil
	}

	geometries := make([]geom.Geometry, gc.NumGeometries())
	for i := 0; i < gc.NumGeometries(); i++ {
		geom, err := TransformGeometry(t, gc.GeometryN(i))
		if err != nil {
			return nil, fmt.Errorf("transforming collection component %d: %w", i, err)
		}
		geometries[i] = geom
	}

	result := geom.NewGeometryCollection(geometries)
	result.SetSRID(gc.SRID())
	return result, nil
}
