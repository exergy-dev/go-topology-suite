package projection

import (
	"fmt"
	"math"
)

// Mercator implements the Mercator projection (conformal cylindrical).
// The Mercator projection is widely used for navigation and web mapping.
// It preserves angles and shapes locally but distorts areas, especially
// near the poles.
type Mercator struct {
	Name            string
	CentralMeridian float64     // Central meridian in degrees
	FalseEasting    float64     // False easting in meters
	FalseNorthing   float64     // False northing in meters
	Ellipsoid       *Ellipsoid  // Reference ellipsoid
}

// NewMercator creates a new Mercator projection with the specified parameters.
func NewMercator(ellipsoid *Ellipsoid, centralMeridian, falseEasting, falseNorthing float64) *Mercator {
	return &Mercator{
		Name:            "Mercator",
		CentralMeridian: centralMeridian,
		FalseEasting:    falseEasting,
		FalseNorthing:   falseNorthing,
		Ellipsoid:       ellipsoid,
	}
}

// WebMercator creates a Web Mercator projection (EPSG:3857).
// This is the projection used by Google Maps, OpenStreetMap, and most web mapping applications.
// It uses a spherical formula with the WGS84 semi-major axis.
func WebMercator() *Mercator {
	return &Mercator{
		Name:            "Web Mercator (EPSG:3857)",
		CentralMeridian: 0,
		FalseEasting:    0,
		FalseNorthing:   0,
		Ellipsoid:       Sphere(6378137.0), // WGS84 semi-major axis
	}
}

// ProjectionName returns the name of this projection.
func (m *Mercator) ProjectionName() string {
	return m.Name
}

// Forward transforms from geographic coordinates (lon, lat in degrees)
// to projected coordinates (x, y in meters).
func (m *Mercator) Forward(lon, lat float64) (x, y float64, err error) {
	// Check latitude bounds (Mercator is undefined at the poles)
	if lat <= -90 || lat >= 90 {
		return 0, 0, fmt.Errorf("latitude %f is out of bounds for Mercator projection", lat)
	}

	// Convert degrees to radians
	lonRad := lon * math.Pi / 180.0
	latRad := lat * math.Pi / 180.0
	centralMeridianRad := m.CentralMeridian * math.Pi / 180.0

	a := m.Ellipsoid.A

	if m.Ellipsoid.IsSpherical() {
		// Spherical formulas (used by Web Mercator)
		x = a * (lonRad - centralMeridianRad)
		y = a * math.Log(math.Tan(math.Pi/4.0 + latRad/2.0))
	} else {
		// Ellipsoidal formulas
		e := math.Sqrt(m.Ellipsoid.Eccentricity)
		sinLat := math.Sin(latRad)

		x = a * (lonRad - centralMeridianRad)

		// y = a * ln(tan(π/4 + φ/2) * ((1 - e*sin(φ)) / (1 + e*sin(φ)))^(e/2))
		esinLat := e * sinLat
		conformalLat := math.Tan(math.Pi/4.0 + latRad/2.0) *
			math.Pow((1.0-esinLat)/(1.0+esinLat), e/2.0)
		y = a * math.Log(conformalLat)
	}

	// Apply false easting and northing
	x += m.FalseEasting
	y += m.FalseNorthing

	return x, y, nil
}

// Inverse transforms from projected coordinates (x, y in meters)
// to geographic coordinates (lon, lat in degrees).
func (m *Mercator) Inverse(x, y float64) (lon, lat float64, err error) {
	// Remove false easting and northing
	x -= m.FalseEasting
	y -= m.FalseNorthing

	a := m.Ellipsoid.A
	centralMeridianRad := m.CentralMeridian * math.Pi / 180.0

	// Calculate longitude
	lonRad := x/a + centralMeridianRad
	lon = lonRad * 180.0 / math.Pi

	if m.Ellipsoid.IsSpherical() {
		// Spherical formula (used by Web Mercator)
		latRad := 2.0*math.Atan(math.Exp(y/a)) - math.Pi/2.0
		lat = latRad * 180.0 / math.Pi
	} else {
		// Ellipsoidal formula - use iterative solution
		// The inverse of: y = a * ln(tan(π/4 + φ/2) * ((1 - e*sin(φ)) / (1 + e*sin(φ)))^(e/2))
		e := math.Sqrt(m.Ellipsoid.Eccentricity)

		// t = exp(y/a)
		t := math.Exp(y / a)

		// Initial approximation using spherical formula
		latRad := 2.0*math.Atan(t) - math.Pi/2.0

		// Iterate to refine latitude using:
		// φ = 2*atan(t * ((1 + e*sin(φ)) / (1 - e*sin(φ)))^(e/2)) - π/2
		const maxIterations = 15
		const tolerance = 1e-12

		for i := 0; i < maxIterations; i++ {
			sinLat := math.Sin(latRad)
			esinLat := e * sinLat

			// Compute new latitude estimate
			latRadNew := 2.0*math.Atan(t*math.Pow((1.0+esinLat)/(1.0-esinLat), e/2.0)) - math.Pi/2.0

			if math.Abs(latRadNew-latRad) < tolerance {
				latRad = latRadNew
				break
			}
			latRad = latRadNew
		}

		lat = latRad * 180.0 / math.Pi
	}

	// Normalize longitude to [-180, 180]
	for lon > 180 {
		lon -= 360
	}
	for lon < -180 {
		lon += 360
	}

	return lon, lat, nil
}
