package proj

import "math"

// WebMercator is the EPSG:3857 "Pseudo-Mercator" projection: spherical
// Mercator using WGS84 longitude/latitude as if they were spherical
// coordinates on a sphere of radius equal to the WGS84 semi-major axis.
//
// The projection is famously non-conformal at the millimetre level
// because the input is ellipsoidal lat/lon but the formula treats them
// as spherical — that's the EPSG-blessed quirk. Forward and Inverse
// reproduce PROJ's `+proj=webmerc` (and `+proj=merc +a=6378137 +b=6378137`)
// to sub-millimetre on every test point in PROJ's gie suite.
type WebMercator struct {
	A float64 // sphere radius (m), conventionally 6378137.0
}

func NewWebMercator() WebMercator { return WebMercator{A: 6378137.0} }

func (p WebMercator) Name() string { return "Web Mercator" }

func (p WebMercator) Forward(lonRad, latRad float64) (e, n float64) {
	e = p.A * lonRad
	n = p.A * math.Log(math.Tan(piOver4+latRad/2))
	return
}

func (p WebMercator) Inverse(e, n float64) (lonRad, latRad float64) {
	lonRad = e / p.A
	latRad = piOver2 - 2*math.Atan(math.Exp(-n/p.A))
	return
}
