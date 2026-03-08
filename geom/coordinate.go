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
// When Z or M is absent, the value is math.NaN().
//
// IMPORTANT: Always use constructors (NewCoordinate, NewCoordinateZ, etc.)
// to create Coordinate values. A bare struct literal like Coordinate{X: 1, Y: 2}
// sets Z and M to 0 (not NaN), causing HasZ() and HasM() to return true incorrectly.
// Coordinate must NOT be used as a Go map key because NaN != NaN; use CoordinateXY instead.
type Coordinate struct {
	X, Y float64
	Z    float64 // NaN if absent
	M    float64 // NaN if absent
}

// NewCoordinate creates a new 2D coordinate.
func NewCoordinate(x, y float64) Coordinate {
	return Coordinate{X: x, Y: y, Z: math.NaN(), M: math.NaN()}
}

// NewCoordinateZ creates a new 3D coordinate with Z value.
func NewCoordinateZ(x, y, z float64) Coordinate {
	return Coordinate{X: x, Y: y, Z: z, M: math.NaN()}
}

// NewCoordinateZM creates a new coordinate with Z and M values.
func NewCoordinateZM(x, y, z, m float64) Coordinate {
	return Coordinate{X: x, Y: y, Z: z, M: m}
}

// NewCoordinateM creates a new coordinate with M value (no Z).
func NewCoordinateM(x, y, m float64) Coordinate {
	return Coordinate{X: x, Y: y, Z: math.NaN(), M: m}
}

// NewCoordinateNaN creates a coordinate with NaN values, useful for marking
// "no value" in optional coordinate fields.
func NewCoordinateNaN() Coordinate {
	return Coordinate{X: math.NaN(), Y: math.NaN(), Z: math.NaN(), M: math.NaN()}
}

// Clone returns a copy of the coordinate.
func (c Coordinate) Clone() Coordinate {
	return Coordinate{X: c.X, Y: c.Y, Z: c.Z, M: c.M}
}

// String returns a string representation of the coordinate.
func (c Coordinate) String() string {
	if c.HasZ() && c.HasM() {
		return fmt.Sprintf("(%g, %g, %g, %g)", c.X, c.Y, c.Z, c.M)
	}
	if c.HasZ() {
		return fmt.Sprintf("(%g, %g, %g)", c.X, c.Y, c.Z)
	}
	if c.HasM() {
		return fmt.Sprintf("(%g, %g, M=%g)", c.X, c.Y, c.M)
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
	if c.HasZ() != other.HasZ() {
		return false
	}
	if c.HasZ() && math.Abs(c.Z-other.Z) >= epsilon {
		return false
	}
	// Check M values
	if c.HasM() != other.HasM() {
		return false
	}
	if c.HasM() && math.Abs(c.M-other.M) >= epsilon {
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

// IsNaN returns true if any of the X or Y values are NaN.
func (c Coordinate) IsNaN() bool {
	return math.IsNaN(c.X) || math.IsNaN(c.Y)
}

// CoordinateXY is a 2D coordinate suitable for use as a Go map key.
// Unlike Coordinate, it contains no NaN fields and is safe for == comparison.
type CoordinateXY struct {
	X, Y float64
}

// XY returns a CoordinateXY suitable for use as a Go map key.
// This is necessary because Coordinate contains NaN fields (Z, M)
// and NaN != NaN in IEEE 754, making Coordinate unusable as a map key.
func (c Coordinate) XY() CoordinateXY {
	return CoordinateXY{X: c.X, Y: c.Y}
}

// HasZ returns true if this coordinate has a Z value (not NaN).
func (c Coordinate) HasZ() bool {
	return !math.IsNaN(c.Z)
}

// HasM returns true if this coordinate has an M value (not NaN).
func (c Coordinate) HasM() bool {
	return !math.IsNaN(c.M)
}

// GetZ returns the Z value or 0 if not set.
func (c Coordinate) GetZ() float64 {
	if !c.HasZ() {
		return 0
	}
	return c.Z
}

// GetM returns the M value or 0 if not set.
func (c Coordinate) GetM() float64 {
	if !c.HasM() {
		return 0
	}
	return c.M
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
// Returns an error if the number of values is odd.
func NewCoordinateSequenceXY(values ...float64) (CoordinateSequence, error) {
	if len(values)%2 != 0 {
		return nil, fmt.Errorf("NewCoordinateSequenceXY requires even number of values, got %d", len(values))
	}
	seq := make(CoordinateSequence, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		seq[i/2] = NewCoordinate(values[i], values[i+1])
	}
	return seq, nil
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

