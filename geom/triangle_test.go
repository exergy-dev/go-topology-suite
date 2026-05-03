package geom

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTriangleSignedArea(t *testing.T) {
	// CCW oriented triangle should give negative signed area per JTS.
	a := XY{0, 0}
	b := XY{4, 0}
	c := XY{0, 3}
	got := TriangleSignedArea(a, b, c)
	assert.InDelta(t, -6.0, got, 1e-12, "CCW signed area")

	// CW oriented triangle should give positive signed area.
	got = TriangleSignedArea(a, c, b)
	assert.InDelta(t, 6.0, got, 1e-12, "CW signed area")
}

func TestTriangleArea(t *testing.T) {
	a := XY{0, 0}
	b := XY{4, 0}
	c := XY{0, 3}
	assert.InDelta(t, 6.0, TriangleArea(a, b, c), 1e-12, "unsigned area CCW")
	assert.InDelta(t, 6.0, TriangleArea(a, c, b), 1e-12, "unsigned area CW")

	// Degenerate (colinear) triangle has zero area.
	d := XY{2, 0}
	assert.InDelta(t, 0.0, TriangleArea(a, b, d), 1e-12, "colinear area")
}

func TestTriangleCentroid(t *testing.T) {
	a := XY{0, 0}
	b := XY{6, 0}
	c := XY{0, 9}
	g := TriangleCentroid(a, b, c)
	assert.InDelta(t, 2.0, g.X, 1e-12, "centroid X")
	assert.InDelta(t, 3.0, g.Y, 1e-12, "centroid Y")
}

func TestTriangleCircumcentre(t *testing.T) {
	// Right triangle: circumcentre is midpoint of hypotenuse.
	a := XY{0, 0}
	b := XY{4, 0}
	c := XY{0, 4}
	cc := TriangleCircumcentre(a, b, c)
	assert.InDelta(t, 2.0, cc.X, 1e-9, "circumcentre X")
	assert.InDelta(t, 2.0, cc.Y, 1e-9, "circumcentre Y")

	// Verify equidistance: |cc-a| == |cc-b| == |cc-c|.
	r := triDistance(cc, a)
	assert.InDelta(t, r, triDistance(cc, b), 1e-9, "equidistant b")
	assert.InDelta(t, r, triDistance(cc, c), 1e-9, "equidistant c")
}

func TestTriangleInCentre(t *testing.T) {
	// 3-4-5 right triangle: inradius = (a+b-c)/2 = (3+4-5)/2 = 1, incentre at (1,1).
	a := XY{0, 0}
	b := XY{4, 0}
	c := XY{0, 3}
	ic := TriangleInCentre(a, b, c)
	assert.InDelta(t, 1.0, ic.X, 1e-12, "incentre X")
	assert.InDelta(t, 1.0, ic.Y, 1e-12, "incentre Y")
}

func TestTriangleArea3D(t *testing.T) {
	// Triangle in the XY plane: 3D area should equal 2D area.
	a := XYZ{0, 0, 0}
	b := XYZ{4, 0, 0}
	c := XYZ{0, 3, 0}
	assert.InDelta(t, 6.0, TriangleArea3D(a, b, c), 1e-12, "flat triangle 3D area")

	// Vertical triangle: 3-4-5 right triangle in XZ plane, area = 6.
	a2 := XYZ{0, 0, 0}
	b2 := XYZ{4, 0, 0}
	c2 := XYZ{0, 0, 3}
	assert.InDelta(t, 6.0, TriangleArea3D(a2, b2, c2), 1e-12, "vertical triangle 3D area")
}

func TestTriangleInterpolateZ(t *testing.T) {
	// Plane z = x + y over triangle (0,0,0)-(1,0,1)-(0,1,1).
	v0 := XYZ{0, 0, 0}
	v1 := XYZ{1, 0, 1}
	v2 := XYZ{0, 1, 1}

	z := TriangleInterpolateZ(XY{0.5, 0.5}, v0, v1, v2)
	assert.InDelta(t, 1.0, z, 1e-12, "interpolated Z mid")

	z = TriangleInterpolateZ(XY{0, 0}, v0, v1, v2)
	assert.InDelta(t, 0.0, z, 1e-12, "interpolated Z at v0")

	z = TriangleInterpolateZ(XY{0.25, 0.25}, v0, v1, v2)
	assert.InDelta(t, 0.5, z, 1e-12, "interpolated Z quarter")
}

func TestTriangleMethods(t *testing.T) {
	tri := NewTriangle(XYZ{0, 0, 0}, XYZ{4, 0, 0}, XYZ{0, 3, 0})
	assert.InDelta(t, 6.0, tri.Area(), 1e-12, "Triangle.Area")
	assert.InDelta(t, -6.0, tri.SignedArea(), 1e-12, "Triangle.SignedArea")
	g := tri.Centroid()
	assert.InDelta(t, 4.0/3.0, g.X, 1e-12, "Triangle.Centroid X")
	assert.InDelta(t, 1.0, g.Y, 1e-12, "Triangle.Centroid Y")
	assert.InDelta(t, 6.0, tri.Area3D(), 1e-12, "Triangle.Area3D")

	ic := tri.InCentre()
	assert.InDelta(t, 1.0, ic.X, 1e-12, "Triangle.InCentre X")
	assert.InDelta(t, 1.0, ic.Y, 1e-12, "Triangle.InCentre Y")

	cc := tri.Circumcentre()
	// Hypotenuse midpoint of right triangle (0,0)-(4,0)-(0,3) is (2,1.5).
	assert.InDelta(t, 2.0, cc.X, 1e-9, "Triangle.Circumcentre X")
	assert.InDelta(t, 1.5, cc.Y, 1e-9, "Triangle.Circumcentre Y")

	// Plane z = x + y triangle.
	tri2 := NewTriangle(XYZ{0, 0, 0}, XYZ{1, 0, 1}, XYZ{0, 1, 1})
	z := tri2.InterpolateZ(XY{0.5, 0.5})
	assert.InDelta(t, 1.0, z, 1e-12, "Triangle.InterpolateZ")
	assert.False(t, math.IsNaN(z), "InterpolateZ not NaN")
}
