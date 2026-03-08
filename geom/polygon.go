package geom

import (
	"fmt"
	"math"
	"strings"
)

// LinearRing is a closed, simple LineString that forms the boundary of a polygon.
// A valid LinearRing has at least 4 coordinates (including the closing point),
// where the first and last coordinates are equal. When used as a polygon shell,
// it should be counter-clockwise; when used as a hole, clockwise.
// NewLinearRing automatically closes the ring if needed.
type LinearRing struct {
	*LineString
}

// NewLinearRing creates a new LinearRing from coordinates.
// If the ring is not closed, the first coordinate is appended automatically.
func NewLinearRing(coords CoordinateSequence) *LinearRing {
	// Ensure the ring is closed
	if len(coords) > 0 && !coords.IsClosed(DefaultEpsilon) {
		coords = append(coords.Clone(), coords[0].Clone())
	} else {
		coords = coords.Clone()
	}

	return &LinearRing{
		LineString: &LineString{
			coords: coords,
		},
	}
}

// NewLinearRingXY creates a LinearRing from x,y pairs.
func NewLinearRingXY(values ...float64) *LinearRing {
	return NewLinearRing(NewCoordinateSequenceXY(values...))
}

// NewLinearRingEmpty creates an empty LinearRing.
func NewLinearRingEmpty() *LinearRing {
	return &LinearRing{
		LineString: NewLineStringEmpty(),
	}
}

// GeometryType returns "LinearRing".
func (lr *LinearRing) GeometryType() string {
	return "LinearRing"
}

// Clone returns a deep copy.
func (lr *LinearRing) Clone() Geometry {
	return &LinearRing{
		LineString: lr.LineString.Clone().(*LineString),
	}
}

// IsValid returns true if the ring is valid.
// A valid ring has at least 4 points (including closure), is closed,
// and has no self-intersections.
func (lr *LinearRing) IsValid() bool {
	if lr.IsEmpty() {
		return true
	}
	if len(lr.coords) < 4 {
		return false
	}
	if !lr.IsClosed() {
		return false
	}
	// Check for self-intersection
	if hasRingSelfIntersection(lr.coords) {
		return false
	}
	return true
}

// IsCCW returns true if the ring is counter-clockwise oriented.
func (lr *LinearRing) IsCCW() bool {
	return SignedArea(lr.coords) > 0
}

// IsCW returns true if the ring is clockwise oriented.
func (lr *LinearRing) IsCW() bool {
	return SignedArea(lr.coords) < 0
}

// Area returns the absolute area of the ring.
func (lr *LinearRing) Area() float64 {
	return math.Abs(SignedArea(lr.coords))
}

// SignedArea computes the signed area of a ring.
// Positive if counter-clockwise, negative if clockwise.
func SignedArea(coords CoordinateSequence) float64 {
	if len(coords) < 3 {
		return 0
	}

	sum := 0.0
	n := len(coords)
	for i := 0; i < n-1; i++ {
		sum += coords[i].X*coords[i+1].Y - coords[i+1].X*coords[i].Y
	}
	if !coords.IsClosed(DefaultEpsilon) {
		sum += coords[n-1].X*coords[0].Y - coords[0].X*coords[n-1].Y
	}
	return sum / 2
}

// Reverse returns a new ring with reversed winding order.
func (lr *LinearRing) Reverse() *LinearRing {
	return &LinearRing{
		LineString: lr.LineString.Reverse(),
	}
}

// Normalize normalizes the ring to canonical form.
func (lr *LinearRing) Normalize() {
	if lr.IsEmpty() || len(lr.coords) < 4 {
		return
	}

	// Find the minimum coordinate
	minIdx := 0
	for i := 1; i < len(lr.coords)-1; i++ { // Exclude the closing point
		if lr.coords[i].X < lr.coords[minIdx].X ||
			(lr.coords[i].X == lr.coords[minIdx].X && lr.coords[i].Y < lr.coords[minIdx].Y) {
			minIdx = i
		}
	}

	// Rotate the ring so the minimum point is first
	if minIdx > 0 {
		n := len(lr.coords) - 1 // Exclude closing point
		newCoords := make(CoordinateSequence, n+1)
		for i := 0; i < n; i++ {
			newCoords[i] = lr.coords[(i+minIdx)%n].Clone()
		}
		newCoords[n] = newCoords[0].Clone()
		lr.coords = newCoords
	}
}

