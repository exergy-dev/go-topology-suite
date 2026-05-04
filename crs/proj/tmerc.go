package proj

import (
	"math"

	"github.com/terra-geo/terra/crs"
)

// TransverseMercator is the ellipsoidal Transverse Mercator projection
// implemented via the Krüger n-series (Karney 2011, Engsager-Poder 2007).
// Accuracy is sub-millimetre over the entire globe.
//
// Constructed values are immutable; Forward/Inverse are pure on the
// constructed value and safe for concurrent use.
type TransverseMercator struct {
	a, e2          float64
	lon0, lat0     float64
	k0             float64
	fe, fn         float64
	a0             float64    // rectifying radius × k0
	alpha          [6]float64 // forward series coefficients
	beta           [6]float64 // inverse series coefficients
	xi0            float64    // ξ₀ = rectifying latitude of lat0
}

// NewTransverseMercator builds a TM projection with the given parameters.
// All angles in radians. Coefficients are computed once at construction.
func NewTransverseMercator(a, e2, lon0, lat0, k0, fe, fn float64) *TransverseMercator {
	p := &TransverseMercator{a: a, e2: e2, lon0: lon0, lat0: lat0, k0: k0, fe: fe, fn: fn}

	var n float64
	if e2 != 0 {
		f := 1 - math.Sqrt(1-e2)
		n = f / (2 - f)
	}
	n2 := n * n
	n3 := n2 * n
	n4 := n3 * n
	n5 := n4 * n
	n6 := n5 * n

	// Karney 2011, eq. 14: rectifying radius.
	A := a / (1 + n) * (1 + n2/4 + n4/64 + n6/256)

	// Karney 2011, eq. 35: forward series α_j, n^6 truncation.
	p.alpha[0] = 1.0/2*n - 2.0/3*n2 + 5.0/16*n3 + 41.0/180*n4 - 127.0/288*n5 + 7891.0/37800*n6
	p.alpha[1] = 13.0/48*n2 - 3.0/5*n3 + 557.0/1440*n4 + 281.0/630*n5 - 1983433.0/1935360*n6
	p.alpha[2] = 61.0/240*n3 - 103.0/140*n4 + 15061.0/26880*n5 + 167603.0/181440*n6
	p.alpha[3] = 49561.0/161280*n4 - 179.0/168*n5 + 6601661.0/7257600*n6
	p.alpha[4] = 34729.0/80640*n5 - 3418889.0/1995840*n6
	p.alpha[5] = 212378941.0 / 319334400 * n6

	// Karney 2011, eq. 36: inverse series β_j, n^6 truncation.
	p.beta[0] = 1.0/2*n - 2.0/3*n2 + 37.0/96*n3 - 1.0/360*n4 - 81.0/512*n5 + 96199.0/604800*n6
	p.beta[1] = 1.0/48*n2 + 1.0/15*n3 - 437.0/1440*n4 + 46.0/105*n5 - 1118711.0/3870720*n6
	p.beta[2] = 17.0/480*n3 - 37.0/840*n4 - 209.0/4480*n5 + 5569.0/90720*n6
	p.beta[3] = 4397.0/161280*n4 - 11.0/504*n5 - 830251.0/7257600*n6
	p.beta[4] = 4583.0/161280*n5 - 108847.0/3991680*n6
	p.beta[5] = 20648693.0 / 638668800 * n6

	p.a0 = k0 * A
	if lat0 != 0 && e2 != 0 {
		e := math.Sqrt(e2)
		chi0 := conformalLatitude(lat0, e)
		p.xi0 = chi0 + sumAlphaSin(chi0, p.alpha)
	}
	return p
}

