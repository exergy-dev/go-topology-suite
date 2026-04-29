package predicate

import (
	"testing"

	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/internal/proptest"
	"pgregory.net/rapid"
)

// TestIntersectsSymmetric: Intersects is symmetric over the operands.
func TestIntersectsSymmetric(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		a := geom.NewPoint(nil, proptest.SmallXY(t))
		b := geom.NewPoint(nil, proptest.SmallXY(t))
		ab, _ := Intersects(a, b)
		ba, _ := Intersects(b, a)
		if ab != ba {
			t.Fatalf("Intersects not symmetric: %v vs %v", ab, ba)
		}
	})
}

// TestEqualsReflexive: Equals(g, g) is always true.
func TestEqualsReflexive(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		p := geom.NewPoint(nil, proptest.AnyXY(t))
		eq, err := Equals(p, p)
		if err != nil {
			t.Fatal(err)
		}
		if !eq {
			t.Fatalf("Equals not reflexive for %v", p.XY())
		}
	})
}

// TestDisjointIsComplementOfIntersects.
func TestDisjointIsComplementOfIntersects(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		a := geom.NewPoint(nil, proptest.SmallXY(t))
		b := geom.NewPoint(nil, proptest.SmallXY(t))
		i, _ := Intersects(a, b)
		d, _ := Disjoint(a, b)
		if i == d {
			t.Fatalf("Disjoint and Intersects should never agree (got %v, %v)", i, d)
		}
	})
}

// TestContainsImpliesIntersects.
func TestContainsImpliesIntersects(t *testing.T) {
	// Build a random outer square and a random inner point. If the
	// point is inside the square (Contains), it must also Intersect.
	rapid.Check(t, func(t *rapid.T) {
		x0 := rapid.Float64Range(-50, 50).Draw(t, "x")
		y0 := rapid.Float64Range(-50, 50).Draw(t, "y")
		w := rapid.Float64Range(1, 50).Draw(t, "w")
		h := rapid.Float64Range(1, 50).Draw(t, "h")
		poly := geom.NewPolygon(nil,
			[]geom.XY{
				{X: x0, Y: y0},
				{X: x0 + w, Y: y0},
				{X: x0 + w, Y: y0 + h},
				{X: x0, Y: y0 + h},
				{X: x0, Y: y0},
			})
		px := rapid.Float64Range(x0+0.01*w, x0+0.99*w).Draw(t, "px")
		py := rapid.Float64Range(y0+0.01*h, y0+0.99*h).Draw(t, "py")
		pt := geom.NewPoint(nil, geom.XY{X: px, Y: py})
		c, _ := Contains(poly, pt)
		i, _ := Intersects(poly, pt)
		if c && !i {
			t.Fatalf("Contains true but Intersects false")
		}
	})
}

// TestCoversIsLooserThanContains: Contains implies Covers.
func TestCoversIsLooserThanContains(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		x0 := rapid.Float64Range(-50, 50).Draw(t, "x")
		y0 := rapid.Float64Range(-50, 50).Draw(t, "y")
		w := rapid.Float64Range(1, 50).Draw(t, "w")
		h := rapid.Float64Range(1, 50).Draw(t, "h")
		poly := geom.NewPolygon(nil,
			[]geom.XY{
				{X: x0, Y: y0},
				{X: x0 + w, Y: y0},
				{X: x0 + w, Y: y0 + h},
				{X: x0, Y: y0 + h},
				{X: x0, Y: y0},
			})
		px := rapid.Float64Range(-100, 100).Draw(t, "px")
		py := rapid.Float64Range(-100, 100).Draw(t, "py")
		pt := geom.NewPoint(nil, geom.XY{X: px, Y: py})
		c, _ := Contains(poly, pt)
		cov, _ := Covers(poly, pt)
		if c && !cov {
			t.Fatalf("Contains=true but Covers=false")
		}
	})
}
