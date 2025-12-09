package geom

import (
	"math"
)

// Spatial predicate implementations for geometry types.
// These implement the OGC Simple Features spatial relationship predicates.

// Intersects returns true if the geometries share any portion of space.
func Intersects(g1, g2 Geometry) bool {
	// Quick rejection using envelopes
	if !g1.Envelope().Intersects(g2.Envelope()) {
		return false
	}

	return intersectsImpl(g1, g2)
}

func intersectsImpl(g1, g2 Geometry) bool {
	switch a := g1.(type) {
	case *Point:
		return pointIntersects(a, g2)
	case *LineString:
		return lineStringIntersects(a, g2)
	case *Polygon:
		return polygonIntersects(a, g2)
	case *MultiPoint:
		for i := 0; i < a.NumGeometries(); i++ {
			if pointIntersects(a.GeometryN(i).(*Point), g2) {
				return true
			}
		}
		return false
	case *MultiLineString:
		for i := 0; i < a.NumGeometries(); i++ {
			if lineStringIntersects(a.GeometryN(i).(*LineString), g2) {
				return true
			}
		}
		return false
	case *MultiPolygon:
		for i := 0; i < a.NumGeometries(); i++ {
			if polygonIntersects(a.GeometryN(i).(*Polygon), g2) {
				return true
			}
		}
		return false
	case *GeometryCollection:
		for i := 0; i < a.NumGeometries(); i++ {
			if intersectsImpl(a.GeometryN(i), g2) {
				return true
			}
		}
		return false
	}
	return false
}

func pointIntersects(p *Point, g Geometry) bool {
	if p.IsEmpty() {
		return false
	}
	switch b := g.(type) {
	case *Point:
		return !b.IsEmpty() && p.coord.Equals2D(b.coord, DefaultEpsilon)
	case *LineString:
		return pointOnLineString(p.coord, b)
	case *Polygon:
		return pointInPolygon(p.coord, b) != LocationExterior
	case *MultiPoint:
		for i := 0; i < b.NumGeometries(); i++ {
			if p.coord.Equals2D(b.GeometryN(i).(*Point).coord, DefaultEpsilon) {
				return true
			}
		}
		return false
	case *MultiLineString:
		for i := 0; i < b.NumGeometries(); i++ {
			if pointOnLineString(p.coord, b.GeometryN(i).(*LineString)) {
				return true
			}
		}
		return false
	case *MultiPolygon:
		for i := 0; i < b.NumGeometries(); i++ {
			if pointInPolygon(p.coord, b.GeometryN(i).(*Polygon)) != LocationExterior {
				return true
			}
		}
		return false
	case *GeometryCollection:
		for i := 0; i < b.NumGeometries(); i++ {
			if pointIntersects(p, b.GeometryN(i)) {
				return true
			}
		}
		return false
	}
	return false
}

func lineStringIntersects(ls *LineString, g Geometry) bool {
	if ls.IsEmpty() {
		return false
	}
	switch b := g.(type) {
	case *Point:
		return pointOnLineString(b.coord, ls)
	case *LineString:
		return lineStringsIntersect(ls, b)
	case *Polygon:
		return lineStringPolygonIntersect(ls, b)
	case *MultiPoint:
		for i := 0; i < b.NumGeometries(); i++ {
			if pointOnLineString(b.GeometryN(i).(*Point).coord, ls) {
				return true
			}
		}
		return false
	case *MultiLineString:
		for i := 0; i < b.NumGeometries(); i++ {
			if lineStringsIntersect(ls, b.GeometryN(i).(*LineString)) {
				return true
			}
		}
		return false
	case *MultiPolygon:
		for i := 0; i < b.NumGeometries(); i++ {
			if lineStringPolygonIntersect(ls, b.GeometryN(i).(*Polygon)) {
				return true
			}
		}
		return false
	case *GeometryCollection:
		for i := 0; i < b.NumGeometries(); i++ {
			if lineStringIntersects(ls, b.GeometryN(i)) {
				return true
			}
		}
		return false
	}
	return false
}

