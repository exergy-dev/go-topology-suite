package geom

// GeometryVisitor provides type-specific callbacks for geometry dispatch.
type GeometryVisitor struct {
	Point              func(*Point)
	LineString         func(*LineString)
	LinearRing         func(*LinearRing)
	Polygon            func(*Polygon)
	MultiPoint         func(*MultiPoint)
	MultiLineString    func(*MultiLineString)
	MultiPolygon       func(*MultiPolygon)
	GeometryCollection func(*GeometryCollection)
	Default            func(Geometry)
}

// VisitGeometry dispatches to the first matching visitor callback.
func VisitGeometry(g Geometry, visitor GeometryVisitor) {
	switch v := g.(type) {
	case *Point:
		if visitor.Point != nil {
			visitor.Point(v)
			return
		}
	case *LineString:
		if visitor.LineString != nil {
			visitor.LineString(v)
			return
		}
	case *LinearRing:
		if visitor.LinearRing != nil {
			visitor.LinearRing(v)
			return
		}
	case *Polygon:
		if visitor.Polygon != nil {
			visitor.Polygon(v)
			return
		}
	case *MultiPoint:
		if visitor.MultiPoint != nil {
			visitor.MultiPoint(v)
			return
		}
	case *MultiLineString:
		if visitor.MultiLineString != nil {
			visitor.MultiLineString(v)
			return
		}
	case *MultiPolygon:
		if visitor.MultiPolygon != nil {
			visitor.MultiPolygon(v)
			return
		}
	case *GeometryCollection:
		if visitor.GeometryCollection != nil {
			visitor.GeometryCollection(v)
			return
		}
	}
	if visitor.Default != nil {
		visitor.Default(g)
	}
}
