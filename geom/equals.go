package geom

import "slices"

// EqualsExact reports whether a and b are structurally identical: the same
// concrete type, the same Layout, and the same vertices in the same order
// with the same structure (ring boundaries for a Polygon, member sequence
// for a collection). It is the exact, tolerance-free counterpart of JTS
// equalsExact and is NOT a topological equality test — use predicate.Equals
// for that. A LineString and a LinearRing with identical coordinates are
// not equal, nor are reordered vertices or the same coordinates split into
// different rings.
//
// Every ordinate at the geometry's full stride is compared, so Z and M
// participate. Ordinates compare with ==, except that NaN equals NaN
// (consistent with XY.Equal). Two empty geometries of the same type and
// layout are equal. The CRS is deliberately NOT compared, matching JTS,
// which ignores the SRID; check CRS separately if it matters.
//
// Nil handling: two nil geometries are equal and a nil geometry never
// equals a non-nil one. A typed-nil pointer (e.g. (*Point)(nil)) is
// treated the same as an interface nil rather than panicking.
//
// No coordinates are copied.
//
// JTS: Geometry.equalsExact(Geometry).
func EqualsExact(a, b Geometry) bool {
	aNil, bNil := isNilGeometry(a), isNilGeometry(b)
	if aNil || bNil {
		return aNil && bNil
	}
	if a.Type() != b.Type() || a.Layout() != b.Layout() {
		return false
	}
	aEmpty, bEmpty := a.IsEmpty(), b.IsEmpty()
	if aEmpty || bEmpty {
		return aEmpty && bEmpty
	}
	switch av := a.(type) {
	case *Point:
		bv, ok := b.(*Point)
		return ok && equalFlat(av.coords, bv.coords)
	case *LineString:
		bv, ok := b.(*LineString)
		return ok && equalFlat(av.coords, bv.coords)
	case *LinearRing:
		bv, ok := b.(*LinearRing)
		return ok && equalFlat(av.coords, bv.coords)
	case *MultiPoint:
		bv, ok := b.(*MultiPoint)
		return ok && equalFlat(av.coords, bv.coords)
	case *Polygon:
		bv, ok := b.(*Polygon)
		return ok && slices.Equal(av.ringStarts, bv.ringStarts) &&
			equalFlat(av.coords, bv.coords)
	case *MultiLineString:
		bv, ok := b.(*MultiLineString)
		return ok && equalParts(av.parts, bv.parts)
	case *MultiPolygon:
		bv, ok := b.(*MultiPolygon)
		return ok && equalParts(av.parts, bv.parts)
	case *GeometryCollection:
		bv, ok := b.(*GeometryCollection)
		return ok && equalParts(av.parts, bv.parts)
	}
	return false
}

// isNilGeometry reports whether g is an interface nil or a typed-nil
// pointer to one of the concrete geometry types.
func isNilGeometry(g Geometry) bool {
	switch v := g.(type) {
	case nil:
		return true
	case *Point:
		return v == nil
	case *LineString:
		return v == nil
	case *LinearRing:
		return v == nil
	case *Polygon:
		return v == nil
	case *MultiPoint:
		return v == nil
	case *MultiLineString:
		return v == nil
	case *MultiPolygon:
		return v == nil
	case *GeometryCollection:
		return v == nil
	}
	return false
}

// equalFlat compares two flat coordinate buffers element-wise, treating
// NaN as equal to NaN.
func equalFlat(a, b []float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !equalOrNaN(a[i], b[i]) {
			return false
		}
	}
	return true
}

func equalParts[T Geometry](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !EqualsExact(a[i], b[i]) {
			return false
		}
	}
	return true
}
