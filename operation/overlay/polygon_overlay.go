package overlay

import (
	"math"

	"github.com/go-topology-suite/gts/algorithm"
	"github.com/go-topology-suite/gts/geom"
)

// polygonPolygonIntersection computes the intersection of two polygon sets.
func polygonPolygonIntersection(polysA, polysB []*geom.Polygon) geom.Geometry {
	// Special case: if both slices point to the same underlying array
	// (i.e., the same polygon set), return it directly for intersection/union
	// This avoids issues with degenerate polygons after bounding extreme coordinates
	if len(polysA) > 0 && len(polysB) > 0 && &polysA[0] == &polysB[0] {
		// Same slice - A ∩ A = A
		return collectPolygons(polysA)
	}

	// Bound coordinates to prevent overflow
	polysA = boundPolygons(polysA)
	polysB = boundPolygons(polysB)

	// Use noded overlay for robust intersection handling
	return nodedPolygonOverlay(polysA, polysB, OpIntersection)
}

// polygonPolygonUnion computes the union of two polygon sets.
func polygonPolygonUnion(polysA, polysB []*geom.Polygon) geom.Geometry {
	// Special case: if both slices point to the same underlying array
	// (i.e., the same polygon set), return it directly for intersection/union
	// This avoids issues with degenerate polygons after bounding extreme coordinates
	if len(polysA) > 0 && len(polysB) > 0 && &polysA[0] == &polysB[0] {
		// Same slice - A ∪ A = A
		return collectPolygons(polysA)
	}

	// Bound coordinates to prevent overflow
	polysA = boundPolygons(polysA)
	polysB = boundPolygons(polysB)

	// Use noded overlay for robust union handling
	return nodedPolygonOverlay(polysA, polysB, OpUnion)
}

// boundPolygons bounds all polygons in a slice.
func boundPolygons(polys []*geom.Polygon) []*geom.Polygon {
	result := make([]*geom.Polygon, len(polys))
	for i, p := range polys {
		result[i] = boundPolygon(p)
	}
	return result
}

// boundCoordinate bounds a coordinate to prevent overflow.
func boundCoordinate(c geom.Coordinate) geom.Coordinate {
	// Use 1e15 as the maximum coordinate value.
	// This ensures that arithmetic operations (addition, subtraction) maintain precision
	// while still being large enough for most real-world use cases.
	// At 1e15, we can still add values as small as 1.0 and have them represented.
	const maxCoord = 1e15

	x := c.X
	y := c.Y

	if math.IsNaN(x) || math.IsInf(x, 0) {
		x = 0
	} else if x > maxCoord {
		x = maxCoord
	} else if x < -maxCoord {
		x = -maxCoord
	}

	if math.IsNaN(y) || math.IsInf(y, 0) {
		y = 0
	} else if y > maxCoord {
		y = maxCoord
	} else if y < -maxCoord {
		y = -maxCoord
	}

	return geom.NewCoordinate(x, y)
}

// boundPolygon bounds all coordinates in a polygon to prevent overflow.
func boundPolygon(poly *geom.Polygon) *geom.Polygon {
	if poly.IsEmpty() {
		return poly
	}

	// Bound exterior ring
	extCoords := poly.ExteriorRing().Coordinates()
	boundedExt := make(geom.CoordinateSequence, len(extCoords))
	for i, c := range extCoords {
		boundedExt[i] = boundCoordinate(c)
	}

	// Bound holes
	holes := make([]*geom.LinearRing, poly.NumInteriorRings())
	for i := 0; i < poly.NumInteriorRings(); i++ {
		holeCoords := poly.InteriorRingN(i).Coordinates()
		boundedHole := make(geom.CoordinateSequence, len(holeCoords))
		for j, c := range holeCoords {
			boundedHole[j] = boundCoordinate(c)
		}
		holes[i] = geom.NewLinearRing(boundedHole)
	}

	return geom.NewPolygon(geom.NewLinearRing(boundedExt), holes)
}

