package wkt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/terra-geo/terra/geom"
)

func TestRoundTripPointXYM(t *testing.T) {
	in := "POINT M (1 2 99)"
	g, err := Unmarshal(in)
	require.NoError(t, err)
	assert.Equal(t, geom.LayoutXYM, g.Layout(), "layout")
	out, _ := Marshal(g)
	assert.Equal(t, in, out)
}

func TestRoundTripPointXYZM(t *testing.T) {
	in := "POINT ZM (1 2 3 4)"
	g, err := Unmarshal(in)
	require.NoError(t, err)
	assert.Equal(t, geom.LayoutXYZM, g.Layout(), "layout")
	pp := g.(*geom.Point)
	assert.Equal(t, float64(3), pp.Z(), "Z")
	assert.Equal(t, float64(4), pp.M(), "M")
	out, _ := Marshal(g)
	assert.Equal(t, in, out)
}

// TestGluedDimensionSuffix verifies JTS-compatible glued forms POINTZ,
// LINESTRINGZM, POLYGONM, etc. (no whitespace between type and modifier).
func TestGluedDimensionSuffix(t *testing.T) {
	cases := []struct {
		in     string
		layout geom.Layout
	}{
		{"POINTZ (1 2 3)", geom.LayoutXYZ},
		{"POINTM (1 2 99)", geom.LayoutXYM},
		{"POINTZM (1 2 3 4)", geom.LayoutXYZM},
		{"LINESTRINGZ (0 0 1, 1 1 2)", geom.LayoutXYZ},
		{"LINESTRINGZM (0 0 1 5, 1 1 2 6)", geom.LayoutXYZM},
		// POLYGON ring storage is intrinsically XY in this codebase, so
		// we only assert the type token is recognised, not that Z is
		// preserved on rings (matches the existing POLYGON Z behaviour).
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			g, err := Unmarshal(tc.in)
			require.NoError(t, err)
			assert.Equal(t, tc.layout, g.Layout())
		})
	}
	// POLYGONZ should at least parse without error even though the
	// polygon model only retains XY ordinates.
	_, err := Unmarshal("POLYGONZ ((0 0 1, 1 0 1, 1 1 1, 0 0 1))")
	require.NoError(t, err)
	_, err = Unmarshal("MULTIPOLYGONZM (((0 0 1 5, 1 0 1 5, 1 1 1 5, 0 0 1 5)))")
	require.NoError(t, err)
}
