// Package crs provides coordinate reference system (CRS) support for the
// Go Topology Suite. It implements types and interfaces for representing
// geographic, projected, and other coordinate reference systems according
// to standards like EPSG and OGC.
//
// A coordinate reference system defines how coordinates relate to positions
// on Earth. This includes the datum (reference frame), ellipsoid (Earth model),
// and coordinate system (axes and units).
package crs

// CRSType represents the type of coordinate reference system.
type CRSType int

const (
	// Geographic CRS uses latitude and longitude coordinates.
	Geographic CRSType = iota
	// Projected CRS uses planar coordinates (e.g., UTM, State Plane).
	Projected
	// Geocentric CRS uses 3D Cartesian coordinates centered at Earth's center.
	Geocentric
	// Vertical CRS measures heights or depths relative to a vertical datum.
	Vertical
	// Compound CRS combines horizontal and vertical CRS.
	Compound
)

// String returns the string representation of a CRSType.
func (t CRSType) String() string {
	switch t {
	case Geographic:
		return "Geographic"
	case Projected:
		return "Projected"
	case Geocentric:
		return "Geocentric"
	case Vertical:
		return "Vertical"
	case Compound:
		return "Compound"
	default:
		return "Unknown"
	}
}

// CRS defines the interface for coordinate reference systems.
// A CRS specifies how coordinates relate to positions in the real world.
type CRS interface {
	// Code returns the authority code (e.g., "EPSG:4326" for WGS 84).
	Code() string

	// Name returns the human-readable name of the CRS.
	Name() string

	// Type returns the type of coordinate reference system.
	Type() CRSType

	// IsGeographic returns true if this is a geographic CRS (lat/lon).
	IsGeographic() bool

	// Datum returns the geodetic datum used by this CRS.
	Datum() Datum

	// CoordinateSystem returns the coordinate system (axes and units).
	CoordinateSystem() CoordinateSystem

	// AreaOfUse returns the geographic area where this CRS is valid.
	// Returns (minLon, minLat, maxLon, maxLat) in degrees.
	AreaOfUse() (minLon, minLat, maxLon, maxLat float64)

	// WKT returns the Well-Known Text representation of this CRS.
	WKT() string
}

// Datum defines the interface for geodetic datums.
// A datum specifies the reference frame for coordinate measurements,
// including the ellipsoid model and its orientation relative to Earth.
type Datum interface {
	// Name returns the name of the datum.
	Name() string

	// Ellipsoid returns the ellipsoid (Earth model) used by this datum.
	Ellipsoid() Ellipsoid

	// PrimeMeridian returns the longitude of the prime meridian in degrees
	// from Greenwich (0 for Greenwich, others for historical datums).
	PrimeMeridian() float64

	// ToWGS84Params returns the 7-parameter Helmert transformation
	// parameters to convert from this datum to WGS84.
	// Returns (dx, dy, dz, rx, ry, rz, ds) where:
	//   - dx, dy, dz: translation in meters
	//   - rx, ry, rz: rotation in arc-seconds
	//   - ds: scale factor in parts per million
	ToWGS84Params() (dx, dy, dz, rx, ry, rz, ds float64)
}

// Ellipsoid defines the interface for reference ellipsoids.
// An ellipsoid is a mathematical model of Earth's shape.
type Ellipsoid interface {
	// Name returns the name of the ellipsoid.
	Name() string

	// SemiMajorAxis returns the semi-major axis (equatorial radius) in meters.
	SemiMajorAxis() float64

	// InverseFlattening returns the inverse flattening (1/f).
	// A value of 0 indicates a sphere.
	InverseFlattening() float64

	// SemiMinorAxis returns the semi-minor axis (polar radius) in meters.
	SemiMinorAxis() float64

	// Eccentricity returns the first eccentricity of the ellipsoid.
	Eccentricity() float64

	// EccentricitySquared returns the square of the first eccentricity.
	EccentricitySquared() float64
}

// CoordinateSystem defines the interface for coordinate systems.
// A coordinate system specifies the axes (directions and units) used
// for coordinates.
type CoordinateSystem interface {
	// Dimension returns the number of dimensions (typically 2 or 3).
	Dimension() int

	// Axis returns the i-th axis (0-indexed).
	Axis(i int) Axis
}

// Axis represents a coordinate system axis with its properties.
type Axis struct {
	// Name is the axis name (e.g., "Longitude", "Easting", "Height").
	Name string

	// Direction is the positive direction of the axis.
	Direction Direction

	// Unit is the unit of measurement for this axis.
	Unit Unit
}

// Direction represents the positive direction of a coordinate axis.
type Direction int

const (
	// North indicates increasing northward.
	North Direction = iota
	// South indicates increasing southward.
	South
	// East indicates increasing eastward.
	East
	// West indicates increasing westward.
	West
	// Up indicates increasing upward (away from Earth center).
	Up
	// Down indicates increasing downward (toward Earth center).
	Down
)

// String returns the string representation of a Direction.
func (d Direction) String() string {
	switch d {
	case North:
		return "North"
	case South:
		return "South"
	case East:
		return "East"
	case West:
		return "West"
	case Up:
		return "Up"
	case Down:
		return "Down"
	default:
		return "Unknown"
	}
}

// coordinateSystem is the default implementation of CoordinateSystem.
type coordinateSystem struct {
	axes []Axis
}

// NewCoordinateSystem creates a new coordinate system with the given axes.
func NewCoordinateSystem(axes []Axis) CoordinateSystem {
	if len(axes) == 0 {
		panic("coordinate system must have at least one axis")
	}
	return &coordinateSystem{axes: axes}
}

// Dimension returns the number of dimensions.
func (cs *coordinateSystem) Dimension() int {
	return len(cs.axes)
}

// Axis returns the i-th axis.
func (cs *coordinateSystem) Axis(i int) Axis {
	if i < 0 || i >= len(cs.axes) {
		panic("axis index out of range")
	}
	return cs.axes[i]
}

// Standard coordinate systems used by common CRS types.
var (
	// EllipsoidalCS2D is the standard 2D geographic coordinate system
	// with longitude and latitude in degrees.
	EllipsoidalCS2D = NewCoordinateSystem([]Axis{
		{Name: "Longitude", Direction: East, Unit: Degree},
		{Name: "Latitude", Direction: North, Unit: Degree},
	})

	// CartesianCS2D is the standard 2D Cartesian coordinate system
	// with easting and northing in meters.
	CartesianCS2D = NewCoordinateSystem([]Axis{
		{Name: "Easting", Direction: East, Unit: Metre},
		{Name: "Northing", Direction: North, Unit: Metre},
	})
)
