package crs

import "math"

// geodeticToGeocentric converts ellipsoidal (lon, lat, h) — radians, radians,
// metres — to geocentric Cartesian (X, Y, Z) on the given ellipsoid.
//
// Closed-form. See EPSG Guidance Note 7-2 §2.2.1.
func geodeticToGeocentric(lonRad, latRad, h float64, e Ellipsoid) (x, y, z float64) {
	sinLat, cosLat := math.Sincos(latRad)
	sinLon, cosLon := math.Sincos(lonRad)
	a := e.A
	e2 := e.E2()
	// Prime vertical radius of curvature.
	N := a / math.Sqrt(1-e2*sinLat*sinLat)
	x = (N + h) * cosLat * cosLon
	y = (N + h) * cosLat * sinLon
	z = (N*(1-e2) + h) * sinLat
	return
}

// geocentricToGeodetic converts (X, Y, Z) to (lon, lat, h) on the given
// ellipsoid using Bowring's 1985 closed-form solution. Accurate to better
// than 10⁻¹¹ rad anywhere on or near the surface; we do a single Newton
// refinement for full machine precision.
func geocentricToGeodetic(x, y, z float64, e Ellipsoid) (lonRad, latRad, h float64) {
	a := e.A
	b := e.B()
	e2 := e.E2()
	ep2 := e.EP2()

	p := math.Hypot(x, y)
	if p == 0 {
		// On the polar axis.
		lonRad = 0
		if z >= 0 {
			latRad = math.Pi / 2
			h = z - b
		} else {
			latRad = -math.Pi / 2
			h = -z - b
		}
		return
	}

	lonRad = math.Atan2(y, x)

	// Bowring's initial parametric latitude.
	theta := math.Atan2(z*a, p*b)
	sinT, cosT := math.Sincos(theta)
	latRad = math.Atan2(z+ep2*b*sinT*sinT*sinT, p-e2*a*cosT*cosT*cosT)

	// One Newton iteration for max precision (overkill but cheap).
	for i := 0; i < 2; i++ {
		sinLat, cosLat := math.Sincos(latRad)
		N := a / math.Sqrt(1-e2*sinLat*sinLat)
		h = p/cosLat - N
		// Newton step using d(lat)/d(...) recurrence; equivalent to
		// re-deriving lat from p, z, and the updated h.
		latRad = math.Atan2(z, p*(1-e2*N/(N+h)))
	}
	return
}

// helmert7 applies a 7-parameter (Bursa-Wolf) transformation to a
// geocentric coordinate. Parameters in p are (dx, dy, dz, rx, ry, rz, ds)
// with dx,dy,dz in metres, rx,ry,rz in arc-seconds, ds in ppm.
//
// PositionVector convention: rotates the position vector. CoordinateFrame
// flips the rotation sign.
func helmert7(x, y, z float64, p [7]float64, conv HelmertConvention) (xo, yo, zo float64) {
	const arcsec2rad = math.Pi / (180.0 * 3600.0)
	const ppm = 1e-6

	dx, dy, dz := p[0], p[1], p[2]
	rx, ry, rz := p[3]*arcsec2rad, p[4]*arcsec2rad, p[5]*arcsec2rad
	s := 1.0 + p[6]*ppm
	if conv == CoordinateFrame {
		rx, ry, rz = -rx, -ry, -rz
	}
	// Small-angle rotation matrix multiplication; retains full accuracy
	// for the parameter ranges that show up in real datum shifts (< 30").
	xo = dx + s*(x-rz*y+ry*z)
	yo = dy + s*(rz*x+y-rx*z)
	zo = dz + s*(-ry*x+rx*y+z)
	return
}

// shiftDatum transforms a geographic coordinate (radians, height in m)
// from src to dst by going through the WGS84 hub. The function is a no-op
// when src == dst by name. Both ToWGS84 vectors are interpreted with their
// respective conventions.
//
// Path: src lat/lon/h → src geocentric → WGS84 geocentric → dst geocentric
// → dst lat/lon/h. Identity ToWGS84 vectors skip the corresponding
// Helmert step entirely.
func shiftDatum(lonRad, latRad, h float64, src, dst Datum) (float64, float64, float64) {
	if src.Name == dst.Name {
		return lonRad, latRad, h
	}
	x, y, z := geodeticToGeocentric(lonRad, latRad, h, src.Ellipsoid)
	if !src.IsIdentityToWGS84() {
		x, y, z = helmert7(x, y, z, src.ToWGS84, src.Convention)
	}
	if !dst.IsIdentityToWGS84() {
		x, y, z = helmert7Inverse(x, y, z, dst.ToWGS84, dst.Convention)
	}
	return geocentricToGeodetic(x, y, z, dst.Ellipsoid)
}

// helmert7Inverse applies the inverse of a 7-parameter Helmert. For the
// small parameter magnitudes seen in real datum shifts, negating the
// translation/rotation/scale and reapplying the forward formula is
// accurate to better than a millimetre — far below the inherent
// uncertainty of published Helmert sets.
func helmert7Inverse(x, y, z float64, p [7]float64, conv HelmertConvention) (float64, float64, float64) {
	var inv [7]float64
	for i := range p {
		inv[i] = -p[i]
	}
	return helmert7(x, y, z, inv, conv)
}