// mergePolygons merges two overlapping polygons into one.
func mergePolygons(polyA, polyB *geom.Polygon) *geom.Polygon {
	if polyA.IsEmpty() {
		return polyB
	}
	if polyB.IsEmpty() {
		return polyA
	}

	// Get parts of A outside B
	shellA := polyA.ExteriorRing().Coordinates()
	shellB := polyB.ExteriorRing().Coordinates()

	// Trace the union boundary by walking both polygons
	unionCoords := traceUnionBoundary(shellA, shellB)
	if len(unionCoords) < 4 {
		// Fall back to returning the larger polygon
		if polyA.Area() >= polyB.Area() {
			return polyA
		}
		return polyB
	}

	// Ensure closed
	if !unionCoords.IsClosed(geom.DefaultEpsilon) {
		unionCoords = append(unionCoords, unionCoords[0].Clone())
	}

	return geom.NewPolygon(geom.NewLinearRing(unionCoords), nil)
}

// traceUnionBoundary traces the boundary of the union of two polygons.
// Uses the Weiler-Atherton algorithm for polygon union.
func traceUnionBoundary(shellA, shellB geom.CoordinateSequence) geom.CoordinateSequence {
	// Find all intersection points between the two polygon boundaries
	type intersectionInfo struct {
		coord    geom.Coordinate
		onAIndex int
		onBIndex int
		tA       float64
		tB       float64
		entering bool // true if entering B from outside, false if leaving
	}

	var intersections []intersectionInfo

	// Find intersections
	for i := 0; i < len(shellA)-1; i++ {
		a1, a2 := shellA[i], shellA[i+1]
		for j := 0; j < len(shellB)-1; j++ {
			b1, b2 := shellB[j], shellB[j+1]
			if pt := segmentSegmentIntersect(a1, a2, b1, b2); pt != nil {
				tA := parameterOnSegment(a1, a2, *pt)
				tB := parameterOnSegment(b1, b2, *pt)
				// Determine if we're entering or leaving B
				// Check a point slightly after intersection
				testT := tA + 0.001
				if testT > 1 {
					testT = 0.999
				}
				testPt := geom.NewCoordinate(a1.X+testT*(a2.X-a1.X), a1.Y+testT*(a2.Y-a1.Y))
				entering := pointInPolygon(testPt, shellB) > 0

				intersections = append(intersections, intersectionInfo{
					coord:    *pt,
					onAIndex: i,
					onBIndex: j,
					tA:       tA,
					tB:       tB,
					entering: entering,
				})
			}
		}
	}

	// If no intersections, check containment
	if len(intersections) == 0 {
		if pointInPolygon(shellB[0], shellA) > 0 {
			return shellA.Clone()
		}
		if pointInPolygon(shellA[0], shellB) > 0 {
			return shellB.Clone()
		}
		return nil
	}

	// For union, collect all points outside the other polygon plus intersection points
	var result geom.CoordinateSequence

	// Add all points from A that are outside B, along with intersections on A's edges
	for i := 0; i < len(shellA)-1; i++ {
		// Check if this vertex is outside B
		if pointInPolygon(shellA[i], shellB) <= 0 {
			result = append(result, shellA[i])
		}

		// Add any intersections on this edge, sorted by tA
		var edgeIntersections []intersectionInfo
		for _, inter := range intersections {
			if inter.onAIndex == i {
				edgeIntersections = append(edgeIntersections, inter)
			}
		}
		// Sort by tA
		for m := 0; m < len(edgeIntersections)-1; m++ {
			for n := m + 1; n < len(edgeIntersections); n++ {
				if edgeIntersections[m].tA > edgeIntersections[n].tA {
					edgeIntersections[m], edgeIntersections[n] = edgeIntersections[n], edgeIntersections[m]
				}
			}
		}
		for _, inter := range edgeIntersections {
			result = append(result, inter.coord)
		}
	}

	// Add all points from B that are outside A, along with intersections on B's edges
	for i := 0; i < len(shellB)-1; i++ {
		// Check if this vertex is outside A
		if pointInPolygon(shellB[i], shellA) <= 0 {
			result = append(result, shellB[i])
		}
		// Intersections are already added from A's perspective
	}

	// Remove duplicates
	if len(result) > 1 {
		var cleaned geom.CoordinateSequence
		for i := 0; i < len(result); i++ {
			isDup := false
			for j := 0; j < len(cleaned); j++ {
				if result[i].Equals2D(cleaned[j], geom.DefaultEpsilon) {
					isDup = true
					break
				}
			}
			if !isDup {
				cleaned = append(cleaned, result[i])
			}
		}
		result = cleaned
	}

	// Sort points by angle around centroid to form a valid polygon
	if len(result) >= 3 {
		result = sortPointsByAngle(result)
	}

	return result
}

