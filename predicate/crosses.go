package predicate

import (
	terra "github.com/terra-geo/terra"
	"github.com/terra-geo/terra/crs"
	"github.com/terra-geo/terra/geom"
)

// Crosses reports whether the geometries' interiors share at least one
// point but the dimension of their intersection is strictly less than
// max(dim(a), dim(b)).
//
// Per OGC, the matrix patterns depend on the dimensions of a and b:
//
//   - dim(a) < dim(b):     T*T******
//   - dim(a) = dim(b) = 1: 0********
//   - dim(a) > dim(b):     T*****T**
//
// Same-dim area-area is undefined and returns false.
func Crosses(a, b geom.Geometry, opts ...Option) (bool, error) {
	if !crs.Equal(a.CRS(), b.CRS()) {
		return false, terra.ErrCRSMismatch
	}
	a = unwrapLinearRing(a)
	b = unwrapLinearRing(b)
	if a.IsEmpty() || b.IsEmpty() {
		return false, nil
	}
	c := resolve(a, opts)
	if c.kernel.Name() == "planar" && !a.Envelope().Intersects(b.Envelope()) {
		return false, nil
	}
	dA := dimensionOf(a)
	dB := dimensionOf(b)

	d, err := Relate(a, b, opts...)
	if err != nil {
		return false, err
	}
	switch {
	case dA == 1 && dB == 1:
		return d.Matches("0********"), nil
	case dA < dB:
		return d.Matches("T*T******"), nil
	case dA > dB:
		return d.Matches("T*****T**"), nil
	}
	return false, nil
}
