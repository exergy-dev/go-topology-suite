// Package overlay provides geometry overlay operations.
//
// Overlay operations combine two geometries to produce a result based on
// set-theoretic operations: intersection, union, difference, and symmetric difference.
package overlay

import (
	"github.com/robert-malhotra/go-topology-suite/algorithm"
	"github.com/robert-malhotra/go-topology-suite/geom"
)

// Op represents the type of overlay operation.
type Op int

const (
	// OpIntersection computes the geometry that is in both inputs.
	OpIntersection Op = iota
	// OpUnion computes the geometry that is in either input.
	OpUnion
	// OpDifference computes the geometry that is in A but not B.
	OpDifference
	// OpSymDifference computes the geometry in either input but not both.
	OpSymDifference
)

// Intersection computes the geometry that is in both A and B.
func Intersection(a, b geom.Geometry) geom.Geometry {
	return Overlay(a, b, OpIntersection)
}

// Union computes the geometry that is in either A or B.
func Union(a, b geom.Geometry) geom.Geometry {
	return Overlay(a, b, OpUnion)
}

// Difference computes the geometry that is in A but not in B.
func Difference(a, b geom.Geometry) geom.Geometry {
	return Overlay(a, b, OpDifference)
}

// SymDifference computes the geometry that is in A or B but not both.
func SymDifference(a, b geom.Geometry) geom.Geometry {
	return Overlay(a, b, OpSymDifference)
}

// Overlay performs the specified overlay operation on two geometries.
func Overlay(a, b geom.Geometry, op Op) geom.Geometry {
	if a == nil || a.IsEmpty() {
		return handleEmptyA(b, op)
	}
	if b == nil || b.IsEmpty() {
		return handleEmptyB(a, op)
	}

	// Special case: if both geometries are the same reference
	// This handles self-intersection/union efficiently and avoids issues
	// with extreme coordinates becoming degenerate after bounding
	if a == b {
		switch op {
		case OpIntersection, OpUnion:
			// A ∩ A = A, A ∪ A = A
			return a.Clone()
		case OpDifference, OpSymDifference:
			// A - A = ∅, A △ A = ∅
			return geom.NewGeometryCollectionEmpty()
		}
	}

	// Quick envelope check
	if !a.Envelope().Intersects(b.Envelope()) {
		return handleDisjoint(a, b, op)
	}

	// Dispatch based on geometry types
	return computeOverlay(a, b, op)
}

// handleEmptyA handles the case where A is empty.
func handleEmptyA(b geom.Geometry, op Op) geom.Geometry {
	switch op {
	case OpIntersection:
		return geom.NewGeometryCollectionEmpty()
	case OpUnion:
		if b != nil && !b.IsEmpty() {
			return b.Clone()
		}
		return geom.NewGeometryCollectionEmpty()
	case OpDifference:
		return geom.NewGeometryCollectionEmpty()
	case OpSymDifference:
		if b != nil && !b.IsEmpty() {
			return b.Clone()
		}
		return geom.NewGeometryCollectionEmpty()
	default:
		return geom.NewGeometryCollectionEmpty()
	}
}

// handleEmptyB handles the case where B is empty.
func handleEmptyB(a geom.Geometry, op Op) geom.Geometry {
	switch op {
	case OpIntersection:
		return geom.NewGeometryCollectionEmpty()
	case OpUnion:
		return a.Clone()
	case OpDifference:
		return a.Clone()
	case OpSymDifference:
		return a.Clone()
	default:
		return geom.NewGeometryCollectionEmpty()
	}
}

// handleDisjoint handles the case where envelopes don't intersect.
func handleDisjoint(a, b geom.Geometry, op Op) geom.Geometry {
	switch op {
	case OpIntersection:
		return geom.NewGeometryCollectionEmpty()
	case OpUnion:
		return collectGeometries(a, b)
	case OpDifference:
		return a.Clone()
	case OpSymDifference:
		return collectGeometries(a, b)
	default:
		return geom.NewGeometryCollectionEmpty()
	}
}

// computeOverlay dispatches to the appropriate overlay computation.
func computeOverlay(a, b geom.Geometry, op Op) geom.Geometry {
	// Handle by dimension combination
	dimA := a.Dimension()
	dimB := b.Dimension()

	// Point cases
	if dimA == geom.DimensionPoint {
		return overlayPointWith(a, b, op)
	}
	if dimB == geom.DimensionPoint {
		return overlayWithPoint(a, b, op)
	}

	// Line cases
	if dimA == geom.DimensionLine && dimB == geom.DimensionLine {
		return overlayLineLine(a, b, op)
	}
	if dimA == geom.DimensionLine && dimB == geom.DimensionArea {
		return overlayLinePolygon(a, b, op)
	}
	if dimA == geom.DimensionArea && dimB == geom.DimensionLine {
		// Swap and adjust operation for some cases
		return overlayPolygonLine(a, b, op)
	}

	// Polygon/Polygon case
	if dimA == geom.DimensionArea && dimB == geom.DimensionArea {
		return overlayPolygonPolygon(a, b, op)
	}

	return geom.NewGeometryCollectionEmpty()
}

