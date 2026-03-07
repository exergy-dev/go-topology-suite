package spherical

import (
	"github.com/robert-malhotra/go-topology-suite/geom"
)

func isNilGeometry(g geom.Geometry) bool {
	if g == nil {
		return true
	}
	switch v := g.(type) {
	case *geom.Point:
		return v == nil
	case *geom.LineString:
		return v == nil
	case *geom.LinearRing:
		return v == nil
	case *geom.Polygon:
		return v == nil
	case *geom.MultiPoint:
		return v == nil
	case *geom.MultiLineString:
		return v == nil
	case *geom.MultiPolygon:
		return v == nil
	case *geom.GeometryCollection:
		return v == nil
	default:
		return false
	}
}

func isEmptyGeometry(g geom.Geometry) bool {
	if isNilGeometry(g) {
		return true
	}
	return g.IsEmpty()
}
