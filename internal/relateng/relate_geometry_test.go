package relateng

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/exergy-dev/go-topology-suite/geom"
	"github.com/exergy-dev/go-topology-suite/wkt"
)

func mustParse(t *testing.T, s string) geom.Geometry {
	t.Helper()
	g, err := wkt.Unmarshal(s)
	require.NoError(t, err, "wkt")
	return g
}

func TestRelateGeometry_Dimension(t *testing.T) {
	cases := []struct {
		wkt  string
		dim  int
		hasP bool
		hasL bool
		hasA bool
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
			assert.Equal(t, tc.dim, rg.Dimension(), "dim")
			assert.Equal(t, tc.hasP, rg.HasDimension(DimP), "hasP")
			assert.Equal(t, tc.hasL, rg.HasDimension(DimL), "hasL")
			assert.Equal(t, tc.hasA, rg.HasDimension(DimA), "hasA")
		})
	}
}

func TestRelateGeometry_DimensionReal_ZeroLengthLine(t *testing.T) {
	rg := NewGeometry(mustParse(t, "LINESTRING (5 5, 5 5)"))
	assert.Equal(t, DimP, rg.DimensionReal(), "zero-length line: DimensionReal")
}

func TestRelateGeometry_HasBoundary(t *testing.T) {
	cases := []struct {
		wkt  string
		want bool
	}{
		{"POINT (1 2)", false},
		{"LINESTRING (0 0, 1 1)", true},
		{"LINESTRING (0 0, 1 0, 0 0)", false},      // closed → no boundary under Mod2
		{"POLYGON ((0 0,1 0,1 1,0 1,0 0))", false}, // areal: locator has no line boundary
	}
	for _, tc := range cases {
		t.Run(tc.wkt, func(t *testing.T) {
			rg := NewGeometry(mustParse(t, tc.wkt))
			assert.Equal(t, tc.want, rg.HasBoundary(), "HasBoundary")
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
			assert.Equal(t, tc.want, rg.IsSelfNodingRequired(), "IsSelfNodingRequired")
		})
	}
}
