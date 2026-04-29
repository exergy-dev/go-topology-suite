package predicate

import (
	"errors"
	"testing"

	terra "github.com/terra-geo/terra"
	"github.com/terra-geo/terra/crs"
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/wkt"
)

func mustParse(t *testing.T, s string) geom.Geometry {
	t.Helper()
	g, err := wkt.Unmarshal(s)
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func TestIntersectsBasics(t *testing.T) {
	cases := []struct {
		name string
		a, b string
		want bool
	}{
		{"two points equal", "POINT (1 2)", "POINT (1 2)", true},
		{"two points distinct", "POINT (1 2)", "POINT (3 4)", false},
		{"point on line", "POINT (1 1)", "LINESTRING (0 0, 2 2)", true},
		{"point off line", "POINT (1 0)", "LINESTRING (0 0, 2 2)", false},
		{"point in polygon", "POINT (5 5)", "POLYGON ((0 0, 0 10, 10 10, 10 0, 0 0))", true},
		{"point outside polygon", "POINT (15 5)", "POLYGON ((0 0, 0 10, 10 10, 10 0, 0 0))", false},
		{"point on polygon boundary", "POINT (0 5)", "POLYGON ((0 0, 0 10, 10 10, 10 0, 0 0))", true},
		{"two lines crossing", "LINESTRING (0 0, 2 2)", "LINESTRING (0 2, 2 0)", true},
		{"two lines parallel", "LINESTRING (0 0, 2 0)", "LINESTRING (0 1, 2 1)", false},
		{"line touching polygon", "LINESTRING (5 -5, 5 5)", "POLYGON ((0 0, 0 10, 10 10, 10 0, 0 0))", true},
		{"line in polygon hole", "LINESTRING (3 3, 3.5 3.5)",
			"POLYGON ((0 0, 0 10, 10 10, 10 0, 0 0), (2 2, 2 4, 4 4, 4 2, 2 2))", false},
		{"polygons overlapping", "POLYGON ((0 0, 0 10, 10 10, 10 0, 0 0))",
			"POLYGON ((5 5, 5 15, 15 15, 15 5, 5 5))", true},
		{"polygons disjoint", "POLYGON ((0 0, 0 5, 5 5, 5 0, 0 0))",
			"POLYGON ((10 10, 10 15, 15 15, 15 10, 10 10))", false},
		{"polygons one inside other", "POLYGON ((0 0, 0 10, 10 10, 10 0, 0 0))",
			"POLYGON ((2 2, 2 4, 4 4, 4 2, 2 2))", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a, b := mustParse(t, tc.a), mustParse(t, tc.b)
			got, err := Intersects(a, b)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("Intersects = %v, want %v", got, tc.want)
			}
			// Disjoint must be the complement.
			d, err := Disjoint(a, b)
			if err != nil {
				t.Fatal(err)
			}
			if d == got {
				t.Errorf("Disjoint = %v, but Intersects = %v (must differ)", d, got)
			}
		})
	}
}

func TestContainsPolygonPoint(t *testing.T) {
	poly := mustParse(t, "POLYGON ((0 0, 0 10, 10 10, 10 0, 0 0))")
	inside := mustParse(t, "POINT (5 5)")
	outside := mustParse(t, "POINT (15 5)")
	boundary := mustParse(t, "POINT (0 5)")

	got, _ := Contains(poly, inside)
	if !got {
		t.Errorf("polygon should contain interior point")
	}
	got, _ = Contains(poly, outside)
	if got {
		t.Errorf("polygon should not contain external point")
	}
	got, _ = Contains(poly, boundary)
	if got {
		t.Errorf("polygon should not Contain boundary point (Covers does)")
	}
}

func TestContainsPolygonPolygon(t *testing.T) {
	outer := mustParse(t, "POLYGON ((0 0, 0 10, 10 10, 10 0, 0 0))")
	inner := mustParse(t, "POLYGON ((2 2, 2 4, 4 4, 4 2, 2 2))")
	overlap := mustParse(t, "POLYGON ((5 5, 5 15, 15 15, 15 5, 5 5))")

	got, _ := Contains(outer, inner)
	if !got {
		t.Errorf("outer should contain inner")
	}
	got, _ = Contains(outer, overlap)
	if got {
		t.Errorf("partial overlap should not be Contains")
	}
}

func TestEquals(t *testing.T) {
	a := mustParse(t, "POINT (1 2)")
	b := mustParse(t, "POINT (1 2)")
	c := mustParse(t, "POINT (1 3)")

	if got, _ := Equals(a, b); !got {
		t.Errorf("identical points should be Equal")
	}
	if got, _ := Equals(a, c); got {
		t.Errorf("differing points should not be Equal")
	}
	d := mustParse(t, "LINESTRING (1 2, 3 4)")
	if got, _ := Equals(a, d); got {
		t.Errorf("different types should not be Equal")
	}
}

func TestCRSMismatch(t *testing.T) {
	a := geom.NewPoint(crs.WGS84, geom.XY{X: 1, Y: 2})
	b := geom.NewPoint(crs.WebMercator, geom.XY{X: 1, Y: 2})
	_, err := Intersects(a, b)
	if !errors.Is(err, terra.ErrCRSMismatch) {
		t.Errorf("expected ErrCRSMismatch, got %v", err)
	}
}
