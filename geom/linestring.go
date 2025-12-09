package geom

import (
	"fmt"
	"math"
	"strings"
)

// LineString represents a sequence of connected line segments.
type LineString struct {
	baseGeometry
	coords CoordinateSequence
}

// NewLineString creates a new LineString from a coordinate sequence.
func NewLineString(coords CoordinateSequence) *LineString {
	return &LineString{
		coords: coords.Clone(),
	}
}

// NewLineStringXY creates a new LineString from x,y pairs.
func NewLineStringXY(values ...float64) *LineString {
	return NewLineString(NewCoordinateSequenceXY(values...))
}

// NewLineStringEmpty creates an empty LineString.
func NewLineStringEmpty() *LineString {
	return &LineString{
		coords: CoordinateSequence{},
	}
}

// GeometryType returns "LineString".
func (ls *LineString) GeometryType() string {
	return "LineString"
}

// Envelope returns the bounding box.
func (ls *LineString) Envelope() *Envelope {
	if ls.envelope == nil {
		ls.envelope = ls.coords.Envelope()
	}
	return ls.envelope.Clone()
}

// IsEmpty returns true if the linestring has no coordinates.
func (ls *LineString) IsEmpty() bool {
	return len(ls.coords) == 0
}

// IsSimple returns true if the linestring has no self-intersections.
func (ls *LineString) IsSimple() bool {
	if len(ls.coords) <= 3 {
		return true
	}
	// Simple check: no segment intersections except at consecutive endpoints
	// Full implementation would check all segment pairs
	return ls.checkSimple()
}

func (ls *LineString) checkSimple() bool {
	n := len(ls.coords)
	if n <= 3 {
		return true
	}

	for i := 0; i < n-1; i++ {
		for j := i + 2; j < n-1; j++ {
			// Don't check consecutive segments
			if i == 0 && j == n-2 && ls.IsClosed() {
				continue
			}
			if ls.segmentsIntersect(i, j) {
				return false
			}
		}
	}
	return true
}

func (ls *LineString) segmentsIntersect(i, j int) bool {
	p1 := ls.coords[i]
	p2 := ls.coords[i+1]
	p3 := ls.coords[j]
	p4 := ls.coords[j+1]

	// Check if segments intersect in their interiors
	d1 := direction(p3, p4, p1)
	d2 := direction(p3, p4, p2)
	d3 := direction(p1, p2, p3)
	d4 := direction(p1, p2, p4)

	if ((d1 > 0 && d2 < 0) || (d1 < 0 && d2 > 0)) &&
		((d3 > 0 && d4 < 0) || (d3 < 0 && d4 > 0)) {
		return true
	}

	return false
}

func direction(p1, p2, p3 Coordinate) float64 {
	return (p3.X-p1.X)*(p2.Y-p1.Y) - (p2.X-p1.X)*(p3.Y-p1.Y)
}

// IsValid returns true if the linestring is valid.
// A valid linestring has either 0 or >= 2 points.
func (ls *LineString) IsValid() bool {
	return len(ls.coords) == 0 || len(ls.coords) >= 2
}

// Dimension returns 1 for LineString.
func (ls *LineString) Dimension() Dimension {
	return DimensionLine
}

// Boundary returns the boundary (endpoints for non-closed, empty for closed).
func (ls *LineString) Boundary() Geometry {
	if ls.IsEmpty() || ls.IsClosed() {
		return NewMultiPointEmpty()
	}
	points := []*Point{
		NewPointFromCoordinate(ls.coords.First()),
		NewPointFromCoordinate(ls.coords.Last()),
	}
	return NewMultiPoint(points)
}

// Coordinates returns the coordinate sequence.
func (ls *LineString) Coordinates() CoordinateSequence {
	return ls.coords.Clone()
}

// NumGeometries returns 1 for LineString.
func (ls *LineString) NumGeometries() int {
	return 1
}

// GeometryN returns the linestring itself (for n=0).
func (ls *LineString) GeometryN(n int) Geometry {
	if n != 0 {
		return nil
	}
	return ls
}

// Clone returns a deep copy.
func (ls *LineString) Clone() Geometry {
	clone := NewLineString(ls.coords)
	clone.srid = ls.srid
	return clone
}

// Normalize normalizes the linestring to canonical form.
func (ls *LineString) Normalize() {
	if ls.IsEmpty() {
		return
	}
	// For non-closed linestrings, ensure first point < last point
	if !ls.IsClosed() && Compare(NewPointFromCoordinate(ls.coords.First()),
		NewPointFromCoordinate(ls.coords.Last())) > 0 {
		ls.coords = ls.coords.Reverse()
	}
}

// EqualsExact returns true if the linestrings are exactly equal.
func (ls *LineString) EqualsExact(other Geometry, tolerance float64) bool {
	if other == nil {
		return false
	}
	otherLS, ok := other.(*LineString)
	if !ok {
		return false
	}
	if len(ls.coords) != len(otherLS.coords) {
		return false
	}
	for i, c := range ls.coords {
		if !c.Equals2D(otherLS.coords[i], tolerance) {
			return false
		}
	}
	return true
}

