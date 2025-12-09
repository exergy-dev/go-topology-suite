package projection

import (
	"fmt"
	"math"
)

// TransverseMercator implements the Transverse Mercator projection.
// This projection is used for UTM (Universal Transverse Mercator) zones
// and many national mapping systems. It is conformal and best suited
// for regions that extend north-south.
type TransverseMercator struct {
	Name             string
	CentralMeridian  float64    // Central meridian in degrees
	LatitudeOfOrigin float64    // Latitude of origin in degrees
	ScaleFactor      float64    // Scale factor at central meridian
	FalseEasting     float64    // False easting in meters
	FalseNorthing    float64    // False northing in meters
	Ellipsoid        *Ellipsoid // Reference ellipsoid
}

// NewTransverseMercator creates a new Transverse Mercator projection.
func NewTransverseMercator(ellipsoid *Ellipsoid, centralMeridian, latitudeOfOrigin, scaleFactor, falseEasting, falseNorthing float64) *TransverseMercator {
	return &TransverseMercator{
		Name:             "Transverse Mercator",
		CentralMeridian:  centralMeridian,
		LatitudeOfOrigin: latitudeOfOrigin,
		ScaleFactor:      scaleFactor,
		FalseEasting:     falseEasting,
		FalseNorthing:    falseNorthing,
		Ellipsoid:        ellipsoid,
	}
}

// UTM creates a Transverse Mercator projection configured for a UTM zone.
// Zone numbers range from 1 to 60.
// north indicates whether this is a northern (true) or southern (false) hemisphere zone.
func UTM(zone int, north bool, ellipsoid *Ellipsoid) *TransverseMercator {
	if zone < 1 || zone > 60 {
		zone = 1 // Default to zone 1 if invalid
	}

	if ellipsoid == nil {
		ellipsoid = WGS84()
	}

	// Calculate central meridian: -183 + zone * 6
	centralMeridian := float64(-183 + zone*6)

	// UTM parameters
	scaleFactor := 0.9996
	falseEasting := 500000.0
	falseNorthing := 0.0
	if !north {
		falseNorthing = 10000000.0 // Southern hemisphere
	}

	name := fmt.Sprintf("UTM Zone %d%s", zone, map[bool]string{true: "N", false: "S"}[north])

	return &TransverseMercator{
		Name:             name,
		CentralMeridian:  centralMeridian,
		LatitudeOfOrigin: 0,
		ScaleFactor:      scaleFactor,
		FalseEasting:     falseEasting,
		FalseNorthing:    falseNorthing,
		Ellipsoid:        ellipsoid,
	}
}

// ProjectionName returns the name of this projection.
func (tm *TransverseMercator) ProjectionName() string {
	return tm.Name
}

// Forward transforms from geographic coordinates (lon, lat in degrees)
// to projected coordinates (x, y in meters).
func (tm *TransverseMercator) Forward(lon, lat float64) (x, y float64, err error) {
	// Convert degrees to radians
	lonRad := lon * math.Pi / 180.0
	latRad := lat * math.Pi / 180.0
	centralMeridianRad := tm.CentralMeridian * math.Pi / 180.0
	latOriginRad := tm.LatitudeOfOrigin * math.Pi / 180.0

	// Longitude relative to central meridian
	dLon := lonRad - centralMeridianRad

	// Ellipsoid parameters
	a := tm.Ellipsoid.A
	e2 := tm.Ellipsoid.Eccentricity

	// Calculate frequently used values
	sinLat := math.Sin(latRad)
	cosLat := math.Cos(latRad)
	tanLat := math.Tan(latRad)

	// Calculate N (radius of curvature in prime vertical)
	N := a / math.Sqrt(1.0-e2*sinLat*sinLat)

	// Calculate T, C, A
	T := tanLat * tanLat
	C := e2 / (1.0 - e2) * cosLat * cosLat
	A := dLon * cosLat

	// Calculate M (meridional arc)
	M := tm.meridionalArc(latRad)
	M0 := tm.meridionalArc(latOriginRad)

	// Calculate x and y using series expansion
	k0 := tm.ScaleFactor

	x = k0 * N * (A +
		(1.0-T+C)*A*A*A/6.0 +
		(5.0-18.0*T+T*T+72.0*C-58.0*e2/(1.0-e2))*A*A*A*A*A/120.0)

	y = k0 * (M - M0 +
		N*tanLat*(A*A/2.0 +
			(5.0-T+9.0*C+4.0*C*C)*A*A*A*A/24.0 +
			(61.0-58.0*T+T*T+600.0*C-330.0*e2/(1.0-e2))*A*A*A*A*A*A/720.0))

	// Apply false easting and northing
	x += tm.FalseEasting
	y += tm.FalseNorthing

	return x, y, nil
}

