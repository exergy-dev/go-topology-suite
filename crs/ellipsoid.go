package crs

import (
	"fmt"
	"math"
)

// ellipsoid is the default implementation of the Ellipsoid interface.
type ellipsoid struct {
	name              string
	semiMajorAxis     float64
	inverseFlattening float64
	semiMinorAxis     float64
	eccentricity      float64
	eccentricitySq    float64
}

// NewEllipsoid creates a new ellipsoid from semi-major axis and inverse flattening.
// The semi-major axis (a) is the equatorial radius in meters.
// The inverse flattening (1/f) defines how much the ellipsoid deviates from a sphere.
// An inverse flattening of 0 indicates a perfect sphere.
func NewEllipsoid(name string, semiMajorAxis, inverseFlattening float64) (Ellipsoid, error) {
	if semiMajorAxis <= 0 {
		return nil, fmt.Errorf("semi-major axis must be positive, got %f", semiMajorAxis)
	}
	if inverseFlattening < 0 {
		return nil, fmt.Errorf("inverse flattening must be non-negative, got %f", inverseFlattening)
	}

	var semiMinorAxis, eccentricitySq, eccentricity float64

	if inverseFlattening == 0 {
		// Sphere
		semiMinorAxis = semiMajorAxis
		eccentricitySq = 0
		eccentricity = 0
	} else {
		// Calculate flattening: f = 1 / inverseFlattening
		f := 1.0 / inverseFlattening

		// Calculate semi-minor axis: b = a * (1 - f)
		semiMinorAxis = semiMajorAxis * (1.0 - f)

		// Calculate eccentricity squared: e² = (a² - b²) / a²
		// This can be simplified to: e² = 2f - f²
		eccentricitySq = 2.0*f - f*f

		// Calculate eccentricity: e = √(e²)
		eccentricity = math.Sqrt(eccentricitySq)
	}

	return &ellipsoid{
		name:              name,
		semiMajorAxis:     semiMajorAxis,
		inverseFlattening: inverseFlattening,
		semiMinorAxis:     semiMinorAxis,
		eccentricity:      eccentricity,
		eccentricitySq:    eccentricitySq,
	}, nil
}

// NewEllipsoidFromAF creates a new ellipsoid from semi-major axis and flattening.
// The flattening (f) is (a - b) / a where a is the semi-major axis and b is the
// semi-minor axis. This is a convenience function that converts flattening to
// inverse flattening.
func NewEllipsoidFromAF(name string, semiMajorAxis, flattening float64) (Ellipsoid, error) {
	if flattening < 0 || flattening >= 1 {
		return nil, fmt.Errorf("flattening must be in range [0, 1), got %f", flattening)
	}

	var inverseFlattening float64
	if flattening == 0 {
		inverseFlattening = 0
	} else {
		inverseFlattening = 1.0 / flattening
	}

	return NewEllipsoid(name, semiMajorAxis, inverseFlattening)
}

// Name returns the name of the ellipsoid.
func (e *ellipsoid) Name() string {
	return e.name
}

// SemiMajorAxis returns the semi-major axis (equatorial radius) in meters.
func (e *ellipsoid) SemiMajorAxis() float64 {
	return e.semiMajorAxis
}

// InverseFlattening returns the inverse flattening (1/f).
func (e *ellipsoid) InverseFlattening() float64 {
	return e.inverseFlattening
}

// SemiMinorAxis returns the semi-minor axis (polar radius) in meters.
func (e *ellipsoid) SemiMinorAxis() float64 {
	return e.semiMinorAxis
}

// Eccentricity returns the first eccentricity of the ellipsoid.
func (e *ellipsoid) Eccentricity() float64 {
	return e.eccentricity
}

// EccentricitySquared returns the square of the first eccentricity.
func (e *ellipsoid) EccentricitySquared() float64 {
	return e.eccentricitySq
}

// Common ellipsoids used in geodetic datums.
var (
	// WGS84Ellipsoid is the World Geodetic System 1984 ellipsoid.
	// Used by GPS and modern geographic coordinate systems.
	WGS84Ellipsoid Ellipsoid

	// GRS80Ellipsoid is the Geodetic Reference System 1980 ellipsoid.
	// Used by NAD83 and many modern datums. Nearly identical to WGS84.
	GRS80Ellipsoid Ellipsoid

	// Clarke1866Ellipsoid is the Clarke 1866 ellipsoid.
	// Used by NAD27 and many historical North American datums.
	Clarke1866Ellipsoid Ellipsoid

	// Airy1830Ellipsoid is the Airy 1830 ellipsoid.
	// Used by OSGB36 and other British datums.
	Airy1830Ellipsoid Ellipsoid

	// Bessel1841Ellipsoid is the Bessel 1841 ellipsoid.
	// Used by many European and Asian datums.
	Bessel1841Ellipsoid Ellipsoid

	// International1924Ellipsoid is the International 1924 ellipsoid (Hayford).
	// Used by many older datums worldwide.
	International1924Ellipsoid Ellipsoid
)

// initEllipsoids initializes the common ellipsoid instances.
// This is called from init.go.
func initEllipsoids() {
	var err error

	// WGS 84 ellipsoid parameters
	WGS84Ellipsoid, err = NewEllipsoid("WGS 84", 6378137.0, 298.257223563)
	if err != nil {
		panic(fmt.Sprintf("failed to create WGS84 ellipsoid: %v", err))
	}

	// GRS 1980 ellipsoid parameters
	GRS80Ellipsoid, err = NewEllipsoid("GRS 1980", 6378137.0, 298.257222101)
	if err != nil {
		panic(fmt.Sprintf("failed to create GRS80 ellipsoid: %v", err))
	}

	// Clarke 1866 ellipsoid parameters
	Clarke1866Ellipsoid, err = NewEllipsoid("Clarke 1866", 6378206.4, 294.978698214)
	if err != nil {
		panic(fmt.Sprintf("failed to create Clarke1866 ellipsoid: %v", err))
	}

	// Airy 1830 ellipsoid parameters
	Airy1830Ellipsoid, err = NewEllipsoid("Airy 1830", 6377563.396, 299.3249646)
	if err != nil {
		panic(fmt.Sprintf("failed to create Airy1830 ellipsoid: %v", err))
	}

	// Bessel 1841 ellipsoid parameters
	Bessel1841Ellipsoid, err = NewEllipsoid("Bessel 1841", 6377397.155, 299.1528128)
	if err != nil {
		panic(fmt.Sprintf("failed to create Bessel1841 ellipsoid: %v", err))
	}

	// International 1924 ellipsoid parameters
	International1924Ellipsoid, err = NewEllipsoid("International 1924", 6378388.0, 297.0)
	if err != nil {
		panic(fmt.Sprintf("failed to create International1924 ellipsoid: %v", err))
	}
}
