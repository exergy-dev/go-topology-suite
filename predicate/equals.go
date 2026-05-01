package predicate

import (
	terra "github.com/terra-geo/terra"
	"github.com/terra-geo/terra/crs"
	"github.com/terra-geo/terra/geom"
)

// Equals reports whether a and b describe the same point set.
//
// Fast path: a strict structural comparison (identical type, layout, and
// coordinate buffer) returns true immediately. When that fails — e.g.
// rings start at different vertices, or a Polygon vs MultiPolygon-of-one
// is being compared — Equals falls back to topological equality via the
// DE-9IM matrix (pattern "T*F**FFF*").
//
// Empty inputs of the same type compare equal; an empty geometry is
// considered topologically equal to any other empty geometry of any
// type.
func Equals(a, b geom.Geometry, opts ...Option) (bool, error) {
	if !crs.Equal(a.CRS(), b.CRS()) {
		return false, terra.ErrCRSMismatch
	}
	if a.IsEmpty() && b.IsEmpty() {
		// JTS semantics: any two empty geometries are equal regardless
		// of advertised type.
		return true, nil
	}
	if a.IsEmpty() != b.IsEmpty() {
		return false, nil
	}
	if a.Type() == b.Type() && a.Layout() == b.Layout() && structuralEqual(a, b) {
		return true, nil
	}
	if pointZeroLengthLinePair(a, b) {
		return false, nil
	}
	d, err := Relate(a, b, opts...)
	if err != nil {
		return false, err
	}
	return d.Matches("T*F**FFF*"), nil
}

func pointZeroLengthLinePair(a, b geom.Geometry) bool {
	if _, ok := a.(*geom.Point); ok {
		if ls, ok := b.(*geom.LineString); ok {
			return isZeroLengthLine(ls)
		}
	}
	if _, ok := b.(*geom.Point); ok {
		if ls, ok := a.(*geom.LineString); ok {
			return isZeroLengthLine(ls)
		}
	}
	return false
}

func structuralEqual(a, b geom.Geometry) bool {
	// Recursive paths (GeometryCollection, MultiPolygon child loops)
	// dispatch on a's concrete type without checking that b matches —
	// type assertions below panic if the children disagree. The
	// outer-level Equals() guards the top level; this guards the
	// recursive calls.
	if a.Type() != b.Type() {
		return false
	}
	if a.Layout() != b.Layout() {
		return false
	}
	switch va := a.(type) {
	case *geom.Point:
		return va.XY() == b.(*geom.Point).XY()
	case *geom.LineString:
		return flatEqual(va.FlatCoords(), b.(*geom.LineString).FlatCoords())
	case *geom.Polygon:
		vb := b.(*geom.Polygon)
		if va.NumRings() != vb.NumRings() {
			return false
		}
		for i := 0; i < va.NumRings(); i++ {
			if !ringEqualXY(va.Ring(i), vb.Ring(i)) {
				return false
			}
		}
		return true
	case *geom.MultiPoint:
		return flatEqual(va.FlatCoords(), b.(*geom.MultiPoint).FlatCoords())
	case *geom.MultiLineString:
		vb := b.(*geom.MultiLineString)
		if va.NumGeometries() != vb.NumGeometries() {
			return false
		}
		for i := 0; i < va.NumGeometries(); i++ {
			if !structuralEqual(va.LineStringAt(i), vb.LineStringAt(i)) {
				return false
			}
		}
		return true
	case *geom.MultiPolygon:
		vb := b.(*geom.MultiPolygon)
		if va.NumGeometries() != vb.NumGeometries() {
			return false
		}
		for i := 0; i < va.NumGeometries(); i++ {
			if !structuralEqual(va.PolygonAt(i), vb.PolygonAt(i)) {
				return false
			}
		}
		return true
	case *geom.GeometryCollection:
		vb := b.(*geom.GeometryCollection)
		if va.NumGeometries() != vb.NumGeometries() {
			return false
		}
		for i := 0; i < va.NumGeometries(); i++ {
			if !structuralEqual(va.GeometryAt(i), vb.GeometryAt(i)) {
				return false
			}
		}
		return true
	}
	return false
}

func flatEqual(a, b []float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func ringEqualXY(a, b []geom.XY) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
