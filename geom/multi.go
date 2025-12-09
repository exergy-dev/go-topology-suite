package geom

import (
	"fmt"
	"strings"
)

// MultiPoint represents a collection of Points.
type MultiPoint struct {
	baseGeometry
	points []*Point
}

// NewMultiPoint creates a new MultiPoint from Points.
func NewMultiPoint(points []*Point) *MultiPoint {
	mp := &MultiPoint{
		points: make([]*Point, len(points)),
	}
	for i, p := range points {
		mp.points[i] = p.Clone().(*Point)
	}
	return mp
}

// NewMultiPointFromCoords creates a MultiPoint from coordinates.
func NewMultiPointFromCoords(coords CoordinateSequence) *MultiPoint {
	points := make([]*Point, len(coords))
	for i, c := range coords {
		points[i] = NewPointFromCoordinate(c)
	}
	return &MultiPoint{points: points}
}

// NewMultiPointEmpty creates an empty MultiPoint.
func NewMultiPointEmpty() *MultiPoint {
	return &MultiPoint{points: []*Point{}}
}

// GeometryType returns "MultiPoint".
func (mp *MultiPoint) GeometryType() string {
	return "MultiPoint"
}

// Envelope returns the bounding box.
func (mp *MultiPoint) Envelope() *Envelope {
	if mp.envelope == nil {
		mp.envelope = NewEnvelopeEmpty()
		for _, p := range mp.points {
			mp.envelope.ExpandToInclude(p.Envelope())
		}
	}
	return mp.envelope.Clone()
}

// IsEmpty returns true if there are no points.
func (mp *MultiPoint) IsEmpty() bool {
	return len(mp.points) == 0
}

// IsSimple returns true if all points are distinct.
func (mp *MultiPoint) IsSimple() bool {
	for i := 0; i < len(mp.points); i++ {
		for j := i + 1; j < len(mp.points); j++ {
			if mp.points[i].EqualsExact(mp.points[j], DefaultEpsilon) {
				return false
			}
		}
	}
	return true
}

// IsValid returns true (MultiPoints are always valid).
func (mp *MultiPoint) IsValid() bool {
	return true
}

// Dimension returns 0 for MultiPoint.
func (mp *MultiPoint) Dimension() Dimension {
	return DimensionPoint
}

// Boundary returns an empty GeometryCollection (points have no boundary).
func (mp *MultiPoint) Boundary() Geometry {
	return NewGeometryCollectionEmpty()
}

// Coordinates returns all point coordinates.
func (mp *MultiPoint) Coordinates() CoordinateSequence {
	coords := make(CoordinateSequence, len(mp.points))
	for i, p := range mp.points {
		coords[i] = p.coord.Clone()
	}
	return coords
}

// NumGeometries returns the number of points.
func (mp *MultiPoint) NumGeometries() int {
	return len(mp.points)
}

// GeometryN returns the nth point (0-indexed).
func (mp *MultiPoint) GeometryN(n int) Geometry {
	if n < 0 || n >= len(mp.points) {
		return nil
	}
	return mp.points[n]
}

// Clone returns a deep copy.
func (mp *MultiPoint) Clone() Geometry {
	clone := NewMultiPoint(mp.points)
	clone.srid = mp.srid
	return clone
}

// Normalize normalizes by sorting points.
func (mp *MultiPoint) Normalize() {
	// Sort points by coordinates
	for i := 0; i < len(mp.points); i++ {
		for j := i + 1; j < len(mp.points); j++ {
			if Compare(mp.points[i], mp.points[j]) > 0 {
				mp.points[i], mp.points[j] = mp.points[j], mp.points[i]
			}
		}
	}
}

// EqualsExact returns true if the MultiPoints are exactly equal.
func (mp *MultiPoint) EqualsExact(other Geometry, tolerance float64) bool {
	if other == nil {
		return false
	}
	otherMP, ok := other.(*MultiPoint)
	if !ok {
		return false
	}
	if len(mp.points) != len(otherMP.points) {
		return false
	}
	for i, p := range mp.points {
		if !p.EqualsExact(otherMP.points[i], tolerance) {
			return false
		}
	}
	return true
}