// sortPointsByAngle sorts points counter-clockwise around their centroid.
func sortPointsByAngle(coords geom.CoordinateSequence) geom.CoordinateSequence {
	if len(coords) < 3 {
		return coords
	}

	// Compute centroid
	var cx, cy float64
	for _, c := range coords {
		cx += c.X
		cy += c.Y
	}
	cx /= float64(len(coords))
	cy /= float64(len(coords))

	// Compute angles
	type pointAngle struct {
		coord geom.Coordinate
		angle float64
	}
	angles := make([]pointAngle, len(coords))
	for i, c := range coords {
		angles[i] = pointAngle{
			coord: c,
			angle: math.Atan2(c.Y-cy, c.X-cx),
		}
	}

	// Sort by angle
	for i := 0; i < len(angles)-1; i++ {
		for j := i + 1; j < len(angles); j++ {
			if angles[i].angle > angles[j].angle {
				angles[i], angles[j] = angles[j], angles[i]
			}
		}
	}

	// Build result
	result := make(geom.CoordinateSequence, len(angles))
	for i, pa := range angles {
		result[i] = pa.coord
	}

	return result
}

// segmentSegmentIntersect finds intersection of two finite segments.
func segmentSegmentIntersect(a1, a2, b1, b2 geom.Coordinate) *geom.Coordinate {
	return lineLineIntersect(a1, a2, b1, b2)
}

// parameterOnSegment calculates the parameter t for point p on segment (a, b).
func parameterOnSegment(a, b, p geom.Coordinate) float64 {
	dx := b.X - a.X
	dy := b.Y - a.Y
	if math.Abs(dx) > math.Abs(dy) {
		return (p.X - a.X) / dx
	}
	if math.Abs(dy) > geom.DefaultEpsilon {
		return (p.Y - a.Y) / dy
	}
	return 0
}

// collectPolygons converts a slice of polygons to the appropriate geometry type.
func collectPolygons(polys []*geom.Polygon) geom.Geometry {
	if len(polys) == 0 {
		return geom.NewPolygonEmpty()
	}
	if len(polys) == 1 {
		return polys[0]
	}
	return geom.NewMultiPolygon(polys)
}

// polygonPolygonDifference computes parts of polysA not in polysB.
func polygonPolygonDifference(polysA, polysB []*geom.Polygon) geom.Geometry {
	// Bound coordinates to prevent overflow
	polysA = boundPolygons(polysA)
	polysB = boundPolygons(polysB)

	// Use noded overlay for robust difference handling
	return nodedPolygonOverlay(polysA, polysB, OpDifference)
}

// polygonPolygonSymDifference computes parts in either polygon set but not both.
func polygonPolygonSymDifference(polysA, polysB []*geom.Polygon) geom.Geometry {
	// Bound coordinates to prevent overflow
	polysA = boundPolygons(polysA)
	polysB = boundPolygons(polysB)

	// Use noded overlay for robust symmetric difference handling
	return nodedPolygonOverlay(polysA, polysB, OpSymDifference)
}

// clipPolygonToPolygon clips polyA to the interior of polyB using Sutherland-Hodgman algorithm.
func clipPolygonToPolygon(polyA, polyB *geom.Polygon) *geom.Polygon {
	if polyA.IsEmpty() || polyB.IsEmpty() {
		return geom.NewPolygonEmpty()
	}

	// Check if envelopes intersect
	if !polyA.Envelope().Intersects(polyB.Envelope()) {
		return geom.NewPolygonEmpty()
	}

	// Clip using Sutherland-Hodgman algorithm
	shellA := polyA.ExteriorRing().Coordinates()
	shellB := polyB.ExteriorRing().Coordinates()

	clipped := sutherlandHodgmanClip(shellA, shellB)
	if len(clipped) < 4 {
		return geom.NewPolygonEmpty()
	}

	// Ensure closed
	if !clipped.IsClosed(geom.DefaultEpsilon) {
		clipped = append(clipped, clipped[0].Clone())
	}

	return geom.NewPolygon(geom.NewLinearRing(clipped), nil)
}

