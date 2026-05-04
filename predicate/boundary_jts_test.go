package predicate

import (
	"testing"

	"github.com/terra-geo/terra/wkt"
)

// TestBoundaryContainsCoversSplit verifies the JTS contract:
// Polygon.contains(Point on boundary) = false (interior only).
// Polygon.covers(Point on boundary) = true.
// Same for Polygon-LineString-on-boundary, etc.
func TestBoundaryContainsCoversSplit(t *testing.T) {
	poly, _ := wkt.Unmarshal("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	ptBoundary, _ := wkt.Unmarshal("POINT (5 0)")  // on edge
	ptCorner, _ := wkt.Unmarshal("POINT (0 0)")    // on corner
	ptInterior, _ := wkt.Unmarshal("POINT (5 5)")  // interior
	ptOutside, _ := wkt.Unmarshal("POINT (-1 -1)") // outside

	cases := []struct {
		name string
		fn   func() (bool, error)
		want bool
	}{
		{"contains(poly, edge)", func() (bool, error) { return Contains(poly, ptBoundary) }, false},
		{"contains(poly, corner)", func() (bool, error) { return Contains(poly, ptCorner) }, false},
		{"contains(poly, interior)", func() (bool, error) { return Contains(poly, ptInterior) }, true},
		{"contains(poly, outside)", func() (bool, error) { return Contains(poly, ptOutside) }, false},
		{"covers(poly, edge)", func() (bool, error) { return Covers(poly, ptBoundary) }, true},
		{"covers(poly, corner)", func() (bool, error) { return Covers(poly, ptCorner) }, true},
		{"covers(poly, interior)", func() (bool, error) { return Covers(poly, ptInterior) }, true},
		{"covers(poly, outside)", func() (bool, error) { return Covers(poly, ptOutside) }, false},
	}
	for _, c := range cases {
		got, err := c.fn()
		if err != nil {
			t.Errorf("%s: err %v", c.name, err)
			continue
		}
		if got != c.want {
			t.Errorf("%s: got %v want %v", c.name, got, c.want)
		}
	}
}