func polygonIntersects(p *Polygon, g Geometry) bool {
	if p.IsEmpty() {
		return false
	}
	switch b := g.(type) {
	case *Point:
		return pointInPolygon(b.coord, p) != LocationExterior
	case *LineString:
		return lineStringPolygonIntersect(b, p)
	case *Polygon:
		return polygonsIntersect(p, b)
	case *MultiPoint:
		for i := 0; i < b.NumGeometries(); i++ {
			if pointInPolygon(b.GeometryN(i).(*Point).coord, p) != LocationExterior {
				return true
			}
		}
		return false
	case *MultiLineString:
		for i := 0; i < b.NumGeometries(); i++ {
			if lineStringPolygonIntersect(b.GeometryN(i).(*LineString), p) {
				return true
			}
		}
		return false
	case *MultiPolygon:
		for i := 0; i < b.NumGeometries(); i++ {
			if polygonsIntersect(p, b.GeometryN(i).(*Polygon)) {
				return true
			}
		}
		return false
	case *GeometryCollection:
		for i := 0; i < b.NumGeometries(); i++ {
			if polygonIntersects(p, b.GeometryN(i)) {
				return true
			}
		}
		return false
	}
	return false
}

// Contains returns true if g1 completely contains g2.
func Contains(g1, g2 Geometry) bool {
	// Quick rejection using envelopes
	if !g1.Envelope().ContainsEnvelope(g2.Envelope()) {
		return false
	}
	return containsImpl(g1, g2)
}

func containsImpl(g1, g2 Geometry) bool {
	// All points of g2 must be in the interior or boundary of g1
	// AND at least one point of g2 must be in the interior of g1
	coords2 := g2.Coordinates()
	if len(coords2) == 0 {
		return true // Empty geometry is contained by everything
	}

	hasInterior := false
	for _, c := range coords2 {
		loc := locatePointIn(c, g1)
		if loc == LocationExterior {
			return false
		}
		if loc == LocationInterior {
			hasInterior = true
		}
	}

	// For proper contains, need interior intersection
	// For polygons containing lines/polygons, check edges too
	if !hasInterior {
		// Check if any point is in interior
		switch a := g1.(type) {
		case *Polygon:
			// Check centroid or midpoints
			centroid := g2.Envelope().Centre()
			if locatePointIn(centroid, a) == LocationInterior {
				hasInterior = true
			}
		}
	}

	return hasInterior
}

// Within returns true if g1 is completely within g2.
func Within(g1, g2 Geometry) bool {
	return Contains(g2, g1)
}

// Covers returns true if g1 covers g2 (no point of g2 is outside g1).
func Covers(g1, g2 Geometry) bool {
	if !g1.Envelope().ContainsEnvelope(g2.Envelope()) {
		return false
	}

	coords2 := g2.Coordinates()
	for _, c := range coords2 {
		if locatePointIn(c, g1) == LocationExterior {
			return false
		}
	}
	return true
}

// CoveredBy returns true if g1 is covered by g2.
func CoveredBy(g1, g2 Geometry) bool {
	return Covers(g2, g1)
}

// Disjoint returns true if the geometries have no point in common.
func Disjoint(g1, g2 Geometry) bool {
	return !Intersects(g1, g2)
}

