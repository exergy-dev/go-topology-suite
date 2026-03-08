package geom

import (
	"math"
	"sort"
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
		// Check if any point from g2's interior is in g1's interior
		switch container := g1.(type) {
		case *Polygon:
			// Use an envelope sample point as a fallback for polygonal containers
			centroid := g2.Envelope().Centre()
			if locatePointIn(centroid, container) == LocationInterior {
				hasInterior = true
			}
		case *LineString:
			hasInterior = linealHasInteriorPointIn(container, g2)
		case *MultiLineString:
			hasInterior = linealHasInteriorPointIn(container, g2)
		}
	}

	if !hasInterior {
		return false
	}

	// After vertex checks pass, verify that no edge of g2 crosses the boundary of g1
	// A geometry can have all its vertices inside another geometry but still have edges that cross out
	if !edgesContainedIn(g2, g1) {
		return false
	}

	// Special handling for lineal containers (LineString, MultiLineString)
	// For a line to contain another geometry, segments must lie on the line
	switch g1.(type) {
	case *LineString, *MultiLineString:
		return linealContainsGeometry(g1, g2)
	}

	return true
}

// edgesContainedIn checks if all edges of g2 are contained within g1.
// This is necessary because vertices can all be inside while edges cross out.
func edgesContainedIn(g2, g1 Geometry) bool {
	switch container := g1.(type) {
	case *Polygon:
		return edgesContainedInPolygon(g2, container)
	case *MultiPolygon:
		// For MultiPolygon, we need to check if g2 is fully contained in any single polygon
		// or if it's properly distributed across them (more complex case)
		// For simplicity, check if edges don't cross to exterior
		return edgesContainedInMultiPolygon(g2, container)
	}
	// For other container types, vertex check is sufficient
	return true
}

// edgesContainedInPolygon checks if all edges of g2 are contained within the polygon
func edgesContainedInPolygon(g2 Geometry, poly *Polygon) bool {
	switch inner := g2.(type) {
	case *LineString:
		return lineStringEdgesContainedInPolygon(inner, poly)
	case *Polygon:
		return polygonEdgesContainedInPolygon(inner, poly)
	case *MultiLineString:
		for i := 0; i < inner.NumGeometries(); i++ {
			if !lineStringEdgesContainedInPolygon(inner.GeometryN(i).(*LineString), poly) {
				return false
			}
		}
		return true
	case *MultiPolygon:
		for i := 0; i < inner.NumGeometries(); i++ {
			if !polygonEdgesContainedInPolygon(inner.GeometryN(i).(*Polygon), poly) {
				return false
			}
		}
		return true
	case *GeometryCollection:
		for i := 0; i < inner.NumGeometries(); i++ {
			if !edgesContainedInPolygon(inner.GeometryN(i), poly) {
				return false
			}
		}
		return true
	}
	return true
}

// lineStringEdgesContainedInPolygon checks if all edges of a linestring stay within the polygon
func lineStringEdgesContainedInPolygon(ls *LineString, poly *Polygon) bool {
	if ls.IsEmpty() {
		return true
	}
	coords := ls.Coordinates()
	if len(coords) < 2 {
		return true
	}

	// Check each segment of the linestring
	for i := 1; i < len(coords); i++ {
		if !segmentContainedInPolygon(coords[i-1], coords[i], poly) {
			return false
		}
	}
	return true
}

// polygonEdgesContainedInPolygon checks if all edges of the inner polygon stay within the container
func polygonEdgesContainedInPolygon(inner, container *Polygon) bool {
	if inner.IsEmpty() {
		return true
	}

	// Check shell edges
	shellCoords := inner.shell.Coordinates()
	for i := 1; i < len(shellCoords); i++ {
		if !segmentContainedInPolygon(shellCoords[i-1], shellCoords[i], container) {
			return false
		}
	}

	// Check hole edges
	for _, hole := range inner.holes {
		holeCoords := hole.Coordinates()
		for i := 1; i < len(holeCoords); i++ {
			if !segmentContainedInPolygon(holeCoords[i-1], holeCoords[i], container) {
				return false
			}
		}
	}

	return true
}

