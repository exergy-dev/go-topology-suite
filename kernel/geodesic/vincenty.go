package geodesic

import "math"

const (
	vincentyMaxIter = 200
	vincentyTol     = 1e-12
)

// vincentyInverse implements Vincenty's inverse formula. Returns the
// distance s in metres, the initial bearing α1 in radians, and the final
// bearing α2 in radians. ok=false signals non-convergence (typically
// near-antipodal pairs); the caller should fall back to a spherical
// approximation.
func vincentyInverse(lon1, lat1, lon2, lat2 float64) (s, alpha1, alpha2 float64, ok bool) {
	a := SemiMajorA
	b := SemiMinorB
	f := Flattening

	la1 := lat1 * math.Pi / 180
	la2 := lat2 * math.Pi / 180
	L := (lon2 - lon1) * math.Pi / 180

	U1 := math.Atan((1 - f) * math.Tan(la1))
	U2 := math.Atan((1 - f) * math.Tan(la2))
	sinU1, cosU1 := math.Sincos(U1)
	sinU2, cosU2 := math.Sincos(U2)

	lambda := L
	var sinLambda, cosLambda, sinSigma, cosSigma, sigma, sinAlpha, cos2Alpha, cos2SigmaM float64
	for iter := 0; iter < vincentyMaxIter; iter++ {
		sinLambda, cosLambda = math.Sincos(lambda)
		sinSigma = math.Sqrt(
			(cosU2*sinLambda)*(cosU2*sinLambda) +
				(cosU1*sinU2-sinU1*cosU2*cosLambda)*(cosU1*sinU2-sinU1*cosU2*cosLambda),
		)
		if sinSigma == 0 {
			// Coincident points.
			return 0, 0, 0, true
		}
		cosSigma = sinU1*sinU2 + cosU1*cosU2*cosLambda
		sigma = math.Atan2(sinSigma, cosSigma)
		sinAlpha = cosU1 * cosU2 * sinLambda / sinSigma
		cos2Alpha = 1 - sinAlpha*sinAlpha
		if cos2Alpha == 0 {
			cos2SigmaM = 0 // Equatorial line.
		} else {
			cos2SigmaM = cosSigma - 2*sinU1*sinU2/cos2Alpha
		}
		C := f / 16 * cos2Alpha * (4 + f*(4-3*cos2Alpha))
		lambdaPrev := lambda
		lambda = L + (1-C)*f*sinAlpha*(sigma+C*sinSigma*(cos2SigmaM+C*cosSigma*(-1+2*cos2SigmaM*cos2SigmaM)))
		if math.Abs(lambda-lambdaPrev) < vincentyTol {
			ok = true
			break
		}
	}
	if !ok {
		return 0, 0, 0, false
	}

	u2 := cos2Alpha * (a*a - b*b) / (b * b)
	A := 1 + u2/16384*(4096+u2*(-768+u2*(320-175*u2)))
	B := u2 / 1024 * (256 + u2*(-128+u2*(74-47*u2)))
	dSigma := B * sinSigma * (cos2SigmaM + B/4*(cosSigma*(-1+2*cos2SigmaM*cos2SigmaM)-
		B/6*cos2SigmaM*(-3+4*sinSigma*sinSigma)*(-3+4*cos2SigmaM*cos2SigmaM)))

	s = b * A * (sigma - dSigma)
	alpha1 = math.Atan2(cosU2*sinLambda, cosU1*sinU2-sinU1*cosU2*cosLambda)
	alpha2 = math.Atan2(cosU1*sinLambda, -sinU1*cosU2+cosU1*sinU2*cosLambda)
	return s, alpha1, alpha2, true
}

// vincentyDirect computes the destination point given start, initial
// bearing α1 (radians), and distance s (metres) on the WGS84 ellipsoid.
// Returns lon2, lat2 in degrees.
func vincentyDirect(lon1, lat1, alpha1, s float64) (lon2, lat2 float64) {
	a := SemiMajorA
	b := SemiMinorB
	f := Flattening

	la1 := lat1 * math.Pi / 180
	lo1 := lon1 * math.Pi / 180

	sinAlpha1, cosAlpha1 := math.Sincos(alpha1)
	tanU1 := (1 - f) * math.Tan(la1)
	cosU1 := 1 / math.Sqrt(1+tanU1*tanU1)
	sinU1 := tanU1 * cosU1

	sigma1 := math.Atan2(tanU1, cosAlpha1)
	sinAlpha := cosU1 * sinAlpha1
	cos2Alpha := 1 - sinAlpha*sinAlpha
	u2 := cos2Alpha * (a*a - b*b) / (b * b)
	A := 1 + u2/16384*(4096+u2*(-768+u2*(320-175*u2)))
	B := u2 / 1024 * (256 + u2*(-128+u2*(74-47*u2)))

	sigma := s / (b * A)
	var sigmaPrev, sin2SigmaM, sinSigma, cosSigma, dSigma float64
	for iter := 0; iter < vincentyMaxIter; iter++ {
		twoSigmaM := 2*sigma1 + sigma
		sin2SigmaM = math.Sin(twoSigmaM)
		cos2SigmaM := math.Cos(twoSigmaM)
		sinSigma, cosSigma = math.Sincos(sigma)
		dSigma = B * sinSigma * (cos2SigmaM + B/4*(cosSigma*(-1+2*cos2SigmaM*cos2SigmaM)-
			B/6*cos2SigmaM*(-3+4*sinSigma*sinSigma)*(-3+4*cos2SigmaM*cos2SigmaM)))
		sigmaPrev = sigma
		sigma = s/(b*A) + dSigma
		if math.Abs(sigma-sigmaPrev) < vincentyTol {
			break
		}
	}
	_ = sin2SigmaM

	la2 := math.Atan2(
		sinU1*cosSigma+cosU1*sinSigma*cosAlpha1,
		(1-f)*math.Sqrt(sinAlpha*sinAlpha+(sinU1*sinSigma-cosU1*cosSigma*cosAlpha1)*(sinU1*sinSigma-cosU1*cosSigma*cosAlpha1)),
	)
	lambda := math.Atan2(sinSigma*sinAlpha1, cosU1*cosSigma-sinU1*sinSigma*cosAlpha1)
	C := f / 16 * cos2Alpha * (4 + f*(4-3*cos2Alpha))
	L := lambda - (1-C)*f*sinAlpha*(sigma+C*sinSigma*(math.Cos(2*sigma1+sigma)+C*cosSigma*(-1+2*math.Cos(2*sigma1+sigma)*math.Cos(2*sigma1+sigma))))
	lo2 := lo1 + L
	return lo2 * 180 / math.Pi, la2 * 180 / math.Pi
}
