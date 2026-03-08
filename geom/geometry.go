package geom

// Dimension represents the topological dimension of a geometry.
type Dimension int

const (
	// DimensionEmpty indicates an empty geometry.
	DimensionEmpty Dimension = -1
	// DimensionPoint indicates a 0-dimensional geometry (point).
	DimensionPoint Dimension = 0
	// DimensionLine indicates a 1-dimensional geometry (line).
	DimensionLine Dimension = 1
	// DimensionArea indicates a 2-dimensional geometry (polygon).
	DimensionArea Dimension = 2
)

// Location represents the position of a point relative to a geometry.
type Location int

const (
	// LocationInterior indicates a point in the interior of a geometry.
	LocationInterior Location = 0
	// LocationBoundary indicates a point on the boundary of a geometry.
	LocationBoundary Location = 1
	// LocationExterior indicates a point outside a geometry.
	LocationExterior Location = 2
	// LocationNone indicates no location (for empty geometries).
	LocationNone Location = -1
)

// Geometry is the base interface for all geometric objects.
// It follows the OGC Simple Features Specification.
type Geometry interface {
	// GeometryType returns the type name (e.g., "Point", "LineString").
	GeometryType() string

	// SRID returns the Spatial Reference System ID.
	SRID() int

	// SetSRID sets the Spatial Reference System ID.
	SetSRID(srid int)

	// Envelope returns the bounding box of the geometry.
	Envelope() *Envelope

	// IsEmpty returns true if the geometry has no points.
	IsEmpty() bool

	// IsSimple returns true if the geometry has no self-intersections.
	IsSimple() bool

	// IsValid returns true if the geometry is topologically valid.
	IsValid() bool

	// Dimension returns the topological dimension.
	Dimension() Dimension

	// Boundary returns the boundary of the geometry.
	Boundary() Geometry

	// Coordinates returns all coordinates of the geometry.
	Coordinates() CoordinateSequence

	// NumGeometries returns the number of component geometries.
	// For atomic geometries, returns 1.
	NumGeometries() int

	// GeometryN returns the nth geometry (0-indexed).
	// Returns the geometry itself for atomic types.
	GeometryN(n int) Geometry

	// Clone returns a deep copy of the geometry.
	Clone() Geometry

	// Normalize normalizes the geometry to canonical form.
	Normalize()

	// EqualsExact returns true if the geometries are exactly equal.
	EqualsExact(other Geometry, tolerance float64) bool

	// String returns the WKT representation.
	String() string
}

// Polygonal represents a geometry with area (Polygon or MultiPolygon).
type Polygonal interface {
	Geometry
	Area() float64
}

// Lineal represents a geometry with length (LineString or MultiLineString).
type Lineal interface {
	Geometry
	Length() float64
}

// Puntal represents a point geometry (Point or MultiPoint).
type Puntal interface {
	Geometry
}

// GeometryComponentFilter is a visitor for geometry components.
type GeometryComponentFilter interface {
	Filter(geom Geometry)
}

// CoordinateFilter is a visitor for coordinates.
type CoordinateFilter interface {
	Filter(coord *Coordinate)
}

// CoordinateFilterer applies a coordinate filter to a geometry.
type CoordinateFilterer interface {
	ApplyCoordinateFilter(filter CoordinateFilter)
}

// baseGeometry provides common fields for all geometry implementations.
type baseGeometry struct {
	srid     int
	envelope *Envelope
}

func (b *baseGeometry) SRID() int {
	return b.srid
}

func (b *baseGeometry) SetSRID(srid int) {
	b.srid = srid
}

// invalidateEnvelope clears the cached envelope.
func (b *baseGeometry) invalidateEnvelope() {
	b.envelope = nil
}

// Compare compares two geometries for ordering.
// Returns negative if a < b, zero if equal, positive if a > b.
func Compare(a, b Geometry) int {
	// First compare by type
	typeA := a.GeometryType()
	typeB := b.GeometryType()
	if typeA < typeB {
		return -1
	}
	if typeA > typeB {
		return 1
	}

	// Then compare by coordinates
	coordsA := a.Coordinates()
	coordsB := b.Coordinates()

	for i := 0; i < len(coordsA) && i < len(coordsB); i++ {
		if coordsA[i].X < coordsB[i].X {
			return -1
		}
		if coordsA[i].X > coordsB[i].X {
			return 1
		}
		if coordsA[i].Y < coordsB[i].Y {
			return -1
		}
		if coordsA[i].Y > coordsB[i].Y {
			return 1
		}
	}

	// Compare by length
	if len(coordsA) < len(coordsB) {
		return -1
	}
	if len(coordsA) > len(coordsB) {
		return 1
	}

	return 0
}

