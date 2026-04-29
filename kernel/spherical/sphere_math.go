package spherical

import "math"

// EarthRadius is the IUGG mean Earth radius in metres. The package uses
// this for distance/area conversions; users wanting a different sphere
// can construct a Kernel with a custom radius via NewWithRadius.
const EarthRadius = 6371008.8

// vec3 is a unit vector on the sphere.
type vec3 struct{ X, Y, Z float64 }

func deg2rad(d float64) float64 { return d * math.Pi / 180 }
func rad2deg(r float64) float64 { return r * 180 / math.Pi }

// lonLatToVec converts (lon, lat) in degrees to a 3D unit vector.
func lonLatToVec(lon, lat float64) vec3 {
	la := deg2rad(lat)
	lo := deg2rad(lon)
	cl := math.Cos(la)
	return vec3{X: cl * math.Cos(lo), Y: cl * math.Sin(lo), Z: math.Sin(la)}
}

// vecToLonLat converts a 3D unit vector to (lon, lat) in degrees.
func vecToLonLat(v vec3) (lon, lat float64) {
	lat = rad2deg(math.Asin(clamp(v.Z, -1, 1)))
	lon = rad2deg(math.Atan2(v.Y, v.X))
	return
}

func clamp(v, lo, hi float64) float64 {
	switch {
	case v < lo:
		return lo
	case v > hi:
		return hi
	}
	return v
}

func (a vec3) dot(b vec3) float64 { return a.X*b.X + a.Y*b.Y + a.Z*b.Z }

func (a vec3) cross(b vec3) vec3 {
	return vec3{
		X: a.Y*b.Z - a.Z*b.Y,
		Y: a.Z*b.X - a.X*b.Z,
		Z: a.X*b.Y - a.Y*b.X,
	}
}

func (a vec3) norm() float64 { return math.Sqrt(a.dot(a)) }

func (a vec3) normalize() vec3 {
	n := a.norm()
	if n == 0 {
		return a
	}
	return vec3{a.X / n, a.Y / n, a.Z / n}
}

func (a vec3) neg() vec3 { return vec3{-a.X, -a.Y, -a.Z} }

// haversineCentralAngle returns the central angle between two lon/lat
// points in radians, using the haversine formula (numerically stable for
// small distances).
func haversineCentralAngle(lon1, lat1, lon2, lat2 float64) float64 {
	la1 := deg2rad(lat1)
	la2 := deg2rad(lat2)
	dLat := la2 - la1
	dLon := deg2rad(lon2 - lon1)
	s1 := math.Sin(dLat / 2)
	s2 := math.Sin(dLon / 2)
	a := s1*s1 + math.Cos(la1)*math.Cos(la2)*s2*s2
	return 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

// initialBearingDeg returns the initial great-circle bearing from
// (lon1,lat1) to (lon2,lat2) in degrees, normalised to [0, 360).
func initialBearingDeg(lon1, lat1, lon2, lat2 float64) float64 {
	la1 := deg2rad(lat1)
	la2 := deg2rad(lat2)
	dLon := deg2rad(lon2 - lon1)
	y := math.Sin(dLon) * math.Cos(la2)
	x := math.Cos(la1)*math.Sin(la2) - math.Sin(la1)*math.Cos(la2)*math.Cos(dLon)
	d := rad2deg(math.Atan2(y, x))
	if d < 0 {
		d += 360
	}
	return d
}

// destinationLonLat returns the point at great-circle distance d (metres)
// and bearing (degrees) from (lon, lat) on a sphere of given radius.
func destinationLonLat(lon, lat, bearingDeg, distance, radius float64) (float64, float64) {
	la1 := deg2rad(lat)
	lo1 := deg2rad(lon)
	br := deg2rad(bearingDeg)
	ang := distance / radius

	la2 := math.Asin(math.Sin(la1)*math.Cos(ang) + math.Cos(la1)*math.Sin(ang)*math.Cos(br))
	lo2 := lo1 + math.Atan2(
		math.Sin(br)*math.Sin(ang)*math.Cos(la1),
		math.Cos(ang)-math.Sin(la1)*math.Sin(la2),
	)
	return rad2deg(lo2), rad2deg(la2)
}

// signedAngleBetween returns the signed angle in radians from vector v1
// to v2, in (-π, π]. Sign is determined by the right-hand rule about
// the reference normal n (which should be a unit vector roughly normal
// to the v1-v2 plane).
func signedAngleBetween(v1, v2, n vec3) float64 {
	c := v1.cross(v2)
	sin := c.norm()
	if c.dot(n) < 0 {
		sin = -sin
	}
	cos := v1.dot(v2)
	return math.Atan2(sin, cos)
}
