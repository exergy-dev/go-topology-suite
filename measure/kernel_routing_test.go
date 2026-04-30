package measure

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/terra-geo/terra/crs"
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/kernel/planar"
)

// When a geometry carries a geographic CRS the default kernel is
// geodesic — the same coordinate pair therefore returns metres on the
// ellipsoid rather than degrees in the plane. This is the load-bearing
// promise of the design memo.
func TestDistanceDefaultsToGeodesicForGeographic(t *testing.T) {
	a := geom.NewPoint(crs.WGS84, geom.XY{X: -73.7781, Y: 40.6413})
	b := geom.NewPoint(crs.WGS84, geom.XY{X: -118.4085, Y: 33.9416})
	d, err := Distance(a, b)
	require.NoError(t, err)
	// Expect ~3,983 km on WGS84 (Vincenty), not ~45 in degree-units.
	assert.True(t, d >= 3.9e6 && d <= 4.0e6, "expected ~3,983 km in metres, got %v", d)
}

func TestDistanceDefaultsToPlanarForUnsetCRS(t *testing.T) {
	a := geom.NewPoint(nil, geom.XY{X: 0, Y: 0})
	b := geom.NewPoint(nil, geom.XY{X: 3, Y: 4})
	d, err := Distance(a, b)
	require.NoError(t, err)
	assert.InDelta(t, 5.0, d, 1e-9, "planar default expected 5, got %v", d)
}

func TestWithKernelOverridesDefault(t *testing.T) {
	// Geographic CRS would normally pick geodesic; force planar.
	a := geom.NewPoint(crs.WGS84, geom.XY{X: 0, Y: 0})
	b := geom.NewPoint(crs.WGS84, geom.XY{X: 3, Y: 4})
	d, _ := Distance(a, b, WithKernel(planar.Default))
	assert.InDelta(t, 5.0, d, 1e-9, "planar override expected 5, got %v", d)
}