// String returns the WKT representation.
func (mp *MultiPoint) String() string {
	if mp.IsEmpty() {
		return "MULTIPOINT EMPTY"
	}

	var sb strings.Builder
	sb.WriteString("MULTIPOINT (")
	for i, p := range mp.points {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("(%g %g)", p.coord.X, p.coord.Y))
	}
	sb.WriteString(")")
	return sb.String()
}

// PointN returns the nth point (0-indexed).
func (mp *MultiPoint) PointN(n int) *Point {
	if n < 0 || n >= len(mp.points) {
		return nil
	}
	return mp.points[n]
}

// MultiLineString represents a collection of LineStrings.
type MultiLineString struct {
	baseGeometry
	lines []*LineString
}

// NewMultiLineString creates a new MultiLineString from LineStrings.
func NewMultiLineString(lines []*LineString) *MultiLineString {
	mls := &MultiLineString{
		lines: make([]*LineString, len(lines)),
	}
	for i, l := range lines {
		mls.lines[i] = l.Clone().(*LineString)
	}
	return mls
}

// NewMultiLineStringEmpty creates an empty MultiLineString.
func NewMultiLineStringEmpty() *MultiLineString {
	return &MultiLineString{lines: []*LineString{}}
}

// GeometryType returns "MultiLineString".
func (mls *MultiLineString) GeometryType() string {
	return "MultiLineString"
}

// Envelope returns the bounding box.
func (mls *MultiLineString) Envelope() *Envelope {
	if mls.envelope == nil {
		mls.envelope = NewEnvelopeEmpty()
		for _, l := range mls.lines {
			mls.envelope.ExpandToInclude(l.Envelope())
		}
	}
	return mls.envelope.Clone()
}

// IsEmpty returns true if there are no linestrings.
func (mls *MultiLineString) IsEmpty() bool {
	return len(mls.lines) == 0
}

// IsSimple returns true if no linestrings self-intersect or intersect each other
// (except at endpoints).
func (mls *MultiLineString) IsSimple() bool {
	for _, l := range mls.lines {
		if !l.IsSimple() {
			return false
		}
	}
	// Full implementation would check inter-linestring intersections
	return true
}

// IsValid returns true if all linestrings are valid.
func (mls *MultiLineString) IsValid() bool {
	for _, l := range mls.lines {
		if !l.IsValid() {
			return false
		}
	}
	return true
}

// Dimension returns 1 for MultiLineString.
func (mls *MultiLineString) Dimension() Dimension {
	return DimensionLine
}

// Boundary returns the boundary (endpoints with odd degree).
func (mls *MultiLineString) Boundary() Geometry {
	if mls.IsEmpty() {
		return NewMultiPointEmpty()
	}

	// Count endpoint occurrences
	endpoints := make(map[string]int)
	coordKey := func(c Coordinate) string {
		return fmt.Sprintf("%g,%g", c.X, c.Y)
	}

	for _, l := range mls.lines {
		if l.IsEmpty() {
			continue
		}
		if l.IsClosed() {
			continue
		}
		endpoints[coordKey(l.coords.First())]++
		endpoints[coordKey(l.coords.Last())]++
	}

	// Collect points with odd degree
	var points []*Point
	for _, l := range mls.lines {
		if l.IsEmpty() || l.IsClosed() {
			continue
		}
		if endpoints[coordKey(l.coords.First())]%2 == 1 {
			points = append(points, NewPointFromCoordinate(l.coords.First()))
		}
		if endpoints[coordKey(l.coords.Last())]%2 == 1 {
			points = append(points, NewPointFromCoordinate(l.coords.Last()))
		}
	}

	return NewMultiPoint(points)
}

// Coordinates returns all coordinates from all linestrings.
func (mls *MultiLineString) Coordinates() CoordinateSequence {
	var coords CoordinateSequence
	for _, l := range mls.lines {
		coords = append(coords, l.coords...)
	}
	return coords
}

