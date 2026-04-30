package predicate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/terra-geo/terra/crs"
	"github.com/terra-geo/terra/geom"
)

// A polygon enclosing the antimeridian is correctly tested only by the
// spherical kernel — the planar kernel would treat the same lon/lat
// coordinates as a non-overlapping pair of pieces. Here we verify that
// passing a WGS84-tagged ring routes through spherical.
func TestPredicateRoutesGeographicToSpherical(t *testing.T) {
	// Ring straddling the antimeridian: lon 170..-170, lat 0..10.
	ring := []geom.XY{
		{X: 170, Y: 0}, {X: -170, Y: 0}, {X: -170, Y: 10}, {X: 170, Y: 10}, {X: 170, Y: 0},
	}
	poly := geom.NewPolygon(crs.WGS84, ring)
	inside := geom.NewPoint(crs.WGS84, geom.XY{X: 180, Y: 5})
	outside := geom.NewPoint(crs.WGS84, geom.XY{X: 0, Y: 5})

	got, _ := Contains(poly, inside)
	assert.True(t, got, "antimeridian polygon should Contain (180, 5) under spherical kernel")
	got, _ = Contains(poly, outside)
	assert.False(t, got, "antimeridian polygon should not Contain (0, 5)")
}
