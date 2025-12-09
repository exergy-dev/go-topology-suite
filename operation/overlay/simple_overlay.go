package overlay

import (
	"math"

	"github.com/go-topology-suite/gts/algorithm"
	"github.com/go-topology-suite/gts/geom"
)

// NOTE: This file contains a simplified overlay implementation that was used
// as a fallback. The noded overlay implementation (noded_overlay.go) is now
// the primary implementation. This file is kept for reference and potential
// future use for specific edge cases, but is not actively used in production.

// simplePolygonOverlay implements polygon overlay using a simpler approach
// that doesn't rely on the complex noding framework.
// DEPRECATED: Use nodedPolygonOverlay instead.
func simplePolygonOverlay(polysA, polysB []*geom.Polygon, op Op) geom.Geometry {
	if len(polysA) == 0 && len(polysB) == 0 {
		return geom.NewPolygonEmpty()
	}
	if len(polysA) == 0 {
		return handleEmptyPolyA(polysB, op)
	}
	if len(polysB) == 0 {
		return handleEmptyPolyB(polysA, op)
	}

	// For now, handle single polygon case
	if len(polysA) == 1 && len(polysB) == 1 {
		return simpleSinglePolygonOverlay(polysA[0], polysB[0], op)
	}

	// For multiple polygons, process each pair and combine results
	// This is a simple approach that works for basic cases
	var results []*geom.Polygon
	for _, polyA := range polysA {
		for _, polyB := range polysB {
			result := simpleSinglePolygonOverlay(polyA, polyB, op)
			if !result.IsEmpty() {
				if poly, ok := result.(*geom.Polygon); ok {
					results = append(results, poly)
				} else if mp, ok := result.(*geom.MultiPolygon); ok {
					for i := 0; i < mp.NumGeometries(); i++ {
						results = append(results, mp.GeometryN(i).(*geom.Polygon))
					}
				}
			}
		}
	}
	return collectPolygons(results)
}

// simpleSinglePolygonOverlay handles overlay of two single polygons.
func simpleSinglePolygonOverlay(polyA, polyB *geom.Polygon, op Op) geom.Geometry {
	// Quick envelope check
	if !polyA.Envelope().Intersects(polyB.Envelope()) {
		// Disjoint case
		switch op {
		case OpIntersection:
			return geom.NewPolygonEmpty()
		case OpUnion, OpSymDifference:
			return geom.NewMultiPolygon([]*geom.Polygon{polyA, polyB})
		case OpDifference:
			return polyA
		}
	}

	// Check containment
	centroidA := polyA.Centroid()
	centroidB := polyB.Centroid()

	aContainsB := false
	bContainsA := false

	if !centroidB.IsEmpty() {
		locBinA := algorithm.PointLocationInPolygon(centroidB.Coordinate(), polyA)
		if locBinA == geom.LocationInterior {
			// B might be inside A - check all vertices
			allInside := true
			for _, c := range polyB.ExteriorRing().Coordinates() {
				if algorithm.PointLocationInPolygon(c, polyA) == geom.LocationExterior {
					allInside = false
					break
				}
			}
			aContainsB = allInside
		}
	}

	if !centroidA.IsEmpty() {
		locAinB := algorithm.PointLocationInPolygon(centroidA.Coordinate(), polyB)
		if locAinB == geom.LocationInterior {
			// A might be inside B - check all vertices
			allInside := true
			for _, c := range polyA.ExteriorRing().Coordinates() {
				if algorithm.PointLocationInPolygon(c, polyB) == geom.LocationExterior {
					allInside = false
					break
				}
			}
			bContainsA = allInside
		}
	}

	// Handle containment cases
	if aContainsB {
		switch op {
		case OpIntersection:
			return polyB
		case OpUnion:
			return polyA
		case OpDifference:
			// A - B where B is inside A: create hole
			shellA := polyA.ExteriorRing()
			holesA := make([]*geom.LinearRing, polyA.NumInteriorRings())
			for i := 0; i < polyA.NumInteriorRings(); i++ {
				holesA[i] = polyA.InteriorRingN(i)
			}
			// Add B's exterior as a hole (ensure clockwise for hole)
			holeCoords := polyB.ExteriorRing().Coordinates()
			if geom.SignedArea(holeCoords) > 0 {
				// Counter-clockwise, reverse it
				holeCoords = holeCoords.Reverse()
			}
			holesA = append(holesA, geom.NewLinearRing(holeCoords))
			return geom.NewPolygon(shellA, holesA)
		case OpSymDifference:
			// A △ B where B inside A: same as A - B
			shellA := polyA.ExteriorRing()
			holesA := make([]*geom.LinearRing, polyA.NumInteriorRings())
			for i := 0; i < polyA.NumInteriorRings(); i++ {
				holesA[i] = polyA.InteriorRingN(i)
			}
			holeCoords := polyB.ExteriorRing().Coordinates()
			if geom.SignedArea(holeCoords) > 0 {
				holeCoords = holeCoords.Reverse()
			}
			holesA = append(holesA, geom.NewLinearRing(holeCoords))
			return geom.NewPolygon(shellA, holesA)
		}
	}

	if bContainsA {
		switch op {
		case OpIntersection:
			return polyA
		case OpUnion:
			return polyB
		case OpDifference:
			// A - B where A is inside B: empty
			return geom.NewPolygonEmpty()
		case OpSymDifference:
			// B - A
			shellB := polyB.ExteriorRing()
			holesB := make([]*geom.LinearRing, polyB.NumInteriorRings())
			for i := 0; i < polyB.NumInteriorRings(); i++ {
				holesB[i] = polyB.InteriorRingN(i)
			}
			holeCoords := polyA.ExteriorRing().Coordinates()
			if geom.SignedArea(holeCoords) > 0 {
				holeCoords = holeCoords.Reverse()
			}
			holesB = append(holesB, geom.NewLinearRing(holeCoords))
			return geom.NewPolygon(shellB, holesB)
		}
	}

	// General case: handle overlapping polygons using clipping
	switch op {
	case OpIntersection:
		// Use Sutherland-Hodgman clipping
		return clipPolygonToPolygon(polyA, polyB)

	case OpDifference:
		// A - B: parts of A not in B
		// Use the clipping approach
		result := clipPolygonDifference(polyA, polyB)
		return collectPolygons(result)

	case OpUnion:
		// Try merging the polygons
		merged := mergePolygons(polyA, polyB)
		if merged != nil && !merged.IsEmpty() {
			return merged
		}
		// Fall back to returning both polygons as a multi-polygon
		return geom.NewMultiPolygon([]*geom.Polygon{polyA, polyB})

	case OpSymDifference:
		// Symmetric difference: (A - B) union (B - A)
		aMinusB := clipPolygonDifference(polyA, polyB)
		bMinusA := clipPolygonDifference(polyB, polyA)

		// Combine results
		allPolys := append(aMinusB, bMinusA...)
		if len(allPolys) == 0 {
			return geom.NewPolygonEmpty()
		}
		if len(allPolys) == 1 {
			return allPolys[0]
		}
		return geom.NewMultiPolygon(allPolys)

	default:
		return geom.NewPolygonEmpty()
	}
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