// segmentContainedInPolygon checks if a segment is fully contained within the polygon
// A segment is contained if:
// 1. It doesn't properly cross the shell (exit the polygon)
// 2. It doesn't enter any hole
func segmentContainedInPolygon(a, b Coordinate, poly *Polygon) bool {
	// Check against shell - segment should not properly cross out
	shellCoords := poly.shell.Coordinates()
	for i := 1; i < len(shellCoords); i++ {
		if segmentsCrossProper(a, b, shellCoords[i-1], shellCoords[i]) {
			return false
		}
	}

	// Check against holes - segment should not enter any hole
	for _, hole := range poly.holes {
		holeCoords := hole.Coordinates()
		// First check if segment crosses hole boundary
		for i := 1; i < len(holeCoords); i++ {
			if segmentsCrossProper(a, b, holeCoords[i-1], holeCoords[i]) {
				return false
			}
		}
		// Also check if midpoint of segment is inside the hole
		mid := Coordinate{X: (a.X + b.X) / 2, Y: (a.Y + b.Y) / 2}
		if PointInRing(mid, hole) && !PointOnRing(mid, hole) {
			return false
		}
	}

	return true
}

// segmentsCrossProper returns true if segments properly cross (not at endpoints, not collinear)
// This is similar to segmentsCross but we need to ensure it's detecting proper crossings
func segmentsCrossProper(a1, a2, b1, b2 Coordinate) bool {
	o1 := orientation(a1, a2, b1)
	o2 := orientation(a1, a2, b2)
	o3 := orientation(b1, b2, a1)
	o4 := orientation(b1, b2, a2)

	// Proper crossing: opposite orientations on both segments
	// and none of the orientations are collinear (0)
	if o1 != o2 && o3 != o4 && o1 != 0 && o2 != 0 && o3 != 0 && o4 != 0 {
		return true
	}

	return false
}

// edgesContainedInMultiPolygon checks if edges are contained in the multipolygon
func edgesContainedInMultiPolygon(g2 Geometry, mp *MultiPolygon) bool {
	// For each edge of g2, check that it doesn't cross to exterior of all polygons
	// This is a simplified check - a more rigorous approach would track which
	// polygon each part of g2 is in
	switch inner := g2.(type) {
	case *LineString:
		return lineStringEdgesContainedInMultiPolygon(inner, mp)
	case *Polygon:
		return polygonEdgesContainedInMultiPolygon(inner, mp)
	case *MultiLineString:
		for i := 0; i < inner.NumGeometries(); i++ {
			if !lineStringEdgesContainedInMultiPolygon(inner.GeometryN(i).(*LineString), mp) {
				return false
			}
		}
		return true
	case *MultiPolygon:
		for i := 0; i < inner.NumGeometries(); i++ {
			if !polygonEdgesContainedInMultiPolygon(inner.GeometryN(i).(*Polygon), mp) {
				return false
			}
		}
		return true
	case *GeometryCollection:
		for i := 0; i < inner.NumGeometries(); i++ {
			if !edgesContainedInMultiPolygon(inner.GeometryN(i), mp) {
				return false
			}
		}
		return true
	}
	return true
}

// lineStringEdgesContainedInMultiPolygon checks each segment
func lineStringEdgesContainedInMultiPolygon(ls *LineString, mp *MultiPolygon) bool {
	if ls.IsEmpty() {
		return true
	}
	coords := ls.Coordinates()
	if len(coords) < 2 {
		return true
	}

	for i := 1; i < len(coords); i++ {
		if !segmentContainedInMultiPolygon(coords[i-1], coords[i], mp) {
			return false
		}
	}
	return true
}

// polygonEdgesContainedInMultiPolygon checks shell and hole edges
func polygonEdgesContainedInMultiPolygon(inner *Polygon, mp *MultiPolygon) bool {
	if inner.IsEmpty() {
		return true
	}

	shellCoords := inner.shell.Coordinates()
	for i := 1; i < len(shellCoords); i++ {
		if !segmentContainedInMultiPolygon(shellCoords[i-1], shellCoords[i], mp) {
			return false
		}
	}

	for _, hole := range inner.holes {
		holeCoords := hole.Coordinates()
		for i := 1; i < len(holeCoords); i++ {
			if !segmentContainedInMultiPolygon(holeCoords[i-1], holeCoords[i], mp) {
				return false
			}
		}
	}

	return true
}