// NumGeometries returns the number of linestrings.
func (mls *MultiLineString) NumGeometries() int {
	return len(mls.lines)
}

// GeometryN returns the nth linestring (0-indexed).
func (mls *MultiLineString) GeometryN(n int) Geometry {
	if n < 0 || n >= len(mls.lines) {
		return nil
	}
	return mls.lines[n]
}

// Clone returns a deep copy.
func (mls *MultiLineString) Clone() Geometry {
	clone := NewMultiLineString(mls.lines)
	clone.srid = mls.srid
	return clone
}

// Normalize normalizes all component linestrings.
func (mls *MultiLineString) Normalize() {
	for _, l := range mls.lines {
		l.Normalize()
	}
	// Sort linestrings
	for i := 0; i < len(mls.lines); i++ {
		for j := i + 1; j < len(mls.lines); j++ {
			if Compare(mls.lines[i], mls.lines[j]) > 0 {
				mls.lines[i], mls.lines[j] = mls.lines[j], mls.lines[i]
			}
		}
	}
}

// EqualsExact returns true if the MultiLineStrings are exactly equal.
func (mls *MultiLineString) EqualsExact(other Geometry, tolerance float64) bool {
	if other == nil {
		return false
	}
	otherMLS, ok := other.(*MultiLineString)
	if !ok {
		return false
	}
	if len(mls.lines) != len(otherMLS.lines) {
		return false
	}
	for i, l := range mls.lines {
		if !l.EqualsExact(otherMLS.lines[i], tolerance) {
			return false
		}
	}
	return true
}

// String returns the WKT representation.
func (mls *MultiLineString) String() string {
	if mls.IsEmpty() {
		return "MULTILINESTRING EMPTY"
	}

	var sb strings.Builder
	sb.WriteString("MULTILINESTRING (")
	for i, l := range mls.lines {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("(")
		for j, c := range l.coords {
			if j > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%g %g", c.X, c.Y))
		}
		sb.WriteString(")")
	}
	sb.WriteString(")")
	return sb.String()
}

// Length returns the total length of all linestrings.
func (mls *MultiLineString) Length() float64 {
	length := 0.0
	for _, l := range mls.lines {
		length += l.Length()
	}
	return length
}

// IsClosed returns true if all linestrings are closed.
func (mls *MultiLineString) IsClosed() bool {
	for _, l := range mls.lines {
		if !l.IsClosed() {
			return false
		}
	}
	return true
}

// LineStringN returns the nth linestring (0-indexed).
func (mls *MultiLineString) LineStringN(n int) *LineString {
	if n < 0 || n >= len(mls.lines) {
		return nil
	}
	return mls.lines[n]
}

// MultiPolygon represents a collection of Polygons.
type MultiPolygon struct {
	baseGeometry
	polygons []*Polygon
}

// NewMultiPolygon creates a new MultiPolygon from Polygons.
func NewMultiPolygon(polygons []*Polygon) *MultiPolygon {
	mp := &MultiPolygon{
		polygons: make([]*Polygon, len(polygons)),
	}
	for i, p := range polygons {
		mp.polygons[i] = p.Clone().(*Polygon)
	}
	return mp
}

// NewMultiPolygonEmpty creates an empty MultiPolygon.
func NewMultiPolygonEmpty() *MultiPolygon {
	return &MultiPolygon{polygons: []*Polygon{}}
}

// GeometryType returns "MultiPolygon".
func (mp *MultiPolygon) GeometryType() string {
	return "MultiPolygon"
}

// Envelope returns the bounding box.
func (mp *MultiPolygon) Envelope() *Envelope {
	if mp.envelope == nil {
		mp.envelope = NewEnvelopeEmpty()
		for _, p := range mp.polygons {
			mp.envelope.ExpandToInclude(p.Envelope())
		}
	}
	return mp.envelope.Clone()
}

// IsEmpty returns true if there are no polygons.
func (mp *MultiPolygon) IsEmpty() bool {
	return len(mp.polygons) == 0
}

// IsSimple returns true (MultiPolygons are simple by definition).
func (mp *MultiPolygon) IsSimple() bool {
	return true
}

