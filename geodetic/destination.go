package geodetic

import "math"

// DestinationPoint calculates the destination point from a starting point,
// given an initial bearing and distance along the geodesic.
//
// Parameters:
//   - lat: latitude of starting point in degrees
//   - lon: longitude of starting point in degrees
//   - bearing: initial bearing in degrees (0-360)
//   - distance: distance to travel in meters
//   - e: reference ellipsoid to use for calculation
//
// Returns:
//   - lat2: latitude of destination point in degrees
//   - lon2: longitude of destination point in degrees
//
// This function uses Vincenty's direct formula. If the calculation fails
// (which is rare), it falls back to a spherical approximation.
func DestinationPoint(lat, lon, bearing, distance float64, e *Ellipsoid) (lat2, lon2 float64) {
	lat2, lon2, _, err := Direct(lat, lon, bearing, distance, e)
	if err != nil {
		// Fallback to spherical approximation
		return destinationPointSpherical(lat, lon, bearing, distance, e.a)
	}
	return lat2, lon2
}

// Direct solves the direct geodesic problem: given a starting point, initial
// azimuth, and distance, calculate the destination point and final azimuth.
//
// Parameters:
//   - lat1: latitude of starting point in degrees
//   - lon1: longitude of starting point in degrees
//   - azimuth1: initial azimuth in degrees (0-360)
//   - distance: distance to travel in meters
//   - e: reference ellipsoid to use for calculation
//
// Returns:
//   - lat2: latitude of destination point in degrees
//   - lon2: longitude of destination point in degrees
//   - azimuth2: final azimuth at destination point in degrees (0-360)
//   - err: error if calculation fails
//
// This implements Vincenty's direct formula, which is accurate to within
// 0.5mm on the Earth ellipsoid.
//
// Reference: Vincenty, T. (1975) "Direct and Inverse Solutions of Geodesics
// on the Ellipsoid with application of nested equations"
func Direct(lat1, lon1, azimuth1, distance float64, e *Ellipsoid) (lat2, lon2, azimuth2 float64, err error) {
	// Convert to radians
	φ1 := deg2rad(lat1)
	λ1 := deg2rad(lon1)
	α1 := deg2rad(azimuth1)
	s := distance

	a := e.a
	b := e.b
	f := e.f

	sinα1 := math.Sin(α1)
	cosα1 := math.Cos(α1)

	tanU1 := (1 - f) * math.Tan(φ1)
	cosU1 := 1 / math.Sqrt(1+tanU1*tanU1)
	sinU1 := tanU1 * cosU1

	σ1 := math.Atan2(tanU1, cosα1)
	sinα := cosU1 * sinα1
	cos2α := 1 - sinα*sinα

	uSq := cos2α * (a*a - b*b) / (b * b)
	A := 1 + uSq/16384*(4096+uSq*(-768+uSq*(320-175*uSq)))
	B := uSq / 1024 * (256 + uSq*(-128+uSq*(74-47*uSq)))

	σ := s / (b * A)
	var σPrev float64
	var sinσ, cosσ float64
	var cos2σm float64
	var Δσ float64

	iterCount := 0
	for {
		iterCount++
		if iterCount > maxIterations {
			return 0, 0, 0, ErrNoConvergence
		}

		cos2σm = math.Cos(2*σ1 + σ)
		sinσ = math.Sin(σ)
		cosσ = math.Cos(σ)

		Δσ = B * sinσ * (cos2σm + B/4*(cosσ*(-1+2*cos2σm*cos2σm)-
			B/6*cos2σm*(-3+4*sinσ*sinσ)*(-3+4*cos2σm*cos2σm)))

		σPrev = σ
		σ = s/(b*A) + Δσ

		// Check for convergence
		if math.Abs(σ-σPrev) < convergenceThreshold {
			break
		}
	}

	tmp := sinU1*sinσ - cosU1*cosσ*cosα1
	φ2 := math.Atan2(sinU1*cosσ+cosU1*sinσ*cosα1,
		(1-f)*math.Sqrt(sinα*sinα+tmp*tmp))

	λ := math.Atan2(sinσ*sinα1, cosU1*cosσ-sinU1*sinσ*cosα1)
	C := f / 16 * cos2α * (4 + f*(4-3*cos2α))
	L := λ - (1-C)*f*sinα*
		(σ+C*sinσ*(cos2σm+C*cosσ*(-1+2*cos2σm*cos2σm)))

	λ2 := λ1 + L

	α2 := math.Atan2(sinα, -tmp)

	lat2 = rad2deg(φ2)
	lon2 = normalizeLongitude(rad2deg(λ2))
	azimuth2 = normalizeAzimuth(rad2deg(α2))

	return lat2, lon2, azimuth2, nil
}

// destinationPointSpherical calculates the destination point using spherical
// Earth approximation. Used as a fallback when Vincenty fails.
func destinationPointSpherical(lat, lon, bearing, distance, radius float64) (lat2, lon2 float64) {
	φ1 := deg2rad(lat)
	λ1 := deg2rad(lon)
	θ := deg2rad(bearing)
	δ := distance / radius // angular distance

	φ2 := math.Asin(math.Sin(φ1)*math.Cos(δ) +
		math.Cos(φ1)*math.Sin(δ)*math.Cos(θ))

	λ2 := λ1 + math.Atan2(math.Sin(θ)*math.Sin(δ)*math.Cos(φ1),
		math.Cos(δ)-math.Sin(φ1)*math.Sin(φ2))

	lat2 = rad2deg(φ2)
	lon2 = normalizeLongitude(rad2deg(λ2))

	return lat2, lon2
}
