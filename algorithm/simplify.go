package algorithm

import (
	"math"

	"github.com/go-topology-suite/gts/geom"
)

// DouglasPeucker simplifies a geometry using the Douglas-Peucker algorithm.
// The tolerance specifies the maximum perpendicular distance a point can be
// from the simplified line before it must be kept.
func DouglasPeucker(g geom.Geometry, tolerance float64) geom.Geometry {
	switch v := g.(type) {
	case *geom.Point:
		return v.Clone()
	case *geom.LineString:
		return simplifyLineString(v, tolerance)
	case *geom.LinearRing:
		return simplifyLinearRing(v, tolerance)
	case *geom.Polygon:
		return simplifyPolygon(v, tolerance)
	case *geom.MultiPoint:
		return v.Clone()
	case *geom.MultiLineString:
		return simplifyMultiLineString(v, tolerance)
	case *geom.MultiPolygon:
		return simplifyMultiPolygon(v, tolerance)
	case *geom.GeometryCollection:
		return simplifyGeometryCollection(v, tolerance)
	default:
		return g.Clone()
	}
}

func simplifyLineString(ls *geom.LineString, tolerance float64) *geom.LineString {
	if ls.NumPoints() <= 2 {
		return ls.Clone().(*geom.LineString)
	}

	coords := ls.Coordinates()
	simplified := douglasPeuckerSimplify(coords, tolerance)

	// Ensure at least 2 points
	if len(simplified) < 2 {
		simplified = geom.CoordinateSequence{coords[0], coords[len(coords)-1]}
	}

	return geom.NewLineString(simplified)
}

func simplifyLinearRing(lr *geom.LinearRing, tolerance float64) *geom.LinearRing {
	if lr.NumPoints() <= 4 {
		return lr.Clone().(*geom.LinearRing)
	}

	coords := lr.Coordinates()
	// Remove closing point for simplification
	open := coords[:len(coords)-1]
	simplified := douglasPeuckerSimplify(open, tolerance)

	// Ensure at least 3 points (4 with closure)
	if len(simplified) < 3 {
		// Keep first, middle, and last points
		mid := len(open) / 2
		simplified = geom.CoordinateSequence{open[0], open[mid], open[len(open)-1]}
	}

	// Re-close the ring
	simplified = append(simplified, simplified[0])

	return geom.NewLinearRing(simplified)
}

func simplifyPolygon(p *geom.Polygon, tolerance float64) *geom.Polygon {
	if p.IsEmpty() {
		return p.Clone().(*geom.Polygon)
	}

	shell := simplifyLinearRing(p.ExteriorRing(), tolerance)

	holes := make([]*geom.LinearRing, 0, p.NumInteriorRings())
	for i := 0; i < p.NumInteriorRings(); i++ {
		hole := simplifyLinearRing(p.InteriorRingN(i), tolerance)
		// Only keep holes with at least 4 points
		if hole.NumPoints() >= 4 {
			holes = append(holes, hole)
		}
	}

	return geom.NewPolygon(shell, holes)
}

func simplifyMultiLineString(mls *geom.MultiLineString, tolerance float64) *geom.MultiLineString {
	lines := make([]*geom.LineString, 0, mls.NumGeometries())
	for i := 0; i < mls.NumGeometries(); i++ {
		ls := mls.GeometryN(i).(*geom.LineString)
		simplified := simplifyLineString(ls, tolerance)
		if simplified.NumPoints() >= 2 {
			lines = append(lines, simplified)
		}
	}
	return geom.NewMultiLineString(lines)
}

func simplifyMultiPolygon(mp *geom.MultiPolygon, tolerance float64) *geom.MultiPolygon {
	polys := make([]*geom.Polygon, 0, mp.NumGeometries())
	for i := 0; i < mp.NumGeometries(); i++ {
		p := mp.GeometryN(i).(*geom.Polygon)
		simplified := simplifyPolygon(p, tolerance)
		if !simplified.IsEmpty() {
			polys = append(polys, simplified)
		}
	}
	return geom.NewMultiPolygon(polys)
}

func simplifyGeometryCollection(gc *geom.GeometryCollection, tolerance float64) *geom.GeometryCollection {
	geoms := make([]geom.Geometry, 0, gc.NumGeometries())
	for i := 0; i < gc.NumGeometries(); i++ {
		simplified := DouglasPeucker(gc.GeometryN(i), tolerance)
		if !simplified.IsEmpty() {
			geoms = append(geoms, simplified)
		}
	}
	return geom.NewGeometryCollection(geoms)
}

// douglasPeuckerSimplify implements the Douglas-Peucker algorithm.
func douglasPeuckerSimplify(coords geom.CoordinateSequence, tolerance float64) geom.CoordinateSequence {
	if len(coords) <= 2 {
		return coords.Clone()
	}

	// Find the point with maximum distance
	maxDist := 0.0
	maxIdx := 0

	start := coords[0]
	end := coords[len(coords)-1]

	for i := 1; i < len(coords)-1; i++ {
		dist := DistancePointToSegment(coords[i], start, end)
		if dist > maxDist {
			maxDist = dist
			maxIdx = i
		}
	}

	// If max distance is greater than tolerance, recursively simplify
	if maxDist > tolerance {
		// Recursive call
		left := douglasPeuckerSimplify(coords[:maxIdx+1], tolerance)
		right := douglasPeuckerSimplify(coords[maxIdx:], tolerance)

		// Concatenate (removing duplicate point)
		result := make(geom.CoordinateSequence, 0, len(left)+len(right)-1)
		result = append(result, left[:len(left)-1]...)
		result = append(result, right...)
		return result
	}

	// Return just the endpoints
	return geom.CoordinateSequence{start, end}
}