// Touches returns true if the geometries touch at their boundaries only.
func Touches(g1, g2 Geometry) bool {
	if !g1.Envelope().Intersects(g2.Envelope()) {
		return false
	}

	// Must have boundary intersection but no interior intersection
	hasCommonPoint := false
	hasInteriorIntersection := false

	coords1 := g1.Coordinates()
	coords2 := g2.Coordinates()

	for _, c := range coords1 {
		loc := locatePointIn(c, g2)
		if loc == LocationBoundary {
			hasCommonPoint = true
		} else if loc == LocationInterior {
			hasInteriorIntersection = true
			break
		}
	}

	if hasInteriorIntersection {
		return false
	}

	for _, c := range coords2 {
		loc := locatePointIn(c, g1)
		if loc == LocationBoundary {
			hasCommonPoint = true
		} else if loc == LocationInterior {
			hasInteriorIntersection = true
			break
		}
	}

	if hasInteriorIntersection {
		return false
	}

	// For polygons, also check if boundaries intersect in a way that causes interior overlap
	// Two overlapping polygons have interior-interior intersection even if no vertex is in the other's interior
	if g1.Dimension() == DimensionArea && g2.Dimension() == DimensionArea {
		if hasPolygonInteriorIntersection(g1, g2) {
			return false
		}
	}

	return hasCommonPoint
}

// Crosses returns true if the geometries cross each other.
func Crosses(g1, g2 Geometry) bool {
	if !g1.Envelope().Intersects(g2.Envelope()) {
		return false
	}

	dim1 := g1.Dimension()
	dim2 := g2.Dimension()

	// Point/Point and Area/Area cannot cross
	if dim1 == dim2 && (dim1 == DimensionPoint || dim1 == DimensionArea) {
		return false
	}

	// Line/Line: must have point intersection but not share a line segment
	if dim1 == DimensionLine && dim2 == DimensionLine {
		return linesCross(g1, g2)
	}

	// Point/Line or Line/Point: point must be in interior of line
	if (dim1 == DimensionPoint && dim2 == DimensionLine) ||
		(dim1 == DimensionLine && dim2 == DimensionPoint) {
		return false // Points don't cross lines
	}

	// Line/Area: line must pass through interior and exterior
	if dim1 == DimensionLine && dim2 == DimensionArea {
		return lineCrossesArea(g1, g2)
	}
	if dim1 == DimensionArea && dim2 == DimensionLine {
		return lineCrossesArea(g2, g1)
	}

	return false
}

// Overlaps returns true if the geometries overlap.
func Overlaps(g1, g2 Geometry) bool {
	if !g1.Envelope().Intersects(g2.Envelope()) {
		return false
	}

	dim1 := g1.Dimension()
	dim2 := g2.Dimension()

	// Must have same dimension
	if dim1 != dim2 {
		return false
	}

	// Must intersect
	if !Intersects(g1, g2) {
		return false
	}

	// Neither must contain the other
	if Contains(g1, g2) || Contains(g2, g1) {
		return false
	}

	return true
}

// Equals returns true if the geometries are topologically equal.
func Equals(g1, g2 Geometry) bool {
	if g1.GeometryType() != g2.GeometryType() {
		return false
	}

	if g1.IsEmpty() && g2.IsEmpty() {
		return true
	}

	if g1.IsEmpty() || g2.IsEmpty() {
		return false
	}

	// Check if each covers the other
	return Covers(g1, g2) && Covers(g2, g1)
}

// Helper functions

func pointOnLineString(p Coordinate, ls *LineString) bool {
	coords := ls.Coordinates()
	for i := 1; i < len(coords); i++ {
		if pointOnSegment(p, coords[i-1], coords[i]) {
			return true
		}
	}
	return false
}

func pointOnSegment(p, a, b Coordinate) bool {
	// Check collinearity
	cross := (p.X-a.X)*(b.Y-a.Y) - (p.Y-a.Y)*(b.X-a.X)
	if math.Abs(cross) > DefaultEpsilon {
		return false
	}

	// Check if within segment bounds
	if p.X < math.Min(a.X, b.X)-DefaultEpsilon || p.X > math.Max(a.X, b.X)+DefaultEpsilon {
		return false
	}
	if p.Y < math.Min(a.Y, b.Y)-DefaultEpsilon || p.Y > math.Max(a.Y, b.Y)+DefaultEpsilon {
		return false
	}

	return true
}

