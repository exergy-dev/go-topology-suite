package algorithm

import (
	"fmt"
	"sort"

	"github.com/go-topology-suite/gts/geom"
)

// ConvexHull computes the convex hull of a geometry.
// Returns a Polygon for 3+ points, LineString for 2 points, Point for 1 point.
func ConvexHull(g geom.Geometry) geom.Geometry {
	coords := g.Coordinates()
	return ConvexHullFromCoords(coords)
}

// ConvexHullFromCoords computes the convex hull of a set of coordinates.
func ConvexHullFromCoords(coords geom.CoordinateSequence) geom.Geometry {
	if len(coords) == 0 {
		return geom.NewPointEmpty()
	}
	if len(coords) == 1 {
		return geom.NewPointFromCoordinate(coords[0])
	}
	if len(coords) == 2 {
		return geom.NewLineString(coords)
	}

	// Remove duplicate points
	unique := uniqueCoords(coords)
	if len(unique) < 3 {
		if len(unique) == 1 {
			return geom.NewPointFromCoordinate(unique[0])
		}
		return geom.NewLineString(unique)
	}

	// Graham scan algorithm
	hull := grahamScan(unique)

	if len(hull) < 3 {
		if len(hull) == 2 {
			return geom.NewLineString(hull)
		}
		return geom.NewPointFromCoordinate(hull[0])
	}

	// Close the ring
	hull = append(hull, hull[0])
	ring := geom.NewLinearRing(hull)
	return geom.NewPolygon(ring, nil)
}

func uniqueCoords(coords geom.CoordinateSequence) geom.CoordinateSequence {
	seen := make(map[string]bool)
	result := make(geom.CoordinateSequence, 0, len(coords))

	for _, c := range coords {
		key := coordKey(c)
		if !seen[key] {
			seen[key] = true
			result = append(result, c)
		}
	}
	return result
}

func coordKey(c geom.Coordinate) string {
	return fmt.Sprintf("%.15g,%.15g", c.X, c.Y)
}

// grahamScan implements the Graham scan algorithm for convex hull.
func grahamScan(coords geom.CoordinateSequence) geom.CoordinateSequence {
	// Find the lowest point (and leftmost if tie)
	lowestIdx := 0
	for i := 1; i < len(coords); i++ {
		if coords[i].Y < coords[lowestIdx].Y ||
			(coords[i].Y == coords[lowestIdx].Y && coords[i].X < coords[lowestIdx].X) {
			lowestIdx = i
		}
	}

	// Swap lowest to first position
	coords[0], coords[lowestIdx] = coords[lowestIdx], coords[0]
	pivot := coords[0]

	// Sort remaining points by polar angle relative to pivot
	toSort := coords[1:]
	sort.Slice(toSort, func(i, j int) bool {
		o := OrientationIndex(pivot, toSort[i], toSort[j])
		if o == Collinear {
			// If collinear, closer point comes first
			return pivot.Distance(toSort[i]) < pivot.Distance(toSort[j])
		}
		return o == CounterClockwise
	})

	// Remove collinear points (keep farthest)
	filtered := geom.CoordinateSequence{pivot}
	for i := 0; i < len(toSort); i++ {
		// Skip points collinear with the pivot and previous point
		for i < len(toSort)-1 && OrientationIndex(pivot, toSort[i], toSort[i+1]) == Collinear {
			i++
		}
		filtered = append(filtered, toSort[i])
	}

	if len(filtered) < 3 {
		return filtered
	}

	// Build hull using stack
	stack := geom.CoordinateSequence{filtered[0], filtered[1]}

	for i := 2; i < len(filtered); i++ {
		// Pop while the turn is not counter-clockwise
		for len(stack) > 1 && OrientationIndex(stack[len(stack)-2], stack[len(stack)-1], filtered[i]) != CounterClockwise {
			stack = stack[:len(stack)-1]
		}
		stack = append(stack, filtered[i])
	}

	return stack
}

// MonotoneChain computes convex hull using the monotone chain algorithm.
// This is an alternative to Graham scan with similar O(n log n) complexity.
func MonotoneChain(coords geom.CoordinateSequence) geom.Geometry {
	if len(coords) == 0 {
		return geom.NewPointEmpty()
	}
	if len(coords) == 1 {
		return geom.NewPointFromCoordinate(coords[0])
	}

	// Sort points lexicographically
	sorted := make(geom.CoordinateSequence, len(coords))
	copy(sorted, coords)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].X != sorted[j].X {
			return sorted[i].X < sorted[j].X
		}
		return sorted[i].Y < sorted[j].Y
	})

	// Remove duplicates
	unique := make(geom.CoordinateSequence, 0, len(sorted))
	for i, c := range sorted {
		if i == 0 || !c.Equals2D(unique[len(unique)-1], geom.DefaultEpsilon) {
			unique = append(unique, c)
		}
	}

	if len(unique) < 3 {
		if len(unique) == 2 {
			return geom.NewLineString(unique)
		}
		return geom.NewPointFromCoordinate(unique[0])
	}

	// Build lower hull
	lower := make(geom.CoordinateSequence, 0, len(unique))
	for _, c := range unique {
		for len(lower) >= 2 && OrientationIndex(lower[len(lower)-2], lower[len(lower)-1], c) != CounterClockwise {
			lower = lower[:len(lower)-1]
		}
		lower = append(lower, c)
	}

	// Build upper hull
	upper := make(geom.CoordinateSequence, 0, len(unique))
	for i := len(unique) - 1; i >= 0; i-- {
		c := unique[i]
		for len(upper) >= 2 && OrientationIndex(upper[len(upper)-2], upper[len(upper)-1], c) != CounterClockwise {
			upper = upper[:len(upper)-1]
		}
		upper = append(upper, c)
	}

	// Concatenate (remove last point of each half to avoid duplication)
	lower = lower[:len(lower)-1]
	upper = upper[:len(upper)-1]
	hull := append(lower, upper...)

	if len(hull) < 3 {
		if len(hull) == 2 {
			return geom.NewLineString(hull)
		}
		return geom.NewPointFromCoordinate(hull[0])
	}

	// Close the ring
	hull = append(hull, hull[0])
	ring := geom.NewLinearRing(hull)
	return geom.NewPolygon(ring, nil)
}

// IsConvex returns true if the given polygon is convex.
func IsConvex(p *geom.Polygon) bool {
	if p.IsEmpty() || p.NumInteriorRings() > 0 {
		return false
	}

	coords := p.ExteriorRing().Coordinates()
	if len(coords) < 4 {
		return true
	}

	n := len(coords) - 1 // Exclude closing point
	sign := 0

	for i := 0; i < n; i++ {
		p1 := coords[i]
		p2 := coords[(i+1)%n]
		p3 := coords[(i+2)%n]

		o := OrientationIndex(p1, p2, p3)
		if o != Collinear {
			if sign == 0 {
				sign = o
			} else if sign != o {
				return false
			}
		}
	}

	return true
}