// String returns the WKT representation.
func (lr *LinearRing) String() string {
	if lr.IsEmpty() {
		return "LINEARRING EMPTY"
	}
	return strings.Replace(lr.LineString.String(), "LINESTRING", "LINEARRING", 1)
}

// Polygon represents a planar surface defined by an exterior ring and zero or more holes.
type Polygon struct {
	baseGeometry
	shell *LinearRing
	holes []*LinearRing
}

// NewPolygon creates a new Polygon with an exterior ring and optional holes.
// The shell and holes are cloned to prevent external mutation.
func NewPolygon(shell *LinearRing, holes []*LinearRing) *Polygon {
	var clonedShell *LinearRing
	if shell != nil {
		clonedShell = shell.Clone().(*LinearRing)
	}
	clonedHoles := make([]*LinearRing, len(holes))
	for i, h := range holes {
		if h != nil {
			clonedHoles[i] = h.Clone().(*LinearRing)
		}
	}
	return &Polygon{
		shell: clonedShell,
		holes: clonedHoles,
	}
}

// NewPolygonFromCoords creates a Polygon from coordinate sequences.
func NewPolygonFromCoords(shell CoordinateSequence, holes ...CoordinateSequence) *Polygon {
	shellRing := NewLinearRing(shell)
	holeRings := make([]*LinearRing, len(holes))
	for i, h := range holes {
		holeRings[i] = NewLinearRing(h)
	}
	return NewPolygon(shellRing, holeRings)
}

// NewPolygonEmpty creates an empty Polygon.
func NewPolygonEmpty() *Polygon {
	return &Polygon{
		shell: NewLinearRingEmpty(),
		holes: []*LinearRing{},
	}
}

// GeometryType returns "Polygon".
func (p *Polygon) GeometryType() string {
	return "Polygon"
}

// Envelope returns the bounding box.
func (p *Polygon) Envelope() *Envelope {
	if p.envelope == nil {
		if p.shell == nil || p.shell.IsEmpty() {
			p.envelope = NewEnvelopeEmpty()
		} else {
			p.envelope = p.shell.Envelope()
		}
	}
	return p.envelope.Clone()
}

// IsEmpty returns true if the polygon has no shell.
func (p *Polygon) IsEmpty() bool {
	return p.shell == nil || p.shell.IsEmpty()
}

// IsSimple returns true (polygons are always simple by definition).
func (p *Polygon) IsSimple() bool {
	return true
}

// IsValid returns true if the polygon is valid.
// A valid polygon has:
// - A valid shell with counter-clockwise orientation
// - Valid holes with clockwise orientation
// - Holes inside the shell
// - No shell/hole crossings
// - No nested holes
// - No crossing holes
func (p *Polygon) IsValid() bool {
	if p.IsEmpty() {
		return true
	}

	// Check shell is valid
	if !p.shell.IsValid() {
		return false
	}

	// Check shell has correct orientation (counter-clockwise)
	if !p.shell.IsCCW() {
		return false
	}

	// Check holes are valid and have correct orientation (clockwise)
	for _, hole := range p.holes {
		if !hole.IsValid() {
			return false
		}
		if !hole.IsCW() {
			return false
		}

		// Check hole is inside shell
		if !isRingInsideRing(hole, p.shell) {
			return false
		}

		// Check shell and hole don't cross
		if ringsProperlyIntersect(p.shell, hole) {
			return false
		}
	}

	// Check holes don't nest within each other and don't cross
	for i := 0; i < len(p.holes); i++ {
		for j := i + 1; j < len(p.holes); j++ {
			// Check for nested holes
			if isRingInsideRing(p.holes[i], p.holes[j]) || isRingInsideRing(p.holes[j], p.holes[i]) {
				return false
			}
			// Check for crossing holes
			if ringsProperlyIntersect(p.holes[i], p.holes[j]) {
				return false
			}
		}
	}

	return true
}

// Dimension returns 2 for Polygon.
func (p *Polygon) Dimension() Dimension {
	return DimensionArea
}

