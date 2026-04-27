package topology

import (
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
)

func TestPointLocationInPolygonWithHole(t *testing.T) {
	polygon := geom.NewPolygon(
		mustLinearRingXY(0, 0, 20, 0, 20, 20, 0, 20, 0, 0),
		[]*geom.LinearRing{
			mustLinearRingXY(5, 5, 15, 5, 15, 15, 5, 15, 5, 5),
		},
	)

	tests := []struct {
		name string
		pt   geom.Coordinate
		want geom.Location
	}{
		{name: "shell interior", pt: geom.NewCoordinate(2, 2), want: geom.LocationInterior},
		{name: "shell boundary", pt: geom.NewCoordinate(0, 10), want: geom.LocationBoundary},
		{name: "hole interior", pt: geom.NewCoordinate(10, 10), want: geom.LocationExterior},
		{name: "hole boundary", pt: geom.NewCoordinate(5, 10), want: geom.LocationBoundary},
		{name: "outside", pt: geom.NewCoordinate(30, 30), want: geom.LocationExterior},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PointLocationInPolygon(tt.pt, polygon); got != tt.want {
				t.Fatalf("PointLocationInPolygon() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPointLocationMultiGeometries(t *testing.T) {
	multiPoint := geom.NewMultiPoint([]*geom.Point{
		geom.NewPointEmpty(),
		geom.NewPoint(1, 1),
	})
	if got := PointLocation(geom.NewCoordinate(0, 0), multiPoint); got != geom.LocationExterior {
		t.Fatalf("empty point component should not locate as interior, got %v", got)
	}
	if got := PointLocation(geom.NewCoordinate(1, 1), multiPoint); got != geom.LocationInterior {
		t.Fatalf("matching point should locate as interior, got %v", got)
	}

	multiPolygon := geom.NewMultiPolygon([]*geom.Polygon{
		geom.NewPolygon(mustLinearRingXY(0, 0, 5, 0, 5, 5, 0, 5, 0, 0), nil),
		geom.NewPolygon(mustLinearRingXY(10, 0, 15, 0, 15, 5, 10, 5, 10, 0), nil),
	})
	if got := PointLocation(geom.NewCoordinate(5, 2), multiPolygon); got != geom.LocationBoundary {
		t.Fatalf("boundary point should locate as boundary, got %v", got)
	}
	if got := PointLocation(geom.NewCoordinate(12, 2), multiPolygon); got != geom.LocationInterior {
		t.Fatalf("interior point should locate as interior, got %v", got)
	}
}

func TestPointLocationInPolygonSet(t *testing.T) {
	polygons := []*geom.Polygon{
		geom.NewPolygon(mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil),
		geom.NewPolygon(mustLinearRingXY(10, 0, 20, 0, 20, 10, 10, 10, 10, 0), nil),
	}

	tests := []struct {
		name string
		pt   geom.Coordinate
		want geom.Location
	}{
		{name: "component interior", pt: geom.NewCoordinate(5, 5), want: geom.LocationInterior},
		{name: "shared edge interior", pt: geom.NewCoordinate(10, 5), want: geom.LocationInterior},
		{name: "outside set", pt: geom.NewCoordinate(30, 30), want: geom.LocationExterior},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PointLocationInPolygonSet(tt.pt, polygons); got != tt.want {
				t.Fatalf("PointLocationInPolygonSet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPointLocationInPolygonSetSharedEdgeIsInterior(t *testing.T) {
	left := geom.NewPolygon(
		mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0),
		nil,
	)
	right := geom.NewPolygon(
		mustLinearRingXY(10, 0, 20, 0, 20, 10, 10, 10, 10, 0),
		nil,
	)

	loc := PointLocationInPolygonSet(geom.NewCoordinate(10, 5), []*geom.Polygon{left, right})
	if loc != geom.LocationInterior {
		t.Fatalf("expected shared polygon-set edge to be interior, got %v", loc)
	}

	loc = PointLocationInPolygonSet(geom.NewCoordinate(10, 10), []*geom.Polygon{left, right})
	if loc != geom.LocationBoundary {
		t.Fatalf("expected outer shared endpoint to remain boundary, got %v", loc)
	}
}

func mustLinearRingXY(values ...float64) *geom.LinearRing {
	seq, err := geom.NewCoordinateSequenceXY(values...)
	if err != nil {
		panic(err)
	}
	return geom.NewLinearRing(seq)
}
