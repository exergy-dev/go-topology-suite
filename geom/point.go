package geom

import (
	"math"

	"github.com/terra-geo/terra/crs"
)

func nan() float64 { return math.NaN() }

// Point is a single coordinate.
type Point struct {
	baseGeom
}

// NewPoint constructs a Point from an XY coordinate.
// The CRS may be nil; callers donate the value semantics — Point holds no
// reference to caller-owned slices.
func NewPoint(c *crs.CRS, p XY) *Point {
	return &Point{
		baseGeom: baseGeom{
			layout: LayoutXY,
			coords: []float64{p.X, p.Y},
			crs:    c,
		},
	}
}

// NewPointXYZ constructs a 3D Point.
func NewPointXYZ(c *crs.CRS, p XYZ) *Point {
	return &Point{
		baseGeom: baseGeom{
			layout: LayoutXYZ,
			coords: []float64{p.X, p.Y, p.Z},
			crs:    c,
		},
	}
}

// NewPointXYM constructs a 2D+M Point.
func NewPointXYM(c *crs.CRS, p XYM) *Point {
	return &Point{
		baseGeom: baseGeom{
			layout: LayoutXYM,
			coords: []float64{p.X, p.Y, p.M},
			crs:    c,
		},
	}
}

// NewPointXYZM constructs a 3D+M Point.
func NewPointXYZM(c *crs.CRS, p XYZM) *Point {
	return &Point{
		baseGeom: baseGeom{
			layout: LayoutXYZM,
			coords: []float64{p.X, p.Y, p.Z, p.M},
			crs:    c,
		},
	}
}

// NewEmptyPoint constructs a POINT EMPTY in the given layout.
func NewEmptyPoint(c *crs.CRS, layout Layout) *Point {
	return &Point{baseGeom: baseGeom{layout: layout, crs: c}}
}

// Z returns the Z value if the layout has one, otherwise NaN.
func (p *Point) Z() float64 {
	if p.IsEmpty() || !p.layout.HasZ() {
		return nan()
	}
	return p.coords[2]
}

// M returns the M value if the layout has one, otherwise NaN. M is at
// index 2 for XYM and index 3 for XYZM.
func (p *Point) M() float64 {
	if p.IsEmpty() || !p.layout.HasM() {
		return nan()
	}
	switch p.layout {
	case LayoutXYM:
		return p.coords[2]
	case LayoutXYZM:
		return p.coords[3]
	}
	return nan()
}

func (p *Point) isGeometry()      {}
func (p *Point) Type() Type       { return PointType }
func (p *Point) Envelope() Envelope { return p.envelope() }
func (p *Point) IsEmpty() bool      { return len(p.coords) == 0 }
func (p *Point) NumGeometries() int { return 1 }

// XY returns the 2D projection of the point.
// Returns the zero XY for an empty point.
func (p *Point) XY() XY {
	if p.IsEmpty() {
		return XY{}
	}
	return XY{p.coords[0], p.coords[1]}
}
