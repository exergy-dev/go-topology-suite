package relateng

import (
	"testing"

	"github.com/exergy-dev/go-topology-suite/geom"
)

func TestPointLocator_Polygon(t *testing.T) {
	g := mustParse(t, "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	loc := NewPointLocator(g)
	cases := []struct {
		p       geom.XY
		want    int
		wantDim int
	}{
		{geom.XY{X: 5, Y: 5}, LocInterior, DLAreaInterior},
		{geom.XY{X: 5, Y: 0}, LocBoundary, DLAreaBoundary},
		{geom.XY{X: 0, Y: 0}, LocBoundary, DLAreaBoundary},
		{geom.XY{X: 100, Y: 100}, LocExterior, DLExterior},
	}
	for _, tc := range cases {
		got := loc.Locate(tc.p)
		gotDim := loc.LocateWithDim(tc.p)
		if got != tc.want || gotDim != tc.wantDim {
			t.Errorf("Locate(%v) = (%d,%d), want (%d,%d)", tc.p, got, gotDim, tc.want, tc.wantDim)
		}
	}
}

func TestPointLocator_LineString(t *testing.T) {
	g := mustParse(t, "LINESTRING (0 0, 10 0)")
	loc := NewPointLocator(g)
	cases := []struct {
		p       geom.XY
		wantDim int
	}{
		{geom.XY{X: 5, Y: 0}, DLLineInterior},
		{geom.XY{X: 0, Y: 0}, DLLineBoundary},
		{geom.XY{X: 10, Y: 0}, DLLineBoundary},
		{geom.XY{X: 5, Y: 1}, DLExterior},
	}
	for _, tc := range cases {
		gotDim := loc.LocateWithDim(tc.p)
		if gotDim != tc.wantDim {
			t.Errorf("LocateWithDim(%v) = %d, want %d", tc.p, gotDim, tc.wantDim)
		}
	}
}

func TestPointLocator_ClosedLine_BoundaryEmpty(t *testing.T) {
	g := mustParse(t, "LINESTRING (0 0, 10 0, 10 10, 0 0)")
	loc := NewPointLocator(g)
	if loc.HasBoundary() {
		t.Error("closed line should have no boundary under Mod2")
	}
	if got := loc.LocateWithDim(geom.XY{X: 0, Y: 0}); got != DLLineInterior {
		t.Errorf("closed line endpoint: dim/loc = %d, want %d", got, DLLineInterior)
	}
}

func TestPointLocator_Point(t *testing.T) {
	g := mustParse(t, "MULTIPOINT (1 1, 2 2)")
	loc := NewPointLocator(g)
	if got := loc.LocateWithDim(geom.XY{X: 1, Y: 1}); got != DLPointInterior {
		t.Errorf("MP hit: dim/loc = %d, want %d", got, DLPointInterior)
	}
	if got := loc.LocateWithDim(geom.XY{X: 3, Y: 3}); got != DLExterior {
		t.Errorf("MP miss: dim/loc = %d, want EXTERIOR", got)
	}
}

func TestPointLocator_MixedGC_PrefersHigherDim(t *testing.T) {
	// Point at (5,5) lies on a line interior AND in a polygon interior.
	// The locator must report the higher-dim element (AREA_INTERIOR).
	g := mustParse(t, "GEOMETRYCOLLECTION (LINESTRING(0 5, 10 5), POLYGON((0 0,10 0,10 10,0 10,0 0)))")
	loc := NewPointLocator(g)
	if got := loc.LocateWithDim(geom.XY{X: 5, Y: 5}); got != DLAreaInterior {
		t.Errorf("mixed GC at interior: dim/loc = %d, want AREA_INTERIOR", got)
	}
}

func TestDimensionLocation_Encoding(t *testing.T) {
	cases := []struct {
		dl  int
		loc int
		dim int
	}{
		{DLExterior, LocExterior, DimFalse},
		{DLPointInterior, LocInterior, DimP},
		{DLLineInterior, LocInterior, DimL},
		{DLLineBoundary, LocBoundary, DimL},
		{DLAreaInterior, LocInterior, DimA},
		{DLAreaBoundary, LocBoundary, DimA},
	}
	for _, tc := range cases {
		if got := Location(tc.dl); got != tc.loc {
			t.Errorf("Location(%d) = %d, want %d", tc.dl, got, tc.loc)
		}
		if got := Dimension(tc.dl); got != tc.dim {
			t.Errorf("Dimension(%d) = %d, want %d", tc.dl, got, tc.dim)
		}
	}
}

func TestLinearBoundary_Mod2(t *testing.T) {
	// Two open lines sharing one endpoint:
	//   (0 0)→(1 0) and (1 0)→(2 0)
	// (1 0) appears twice (degree 2, even → not boundary)
	// (0 0) and (2 0) appear once each (degree 1, odd → boundary)
	g := mustParse(t, "MULTILINESTRING ((0 0, 1 0), (1 0, 2 0))")
	loc := NewPointLocator(g)
	if !loc.HasBoundary() {
		t.Fatal("expected non-empty boundary")
	}
	if !loc.lineBoundary.IsBoundary(geom.XY{X: 0, Y: 0}) {
		t.Error("(0,0) should be boundary")
	}
	if loc.lineBoundary.IsBoundary(geom.XY{X: 1, Y: 0}) {
		t.Error("(1,0) should not be boundary (degree 2)")
	}
	if !loc.lineBoundary.IsBoundary(geom.XY{X: 2, Y: 0}) {
		t.Error("(2,0) should be boundary")
	}
}