// Boundary returns the boundary (exterior ring + holes as MultiLineString).
func (p *Polygon) Boundary() Geometry {
	if p.IsEmpty() {
		return NewMultiLineStringEmpty()
	}

	rings := make([]*LineString, 1+len(p.holes))
	rings[0] = p.shell.LineString
	for i, hole := range p.holes {
		rings[i+1] = hole.LineString
	}

	return NewMultiLineString(rings)
}

// Coordinates returns all coordinates (shell + holes).
func (p *Polygon) Coordinates() CoordinateSequence {
	if p.IsEmpty() {
		return CoordinateSequence{}
	}

	total := len(p.shell.coords)
	for _, hole := range p.holes {
		total += len(hole.coords)
	}

	coords := make(CoordinateSequence, 0, total)
	coords = append(coords, p.shell.coords...)
	for _, hole := range p.holes {
		coords = append(coords, hole.coords...)
	}

	return coords
}

// ApplyCoordinateFilter applies a coordinate filter to the polygon.
func (p *Polygon) ApplyCoordinateFilter(filter CoordinateFilter) {
	if p.IsEmpty() || filter == nil {
		return
	}
	p.shell.ApplyCoordinateFilter(filter)
	for _, hole := range p.holes {
		hole.ApplyCoordinateFilter(filter)
	}
	p.invalidateEnvelope()
}

// NumGeometries returns 1 for Polygon.
func (p *Polygon) NumGeometries() int {
	return 1
}

// GeometryN returns the polygon itself (for n=0).
func (p *Polygon) GeometryN(n int) Geometry {
	if n != 0 {
		return nil
	}
	return p
}

// Clone returns a deep copy.
func (p *Polygon) Clone() Geometry {
	clone := NewPolygon(p.shell, p.holes)
	clone.srid = p.srid
	return clone
}

// Normalize normalizes the polygon to canonical form.
func (p *Polygon) Normalize() {
	if p.IsEmpty() {
		return
	}

	p.shell.Normalize()

	// Ensure counter-clockwise orientation for shell
	if p.shell.IsCW() {
		p.shell = p.shell.Reverse()
	}

	for i, hole := range p.holes {
		hole.Normalize()
		// Ensure clockwise orientation for holes
		if hole.IsCCW() {
			p.holes[i] = hole.Reverse()
		}
	}
}

// EqualsExact returns true if the polygons are exactly equal.
func (p *Polygon) EqualsExact(other Geometry, tolerance float64) bool {
	if other == nil {
		return false
	}
	otherPoly, ok := other.(*Polygon)
	if !ok {
		return false
	}
	if p.IsEmpty() && otherPoly.IsEmpty() {
		return true
	}
	if p.IsEmpty() || otherPoly.IsEmpty() {
		return false
	}

	if !p.shell.EqualsExact(otherPoly.shell, tolerance) {
		return false
	}

	if len(p.holes) != len(otherPoly.holes) {
		return false
	}

	for i, hole := range p.holes {
		if !hole.EqualsExact(otherPoly.holes[i], tolerance) {
			return false
		}
	}

	return true
}

// String returns the WKT representation.
func (p *Polygon) String() string {
	if p.IsEmpty() {
		return "POLYGON EMPTY"
	}

	hasZ := p.shell.coords.HasZ()
	hasM := p.shell.coords.HasM()

	var sb strings.Builder
	sb.WriteString("POLYGON ")

	if hasZ && hasM {
		sb.WriteString("ZM ")
	} else if hasZ {
		sb.WriteString("Z ")
	} else if hasM {
		sb.WriteString("M ")
	}

	sb.WriteString("(")
	sb.WriteString(ringCoordsToString(p.shell.coords, hasZ, hasM))
	for _, hole := range p.holes {
		sb.WriteString(", ")
		sb.WriteString(ringCoordsToString(hole.coords, hasZ, hasM))
	}
	sb.WriteString(")")

	return sb.String()
}

