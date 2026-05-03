package relateng

import (
	"testing"

	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/wkt"
)

func mustParse(t *testing.T, s string) geom.Geometry {
	t.Helper()
	g, err := wkt.Unmarshal(s)
	if err != nil {
		t.Fatalf("wkt: %v", err)
	}
	return g
}

func TestRelateGeometry_Dimension(t *testing.T) {
	cases := []struct {
		wkt    string
		dim    int
		hasP   bool
		hasL   bool
		hasA   bool
	}{
		{"POINT (1 2)", DimP, true, false, false},
		{"MULTIPOINT (1 2, 3 4)", DimP, true, false, false},
		{"LINESTRING (0 0, 1 1)", DimL, false, true, false},
		{"POLYGON ((0 0, 1 0, 1 1, 0 1, 0 0))", DimA, false, false, true},
		{"GEOMETRYCOLLECTION (POINT(1 2), LINESTRING(0 0, 1 0))", DimL, true, true, false},
		{"GEOMETRYCOLLECTION (POINT(0 0), POLYGON((0 0,1 0,1 1,0 1,0 0)))", DimA, true, false, true},
	}
	for _, tc := range cases {
		t.Run(tc.wkt, func(t *testing.T) {
			rg := NewGeometry(mustParse(t, tc.wkt))
			if rg.Dimension() != tc.dim {
				t.Errorf("dim = %d, want %d", rg.Dimension(), tc.dim)
			}
			if rg.HasDimension(DimP) != tc.hasP {
				t.Errorf("hasP = %v, want %v", rg.HasDimension(DimP), tc.hasP)
			}
			if rg.HasDimension(DimL) != tc.hasL {
				t.Errorf("hasL = %v, want %v", rg.HasDimension(DimL), tc.hasL)
			}
			if rg.HasDimension(DimA) != tc.hasA {
				t.Errorf("hasA = %v, want %v", rg.HasDimension(DimA), tc.hasA)
			}
		})
	}
}

func TestRelateGeometry_DimensionReal_ZeroLengthLine(t *testing.T) {
	rg := NewGeometry(mustParse(t, "LINESTRING (5 5, 5 5)"))
	if got := rg.DimensionReal(); got != DimP {
		t.Errorf("zero-length line: DimensionReal = %d, want %d", got, DimP)
	}
}

func TestRelateGeometry_HasBoundary(t *testing.T) {
	cases := []struct {
		wkt  string
		want bool
	}{
		{"POINT (1 2)", false},
		{"LINESTRING (0 0, 1 1)", true},
		{"LINESTRING (0 0, 1 0, 0 0)", false}, // closed → no boundary under Mod2
		{"POLYGON ((0 0,1 0,1 1,0 1,0 0))", false}, // areal: locator has no line boundary
	}
	for _, tc := range cases {
		t.Run(tc.wkt, func(t *testing.T) {
			rg := NewGeometry(mustParse(t, tc.wkt))
			if got := rg.HasBoundary(); got != tc.want {
				t.Errorf("HasBoundary = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestRelateGeometry_IsSelfNodingRequired(t *testing.T) {
	cases := []struct {
		wkt  string
		want bool
	}{
		{"POINT (1 2)", false},
		{"POLYGON ((0 0,1 0,1 1,0 1,0 0))", false},
		{"MULTIPOLYGON (((0 0,1 0,1 1,0 1,0 0)))", false},
		{"LINESTRING (0 0, 1 1)", true},
		{"GEOMETRYCOLLECTION (POINT(0 0), LINESTRING(1 1, 2 2))", true},
	}
	for _, tc := range cases {
		t.Run(tc.wkt, func(t *testing.T) {
			rg := NewGeometry(mustParse(t, tc.wkt))
			if got := rg.IsSelfNodingRequired(); got != tc.want {
				t.Errorf("IsSelfNodingRequired = %v, want %v", got, tc.want)
			}
		})
	}
}