// overlayPointWith overlays a point geometry with another geometry.
func overlayPointWith(a, b geom.Geometry, op Op) geom.Geometry {
	points := geom.ExtractPoints(a)
	var resultPoints []*geom.Point

	for _, p := range points {
		coord := p.Coordinate()
		loc := algorithm.PointLocation(coord, b)

		switch op {
		case OpIntersection:
			if loc != geom.LocationExterior {
				resultPoints = append(resultPoints, p)
			}
		case OpUnion:
			resultPoints = append(resultPoints, p)
		case OpDifference:
			if loc == geom.LocationExterior {
				resultPoints = append(resultPoints, p)
			}
		case OpSymDifference:
			if loc == geom.LocationExterior {
				resultPoints = append(resultPoints, p)
			}
		}
	}

	// For union and symDifference, include B as well
	if op == OpUnion || op == OpSymDifference {
		return collectGeometries(createPointResult(resultPoints), b.Clone())
	}

	return createPointResult(resultPoints)
}

// overlayWithPoint overlays a geometry with a point geometry.
func overlayWithPoint(a, b geom.Geometry, op Op) geom.Geometry {
	points := geom.ExtractPoints(b)
	var resultPoints []*geom.Point

	for _, p := range points {
		coord := p.Coordinate()
		loc := algorithm.PointLocation(coord, a)

		switch op {
		case OpIntersection:
			if loc != geom.LocationExterior {
				resultPoints = append(resultPoints, p)
			}
		case OpUnion:
			// Include point only if exterior to A
			if loc == geom.LocationExterior {
				resultPoints = append(resultPoints, p)
			}
		case OpDifference:
			// A minus point - keep A, remove points that are in A
			// Points don't affect area/line geometries
		case OpSymDifference:
			if loc == geom.LocationExterior {
				resultPoints = append(resultPoints, p)
			}
		}
	}

	switch op {
	case OpIntersection:
		return createPointResult(resultPoints)
	case OpUnion:
		return collectGeometries(a.Clone(), createPointResult(resultPoints))
	case OpDifference:
		return a.Clone()
	case OpSymDifference:
		return collectGeometries(a.Clone(), createPointResult(resultPoints))
	default:
		return geom.NewGeometryCollectionEmpty()
	}
}

// overlayLineLine overlays two line geometries.
func overlayLineLine(a, b geom.Geometry, op Op) geom.Geometry {
	linesA := geom.ExtractLineStringsWithRings(a)
	linesB := geom.ExtractLineStringsWithRings(b)

	switch op {
	case OpIntersection:
		return lineLineIntersection(linesA, linesB)
	case OpUnion:
		return lineLineUnion(linesA, linesB)
	case OpDifference:
		return lineLineDifference(linesA, linesB)
	case OpSymDifference:
		return lineLineSymDifference(linesA, linesB)
	default:
		return geom.NewGeometryCollectionEmpty()
	}
}

// overlayLinePolygon overlays a line with a polygon.
func overlayLinePolygon(a, b geom.Geometry, op Op) geom.Geometry {
	lines := geom.ExtractLineStringsWithRings(a)
	polygons := geom.ExtractPolygons(b)

	switch op {
	case OpIntersection:
		return linePolygonIntersection(lines, polygons)
	case OpUnion:
		// Union of line and polygon is the polygon plus exterior parts of line
		return collectGeometries(b.Clone(), linePolygonDifference(lines, polygons))
	case OpDifference:
		return linePolygonDifference(lines, polygons)
	case OpSymDifference:
		// Parts of line not in polygon, and polygon
		return collectGeometries(b.Clone(), linePolygonDifference(lines, polygons))
	default:
		return geom.NewGeometryCollectionEmpty()
	}
}

// overlayPolygonLine overlays a polygon with a line.
func overlayPolygonLine(a, b geom.Geometry, op Op) geom.Geometry {
	lines := geom.ExtractLineStringsWithRings(b)
	polygons := geom.ExtractPolygons(a)

	switch op {
	case OpIntersection:
		return linePolygonIntersection(lines, polygons)
	case OpUnion:
		return collectGeometries(a.Clone(), linePolygonDifference(lines, polygons))
	case OpDifference:
		// Polygon minus line - line doesn't affect polygon area
		return a.Clone()
	case OpSymDifference:
		return collectGeometries(a.Clone(), linePolygonDifference(lines, polygons))
	default:
		return geom.NewGeometryCollectionEmpty()
	}
}

// overlayPolygonPolygon overlays two polygon geometries.
func overlayPolygonPolygon(a, b geom.Geometry, op Op) geom.Geometry {
	polysA := geom.ExtractPolygons(a)
	polysB := geom.ExtractPolygons(b)

	switch op {
	case OpIntersection:
		return polygonPolygonIntersection(polysA, polysB)
	case OpUnion:
		return polygonPolygonUnion(polysA, polysB)
	case OpDifference:
		return polygonPolygonDifference(polysA, polysB)
	case OpSymDifference:
		return polygonPolygonSymDifference(polysA, polysB)
	default:
		return geom.NewGeometryCollectionEmpty()
	}
}


// createPointResult creates a geometry from a list of points.
func createPointResult(points []*geom.Point) geom.Geometry {
	if len(points) == 0 {
		return geom.NewPointEmpty()
	}
	if len(points) == 1 {
		return points[0]
	}
	return geom.NewMultiPoint(points)
}

// collectGeometries combines geometries into a collection.
func collectGeometries(geoms ...geom.Geometry) geom.Geometry {
	var nonEmpty []geom.Geometry
	for _, g := range geoms {
		if g != nil && !g.IsEmpty() {
			nonEmpty = append(nonEmpty, g)
		}
	}
	if len(nonEmpty) == 0 {
		return geom.NewGeometryCollectionEmpty()
	}
	if len(nonEmpty) == 1 {
		return nonEmpty[0]
	}
	return geom.NewGeometryCollection(nonEmpty)
}
