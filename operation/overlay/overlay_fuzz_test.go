package overlay

import (
	"testing"

	"github.com/go-topology-suite/gts/geom"
)

// createTestPolygon creates a simple square polygon for testing.
func createTestPolygon(cx, cy, size float64) *geom.Polygon {
	if size < 0.1 {
		size = 0.1
	}
	if size > 1000 {
		size = 1000
	}
	half := size / 2
	shell := geom.NewLinearRingXY(
		cx-half, cy-half,
		cx+half, cy-half,
		cx+half, cy+half,
		cx-half, cy+half,
		cx-half, cy-half,
	)
	return geom.NewPolygon(shell, nil)
}

// FuzzIntersectionPolygons tests that polygon intersection doesn't panic.
func FuzzIntersectionPolygons(f *testing.F) {
	// Add seed corpus
	f.Add(0.0, 0.0, 10.0, 5.0, 5.0, 10.0)
	f.Add(-10.0, -10.0, 20.0, 10.0, 10.0, 20.0)
	f.Add(0.0, 0.0, 5.0, 100.0, 100.0, 5.0)

	f.Fuzz(func(t *testing.T, cx1, cy1, size1, cx2, cy2, size2 float64) {
		poly1 := createTestPolygon(cx1, cy1, size1)
		poly2 := createTestPolygon(cx2, cy2, size2)

		result := Intersection(poly1, poly2)
		if result == nil {
			t.Error("Intersection returned nil")
		}
	})
}

// FuzzUnionPolygons tests that polygon union doesn't panic.
func FuzzUnionPolygons(f *testing.F) {
	// Add seed corpus
	f.Add(0.0, 0.0, 10.0, 5.0, 5.0, 10.0)
	f.Add(-10.0, -10.0, 20.0, 10.0, 10.0, 20.0)
	f.Add(0.0, 0.0, 5.0, 100.0, 100.0, 5.0)

	f.Fuzz(func(t *testing.T, cx1, cy1, size1, cx2, cy2, size2 float64) {
		poly1 := createTestPolygon(cx1, cy1, size1)
		poly2 := createTestPolygon(cx2, cy2, size2)

		result := Union(poly1, poly2)
		if result == nil {
			t.Error("Union returned nil")
		}
	})
}

// FuzzDifferencePolygons tests that polygon difference doesn't panic.
func FuzzDifferencePolygons(f *testing.F) {
	// Add seed corpus
	f.Add(0.0, 0.0, 10.0, 5.0, 5.0, 10.0)
	f.Add(-10.0, -10.0, 20.0, 10.0, 10.0, 20.0)
	f.Add(0.0, 0.0, 5.0, 100.0, 100.0, 5.0)

	f.Fuzz(func(t *testing.T, cx1, cy1, size1, cx2, cy2, size2 float64) {
		poly1 := createTestPolygon(cx1, cy1, size1)
		poly2 := createTestPolygon(cx2, cy2, size2)

		result := Difference(poly1, poly2)
		if result == nil {
			t.Error("Difference returned nil")
		}
	})
}

// FuzzSymDifferencePolygons tests that polygon symmetric difference doesn't panic.
func FuzzSymDifferencePolygons(f *testing.F) {
	// Add seed corpus
	f.Add(0.0, 0.0, 10.0, 5.0, 5.0, 10.0)
	f.Add(-10.0, -10.0, 20.0, 10.0, 10.0, 20.0)
	f.Add(0.0, 0.0, 5.0, 100.0, 100.0, 5.0)

	f.Fuzz(func(t *testing.T, cx1, cy1, size1, cx2, cy2, size2 float64) {
		poly1 := createTestPolygon(cx1, cy1, size1)
		poly2 := createTestPolygon(cx2, cy2, size2)

		result := SymDifference(poly1, poly2)
		if result == nil {
			t.Error("SymDifference returned nil")
		}
	})
}

// FuzzIntersectionPointPolygon tests point-polygon intersection.
func FuzzIntersectionPointPolygon(f *testing.F) {
	f.Add(5.0, 5.0, 0.0, 0.0, 10.0)
	f.Add(15.0, 15.0, 0.0, 0.0, 10.0)
	f.Add(0.0, 0.0, -10.0, -10.0, 5.0)

	f.Fuzz(func(t *testing.T, px, py, cx, cy, size float64) {
		p := geom.NewPoint(px, py)
		poly := createTestPolygon(cx, cy, size)

		result := Intersection(p, poly)
		if result == nil {
			t.Error("Intersection returned nil")
		}
	})
}

// FuzzIntersectionLinePolygon tests line-polygon intersection.
func FuzzIntersectionLinePolygon(f *testing.F) {
	f.Add(5.0, -5.0, 5.0, 15.0, 0.0, 0.0, 10.0)
	f.Add(0.0, 0.0, 10.0, 10.0, 5.0, 5.0, 10.0)

	f.Fuzz(func(t *testing.T, x1, y1, x2, y2, cx, cy, size float64) {
		ls := geom.NewLineStringXY(x1, y1, x2, y2)
		poly := createTestPolygon(cx, cy, size)

		result := Intersection(ls, poly)
		if result == nil {
			t.Error("Intersection returned nil")
		}
	})
}

// FuzzOverlayOp tests all overlay operations.
func FuzzOverlayOp(f *testing.F) {
	f.Add(0.0, 0.0, 10.0, 5.0, 5.0, 10.0, 0)
	f.Add(0.0, 0.0, 10.0, 5.0, 5.0, 10.0, 1)
	f.Add(0.0, 0.0, 10.0, 5.0, 5.0, 10.0, 2)
	f.Add(0.0, 0.0, 10.0, 5.0, 5.0, 10.0, 3)

	f.Fuzz(func(t *testing.T, cx1, cy1, size1, cx2, cy2, size2 float64, op int) {
		poly1 := createTestPolygon(cx1, cy1, size1)
		poly2 := createTestPolygon(cx2, cy2, size2)

		var result geom.Geometry
		switch Op(op % 4) {
		case OpIntersection:
			result = Intersection(poly1, poly2)
		case OpUnion:
			result = Union(poly1, poly2)
		case OpDifference:
			result = Difference(poly1, poly2)
		case OpSymDifference:
			result = SymDifference(poly1, poly2)
		}

		if result == nil {
			t.Error("Overlay returned nil")
		}
	})
}
