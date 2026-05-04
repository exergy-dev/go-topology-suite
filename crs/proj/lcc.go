package proj

import "math"

// LambertConformalConic2SP is the two-standard-parallel form of the
// Lambert Conformal Conic projection. EPSG method 9802; Snyder PP1395 §15.
//
// Construct via NewLambertConformalConic2SP — the derived constants
// (n, F, ρ₀) are computed once at construction so Forward/Inverse are
// safe for concurrent use.
type LambertConformalConic2SP struct {
	a, e2          float64
	lon0           float64
	k0             float64 // additional scale factor (PROJ extension; 1.0 for EPSG 9802)
	fe, fn         float64
	n, f, rho0, e1 float64
}

// NewLambertConformalConic2SP builds a Lambert Conformal Conic 2SP
// projection from its EPSG parameters. All angles in radians.
// k_0 defaults to 1.0; use NewLambertConformalConic2SPWithK to set a
// PROJ-style scale factor.
func NewLambertConformalConic2SP(
	a, e2 float64,
	lon0, lat0, lat1, lat2 float64,
	fe, fn float64,
) *LambertConformalConic2SP {
	return NewLambertConformalConic2SPWithK(a, e2, lon0, lat0, lat1, lat2, 1.0, fe, fn)
}

// NewLambertConformalConic2SPWithK builds an LCC 2SP with an explicit
// scale factor k_0 (PROJ extension; not in EPSG 9802 proper).
func NewLambertConformalConic2SPWithK(
	a, e2 float64,
	lon0, lat0, lat1, lat2 float64,
	k0 float64,
	fe, fn float64,
) *LambertConformalConic2SP {
	p := &LambertConformalConic2SP{a: a, e2: e2, lon0: lon0, k0: k0, fe: fe, fn: fn}
	p.e1 = math.Sqrt(e2)
	t1 := tFromLat(lat1, p.e1)
	t2 := tFromLat(lat2, p.e1)
	t0 := tFromLat(lat0, p.e1)
	m1 := math.Cos(lat1) / math.Sqrt(1-e2*math.Sin(lat1)*math.Sin(lat1))
	m2 := math.Cos(lat2) / math.Sqrt(1-e2*math.Sin(lat2)*math.Sin(lat2))
	if math.Abs(lat1-lat2) < 1e-12 {
		p.n = math.Sin(lat1)
	} else {
		p.n = (math.Log(m1) - math.Log(m2)) / (math.Log(t1) - math.Log(t2))
	}
	p.f = m1 / (p.n * math.Pow(t1, p.n))
	p.rho0 = a * p.f * math.Pow(t0, p.n)
	return p
}

// tFromLat computes Snyder's "t" for latitude phi at first eccentricity e.
//
//	t = tan(π/4 - φ/2) · ((1+e·sinφ)/(1-e·sinφ))^(e/2)
//
// For the spherical case (e==0) the eccentricity correction is 1, so t
// reduces to tan(π/4 - φ/2).
func tFromLat(phi, e float64) float64 {
	if e == 0 {
		return math.Tan(piOver4 - phi/2)
	}
	sinPhi := math.Sin(phi)
	return math.Tan(piOver4-phi/2) *
		math.Pow((1+e*sinPhi)/(1-e*sinPhi), e/2)
}

// latFromT inverts tFromLat. Closed-form for the sphere; Newton-on-
// isometric for the ellipsoid (quadratic convergence, never the
// fixed-point oscillation that bites near the equator).
func latFromT(t, e float64) float64 {
	if e == 0 {
		return piOver2 - 2*math.Atan(t)
	}
	// On ellipsoid: ψ = ln(t)·(-1) is the isometric latitude (ψ = -ln t).
	// Equivalently, atanh(sin χ) = -ln t where χ is the conformal
	// latitude. Newton iterate on ψ_geo(φ) = -ln t.
	psiTarget := -math.Log(t)
	e2 := e * e
	phi := piOver2 - 2*math.Atan(t)
	for i := 0; i < 12; i++ {
		sinPhi := math.Sin(phi)
		cosPhi := math.Cos(phi)
		f := math.Atanh(sinPhi) - e*math.Atanh(e*sinPhi) - psiTarget
		fPrime := (1 - e2) / (cosPhi * (1 - e2*sinPhi*sinPhi))
		dphi := -f / fPrime
		phi += dphi
		if math.Abs(dphi) < 1e-14 {
			return phi
		}
	}
	return phi
}

func (p *LambertConformalConic2SP) Name() string { return "Lambert Conformal Conic 2SP" }

func (p *LambertConformalConic2SP) Forward(lonRad, latRad float64) (easting, northing float64) {
	t := tFromLat(latRad, p.e1)
	rho := p.k0 * p.a * p.f * math.Pow(t, p.n)
	theta := p.n * (lonRad - p.lon0)
	easting = p.fe + rho*math.Sin(theta)
	northing = p.fn + p.k0*p.rho0 - rho*math.Cos(theta)
	return
}

func (p *LambertConformalConic2SP) Inverse(easting, northing float64) (lonRad, latRad float64) {
	dE := easting - p.fe
	dN := p.k0*p.rho0 - (northing - p.fn)
	rho := math.Copysign(math.Hypot(dE, dN), p.n)
	var theta float64
	if p.n >= 0 {
		theta = math.Atan2(dE, dN)
	} else {
		theta = math.Atan2(-dE, -dN)
	}
	if rho == 0 {
		latRad = math.Copysign(piOver2, p.n)
		lonRad = p.lon0
		return
	}
	t := math.Pow(rho/(p.k0*p.a*p.f), 1.0/p.n)
	latRad = latFromT(t, p.e1)
	lonRad = theta/p.n + p.lon0
	return
}
