package predicate

import (
	terra "github.com/terra-geo/terra"
	"github.com/terra-geo/terra/crs"
	"github.com/terra-geo/terra/geom"
)

// Touches reports whether a and b share at least one boundary point but
// have no interior points in common.
//
// Defined for all type pairs except (Point, Point) (which always returns
// false: points have no boundary). Derived from the DE-9IM matrix per
// OGC: II=F AND any of {IB, BI, BB} is non-F.
func Touches(a, b geom.Geometry, opts ...Option) (bool, error) {
	if !crs.Equal(a.CRS(), b.CRS()) {
		return false, terra.ErrCRSMismatch
	}
	if a.IsEmpty() || b.IsEmpty() {
		return false, nil
	}
	// Two pure points cannot Touches (no boundaries).
	if dimensionOf(a) == 0 && dimensionOf(b) == 0 && !isMulti(a) && !isMulti(b) {
		return false, nil
	}
	c := resolve(a, opts)
	if c.kernel.Name() == "planar" && !a.Envelope().Intersects(b.Envelope()) {
		return false, nil
	}
	d, err := Relate(a, b, opts...)
	if err != nil {
		return false, err
	}
	return d.Matches("FT*******") || d.Matches("F**T*****") || d.Matches("F***T****"), nil
}
