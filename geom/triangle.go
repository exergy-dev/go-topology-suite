package geom

import "math"

// Triangle is a planar triangle defined by three vertices, providing a
// collection of geometric properties (area, centroid, circumcentre, etc.).
//
// Ported from JTS org.locationtech.jts.geom.Triangle (Vivid Solutions, EPL).
// The vertex Z values participate only in Area3D and InterpolateZ; all other
// helpers operate on the XY projection.
type Triangle struct {
	P0, P1, P2 XYZ
}

// NewTriangle returns a triangle with the given vertices. The vertices are
// stored as XYZ; pass NaN Z values for purely 2D use.
func NewTriangle(p0, p1, p2 XYZ) Triangle {
	return Triangle{P0: p0, P1: p1, P2: p2}
}

// TriangleSignedArea returns the signed 2D area of the triangle a-b-c.
// Positive for clockwise orientation, negative for counter-clockwise.
//
// JTS: Triangle.signedArea(Coordinate, Coordinate, Coordinate)
func TriangleSignedArea(a, b, c XY) float64 {
	return ((c.X-a.X)*(b.Y-a.Y) - (b.X-a.X)*(c.Y-a.Y)) / 2
}

// TriangleArea returns the unsigned 2D area of the triangle a-b-c.
//
// JTS: Triangle.area(Coordinate, Coordinate, Coordinate)
func TriangleArea(a, b, c XY) float64 {
	return math.Abs(TriangleSignedArea(a, b, c))
}

// TriangleCentroid returns the centroid (mean of vertex coordinates) of
// triangle a-b-c. The centroid always lies within the triangle.
//
// JTS: Triangle.centroid(Coordinate, Coordinate, Coordinate)
func TriangleCentroid(a, b, c XY) XY {
	return XY{
		X: (a.X + b.X + c.X) / 3,
		Y: (a.Y + b.Y + c.Y) / 3,
	}
}

// TriangleCircumcentre returns the circumcentre of triangle a-b-c — the
// point equidistant from all three vertices. Uses the Shewchuk-style
// origin-normalisation for improved floating-point accuracy.
//
// JTS: Triangle.circumcentre(Coordinate, Coordinate, Coordinate)
func TriangleCircumcentre(a, b, c XY) XY {
	cx := c.X
	cy := c.Y
	ax := a.X - cx
	ay := a.Y - cy
	bx := b.X - cx
	by := b.Y - cy

	denom := 2 * triDet(ax, ay, bx, by)
	numx := triDet(ay, ax*ax+ay*ay, by, bx*bx+by*by)
	numy := triDet(ax, ax*ax+ay*ay, bx, bx*bx+by*by)

	return XY{X: cx - numx/denom, Y: cy + numy/denom}
}

// TriangleInCentre returns the incentre of triangle a-b-c — the point
// equidistant from all three sides, which is the centre of the inscribed
// circle. Always lies within the triangle.
//
// JTS: Triangle.inCentre(Coordinate, Coordinate, Coordinate)
func TriangleInCentre(a, b, c XY) XY {
	lenAB := triDistance(a, b)
	lenBC := triDistance(b, c)
	lenCA := triDistance(c, a)
	circum := lenBC + lenCA + lenAB
	return XY{
		X: (lenBC*a.X + lenCA*b.X + lenAB*c.X) / circum,
		Y: (lenBC*a.Y + lenCA*b.Y + lenAB*c.Y) / circum,
	}
}

// TriangleArea3D returns the unsigned 3D area of triangle a-b-c. Computed
// as half the magnitude of the cross product of the two edge vectors.
//
// JTS: Triangle.area3D(Coordinate, Coordinate, Coordinate)
func TriangleArea3D(a, b, c XYZ) float64 {
	ux := b.X - a.X
	uy := b.Y - a.Y
	uz := b.Z - a.Z

	vx := c.X - a.X
	vy := c.Y - a.Y
	vz := c.Z - a.Z

	crossx := uy*vz - uz*vy
	crossy := uz*vx - ux*vz
	crossz := ux*vy - uy*vx

	return math.Sqrt(crossx*crossx+crossy*crossy+crossz*crossz) / 2
}

// TriangleInterpolateZ returns the Z value of the planar surface defined by
// triangle v0-v1-v2 evaluated at p (X,Y). The triangle must be non-degenerate
// in XY and must not be parallel to the Z-axis.
//
// JTS: Triangle.interpolateZ(Coordinate, Coordinate, Coordinate, Coordinate)
func TriangleInterpolateZ(p XY, v0, v1, v2 XYZ) float64 {
	x0 := v0.X
	y0 := v0.Y
	a := v1.X - x0
	b := v2.X - x0
	c := v1.Y - y0
	d := v2.Y - y0
	det := a*d - b*c
	dx := p.X - x0
	dy := p.Y - y0
	t := (d*dx - b*dy) / det
	u := (-c*dx + a*dy) / det
	return v0.Z + t*(v1.Z-v0.Z) + u*(v2.Z-v0.Z)
}

// SignedArea returns the signed 2D area of this triangle (positive for CW,
// negative for CCW orientation).
func (t Triangle) SignedArea() float64 { return TriangleSignedArea(t.P0.AsXY(), t.P1.AsXY(), t.P2.AsXY()) }

// Area returns the unsigned 2D area of this triangle.
func (t Triangle) Area() float64 { return TriangleArea(t.P0.AsXY(), t.P1.AsXY(), t.P2.AsXY()) }

// Centroid returns the centroid (centre of mass) of this triangle.
func (t Triangle) Centroid() XY { return TriangleCentroid(t.P0.AsXY(), t.P1.AsXY(), t.P2.AsXY()) }

// Circumcentre returns the circumcentre of this triangle.
func (t Triangle) Circumcentre() XY {
	return TriangleCircumcentre(t.P0.AsXY(), t.P1.AsXY(), t.P2.AsXY())
}

// InCentre returns the incentre of this triangle.
func (t Triangle) InCentre() XY { return TriangleInCentre(t.P0.AsXY(), t.P1.AsXY(), t.P2.AsXY()) }

// Area3D returns the 3D area of this triangle, using the Z values of each
// vertex. The result is always non-negative.
func (t Triangle) Area3D() float64 { return TriangleArea3D(t.P0, t.P1, t.P2) }

// InterpolateZ returns the Z value of the planar surface defined by this
// triangle at the supplied XY point.
func (t Triangle) InterpolateZ(p XY) float64 { return TriangleInterpolateZ(p, t.P0, t.P1, t.P2) }

func triDet(m00, m01, m10, m11 float64) float64 {
	return m00*m11 - m01*m10
}

func triDistance(a, b XY) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	return math.Sqrt(dx*dx + dy*dy)
}
