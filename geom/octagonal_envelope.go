package geom

import (
	"math"

	"github.com/exergy-dev/go-topology-suite/crs"
)

// OctagonalEnvelope is a tight bounding container shaped like a general
// octagon: an axis-aligned envelope intersected with a 45°-rotated
// envelope. Tighter than Envelope and cheap to compute, it is useful as
// a quick reject step before exact geometric tests.
//
// The eight extents are: the four axis-aligned bounds (minX, maxX,
// minY, maxY) and the four diagonal bounds in the (A=x+y, B=x-y) basis
// (minA, maxA, minB, maxB). An octagonal envelope can degenerate to any
// shape from a rectangle through a hexagon down to a line or point.
//
// The zero value is the "null" envelope (IsNull reports true). Use
// NewOctagonalEnvelope() to build one explicitly.
//
// Ported from JTS org.locationtech.jts.geom.OctagonalEnvelope.
type OctagonalEnvelope struct {
	minX, maxX float64
	minY, maxY float64
	minA, maxA float64
	minB, maxB float64
	hasValue   bool
}

// NewOctagonalEnvelope returns a new null OctagonalEnvelope.
func NewOctagonalEnvelope() *OctagonalEnvelope {
	return &OctagonalEnvelope{}
}

// NewOctagonalEnvelopeFromGeometry returns the OctagonalEnvelope of g.
func NewOctagonalEnvelopeFromGeometry(g Geometry) *OctagonalEnvelope {
	oe := NewOctagonalEnvelope()
	oe.ExpandToIncludeGeometry(g)
	return oe
}

// IsNull reports whether the envelope contains no points.
func (oe *OctagonalEnvelope) IsNull() bool { return !oe.hasValue }

// MinX/MaxX/MinY/MaxY return the axis-aligned bounds.
func (oe *OctagonalEnvelope) MinX() float64 { return oe.minX }

// MaxX returns the maximum X bound.
func (oe *OctagonalEnvelope) MaxX() float64 { return oe.maxX }

// MinY returns the minimum Y bound.
func (oe *OctagonalEnvelope) MinY() float64 { return oe.minY }

// MaxY returns the maximum Y bound.
func (oe *OctagonalEnvelope) MaxY() float64 { return oe.maxY }

// MinA/MaxA/MinB/MaxB return the rotated (45°) diagonal bounds.
// A = X + Y, B = X - Y.
func (oe *OctagonalEnvelope) MinA() float64 { return oe.minA }

// MaxA returns the maximum diagonal bound A=x+y.
func (oe *OctagonalEnvelope) MaxA() float64 { return oe.maxA }

// MinB returns the minimum diagonal bound B=x-y.
func (oe *OctagonalEnvelope) MinB() float64 { return oe.minB }

// MaxB returns the maximum diagonal bound B=x-y.
func (oe *OctagonalEnvelope) MaxB() float64 { return oe.maxB }

// SetToNull resets the envelope to the empty state.
func (oe *OctagonalEnvelope) SetToNull() { oe.hasValue = false }

// ExpandToInclude expands oe to include the point (x, y).
func (oe *OctagonalEnvelope) ExpandToInclude(x, y float64) *OctagonalEnvelope {
	a := x + y
	b := x - y
	if !oe.hasValue {
		oe.minX, oe.maxX = x, x
		oe.minY, oe.maxY = y, y
		oe.minA, oe.maxA = a, a
		oe.minB, oe.maxB = b, b
		oe.hasValue = true
		return oe
	}
	if x < oe.minX {
		oe.minX = x
	}
	if x > oe.maxX {
		oe.maxX = x
	}
	if y < oe.minY {
		oe.minY = y
	}
	if y > oe.maxY {
		oe.maxY = y
	}
	if a < oe.minA {
		oe.minA = a
	}
	if a > oe.maxA {
		oe.maxA = a
	}
	if b < oe.minB {
		oe.minB = b
	}
	if b > oe.maxB {
		oe.maxB = b
	}
	return oe
}

