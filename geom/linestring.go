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
// This checks all segment pairs for interior intersections.
func (ls *LineString) IsSimple() bool {
	if len(ls.coords) <= 3 {
		return true
	}
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

	info := segmentIntersectionInfo(p1, p2, p3, p4)
	if !info.intersects {
		return false
	}
	if ls.IsClosed() {
		return true
	}
	if info.proper || info.collinearOverlap {
		return true
	}
	for _, p := range info.points {
		if !ls.isBoundaryPoint(p) {
			return true
		}
	}
	// If we couldn't identify an intersection point, treat it as non-simple.
	return len(info.points) == 0
}

type segmentIntersection struct {
	intersects       bool
	proper           bool
	collinearOverlap bool
	points           []Coordinate
}

func segmentIntersectionInfo(a1, a2, b1, b2 Coordinate) segmentIntersection {
	o1 := orientation(a1, a2, b1)
	o2 := orientation(a1, a2, b2)
	o3 := orientation(b1, b2, a1)
	o4 := orientation(b1, b2, a2)

	// Proper crossing (interior-interior)
	if o1 != o2 && o3 != o4 && o1 != 0 && o2 != 0 && o3 != 0 && o4 != 0 {
		return segmentIntersection{intersects: true, proper: true}
	}

	// Collinear case
	if o1 == 0 && o2 == 0 && o3 == 0 && o4 == 0 {
		return collinearIntersectionInfo(a1, a2, b1, b2)
	}

	info := segmentIntersection{}
	if o1 == 0 && onSegmentBounds(a1, b1, a2) {
		info = addIntersectionPoint(info, b1)
	}
	if o2 == 0 && onSegmentBounds(a1, b2, a2) {
		info = addIntersectionPoint(info, b2)
	}
	if o3 == 0 && onSegmentBounds(b1, a1, b2) {
		info = addIntersectionPoint(info, a1)
	}
	if o4 == 0 && onSegmentBounds(b1, a2, b2) {
		info = addIntersectionPoint(info, a2)
	}
	if len(info.points) > 0 {
		info.intersects = true
	}
	return info
}

func collinearIntersectionInfo(a1, a2, b1, b2 Coordinate) segmentIntersection {
	info := segmentIntersection{}

	onA1 := onSegmentBounds(b1, a1, b2)
	onA2 := onSegmentBounds(b1, a2, b2)
	onB1 := onSegmentBounds(a1, b1, a2)
	onB2 := onSegmentBounds(a1, b2, a2)

	if !onA1 && !onA2 && !onB1 && !onB2 {
		return info
	}

	info.intersects = true

	shared := []Coordinate{}
	shared = addSharedEndpoint(shared, a1, b1)
	shared = addSharedEndpoint(shared, a1, b2)
	shared = addSharedEndpoint(shared, a2, b1)
	shared = addSharedEndpoint(shared, a2, b2)

	overlap := len(shared) >= 2
	if (onA1 && !isSharedPoint(shared, a1)) ||
		(onA2 && !isSharedPoint(shared, a2)) ||
		(onB1 && !isSharedPoint(shared, b1)) ||
		(onB2 && !isSharedPoint(shared, b2)) {
		overlap = true
	}

	if overlap {
		info.collinearOverlap = true
		return info
	}

	for _, p := range shared {
		info = addIntersectionPoint(info, p)
	}

	return info
}

func addSharedEndpoint(shared []Coordinate, a, b Coordinate) []Coordinate {
	if a.Equals2D(b, DefaultEpsilon) && !isSharedPoint(shared, a) {
		return append(shared, a.Clone())
	}
	return shared
}

func isSharedPoint(shared []Coordinate, p Coordinate) bool {
	for _, s := range shared {
		if s.Equals2D(p, DefaultEpsilon) {
			return true
		}
	}
	return false
}

func addIntersectionPoint(info segmentIntersection, p Coordinate) segmentIntersection {
	for _, existing := range info.points {
		if existing.Equals2D(p, DefaultEpsilon) {
			return info
		}
	}
	info.points = append(info.points, p.Clone())
	return info
}

func (ls *LineString) isBoundaryPoint(p Coordinate) bool {
	if ls.IsClosed() || ls.IsEmpty() {
		return false
	}
	return p.Equals2D(ls.coords.First(), DefaultEpsilon) ||
		p.Equals2D(ls.coords.Last(), DefaultEpsilon)
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

// ApplyCoordinateFilter applies a coordinate filter to the linestring.
func (ls *LineString) ApplyCoordinateFilter(filter CoordinateFilter) {
	if filter == nil {
		return
	}
	for i := range ls.coords {
		filter.Filter(&ls.coords[i])
	}
	ls.invalidateEnvelope()
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
	if totalLength == 0 {
		return NewPointFromCoordinate(ls.coords.First())
	}
	targetLength := totalLength * fraction

	currentLength := 0.0
	for i := 1; i < len(ls.coords); i++ {
		segmentLength := ls.coords[i-1].Distance(ls.coords[i])
		if segmentLength == 0 {
			continue
		}
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
		p := ClosestPointOnSegment(c, ls.coords[i-1], ls.coords[i])
		dist := c.Distance(p)
		if dist < minDist {
			minDist = dist
			closest = p
		}
	}

	return NewPointFromCoordinate(closest)
}

// ClosestPointOnSegment returns the closest point on segment (a,b) to point p.
func ClosestPointOnSegment(p, a, b Coordinate) Coordinate {
	dx := b.X - a.X
	dy := b.Y - a.Y

	if a.Equals2D(b, DefaultEpsilon) {
		return a
	}

	t := ((p.X-a.X)*dx + (p.Y-a.Y)*dy) / (dx*dx + dy*dy)
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}

	return NewCoordinate(a.X+t*dx, a.Y+t*dy)
}
