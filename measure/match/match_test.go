package match

import (
	"math"
	"testing"

	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/measure"
)

func TestAreaSimilarity_Identical(t *testing.T) {
	p := geom.NewPolygon(nil, []geom.XY{{0, 0}, {10, 0}, {10, 10}, {0, 10}, {0, 0}})
	got := measure.AreaSimilarity(p, p)
	if math.Abs(got-1) > 1e-9 {
		t.Fatalf("identical: want 1, got %v", got)
	}
}

func TestAreaSimilarity_Disjoint(t *testing.T) {
	a := geom.NewPolygon(nil, []geom.XY{{0, 0}, {1, 0}, {1, 1}, {0, 1}, {0, 0}})
	b := geom.NewPolygon(nil, []geom.XY{{5, 5}, {6, 5}, {6, 6}, {5, 6}, {5, 5}})
	got := measure.AreaSimilarity(a, b)
	if math.Abs(got) > 1e-9 {
		t.Fatalf("disjoint: want 0, got %v", got)
	}
}

func TestAreaSimilarity_HalfOverlap(t *testing.T) {
	// Two unit squares with horizontal offset 0.5 → intersection area
	// 0.5, union area 1.5, similarity = 1/3.
	a := geom.NewPolygon(nil, []geom.XY{{0, 0}, {1, 0}, {1, 1}, {0, 1}, {0, 0}})
	b := geom.NewPolygon(nil, []geom.XY{{0.5, 0}, {1.5, 0}, {1.5, 1}, {0.5, 1}, {0.5, 0}})
	got := measure.AreaSimilarity(a, b)
	want := 1.0 / 3.0
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("half overlap: want %v, got %v", want, got)
	}
}

func TestAreaSimilarity_Empty(t *testing.T) {
	a := geom.NewEmptyPolygon(nil, geom.LayoutXY)
	b := geom.NewEmptyPolygon(nil, geom.LayoutXY)
	if got := measure.AreaSimilarity(a, b); got != 1 {
		t.Fatalf("both empty: want 1, got %v", got)
	}
	c := geom.NewPolygon(nil, []geom.XY{{0, 0}, {1, 0}, {1, 1}, {0, 1}, {0, 0}})
	if got := measure.AreaSimilarity(a, c); got != 0 {
		t.Fatalf("one empty: want 0, got %v", got)
	}
}
