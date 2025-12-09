package geodetic

import "math"

// InitialBearing calculates the initial bearing (forward azimuth) from the
// first point to the second point along the great circle path.
//
// Parameters:
//   - lat1, lon1: latitude and longitude of first point in degrees
//   - lat2, lon2: latitude and longitude of second point in degrees
//
// Returns the initial bearing in degrees (0-360), where:
//   - 0/360 = North
//   - 90 = East
//   - 180 = South
//   - 270 = West
//
// This uses the spherical approximation and is suitable for most applications.
// For high-precision bearings on an ellipsoid, use Inverse().
func InitialBearing(lat1, lon1, lat2, lon2 float64) float64 {
	φ1 := deg2rad(lat1)
	φ2 := deg2rad(lat2)
	Δλ := deg2rad(lon2 - lon1)

	y := math.Sin(Δλ) * math.Cos(φ2)
	x := math.Cos(φ1)*math.Sin(φ2) -
		math.Sin(φ1)*math.Cos(φ2)*math.Cos(Δλ)

	θ := math.Atan2(y, x)

	return normalizeAzimuth(rad2deg(θ))
}

// FinalBearing calculates the final bearing (reverse azimuth) when arriving at
// the second point from the first point along the great circle path.
//
// Parameters:
//   - lat1, lon1: latitude and longitude of first point in degrees
//   - lat2, lon2: latitude and longitude of second point in degrees
//
// Returns the final bearing in degrees (0-360).
//
// The final bearing is the initial bearing from point 2 to point 1, reversed.
func FinalBearing(lat1, lon1, lat2, lon2 float64) float64 {
	// The final bearing is the reverse of the initial bearing from point 2 to point 1
	bearing := InitialBearing(lat2, lon2, lat1, lon1)
	return normalizeAzimuth(bearing + 180)
}

// Inverse solves the inverse geodesic problem: given two points, calculate
// the distance between them and the forward and reverse azimuths.
//
// Parameters:
//   - lat1, lon1: latitude and longitude of first point in degrees
//   - lat2, lon2: latitude and longitude of second point in degrees
//   - e: reference ellipsoid to use for calculation
//
// Returns:
//   - distance: geodesic distance in meters
//   - azimuth1: forward azimuth at first point in degrees (0-360)
//   - azimuth2: forward azimuth at second point in degrees (0-360)
//   - err: error if calculation fails (e.g., for antipodal points)
//
// This implements Vincenty's inverse formula, which is accurate to within
// 0.5mm on the Earth ellipsoid.
//
// Note: azimuth2 is the forward azimuth at the second point, not the back
// azimuth. To get the back azimuth, add 180 degrees to azimuth2.
func Inverse(lat1, lon1, lat2, lon2 float64, e *Ellipsoid) (distance, azimuth1, azimuth2 float64, err error) {
	// Convert to radians
	φ1 := deg2rad(lat1)
	λ1 := deg2rad(lon1)
	φ2 := deg2rad(lat2)
	λ2 := deg2rad(lon2)

	// Quick check for identical points
	if φ1 == φ2 && λ1 == λ2 {
		return 0, 0, 0, nil
	}

	a := e.a
	b := e.b
	f := e.f

	L := λ2 - λ1
	U1 := math.Atan((1 - f) * math.Tan(φ1))
	U2 := math.Atan((1 - f) * math.Tan(φ2))

	sinU1 := math.Sin(U1)
	cosU1 := math.Cos(U1)
	sinU2 := math.Sin(U2)
	cosU2 := math.Cos(U2)

	λ := L
	var λPrev float64
	var sinλ, cosλ float64
	var sinσ, cosσ, σ float64
	var sinα, cos2α float64
	var cos2σm float64

	iterCount := 0
	for {
		iterCount++
		if iterCount > maxIterations {
			return 0, 0, 0, ErrNoConvergence
		}

		sinλ = math.Sin(λ)
		cosλ = math.Cos(λ)

		sinσ = math.Sqrt((cosU2*sinλ)*(cosU2*sinλ) +
			(cosU1*sinU2-sinU1*cosU2*cosλ)*(cosU1*sinU2-sinU1*cosU2*cosλ))

		if sinσ == 0 {
			return 0, 0, 0, nil // Coincident points
		}

		cosσ = sinU1*sinU2 + cosU1*cosU2*cosλ
		σ = math.Atan2(sinσ, cosσ)

		sinα = cosU1 * cosU2 * sinλ / sinσ
		cos2α = 1 - sinα*sinα

		// Handle equatorial line (division by zero)
		if cos2α == 0 {
			cos2σm = 0
		} else {
			cos2σm = cosσ - 2*sinU1*sinU2/cos2α
		}

		C := f / 16 * cos2α * (4 + f*(4-3*cos2α))

		λPrev = λ
		λ = L + (1-C)*f*sinα*
			(σ+C*sinσ*(cos2σm+C*cosσ*(-1+2*cos2σm*cos2σm)))

		// Check for convergence
		if math.Abs(λ-λPrev) < convergenceThreshold {
			break
		}
	}

	uSq := cos2α * (a*a - b*b) / (b * b)
	A := 1 + uSq/16384*(4096+uSq*(-768+uSq*(320-175*uSq)))
	B := uSq / 1024 * (256 + uSq*(-128+uSq*(74-47*uSq)))

	Δσ := B * sinσ * (cos2σm + B/4*(cosσ*(-1+2*cos2σm*cos2σm)-
		B/6*cos2σm*(-3+4*sinσ*sinσ)*(-3+4*cos2σm*cos2σm)))

	s := b * A * (σ - Δσ)

	// Calculate azimuths
	α1 := math.Atan2(cosU2*sinλ, cosU1*sinU2-sinU1*cosU2*cosλ)
	α2 := math.Atan2(cosU1*sinλ, -sinU1*cosU2+cosU1*sinU2*cosλ)

	azimuth1 = normalizeAzimuth(rad2deg(α1))
	azimuth2 = normalizeAzimuth(rad2deg(α2))

	return s, azimuth1, azimuth2, nil
}

// InverseWGS84 solves the inverse geodesic problem using the WGS84 ellipsoid.
//
// Parameters:
//   - lat1, lon1: latitude and longitude of first point in degrees
//   - lat2, lon2: latitude and longitude of second point in degrees
//
// Returns:
//   - distance: geodesic distance in meters
//   - azimuth1: forward azimuth at first point in degrees (0-360)
//   - azimuth2: forward azimuth at second point in degrees (0-360)
//   - err: error if calculation fails
func InverseWGS84(lat1, lon1, lat2, lon2 float64) (distance, azimuth1, azimuth2 float64, err error) {
	return Inverse(lat1, lon1, lat2, lon2, WGS84)
}
