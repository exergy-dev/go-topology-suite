package topology

import (
	"math"

	"github.com/robert-malhotra/go-topology-suite/geom"
)

// PolygonBoundaryLines extracts shell and hole rings from a polygon set as
// line strings suitable for shared noding and graph construction.
func PolygonBoundaryLines(polygons []*geom.Polygon) []*geom.LineString {
	var lines []*geom.LineString
	for _, polygon := range polygons {
		if polygon == nil || polygon.IsEmpty() {
			continue
		}
		lines = append(lines, polygon.ExteriorRing().LineString)
		for i := 0; i < polygon.NumInteriorRings(); i++ {
			lines = append(lines, polygon.InteriorRingN(i).LineString)
		}
	}
	return lines
}

// NodePolygonBoundaries nodes the boundary rings of two polygon sets together.
// The returned segments are labeled by the polygon set whose boundary contains
// each segment. Face labeling and operation-specific selection are intentionally
// separate steps layered on top of this primitive.
func NodePolygonBoundaries(polygonsA, polygonsB []*geom.Polygon) []NodedLineSegment {
	return NodeLineSets(PolygonBoundaryLines(polygonsA), PolygonBoundaryLines(polygonsB))
}

// NodePolygonBoundariesWithPrecision applies a precision model to cloned
// polygon boundary coordinates before noding. Input polygons are not mutated.
func NodePolygonBoundariesWithPrecision(polygonsA, polygonsB []*geom.Polygon, pm geom.PrecisionModel) []NodedLineSegment {
	return NodeLineSetsWithPrecision(PolygonBoundaryLines(polygonsA), PolygonBoundaryLines(polygonsB), pm)
}

func makePrecisePolygons(polygons []*geom.Polygon, pm geom.PrecisionModel) []*geom.Polygon {
	precise := make([]*geom.Polygon, 0, len(polygons))
	for _, polygon := range polygons {
		if polygon == nil {
			continue
		}
		if polygon.IsEmpty() {
			precise = append(precise, geom.NewPolygonEmpty())
			continue
		}
		precise = append(precise, geom.NewPolygon(
			makePreciseRing(polygon.ExteriorRing(), pm),
			makePreciseInteriorRings(polygon, pm),
		))
	}
	return precise
}

func makePreciseInteriorRings(polygon *geom.Polygon, pm geom.PrecisionModel) []*geom.LinearRing {
	holes := make([]*geom.LinearRing, 0, polygon.NumInteriorRings())
	for i := 0; i < polygon.NumInteriorRings(); i++ {
		hole := polygon.InteriorRingN(i)
		if hole == nil {
			continue
		}
		holes = append(holes, makePreciseRing(hole, pm))
	}
	return holes
}

func makePreciseRing(ring *geom.LinearRing, pm geom.PrecisionModel) *geom.LinearRing {
	if ring == nil {
		return nil
	}
	coords := ring.Coordinates()
	geom.MakePreciseSequence(pm, coords)
	return geom.NewLinearRing(coords)
}

// PolygonBoundaryIntersectionDimension reports the highest dimension shared by
// two polygon boundary sets. The bool is false when the boundaries are disjoint.
func PolygonBoundaryIntersectionDimension(polygonsA, polygonsB []*geom.Polygon) (geom.Dimension, bool) {
	nodedSegments := NodePolygonBoundaries(polygonsA, polygonsB)
	for _, segment := range nodedSegments {
		if segment.InA() && segment.InB() {
			return geom.DimensionLine, true
		}
	}

	boundaryLinesA := PolygonBoundaryLines(polygonsA)
	boundaryLinesB := PolygonBoundaryLines(polygonsB)
	for _, segment := range nodedSegments {
		if pointOnLineSet(segment.Start, boundaryLinesA) && pointOnLineSet(segment.Start, boundaryLinesB) {
			return geom.DimensionPoint, true
		}
		if pointOnLineSet(segment.End, boundaryLinesA) && pointOnLineSet(segment.End, boundaryLinesB) {
			return geom.DimensionPoint, true
		}
	}
	return 0, false
}

func pointOnLineSet(point geom.Coordinate, lines []*geom.LineString) bool {
	for _, line := range lines {
		coords := line.Coordinates()
		for i := 0; i < len(coords)-1; i++ {
			if geom.PointOnSegment(point, coords[i], coords[i+1]) {
				return true
			}
		}
	}
	return false
}