// String returns the WKT representation.
func (ls *LineString) String() string {
	if ls.IsEmpty() {
		return "LINESTRING EMPTY"
	}

	hasZ := ls.coords.HasZ()
	hasM := ls.coords.HasM()

	var sb strings.Builder
	sb.WriteString("LINESTRING ")

	if hasZ && hasM {
		sb.WriteString("ZM ")
	} else if hasZ {
		sb.WriteString("Z ")
	} else if hasM {
		sb.WriteString("M ")
	}

	sb.WriteString("(")
	for i, c := range ls.coords {
		if i > 0 {
			sb.WriteString(", ")
		}
		if hasZ && hasM {
			sb.WriteString(fmt.Sprintf("%g %g %g %g", c.X, c.Y, c.GetZ(), c.GetM()))
		} else if hasZ {
			sb.WriteString(fmt.Sprintf("%g %g %g", c.X, c.Y, c.GetZ()))
		} else if hasM {
			sb.WriteString(fmt.Sprintf("%g %g %g", c.X, c.Y, c.GetM()))
		} else {
			sb.WriteString(fmt.Sprintf("%g %g", c.X, c.Y))
		}
	}
	sb.WriteString(")")
	return sb.String()
}

// NumPoints returns the number of points.
func (ls *LineString) NumPoints() int {
	return len(ls.coords)
}

// PointN returns the nth point (0-indexed).
func (ls *LineString) PointN(n int) *Point {
	if n < 0 || n >= len(ls.coords) {
		return nil
	}
	return NewPointFromCoordinate(ls.coords[n])
}

// StartPoint returns the first point.
func (ls *LineString) StartPoint() *Point {
	if ls.IsEmpty() {
		return nil
	}
	return NewPointFromCoordinate(ls.coords.First())
}

// EndPoint returns the last point.
func (ls *LineString) EndPoint() *Point {
	if ls.IsEmpty() {
		return nil
	}
	return NewPointFromCoordinate(ls.coords.Last())
}

// IsClosed returns true if the first and last points are equal.
func (ls *LineString) IsClosed() bool {
	if ls.IsEmpty() {
		return false
	}
	return ls.coords.IsClosed(DefaultEpsilon)
}

// IsRing returns true if the linestring is closed and simple.
func (ls *LineString) IsRing() bool {
	return ls.IsClosed() && ls.IsSimple()
}

// Length returns the length of the linestring.
func (ls *LineString) Length() float64 {
	if len(ls.coords) < 2 {
		return 0
	}
	length := 0.0
	for i := 1; i < len(ls.coords); i++ {
		length += ls.coords[i-1].Distance(ls.coords[i])
	}
	return length
}

// Reverse returns a new linestring with reversed coordinate order.
func (ls *LineString) Reverse() *LineString {
	reversed := NewLineString(ls.coords.Reverse())
	reversed.srid = ls.srid
	return reversed
}

// CoordinateN returns the nth coordinate (0-indexed).
func (ls *LineString) CoordinateN(n int) Coordinate {
	return ls.coords[n]
}

// SegmentLength returns the length of the nth segment (0-indexed).
func (ls *LineString) SegmentLength(n int) float64 {
	if n < 0 || n >= len(ls.coords)-1 {
		return 0
	}
	return ls.coords[n].Distance(ls.coords[n+1])
}

// PointAlong returns the point at the given fraction along the linestring.
// Fraction should be between 0 and 1.
func (ls *LineString) PointAlong(fraction float64) *Point {
	if ls.IsEmpty() {
		return NewPointEmpty()
	}
	if fraction <= 0 {
		return ls.StartPoint()
	}
	if fraction >= 1 {
		return ls.EndPoint()
	}

	totalLength := ls.Length()
	targetLength := totalLength * fraction

	currentLength := 0.0
	for i := 1; i < len(ls.coords); i++ {
		segmentLength := ls.coords[i-1].Distance(ls.coords[i])
		if currentLength+segmentLength >= targetLength {
			// Interpolate within this segment
			segmentFraction := (targetLength - currentLength) / segmentLength
			x := ls.coords[i-1].X + segmentFraction*(ls.coords[i].X-ls.coords[i-1].X)
			y := ls.coords[i-1].Y + segmentFraction*(ls.coords[i].Y-ls.coords[i-1].Y)
			return NewPoint(x, y)
		}
		currentLength += segmentLength
	}

	return ls.EndPoint()
}

// Centroid returns the centroid of the linestring.
func (ls *LineString) Centroid() *Point {
	if ls.IsEmpty() {
		return NewPointEmpty()
	}

	totalLength := 0.0
	sumX := 0.0
	sumY := 0.0

	for i := 1; i < len(ls.coords); i++ {
		p1 := ls.coords[i-1]
		p2 := ls.coords[i]
		segLength := p1.Distance(p2)
		midX := (p1.X + p2.X) / 2
		midY := (p1.Y + p2.Y) / 2
		sumX += midX * segLength
		sumY += midY * segLength
		totalLength += segLength
	}

	if totalLength == 0 {
		return NewPointFromCoordinate(ls.coords[0])
	}

	return NewPoint(sumX/totalLength, sumY/totalLength)
}

// ClosestPoint returns the point on this linestring closest to the given coordinate.
func (ls *LineString) ClosestPoint(c Coordinate) *Point {
	if ls.IsEmpty() {
		return NewPointEmpty()
	}

	minDist := math.MaxFloat64
	var closest Coordinate

	for i := 1; i < len(ls.coords); i++ {
		p := closestPointOnSegment(c, ls.coords[i-1], ls.coords[i])
		dist := c.Distance(p)
		if dist < minDist {
			minDist = dist
			closest = p
		}
	}

	return NewPointFromCoordinate(closest)
}

func closestPointOnSegment(p, a, b Coordinate) Coordinate {
	dx := b.X - a.X
	dy := b.Y - a.Y

	if dx == 0 && dy == 0 {
		return a
	}

	t := ((p.X-a.X)*dx + (p.Y-a.Y)*dy) / (dx*dx + dy*dy)
	t = math.Max(0, math.Min(1, t))

	return NewCoordinate(a.X+t*dx, a.Y+t*dy)
}
