package hull

import (
	"testing"

	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/wkt"
)

func TestSquareHull(t *testing.T) {
	pts := geom.NewMultiPoint(nil, []geom.XY{
		{X: 0, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 1}, {X: 1, Y: 0}, {X: 0.5, Y: 0.5},
	})
	hull := ConvexHull(pts)
	if hull.Type() != geom.PolygonType {
		t.Fatalf("hull type = %v", hull.Type())
	}
	got, _ := wkt.Marshal(hull)
	want := "POLYGON ((0 0, 1 0, 1 1, 0 1, 0 0))"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestCollinearHull(t *testing.T) {
	pts := geom.NewMultiPoint(nil, []geom.XY{
		{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0},
	})
	hull := ConvexHull(pts)
	if hull.Type() != geom.LineStringType {
		t.Errorf("collinear hull should be LineString, got %v", hull.Type())
	}
}

func TestSinglePointHull(t *testing.T) {
	pts := geom.NewMultiPoint(nil, []geom.XY{{X: 5, Y: 5}})
	hull := ConvexHull(pts)
	if hull.Type() != geom.PointType {
		t.Errorf("single-point hull = %v", hull.Type())
	}
}

func TestEmptyHull(t *testing.T) {
	pts := geom.NewMultiPoint(nil, nil)
	hull := ConvexHull(pts)
	if !hull.IsEmpty() {
		t.Errorf("empty hull should be empty")
	}
}

func TestHullOfPolygon(t *testing.T) {
	// L-shaped polygon's hull is its bounding box.
	g, _ := wkt.Unmarshal("POLYGON ((0 0, 0 4, 2 4, 2 2, 4 2, 4 0, 0 0))")
	hull := ConvexHull(g)
	got, _ := wkt.Marshal(hull)
	want := "POLYGON ((0 0, 4 0, 4 2, 2 4, 0 4, 0 0))"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}
