package overlay

import (
	"sort"

	"github.com/robert-malhotra/go-topology-suite/algorithm"
	"github.com/robert-malhotra/go-topology-suite/geom"
)

// lineLineIntersection computes intersection points and segments of two line sets.
func lineLineIntersection(linesA, linesB []*geom.LineString) geom.Geometry {
	var resultPoints []*geom.Point
	var resultLines []*geom.LineString

	for _, lineA := range linesA {
		coordsA := lineA.Coordinates()
		for _, lineB := range linesB {
			coordsB := lineB.Coordinates()

			// Check each segment pair
			for i := 1; i < len(coordsA); i++ {
				for j := 1; j < len(coordsB); j++ {
					result := algorithm.LineIntersection(
						coordsA[i-1], coordsA[i],
						coordsB[j-1], coordsB[j],
					)
					if result.HasIntersection {
						if result.IsCollinear && !result.Intersection2.IsNaN() {
							// Collinear overlap: emit a LineString
							resultLines = append(resultLines,
								geom.NewLineString(geom.CoordinateSequence{
									result.Intersection, result.Intersection2,
								}))
						} else {
							resultPoints = append(resultPoints,
								geom.NewPointFromCoordinate(result.Intersection))
						}
					}
				}
			}
		}
	}

	resultLines = mergeAdjacentLines(resultLines)
	return createMixedResult(resultPoints, resultLines)
}

// lineLineUnion computes the union of two line sets.
func lineLineUnion(linesA, linesB []*geom.LineString) geom.Geometry {
	// Simple union - just combine all lines
	var allLines []*geom.LineString
	allLines = append(allLines, linesA...)
	allLines = append(allLines, linesB...)

	if len(allLines) == 0 {
		return geom.NewLineStringEmpty()
	}
	if len(allLines) == 1 {
		return allLines[0]
	}
	return geom.NewMultiLineString(allLines)
}

// lineLineDifference computes parts of linesA not in linesB.
func lineLineDifference(linesA, linesB []*geom.LineString) geom.Geometry {
	if len(linesA) == 0 {
		return geom.NewLineStringEmpty()
	}

	var resultLines []*geom.LineString

	for _, lineA := range linesA {
		coordsA := lineA.Coordinates()

		// Process each segment of lineA
		for i := 0; i < len(coordsA)-1; i++ {
			p0 := coordsA[i]
			p1 := coordsA[i+1]

			// Collect covered intervals [tMin, tMax] on this segment
			var covered [][2]float64

			for _, lineB := range linesB {
				coordsB := lineB.Coordinates()
				for j := 0; j < len(coordsB)-1; j++ {
					result := algorithm.LineIntersection(p0, p1, coordsB[j], coordsB[j+1])
					if !result.HasIntersection {
						continue
					}
					if result.IsCollinear && !result.Intersection2.IsNaN() {
						// Compute t-parameters for the overlap on segment (p0, p1)
						t1 := tParam(p0, p1, result.Intersection)
						t2 := tParam(p0, p1, result.Intersection2)
						tMin := t1
						tMax := t2
						if tMin > tMax {
							tMin, tMax = tMax, tMin
						}
						// Clamp to [0, 1]
						if tMin < 0 {
							tMin = 0
						}
						if tMax > 1 {
							tMax = 1
						}
						if tMax > tMin+geom.DefaultEpsilon {
							covered = append(covered, [2]float64{tMin, tMax})
						}
					}
					// Single-point intersections don't remove length, skip them
				}
			}

			// Sort and merge covered intervals
			covered = mergeIntervals(covered)

			// Compute complement intervals on [0, 1]
			complement := complementIntervals(covered)

			// Create line segments from complement intervals
			for _, interval := range complement {
				startCoord := interpolate(p0, p1, interval[0])
				endCoord := interpolate(p0, p1, interval[1])
				if startCoord.Distance(endCoord) > geom.DefaultEpsilon {
					resultLines = append(resultLines, geom.NewLineString(
						geom.CoordinateSequence{startCoord, endCoord}))
				}
			}
		}
	}

	resultLines = mergeAdjacentLines(resultLines)
	return createLineResult(resultLines)
}

// lineLineSymDifference computes parts in either line set but not both.
func lineLineSymDifference(linesA, linesB []*geom.LineString) geom.Geometry {
	diffAB := lineLineDifference(linesA, linesB)
	diffBA := lineLineDifference(linesB, linesA)
	return collectGeometries(diffAB, diffBA)
}

