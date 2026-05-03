package predicate

import (
	"testing"

	"github.com/terra-geo/terra/geom"
)

// TestShortCircuitDimMismatch exercises the JTS RelateNG-style
// dim-mismatch fast paths in scCovers/scContains. A line cannot cover
// or contain a polygon, etc. — we should reach this conclusion without
// invoking the topology graph.
func TestShortCircuitDimMismatch(t *testing.T) {
	pt := geom.NewPoint(nil, geom.XY{X: 0, Y: 0})
	line := geom.NewLineString(nil, []geom.XY{{X: 0, Y: 0}, {X: 1, Y: 1}})
	poly := scTestPoly()

	cases := []struct {
		name     string
		a, b     geom.Geometry
		contains bool
		covers   bool
	}{
		// Point cannot contain a line/polygon.
		{"point.contains(line)", pt, line, false, false},
		{"point.contains(poly)", pt, poly, false, false},
		// Line cannot contain a polygon.
		{"line.contains(poly)", line, poly, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Contains(tc.a, tc.b)
			if err != nil {
				t.Fatalf("Contains: %v", err)
			}
			if got != tc.contains {
				t.Errorf("Contains: want %v got %v", tc.contains, got)
			}
			gotC, err := Covers(tc.a, tc.b)
			if err != nil {
				t.Fatalf("Covers: %v", err)
			}
			if gotC != tc.covers {
				t.Errorf("Covers: want %v got %v", tc.covers, gotC)
			}
		})
	}
}

// TestShortCircuitOverlapsDimMismatch verifies that Overlaps shortcuts
// to false when A and B have different topological dimensions.
func TestShortCircuitOverlapsDimMismatch(t *testing.T) {
	pt := geom.NewPoint(nil, geom.XY{X: 1, Y: 1})
	line := geom.NewLineString(nil, []geom.XY{{X: 0, Y: 0}, {X: 2, Y: 2}})
	got, err := Overlaps(pt, line)
	if err != nil {
		t.Fatal(err)
	}
	if got {
		t.Errorf("Overlaps(point, line): want false got true")
	}
}

// TestShortCircuitCrossesPP exercises P/P short-circuit (always false
// per JTS).
func TestShortCircuitCrossesPP(t *testing.T) {
	a := geom.NewPoint(nil, geom.XY{X: 0, Y: 0})
	b := geom.NewPoint(nil, geom.XY{X: 0, Y: 0})
	got, err := Crosses(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if got {
		t.Errorf("Crosses(point, point): want false got true")
	}
}

// TestShortCircuitTouchesPP: two pure points cannot touch.
func TestShortCircuitTouchesPP(t *testing.T) {
	a := geom.NewPoint(nil, geom.XY{X: 0, Y: 0})
	b := geom.NewPoint(nil, geom.XY{X: 0, Y: 0})
	got, err := Touches(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if got {
		t.Errorf("Touches(point, point): want false got true")
	}
}

// TestShortCircuitEnvelopeDisjoint verifies the envelope-disjoint
// fast paths shared by Intersects/Crosses/Touches/Overlaps.
func TestShortCircuitEnvelopeDisjoint(t *testing.T) {
	a := scTestPoly()
	b := scTestPolyAt(10, 10)

	if got, _ := Intersects(a, b); got {
		t.Errorf("Intersects: want false")
	}
	if got, _ := Disjoint(a, b); !got {
		t.Errorf("Disjoint: want true")
	}
	if got, _ := Touches(a, b); got {
		t.Errorf("Touches: want false")
	}
	if got, _ := Overlaps(a, b); got {
		t.Errorf("Overlaps: want false")
	}
	if got, _ := Crosses(a, b); got {
		t.Errorf("Crosses: want false")
	}
}

// TestShortCircuitEqualsEnvelopeDiff: differing envelopes resolve
// false without topology.
func TestShortCircuitEqualsEnvelopeDiff(t *testing.T) {
	a := scTestPoly()
	b := scTestPolyAt(0, 0) // larger
	got, err := Equals(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if got {
		t.Errorf("Equals(envelope-diff polygons): want false got true")
	}
}

// TestRelateNGSurface exercises the experimental RelateNG type to
// verify it dispatches to the same package-level predicates.
func TestRelateNGSurface(t *testing.T) {
	a := geom.NewPolygon(nil, []geom.XY{
		{X: 0, Y: 0}, {X: 0, Y: 4}, {X: 4, Y: 4}, {X: 4, Y: 0}, {X: 0, Y: 0},
	})
	pIn := geom.NewPoint(nil, geom.XY{X: 1, Y: 1})
	pOut := geom.NewPoint(nil, geom.XY{X: 9, Y: 9})

	r := NewRelateNG(a)
	if got, _ := r.Contains(pIn); !got {
		t.Errorf("RelateNG.Contains(inner point): want true")
	}
	if got, _ := r.Contains(pOut); got {
		t.Errorf("RelateNG.Contains(outer point): want false")
	}
	if got, _ := r.Disjoint(pOut); !got {
		t.Errorf("RelateNG.Disjoint(outer point): want true")
	}
	if got, _ := r.Intersects(pIn); !got {
		t.Errorf("RelateNG.Intersects(inner point): want true")
	}
}

func scTestPoly() *geom.Polygon {
	return geom.NewPolygon(nil, []geom.XY{
		{X: 0, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 1}, {X: 1, Y: 0}, {X: 0, Y: 0},
	})
}

func scTestPolyAt(x, y float64) *geom.Polygon {
	// 2x2 box at the given origin — distinct envelope from scTestPoly.
	return geom.NewPolygon(nil, []geom.XY{
		{X: x, Y: y}, {X: x, Y: y + 2}, {X: x + 2, Y: y + 2}, {X: x + 2, Y: y}, {X: x, Y: y},
	})
}
