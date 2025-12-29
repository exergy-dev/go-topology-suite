// Package geom provides types and functions for representing
// and manipulating geometric objects in 2D space.
//
// The geometry model follows the OGC Simple Features Specification.
// All geometries implement the Geometry interface which provides
// standard operations like intersection, union, and spatial predicates.
package geom

import (
	"fmt"
	"math"
)

// DefaultEpsilon is the default tolerance for coordinate comparisons.
const DefaultEpsilon = 1e-10

// Coordinate represents a location in 2D space with optional Z and M values.
// X and Y are required; Z (elevation) and M (measure) are optional.
type Coordinate struct {
	X, Y float64
	Z    *float64 // Optional elevation
	M    *float64 // Optional measure value
}

// NewCoordinate creates a new 2D coordinate.
func NewCoordinate(x, y float64) Coordinate {
	return Coordinate{X: x, Y: y}
}

// NewCoordinateZ creates a new 3D coordinate with Z value.
func NewCoordinateZ(x, y, z float64) Coordinate {
	return Coordinate{X: x, Y: y, Z: &z}
}

// NewCoordinateZM creates a new coordinate with Z and M values.
func NewCoordinateZM(x, y, z, m float64) Coordinate {
	return Coordinate{X: x, Y: y, Z: &z, M: &m}
}

// NewCoordinateM creates a new coordinate with M value (no Z).
func NewCoordinateM(x, y, m float64) Coordinate {
	return Coordinate{X: x, Y: y, M: &m}
}

// NewCoordinateNaN creates a coordinate with NaN values, useful for marking
// "no value" in optional coordinate fields.
func NewCoordinateNaN() Coordinate {
	return Coordinate{X: math.NaN(), Y: math.NaN()}
}

// Clone returns a deep copy of the coordinate.
func (c Coordinate) Clone() Coordinate {
	clone := Coordinate{X: c.X, Y: c.Y}
	if c.Z != nil {
		z := *c.Z
		clone.Z = &z
	}
	if c.M != nil {
		m := *c.M
		clone.M = &m
	}
	return clone
}

// String returns a string representation of the coordinate.
func (c Coordinate) String() string {
	if c.Z != nil && c.M != nil {
		return fmt.Sprintf("(%g, %g, %g, %g)", c.X, c.Y, *c.Z, *c.M)
	}
	if c.Z != nil {
		return fmt.Sprintf("(%g, %g, %g)", c.X, c.Y, *c.Z)
	}
	if c.M != nil {
		return fmt.Sprintf("(%g, %g, M=%g)", c.X, c.Y, *c.M)
	}
	return fmt.Sprintf("(%g, %g)", c.X, c.Y)
}

// Equals2D returns true if the X and Y values are equal within epsilon.
func (c Coordinate) Equals2D(other Coordinate, epsilon float64) bool {
	return math.Abs(c.X-other.X) < epsilon && math.Abs(c.Y-other.Y) < epsilon
}

// Equals returns true if all coordinate values are equal within epsilon.
func (c Coordinate) Equals(other Coordinate, epsilon float64) bool {
	if !c.Equals2D(other, epsilon) {
		return false
	}
	// Check Z values
	if (c.Z == nil) != (other.Z == nil) {
		return false
	}
	if c.Z != nil && math.Abs(*c.Z-*other.Z) >= epsilon {
		return false
	}
	// Check M values
	if (c.M == nil) != (other.M == nil) {
		return false
	}
	if c.M != nil && math.Abs(*c.M-*other.M) >= epsilon {
		return false
	}
	return true
}

