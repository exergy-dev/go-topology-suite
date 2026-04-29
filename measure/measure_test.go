package measure

import (
	"math"
	"testing"

	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/wkt"
)

func mustParse(t *testing.T, s string) geom.Geometry {
	t.Helper()
	g, err := wkt.Unmarshal(s)
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func almostEqual(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestDistancePointPoint(t *testing.T) {
	a := mustParse(t, "POINT (0 0)")
	b := mustParse(t, "POINT (3 4)")
	d, err := Distance(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if d != 5 {
		t.Errorf("Distance = %v, want 5", d)
	}
}

func TestDistancePointToLine(t *testing.T) {
	p := mustParse(t, "POINT (0 1)")
	ls := mustParse(t, "LINESTRING (0 0, 10 0)")
	d, _ := Distance(p, ls)
	if !almostEqual(d, 1) {
		t.Errorf("perpendicular distance = %v, want 1", d)
	}
}

func TestDistanceEmpty(t *testing.T) {
	a := mustParse(t, "POINT EMPTY")
	b := mustParse(t, "POINT (1 2)")
	d, err := Distance(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if d != 0 {
		t.Errorf("empty-input distance should be 0, got %v", d)
	}
}

func TestLength(t *testing.T) {
	cases := []struct {
		wkt  string
		want float64
	}{
		{"LINESTRING (0 0, 3 0, 3 4)", 7},
		{"POINT (1 2)", 0},
		{"POLYGON ((0 0, 0 1, 1 1, 1 0, 0 0))", 4}, // perimeter
		{"MULTILINESTRING ((0 0, 1 0), (0 1, 0 4))", 4},
	}
	for _, tc := range cases {
		g := mustParse(t, tc.wkt)
		got := Length(g)
		if !almostEqual(got, tc.want) {
			t.Errorf("Length(%s) = %v, want %v", tc.wkt, got, tc.want)
		}
	}
}

func TestArea(t *testing.T) {
	cases := []struct {
		wkt  string
		want float64
	}{
		{"POLYGON ((0 0, 0 10, 10 10, 10 0, 0 0))", 100},
		{"POLYGON ((0 0, 0 10, 10 10, 10 0, 0 0), (2 2, 2 4, 4 4, 4 2, 2 2))", 96}, // 100 - 4
		{"MULTIPOLYGON (((0 0, 0 1, 1 1, 1 0, 0 0)), ((10 10, 10 11, 11 11, 11 10, 10 10)))", 2},
		{"LINESTRING (0 0, 1 1)", 0},
		{"POINT (1 2)", 0},
	}
	for _, tc := range cases {
		g := mustParse(t, tc.wkt)
		got := Area(g)
		if !almostEqual(got, tc.want) {
			t.Errorf("Area(%s) = %v, want %v", tc.wkt, got, tc.want)
		}
	}
}

func TestCentroidPoint(t *testing.T) {
	g := mustParse(t, "POINT (3 4)")
	c := Centroid(g)
	if c.XY().X != 3 || c.XY().Y != 4 {
		t.Errorf("centroid of point = %+v", c.XY())
	}
}

func TestCentroidLineString(t *testing.T) {
	g := mustParse(t, "LINESTRING (0 0, 10 0)")
	c := Centroid(g)
	if !almostEqual(c.XY().X, 5) || !almostEqual(c.XY().Y, 0) {
		t.Errorf("centroid line = %+v", c.XY())
	}
}

func TestCentroidSquarePolygon(t *testing.T) {
	g := mustParse(t, "POLYGON ((0 0, 0 10, 10 10, 10 0, 0 0))")
	c := Centroid(g)
	if !almostEqual(c.XY().X, 5) || !almostEqual(c.XY().Y, 5) {
		t.Errorf("square centroid = %+v", c.XY())
	}
}

func TestCentroidMultiPoint(t *testing.T) {
	g := mustParse(t, "MULTIPOINT ((0 0), (4 0), (0 4))")
	c := Centroid(g)
	want := geom.XY{X: 4.0 / 3, Y: 4.0 / 3}
	if !almostEqual(c.XY().X, want.X) || !almostEqual(c.XY().Y, want.Y) {
		t.Errorf("multipoint centroid = %+v, want %+v", c.XY(), want)
	}
}