// Inverse transforms from projected coordinates (x, y in meters)
// to geographic coordinates (lon, lat in degrees).
func (tm *TransverseMercator) Inverse(x, y float64) (lon, lat float64, err error) {
	// Remove false easting and northing
	x -= tm.FalseEasting
	y -= tm.FalseNorthing

	// Ellipsoid parameters
	a := tm.Ellipsoid.A
	e2 := tm.Ellipsoid.Eccentricity
	k0 := tm.ScaleFactor

	// Calculate M
	M0 := tm.meridionalArc(tm.LatitudeOfOrigin * math.Pi / 180.0)
	M := M0 + y/k0

	// Calculate footpoint latitude μ (iterative)
	mu := tm.footpointLatitude(M)

	// Calculate e1
	e1 := (1.0 - math.Sqrt(1.0-e2)) / (1.0 + math.Sqrt(1.0-e2))

	// Calculate latitude using series
	phi1 := mu +
		(3.0*e1/2.0-27.0*e1*e1*e1/32.0)*math.Sin(2.0*mu) +
		(21.0*e1*e1/16.0-55.0*e1*e1*e1*e1/32.0)*math.Sin(4.0*mu) +
		(151.0*e1*e1*e1/96.0)*math.Sin(6.0*mu) +
		(1097.0*e1*e1*e1*e1/512.0)*math.Sin(8.0*mu)

	sinPhi1 := math.Sin(phi1)
	cosPhi1 := math.Cos(phi1)
	tanPhi1 := math.Tan(phi1)

	// Calculate N1, T1, C1, R1, D
	N1 := a / math.Sqrt(1.0-e2*sinPhi1*sinPhi1)
	T1 := tanPhi1 * tanPhi1
	C1 := e2 / (1.0 - e2) * cosPhi1 * cosPhi1
	R1 := a * (1.0 - e2) / math.Pow(1.0-e2*sinPhi1*sinPhi1, 1.5)
	D := x / (N1 * k0)

	// Calculate latitude
	latRad := phi1 -
		(N1*tanPhi1/R1)*(D*D/2.0-
			(5.0+3.0*T1+10.0*C1-4.0*C1*C1-9.0*e2/(1.0-e2))*D*D*D*D/24.0+
			(61.0+90.0*T1+298.0*C1+45.0*T1*T1-252.0*e2/(1.0-e2)-3.0*C1*C1)*D*D*D*D*D*D/720.0)

	// Calculate longitude
	lonRad := tm.CentralMeridian*math.Pi/180.0 +
		(D-(1.0+2.0*T1+C1)*D*D*D/6.0+
			(5.0-2.0*C1+28.0*T1-3.0*C1*C1+8.0*e2/(1.0-e2)+24.0*T1*T1)*D*D*D*D*D/120.0)/cosPhi1

	// Convert to degrees
	lat = latRad * 180.0 / math.Pi
	lon = lonRad * 180.0 / math.Pi

	// Normalize longitude to [-180, 180]
	for lon > 180 {
		lon -= 360
	}
	for lon < -180 {
		lon += 360
	}

	return lon, lat, nil
}

// meridionalArc calculates the meridional arc length from the equator
// to the given latitude.
func (tm *TransverseMercator) meridionalArc(lat float64) float64 {
	a := tm.Ellipsoid.A
	e2 := tm.Ellipsoid.Eccentricity

	// Calculate coefficients
	e4 := e2 * e2
	e6 := e4 * e2
	e8 := e6 * e2

	A0 := 1.0 - e2/4.0 - 3.0*e4/64.0 - 5.0*e6/256.0 - 175.0*e8/16384.0
	A2 := 3.0/8.0*(e2 + e4/4.0 + 15.0*e6/128.0 - 455.0*e8/4096.0)
	A4 := 15.0/256.0*(e4 + 3.0*e6/4.0 - 77.0*e8/128.0)
	A6 := 35.0/3072.0*(e6 - 41.0*e8/32.0)
	A8 := -315.0 / 131072.0 * e8

	M := a * (A0*lat -
		A2*math.Sin(2.0*lat) +
		A4*math.Sin(4.0*lat) -
		A6*math.Sin(6.0*lat) +
		A8*math.Sin(8.0*lat))

	return M
}

// footpointLatitude calculates the footpoint latitude from the meridional arc.
func (tm *TransverseMercator) footpointLatitude(M float64) float64 {
	a := tm.Ellipsoid.A
	e2 := tm.Ellipsoid.Eccentricity

	e4 := e2 * e2
	e6 := e4 * e2
	e8 := e6 * e2

	A0 := 1.0 - e2/4.0 - 3.0*e4/64.0 - 5.0*e6/256.0 - 175.0*e8/16384.0

	// Initial approximation
	mu := M / (a * A0)

	return mu
}
