package proj

import "math"

// Constants reused across projections.
const (
	piOver2 = math.Pi / 2
	piOver4 = math.Pi / 4
)

// conformalLatitude returns the conformal latitude χ on the ellipsoid:
//
//	χ = 2·atan( tan(π/4 + φ/2) · ((1 - e·sin φ)/(1 + e·sin φ))^(e/2) ) - π/2
//
// Used by Mercator and Transverse Mercator. e is first eccentricity (not
// squared).
//
// It is evaluated in the algebraically identical isometric-latitude form
//
//	ψ = atanh(sin φ) - e·atanh(e·sin φ),  χ = atan(sinh ψ)
//
// (the same ψ that inverseConformalLatitudeKr solves for). The tan(π/4 +
// φ/2) form cancels catastrophically near the equator: at φ = 0 it
// returned ±ulp noise depending on how math.Tan(π/4) rounds (1 on amd64,
// 1-2⁻⁵² on arm64 where tan's argument reduction is FMA-fused), giving
// UTM northings of -1.4e-9 m instead of 0. This form is odd in φ and
// exact (+0) at φ = 0.
func conformalLatitude(phi, e float64) float64 {
	if e == 0 {
		return phi
	}
	sinPhi := math.Sin(phi)
	psi := math.Atanh(sinPhi) - e*math.Atanh(e*sinPhi)
	return math.Atan(math.Sinh(psi))
}

// normaliseLon wraps a longitude in radians to [-π, π].
func normaliseLon(lon float64) float64 {
	for lon > math.Pi {
		lon -= 2 * math.Pi
	}
	for lon < -math.Pi {
		lon += 2 * math.Pi
	}
	return lon
}