// segmentContainedInMultiPolygon checks if a segment is contained in at least one polygon
// and doesn't cross out of the multipolygon
func segmentContainedInMultiPolygon(a, b Coordinate, mp *MultiPolygon) bool {
	// Check if segment is contained in any single polygon
	for i := 0; i < mp.NumGeometries(); i++ {
		poly := mp.GeometryN(i).(*Polygon)
		if segmentContainedInPolygon(a, b, poly) {
			return true
		}
	}

	// If not contained in any single polygon, check if the midpoint is inside the multipolygon
	// and the segment doesn't cross all polygon boundaries to exterior
	mid := Coordinate{X: (a.X + b.X) / 2, Y: (a.Y + b.Y) / 2}
	midInside := false
	for i := 0; i < mp.NumGeometries(); i++ {
		poly := mp.GeometryN(i).(*Polygon)
		if pointInPolygon(mid, poly) != LocationExterior {
			midInside = true
			break
		}
	}

	return midInside
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

	// Check that edges of g2 don't exit g1 (handles concave containers)
	if !edgesContainedIn(g2, g1) {
		return false
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

	// Must have boundary intersection but no interior-interior intersection
	hasCommonPoint := false
	hasInteriorInterior := false

	coords1 := g1.Coordinates()
	coords2 := g2.Coordinates()

	// Check coords of g1 against g2
	for _, c := range coords1 {
		locIn2 := locatePointIn(c, g2)
		if locIn2 == LocationBoundary {
			hasCommonPoint = true
		} else if locIn2 == LocationInterior {
			// c is in g2's interior. Check if c is also in g1's interior (not boundary)
			locIn1 := locatePointInSelf(c, g1)
			if locIn1 == LocationInterior {
				// Interior-interior intersection - NOT touches
				hasInteriorInterior = true
				break
			} else if locIn1 == LocationBoundary {
				// c is on g1's boundary but in g2's interior
				// This is boundary-interior intersection, which counts as touching
				hasCommonPoint = true
			}
		}
	}

	if hasInteriorInterior {
		return false
	}

	// Check coords of g2 against g1
	for _, c := range coords2 {
		locIn1 := locatePointIn(c, g1)
		if locIn1 == LocationBoundary {
			hasCommonPoint = true
		} else if locIn1 == LocationInterior {
			// c is in g1's interior. Check if c is also in g2's interior
			locIn2 := locatePointInSelf(c, g2)
			if locIn2 == LocationInterior {
				hasInteriorInterior = true
				break
			} else if locIn2 == LocationBoundary {
				// c is on g2's boundary but in g1's interior
				// This is boundary-interior intersection, not interior-interior
				// For line-polygon: polygon boundary point in line interior is OK for touches
				hasCommonPoint = true
			}
		}
	}

	if hasInteriorInterior {
		return false
	}

	// For polygons, also check if boundaries intersect in a way that causes interior overlap
	// Two overlapping polygons have interior-interior intersection even if no vertex is in the other's interior
	if g1.Dimension() == DimensionArea && g2.Dimension() == DimensionArea {
		if hasPolygonInteriorIntersection(g1, g2) {
			return false
		}
	}

	// For line-line and line-area, ensure there is no interior-interior intersection
	if g1.Dimension() == DimensionLine && g2.Dimension() == DimensionLine {
		if linesHaveInteriorIntersection(g1, g2) {
			return false
		}
	}
	if g1.Dimension() == DimensionLine && g2.Dimension() == DimensionArea {
		if lineIntersectsAreaInterior(g1, g2) {
			return false
		}
	}
	if g1.Dimension() == DimensionArea && g2.Dimension() == DimensionLine {
		if lineIntersectsAreaInterior(g2, g1) {
			return false
		}
	}

	// If no common point found via vertices, check for mid-segment boundary overlaps
	// This handles cases like a line overlapping a polygon edge
	if !hasCommonPoint {
		hasCommonPoint = boundariesIntersect(g1, g2)
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

	// Equal geometries do not overlap (OGC definition)
	if Equals(g1, g2) {
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
		if PointOnSegment(p, coords[i-1], coords[i]) {
			return true
		}
	}
	return false
}

// PointOnSegment reports whether p lies on segment (a,b), within DefaultEpsilon.
func PointOnSegment(p, a, b Coordinate) bool {
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
	if PointOnRing(p, poly.shell) {
		return LocationBoundary
	}
	for _, hole := range poly.holes {
		if PointOnRing(p, hole) {
			return LocationBoundary
		}
	}

	// Check interior
	if !PointInRing(p, poly.shell) {
		return LocationExterior
	}

	// Check if in any hole
	for _, hole := range poly.holes {
		if PointInRing(p, hole) {
			return LocationExterior
		}
	}

	return LocationInterior
}

// PointOnRing reports whether p lies on any segment of ring.
func PointOnRing(p Coordinate, ring *LinearRing) bool {
	coords := ring.coords
	for i := 1; i < len(coords); i++ {
		if PointOnSegment(p, coords[i-1], coords[i]) {
			return true
		}
	}
	return false
}

// PointInRing determines whether p is inside ring using ray casting.
func PointInRing(p Coordinate, ring *LinearRing) bool {
	coords := ring.coords
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

// locateInMulti dispatches location queries across multi-geometry children using the given locator function.
func locateInMulti(p Coordinate, g Geometry, locator func(Coordinate, Geometry) Location) Location {
	for i := 0; i < g.NumGeometries(); i++ {
		loc := locator(p, g.GeometryN(i))
		if loc != LocationExterior {
			return loc
		}
	}
	return LocationExterior
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
	case *LinearRing:
		// Closed ring: boundary is the ring itself, no boundary endpoints
		if PointOnRing(p, v) {
			return LocationBoundary
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
	case *MultiLineString, *MultiPolygon, *GeometryCollection:
		return locateInMulti(p, g, locatePointIn)
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
	return SegmentsIntersect(a1, a2, b1, b2)
}

func orientation(p, q, r Coordinate) int {
	return OrientationIndex(p, q, r)
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
				switch loc {
				case LocationInterior:
					hasInterior = true
				case LocationExterior:
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

// locatePointInSelf determines where a point lies within its own geometry
// Used to distinguish boundary vs interior points of a geometry
func locatePointInSelf(p Coordinate, g Geometry) Location {
	switch v := g.(type) {
	case *Point:
		if p.Equals2D(v.coord, DefaultEpsilon) {
			return LocationInterior // A point is its own interior
		}
		return LocationExterior
	case *LinearRing:
		if PointOnRing(p, v) {
			return LocationBoundary
		}
		return LocationExterior
	case *LineString:
		coords := v.Coordinates()
		if len(coords) < 2 {
			return LocationExterior
		}
		// Endpoints are boundary
		if p.Equals2D(coords[0], DefaultEpsilon) || p.Equals2D(coords[len(coords)-1], DefaultEpsilon) {
			if !v.IsClosed() {
				return LocationBoundary
			}
		}
		// Other points on line are interior
		if pointOnLineString(p, v) {
			return LocationInterior
		}
		return LocationExterior
	case *Polygon:
		// Polygon boundary is the rings
		if PointOnRing(p, v.shell) {
			return LocationBoundary
		}
		for _, hole := range v.holes {
			if PointOnRing(p, hole) {
				return LocationBoundary
			}
		}
		if PointInRing(p, v.shell) {
			// Check not in any hole
			for _, hole := range v.holes {
				if PointInRing(p, hole) {
					return LocationExterior
				}
			}
			return LocationInterior
		}
		return LocationExterior
	case *MultiPoint:
		for i := 0; i < v.NumGeometries(); i++ {
			if p.Equals2D(v.GeometryN(i).(*Point).coord, DefaultEpsilon) {
				return LocationInterior
			}
		}
		return LocationExterior
	case *MultiLineString, *MultiPolygon, *GeometryCollection:
		return locateInMulti(p, g, locatePointInSelf)
	}
	return LocationExterior
}

// boundariesIntersect checks if the boundaries of two geometries share any points
// This handles mid-segment overlaps that vertex checks miss
func boundariesIntersect(g1, g2 Geometry) bool {
	// Get boundary segments from each geometry
	segs1 := BoundarySegments(g1)
	segs2 := BoundarySegments(g2)

	// Check for segment-segment intersection
	for _, s1 := range segs1 {
		for _, s2 := range segs2 {
			if SegmentsIntersect(s1.P0, s1.P1, s2.P0, s2.P1) {
				return true
			}
		}
	}
	return false
}

// linealHasInteriorPointIn checks if any interior point of a lineal geometry is inside the container.
func linealHasInteriorPointIn(container Geometry, lineal Geometry) bool {
	lines := getLineStrings(lineal)
	for _, ls := range lines {
		coords := ls.Coordinates()
		for i := 1; i < len(coords); i++ {
			mid := Coordinate{X: (coords[i-1].X + coords[i].X) / 2, Y: (coords[i-1].Y + coords[i].Y) / 2}
			if locatePointIn(mid, container) == LocationInterior {
				return true
			}
		}
	}
	return false
}

// lineIntersectsAreaInterior returns true if a line has interior-interior intersection with an area.
func lineIntersectsAreaInterior(lineGeom, areaGeom Geometry) bool {
	if lineCrossesArea(lineGeom, areaGeom) {
		return true
	}
	return Contains(areaGeom, lineGeom)
}

// linesHaveInteriorIntersection returns true if two lineal geometries intersect in their interiors.
func linesHaveInteriorIntersection(g1, g2 Geometry) bool {
	ls1 := getLineStrings(g1)
	ls2 := getLineStrings(g2)

	for _, l1 := range ls1 {
		c1 := l1.Coordinates()
		for _, l2 := range ls2 {
			c2 := l2.Coordinates()
			for i := 1; i < len(c1); i++ {
				a1 := c1[i-1]
				a2 := c1[i]
				for j := 1; j < len(c2); j++ {
					b1 := c2[j-1]
					b2 := c2[j]
					info := segmentIntersectionInfo(a1, a2, b1, b2)
					if !info.intersects {
						continue
					}
					if info.proper || info.collinearOverlap {
						return true
					}
					for _, p := range info.points {
						if pointInSegmentInterior(p, a1, a2) && pointInSegmentInterior(p, b1, b2) {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

func pointInSegmentInterior(p, a, b Coordinate) bool {
	if !PointOnSegment(p, a, b) {
		return false
	}
	if p.Equals2D(a, DefaultEpsilon) || p.Equals2D(b, DefaultEpsilon) {
		return false
	}
	return true
}

// linealContainsGeometry checks if a lineal geometry contains another geometry.
func linealContainsGeometry(container Geometry, g Geometry) bool {
	lines := getLineStrings(container)
	if len(lines) == 0 {
		return false
	}
	switch inner := g.(type) {
	case *Point:
		return pointOnAnyLine(inner.coord, lines)
	case *LineString:
		return lineContainedByLines(lines, inner)
	case *MultiPoint:
		for i := 0; i < inner.NumGeometries(); i++ {
			if !pointOnAnyLine(inner.GeometryN(i).(*Point).coord, lines) {
				return false
			}
		}
		return true
	case *MultiLineString:
		for i := 0; i < inner.NumGeometries(); i++ {
			if !lineContainedByLines(lines, inner.GeometryN(i).(*LineString)) {
				return false
			}
		}
		return true
	}
	return false
}

func pointOnAnyLine(p Coordinate, lines []*LineString) bool {
	for _, line := range lines {
		if pointOnLineString(p, line) {
			return true
		}
	}
	return false
}

func lineContainedByLines(lines []*LineString, inner *LineString) bool {
	if inner.IsEmpty() {
		return true
	}
	innerCoords := inner.Coordinates()
	if len(innerCoords) < 2 {
		return len(innerCoords) == 0 || pointOnAnyLine(innerCoords[0], lines)
	}

	for i := 1; i < len(innerCoords); i++ {
		if !segmentCoveredByAnyLine(innerCoords[i-1], innerCoords[i], lines) {
			return false
		}
	}
	return true
}

func segmentCoveredByAnyLine(a, b Coordinate, lines []*LineString) bool {
	for _, line := range lines {
		if segmentCoveredByLine(a, b, line) {
			return true
		}
	}
	return false
}

// segmentCoveredByLine checks if segment (a,b) is fully covered by the line's segments
// The segment must be collinear with and overlap one or more consecutive segments of the line
func segmentCoveredByLine(a, b Coordinate, line *LineString) bool {
	coords := line.Coordinates()
	if len(coords) < 2 {
		return false
	}

	if a.Equals2D(b, DefaultEpsilon) {
		return pointOnLineString(a, line)
	}

	// Check if both endpoints are on the line
	if !pointOnLineString(a, line) || !pointOnLineString(b, line) {
		return false
	}

	dx := b.X - a.X
	dy := b.Y - a.Y
	useX := math.Abs(dx) >= math.Abs(dy)
	denom := dx
	if !useX {
		denom = dy
	}
	if math.Abs(denom) < DefaultEpsilon {
		return false
	}

	type interval struct {
		start float64
		end   float64
	}
	intervals := make([]interval, 0)

	for i := 1; i < len(coords); i++ {
		p := coords[i-1]
		q := coords[i]
		if orientation(a, b, p) != 0 || orientation(a, b, q) != 0 {
			continue
		}

		var t1, t2 float64
		if useX {
			t1 = (p.X - a.X) / denom
			t2 = (q.X - a.X) / denom
		} else {
			t1 = (p.Y - a.Y) / denom
			t2 = (q.Y - a.Y) / denom
		}

		start := math.Min(t1, t2)
		end := math.Max(t1, t2)
		if end < 0-DefaultEpsilon || start > 1+DefaultEpsilon {
			continue
		}
		if start < 0 {
			start = 0
		}
		if end > 1 {
			end = 1
		}
		intervals = append(intervals, interval{start: start, end: end})
	}

	if len(intervals) == 0 {
		return false
	}

	sort.Slice(intervals, func(i, j int) bool {
		return intervals[i].start < intervals[j].start
	})

	current := intervals[0]
	if current.start > DefaultEpsilon {
		return false
	}
	for i := 1; i < len(intervals); i++ {
		next := intervals[i]
		if next.start <= current.end+DefaultEpsilon {
			if next.end > current.end {
				current.end = next.end
			}
		} else {
			return false
		}
	}

	return current.end >= 1-DefaultEpsilon
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