func pointInPolygon(p Coordinate, poly *Polygon) Location {
	if poly.IsEmpty() {
		return LocationExterior
	}

	// Check boundary first
	if pointOnRing(p, poly.shell) {
		return LocationBoundary
	}
	for _, hole := range poly.holes {
		if pointOnRing(p, hole) {
			return LocationBoundary
		}
	}

	// Check interior
	if !pointInRing(p, poly.shell) {
		return LocationExterior
	}

	// Check if in any hole
	for _, hole := range poly.holes {
		if pointInRing(p, hole) {
			return LocationExterior
		}
	}

	return LocationInterior
}

func pointOnRing(p Coordinate, ring *LinearRing) bool {
	coords := ring.Coordinates()
	for i := 1; i < len(coords); i++ {
		if pointOnSegment(p, coords[i-1], coords[i]) {
			return true
		}
	}
	return false
}

func pointInRing(p Coordinate, ring *LinearRing) bool {
	coords := ring.Coordinates()
	n := len(coords)
	if n < 4 {
		return false
	}

	inside := false
	j := n - 2

	for i := 0; i < n-1; i++ {
		xi, yi := coords[i].X, coords[i].Y
		xj, yj := coords[j].X, coords[j].Y

		if ((yi > p.Y) != (yj > p.Y)) &&
			(p.X < (xj-xi)*(p.Y-yi)/(yj-yi)+xi) {
			inside = !inside
		}
		j = i
	}

	return inside
}

func locatePointIn(p Coordinate, g Geometry) Location {
	switch v := g.(type) {
	case *Point:
		if v.IsEmpty() {
			return LocationExterior
		}
		if p.Equals2D(v.coord, DefaultEpsilon) {
			return LocationInterior
		}
		return LocationExterior
	case *LineString:
		if v.IsEmpty() {
			return LocationExterior
		}
		coords := v.Coordinates()
		// Check endpoints (boundary)
		if p.Equals2D(coords[0], DefaultEpsilon) || p.Equals2D(coords[len(coords)-1], DefaultEpsilon) {
			if !v.IsClosed() {
				return LocationBoundary
			}
		}
		// Check interior
		if pointOnLineString(p, v) {
			return LocationInterior
		}
		return LocationExterior
	case *Polygon:
		return pointInPolygon(p, v)
	case *MultiPoint:
		for i := 0; i < v.NumGeometries(); i++ {
			if p.Equals2D(v.GeometryN(i).(*Point).coord, DefaultEpsilon) {
				return LocationInterior
			}
		}
		return LocationExterior
	case *MultiLineString:
		for i := 0; i < v.NumGeometries(); i++ {
			loc := locatePointIn(p, v.GeometryN(i))
			if loc != LocationExterior {
				return loc
			}
		}
		return LocationExterior
	case *MultiPolygon:
		for i := 0; i < v.NumGeometries(); i++ {
			loc := locatePointIn(p, v.GeometryN(i))
			if loc != LocationExterior {
				return loc
			}
		}
		return LocationExterior
	case *GeometryCollection:
		for i := 0; i < v.NumGeometries(); i++ {
			loc := locatePointIn(p, v.GeometryN(i))
			if loc != LocationExterior {
				return loc
			}
		}
		return LocationExterior
	}
	return LocationExterior
}

func lineStringsIntersect(ls1, ls2 *LineString) bool {
	coords1 := ls1.Coordinates()
	coords2 := ls2.Coordinates()

	for i := 1; i < len(coords1); i++ {
		for j := 1; j < len(coords2); j++ {
			if segmentsIntersect(coords1[i-1], coords1[i], coords2[j-1], coords2[j]) {
				return true
			}
		}
	}
	return false
}

