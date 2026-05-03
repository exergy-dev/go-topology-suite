package predicate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/terra-geo/terra/wkt"
)

// TestRelateNGOptEquivalence verifies that opting into the RelateNG
// driver produces the same DE-9IM matrix as the legacy path on a
// representative set of point-locator-only inputs (the cases the
// current RelateNG port handles). Inputs that depend on edge-segment
// crossings are answered by the legacy fallback inside Relate, so
// this loop also exercises the fallback path.
func TestRelateNGOptEquivalence(t *testing.T) {
	cases := []struct {
		name string
		a, b string
	}{
		// Point/point
		{"pp-equal", "POINT (1 1)", "POINT (1 1)"},
		{"pp-disjoint", "POINT (1 1)", "POINT (2 2)"},
		{"mp-mp-overlap", "MULTIPOINT ((1 1),(2 2))", "MULTIPOINT ((2 2),(3 3))"},
		// Point/polygon
		{"pt-in-poly", "POINT (5 5)", "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))"},
		{"pt-on-boundary", "POINT (5 0)", "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))"},
		{"pt-outside", "POINT (20 20)", "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))"},
		// Disjoint polygons (no edge intersection -> handled by point path)
		{"poly-disjoint", "POLYGON ((0 0, 1 0, 1 1, 0 1, 0 0))", "POLYGON ((10 10, 11 10, 11 11, 10 11, 10 10))"},
		// Disjoint point/line
		{"pt-line-disjoint", "POINT (5 5)", "LINESTRING (0 0, 1 1)"},
		// Empty cases
		{"empty-a", "POINT EMPTY", "POINT (1 1)"},
		{"empty-b", "POINT (1 1)", "POINT EMPTY"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			a, err := wkt.Unmarshal(c.a)
			require.NoError(t, err)
			b, err := wkt.Unmarshal(c.b)
			require.NoError(t, err)

			legacy, err := Relate(a, b)
			require.NoError(t, err)
			ng, err := Relate(a, b, UseRelateNG(true))
			require.NoError(t, err)
			assert.Equal(t, string(legacy), string(ng),
				"RelateNG vs legacy mismatch for %s/%s", c.a, c.b)
		})
	}
}