func ringCoordsToString(coords CoordinateSequence, hasZ, hasM bool) string {
	var sb strings.Builder
	sb.WriteString("(")
	for i, c := range coords {
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

// ExteriorRing returns the exterior ring.
func (p *Polygon) ExteriorRing() *LinearRing {
	return p.shell
}

// NumInteriorRings returns the number of holes.
func (p *Polygon) NumInteriorRings() int {
	return len(p.holes)
}

// InteriorRingN returns the nth hole (0-indexed).
func (p *Polygon) InteriorRingN(n int) *LinearRing {
	if n < 0 || n >= len(p.holes) {
		return nil
	}
	return p.holes[n]
}

// Area returns the area of the polygon (exterior - holes).
func (p *Polygon) Area() float64 {
	if p.IsEmpty() {
		return 0
	}

	area := p.shell.Area()
	for _, hole := range p.holes {
		area -= hole.Area()
	}
	return area
}

// Perimeter returns the total length of all rings.
func (p *Polygon) Perimeter() float64 {
	if p.IsEmpty() {
		return 0
	}

	perimeter := p.shell.Length()
	for _, hole := range p.holes {
		perimeter += hole.Length()
	}
	return perimeter
}

// Centroid returns the centroid of the polygon.
// The centroid is computed as the weighted average of ring centroids,
// where holes have negative contribution.
func (p *Polygon) Centroid() *Point {
	if p.IsEmpty() {
		return NewPointEmpty()
	}

	// Compute shell centroid and area
	shellCx, shellCy, shellArea := ringCentroidAndArea(p.shell.coords)

	totalCx := shellCx * shellArea
	totalCy := shellCy * shellArea
	totalArea := shellArea

	// Subtract hole contributions
	for _, hole := range p.holes {
		holeCx, holeCy, holeArea := ringCentroidAndArea(hole.coords)
		totalCx -= holeCx * holeArea
		totalCy -= holeCy * holeArea
		totalArea -= holeArea
	}

	if math.Abs(totalArea) < DefaultEpsilon {
		return NewPointFromCoordinate(p.shell.coords[0])
	}

	return NewPoint(totalCx/totalArea, totalCy/totalArea)
}

// RingCentroidAndArea computes the centroid and area of a closed ring.
// The input must include the closing coordinate; area is always positive.
func RingCentroidAndArea(coords CoordinateSequence) (cx, cy, area float64) {
	return ringCentroidAndArea(coords)
}

// ringCentroidAndArea computes the centroid and area of a ring using the shoelace formula.
// Returns (cx, cy, area) where area is always positive.
func ringCentroidAndArea(coords CoordinateSequence) (cx, cy, area float64) {
	n := len(coords) - 1 // Exclude closing point
	if n < 3 {
		return 0, 0, 0
	}

	sumX := 0.0
	sumY := 0.0
	signedArea := 0.0

	for i := 0; i < n; i++ {
		x0, y0 := coords[i].X, coords[i].Y
		x1, y1 := coords[(i+1)%n].X, coords[(i+1)%n].Y

		cross := x0*y1 - x1*y0
		signedArea += cross
		sumX += (x0 + x1) * cross
		sumY += (y0 + y1) * cross
	}

	signedArea /= 2
	if math.Abs(signedArea) < DefaultEpsilon {
		return coords[0].X, coords[0].Y, 0
	}

	cx = sumX / (6 * signedArea)
	cy = sumY / (6 * signedArea)
	area = math.Abs(signedArea)
	return cx, cy, area
}

// ContainsPoint returns true if the polygon contains the given coordinate.
func (p *Polygon) ContainsPoint(c Coordinate) bool {
	if p.IsEmpty() {
		return false
	}
	if !p.Envelope().Contains(c) {
		return false
	}
	loc := pointInPolygon(c, p)
	return loc == LocationInterior
}

// hasRingSelfIntersection checks if a ring's non-adjacent segments properly intersect.
func hasRingSelfIntersection(coords CoordinateSequence) bool {
	n := len(coords)
	for i := 0; i < n-1; i++ {
		for j := i + 2; j < n-1; j++ {
			// Skip first/last segment pair (they share closure point)
			if i == 0 && j == n-2 {
				continue
			}
			if segmentsIntersect(coords[i], coords[i+1], coords[j], coords[j+1]) {
				return true
			}
		}
	}
	return false
}

// ringsProperlyIntersect checks if two rings have a proper crossing.
func ringsProperlyIntersect(r1, r2 *LinearRing) bool {
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

// isRingInsideRing checks if inner ring is inside outer ring.
func isRingInsideRing(inner, outer *LinearRing) bool {
	if inner.IsEmpty() {
		return false
	}
	return PointInRing(inner.Coordinates()[0], outer)
}