func segmentsIntersect(a1, a2, b1, b2 Coordinate) bool {
	o1 := orientation(a1, a2, b1)
	o2 := orientation(a1, a2, b2)
	o3 := orientation(b1, b2, a1)
	o4 := orientation(b1, b2, a2)

	if o1 != o2 && o3 != o4 {
		return true
	}

	if o1 == 0 && onSegmentBounds(a1, b1, a2) {
		return true
	}
	if o2 == 0 && onSegmentBounds(a1, b2, a2) {
		return true
	}
	if o3 == 0 && onSegmentBounds(b1, a1, b2) {
		return true
	}
	if o4 == 0 && onSegmentBounds(b1, a2, b2) {
		return true
	}

	return false
}

func orientation(p, q, r Coordinate) int {
	val := (q.Y-p.Y)*(r.X-q.X) - (q.X-p.X)*(r.Y-q.Y)
	if math.Abs(val) < DefaultEpsilon {
		return 0
	}
	if val > 0 {
		return 1
	}
	return -1
}

func onSegmentBounds(p, q, r Coordinate) bool {
	return q.X <= math.Max(p.X, r.X) && q.X >= math.Min(p.X, r.X) &&
		q.Y <= math.Max(p.Y, r.Y) && q.Y >= math.Min(p.Y, r.Y)
}

func lineStringPolygonIntersect(ls *LineString, poly *Polygon) bool {
	coords := ls.Coordinates()

	// Check if any point is inside
	for _, c := range coords {
		if pointInPolygon(c, poly) != LocationExterior {
			return true
		}
	}

	// Check if any segment intersects polygon boundary
	shellCoords := poly.shell.Coordinates()
	for i := 1; i < len(coords); i++ {
		for j := 1; j < len(shellCoords); j++ {
			if segmentsIntersect(coords[i-1], coords[i], shellCoords[j-1], shellCoords[j]) {
				return true
			}
		}
	}

	for _, hole := range poly.holes {
		holeCoords := hole.Coordinates()
		for i := 1; i < len(coords); i++ {
			for j := 1; j < len(holeCoords); j++ {
				if segmentsIntersect(coords[i-1], coords[i], holeCoords[j-1], holeCoords[j]) {
					return true
				}
			}
		}
	}

	return false
}

func polygonsIntersect(p1, p2 *Polygon) bool {
	// Check if any vertex of p1 is inside p2
	for _, c := range p1.shell.Coordinates() {
		if pointInPolygon(c, p2) != LocationExterior {
			return true
		}
	}

	// Check if any vertex of p2 is inside p1
	for _, c := range p2.shell.Coordinates() {
		if pointInPolygon(c, p1) != LocationExterior {
			return true
		}
	}

	// Check if boundaries intersect
	shell1 := p1.shell.Coordinates()
	shell2 := p2.shell.Coordinates()

	for i := 1; i < len(shell1); i++ {
		for j := 1; j < len(shell2); j++ {
			if segmentsIntersect(shell1[i-1], shell1[i], shell2[j-1], shell2[j]) {
				return true
			}
		}
	}

	return false
}

func linesCross(g1, g2 Geometry) bool {
	// Lines cross if they have a point intersection but don't overlap
	ls1 := getLineStrings(g1)
	ls2 := getLineStrings(g2)

	for _, l1 := range ls1 {
		coords1 := l1.Coordinates()
		for _, l2 := range ls2 {
			coords2 := l2.Coordinates()
			for i := 1; i < len(coords1); i++ {
				for j := 1; j < len(coords2); j++ {
					if segmentsCross(coords1[i-1], coords1[i], coords2[j-1], coords2[j]) {
						return true
					}
				}
			}
		}
	}
	return false
}

func segmentsCross(a1, a2, b1, b2 Coordinate) bool {
	o1 := orientation(a1, a2, b1)
	o2 := orientation(a1, a2, b2)
	o3 := orientation(b1, b2, a1)
	o4 := orientation(b1, b2, a2)

	// Proper crossing (not at endpoints, not collinear)
	return o1 != o2 && o3 != o4 && o1 != 0 && o2 != 0 && o3 != 0 && o4 != 0
}

