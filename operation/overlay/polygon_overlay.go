package overlay

import (
	"math"

	"github.com/robert-malhotra/go-topology-suite/geom"
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

