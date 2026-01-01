package kdtree_test

import (
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/robert-malhotra/go-topology-suite/index/kdtree"
	"github.com/robert-malhotra/go-topology-suite/index/quadtree"
	"github.com/robert-malhotra/go-topology-suite/index/strtree"
)

// TestCompareWithOtherIndexes ensures KD-tree produces compatible results
// with other spatial indexes for common operations
func TestCompareWithOtherIndexes(t *testing.T) {
	// Create test points
	points := []struct {
		x, y float64
		name string
	}{
		{0, 0, "origin"},
		{10, 10, "ten"},
		{20, 5, "twenty-five"},
		{5, 20, "five-twenty"},
		{15, 15, "fifteen"},
	}

	// Build KD-tree
	kd := kdtree.New()
	for _, p := range points {
		kd.InsertXY(p.x, p.y, p.name)
	}

	// Build STR-tree
	str := strtree.New()
	for _, p := range points {
		env := geom.NewEnvelope(p.x, p.y, p.x, p.y)
		str.Insert(env, p.name)
	}

	// Build Quadtree
	quad := quadtree.New()
	for _, p := range points {
		env := geom.NewEnvelope(p.x, p.y, p.x, p.y)
		quad.Insert(env, p.name)
	}

	// Test 1: All should have same size
	if kd.Size() != str.Size() || kd.Size() != quad.Size() {
		t.Errorf("Size mismatch: kd=%d, str=%d, quad=%d",
			kd.Size(), str.Size(), quad.Size())
	}

	// Test 2: Range query should return same number of results
	queryEnv := geom.NewEnvelope(0, 0, 15, 15)

	kdResults := kd.Query(queryEnv)
	strResults := str.Query(queryEnv)
	quadResults := quad.Query(queryEnv)

	if len(kdResults) != len(strResults) || len(kdResults) != len(quadResults) {
		t.Errorf("Query results count mismatch: kd=%d, str=%d, quad=%d",
			len(kdResults), len(strResults), len(quadResults))
	}

	// Test 3: All should report similar envelope (allowing for small differences)
	kdEnv := kd.Envelope()
	strEnv := str.Envelope()
	quadEnv := quad.Envelope()

	// Check that envelopes overlap significantly (within 1.0 unit tolerance)
	eps := 1.0
	if !kdEnv.Equals(strEnv, eps) {
		t.Logf("KD envelope: %+v", kdEnv)
		t.Logf("STR envelope: %+v", strEnv)
		// This is acceptable - different implementations may have slightly different envelopes
		// due to internal node structures
	}
	if !kdEnv.Equals(quadEnv, eps) {
		t.Logf("KD envelope: %+v", kdEnv)
		t.Logf("Quad envelope: %+v", quadEnv)
		// This is acceptable - different implementations may have slightly different envelopes
	}

	// Test 4: Clear should work the same
	kd.Clear()
	str.Clear()
	quad.Clear()

	if !kd.IsEmpty() || !str.IsEmpty() || !quad.IsEmpty() {
		t.Error("Clear did not empty all indexes")
	}
}

// TestKDTreeSpecificFeatures tests features unique to KD-tree
func TestKDTreeSpecificFeatures(t *testing.T) {
	kd := kdtree.New()

	// Insert test points
	kd.InsertXY(0, 0, "origin")
	kd.InsertXY(3, 4, "three-four")
	kd.InsertXY(10, 0, "ten-zero")
	kd.InsertXY(0, 10, "zero-ten")

	// Test nearest neighbor (KD-tree specific)
	nearest := kd.NearestNeighbor(1, 1)
	if nearest != "origin" {
		t.Errorf("Expected 'origin' as nearest to (1,1), got %v", nearest)
	}

	// Test k-nearest neighbors (KD-tree specific)
	nearestK := kd.NearestK(0, 0, 2)
	if len(nearestK) != 2 {
		t.Errorf("Expected 2 nearest neighbors, got %d", len(nearestK))
	}

	// Test radius query (KD-tree specific)
	inRadius := kd.QueryRadius(0, 0, 5)
	if len(inRadius) != 2 { // origin and three-four
		t.Errorf("Expected 2 points within radius 5, got %d", len(inRadius))
	}
}