// tParam computes the parameter t of point p along segment (a, b).
// Returns 0.0 if p == a, 1.0 if p == b.
func tParam(a, b, p geom.Coordinate) float64 {
	dx := b.X - a.X
	dy := b.Y - a.Y
	if abs(dx) > abs(dy) {
		return (p.X - a.X) / dx
	}
	if abs(dy) > geom.DefaultEpsilon {
		return (p.Y - a.Y) / dy
	}
	return 0
}

// interpolate returns the point at parameter t along segment (a, b).
func interpolate(a, b geom.Coordinate, t float64) geom.Coordinate {
	return geom.NewCoordinate(
		a.X+t*(b.X-a.X),
		a.Y+t*(b.Y-a.Y),
	)
}

// mergeIntervals sorts and merges overlapping intervals.
func mergeIntervals(intervals [][2]float64) [][2]float64 {
	if len(intervals) == 0 {
		return nil
	}
	sort.Slice(intervals, func(i, j int) bool {
		return intervals[i][0] < intervals[j][0]
	})
	merged := [][2]float64{intervals[0]}
	for _, iv := range intervals[1:] {
		last := &merged[len(merged)-1]
		if iv[0] <= last[1]+geom.DefaultEpsilon {
			if iv[1] > last[1] {
				last[1] = iv[1]
			}
		} else {
			merged = append(merged, iv)
		}
	}
	return merged
}

// complementIntervals returns the complement of the given intervals within [0, 1].
func complementIntervals(intervals [][2]float64) [][2]float64 {
	if len(intervals) == 0 {
		return [][2]float64{{0, 1}}
	}
	var result [][2]float64
	pos := 0.0
	for _, iv := range intervals {
		if iv[0] > pos+geom.DefaultEpsilon {
			result = append(result, [2]float64{pos, iv[0]})
		}
		pos = iv[1]
	}
	if pos < 1-geom.DefaultEpsilon {
		result = append(result, [2]float64{pos, 1})
	}
	return result
}

// createLineResult creates a geometry from a list of LineStrings.
func createLineResult(lines []*geom.LineString) geom.Geometry {
	if len(lines) == 0 {
		return geom.NewLineStringEmpty()
	}
	if len(lines) == 1 {
		return lines[0]
	}
	return geom.NewMultiLineString(lines)
}

// createMixedResult creates a geometry from points and lines.
func createMixedResult(points []*geom.Point, lines []*geom.LineString) geom.Geometry {
	hasPoints := len(points) > 0
	hasLines := len(lines) > 0

	if !hasPoints && !hasLines {
		return geom.NewPointEmpty()
	}
	if hasLines && !hasPoints {
		return createLineResult(lines)
	}
	if hasPoints && !hasLines {
		return createPointResult(points)
	}
	// Both points and lines: create GeometryCollection
	var geoms []geom.Geometry
	for _, l := range lines {
		geoms = append(geoms, l)
	}
	for _, p := range points {
		geoms = append(geoms, p)
	}
	return geom.NewGeometryCollection(geoms)
}

// linePolygonIntersection computes parts of lines inside polygons.
func linePolygonIntersection(lines []*geom.LineString, polygons []*geom.Polygon) geom.Geometry {
	var resultLines []*geom.LineString

	for _, line := range lines {
		for _, poly := range polygons {
			clipped := clipLineToPolygon(line, poly)
			resultLines = append(resultLines, clipped...)
		}
	}

	if len(resultLines) == 0 {
		return geom.NewLineStringEmpty()
	}
	if len(resultLines) == 1 {
		return resultLines[0]
	}
	return geom.NewMultiLineString(resultLines)
}

// linePolygonDifference computes parts of lines outside polygons.
func linePolygonDifference(lines []*geom.LineString, polygons []*geom.Polygon) geom.Geometry {
	var resultLines []*geom.LineString

	for _, line := range lines {
		remaining := []*geom.LineString{line}

		for _, poly := range polygons {
			var newRemaining []*geom.LineString
			for _, rem := range remaining {
				clipped := clipLineOutsidePolygon(rem, poly)
				newRemaining = append(newRemaining, clipped...)
			}
			remaining = newRemaining
		}

		resultLines = append(resultLines, remaining...)
	}

	if len(resultLines) == 0 {
		return geom.NewLineStringEmpty()
	}
	if len(resultLines) == 1 {
		return resultLines[0]
	}
	return geom.NewMultiLineString(resultLines)
}

