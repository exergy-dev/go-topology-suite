// Package projection provides map projection transformations
// for converting between geographic coordinates (longitude, latitude)
// and projected coordinates (x, y in meters).
package projection

// Projection defines an interface for map projections.
// Map projections transform coordinates between geographic (lon/lat in degrees)
// and projected (x/y in meters) coordinate systems.
type Projection interface {
	// Forward transforms from geographic to projected coordinates.
	// Input: longitude and latitude in degrees.
	// Output: x and y coordinates in meters (or the projection's units).
	Forward(lon, lat float64) (x, y float64, err error)

	// Inverse transforms from projected to geographic coordinates.
	// Input: x and y coordinates in meters (or the projection's units).
	// Output: longitude and latitude in degrees.
	Inverse(x, y float64) (lon, lat float64, err error)

	// Name returns the projection name or identifier.
	Name() string
}

// Ellipsoid represents an ellipsoid model of the Earth.
type Ellipsoid struct {
	Name          string  // Name of the ellipsoid (e.g., "WGS84")
	A             float64 // Semi-major axis (equatorial radius) in meters
	B             float64 // Semi-minor axis (polar radius) in meters
	F             float64 // Flattening = (a - b) / a
	Eccentricity  float64 // First eccentricity
	Eccentricity2 float64 // Second eccentricity
}

// NewEllipsoid creates an ellipsoid from semi-major axis and inverse flattening.
func NewEllipsoid(name string, a, inverseFlattening float64) *Ellipsoid {
	f := 1.0 / inverseFlattening
	b := a * (1.0 - f)
	e := (a*a - b*b) / (a * a) // e^2 = (a^2 - b^2) / a^2
	e2 := (a*a - b*b) / (b * b) // e'^2 = (a^2 - b^2) / b^2

	return &Ellipsoid{
		Name:          name,
		A:             a,
		B:             b,
		F:             f,
		Eccentricity:  e,
		Eccentricity2: e2,
	}
}

// WGS84 returns the WGS84 ellipsoid (used by GPS).
func WGS84() *Ellipsoid {
	return NewEllipsoid("WGS84", 6378137.0, 298.257223563)
}

// GRS80 returns the GRS 1980 ellipsoid.
func GRS80() *Ellipsoid {
	return NewEllipsoid("GRS80", 6378137.0, 298.257222101)
}

// Clarke1866 returns the Clarke 1866 ellipsoid (used in NAD27).
func Clarke1866() *Ellipsoid {
	return NewEllipsoid("Clarke1866", 6378206.4, 294.978698214)
}

// Sphere returns a spherical model with the given radius.
// If radius is 0, uses the WGS84 semi-major axis.
func Sphere(radius float64) *Ellipsoid {
	if radius == 0 {
		radius = 6378137.0 // WGS84 semi-major axis
	}
	return &Ellipsoid{
		Name:          "Sphere",
		A:             radius,
		B:             radius,
		F:             0,
		Eccentricity:  0,
		Eccentricity2: 0,
	}
}

// IsSpherical returns true if this is a spherical model (a == b).
func (e *Ellipsoid) IsSpherical() bool {
	return e.A == e.B
}
