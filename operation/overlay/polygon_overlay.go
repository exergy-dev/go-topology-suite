package overlay

import (
	"math"

	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/robert-malhotra/go-topology-suite/internal/topology"
)

// polygonPolygonIntersection computes the intersection of two polygon sets.
func polygonPolygonIntersection(polysA, polysB []*geom.Polygon, pm geom.PrecisionModel) geom.Geometry {
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
	result := nodedPolygonOverlayWithPrecision(polysA, polysB, OpIntersection, pm)
	if result == nil || result.IsEmpty() {
		return polygonBoundaryIntersection(polysA, polysB)
	}
	return result
}

// polygonPolygonUnion computes the union of two polygon sets.
func polygonPolygonUnion(polysA, polysB []*geom.Polygon, pm geom.PrecisionModel) geom.Geometry {
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
	return nodedPolygonOverlayWithPrecision(polysA, polysB, OpUnion, pm)
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
func polygonPolygonDifference(polysA, polysB []*geom.Polygon, pm geom.PrecisionModel) geom.Geometry {
	// Bound coordinates to prevent overflow
	polysA = boundPolygons(polysA)
	polysB = boundPolygons(polysB)

	// Use noded overlay for robust difference handling
	return nodedPolygonOverlayWithPrecision(polysA, polysB, OpDifference, pm)
}

// polygonPolygonSymDifference computes parts in either polygon set but not both.
func polygonPolygonSymDifference(polysA, polysB []*geom.Polygon, pm geom.PrecisionModel) geom.Geometry {
	// Bound coordinates to prevent overflow
	polysA = boundPolygons(polysA)
	polysB = boundPolygons(polysB)

	// Use noded overlay for robust symmetric difference handling
	return nodedPolygonOverlayWithPrecision(polysA, polysB, OpSymDifference, pm)
}

func polygonBoundaryIntersection(polysA, polysB []*geom.Polygon) geom.Geometry {
	var points []*geom.Point
	var lines []*geom.LineString
	seenPoints := make(map[pointKey]struct{})
	seenLines := make(map[boundarySegmentKey]struct{})

	nodedSegments := topology.NodePolygonBoundaries(polysA, polysB)
	boundaryLinesA := topology.PolygonBoundaryLines(polysA)
	for _, segment := range nodedSegments {
		if !segment.InA() || !segment.InB() {
			continue
		}
		key := makeBoundarySegmentKey(segment.Start, segment.End)
		if _, ok := seenLines[key]; ok {
			continue
		}
		seenLines[key] = struct{}{}
		lines = append(lines, orientBoundarySegment(segment, boundaryLinesA))
	}

	boundaryLinesB := topology.PolygonBoundaryLines(polysB)
	for _, segment := range nodedSegments {
		for _, coord := range []geom.Coordinate{segment.Start, segment.End} {
			if pointOnAnyLine(coord, lines) {
				continue
			}
			if !pointOnAnyLine(coord, boundaryLinesA) || !pointOnAnyLine(coord, boundaryLinesB) {
				continue
			}
			key := makePointKey(coord)
			if _, ok := seenPoints[key]; ok {
				continue
			}
			seenPoints[key] = struct{}{}
			points = append(points, geom.NewPointFromCoordinate(coord))
		}
	}

	return createMixedResult(points, lines)
}

func orientBoundarySegment(segment topology.NodedLineSegment, sourceLines []*geom.LineString) *geom.LineString {
	start := segment.Start
	end := segment.End
	for _, line := range sourceLines {
		coords := line.Coordinates()
		for i := 0; i < len(coords)-1; i++ {
			a := coords[i]
			b := coords[i+1]
			if !geom.PointOnSegment(start, a, b) || !geom.PointOnSegment(end, a, b) {
				continue
			}
			if segmentOffset(end, a, b) < segmentOffset(start, a, b) {
				start, end = end, start
			}
			return geom.NewLineString(geom.CoordinateSequence{start, end})
		}
	}
	return segment.LineString()
}

func segmentOffset(point, start, end geom.Coordinate) float64 {
	dx := end.X - start.X
	dy := end.Y - start.Y
	if math.Abs(dx) >= math.Abs(dy) {
		if math.Abs(dx) <= geom.DefaultEpsilon {
			return 0
		}
		return (point.X - start.X) / dx
	}
	if math.Abs(dy) <= geom.DefaultEpsilon {
		return 0
	}
	return (point.Y - start.Y) / dy
}

type boundarySegmentKey struct {
	x1, y1, x2, y2 float64
}

func makeBoundarySegmentKey(a, b geom.Coordinate) boundarySegmentKey {
	if b.X < a.X || (b.X == a.X && b.Y < a.Y) {
		a, b = b, a
	}
	return boundarySegmentKey{a.X, a.Y, b.X, b.Y}
}
