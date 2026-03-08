package geom

import (
	"fmt"
	"math"
)

// Envelope represents a bounding box (axis-aligned rectangle).
// It is defined by minimum and maximum X and Y values.
type Envelope struct {
	MinX, MinY, MaxX, MaxY float64
}

// NewEnvelope creates a new envelope from two coordinates.
func NewEnvelope(x1, y1, x2, y2 float64) *Envelope {
	return &Envelope{
		MinX: math.Min(x1, x2),
		MinY: math.Min(y1, y2),
		MaxX: math.Max(x1, x2),
		MaxY: math.Max(y1, y2),
	}
}

// NewEnvelopeFromCoord creates an envelope from a single coordinate.
func NewEnvelopeFromCoord(c Coordinate) *Envelope {
	return &Envelope{
		MinX: c.X,
		MinY: c.Y,
		MaxX: c.X,
		MaxY: c.Y,
	}
}

// NewEnvelopeFromCoords creates an envelope from two coordinates.
func NewEnvelopeFromCoords(c1, c2 Coordinate) *Envelope {
	return NewEnvelope(c1.X, c1.Y, c2.X, c2.Y)
}

// NewEnvelopeEmpty creates an empty envelope.
func NewEnvelopeEmpty() *Envelope {
	return &Envelope{
		MinX: math.Inf(1),
		MinY: math.Inf(1),
		MaxX: math.Inf(-1),
		MaxY: math.Inf(-1),
	}
}

// Clone returns a copy of the envelope.
func (e *Envelope) Clone() *Envelope {
	if e == nil {
		return nil
	}
	return &Envelope{
		MinX: e.MinX,
		MinY: e.MinY,
		MaxX: e.MaxX,
		MaxY: e.MaxY,
	}
}

// String returns a string representation of the envelope.
func (e *Envelope) String() string {
	if e == nil || e.IsNull() {
		return "Envelope(EMPTY)"
	}
	return fmt.Sprintf("Envelope(%g, %g, %g, %g)", e.MinX, e.MinY, e.MaxX, e.MaxY)
}

// IsNull returns true if this is an empty envelope.
func (e *Envelope) IsNull() bool {
	return e == nil || e.MaxX < e.MinX
}

// Width returns the width of the envelope (MaxX - MinX).
func (e *Envelope) Width() float64 {
	if e.IsNull() {
		return 0
	}
	return e.MaxX - e.MinX
}

// Height returns the height of the envelope (MaxY - MinY).
func (e *Envelope) Height() float64 {
	if e.IsNull() {
		return 0
	}
	return e.MaxY - e.MinY
}

// Area returns the area of the envelope.
func (e *Envelope) Area() float64 {
	return e.Width() * e.Height()
}

// Centre returns the center point of the envelope.
func (e *Envelope) Centre() Coordinate {
	if e.IsNull() {
		return Coordinate{X: math.NaN(), Y: math.NaN()}
	}
	return Coordinate{
		X: (e.MinX + e.MaxX) / 2,
		Y: (e.MinY + e.MaxY) / 2,
	}
}

// ExpandToInclude expands the envelope to include another envelope.
func (e *Envelope) ExpandToInclude(other *Envelope) {
	if other == nil || other.IsNull() {
		return
	}
	if e.IsNull() {
		e.MinX = other.MinX
		e.MinY = other.MinY
		e.MaxX = other.MaxX
		e.MaxY = other.MaxY
	} else {
		e.MinX = math.Min(e.MinX, other.MinX)
		e.MinY = math.Min(e.MinY, other.MinY)
		e.MaxX = math.Max(e.MaxX, other.MaxX)
		e.MaxY = math.Max(e.MaxY, other.MaxY)
	}
}

// ExpandToIncludeCoord expands the envelope to include a coordinate.
func (e *Envelope) ExpandToIncludeCoord(c Coordinate) {
	e.ExpandToIncludeXY(c.X, c.Y)
}

// ExpandToIncludeXY expands the envelope to include a point.
func (e *Envelope) ExpandToIncludeXY(x, y float64) {
	if e.IsNull() {
		e.MinX = x
		e.MaxX = x
		e.MinY = y
		e.MaxY = y
	} else {
		e.MinX = math.Min(e.MinX, x)
		e.MaxX = math.Max(e.MaxX, x)
		e.MinY = math.Min(e.MinY, y)
		e.MaxY = math.Max(e.MaxY, y)
	}
}

// Contains returns true if this envelope contains the given coordinate.
func (e *Envelope) Contains(c Coordinate) bool {
	return e.ContainsXY(c.X, c.Y)
}

// ContainsXY returns true if this envelope contains the given point.
func (e *Envelope) ContainsXY(x, y float64) bool {
	if e.IsNull() {
		return false
	}
	return x >= e.MinX && x <= e.MaxX && y >= e.MinY && y <= e.MaxY
}

// ContainsEnvelope returns true if this envelope completely contains another.
func (e *Envelope) ContainsEnvelope(other *Envelope) bool {
	if e.IsNull() || other.IsNull() {
		return false
	}
	return other.MinX >= e.MinX && other.MaxX <= e.MaxX &&
		other.MinY >= e.MinY && other.MaxY <= e.MaxY
}

// Intersects returns true if this envelope intersects another.
func (e *Envelope) Intersects(other *Envelope) bool {
	if e.IsNull() || other.IsNull() {
		return false
	}
	return !(other.MinX > e.MaxX ||
		other.MaxX < e.MinX ||
		other.MinY > e.MaxY ||
		other.MaxY < e.MinY)
}

// Equals returns true if this envelope equals another within epsilon.
func (e *Envelope) Equals(other *Envelope, epsilon float64) bool {
	if e.IsNull() && other.IsNull() {
		return true
	}
	if e.IsNull() || other.IsNull() {
		return false
	}
	return math.Abs(e.MinX-other.MinX) < epsilon &&
		math.Abs(e.MinY-other.MinY) < epsilon &&
		math.Abs(e.MaxX-other.MaxX) < epsilon &&
		math.Abs(e.MaxY-other.MaxY) < epsilon
}

// Distance returns the distance from this envelope to another.
// Returns 0 if they intersect.
func (e *Envelope) Distance(other *Envelope) float64 {
	if e.Intersects(other) {
		return 0
	}

	var dx, dy float64

	if e.MaxX < other.MinX {
		dx = other.MinX - e.MaxX
	} else if e.MinX > other.MaxX {
		dx = e.MinX - other.MaxX
	}

	if e.MaxY < other.MinY {
		dy = other.MinY - e.MaxY
	} else if e.MinY > other.MaxY {
		dy = e.MinY - other.MaxY
	}

	// Handle edge-only or corner distances
	if dx == 0 {
		return dy
	}
	if dy == 0 {
		return dx
	}
	return math.Sqrt(dx*dx + dy*dy)
}
