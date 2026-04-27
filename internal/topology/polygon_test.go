package topology

import (
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
)

func TestPolygonBoundaryLinesIncludesShellsAndHoles(t *testing.T) {
	polygon := geom.NewPolygon(
		mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0),
		[]*geom.LinearRing{
			mustLinearRingXY(2, 2, 8, 2, 8, 8, 2, 8, 2, 2),
		},
	)

	lines := PolygonBoundaryLines([]*geom.Polygon{nil, geom.NewPolygonEmpty(), polygon})
	if len(lines) != 2 {
		t.Fatalf("expected shell and hole boundary lines, got %d", len(lines))
	}
	if lines[0].Length() != 40 {
		t.Fatalf("expected shell length 40, got %v", lines[0].Length())
	}
	if lines[1].Length() != 24 {
		t.Fatalf("expected hole length 24, got %v", lines[1].Length())
	}
}

func TestNodePolygonBoundariesLabelsSharedEdge(t *testing.T) {
	left := geom.NewPolygon(
		mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0),
		nil,
	)
	right := geom.NewPolygon(
		mustLinearRingXY(10, 0, 20, 0, 20, 10, 10, 10, 10, 0),
		nil,
	)

	segments := NodePolygonBoundaries([]*geom.Polygon{left}, []*geom.Polygon{right})

	var shared int
	for _, segment := range segments {
		if segment.InA() && segment.InB() {
			shared++
			if !segment.Start.Equals2D(geom.NewCoordinate(10, 0), geom.DefaultEpsilon) ||
				!segment.End.Equals2D(geom.NewCoordinate(10, 10), geom.DefaultEpsilon) {
				t.Fatalf("unexpected shared edge: %+v", segment)
			}
		}
	}
	if shared != 1 {
		t.Fatalf("expected one shared boundary segment, got %d", shared)
	}
}

func TestNodePolygonBoundariesSplitsCrossingEdges(t *testing.T) {
	square := geom.NewPolygon(
		mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0),
		nil,
	)
	diamond := geom.NewPolygon(
		mustLinearRingXY(5, -5, 15, 5, 5, 15, -5, 5, 5, -5),
		nil,
	)

	segments := NodePolygonBoundaries([]*geom.Polygon{square}, []*geom.Polygon{diamond})

	for _, want := range []geom.Coordinate{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(10, 0),
		geom.NewCoordinate(10, 10),
		geom.NewCoordinate(0, 10),
	} {
		if !hasSegmentEndpoint(segments, want) {
			t.Fatalf("expected noded boundary endpoint at %v", want)
		}
	}
}

func TestNodePolygonBoundariesWithPrecisionSnapsWithoutMutatingInputs(t *testing.T) {
	a := geom.NewPolygon(
		mustLinearRingXY(0.04, 0.04, 10.04, 0.04, 10.04, 10.04, 0.04, 10.04, 0.04, 0.04),
		nil,
	)
	b := geom.NewPolygon(
		mustLinearRingXY(10.04, 0.04, 12.04, 0.04, 12.04, 10.04, 10.04, 10.04, 10.04, 0.04),
		nil,
	)

	segments := NodePolygonBoundariesWithPrecision(
		[]*geom.Polygon{a},
		[]*geom.Polygon{b},
		geom.NewFixedPrecision(1),
	)

	if !hasEndpoint(segments, geom.NewCoordinate(10, 0)) {
		t.Fatalf("expected snapped polygon boundary endpoint at (10,0), got %#v", segments)
	}
	if got := a.ExteriorRing().Coordinates()[0]; got.X != 0.04 || got.Y != 0.04 {
		t.Fatalf("precision noding mutated input polygon coordinate: %v", got)
	}
}

func TestPolygonBoundaryIntersectionDimension(t *testing.T) {
	tests := []struct {
		name string
		a    *geom.Polygon
		b    *geom.Polygon
		dim  geom.Dimension
		ok   bool
	}{
		{
			name: "point touch",
			a:    geom.NewPolygon(mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil),
			b:    geom.NewPolygon(mustLinearRingXY(10, 10, 20, 10, 20, 20, 10, 20, 10, 10), nil),
			dim:  geom.DimensionPoint,
			ok:   true,
		},
		{
			name: "shared edge",
			a:    geom.NewPolygon(mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil),
			b:    geom.NewPolygon(mustLinearRingXY(10, 0, 20, 0, 20, 10, 10, 10, 10, 0), nil),
			dim:  geom.DimensionLine,
			ok:   true,
		},
		{
			name: "disjoint",
			a:    geom.NewPolygon(mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil),
			b:    geom.NewPolygon(mustLinearRingXY(20, 20, 30, 20, 30, 30, 20, 30, 20, 20), nil),
			ok:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dim, ok := PolygonBoundaryIntersectionDimension([]*geom.Polygon{tt.a}, []*geom.Polygon{tt.b})
			if ok != tt.ok {
				t.Fatalf("ok = %v, want %v", ok, tt.ok)
			}
			if ok && dim != tt.dim {
				t.Fatalf("dimension = %v, want %v", dim, tt.dim)
			}
		})
	}
}

func TestRingInteriorPointFindsPointInsideConcaveRing(t *testing.T) {
	ring := geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(6, 0),
		geom.NewCoordinate(6, 6),
		geom.NewCoordinate(4, 6),
		geom.NewCoordinate(4, 2),
		geom.NewCoordinate(2, 2),
		geom.NewCoordinate(2, 6),
		geom.NewCoordinate(0, 6),
		geom.NewCoordinate(0, 0),
	}

	point := RingInteriorPoint(ring)
	linearRing := geom.NewLinearRing(ring)
	if !geom.PointInRing(point, linearRing) {
		t.Fatalf("point %v should be inside concave ring", point)
	}
	if geom.PointOnRing(point, linearRing) {
		t.Fatalf("point %v should not be on ring boundary", point)
	}
}

func TestPolygonInteriorPointAvoidsHole(t *testing.T) {
	polygon := geom.NewPolygon(
		mustLinearRingXY(0, 0, 20, 0, 20, 20, 0, 20, 0, 0),
		[]*geom.LinearRing{
			mustLinearRingXY(5, 5, 15, 5, 15, 15, 5, 15, 5, 5),
		},
	)

	point, ok := PolygonInteriorPoint(polygon)
	if !ok {
		t.Fatal("expected polygon interior point")
	}
	if loc := PointLocationInPolygon(point, polygon); loc != geom.LocationInterior {
		t.Fatalf("expected interior point, got location %v at %v", loc, point)
	}
	if point.X >= 5 && point.X <= 15 && point.Y >= 5 && point.Y <= 15 {
		t.Fatalf("interior point should not be inside hole envelope: %v", point)
	}
}

func hasSegmentEndpoint(segments []NodedLineSegment, coord geom.Coordinate) bool {
	for _, segment := range segments {
		if segment.Start.Equals2D(coord, geom.DefaultEpsilon) ||
			segment.End.Equals2D(coord, geom.DefaultEpsilon) {
			return true
		}
	}
	return false
}
