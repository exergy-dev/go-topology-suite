package relateng

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
		assert.Equal(t, tc.want, got, "Locate(%v)", tc.p)
		assert.Equal(t, tc.wantDim, gotDim, "LocateWithDim(%v)", tc.p)
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
		assert.Equal(t, tc.wantDim, gotDim, "LocateWithDim(%v)", tc.p)
	}
}

func TestPointLocator_ClosedLine_BoundaryEmpty(t *testing.T) {
	g := mustParse(t, "LINESTRING (0 0, 10 0, 10 10, 0 0)")
	loc := NewPointLocator(g)
	assert.False(t, loc.HasBoundary(), "closed line should have no boundary under Mod2")
	assert.Equal(t, DLLineInterior, loc.LocateWithDim(geom.XY{X: 0, Y: 0}),
		"closed line endpoint")
}

func TestPointLocator_Point(t *testing.T) {
	g := mustParse(t, "MULTIPOINT (1 1, 2 2)")
	loc := NewPointLocator(g)
	assert.Equal(t, DLPointInterior, loc.LocateWithDim(geom.XY{X: 1, Y: 1}), "MP hit")
	assert.Equal(t, DLExterior, loc.LocateWithDim(geom.XY{X: 3, Y: 3}), "MP miss")
}

func TestPointLocator_MixedGC_PrefersHigherDim(t *testing.T) {
	// Point at (5,5) lies on a line interior AND in a polygon interior.
	// The locator must report the higher-dim element (AREA_INTERIOR).
	g := mustParse(t, "GEOMETRYCOLLECTION (LINESTRING(0 5, 10 5), POLYGON((0 0,10 0,10 10,0 10,0 0)))")
	loc := NewPointLocator(g)
	assert.Equal(t, DLAreaInterior, loc.LocateWithDim(geom.XY{X: 5, Y: 5}),
		"mixed GC at interior should prefer AREA_INTERIOR")
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
		assert.Equal(t, tc.loc, Location(tc.dl), "Location(%d)", tc.dl)
		assert.Equal(t, tc.dim, Dimension(tc.dl), "Dimension(%d)", tc.dl)
	}
}

func TestLinearBoundary_Mod2(t *testing.T) {
	// Two open lines sharing one endpoint:
	//   (0 0)→(1 0) and (1 0)→(2 0)
	// (1 0) appears twice (degree 2, even → not boundary)
	// (0 0) and (2 0) appear once each (degree 1, odd → boundary)
	g := mustParse(t, "MULTILINESTRING ((0 0, 1 0), (1 0, 2 0))")
	loc := NewPointLocator(g)
	require.True(t, loc.HasBoundary(), "expected non-empty boundary")
	assert.True(t, loc.lineBoundary.IsBoundary(geom.XY{X: 0, Y: 0}), "(0,0) should be boundary")
	assert.False(t, loc.lineBoundary.IsBoundary(geom.XY{X: 1, Y: 0}), "(1,0) should not be boundary (degree 2)")
	assert.True(t, loc.lineBoundary.IsBoundary(geom.XY{X: 2, Y: 0}), "(2,0) should be boundary")
}
