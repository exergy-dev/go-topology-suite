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