// IsValid returns true if all polygons are valid and don't overlap.
func (mp *MultiPolygon) IsValid() bool {
	for _, p := range mp.polygons {
		if !p.IsValid() {
			return false
		}
	}
	// Full implementation would check for overlaps
	return true
}

// Dimension returns 2 for MultiPolygon.
func (mp *MultiPolygon) Dimension() Dimension {
	return DimensionArea
}

// Boundary returns the boundary (all exterior and hole rings).
func (mp *MultiPolygon) Boundary() Geometry {
	if mp.IsEmpty() {
		return NewMultiLineStringEmpty()
	}

	var rings []*LineString
	for _, p := range mp.polygons {
		if !p.IsEmpty() {
			rings = append(rings, p.shell.LineString)
			for _, hole := range p.holes {
				rings = append(rings, hole.LineString)
			}
		}
	}

	return NewMultiLineString(rings)
}

// Coordinates returns all coordinates from all polygons.
func (mp *MultiPolygon) Coordinates() CoordinateSequence {
	var coords CoordinateSequence
	for _, p := range mp.polygons {
		coords = append(coords, p.Coordinates()...)
	}
	return coords
}

// NumGeometries returns the number of polygons.
func (mp *MultiPolygon) NumGeometries() int {
	return len(mp.polygons)
}

// GeometryN returns the nth polygon (0-indexed).
func (mp *MultiPolygon) GeometryN(n int) Geometry {
	if n < 0 || n >= len(mp.polygons) {
		return nil
	}
	return mp.polygons[n]
}

// Clone returns a deep copy.
func (mp *MultiPolygon) Clone() Geometry {
	clone := NewMultiPolygon(mp.polygons)
	clone.srid = mp.srid
	return clone
}

// Normalize normalizes all component polygons.
func (mp *MultiPolygon) Normalize() {
	for _, p := range mp.polygons {
		p.Normalize()
	}
	// Sort polygons
	for i := 0; i < len(mp.polygons); i++ {
		for j := i + 1; j < len(mp.polygons); j++ {
			if Compare(mp.polygons[i], mp.polygons[j]) > 0 {
				mp.polygons[i], mp.polygons[j] = mp.polygons[j], mp.polygons[i]
			}
		}
	}
}

// EqualsExact returns true if the MultiPolygons are exactly equal.
func (mp *MultiPolygon) EqualsExact(other Geometry, tolerance float64) bool {
	if other == nil {
		return false
	}
	otherMP, ok := other.(*MultiPolygon)
	if !ok {
		return false
	}
	if len(mp.polygons) != len(otherMP.polygons) {
		return false
	}
	for i, p := range mp.polygons {
		if !p.EqualsExact(otherMP.polygons[i], tolerance) {
			return false
		}
	}
	return true
}

// String returns the WKT representation.
func (mp *MultiPolygon) String() string {
	if mp.IsEmpty() {
		return "MULTIPOLYGON EMPTY"
	}

	var sb strings.Builder
	sb.WriteString("MULTIPOLYGON (")
	for i, p := range mp.polygons {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("(")
		sb.WriteString(ringCoordsToString(p.shell.coords, false, false))
		for _, hole := range p.holes {
			sb.WriteString(", ")
			sb.WriteString(ringCoordsToString(hole.coords, false, false))
		}
		sb.WriteString(")")
	}
	sb.WriteString(")")
	return sb.String()
}

// Area returns the total area of all polygons.
func (mp *MultiPolygon) Area() float64 {
	area := 0.0
	for _, p := range mp.polygons {
		area += p.Area()
	}
	return area
}

// Perimeter returns the total perimeter of all polygons.
func (mp *MultiPolygon) Perimeter() float64 {
	perimeter := 0.0
	for _, p := range mp.polygons {
		perimeter += p.Perimeter()
	}
	return perimeter
}

// PolygonN returns the nth polygon (0-indexed).
func (mp *MultiPolygon) PolygonN(n int) *Polygon {
	if n < 0 || n >= len(mp.polygons) {
		return nil
	}
	return mp.polygons[n]
}
