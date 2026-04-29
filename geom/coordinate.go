package geom

import "math"

// XY is a 2D coordinate. It is a value type with no methods that allocate.
//
// Per the design memo: never use XY (or its siblings) as a Go map key when
// any field can be NaN — math.NaN() != math.NaN() will silently break lookups.
// Use a normalized integer-grid key for that purpose.
type XY struct {
	X, Y float64
}

// XYZ is a 3D coordinate.
type XYZ struct {
	X, Y, Z float64
}

// XYM is a 2D coordinate with a linear-referencing measure.
type XYM struct {
	X, Y, M float64
}

// XYZM is a 3D coordinate with a linear-referencing measure.
type XYZM struct {
	X, Y, Z, M float64
}

// Coord is the union of all four coordinate value types. It is the type
// parameter constraint used by generic coordinate-transform helpers like
// geom.Apply.
type Coord interface {
	XY | XYZ | XYM | XYZM
}

// AsXY drops Z/M values and returns the 2D projection of c.
func (c XYZ) AsXY() XY  { return XY{c.X, c.Y} }
func (c XYM) AsXY() XY  { return XY{c.X, c.Y} }
func (c XYZM) AsXY() XY { return XY{c.X, c.Y} }

// Equal compares two XY values exactly. It does not handle NaN; for that
// use EqualOrBothNaN.
func (a XY) Equal(b XY) bool { return a.X == b.X && a.Y == b.Y }

// EqualOrBothNaN compares two XY values treating NaN==NaN as true.
// Use this when XY values may originate from missing-data markers.
func (a XY) EqualOrBothNaN(b XY) bool {
	return (a.X == b.X || (math.IsNaN(a.X) && math.IsNaN(b.X))) &&
		(a.Y == b.Y || (math.IsNaN(a.Y) && math.IsNaN(b.Y)))
}
