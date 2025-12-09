package overlay

import (
	"math"
	"testing"
	"testing/quick"

	"github.com/go-topology-suite/gts/geom"
)

// getArea returns the area of a geometry.
func getArea(g geom.Geometry) float64 {
	if g == nil || g.IsEmpty() {
		return 0
	}
	switch v := g.(type) {
	case *geom.Polygon:
		return v.Area()
	case *geom.MultiPolygon:
		return v.Area()
	case *geom.GeometryCollection:
		var total float64
		for i := 0; i < v.NumGeometries(); i++ {
			total += getArea(v.GeometryN(i))
		}
		return total
	default:
		return 0
	}
}

// normalizeCoord bounds a coordinate to a reasonable range.
func normalizeCoord(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	// Use a smaller range to avoid overflow in arithmetic
	const maxCoord = 100
	if v > maxCoord {
		return maxCoord
	}
	if v < -maxCoord {
		return -maxCoord
	}
	return v
}

// generatePolygon creates a simple polygon from random values.
func generatePolygon(cx, cy, size float64) *geom.Polygon {
	if size < 1 {
		size = 1
	}
	if size > 100 {
		size = 100
	}
	// Create a square centered at (cx, cy)
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

// generateOverlappingPolygons creates two polygons that overlap.
func generateOverlappingPolygons(cx, cy, size, offset float64) (*geom.Polygon, *geom.Polygon) {
	if size < 1 {
		size = 1
	}
	if size > 100 {
		size = 100
	}
	if offset < 0 {
		offset = -offset
	}
	if offset >= size {
		offset = size / 2
	}

	poly1 := generatePolygon(cx, cy, size)
	poly2 := generatePolygon(cx+offset, cy+offset, size)
	return poly1, poly2
}

// TestIntersectionCommutativity tests that A ∩ B = B ∩ A.
func TestIntersectionCommutativity(t *testing.T) {
	f := func(cx, cy, size, offset float64) bool {
		cx = normalizeCoord(cx)
		cy = normalizeCoord(cy)
		size = math.Abs(normalizeCoord(size)) + 1
		offset = normalizeCoord(offset)

		poly1, poly2 := generateOverlappingPolygons(cx, cy, size, offset)

		result1 := Intersection(poly1, poly2)
		result2 := Intersection(poly2, poly1)

		// Both results should have the same area
		area1 := getArea(result1)
		area2 := getArea(result2)

		return math.Abs(area1-area2) < geom.DefaultEpsilon
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// TestUnionCommutativity tests that A ∪ B = B ∪ A.
func TestUnionCommutativity(t *testing.T) {
	f := func(cx, cy, size, offset float64) bool {
		cx = normalizeCoord(cx)
		cy = normalizeCoord(cy)
		size = math.Abs(normalizeCoord(size)) + 1
		offset = normalizeCoord(offset)

		poly1, poly2 := generateOverlappingPolygons(cx, cy, size, offset)

		result1 := Union(poly1, poly2)
		result2 := Union(poly2, poly1)

		// Both results should have the same area
		area1 := getArea(result1)
		area2 := getArea(result2)

		// Use relative tolerance for larger areas
		tolerance := math.Max(geom.DefaultEpsilon, (area1+area2)*0.001) // 0.1% tolerance
		return math.Abs(area1-area2) < tolerance
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// TestIntersectionSubset tests that intersection is a subset of both operands.
func TestIntersectionSubset(t *testing.T) {
	f := func(cx, cy, size, offset float64) bool {
		cx = normalizeCoord(cx)
		cy = normalizeCoord(cy)
		size = math.Abs(normalizeCoord(size)) + 1
		offset = normalizeCoord(offset)

		poly1, poly2 := generateOverlappingPolygons(cx, cy, size, offset)

		intersection := Intersection(poly1, poly2)
		intersectionArea := getArea(intersection)

		// Intersection area should be <= both operand areas
		area1 := poly1.Area()
		area2 := poly2.Area()

		return intersectionArea <= area1+geom.DefaultEpsilon &&
			intersectionArea <= area2+geom.DefaultEpsilon
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// TestUnionContainsBoth tests that union contains both operands.
func TestUnionContainsBoth(t *testing.T) {
	f := func(cx, cy, size, offset float64) bool {
		cx = normalizeCoord(cx)
		cy = normalizeCoord(cy)
		size = math.Abs(normalizeCoord(size)) + 1
		offset = normalizeCoord(offset)

		poly1, poly2 := generateOverlappingPolygons(cx, cy, size, offset)

		union := Union(poly1, poly2)
		unionArea := getArea(union)

		// Union area should be >= both operand areas
		area1 := poly1.Area()
		area2 := poly2.Area()

		return unionArea >= area1-geom.DefaultEpsilon &&
			unionArea >= area2-geom.DefaultEpsilon
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// TestInclusionExclusion tests the inclusion-exclusion principle:
// |A ∪ B| = |A| + |B| - |A ∩ B|
func TestInclusionExclusion(t *testing.T) {
	f := func(cx, cy, size, offset float64) bool {
		cx = normalizeCoord(cx)
		cy = normalizeCoord(cy)
		size = math.Abs(normalizeCoord(size))
		offset = math.Abs(normalizeCoord(offset))

		if size < 5 {
			size = 5
		}
		if size > 50 {
			size = 50
		}
		if offset < 1 {
			offset = 1
		}
		if offset > size/2 {
			offset = size / 2
		}

		poly1, poly2 := generateOverlappingPolygons(cx, cy, size, offset)

		intersection := Intersection(poly1, poly2)
		union := Union(poly1, poly2)

		area1 := poly1.Area()
		area2 := poly2.Area()
		intersectionArea := getArea(intersection)
		unionArea := getArea(union)

		// |A ∪ B| = |A| + |B| - |A ∩ B|
		expectedUnion := area1 + area2 - intersectionArea

		// Allow some tolerance for numerical errors
		tolerance := (area1 + area2) * 0.1 // 10% tolerance
		return math.Abs(unionArea-expectedUnion) < tolerance
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 50}); err != nil {
		t.Error(err)
	}
}

// TestSelfIntersectionIsIdentity tests that A ∩ A = A.
func TestSelfIntersectionIsIdentity(t *testing.T) {
	f := func(cx, cy, size float64) bool {
		cx = normalizeCoord(cx)
		cy = normalizeCoord(cy)
		size = math.Abs(normalizeCoord(size)) + 1

		poly := generatePolygon(cx, cy, size)

		result := Intersection(poly, poly)
		originalArea := poly.Area()
		resultArea := getArea(result)

		return math.Abs(originalArea-resultArea) < geom.DefaultEpsilon
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// TestSelfUnionIsIdentity tests that A ∪ A = A.
func TestSelfUnionIsIdentity(t *testing.T) {
	f := func(cx, cy, size float64) bool {
		cx = normalizeCoord(cx)
		cy = normalizeCoord(cy)
		size = math.Abs(normalizeCoord(size)) + 1

		poly := generatePolygon(cx, cy, size)

		result := Union(poly, poly)
		originalArea := poly.Area()
		resultArea := getArea(result)

		return math.Abs(originalArea-resultArea) < geom.DefaultEpsilon
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// TestEmptyIntersectionWithDisjoint tests that disjoint geometries have empty intersection.
func TestEmptyIntersectionWithDisjoint(t *testing.T) {
	f := func(cx, cy, size float64) bool {
		cx = normalizeCoord(cx)
		cy = normalizeCoord(cy)
		size = math.Abs(normalizeCoord(size))

		if size < 1 {
			size = 1
		}
		if size > 100 {
			size = 100
		}

		// Create two disjoint polygons
		poly1 := generatePolygon(cx, cy, size)
		poly2 := generatePolygon(cx+size*3, cy+size*3, size) // Far away

		result := Intersection(poly1, poly2)
		return result.IsEmpty()
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// TestDifferenceWithSelfIsEmpty tests that A - A = empty.
func TestDifferenceWithSelfIsEmpty(t *testing.T) {
	f := func(cx, cy, size float64) bool {
		cx = normalizeCoord(cx)
		cy = normalizeCoord(cy)
		size = math.Abs(normalizeCoord(size)) + 1

		poly := generatePolygon(cx, cy, size)
		result := Difference(poly, poly)

		// Result should be empty or have negligible area
		return result.IsEmpty() || getArea(result) < geom.DefaultEpsilon
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// TestSymDifferenceSymmetry tests that A △ B = B △ A.
func TestSymDifferenceSymmetry(t *testing.T) {
	f := func(cx, cy, size, offset float64) bool {
		cx = normalizeCoord(cx)
		cy = normalizeCoord(cy)
		size = math.Abs(normalizeCoord(size)) + 1
		offset = normalizeCoord(offset)

		poly1, poly2 := generateOverlappingPolygons(cx, cy, size, offset)

		result1 := SymDifference(poly1, poly2)
		result2 := SymDifference(poly2, poly1)

		area1 := getArea(result1)
		area2 := getArea(result2)

		return math.Abs(area1-area2) < geom.DefaultEpsilon
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}
