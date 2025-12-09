// Package geodetic provides geodetic calculations on ellipsoidal Earth models.
//
// This package implements geodesic distance, azimuth, and area calculations
// using reference ellipsoids. It supports various standard ellipsoids including
// WGS84, GRS80, and Clarke1866.
//
// All coordinate parameters are expected in degrees (not radians) unless
// explicitly stated otherwise.
package geodetic

import "math"

// EarthMeanRadius is the mean radius of Earth in meters.
// This is the average of the three principal radii.
const EarthMeanRadius = 6371008.8

// EarthAuthalicRadius is the radius of a sphere with the same surface area
// as the WGS84 ellipsoid, in meters.
const EarthAuthalicRadius = 6371007.2

// Ellipsoid represents a reference ellipsoid used for geodetic calculations.
// An ellipsoid is defined by its semi-major axis (equatorial radius) and
// semi-minor axis (polar radius), or equivalently by the semi-major axis
// and flattening.
type Ellipsoid struct {
	name string  // Name of the ellipsoid (e.g., "WGS84")
	a    float64 // Semi-major axis (equatorial radius) in meters
	b    float64 // Semi-minor axis (polar radius) in meters
	f    float64 // Flattening
}

// NewEllipsoid creates a new ellipsoid from semi-major and semi-minor axes.
// Parameters:
//   - name: descriptive name for the ellipsoid
//   - a: semi-major axis (equatorial radius) in meters
//   - b: semi-minor axis (polar radius) in meters
func NewEllipsoid(name string, a, b float64) *Ellipsoid {
	f := (a - b) / a
	return &Ellipsoid{
		name: name,
		a:    a,
		b:    b,
		f:    f,
	}
}

// NewEllipsoidFromAF creates a new ellipsoid from semi-major axis and flattening.
// Parameters:
//   - name: descriptive name for the ellipsoid
//   - a: semi-major axis (equatorial radius) in meters
//   - f: flattening (dimensionless)
func NewEllipsoidFromAF(name string, a, f float64) *Ellipsoid {
	b := a * (1 - f)
	return &Ellipsoid{
		name: name,
		a:    a,
		b:    b,
		f:    f,
	}
}

// NewEllipsoidFromAInvF creates a new ellipsoid from semi-major axis and
// inverse flattening (1/f).
// Parameters:
//   - name: descriptive name for the ellipsoid
//   - a: semi-major axis (equatorial radius) in meters
//   - invF: inverse flattening (1/f, dimensionless)
func NewEllipsoidFromAInvF(name string, a, invF float64) *Ellipsoid {
	f := 1.0 / invF
	b := a * (1 - f)
	return &Ellipsoid{
		name: name,
		a:    a,
		b:    b,
		f:    f,
	}
}

// Name returns the name of the ellipsoid.
func (e *Ellipsoid) Name() string {
	return e.name
}

// SemiMajorAxis returns the semi-major axis (equatorial radius) in meters.
func (e *Ellipsoid) SemiMajorAxis() float64 {
	return e.a
}

// SemiMinorAxis returns the semi-minor axis (polar radius) in meters.
func (e *Ellipsoid) SemiMinorAxis() float64 {
	return e.b
}

// Flattening returns the flattening of the ellipsoid.
// Flattening f = (a - b) / a
func (e *Ellipsoid) Flattening() float64 {
	return e.f
}

// InverseFlattening returns the inverse flattening (1/f) of the ellipsoid.
func (e *Ellipsoid) InverseFlattening() float64 {
	if e.f == 0 {
		return 0 // Sphere case
	}
	return 1.0 / e.f
}

// EccentricitySquared returns the square of the first eccentricity.
// e² = (a² - b²) / a² = 2f - f²
func (e *Ellipsoid) EccentricitySquared() float64 {
	return 2*e.f - e.f*e.f
}

// Eccentricity returns the first eccentricity.
// e = sqrt((a² - b²) / a²)
func (e *Ellipsoid) Eccentricity() float64 {
	return math.Sqrt(e.EccentricitySquared())
}

// SecondEccentricitySquared returns the square of the second eccentricity.
// e'² = (a² - b²) / b²
func (e *Ellipsoid) SecondEccentricitySquared() float64 {
	eSq := e.EccentricitySquared()
	return eSq / (1 - eSq)
}

// Standard ellipsoid definitions

var (
	// WGS84 is the World Geodetic System 1984 ellipsoid.
	// This is the most commonly used ellipsoid for GPS and modern mapping.
	// Parameters: a = 6378137 m, 1/f = 298.257223563
	WGS84 = NewEllipsoidFromAInvF("WGS84", 6378137.0, 298.257223563)

	// GRS80 is the Geodetic Reference System 1980 ellipsoid.
	// Used by NAD83 and many national geodetic systems.
	// Parameters: a = 6378137 m, 1/f = 298.257222101
	GRS80 = NewEllipsoidFromAInvF("GRS80", 6378137.0, 298.257222101)

	// Clarke1866 is the Clarke 1866 ellipsoid.
	// Used by NAD27 and historical US mapping.
	// Parameters: a = 6378206.4 m, b = 6356583.8 m
	Clarke1866 = NewEllipsoid("Clarke1866", 6378206.4, 6356583.8)

	// Sphere is a spherical Earth model with mean radius.
	// Useful for simple calculations where high accuracy is not required.
	// Parameters: a = b = 6371000 m (approximately mean Earth radius)
	Sphere = NewEllipsoid("Sphere", 6371000.0, 6371000.0)
)