// clipLineToPolygon clips a line to the interior of a polygon.
func clipLineToPolygon(line *geom.LineString, poly *geom.Polygon) []*geom.LineString {
	if line.IsEmpty() || poly.IsEmpty() {
		return nil
	}

	coords := line.Coordinates()
	if len(coords) < 2 {
		return nil
	}

	var result []*geom.LineString
	shell := poly.ExteriorRing().Coordinates()

	// Process each segment of the line
	for i := 0; i < len(coords)-1; i++ {
		segStart := coords[i]
		segEnd := coords[i+1]

		// Clip this segment to the polygon
		clippedSegments := clipSegmentToPolygon(segStart, segEnd, shell, poly)
		for _, seg := range clippedSegments {
			if len(seg) >= 2 {
				result = append(result, geom.NewLineString(seg))
			}
		}
	}

	// Merge adjacent segments that share endpoints
	result = mergeAdjacentLines(result)

	return result
}

// clipSegmentToPolygon clips a single line segment to a polygon.
func clipSegmentToPolygon(p0, p1 geom.Coordinate, shell geom.CoordinateSequence, poly *geom.Polygon) []geom.CoordinateSequence {
	// Find all intersection points with polygon boundary
	type intersection struct {
		point geom.Coordinate
		t     float64 // parameter along segment
	}

	var intersections []intersection

	// Add endpoints if inside
	loc0 := algorithm.PointLocationInPolygon(p0, poly)
	loc1 := algorithm.PointLocationInPolygon(p1, poly)

	// Check intersections with each edge of the polygon shell
	for i := 0; i < len(shell)-1; i++ {
		result := algorithm.LineIntersection(p0, p1, shell[i], shell[i+1])
		if result.HasIntersection {
			// Calculate t parameter
			dx := p1.X - p0.X
			dy := p1.Y - p0.Y
			var t float64
			if abs(dx) > abs(dy) {
				t = (result.Intersection.X - p0.X) / dx
			} else if abs(dy) > geom.DefaultEpsilon {
				t = (result.Intersection.Y - p0.Y) / dy
			} else {
				t = 0
			}

			if t > geom.DefaultEpsilon && t < 1-geom.DefaultEpsilon {
				intersections = append(intersections, intersection{result.Intersection, t})
			}
		}
	}

	// Sort intersections by t parameter
	sort.Slice(intersections, func(i, j int) bool {
		return intersections[i].t < intersections[j].t
	})

	// Remove duplicate intersections
	if len(intersections) > 1 {
		unique := []intersection{intersections[0]}
		for i := 1; i < len(intersections); i++ {
			if intersections[i].point.Distance(unique[len(unique)-1].point) > geom.DefaultEpsilon {
				unique = append(unique, intersections[i])
			}
		}
		intersections = unique
	}

	// Build list of points along segment
	type pointInfo struct {
		coord geom.Coordinate
		t     float64
	}

	points := []pointInfo{{p0, 0}}
	for _, inter := range intersections {
		points = append(points, pointInfo{inter.point, inter.t})
	}
	points = append(points, pointInfo{p1, 1})

	// Extract segments that are inside the polygon
	var result []geom.CoordinateSequence

	for i := 0; i < len(points)-1; i++ {
		// Check if midpoint is inside
		midX := (points[i].coord.X + points[i+1].coord.X) / 2
		midY := (points[i].coord.Y + points[i+1].coord.Y) / 2
		midpoint := geom.NewCoordinate(midX, midY)

		midLoc := algorithm.PointLocationInPolygon(midpoint, poly)
		if midLoc == geom.LocationInterior || midLoc == geom.LocationBoundary {
			result = append(result, geom.CoordinateSequence{points[i].coord, points[i+1].coord})
		}
	}

	// Handle case where entire segment is inside (no intersections)
	if len(intersections) == 0 {
		if (loc0 == geom.LocationInterior || loc0 == geom.LocationBoundary) &&
			(loc1 == geom.LocationInterior || loc1 == geom.LocationBoundary) {
			return []geom.CoordinateSequence{{p0, p1}}
		}
		return nil
	}

	return result
}

// mergeAdjacentLines merges lines that share endpoints.
func mergeAdjacentLines(lines []*geom.LineString) []*geom.LineString {
	if len(lines) <= 1 {
		return lines
	}

	var result []*geom.LineString
	var current geom.CoordinateSequence

	for _, line := range lines {
		coords := line.Coordinates()
		if len(coords) < 2 {
			continue
		}

		if len(current) == 0 {
			current = coords.Clone()
		} else {
			// Check if this line continues from the current one
			lastPt := current[len(current)-1]
			firstPt := coords[0]

			if lastPt.Distance(firstPt) < geom.DefaultEpsilon {
				// Merge: append all but first point
				current = append(current, coords[1:]...)
			} else {
				// Not adjacent, save current and start new
				if len(current) >= 2 {
					result = append(result, geom.NewLineString(current))
				}
				current = coords.Clone()
			}
		}
	}

	if len(current) >= 2 {
		result = append(result, geom.NewLineString(current))
	}

	return result
}

