package geom

import (
	"fmt"
)

// Point represents a single location in coordinate space.
type Point struct {
	baseGeometry
	coord   Coordinate
	isEmpty bool
}

// NewPoint creates a new Point from x and y coordinates.
func NewPoint(x, y float64) *Point {
	return &Point{
		coord:   NewCoordinate(x, y),
		isEmpty: false,
	}
}

// NewPointFromCoordinate creates a new Point from a Coordinate.
func NewPointFromCoordinate(coord Coordinate) *Point {
	return &Point{
		coord:   coord.Clone(),
		isEmpty: false,
	}
}

// NewPointEmpty creates an empty Point.
func NewPointEmpty() *Point {
	return &Point{
		isEmpty: true,
	}
}

// X returns the X coordinate.
func (p *Point) X() float64 {
	return p.coord.X
}

// Y returns the Y coordinate.
func (p *Point) Y() float64 {
	return p.coord.Y
}

// Z returns the Z coordinate (or nil if not present).
func (p *Point) Z() *float64 {
	return p.coord.Z
}

// M returns the M coordinate (or nil if not present).
func (p *Point) M() *float64 {
	return p.coord.M
}

// Coordinate returns the point's coordinate.
func (p *Point) Coordinate() Coordinate {
	return p.coord
}

// GeometryType returns "Point".
func (p *Point) GeometryType() string {
	return "Point"
}

// Envelope returns the bounding box (a point for Point geometries).
func (p *Point) Envelope() *Envelope {
	if p.isEmpty {
		return NewEnvelopeEmpty()
	}
	if p.envelope == nil {
		p.envelope = NewEnvelopeFromCoord(p.coord)
	}
	return p.envelope.Clone()
}

// IsEmpty returns true if this is an empty point.
func (p *Point) IsEmpty() bool {
	return p.isEmpty
}

// IsSimple returns true (points are always simple).
func (p *Point) IsSimple() bool {
	return true
}

// IsValid returns true if the point is valid.
func (p *Point) IsValid() bool {
	return p.isEmpty || !p.coord.IsNaN()
}

// Dimension returns 0 for Point.
func (p *Point) Dimension() Dimension {
	return DimensionPoint
}

// Boundary returns an empty GeometryCollection (points have no boundary).
func (p *Point) Boundary() Geometry {
	return NewGeometryCollectionEmpty()
}

// Coordinates returns the point's coordinate as a sequence.
func (p *Point) Coordinates() CoordinateSequence {
	if p.isEmpty {
		return CoordinateSequence{}
	}
	return CoordinateSequence{p.coord.Clone()}
}

// NumGeometries returns 1 for Point.
func (p *Point) NumGeometries() int {
	return 1
}

// GeometryN returns the point itself (for n=0).
func (p *Point) GeometryN(n int) Geometry {
	if n != 0 {
		return nil
	}
	return p
}

// Clone returns a deep copy of the point.
func (p *Point) Clone() Geometry {
	if p.isEmpty {
		return NewPointEmpty()
	}
	clone := NewPointFromCoordinate(p.coord)
	clone.srid = p.srid
	return clone
}

// Normalize normalizes the point (no-op for points).
func (p *Point) Normalize() {
	// Points don't need normalization
}

// EqualsExact returns true if the points are exactly equal.
func (p *Point) EqualsExact(other Geometry, tolerance float64) bool {
	if other == nil {
		return false
	}
	otherPoint, ok := other.(*Point)
	if !ok {
		return false
	}
	if p.isEmpty && otherPoint.isEmpty {
		return true
	}
	if p.isEmpty || otherPoint.isEmpty {
		return false
	}
	return p.coord.Equals2D(otherPoint.coord, tolerance)
}

// String returns the WKT representation.
func (p *Point) String() string {
	if p.isEmpty {
		return "POINT EMPTY"
	}
	if p.coord.Z != nil && p.coord.M != nil {
		return fmt.Sprintf("POINT ZM (%g %g %g %g)", p.coord.X, p.coord.Y, *p.coord.Z, *p.coord.M)
	}
	if p.coord.Z != nil {
		return fmt.Sprintf("POINT Z (%g %g %g)", p.coord.X, p.coord.Y, *p.coord.Z)
	}
	if p.coord.M != nil {
		return fmt.Sprintf("POINT M (%g %g %g)", p.coord.X, p.coord.Y, *p.coord.M)
	}
	return fmt.Sprintf("POINT (%g %g)", p.coord.X, p.coord.Y)
}

// Distance returns the distance to another point.
func (p *Point) Distance(other *Point) float64 {
	if p.isEmpty || other.isEmpty {
		return 0
	}
	return p.coord.Distance(other.coord)
}

// Equals returns true if the points are equal within the default epsilon.
func (p *Point) Equals(other *Point) bool {
	return p.EqualsExact(other, DefaultEpsilon)
}
