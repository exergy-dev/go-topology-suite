package predicate

import (
	terra "github.com/terra-geo/terra"
	"github.com/terra-geo/terra/crs"
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/kernel"
)

// Covers reports whether every point of b lies in the closure of a — that
// is, in a's interior OR on a's boundary. This differs from Contains
// only at the boundary: a square covers a vertex on its edge, but does
// not Contain it.
//
// Derived from the DE-9IM matrix per OGC: Covers ⟺ Relate matches any
// of "T*****FF*", "*T****FF*", "***T**FF*", or "****T*FF*".
func Covers(a, b geom.Geometry, opts ...Option) (bool, error) {
	if !crs.Equal(a.CRS(), b.CRS()) {
		return false, terra.ErrCRSMismatch
	}
	a = unwrapLinearRing(a)
	b = unwrapLinearRing(b)
	c := resolve(a, opts)
	// RelateNG short-circuit: empty/empty, dim(b) > dim(a), envelope
	// non-coverage all resolve here without building a topology graph.
	if sc := scCovers(a, b, c.kernel.Name() == "planar"); sc.resolved {
		return sc.get(), nil
	}
	// Prepared fast-path. PreparedPolygon implements Covers(g) directly;
	// fall back to the per-point ContainsPoint loop for the minimal
	// handle interface (works for Point/MultiPoint inputs).
	if c.prepared != nil {
		if pc, ok := c.prepared.(preparedCoverer); ok {
			return pc.Covers(b), nil
		}
		switch vb := b.(type) {
		case *geom.Point:
			return c.prepared.ContainsPoint(vb.XY()) != kernel.Outside, nil
		case *geom.MultiPoint:
			for i := 0; i < vb.NumGeometries(); i++ {
				if c.prepared.ContainsPoint(vb.PointAt(i)) == kernel.Outside {
					return false, nil
				}
			}
			return vb.NumGeometries() > 0, nil
		}
	}
	if ok, handled := coversFastPath(a, b, c.kernel); handled {
		return ok, nil
	}
	d, err := Relate(a, b, opts...)
	if err != nil {
		return false, err
	}
	return d.Matches("T*****FF*") ||
		d.Matches("*T****FF*") ||
		d.Matches("***T**FF*") ||
		d.Matches("****T*FF*"), nil
}

// CoveredBy is Covers with operands swapped.
func CoveredBy(a, b geom.Geometry, opts ...Option) (bool, error) {
	return Covers(b, a, opts...)
}

func coversFastPath(a, b geom.Geometry, k kernel.Kernel) (bool, bool) {
	switch va := a.(type) {
	case *geom.Point:
		switch vb := b.(type) {
		case *geom.Point:
			return va.XY() == vb.XY(), true
		case *geom.LineString:
			for i := 0; i < vb.NumPoints(); i++ {
				if vb.PointAt(i) != va.XY() {
					return false, true
				}
			}
			return vb.NumPoints() > 0, true
		}
	case *geom.LineString:
		switch vb := b.(type) {
		case *geom.Point:
			return pointOnLine(vb.XY(), va, k), true
		case *geom.LineString:
			return lineFullyOn(vb, va, k), true
		}
	case *geom.MultiLineString:
		if p, ok := b.(*geom.Point); ok {
			for i := 0; i < va.NumGeometries(); i++ {
				if pointOnLine(p.XY(), va.LineStringAt(i), k) {
					return true, true
				}
			}
			return false, true
		}
	case *geom.Polygon:
		switch b.(type) {
		case *geom.GeometryCollection, *geom.MultiPoint, *geom.MultiLineString:
			covered, _ := polygonCoversWithInteriorHit(va, b, k)
			return covered, true
		}
	case *geom.MultiPolygon:
		switch b.(type) {
		case *geom.Point, *geom.MultiPoint, *geom.LineString, *geom.MultiLineString, *geom.GeometryCollection:
			covered, _ := collectionCoversWithInteriorHit(multiPolygonAsCollection(va), b, k)
			return covered, true
		}
	case *geom.GeometryCollection:
		covered, _ := collectionCoversWithInteriorHit(va, b, k)
		return covered, true
	}
	return false, false
}
