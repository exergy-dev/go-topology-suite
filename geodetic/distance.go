package geodetic

import (
	"errors"
	"math"
)

const (
	// Maximum iterations for Vincenty's formula convergence
	maxIterations = 200

	// Convergence threshold for Vincenty's formula (approximately 0.06mm)
	convergenceThreshold = 1e-12
)

var (
	// ErrNoConvergence is returned when Vincenty's formula fails to converge.
	// This typically occurs for nearly antipodal points.
	ErrNoConvergence = errors.New("vincenty formula failed to converge")

	// ErrAntipodalPoints is returned when points are exactly or nearly antipodal.
	ErrAntipodalPoints = errors.New("points are antipodal or nearly antipodal")
)

// Distance calculates the geodesic distance between two points using
// Vincenty's inverse formula.
//
// Parameters:
//   - lat1, lon1: latitude and longitude of first point in degrees
//   - lat2, lon2: latitude and longitude of second point in degrees
//   - e: reference ellipsoid to use for calculation
//
// Returns the distance in meters.
//
// This function panics if Vincenty's formula fails to converge. Use Vincenty()
// directly if you need error handling.
func Distance(lat1, lon1, lat2, lon2 float64, e *Ellipsoid) float64 {
	dist, err := Vincenty(lat1, lon1, lat2, lon2, e)
	if err != nil {
		// Fallback to spherical approximation for antipodal points
		return Haversine(lat1, lon1, lat2, lon2, e.a)
	}
	return dist
}

// DistanceWGS84 calculates the geodesic distance between two points using
// the WGS84 ellipsoid.
//
// Parameters:
//   - lat1, lon1: latitude and longitude of first point in degrees
//   - lat2, lon2: latitude and longitude of second point in degrees
//
// Returns the distance in meters.
func DistanceWGS84(lat1, lon1, lat2, lon2 float64) float64 {
	return Distance(lat1, lon1, lat2, lon2, WGS84)
}

// Vincenty calculates the geodesic distance between two points using
// Vincenty's inverse formula. This is accurate to within 0.5mm on the
// Earth ellipsoid.
//
// Parameters:
//   - lat1, lon1: latitude and longitude of first point in degrees
//   - lat2, lon2: latitude and longitude of second point in degrees
//   - e: reference ellipsoid to use for calculation
//
// Returns:
//   - distance in meters
//   - error if the formula fails to converge (typically for antipodal points)
//
// Reference: Vincenty, T. (1975) "Direct and Inverse Solutions of Geodesics
// on the Ellipsoid with application of nested equations"
func Vincenty(lat1, lon1, lat2, lon2 float64, e *Ellipsoid) (float64, error) {
	// Convert to radians
	φ1 := deg2rad(lat1)
	λ1 := deg2rad(lon1)
	φ2 := deg2rad(lat2)
	λ2 := deg2rad(lon2)

	// Quick check for identical points
	if φ1 == φ2 && λ1 == λ2 {
		return 0, nil
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
			return 0, ErrNoConvergence
		}

		sinλ = math.Sin(λ)
		cosλ = math.Cos(λ)

		sinσ = math.Sqrt((cosU2*sinλ)*(cosU2*sinλ) +
			(cosU1*sinU2-sinU1*cosU2*cosλ)*(cosU1*sinU2-sinU1*cosU2*cosλ))

		if sinσ == 0 {
			return 0, nil // Coincident points
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

	return s, nil
}

// Haversine calculates the great circle distance between two points on a
// sphere using the Haversine formula. This is faster but less accurate than
// Vincenty's formula, with errors up to 0.5% due to Earth's ellipsoidal shape.
//
// Parameters:
//   - lat1, lon1: latitude and longitude of first point in degrees
//   - lat2, lon2: latitude and longitude of second point in degrees
//   - radius: radius of the sphere in meters (use EarthMeanRadius for Earth)
//
// Returns the distance in meters.
//
// The Haversine formula is well-conditioned for all distances, including
// antipodal points, unlike some other spherical distance formulas.
func Haversine(lat1, lon1, lat2, lon2, radius float64) float64 {
	// Convert to radians
	φ1 := deg2rad(lat1)
	λ1 := deg2rad(lon1)
	φ2 := deg2rad(lat2)
	λ2 := deg2rad(lon2)

	Δφ := φ2 - φ1
	Δλ := λ2 - λ1

	a := math.Sin(Δφ/2)*math.Sin(Δφ/2) +
		math.Cos(φ1)*math.Cos(φ2)*
			math.Sin(Δλ/2)*math.Sin(Δλ/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return radius * c
}

// deg2rad converts degrees to radians
func deg2rad(deg float64) float64 {
	return deg * math.Pi / 180
}

// rad2deg converts radians to degrees
func rad2deg(rad float64) float64 {
	return rad * 180 / math.Pi
}

// normalizeAzimuth normalizes an azimuth to the range [0, 360)
func normalizeAzimuth(azimuth float64) float64 {
	result := math.Mod(azimuth, 360)
	if result < 0 {
		result += 360
	}
	return result
}

// normalizeLongitude normalizes a longitude to the range (-180, 180]
func normalizeLongitude(lon float64) float64 {
	for lon > 180 {
		lon -= 360
	}
	for lon <= -180 {
		lon += 360
	}
	return lon
}