// UTM returns the Transverse Mercator projection for UTM zone 1..60.
// Central meridian: -177° + 6°·(zone-1). K0 = 0.9996. False easting
// 500 000 m; false northing 0 (north) or 10 000 000 m (south).
func UTM(zone int, southern bool, ellipsoid crs.Ellipsoid) *TransverseMercator {
	lon0 := (-177.0 + 6.0*float64(zone-1)) * math.Pi / 180.0
	fn := 0.0
	if southern {
		fn = 10000000.0
	}
	return NewTransverseMercator(ellipsoid.A, ellipsoid.E2(), lon0, 0, 0.9996, 500000.0, fn)
}

func (p *TransverseMercator) Name() string { return "Transverse Mercator" }

func sumAlphaSin(xi float64, alpha [6]float64) float64 {
	s := 0.0
	for j := 0; j < 6; j++ {
		s += alpha[j] * math.Sin(2*float64(j+1)*xi)
	}
	return s
}

// Forward applies the Krüger TM forward series.
func (p *TransverseMercator) Forward(lonRad, latRad float64) (easting, northing float64) {
	dLon := lonRad - p.lon0
	if p.e2 == 0 {
		B := math.Cos(latRad) * math.Sin(dLon)
		easting = p.fe + p.k0*p.a*0.5*math.Log((1+B)/(1-B))
		northing = p.fn + p.k0*p.a*(math.Atan2(math.Tan(latRad), math.Cos(dLon))-p.lat0)
		return
	}
	e := math.Sqrt(p.e2)
	chi := conformalLatitude(latRad, e)

	sinChi, cosChi := math.Sincos(chi)
	sinDL, cosDL := math.Sincos(dLon)
	xi := math.Atan2(sinChi, cosChi*cosDL)
	eta := math.Atanh(cosChi * sinDL)

	xiP := xi
	etaP := eta
	for j := 0; j < 6; j++ {
		twoJ := 2 * float64(j+1)
		s, c := math.Sincos(twoJ * xi)
		sh, ch := math.Sinh(twoJ*eta), math.Cosh(twoJ*eta)
		xiP += p.alpha[j] * s * ch
		etaP += p.alpha[j] * c * sh
	}

	easting = p.fe + p.a0*etaP
	northing = p.fn + p.a0*(xiP-p.xi0)
	return
}

// Inverse applies the Krüger TM inverse series.
func (p *TransverseMercator) Inverse(easting, northing float64) (lonRad, latRad float64) {
	if p.e2 == 0 {
		D := (northing-p.fn)/(p.k0*p.a) + p.lat0
		x := (easting - p.fe) / (p.k0 * p.a)
		latRad = math.Asin(math.Sin(D) / math.Cosh(x))
		lonRad = p.lon0 + math.Atan2(math.Sinh(x), math.Cos(D))
		return
	}
	xiP := (northing-p.fn)/p.a0 + p.xi0
	etaP := (easting - p.fe) / p.a0

	xi := xiP
	eta := etaP
	for j := 0; j < 6; j++ {
		twoJ := 2 * float64(j+1)
		s, c := math.Sincos(twoJ * xiP)
		sh, ch := math.Sinh(twoJ*etaP), math.Cosh(twoJ*etaP)
		xi -= p.beta[j] * s * ch
		eta -= p.beta[j] * c * sh
	}

	chi := math.Asin(math.Sin(xi) / math.Cosh(eta))
	e := math.Sqrt(p.e2)
	latRad = inverseConformalLatitudeKr(chi, e)
	lonRad = p.lon0 + math.Atan2(math.Sinh(eta), math.Cos(xi))
	return
}

// inverseConformalLatitudeKr inverts conformalLatitude by Newton on
// isometric latitude. Quadratic convergence; 4-5 iterations to machine
// precision. Solves f(φ) = atanh(sinφ) - e·atanh(e·sinφ) - ψ_target = 0.
func inverseConformalLatitudeKr(chi, e float64) float64 {
	if e == 0 {
		return chi
	}
	e2 := e * e
	psiTarget := math.Atanh(math.Sin(chi))
	phi := chi
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
