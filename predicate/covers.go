package predicate

import (
	terra "github.com/terra-geo/terra"
	"github.com/terra-geo/terra/crs"
	"github.com/terra-geo/terra/geom"
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
	if a.IsEmpty() || b.IsEmpty() {
		return false, nil
	}
	c := resolve(a, opts)
	if c.kernel.Name() == "planar" && !a.Envelope().Contains(b.Envelope()) {
		return false, nil
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