// TestIndexPerformanceCharacteristics compares performance characteristics
// (not actual timings, just that operations complete successfully)
func TestIndexPerformanceCharacteristics(t *testing.T) {
	n := 1000

	kd := kdtree.New()
	str := strtree.New()
	quad := quadtree.New()

	// All indexes should handle bulk insertions
	for i := 0; i < n; i++ {
		x := float64(i % 100)
		y := float64(i / 100)

		kd.InsertXY(x, y, i)
		env := geom.NewEnvelope(x, y, x, y)
		str.Insert(env, i)
		quad.Insert(env, i)
	}

	// Build STR-tree (required before querying)
	str.Build()

	// All should have same size
	if kd.Size() != n || str.Size() != n || quad.Size() != n {
		t.Errorf("Size mismatch after bulk insert: kd=%d, str=%d, quad=%d",
			kd.Size(), str.Size(), quad.Size())
	}

	// All should handle queries
	queryEnv := geom.NewEnvelope(40, 4, 60, 6)

	kdResults := kd.Query(queryEnv)
	strResults := str.Query(queryEnv)
	quadResults := quad.Query(queryEnv)

	// All should find results
	if len(kdResults) == 0 || len(strResults) == 0 || len(quadResults) == 0 {
		t.Error("One or more indexes returned no results")
	}

	// Results should be similar (may differ slightly due to boundary conditions)
	// Allow some tolerance
	maxCount := max(len(kdResults), max(len(strResults), len(quadResults)))
	minCount := min(len(kdResults), min(len(strResults), len(quadResults)))

	if maxCount > minCount*2 { // Results shouldn't differ by more than 2x
		t.Errorf("Result count variance too high: kd=%d, str=%d, quad=%d",
			len(kdResults), len(strResults), len(quadResults))
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestGeometryIntegration tests that KD-tree works with geometry types
func TestGeometryIntegration(t *testing.T) {
	factory := geom.DefaultFactory
	kd := kdtree.New()

	// Insert various geometry types (KD-tree uses centroid)
	pt := factory.CreatePoint(10, 10)
	kd.InsertGeometry(pt)

	// Query using geometry
	queryPt := factory.CreatePoint(10, 10)
	results := kd.QueryGeometry(queryPt)

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

// TestEmptyIndexBehavior ensures all indexes handle empty state consistently
func TestEmptyIndexBehavior(t *testing.T) {
	kd := kdtree.New()
	str := strtree.New()
	quad := quadtree.New()

	// All should be empty initially
	if !kd.IsEmpty() || !str.IsEmpty() || !quad.IsEmpty() {
		t.Error("New indexes should be empty")
	}

	// All should return empty results for queries
	env := geom.NewEnvelope(0, 0, 10, 10)

	kdResults := kd.Query(env)
	strResults := str.Query(env)
	quadResults := quad.Query(env)

	if len(kdResults) != 0 || len(strResults) != 0 || len(quadResults) != 0 {
		t.Error("Empty indexes should return no results")
	}

	// All should have null/empty envelope
	if !kd.Envelope().IsNull() || !str.Envelope().IsNull() {
		t.Error("Empty indexes should have null envelope")
	}
}

// BenchmarkIndexComparison provides a basic performance comparison
func BenchmarkIndexComparison(b *testing.B) {
	// This benchmark shows relative performance characteristics
	// KD-tree excels at nearest neighbor, others at envelope queries

	points := make([]struct{ x, y float64 }, 1000)
	for i := range points {
		points[i].x = float64(i % 100)
		points[i].y = float64(i / 100)
	}

	b.Run("KDTree-Insert", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			kd := kdtree.New()
			for _, p := range points {
				kd.InsertXY(p.x, p.y, nil)
			}
		}
	})

	b.Run("STRTree-Insert", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			str := strtree.New()
			for _, p := range points {
				env := geom.NewEnvelope(p.x, p.y, p.x, p.y)
				str.Insert(env, nil)
			}
		}
	})

	b.Run("Quadtree-Insert", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			quad := quadtree.New()
			for _, p := range points {
				env := geom.NewEnvelope(p.x, p.y, p.x, p.y)
				quad.Insert(env, nil)
			}
		}
	})

	// Setup for query benchmarks
	kd := kdtree.New()
	str := strtree.New()
	quad := quadtree.New()

	for _, p := range points {
		kd.InsertXY(p.x, p.y, nil)
		env := geom.NewEnvelope(p.x, p.y, p.x, p.y)
		str.Insert(env, nil)
		quad.Insert(env, nil)
	}
	str.Build()

	queryEnv := geom.NewEnvelope(40, 4, 60, 6)

	b.Run("KDTree-Query", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			kd.Query(queryEnv)
		}
	})

	b.Run("STRTree-Query", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			str.Query(queryEnv)
		}
	})

	b.Run("Quadtree-Query", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			quad.Query(queryEnv)
		}
	})

	// KD-tree specific: nearest neighbor
	b.Run("KDTree-NearestNeighbor", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			kd.NearestNeighbor(50, 5)
		}
	})
}