// ExpandToIncludeXY expands oe to include the point p.
func (oe *OctagonalEnvelope) ExpandToIncludeXY(p XY) *OctagonalEnvelope {
	return oe.ExpandToInclude(p.X, p.Y)
}

// ExpandToIncludeEnvelope expands oe to include all four corners of env.
func (oe *OctagonalEnvelope) ExpandToIncludeEnvelope(env Envelope) *OctagonalEnvelope {
	if env.IsEmpty() {
		return oe
	}
	oe.ExpandToInclude(env.MinX, env.MinY)
	oe.ExpandToInclude(env.MinX, env.MaxY)
	oe.ExpandToInclude(env.MaxX, env.MinY)
	oe.ExpandToInclude(env.MaxX, env.MaxY)
	return oe
}

// ExpandToIncludeOctagonal merges other into oe.
func (oe *OctagonalEnvelope) ExpandToIncludeOctagonal(other *OctagonalEnvelope) *OctagonalEnvelope {
	if other == nil || other.IsNull() {
		return oe
	}
	if !oe.hasValue {
		*oe = *other
		return oe
	}
	if other.minX < oe.minX {
		oe.minX = other.minX
	}
	if other.maxX > oe.maxX {
		oe.maxX = other.maxX
	}
	if other.minY < oe.minY {
		oe.minY = other.minY
	}
	if other.maxY > oe.maxY {
		oe.maxY = other.maxY
	}
	if other.minA < oe.minA {
		oe.minA = other.minA
	}
	if other.maxA > oe.maxA {
		oe.maxA = other.maxA
	}
	if other.minB < oe.minB {
		oe.minB = other.minB
	}
	if other.maxB > oe.maxB {
		oe.maxB = other.maxB
	}
	return oe
}

// ExpandToIncludeGeometry expands oe to include every vertex of g.
func (oe *OctagonalEnvelope) ExpandToIncludeGeometry(g Geometry) *OctagonalEnvelope {
	octVisitVertices(g, func(p XY) { oe.ExpandToInclude(p.X, p.Y) })
	return oe
}

// ExpandBy enlarges oe by distance on every side (axis-aligned by
// distance, diagonal by sqrt(2)*distance). A negative distance can
// collapse the envelope; if so, oe is set to null.
func (oe *OctagonalEnvelope) ExpandBy(distance float64) {
	if !oe.hasValue {
		return
	}
	diag := math.Sqrt2 * distance
	oe.minX -= distance
	oe.maxX += distance
	oe.minY -= distance
	oe.maxY += distance
	oe.minA -= diag
	oe.maxA += diag
	oe.minB -= diag
	oe.maxB += diag
	if !oe.isValid() {
		oe.SetToNull()
	}
}

func (oe *OctagonalEnvelope) isValid() bool {
	if !oe.hasValue {
		return true
	}
	return oe.minX <= oe.maxX && oe.minY <= oe.maxY &&
		oe.minA <= oe.maxA && oe.minB <= oe.maxB
}

// Intersects reports whether oe and other share any point.
func (oe *OctagonalEnvelope) Intersects(other *OctagonalEnvelope) bool {
	if oe.IsNull() || other == nil || other.IsNull() {
		return false
	}
	if oe.minX > other.maxX || oe.maxX < other.minX ||
		oe.minY > other.maxY || oe.maxY < other.minY ||
		oe.minA > other.maxA || oe.maxA < other.minA ||
		oe.minB > other.maxB || oe.maxB < other.minB {
		return false
	}
	return true
}

// IntersectsXY reports whether oe contains the point p.
func (oe *OctagonalEnvelope) IntersectsXY(p XY) bool {
	if oe.IsNull() {
		return false
	}
	if p.X < oe.minX || p.X > oe.maxX || p.Y < oe.minY || p.Y > oe.maxY {
		return false
	}
	a := p.X + p.Y
	b := p.X - p.Y
	if a < oe.minA || a > oe.maxA || b < oe.minB || b > oe.maxB {
		return false
	}
	return true
}