func lineCrossesArea(lineGeom, areaGeom Geometry) bool {
	lines := getLineStrings(lineGeom)
	polys := getPolygons(areaGeom)

	for _, ls := range lines {
		coords := ls.Coordinates()
		for _, poly := range polys {
			hasInterior := false
			hasExterior := false

			// Check endpoints
			for _, c := range coords {
				loc := pointInPolygon(c, poly)
				if loc == LocationInterior {
					hasInterior = true
				} else if loc == LocationExterior {
					hasExterior = true
				}
				if hasInterior && hasExterior {
					return true
				}
			}

			// If only checking endpoints didn't work, check if line crosses polygon boundary
			// which would indicate the line passes through interior from exterior
			if !hasInterior || !hasExterior {
				shellCoords := poly.shell.Coordinates()
				for i := 1; i < len(coords); i++ {
					for j := 1; j < len(shellCoords); j++ {
						if segmentsCross(coords[i-1], coords[i], shellCoords[j-1], shellCoords[j]) {
							// Line properly crosses polygon boundary = crosses interior
							hasInterior = true
							hasExterior = true
							return true
						}
					}
				}
			}
		}
	}
	return false
}

func getLineStrings(g Geometry) []*LineString {
	var result []*LineString
	switch v := g.(type) {
	case *LineString:
		result = append(result, v)
	case *MultiLineString:
		for i := 0; i < v.NumGeometries(); i++ {
			result = append(result, v.GeometryN(i).(*LineString))
		}
	case *GeometryCollection:
		for i := 0; i < v.NumGeometries(); i++ {
			result = append(result, getLineStrings(v.GeometryN(i))...)
		}
	}
	return result
}

func getPolygons(g Geometry) []*Polygon {
	var result []*Polygon
	switch v := g.(type) {
	case *Polygon:
		result = append(result, v)
	case *MultiPolygon:
		for i := 0; i < v.NumGeometries(); i++ {
			result = append(result, v.GeometryN(i).(*Polygon))
		}
	case *GeometryCollection:
		for i := 0; i < v.NumGeometries(); i++ {
			result = append(result, getPolygons(v.GeometryN(i))...)
		}
	}
	return result
}

// hasPolygonInteriorIntersection checks if two area geometries have interior-interior intersection
// This is needed for Touches to detect overlapping polygons even when no vertex is in the other's interior
func hasPolygonInteriorIntersection(g1, g2 Geometry) bool {
	polys1 := getPolygons(g1)
	polys2 := getPolygons(g2)

	for _, p1 := range polys1 {
		for _, p2 := range polys2 {
			// Check if a point inside p1 is also inside p2
			// Use centroid of overlap region or sample point
			env1 := p1.Envelope()
			env2 := p2.Envelope()

			if !env1.Intersects(env2) {
				continue
			}

			// Sample a point from the interior of the envelope intersection
			intersectEnv := NewEnvelope(
				math.Max(env1.MinX, env2.MinX),
				math.Max(env1.MinY, env2.MinY),
				math.Min(env1.MaxX, env2.MaxX),
				math.Min(env1.MaxY, env2.MaxY),
			)

			center := intersectEnv.Centre()
			loc1 := pointInPolygon(center, p1)
			loc2 := pointInPolygon(center, p2)

			if loc1 == LocationInterior && loc2 == LocationInterior {
				return true
			}

			// Also check by testing edges - if boundaries properly cross, there's interior overlap
			shell1 := p1.shell.Coordinates()
			shell2 := p2.shell.Coordinates()

			for i := 1; i < len(shell1); i++ {
				for j := 1; j < len(shell2); j++ {
					if segmentsCross(shell1[i-1], shell1[i], shell2[j-1], shell2[j]) {
						return true
					}
				}
			}
		}
	}
	return false
}
