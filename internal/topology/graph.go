package topology

import (
	"math"

	"github.com/robert-malhotra/go-topology-suite/geom"
)

// PolygonEdgeLabel records point-location labels for both input polygon sets.
type PolygonEdgeLabel struct {
	LocA geom.Location
	LocB geom.Location
}

// PolygonBoundaryEdge is a noded polygon boundary edge labeled on both sides.
type PolygonBoundaryEdge struct {
	Start, End geom.Coordinate
	Sources    LineSource
	Left       PolygonEdgeLabel
	Right      PolygonEdgeLabel
}

// BuildPolygonBoundaryGraph nodes polygon boundaries and labels the left and
// right side of every boundary segment against both input polygon sets.
func BuildPolygonBoundaryGraph(polygonsA, polygonsB []*geom.Polygon) []PolygonBoundaryEdge {
	segments := NodePolygonBoundaries(polygonsA, polygonsB)
	return buildPolygonBoundaryGraph(segments, polygonsA, polygonsB)
}

// BuildPolygonBoundaryGraphWithPrecision applies a precision model to cloned
// polygon coordinates before noding and labeling boundary edges. Input polygons
// are not mutated.
func BuildPolygonBoundaryGraphWithPrecision(polygonsA, polygonsB []*geom.Polygon, pm geom.PrecisionModel) []PolygonBoundaryEdge {
	if pm == nil {
		return BuildPolygonBoundaryGraph(polygonsA, polygonsB)
	}
	preciseA := makePrecisePolygons(polygonsA, pm)
	preciseB := makePrecisePolygons(polygonsB, pm)
	segments := NodePolygonBoundaries(preciseA, preciseB)
	return buildPolygonBoundaryGraph(segments, preciseA, preciseB)
}

func buildPolygonBoundaryGraph(segments []NodedLineSegment, polygonsA, polygonsB []*geom.Polygon) []PolygonBoundaryEdge {
	edges := make([]PolygonBoundaryEdge, 0, len(segments))
	for _, segment := range segments {
		left, right := labelSegmentSides(segment.Start, segment.End, polygonsA, polygonsB)
		edges = append(edges, PolygonBoundaryEdge{
			Start:   segment.Start,
			End:     segment.End,
			Sources: segment.Sources,
			Left:    left,
			Right:   right,
		})
	}
	return edges
}

func labelSegmentSides(start, end geom.Coordinate, polygonsA, polygonsB []*geom.Polygon) (PolygonEdgeLabel, PolygonEdgeLabel) {
	dx := end.X - start.X
	dy := end.Y - start.Y
	length := math.Hypot(dx, dy)
	if length <= geom.DefaultEpsilon {
		exterior := PolygonEdgeLabel{LocA: geom.LocationExterior, LocB: geom.LocationExterior}
		return exterior, exterior
	}

	dx /= length
	dy /= length
	perpX := -dy
	perpY := dx
	offset := math.Max(length*0.1, geom.DefaultEpsilon*1000)
	midX := (start.X + end.X) / 2
	midY := (start.Y + end.Y) / 2

	leftPoint := geom.NewCoordinate(midX+perpX*offset, midY+perpY*offset)
	rightPoint := geom.NewCoordinate(midX-perpX*offset, midY-perpY*offset)

	left := PolygonEdgeLabel{
		LocA: PointLocationInPolygonSet(leftPoint, polygonsA),
		LocB: PointLocationInPolygonSet(leftPoint, polygonsB),
	}
	right := PolygonEdgeLabel{
		LocA: PointLocationInPolygonSet(rightPoint, polygonsA),
		LocB: PointLocationInPolygonSet(rightPoint, polygonsB),
	}
	return left, right
}