// Distance returns the 2D Euclidean distance to another coordinate.
func (c Coordinate) Distance(other Coordinate) float64 {
	dx := c.X - other.X
	dy := c.Y - other.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// Distance3D returns the 3D Euclidean distance to another coordinate.
// Returns 2D distance if either coordinate lacks a Z value.
func (c Coordinate) Distance3D(other Coordinate) float64 {
	if c.Z == nil || other.Z == nil {
		return c.Distance(other)
	}
	dx := c.X - other.X
	dy := c.Y - other.Y
	dz := *c.Z - *other.Z
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

// IsNaN returns true if any of the X or Y values are NaN.
func (c Coordinate) IsNaN() bool {
	return math.IsNaN(c.X) || math.IsNaN(c.Y)
}

// HasZ returns true if this coordinate has a Z value.
func (c Coordinate) HasZ() bool {
	return c.Z != nil
}

// HasM returns true if this coordinate has an M value.
func (c Coordinate) HasM() bool {
	return c.M != nil
}

// GetZ returns the Z value or 0 if not set.
func (c Coordinate) GetZ() float64 {
	if c.Z == nil {
		return 0
	}
	return *c.Z
}

// GetM returns the M value or 0 if not set.
func (c Coordinate) GetM() float64 {
	if c.M == nil {
		return 0
	}
	return *c.M
}

// CoordinateSequence is an ordered list of coordinates.
type CoordinateSequence []Coordinate

// NewCoordinateSequence creates a new coordinate sequence from coordinates.
func NewCoordinateSequence(coords ...Coordinate) CoordinateSequence {
	seq := make(CoordinateSequence, len(coords))
	copy(seq, coords)
	return seq
}

// NewCoordinateSequenceXY creates a coordinate sequence from x,y pairs.
// The values slice must have an even number of elements.
func NewCoordinateSequenceXY(values ...float64) CoordinateSequence {
	if len(values)%2 != 0 {
		panic("NewCoordinateSequenceXY requires even number of values")
	}
	seq := make(CoordinateSequence, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		seq[i/2] = NewCoordinate(values[i], values[i+1])
	}
	return seq
}

// Clone returns a deep copy of the coordinate sequence.
func (cs CoordinateSequence) Clone() CoordinateSequence {
	if cs == nil {
		return nil
	}
	clone := make(CoordinateSequence, len(cs))
	for i, c := range cs {
		clone[i] = c.Clone()
	}
	return clone
}

// Len returns the number of coordinates in the sequence.
func (cs CoordinateSequence) Len() int {
	return len(cs)
}

// IsEmpty returns true if the sequence has no coordinates.
func (cs CoordinateSequence) IsEmpty() bool {
	return len(cs) == 0
}

// Get returns the coordinate at the given index.
func (cs CoordinateSequence) Get(index int) Coordinate {
	return cs[index]
}

// First returns the first coordinate in the sequence.
// Panics if the sequence is empty.
func (cs CoordinateSequence) First() Coordinate {
	return cs[0]
}

// Last returns the last coordinate in the sequence.
// Panics if the sequence is empty.
func (cs CoordinateSequence) Last() Coordinate {
	return cs[len(cs)-1]
}

// IsClosed returns true if the first and last coordinates are equal within epsilon.
func (cs CoordinateSequence) IsClosed(epsilon float64) bool {
	if len(cs) < 2 {
		return false
	}
	return cs.First().Equals2D(cs.Last(), epsilon)
}

// Reverse returns a new sequence with coordinates in reverse order.
func (cs CoordinateSequence) Reverse() CoordinateSequence {
	if cs == nil {
		return nil
	}
	reversed := make(CoordinateSequence, len(cs))
	for i, c := range cs {
		reversed[len(cs)-1-i] = c.Clone()
	}
	return reversed
}

// HasZ returns true if any coordinate has a Z value.
func (cs CoordinateSequence) HasZ() bool {
	for _, c := range cs {
		if c.HasZ() {
			return true
		}
	}
	return false
}

// HasM returns true if any coordinate has an M value.
func (cs CoordinateSequence) HasM() bool {
	for _, c := range cs {
		if c.HasM() {
			return true
		}
	}
	return false
}

// Envelope computes the bounding box of the coordinate sequence.
func (cs CoordinateSequence) Envelope() *Envelope {
	if len(cs) == 0 {
		return NewEnvelopeEmpty()
	}

	env := &Envelope{
		MinX: cs[0].X,
		MaxX: cs[0].X,
		MinY: cs[0].Y,
		MaxY: cs[0].Y,
	}

	for i := 1; i < len(cs); i++ {
		env.ExpandToIncludeXY(cs[i].X, cs[i].Y)
	}

	return env
}

// RemoveRepeatedPoints returns a new sequence with consecutive duplicate points removed.
func (cs CoordinateSequence) RemoveRepeatedPoints(epsilon float64) CoordinateSequence {
	if len(cs) <= 1 {
		return cs.Clone()
	}

	result := make(CoordinateSequence, 0, len(cs))
	result = append(result, cs[0].Clone())

	for i := 1; i < len(cs); i++ {
		if !cs[i].Equals2D(cs[i-1], epsilon) {
			result = append(result, cs[i].Clone())
		}
	}

	return result
}

// SubSequence returns a new sequence containing coordinates from start to end (exclusive).
func (cs CoordinateSequence) SubSequence(start, end int) CoordinateSequence {
	if start < 0 {
		start = 0
	}
	if end > len(cs) {
		end = len(cs)
	}
	if start >= end {
		return CoordinateSequence{}
	}
	return cs[start:end].Clone()
}