// VisvalingamWhyatt simplifies a geometry using the Visvalingam-Whyatt algorithm.
// This algorithm removes points based on the area of the triangle they form.
func VisvalingamWhyatt(g geom.Geometry, areaThreshold float64) geom.Geometry {
	switch v := g.(type) {
	case *geom.LineString:
		return visvalingamLineString(v, areaThreshold)
	case *geom.Polygon:
		return visvalingamPolygon(v, areaThreshold)
	default:
		// Fall back to Douglas-Peucker for other types
		return DouglasPeucker(g, math.Sqrt(areaThreshold))
	}
}

func visvalingamLineString(ls *geom.LineString, threshold float64) *geom.LineString {
	if ls.NumPoints() <= 2 {
		return ls.Clone().(*geom.LineString)
	}

	coords := ls.Coordinates()
	simplified := visvalingamSimplify(coords, threshold)

	if len(simplified) < 2 {
		simplified = geom.CoordinateSequence{coords[0], coords[len(coords)-1]}
	}

	return geom.NewLineString(simplified)
}

func visvalingamPolygon(p *geom.Polygon, threshold float64) *geom.Polygon {
	if p.IsEmpty() {
		return p.Clone().(*geom.Polygon)
	}

	shellCoords := p.ExteriorRing().Coordinates()
	// Remove closing point
	open := shellCoords[:len(shellCoords)-1]
	simplified := visvalingamSimplify(open, threshold)

	if len(simplified) < 3 {
		mid := len(open) / 2
		simplified = geom.CoordinateSequence{open[0], open[mid], open[len(open)-1]}
	}

	simplified = append(simplified, simplified[0])
	shell := geom.NewLinearRing(simplified)

	holes := make([]*geom.LinearRing, 0, p.NumInteriorRings())
	for i := 0; i < p.NumInteriorRings(); i++ {
		holeCoords := p.InteriorRingN(i).Coordinates()
		openHole := holeCoords[:len(holeCoords)-1]
		simplifiedHole := visvalingamSimplify(openHole, threshold)
		if len(simplifiedHole) >= 3 {
			simplifiedHole = append(simplifiedHole, simplifiedHole[0])
			holes = append(holes, geom.NewLinearRing(simplifiedHole))
		}
	}

	return geom.NewPolygon(shell, holes)
}

// visvalingamSimplify implements the Visvalingam-Whyatt algorithm.
func visvalingamSimplify(coords geom.CoordinateSequence, threshold float64) geom.CoordinateSequence {
	if len(coords) <= 2 {
		return coords.Clone()
	}

	// Create a list of indices that are still in the result
	indices := make([]int, len(coords))
	for i := range indices {
		indices[i] = i
	}

	// Compute initial areas
	areas := make([]float64, len(coords))
	for i := 1; i < len(coords)-1; i++ {
		areas[i] = triangleArea(coords[indices[i-1]], coords[indices[i]], coords[indices[i+1]])
	}
	areas[0] = math.Inf(1)
	areas[len(areas)-1] = math.Inf(1)

	// Iteratively remove points with smallest area
	for len(indices) > 2 {
		// Find minimum area
		minArea := math.Inf(1)
		minIdx := -1

		for i := 1; i < len(indices)-1; i++ {
			if areas[indices[i]] < minArea {
				minArea = areas[indices[i]]
				minIdx = i
			}
		}

		if minArea > threshold {
			break
		}

		// Remove the point
		removed := indices[minIdx]
		indices = append(indices[:minIdx], indices[minIdx+1:]...)

		// Update adjacent areas
		if minIdx > 1 && minIdx-1 < len(indices)-1 {
			areas[indices[minIdx-1]] = triangleArea(
				coords[indices[minIdx-2]],
				coords[indices[minIdx-1]],
				coords[indices[minIdx]],
			)
		}
		if minIdx < len(indices)-1 && minIdx > 0 {
			areas[indices[minIdx]] = triangleArea(
				coords[indices[minIdx-1]],
				coords[indices[minIdx]],
				coords[indices[minIdx+1]],
			)
		}
		// Mark removed point
		areas[removed] = math.Inf(1)
	}

	// Build result
	result := make(geom.CoordinateSequence, len(indices))
	for i, idx := range indices {
		result[i] = coords[idx]
	}

	return result
}

// triangleArea computes the area of a triangle formed by three points.
func triangleArea(p1, p2, p3 geom.Coordinate) float64 {
	return math.Abs((p2.X-p1.X)*(p3.Y-p1.Y)-(p3.X-p1.X)*(p2.Y-p1.Y)) / 2
}

// RadialDistance simplifies a geometry by removing points within a distance threshold.
func RadialDistance(g geom.Geometry, threshold float64) geom.Geometry {
	switch v := g.(type) {
	case *geom.LineString:
		return radialDistanceLineString(v, threshold)
	default:
		return DouglasPeucker(g, threshold)
	}
}

func radialDistanceLineString(ls *geom.LineString, threshold float64) *geom.LineString {
	if ls.NumPoints() <= 2 {
		return ls.Clone().(*geom.LineString)
	}

	coords := ls.Coordinates()
	result := geom.CoordinateSequence{coords[0]}

	for i := 1; i < len(coords)-1; i++ {
		lastKept := result[len(result)-1]
		if coords[i].Distance(lastKept) >= threshold {
			result = append(result, coords[i])
		}
	}

	// Always keep the last point
	result = append(result, coords[len(coords)-1])

	return geom.NewLineString(result)
}
