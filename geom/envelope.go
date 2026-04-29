package geom

import "math"

// Envelope is the 2D axis-aligned bounding box used throughout Terra.
// Z and M coordinates are ignored. The zero value is the empty envelope:
// MinX > MaxX signals "no extent yet."
type Envelope struct {
	MinX, MinY, MaxX, MaxY float64
}

// EmptyEnvelope returns the canonical empty envelope: any subsequent Expand
// will replace its bounds with the first inserted coordinate.
func EmptyEnvelope() Envelope {
	return Envelope{
		MinX: math.Inf(+1), MinY: math.Inf(+1),
		MaxX: math.Inf(-1), MaxY: math.Inf(-1),
	}
}

// IsEmpty reports whether e has no extent.
func (e Envelope) IsEmpty() bool { return e.MinX > e.MaxX || e.MinY > e.MaxY }

// Min returns the lower-left corner.
func (e Envelope) Min() XY { return XY{e.MinX, e.MinY} }

// Max returns the upper-right corner.
func (e Envelope) Max() XY { return XY{e.MaxX, e.MaxY} }

// Width returns MaxX-MinX, or 0 if empty.
func (e Envelope) Width() float64 {
	if e.IsEmpty() {
		return 0
	}
	return e.MaxX - e.MinX
}

// Height returns MaxY-MinY, or 0 if empty.
func (e Envelope) Height() float64 {
	if e.IsEmpty() {
		return 0
	}
	return e.MaxY - e.MinY
}

// Area returns Width*Height.
func (e Envelope) Area() float64 { return e.Width() * e.Height() }

// ExpandToIncludeXY returns an envelope that includes p.
// Envelope is a value type, so callers must use the result.
func (e Envelope) ExpandToIncludeXY(p XY) Envelope {
	if e.IsEmpty() {
		return Envelope{p.X, p.Y, p.X, p.Y}
	}
	if p.X < e.MinX {
		e.MinX = p.X
	}
	if p.X > e.MaxX {
		e.MaxX = p.X
	}
	if p.Y < e.MinY {
		e.MinY = p.Y
	}
	if p.Y > e.MaxY {
		e.MaxY = p.Y
	}
	return e
}

// ExpandToInclude merges another envelope into this one.
func (e Envelope) ExpandToInclude(o Envelope) Envelope {
	if o.IsEmpty() {
		return e
	}
	if e.IsEmpty() {
		return o
	}
	if o.MinX < e.MinX {
		e.MinX = o.MinX
	}
	if o.MinY < e.MinY {
		e.MinY = o.MinY
	}
	if o.MaxX > e.MaxX {
		e.MaxX = o.MaxX
	}
	if o.MaxY > e.MaxY {
		e.MaxY = o.MaxY
	}
	return e
}

// Intersects reports whether e and o share at least one point.
// Touching at a corner or edge counts as intersection.
func (e Envelope) Intersects(o Envelope) bool {
	if e.IsEmpty() || o.IsEmpty() {
		return false
	}
	return !(o.MinX > e.MaxX || o.MaxX < e.MinX ||
		o.MinY > e.MaxY || o.MaxY < e.MinY)
}

// ContainsXY reports whether p lies within or on the boundary of e.
func (e Envelope) ContainsXY(p XY) bool {
	if e.IsEmpty() {
		return false
	}
	return p.X >= e.MinX && p.X <= e.MaxX && p.Y >= e.MinY && p.Y <= e.MaxY
}

// Contains reports whether o lies entirely within e (boundary inclusive).
func (e Envelope) Contains(o Envelope) bool {
	if e.IsEmpty() || o.IsEmpty() {
		return false
	}
	return o.MinX >= e.MinX && o.MaxX <= e.MaxX &&
		o.MinY >= e.MinY && o.MaxY <= e.MaxY
}

// envelopeOfFlat builds an envelope from a flat coordinate slice with the
// given stride. It is the routine baseGeom uses to populate its envelope
// cache; exposed at package scope so format decoders can build envelopes
// without going through a Geometry value.
func envelopeOfFlat(coords []float64, stride int) Envelope {
	if len(coords) < stride {
		return EmptyEnvelope()
	}
	minX := coords[0]
	maxX := coords[0]
	minY := coords[1]
	maxY := coords[1]
	for i := stride; i+1 < len(coords); i += stride {
		x, y := coords[i], coords[i+1]
		if x < minX {
			minX = x
		}
		if x > maxX {
			maxX = x
		}
		if y < minY {
			minY = y
		}
		if y > maxY {
			maxY = y
		}
	}
	return Envelope{MinX: minX, MinY: minY, MaxX: maxX, MaxY: maxY}
}
