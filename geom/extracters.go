// Port of org.locationtech.jts.geom.util.PointExtracter,
// LineStringExtracter and PolygonExtracter.
//
// Walks an arbitrary Geometry tree (recursing through GeometryCollection
// and its multi-geometry subclasses) and collects all components of a
// requested type.

package geom

// PointsOf returns every Point component of g, traversing into
// GeometryCollection / MultiPoint members. The returned slice is freshly
// allocated and may be empty.
//
// Mirrors org.locationtech.jts.geom.util.PointExtracter.getPoints.
func PointsOf(g Geometry) []*Point {
	var out []*Point
	walkComponents(g, func(c Geometry) {
		if pt, ok := c.(*Point); ok {
			out = append(out, pt)
		}
	})
	return out
}

// LineStringsOf returns every LineString component of g. LinearRings are
// included as LineStrings (the JTS extracter includes them since LinearRing
// extends LineString).
//
// Mirrors org.locationtech.jts.geom.util.LineStringExtracter.getLineStrings.
func LineStringsOf(g Geometry) []*LineString {
	var out []*LineString
	walkComponents(g, func(c Geometry) {
		switch v := c.(type) {
		case *LineString:
			out = append(out, v)
		case *LinearRing:
			out = append(out, v.AsLineString())
		}
	})
	return out
}

// PolygonsOf returns every Polygon component of g, recursing into
// MultiPolygon and GeometryCollection.
//
// Mirrors org.locationtech.jts.geom.util.PolygonExtracter.getPolygons.
func PolygonsOf(g Geometry) []*Polygon {
	var out []*Polygon
	walkComponents(g, func(c Geometry) {
		if p, ok := c.(*Polygon); ok {
			out = append(out, p)
		}
	})
	return out
}

// walkComponents visits every leaf-level component of g, descending into
// GeometryCollection/MultiPoint/MultiLineString/MultiPolygon. The visitor
// is invoked on the leaf types (Point, LineString, LinearRing, Polygon).
//
// MultiPoint is decomposed into its constituent Points; MultiLineString
// into LineStrings; MultiPolygon into Polygons. This matches the JTS
// GeometryFilter.apply contract, where the filter receives every component
// of the tree.
func walkComponents(g Geometry, visit func(Geometry)) {
	if g == nil {
		return
	}
	switch v := g.(type) {
	case *Point, *LineString, *LinearRing, *Polygon:
		visit(v)
	case *MultiPoint:
		// Decompose into individual Points so callers receive the same
		// component leaves as JTS GeometryFilter would.
		for i := 0; i < v.NumGeometries(); i++ {
			xy := v.PointAt(i)
			pt := NewPoint(v.crs, xy)
			visit(pt)
		}
	case *MultiLineString:
		for _, ls := range v.parts {
			visit(ls)
		}
	case *MultiPolygon:
		for _, p := range v.parts {
			visit(p)
		}
	case *GeometryCollection:
		for _, child := range v.parts {
			walkComponents(child, visit)
		}
	}
}
