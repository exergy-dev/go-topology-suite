package geom

import (
	"fmt"
	"math"
	"sort"
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
	if env := mp.cachedEnvelope(); env != nil {
		return env.Clone()
	}
	env := NewEnvelopeEmpty()
	for _, p := range mp.points {
		env.ExpandToInclude(p.Envelope())
	}
	mp.setCachedEnvelope(env)
	return env.Clone()
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

// ApplyCoordinateFilter applies a coordinate filter to the multipoint.
func (mp *MultiPoint) ApplyCoordinateFilter(filter CoordinateFilter) {
	if filter == nil {
		return
	}
	for _, p := range mp.points {
		p.ApplyCoordinateFilter(filter)
	}
	mp.invalidateEnvelope()
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

// Normalized returns a new MultiPoint with points sorted in canonical order.
func (mp *MultiPoint) Normalized() Geometry {
	clone := mp.Clone().(*MultiPoint)
	sort.Slice(clone.points, func(i, j int) bool {
		return Compare(clone.points[i], clone.points[j]) < 0
	})
	return clone
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

	// Detect dimensions from points
	hasZ, hasM := false, false
	for _, p := range mp.points {
		if p.coord.HasZ() {
			hasZ = true
		}
		if p.coord.HasM() {
			hasM = true
		}
		if hasZ && hasM {
			break
		}
	}

	var sb strings.Builder
	sb.WriteString("MULTIPOINT ")
	if hasZ && hasM {
		sb.WriteString("ZM ")
	} else if hasZ {
		sb.WriteString("Z ")
	} else if hasM {
		sb.WriteString("M ")
	}
	sb.WriteString("(")
	for i, p := range mp.points {
		if i > 0 {
			sb.WriteString(", ")
		}
		if hasZ && hasM {
			sb.WriteString(fmt.Sprintf("(%g %g %g %g)", p.coord.X, p.coord.Y, p.coord.GetZ(), p.coord.GetM()))
		} else if hasZ {
			sb.WriteString(fmt.Sprintf("(%g %g %g)", p.coord.X, p.coord.Y, p.coord.GetZ()))
		} else if hasM {
			sb.WriteString(fmt.Sprintf("(%g %g %g)", p.coord.X, p.coord.Y, p.coord.GetM()))
		} else {
			sb.WriteString(fmt.Sprintf("(%g %g)", p.coord.X, p.coord.Y))
		}
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
	if env := mls.cachedEnvelope(); env != nil {
		return env.Clone()
	}
	env := NewEnvelopeEmpty()
	for _, l := range mls.lines {
		env.ExpandToInclude(l.Envelope())
	}
	mls.setCachedEnvelope(env)
	return env.Clone()
}

// IsEmpty returns true if there are no linestrings.
func (mls *MultiLineString) IsEmpty() bool {
	return len(mls.lines) == 0
}

// IsSimple returns true if no linestrings self-intersect or intersect each other
// improperly (interior intersections are not allowed).
func (mls *MultiLineString) IsSimple() bool {
	// Check each linestring is simple
	for _, l := range mls.lines {
		if !l.IsSimple() {
			return false
		}
	}

	// Check for inter-linestring interior intersections
	for i := 0; i < len(mls.lines); i++ {
		for j := i + 1; j < len(mls.lines); j++ {
			if mls.linesIntersectImproperly(mls.lines[i], mls.lines[j]) {
				return false
			}
		}
	}
	return true
}

// linesIntersectImproperly checks if two linestrings have an interior intersection
// (i.e., they cross each other, not just touch at endpoints).
func (mls *MultiLineString) linesIntersectImproperly(l1, l2 *LineString) bool {
	// Quick envelope check
	if !l1.Envelope().Intersects(l2.Envelope()) {
		return false
	}

	// Check all segment pairs between the two linestrings
	for i := 0; i < len(l1.coords)-1; i++ {
		for j := 0; j < len(l2.coords)-1; j++ {
			info := segmentIntersectionInfo(
				l1.coords[i], l1.coords[i+1],
				l2.coords[j], l2.coords[j+1])
			if !info.intersects {
				continue
			}
			if info.proper || info.collinearOverlap {
				return true
			}
			for _, pt := range info.points {
				if !isLineBoundaryPoint(l1, pt) || !isLineBoundaryPoint(l2, pt) {
					return true
				}
			}
			if len(info.points) == 0 {
				return true
			}
		}
	}
	return false
}

func isLineBoundaryPoint(line *LineString, p Coordinate) bool {
	if line.IsEmpty() || line.IsClosed() {
		return false
	}
	return p.Equals2D(line.coords.First(), DefaultEpsilon) ||
		p.Equals2D(line.coords.Last(), DefaultEpsilon)
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

	// Count endpoint occurrences by exact coordinate value.
	type coordKey struct {
		x uint64
		y uint64
	}
	type endpointInfo struct {
		count int
		coord Coordinate
	}
	endpoints := make(map[coordKey]endpointInfo)
	addEndpoint := func(c Coordinate) {
		key := coordKey{math.Float64bits(c.X), math.Float64bits(c.Y)}
		info := endpoints[key]
		if info.count == 0 {
			info.coord = c.Clone()
		}
		info.count++
		endpoints[key] = info
	}

	for _, l := range mls.lines {
		if l.IsEmpty() {
			continue
		}
		if l.IsClosed() {
			continue
		}
		addEndpoint(l.coords.First())
		addEndpoint(l.coords.Last())
	}

	// Collect points with odd degree
	var points []*Point
	for _, info := range endpoints {
		if info.count%2 == 1 {
			points = append(points, NewPointFromCoordinate(info.coord))
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

// ApplyCoordinateFilter applies a coordinate filter to the multilinestring.
func (mls *MultiLineString) ApplyCoordinateFilter(filter CoordinateFilter) {
	if filter == nil {
		return
	}
	for _, l := range mls.lines {
		l.ApplyCoordinateFilter(filter)
	}
	mls.invalidateEnvelope()
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

// Normalized returns a new MultiLineString with all components normalized.
func (mls *MultiLineString) Normalized() Geometry {
	clone := mls.Clone().(*MultiLineString)
	for i, l := range clone.lines {
		clone.lines[i] = l.Normalized().(*LineString)
	}
	sort.Slice(clone.lines, func(i, j int) bool {
		return Compare(clone.lines[i], clone.lines[j]) < 0
	})
	return clone
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

	hasZ := mls.Coordinates().HasZ()
	hasM := mls.Coordinates().HasM()

	var sb strings.Builder
	sb.WriteString("MULTILINESTRING ")
	if hasZ && hasM {
		sb.WriteString("ZM ")
	} else if hasZ {
		sb.WriteString("Z ")
	} else if hasM {
		sb.WriteString("M ")
	}
	sb.WriteString("(")
	for i, l := range mls.lines {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(ringCoordsToString(l.coords, hasZ, hasM))
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
	if env := mp.cachedEnvelope(); env != nil {
		return env.Clone()
	}
	env := NewEnvelopeEmpty()
	for _, p := range mp.polygons {
		env.ExpandToInclude(p.Envelope())
	}
	mp.setCachedEnvelope(env)
	return env.Clone()
}

// IsEmpty returns true if there are no polygons.
func (mp *MultiPolygon) IsEmpty() bool {
	return len(mp.polygons) == 0
}

// IsSimple returns true (MultiPolygons are simple by definition).
func (mp *MultiPolygon) IsSimple() bool {
	return true
}

// IsValid returns true if all polygons are valid and their interiors don't overlap.
func (mp *MultiPolygon) IsValid() bool {
	for _, p := range mp.polygons {
		if !p.IsValid() {
			return false
		}
	}

	// Check for polygon interior overlaps
	for i := 0; i < len(mp.polygons); i++ {
		for j := i + 1; j < len(mp.polygons); j++ {
			if mp.polygonsOverlap(mp.polygons[i], mp.polygons[j]) {
				return false
			}
		}
	}
	return true
}

// polygonsOverlap checks if two polygons have overlapping interiors.
func (mp *MultiPolygon) polygonsOverlap(p1, p2 *Polygon) bool {
	// Quick envelope check
	if !p1.Envelope().Intersects(p2.Envelope()) {
		return false
	}

	// Proper boundary crossings imply interior overlap.
	if ringsCrossProperly(p1.shell, p2.shell) {
		return true
	}
	if polygonsOverlapAtBoundarySegment(p1, p2) {
		return true
	}

	// If any shell segment midpoint is strictly inside the other polygon,
	// the interiors overlap (covers containment and corner overlaps).
	if shellHasInteriorPointIn(p1, p2) {
		return true
	}
	if shellHasInteriorPointIn(p2, p1) {
		return true
	}

	return false
}

func polygonsOverlapAtBoundarySegment(p1, p2 *Polygon) bool {
	for _, r1 := range polygonBoundaryRings(p1) {
		for _, r2 := range polygonBoundaryRings(p2) {
			if ringsOverlapAtSegment(r1, r2) {
				return true
			}
		}
	}
	return false
}

func polygonBoundaryRings(p *Polygon) []*LinearRing {
	rings := make([]*LinearRing, 0, 1+len(p.holes))
	rings = append(rings, p.shell)
	rings = append(rings, p.holes...)
	return rings
}

// ringsCrossProperly checks if two rings have a proper crossing (not just touching).
func ringsCrossProperly(r1, r2 *LinearRing) bool {
	coords1, coords2 := r1.Coordinates(), r2.Coordinates()
	for i := 0; i < len(coords1)-1; i++ {
		for j := 0; j < len(coords2)-1; j++ {
			if segmentsCrossProper(coords1[i], coords1[i+1], coords2[j], coords2[j+1]) {
				return true
			}
		}
	}
	return false
}

// shellHasInteriorPointIn returns true if any shell segment midpoint lies in the
// interior of the other polygon (not on the boundary).
func shellHasInteriorPointIn(p *Polygon, other *Polygon) bool {
	if p.IsEmpty() || other.IsEmpty() {
		return false
	}

	coords := p.shell.Coordinates()
	for i := 1; i < len(coords); i++ {
		mid := Coordinate{
			X: (coords[i-1].X + coords[i].X) / 2,
			Y: (coords[i-1].Y + coords[i].Y) / 2,
		}
		if pointInPolygon(mid, other) == LocationInterior {
			return true
		}
	}
	return false
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

// ApplyCoordinateFilter applies a coordinate filter to the multipolygon.
func (mp *MultiPolygon) ApplyCoordinateFilter(filter CoordinateFilter) {
	if filter == nil {
		return
	}
	for _, p := range mp.polygons {
		p.ApplyCoordinateFilter(filter)
	}
	mp.invalidateEnvelope()
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

// Normalized returns a new MultiPolygon with all components normalized.
func (mp *MultiPolygon) Normalized() Geometry {
	clone := mp.Clone().(*MultiPolygon)
	for i, p := range clone.polygons {
		clone.polygons[i] = p.Normalized().(*Polygon)
	}
	sort.Slice(clone.polygons, func(i, j int) bool {
		return Compare(clone.polygons[i], clone.polygons[j]) < 0
	})
	return clone
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

	hasZ := mp.Coordinates().HasZ()
	hasM := mp.Coordinates().HasM()

	var sb strings.Builder
	sb.WriteString("MULTIPOLYGON ")
	if hasZ && hasM {
		sb.WriteString("ZM ")
	} else if hasZ {
		sb.WriteString("Z ")
	} else if hasM {
		sb.WriteString("M ")
	}
	sb.WriteString("(")
	for i, p := range mp.polygons {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("(")
		sb.WriteString(ringCoordsToString(p.shell.coords, hasZ, hasM))
		for _, hole := range p.holes {
			sb.WriteString(", ")
			sb.WriteString(ringCoordsToString(hole.coords, hasZ, hasM))
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