// Contains reports whether other lies entirely within oe (boundary inclusive).
func (oe *OctagonalEnvelope) Contains(other *OctagonalEnvelope) bool {
	if oe.IsNull() || other == nil || other.IsNull() {
		return false
	}
	return other.minX >= oe.minX && other.maxX <= oe.maxX &&
		other.minY >= oe.minY && other.maxY <= oe.maxY &&
		other.minA >= oe.minA && other.maxA <= oe.maxA &&
		other.minB >= oe.minB && other.maxB <= oe.maxB
}

// ToGeometry returns the octagonal envelope as a Polygon (or LineString
// / Point in degenerate cases). The CRS is taken from c (use nil for
// CRS-less). Returns an empty Point when oe is null.
//
// Ported from JTS OctagonalEnvelope.toGeometry.
func (oe *OctagonalEnvelope) ToGeometry(c *crs.CRS) Geometry {
	if oe.IsNull() {
		return NewEmptyPoint(c, LayoutXY)
	}
	px00 := XY{X: oe.minX, Y: oe.minA - oe.minX}
	px01 := XY{X: oe.minX, Y: oe.minX - oe.minB}

	px10 := XY{X: oe.maxX, Y: oe.maxX - oe.maxB}
	px11 := XY{X: oe.maxX, Y: oe.maxA - oe.maxX}

	py00 := XY{X: oe.minA - oe.minY, Y: oe.minY}
	py01 := XY{X: oe.minY + oe.maxB, Y: oe.minY}

	py10 := XY{X: oe.maxY + oe.minB, Y: oe.maxY}
	py11 := XY{X: oe.maxA - oe.maxY, Y: oe.maxY}

	candidates := []XY{px00, px01, py10, py11, px11, px10, py01, py00}
	pts := dedupeAdjacent(candidates)

	switch len(pts) {
	case 1:
		return NewPoint(c, pts[0])
	case 2:
		return NewLineString(c, pts)
	}
	// Polygon: close the ring.
	pts = append(pts, pts[0])
	return NewPolygon(c, pts)
}

// dedupeAdjacent removes consecutive duplicate XYs (matching JTS
// CoordinateList.add(Coordinate, false) semantics).
func dedupeAdjacent(in []XY) []XY {
	out := make([]XY, 0, len(in))
	for _, p := range in {
		if len(out) == 0 || out[len(out)-1] != p {
			out = append(out, p)
		}
	}
	if len(out) > 1 && out[0] == out[len(out)-1] {
		out = out[:len(out)-1]
	}
	return out
}

// octVisitVertices visits every vertex of g (mirrors measure.visitVertices).
func octVisitVertices(g Geometry, fn func(XY)) {
	switch v := g.(type) {
	case *Point:
		if !v.IsEmpty() {
			fn(v.XY())
		}
	case *LineString:
		for i := 0; i < v.NumPoints(); i++ {
			fn(v.PointAt(i))
		}
	case *LinearRing:
		for i := 0; i < v.NumPoints(); i++ {
			fn(v.PointAt(i))
		}
	case *Polygon:
		for r := 0; r < v.NumRings(); r++ {
			for _, p := range v.Ring(r) {
				fn(p)
			}
		}
	case *MultiPoint:
		for i := 0; i < v.NumGeometries(); i++ {
			fn(v.PointAt(i))
		}
	case *MultiLineString:
		for i := 0; i < v.NumGeometries(); i++ {
			octVisitVertices(v.LineStringAt(i), fn)
		}
	case *MultiPolygon:
		for i := 0; i < v.NumGeometries(); i++ {
			octVisitVertices(v.PolygonAt(i), fn)
		}
	case *GeometryCollection:
		for i := 0; i < v.NumGeometries(); i++ {
			octVisitVertices(v.GeometryAt(i), fn)
		}
	}
}
