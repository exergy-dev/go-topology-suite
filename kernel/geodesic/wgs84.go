package geodesic

import "math"

// WGS84 ellipsoid parameters.
const (
	// SemiMajorA is the equatorial radius in metres.
	SemiMajorA = 6378137.0
	// Flattening is the standard WGS84 flattening.
	Flattening = 1.0 / 298.257223563
)

// SemiMinorB is the polar radius in metres.
var SemiMinorB = SemiMajorA * (1 - Flattening)

// AuthalicRadius is the radius of the sphere with the same surface area
// as the WGS84 ellipsoid (~6371007.18 m). Used for ring-area on the
// ellipsoid as a high-quality approximation; full geodesic-polygon area
// (Karney) is a Phase 4 follow-up.
var AuthalicRadius = computeAuthalicRadius()

func computeAuthalicRadius() float64 {
	a := SemiMajorA
	b := SemiMinorB
	e2 := 1 - (b*b)/(a*a)
	e := math.Sqrt(e2)
	// Surface area of oblate spheroid:
	//   A = 2π·a² · (1 + ((1-e²)/(2e))·ln((1+e)/(1-e)))
	// Authalic radius R_q = √(A/(4π)) = a · √(q/2) with
	//   q = 1 + ((1-e²)/(2e))·ln((1+e)/(1-e))
	q := 1 + ((1-e2)/(2*e))*math.Log((1+e)/(1-e))
	return a * math.Sqrt(q/2)
}
