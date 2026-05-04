package proj

import "math"

// AlbersEqualAreaConic is the two-standard-parallel Albers Equal-Area
// Conic projection. EPSG method 9822; Snyder PP1395 §14.
//
// Construct via NewAlbersEqualAreaConic. Forward/Inverse are pure on the
// constructed value and safe for concurrent use.
type AlbersEqualAreaConic struct {
	a, e2          float64
	lon0           float64
	fe, fn         float64
	n, c, rho0, e1 float64
}

// NewAlbersEqualAreaConic builds an Albers Equal-Area Conic projection.
// All angles in radians.
func NewAlbersEqualAreaConic(
	a, e2 float64,
	lon0, lat0, lat1, lat2 float64,
	fe, fn float64,
) *AlbersEqualAreaConic {
	p := &AlbersEqualAreaConic{a: a, e2: e2, lon0: lon0, fe: fe, fn: fn}
	p.e1 = math.Sqrt(e2)
	m1sq := math.Cos(lat1) * math.Cos(lat1) / (1 - e2*math.Sin(lat1)*math.Sin(lat1))
	m2sq := math.Cos(lat2) * math.Cos(lat2) / (1 - e2*math.Sin(lat2)*math.Sin(lat2))
	q1 := qOfPhi(lat1, e2)
	q2 := qOfPhi(lat2, e2)
	q0 := qOfPhi(lat0, e2)
	if math.Abs(lat1-lat2) < 1e-12 {
		p.n = math.Sin(lat1)
	} else {
		p.n = (m1sq - m2sq) / (q2 - q1)
	}
	p.c = m1sq + p.n*q1
	p.rho0 = a * math.Sqrt(p.c-p.n*q0) / p.n
	return p
}

// qOfPhi is Snyder's "authalic" factor (PP1395 (3-12)):
//
//	q(φ) = (1-e²) · ( sinφ/(1-e²·sin²φ) - (1/(2e))·ln((1-e·sinφ)/(1+e·sinφ)) )
func qOfPhi(phi, e2 float64) float64 {
	if e2 == 0 {
		return 2 * math.Sin(phi)
	}
	e := math.Sqrt(e2)
	sinPhi := math.Sin(phi)
	return (1 - e2) * (sinPhi/(1-e2*sinPhi*sinPhi) -
		(1/(2*e))*math.Log((1-e*sinPhi)/(1+e*sinPhi)))
}

// phiFromQ inverts qOfPhi by Newton iteration. Snyder PP1395 (3-16).
func phiFromQ(q, e2 float64) float64 {
	if e2 == 0 {
		return math.Asin(q / 2)
	}
	e := math.Sqrt(e2)
	phi := math.Asin(q / 2)
	for i := 0; i < 16; i++ {
		sinPhi := math.Sin(phi)
		denom := 1 - e2*sinPhi*sinPhi
		dphi := (denom * denom / (2 * math.Cos(phi))) *
			(q/(1-e2) -
				sinPhi/denom +
				(1/(2*e))*math.Log((1-e*sinPhi)/(1+e*sinPhi)))
		phi += dphi
		if math.Abs(dphi) < 1e-13 {
			return phi
		}
	}
	return phi
}

func (p *AlbersEqualAreaConic) Name() string { return "Albers Equal-Area Conic" }

func (p *AlbersEqualAreaConic) Forward(lonRad, latRad float64) (easting, northing float64) {
	q := qOfPhi(latRad, p.e2)
	rho := p.a * math.Sqrt(p.c-p.n*q) / p.n
	theta := p.n * (lonRad - p.lon0)
	easting = p.fe + rho*math.Sin(theta)
	northing = p.fn + p.rho0 - rho*math.Cos(theta)
	return
}

func (p *AlbersEqualAreaConic) Inverse(easting, northing float64) (lonRad, latRad float64) {
	dE := easting - p.fe
	dN := p.rho0 - (northing - p.fn)
	rho := math.Copysign(math.Hypot(dE, dN), p.n)
	var theta float64
	if p.n >= 0 {
		theta = math.Atan2(dE, dN)
	} else {
		theta = math.Atan2(-dE, -dN)
	}
	q := (p.c - (rho*p.n/p.a)*(rho*p.n/p.a)) / p.n
	latRad = phiFromQ(q, p.e2)
	lonRad = theta/p.n + p.lon0
	return
}
