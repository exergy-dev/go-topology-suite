package proj

import "math"

// LambertAzimuthalEqualArea is the ellipsoidal Lambert Azimuthal
// Equal-Area projection (general oblique aspect). EPSG method 9820;
// Snyder PP1395 §24.
//
// Construct via NewLambertAzimuthalEqualArea. Forward/Inverse are pure
// on the constructed value and safe for concurrent use.
type LambertAzimuthalEqualArea struct {
	a, e2        float64
	lon0, lat0   float64
	fe, fn       float64
	qp           float64
	beta0        float64
	rq           float64
	d            float64
	sinB0, cosB0 float64
	polar        int8 // 0 = oblique/equatorial, +1 = north polar, -1 = south polar
}

// NewLambertAzimuthalEqualArea builds a Lambert Azimuthal Equal-Area
// projection. All angles in radians.
func NewLambertAzimuthalEqualArea(
	a, e2 float64,
	lon0, lat0 float64,
	fe, fn float64,
) *LambertAzimuthalEqualArea {
	p := &LambertAzimuthalEqualArea{a: a, e2: e2, lon0: lon0, lat0: lat0, fe: fe, fn: fn}
	const polarThreshold = 1e-10
	switch {
	case math.Abs(lat0-math.Pi/2) < polarThreshold:
		p.polar = +1
	case math.Abs(lat0+math.Pi/2) < polarThreshold:
		p.polar = -1
	}
	// Even in the polar case we still need qp (the authalic-area
	// constant); the rest of the constants are skipped.
	if e2 != 0 {
		e := math.Sqrt(e2)
		p.qp = (1 - e2) * (1/(1-e2) - (1/(2*e))*math.Log((1-e)/(1+e)))
	} else {
		p.qp = 2
	}
	if p.polar != 0 {
		return p
	}
	// Oblique aspect.
	if e2 == 0 {
		p.beta0 = lat0
		p.sinB0, p.cosB0 = math.Sincos(lat0)
		p.rq = a
		p.d = 1
		return p
	}
	q0 := qOfPhi(lat0, e2)
	p.beta0 = math.Asin(q0 / p.qp)
	p.sinB0, p.cosB0 = math.Sincos(p.beta0)
	p.rq = a * math.Sqrt(p.qp/2)
	denom := p.rq * p.cosB0 * math.Sqrt(1-e2*math.Sin(lat0)*math.Sin(lat0))
	p.d = a * math.Cos(lat0) / denom
	return p
}

func (p *LambertAzimuthalEqualArea) Name() string { return "Lambert Azimuthal Equal-Area" }

func (p *LambertAzimuthalEqualArea) Forward(lonRad, latRad float64) (easting, northing float64) {
	if p.polar != 0 {
		return p.forwardPolar(lonRad, latRad)
	}
	q := qOfPhi(latRad, p.e2)
	beta := math.Asin(q / p.qp)
	sinB, cosB := math.Sincos(beta)
	dLon := normaliseLon(lonRad - p.lon0)
	cosL := math.Cos(dLon)
	B := p.rq * math.Sqrt(2/(1+p.sinB0*sinB+p.cosB0*cosB*cosL))
	easting = p.fe + B*p.d*(cosB*math.Sin(dLon))
	northing = p.fn + (B/p.d)*(p.cosB0*sinB-p.sinB0*cosB*cosL)
	return
}

func (p *LambertAzimuthalEqualArea) Inverse(easting, northing float64) (lonRad, latRad float64) {
	if p.polar != 0 {
		return p.inversePolar(easting, northing)
	}
	x := easting - p.fe
	y := northing - p.fn
	rho := math.Sqrt((x/p.d)*(x/p.d) + (p.d*y)*(p.d*y))
	if rho < 1e-12 {
		latRad = p.lat0
		lonRad = p.lon0
		return
	}
	c := 2 * math.Asin(rho/(2*p.rq))
	sinC, cosC := math.Sincos(c)
	beta := math.Asin(cosC*p.sinB0 + (p.d*y*sinC*p.cosB0)/rho)
	var q float64
	if p.e2 != 0 {
		q = p.qp * math.Sin(beta)
	} else {
		q = 2 * math.Sin(beta)
	}
	latRad = phiFromQ(q, p.e2)
	lonRad = p.lon0 + math.Atan2(x*sinC, p.d*rho*p.cosB0*cosC-p.d*p.d*y*p.sinB0*sinC)
	return
}

// forwardPolar / inversePolar implement the polar-aspect special case
// (Snyder PP1395 §24, eq. 24-12 to 24-15). For the north-polar aspect
// (lat_0 ≈ +π/2):
//
//	ρ = a · √(qp - q(φ))
//	E = FE + ρ · sin(λ - λ₀)
//	N = FN - ρ · cos(λ - λ₀)
//
// The south-polar aspect mirrors the sign on q and the cosine term.
func (p *LambertAzimuthalEqualArea) forwardPolar(lonRad, latRad float64) (easting, northing float64) {
	q := qOfPhi(latRad, p.e2)
	var rho float64
	if p.polar == +1 {
		rho = p.a * math.Sqrt(p.qp-q)
	} else {
		rho = p.a * math.Sqrt(p.qp+q)
	}
	dLon := normaliseLon(lonRad - p.lon0)
	easting = p.fe + rho*math.Sin(dLon)
	if p.polar == +1 {
		northing = p.fn - rho*math.Cos(dLon)
	} else {
		northing = p.fn + rho*math.Cos(dLon)
	}
	return
}

func (p *LambertAzimuthalEqualArea) inversePolar(easting, northing float64) (lonRad, latRad float64) {
	x := easting - p.fe
	y := northing - p.fn
	rho := math.Hypot(x, y)
	if rho < 1e-12 {
		latRad = p.lat0
		lonRad = p.lon0
		return
	}
	var q float64
	if p.polar == +1 {
		q = p.qp - (rho/p.a)*(rho/p.a)
		latRad = phiFromQ(q, p.e2)
		lonRad = p.lon0 + math.Atan2(x, -y)
	} else {
		q = (rho/p.a)*(rho/p.a) - p.qp
		latRad = phiFromQ(q, p.e2)
		lonRad = p.lon0 + math.Atan2(x, y)
	}
	return
}