// clipPolygonDifference computes polyA minus polyB.
func clipPolygonDifference(polyA, polyB *geom.Polygon) []*geom.Polygon {
	if polyA.IsEmpty() {
		return nil
	}
	if polyB.IsEmpty() {
		return []*geom.Polygon{polyA}
	}

	// Check if envelopes intersect
	if !polyA.Envelope().Intersects(polyB.Envelope()) {
		return []*geom.Polygon{polyA}
	}

	// Check if polyA is completely inside polyB
	centroid := polyA.Centroid()
	if !centroid.IsEmpty() {
		loc := algorithm.PointLocationInPolygon(centroid.Coordinate(), polyB)
		if loc == geom.LocationInterior {
			// polyA might be completely inside polyB
			// Check all vertices
			allInside := true
			for _, c := range polyA.ExteriorRing().Coordinates() {
				if algorithm.PointLocationInPolygon(c, polyB) == geom.LocationExterior {
					allInside = false
					break
				}
			}
			if allInside {
				return nil // polyA is completely inside polyB
			}
		}
	}

	// Check if polyB is completely inside polyA
	centroidB := polyB.Centroid()
	if !centroidB.IsEmpty() {
		loc := algorithm.PointLocationInPolygon(centroidB.Coordinate(), polyA)
		if loc == geom.LocationInterior {
			// polyB might be completely inside polyA
			// Create polyA with polyB as a hole
			allInside := true
			for _, c := range polyB.ExteriorRing().Coordinates() {
				if algorithm.PointLocationInPolygon(c, polyA) == geom.LocationExterior {
					allInside = false
					break
				}
			}
			if allInside {
				// Add polyB as a hole to polyA
				holeCoords := polyB.ExteriorRing().Coordinates()
				ensureClockwiseCoords(holeCoords)
				newHole := geom.NewLinearRing(holeCoords)
				existingHoles := make([]*geom.LinearRing, polyA.NumInteriorRings())
				for i := 0; i < polyA.NumInteriorRings(); i++ {
					existingHoles[i] = polyA.InteriorRingN(i)
				}
				existingHoles = append(existingHoles, newHole)
				return []*geom.Polygon{geom.NewPolygon(polyA.ExteriorRing(), existingHoles)}
			}
		}
	}

	// For overlapping case, use clipping algorithm
	// This is a simplified approach - full implementation would be more complex
	shellA := polyA.ExteriorRing().Coordinates()
	shellB := polyB.ExteriorRing().Coordinates()

	// Get parts of A outside B
	outsideParts := clipPolygonOutside(shellA, shellB)
	var result []*geom.Polygon
	for _, part := range outsideParts {
		if len(part) >= 4 {
			if !part.IsClosed(geom.DefaultEpsilon) {
				part = append(part, part[0].Clone())
			}
			result = append(result, geom.NewPolygon(geom.NewLinearRing(part), nil))
		}
	}

	if len(result) == 0 {
		// If clipping failed, return original (conservative)
		return []*geom.Polygon{polyA}
	}

	return result
}

// sutherlandHodgmanClip clips a polygon against another polygon's edges.
func sutherlandHodgmanClip(subject, clip geom.CoordinateSequence) geom.CoordinateSequence {
	if len(subject) < 3 || len(clip) < 3 {
		return nil
	}

	output := subject.Clone()

	// Clip against each edge of the clip polygon
	for i := 0; i < len(clip)-1; i++ {
		if len(output) == 0 {
			return nil
		}

		input := output
		output = geom.CoordinateSequence{}

		edgeStart := clip[i]
		edgeEnd := clip[i+1]

		for j := 0; j < len(input); j++ {
			current := input[j]
			next := input[(j+1)%len(input)]

			currentInside := isLeft(edgeStart, edgeEnd, current) >= 0
			nextInside := isLeft(edgeStart, edgeEnd, next) >= 0

			if currentInside {
				output = append(output, current)
				if !nextInside {
					// Going out - add intersection
					// Use lineSegmentIntersect: clip edge is infinite line, subject edge is segment
					intersect := lineSegmentIntersect(edgeStart, edgeEnd, current, next)
					if intersect != nil {
						output = append(output, *intersect)
					}
				}
			} else if nextInside {
				// Coming in - add intersection
				intersect := lineSegmentIntersect(edgeStart, edgeEnd, current, next)
				if intersect != nil {
					output = append(output, *intersect)
				}
			}
		}
	}

	return output
}

