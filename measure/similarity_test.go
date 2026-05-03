package measure

import (
	"math"
	"testing"

	"github.com/terra-geo/terra/geom"
)

func TestHausdorffSimilarity_Identical(t *testing.T) {
	p := geom.NewLineString(nil, []geom.XY{{0, 0}, {10, 0}, {10, 10}})
	if got := HausdorffSimilarity(p, p); math.Abs(got-1) > 1e-9 {
		t.Fatalf("identical: want 1, got %v", got)
	}
}

func TestHausdorffSimilarity_Disjoint(t *testing.T) {
	a := geom.NewPoint(nil, geom.XY{0, 0})
	b := geom.NewPoint(nil, geom.XY{10, 0})
	got := HausdorffSimilarity(a, b)
	// Two points: the combined envelope is degenerate (height 0) so
	// diagonal = 10. Hausdorff distance = 10. Similarity = 1 - 10/10 = 0.
	if math.Abs(got) > 1e-9 {
		t.Fatalf("disjoint points: want 0, got %v", got)
	}
}

func TestHausdorffSimilarity_NearlyIdentical(t *testing.T) {
	// A tiny perturbation should yield a similarity close to (but
	// less than) 1.
	a := geom.NewLineString(nil, []geom.XY{{0, 0}, {10, 0}, {10, 10}})
	b := geom.NewLineString(nil, []geom.XY{{0, 0}, {10, 0}, {10, 10.1}})
	got := HausdorffSimilarity(a, b)
	if got <= 0.95 || got >= 1 {
		t.Fatalf("nearly identical: want in (0.95, 1), got %v", got)
	}
}

func TestHausdorffSimilarity_BothEmpty(t *testing.T) {
	a := geom.NewEmptyPolygon(nil, geom.LayoutXY)
	b := geom.NewEmptyPolygon(nil, geom.LayoutXY)
	if got := HausdorffSimilarity(a, b); got != 1 {
		t.Fatalf("both empty: want 1, got %v", got)
	}
}

func TestHausdorffSimilarity_OneEmpty(t *testing.T) {
	a := geom.NewEmptyPolygon(nil, geom.LayoutXY)
	b := geom.NewPoint(nil, geom.XY{0, 0})
	if got := HausdorffSimilarity(a, b); got != 0 {
		t.Fatalf("one empty: want 0, got %v", got)
	}
}

func TestHausdorffSimilarity_NilInputs(t *testing.T) {
	a := geom.NewPoint(nil, geom.XY{0, 0})
	if got := HausdorffSimilarity(nil, a); !math.IsNaN(got) {
		t.Fatalf("nil input: want NaN, got %v", got)
	}
	if got := HausdorffSimilarity(a, nil); !math.IsNaN(got) {
		t.Fatalf("nil input: want NaN, got %v", got)
	}
}
