package proj

import (
	"math"
	"testing"

	"github.com/exergy-dev/go-topology-suite/crs"
	"pgregory.net/rapid"
)

// TestRoundTrips uses property-based testing to verify that for every
// projection go-topology-suite ships, Inverse(Forward(p)) ≈ p within a strict
// tolerance, over the projection's stated domain of validity.
func TestRoundTrips(t *testing.T) {
	d2r := math.Pi / 180.0

	t.Run("WebMercator", func(t *testing.T) {
		rapid.Check(t, func(rt *rapid.T) {
			lon := rapid.Float64Range(-180, 180).Draw(rt, "lon") * d2r
			lat := rapid.Float64Range(-85, 85).Draw(rt, "lat") * d2r
			p := NewWebMercator()
			x, y := p.Forward(lon, lat)
			lon2, lat2 := p.Inverse(x, y)
			if math.Abs(lon2-lon) > 1e-12 || math.Abs(lat2-lat) > 1e-12 {
				rt.Errorf("WebMercator rt: in (%v, %v) → (%v, %v) → (%v, %v)",
					lon, lat, x, y, lon2, lat2)
			}
		})
	})

	t.Run("TransverseMercator/UTM30", func(t *testing.T) {
		p := UTM(30, false, crs.WGS84Ellipsoid) // CM = -3°
		rapid.Check(t, func(rt *rapid.T) {
			// UTM is well-behaved within ±3° of the central meridian.
			lon := rapid.Float64Range(-6, 0).Draw(rt, "lon") * d2r
			lat := rapid.Float64Range(-80, 80).Draw(rt, "lat") * d2r
			x, y := p.Forward(lon, lat)
			lon2, lat2 := p.Inverse(x, y)
			if math.Abs(lon2-lon) > 1e-11 || math.Abs(lat2-lat) > 1e-11 {
				rt.Errorf("UTM30 rt: in (%v, %v) → (%v, %v) → (%v, %v)",
					lon, lat, x, y, lon2, lat2)
			}
		})
	})

	t.Run("LCC2SP/Lambert93", func(t *testing.T) {
		p := NewLambertConformalConic2SP(
			crs.GRS80Ellipsoid.A, crs.GRS80Ellipsoid.E2(),
			3*d2r, 46.5*d2r, 49*d2r, 44*d2r,
			700000.0, 6600000.0,
		)
		rapid.Check(t, func(rt *rapid.T) {
			lon := rapid.Float64Range(-5, 10).Draw(rt, "lon") * d2r
			lat := rapid.Float64Range(40, 52).Draw(rt, "lat") * d2r
			x, y := p.Forward(lon, lat)
			lon2, lat2 := p.Inverse(x, y)
			if math.Abs(lon2-lon) > 1e-11 || math.Abs(lat2-lat) > 1e-11 {
				rt.Errorf("Lambert93 rt: in (%v, %v) → (%v, %v) → (%v, %v)",
					lon, lat, x, y, lon2, lat2)
			}
		})
	})

	t.Run("Albers/CONUS", func(t *testing.T) {
		p := NewAlbersEqualAreaConic(
			crs.GRS80Ellipsoid.A, crs.GRS80Ellipsoid.E2(),
			-96*d2r, 23*d2r, 29.5*d2r, 45.5*d2r,
			0, 0,
		)
		rapid.Check(t, func(rt *rapid.T) {
			lon := rapid.Float64Range(-130, -65).Draw(rt, "lon") * d2r
			lat := rapid.Float64Range(20, 55).Draw(rt, "lat") * d2r
			x, y := p.Forward(lon, lat)
			lon2, lat2 := p.Inverse(x, y)
			if math.Abs(lon2-lon) > 1e-11 || math.Abs(lat2-lat) > 1e-11 {
				rt.Errorf("ConusAlbers rt: in (%v, %v) → (%v, %v) → (%v, %v)",
					lon, lat, x, y, lon2, lat2)
			}
		})
	})

	t.Run("LAEA/EuropeOblique", func(t *testing.T) {
		p := NewLambertAzimuthalEqualArea(
			crs.GRS80Ellipsoid.A, crs.GRS80Ellipsoid.E2(),
			10*d2r, 52*d2r,
			4321000, 3210000,
		)
		rapid.Check(t, func(rt *rapid.T) {
			lon := rapid.Float64Range(-25, 45).Draw(rt, "lon") * d2r
			lat := rapid.Float64Range(34, 71).Draw(rt, "lat") * d2r
			x, y := p.Forward(lon, lat)
			lon2, lat2 := p.Inverse(x, y)
			if math.Abs(lon2-lon) > 1e-10 || math.Abs(lat2-lat) > 1e-10 {
				rt.Errorf("EuropeLAEA rt: in (%v, %v) → (%v, %v) → (%v, %v)",
					lon, lat, x, y, lon2, lat2)
			}
		})
	})

	t.Run("LAEA/NorthPolar", func(t *testing.T) {
		p := NewLambertAzimuthalEqualArea(
			crs.GRS80Ellipsoid.A, crs.GRS80Ellipsoid.E2(),
			0, math.Pi/2, 0, 0,
		)
		rapid.Check(t, func(rt *rapid.T) {
			lon := rapid.Float64Range(-180, 180).Draw(rt, "lon") * d2r
			lat := rapid.Float64Range(0, 89).Draw(rt, "lat") * d2r
			x, y := p.Forward(lon, lat)
			lon2, lat2 := p.Inverse(x, y)
			// At higher latitudes longitude is meaningless; skip near-pole.
			if lat > 89.9*d2r {
				return
			}
			dLon := lon2 - lon
			for dLon > math.Pi {
				dLon -= 2 * math.Pi
			}
			for dLon < -math.Pi {
				dLon += 2 * math.Pi
			}
			if math.Abs(dLon) > 1e-10 || math.Abs(lat2-lat) > 1e-10 {
				rt.Errorf("LAEA polar rt: in (%v, %v) → (%v, %v) → (%v, %v)",
					lon, lat, x, y, lon2, lat2)
			}
		})
	})
}