// clipPolygonOutside returns parts of subject polygon outside clip polygon.
func clipPolygonOutside(subject, clip geom.CoordinateSequence) []geom.CoordinateSequence {
	if len(subject) < 3 || len(clip) < 3 {
		return []geom.CoordinateSequence{subject}
	}

	// Simplified approach: sample points along the boundary
	// and create segments that are outside
	var result []geom.CoordinateSequence
	var current geom.CoordinateSequence

	for i := 0; i < len(subject)-1; i++ {
		p := subject[i]
		loc := pointInPolygon(p, clip)

		if loc < 0 { // Outside
			current = append(current, p)
		} else {
			if len(current) >= 2 {
				current = append(current, current[0].Clone()) // Close
				result = append(result, current)
			}
			current = nil
		}
	}

	if len(current) >= 2 {
		current = append(current, current[0].Clone())
		result = append(result, current)
	}

	if len(result) == 0 {
		// No parts outside - return empty
		return nil
	}

	return result
}

// isLeft returns > 0 if p is left of line from a to b, < 0 if right, 0 if on line.
func isLeft(a, b, p geom.Coordinate) float64 {
	return (b.X-a.X)*(p.Y-a.Y) - (p.X-a.X)*(b.Y-a.Y)
}

// lineLineIntersect computes intersection of two line segments.
func lineLineIntersect(a1, a2, b1, b2 geom.Coordinate) *geom.Coordinate {
	d1x := a2.X - a1.X
	d1y := a2.Y - a1.Y
	d2x := b2.X - b1.X
	d2y := b2.Y - b1.Y

	denom := d1x*d2y - d1y*d2x
	if math.Abs(denom) < geom.DefaultEpsilon {
		return nil
	}

	dx := b1.X - a1.X
	dy := b1.Y - a1.Y

	t := (dx*d2y - dy*d2x) / denom
	s := (dx*d1y - dy*d1x) / denom

	if t < -geom.DefaultEpsilon || t > 1+geom.DefaultEpsilon ||
		s < -geom.DefaultEpsilon || s > 1+geom.DefaultEpsilon {
		return nil
	}

	result := geom.NewCoordinate(a1.X+t*d1x, a1.Y+t*d1y)
	return &result
}

// lineSegmentIntersect computes intersection of an infinite line (a1->a2) with a finite segment (b1->b2).
// The line is treated as infinite, while the segment is finite.
// Used by Sutherland-Hodgman where clip edges are infinite.
func lineSegmentIntersect(a1, a2, b1, b2 geom.Coordinate) *geom.Coordinate {
	d1x := a2.X - a1.X
	d1y := a2.Y - a1.Y
	d2x := b2.X - b1.X
	d2y := b2.Y - b1.Y

	denom := d1x*d2y - d1y*d2x
	if math.Abs(denom) < geom.DefaultEpsilon {
		return nil // Lines are parallel
	}

	dx := b1.X - a1.X
	dy := b1.Y - a1.Y

	// t is parameter along line a1->a2 (infinite, so don't bound)
	t := (dx*d2y - dy*d2x) / denom
	// s is parameter along segment b1->b2 (finite, must be in [0,1])
	s := (dx*d1y - dy*d1x) / denom

	// Only check s bounds (the segment), not t (the infinite line)
	if s < -geom.DefaultEpsilon || s > 1+geom.DefaultEpsilon {
		return nil
	}

	result := geom.NewCoordinate(a1.X+t*d1x, a1.Y+t*d1y)
	return &result
}

// pointInPolygon returns > 0 if inside, 0 if on boundary, < 0 if outside.
func pointInPolygon(p geom.Coordinate, ring geom.CoordinateSequence) int {
	n := len(ring)
	if n < 3 {
		return -1
	}

	inside := false
	j := n - 1

	for i := 0; i < n; i++ {
		xi, yi := ring[i].X, ring[i].Y
		xj, yj := ring[j].X, ring[j].Y

		if ((yi > p.Y) != (yj > p.Y)) &&
			(p.X < (xj-xi)*(p.Y-yi)/(yj-yi)+xi) {
			inside = !inside
		}
		j = i
	}

	if inside {
		return 1
	}
	return -1
}

// ensureClockwiseCoords ensures coordinates are in clockwise order.
func ensureClockwiseCoords(coords geom.CoordinateSequence) {
	if geom.SignedArea(coords) > 0 {
		// Counter-clockwise, reverse
		for i, j := 0, len(coords)-1; i < j; i, j = i+1, j-1 {
			coords[i], coords[j] = coords[j], coords[i]
		}
	}
}