// RingInteriorPoint returns a representative point inside a ring and away from
// its boundary when a non-boundary candidate can be found.
func RingInteriorPoint(ring geom.CoordinateSequence) geom.Coordinate {
	if len(ring) == 0 {
		return geom.Coordinate{}
	}
	if len(ring) < 3 {
		return ring[0]
	}

	if pt, ok := ringAreaCentroid(ring); ok && pointStrictlyInRing(pt, ring) {
		return pt
	}

	offsets := ringOffsets(ring)
	for _, offset := range offsets {
		for i := 0; i < len(ring)-1; i++ {
			start := ring[i]
			end := ring[i+1]
			dx := end.X - start.X
			dy := end.Y - start.Y
			length := math.Hypot(dx, dy)
			if length <= geom.DefaultEpsilon {
				continue
			}

			mid := geom.NewCoordinate((start.X+end.X)/2, (start.Y+end.Y)/2)
			nx := -dy / length
			ny := dx / length
			candidates := []geom.Coordinate{
				geom.NewCoordinate(mid.X+nx*offset, mid.Y+ny*offset),
				geom.NewCoordinate(mid.X-nx*offset, mid.Y-ny*offset),
			}
			for _, candidate := range candidates {
				if pointStrictlyInRing(candidate, ring) {
					return candidate
				}
			}
		}
	}

	if pt, ok := ringVertexAverage(ring); ok {
		return pt
	}
	return ring[0]
}

// PolygonInteriorPoint returns a representative point in the polygon interior
// when one can be found. The boolean is false for nil, empty, or collapsed
// polygons where no non-boundary interior candidate is found.
func PolygonInteriorPoint(poly *geom.Polygon) (geom.Coordinate, bool) {
	if poly == nil || poly.IsEmpty() {
		return geom.Coordinate{}, false
	}

	candidates := []geom.Coordinate{RingInteriorPoint(poly.ExteriorRing().Coordinates())}
	candidates = append(candidates, polygonEdgeInteriorCandidates(poly)...)
	for _, candidate := range candidates {
		if PointLocationInPolygon(candidate, poly) == geom.LocationInterior {
			return candidate, true
		}
	}
	return geom.Coordinate{}, false
}

func polygonEdgeInteriorCandidates(poly *geom.Polygon) []geom.Coordinate {
	shell := poly.ExteriorRing().Coordinates()
	offsets := ringOffsets(shell)
	candidates := make([]geom.Coordinate, 0, len(shell)*len(offsets)*2)
	for _, offset := range offsets {
		for i := 0; i < len(shell)-1; i++ {
			start := shell[i]
			end := shell[i+1]
			dx := end.X - start.X
			dy := end.Y - start.Y
			length := math.Hypot(dx, dy)
			if length <= geom.DefaultEpsilon {
				continue
			}
			mid := geom.NewCoordinate((start.X+end.X)/2, (start.Y+end.Y)/2)
			nx := -dy / length
			ny := dx / length
			candidates = append(candidates,
				geom.NewCoordinate(mid.X+nx*offset, mid.Y+ny*offset),
				geom.NewCoordinate(mid.X-nx*offset, mid.Y-ny*offset),
			)
		}
	}
	return candidates
}

func ringAreaCentroid(ring geom.CoordinateSequence) (geom.Coordinate, bool) {
	var twiceArea, cx, cy float64
	n := ringCoordinateCount(ring)
	if n < 3 {
		return geom.Coordinate{}, false
	}
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		cross := ring[i].X*ring[j].Y - ring[j].X*ring[i].Y
		twiceArea += cross
		cx += (ring[i].X + ring[j].X) * cross
		cy += (ring[i].Y + ring[j].Y) * cross
	}
	if math.Abs(twiceArea) <= geom.DefaultEpsilon {
		return geom.Coordinate{}, false
	}
	return geom.NewCoordinate(cx/(3*twiceArea), cy/(3*twiceArea)), true
}

func ringOffsets(ring geom.CoordinateSequence) []float64 {
	minX, minY := ring[0].X, ring[0].Y
	maxX, maxY := ring[0].X, ring[0].Y
	for _, coord := range ring {
		minX = math.Min(minX, coord.X)
		minY = math.Min(minY, coord.Y)
		maxX = math.Max(maxX, coord.X)
		maxY = math.Max(maxY, coord.Y)
	}
	diag := math.Hypot(maxX-minX, maxY-minY)
	if diag <= geom.DefaultEpsilon {
		diag = 1
	}
	return []float64{
		math.Max(diag*1e-7, geom.DefaultEpsilon*100),
		math.Max(diag*1e-5, geom.DefaultEpsilon*1000),
		math.Max(diag*1e-3, geom.DefaultEpsilon*10000),
	}
}

func ringVertexAverage(ring geom.CoordinateSequence) (geom.Coordinate, bool) {
	n := ringCoordinateCount(ring)
	if n == 0 {
		return geom.Coordinate{}, false
	}
	var x, y float64
	for i := 0; i < n; i++ {
		x += ring[i].X
		y += ring[i].Y
	}
	return geom.NewCoordinate(x/float64(n), y/float64(n)), true
}

func ringCoordinateCount(ring geom.CoordinateSequence) int {
	n := len(ring)
	if n > 1 && ring.IsClosed(geom.DefaultEpsilon) {
		n--
	}
	return n
}

func pointStrictlyInRing(pt geom.Coordinate, ring geom.CoordinateSequence) bool {
	linearRing := geom.NewLinearRing(ring)
	return !geom.PointOnRing(pt, linearRing) && geom.PointInRing(pt, linearRing)
}
