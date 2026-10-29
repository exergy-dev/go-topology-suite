package proj

import (
	"math"
	"testing"

	"github.com/exergy-dev/go-topology-suite/crs"
	"github.com/stretchr/testify/assert"
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
			assert.InDeltaf(rt, lon, lon2, 1e-12,
				"WebMercator rt lon: in (%v, %v) → (%v, %v) → (%v, %v)",
				lon, lat, x, y, lon2, lat2)
			assert.InDeltaf(rt, lat, lat2, 1e-12,
				"WebMercator rt lat: in (%v, %v) → (%v, %v) → (%v, %v)",
				lon, lat, x, y, lon2, lat2)
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
			assert.InDeltaf(rt, lon, lon2, 1e-11,
				"UTM30 rt lon: in (%v, %v) → (%v, %v) → (%v, %v)",
				lon, lat, x, y, lon2, lat2)
			assert.InDeltaf(rt, lat, lat2, 1e-11,
				"UTM30 rt lat: in (%v, %v) → (%v, %v) → (%v, %v)",
				lon, lat, x, y, lon2, lat2)
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
			assert.InDeltaf(rt, lon, lon2, 1e-11,
				"Lambert93 rt lon: in (%v, %v) → (%v, %v) → (%v, %v)",
				lon, lat, x, y, lon2, lat2)
			assert.InDeltaf(rt, lat, lat2, 1e-11,
				"Lambert93 rt lat: in (%v, %v) → (%v, %v) → (%v, %v)",
				lon, lat, x, y, lon2, lat2)
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
			assert.InDeltaf(rt, lon, lon2, 1e-11,
				"ConusAlbers rt lon: in (%v, %v) → (%v, %v) → (%v, %v)",
				lon, lat, x, y, lon2, lat2)
			assert.InDeltaf(rt, lat, lat2, 1e-11,
				"ConusAlbers rt lat: in (%v, %v) → (%v, %v) → (%v, %v)",
				lon, lat, x, y, lon2, lat2)
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
			assert.InDeltaf(rt, lon, lon2, 1e-10,
				"EuropeLAEA rt lon: in (%v, %v) → (%v, %v) → (%v, %v)",
				lon, lat, x, y, lon2, lat2)
			assert.InDeltaf(rt, lat, lat2, 1e-10,
				"EuropeLAEA rt lat: in (%v, %v) → (%v, %v) → (%v, %v)",
				lon, lat, x, y, lon2, lat2)
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
			assert.InDeltaf(rt, 0.0, dLon, 1e-10,
				"LAEA polar rt lon: in (%v, %v) → (%v, %v) → (%v, %v)",
				lon, lat, x, y, lon2, lat2)
			assert.InDeltaf(rt, lat, lat2, 1e-10,
				"LAEA polar rt lat: in (%v, %v) → (%v, %v) → (%v, %v)",
				lon, lat, x, y, lon2, lat2)
		})
	})
}

// TestWebMercatorOriginIsPositiveZero pins the projection origin to an
// exact +0. The textbook ln(tan(π/4 + φ/2)) form gave n ≈ -1.4e-9 m at
// φ = 0 on arm64 (ulp noise from math.Tan(π/4), whose argument reduction
// is FMA-fused there but not on amd64), which surfaced as "-0" when
// formatted and made results platform-dependent.
func TestWebMercatorOriginIsPositiveZero(t *testing.T) {
	p := NewWebMercator()
	e, n := p.Forward(0, 0)
	assert.Equal(t, 0.0, e, "easting at origin")
	assert.Equal(t, 0.0, n, "northing at origin")
	assert.False(t, math.Signbit(e), "easting at origin must be +0, got %v", e)
	assert.False(t, math.Signbit(n), "northing at origin must be +0, got %v", n)

	// Forward is odd in latitude: mirrored latitudes give exactly
	// negated northings, so the equator is not biased to either side.
	for _, deg := range []float64{1e-9, 0.5, 10, 45, 60, 85} {
		lat := deg * math.Pi / 180
		_, nPos := p.Forward(0, lat)
		_, nNeg := p.Forward(0, -lat)
		assert.Equalf(t, nPos, -nNeg, "northing not odd at ±%v°", deg)
	}
}

// TestTransverseMercatorEquatorIsPositiveZero is the conformalLatitude
// counterpart of TestWebMercatorOriginIsPositiveZero: a point on the
// equator must map to a northing of exactly +0 (false northing 0), on
// every GOARCH.
func TestTransverseMercatorEquatorIsPositiveZero(t *testing.T) {
	assert.Equal(t, 0.0, conformalLatitude(0, math.Sqrt(crs.WGS84Ellipsoid.E2())))
	assert.False(t, math.Signbit(conformalLatitude(0, math.Sqrt(crs.WGS84Ellipsoid.E2()))))

	p := UTM(31, false, crs.WGS84Ellipsoid) // CM = 3°
	for _, lonDeg := range []float64{3, 0, 6} {
		_, n := p.Forward(lonDeg*math.Pi/180, 0)
		assert.Equalf(t, 0.0, n, "UTM31 northing on equator at lon %v°", lonDeg)
		assert.Falsef(t, math.Signbit(n), "UTM31 northing on equator at lon %v° must be +0, got %v", lonDeg, n)
	}
}