// abs returns absolute value.
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// clipLineOutsidePolygon clips a line to the exterior of a polygon.
func clipLineOutsidePolygon(line *geom.LineString, poly *geom.Polygon) []*geom.LineString {
	if line.IsEmpty() || poly.IsEmpty() {
		if !line.IsEmpty() {
			return []*geom.LineString{line}
		}
		return nil
	}

	coords := line.Coordinates()
	if len(coords) < 2 {
		return nil
	}

	var result []*geom.LineString
	shell := poly.ExteriorRing().Coordinates()

	// Process each segment of the line
	for i := 0; i < len(coords)-1; i++ {
		segStart := coords[i]
		segEnd := coords[i+1]

		// Clip this segment to outside the polygon
		clippedSegments := clipSegmentOutsidePolygon(segStart, segEnd, shell, poly)
		for _, seg := range clippedSegments {
			if len(seg) >= 2 {
				result = append(result, geom.NewLineString(seg))
			}
		}
	}

	// Merge adjacent segments that share endpoints
	result = mergeAdjacentLines(result)

	return result
}

// clipSegmentOutsidePolygon clips a single line segment to outside a polygon.
func clipSegmentOutsidePolygon(p0, p1 geom.Coordinate, shell geom.CoordinateSequence, poly *geom.Polygon) []geom.CoordinateSequence {
	// Find all intersection points with polygon boundary
	type intersection struct {
		point geom.Coordinate
		t     float64 // parameter along segment
	}

	var intersections []intersection

	// Check intersections with each edge of the polygon shell
	for i := 0; i < len(shell)-1; i++ {
		result := algorithm.LineIntersection(p0, p1, shell[i], shell[i+1])
		if result.HasIntersection {
			// Calculate t parameter
			dx := p1.X - p0.X
			dy := p1.Y - p0.Y
			var t float64
			if abs(dx) > abs(dy) {
				t = (result.Intersection.X - p0.X) / dx
			} else if abs(dy) > geom.DefaultEpsilon {
				t = (result.Intersection.Y - p0.Y) / dy
			} else {
				t = 0
			}

			if t > geom.DefaultEpsilon && t < 1-geom.DefaultEpsilon {
				intersections = append(intersections, intersection{result.Intersection, t})
			}
		}
	}

	// Sort intersections by t parameter
	sort.Slice(intersections, func(i, j int) bool {
		return intersections[i].t < intersections[j].t
	})

	// Remove duplicate intersections
	if len(intersections) > 1 {
		unique := []intersection{intersections[0]}
		for i := 1; i < len(intersections); i++ {
			if intersections[i].point.Distance(unique[len(unique)-1].point) > geom.DefaultEpsilon {
				unique = append(unique, intersections[i])
			}
		}
		intersections = unique
	}

	// Build list of points along segment
	type pointInfo struct {
		coord geom.Coordinate
		t     float64
	}

	points := []pointInfo{{p0, 0}}
	for _, inter := range intersections {
		points = append(points, pointInfo{inter.point, inter.t})
	}
	points = append(points, pointInfo{p1, 1})

	// Extract segments that are outside the polygon
	var result []geom.CoordinateSequence

	for i := 0; i < len(points)-1; i++ {
		// Check if midpoint is outside
		midX := (points[i].coord.X + points[i+1].coord.X) / 2
		midY := (points[i].coord.Y + points[i+1].coord.Y) / 2
		midpoint := geom.NewCoordinate(midX, midY)

		midLoc := algorithm.PointLocationInPolygon(midpoint, poly)
		if midLoc == geom.LocationExterior {
			result = append(result, geom.CoordinateSequence{points[i].coord, points[i+1].coord})
		}
	}

	// Handle case where entire segment is outside (no intersections)
	if len(intersections) == 0 {
		loc0 := algorithm.PointLocationInPolygon(p0, poly)
		loc1 := algorithm.PointLocationInPolygon(p1, poly)
		if loc0 == geom.LocationExterior && loc1 == geom.LocationExterior {
			return []geom.CoordinateSequence{{p0, p1}}
		}
		return nil
	}

	return result
}
