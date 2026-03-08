package kdtree

import (
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
)

func TestNewKDTree(t *testing.T) {
	tree := New()
	if tree == nil {
		t.Fatal("Expected non-nil tree")
	}

	if tree.Size() != 0 {
		t.Errorf("Expected size 0, got %d", tree.Size())
	}

	if !tree.IsEmpty() {
		t.Error("Expected empty tree")
	}
}

func TestInsertAndQuery(t *testing.T) {
	tree := New()

	// Insert points
	tree.InsertXY(0, 0, "origin")
	tree.InsertXY(10, 10, "ten")
	tree.InsertXY(5, 5, "five")

	if tree.Size() != 3 {
		t.Errorf("Expected size 3, got %d", tree.Size())
	}

	// Query a region containing all points
	env := geom.NewEnvelope(0, 0, 10, 10)
	results := tree.Query(env)

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	// Query a region containing only some points
	env = geom.NewEnvelope(0, 0, 5, 5)
	results = tree.Query(env)

	if len(results) != 2 {
		t.Errorf("Expected 2 results (origin and five), got %d", len(results))
	}

	// Query a region with no points
	env = geom.NewEnvelope(100, 100, 110, 110)
	results = tree.Query(env)

	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}

func TestQueryPoint(t *testing.T) {
	tree := New()

	tree.InsertXY(0, 0, "origin")
	tree.InsertXY(10, 10, "ten")
	tree.InsertXY(5, 5, "five")

	// Query exact point
	results := tree.QueryPoint(5, 5)
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
	if results[0] != "five" {
		t.Errorf("Expected 'five', got %v", results[0])
	}

	// Query non-existent point
	results = tree.QueryPoint(3, 3)
	if len(results) != 0 {
		t.Errorf("Expected 0 results for non-existent point, got %d", len(results))
	}
}

func TestQueryRadius(t *testing.T) {
	tree := New()

	tree.InsertXY(0, 0, "origin")
	tree.InsertXY(10, 0, "right")
	tree.InsertXY(0, 10, "up")
	tree.InsertXY(20, 20, "far")

	// Query with radius that includes origin and neighbors
	results := tree.QueryRadius(0, 0, 15)

	// Should include origin (dist=0), right (dist=10), up (dist=10)
	// Should NOT include far (dist=sqrt(800) ≈ 28.28)
	if len(results) != 3 {
		t.Errorf("Expected 3 results within radius 15, got %d", len(results))
	}

	// Query with small radius
	results = tree.QueryRadius(0, 0, 5)
	if len(results) != 1 {
		t.Errorf("Expected 1 result within radius 5, got %d", len(results))
	}

	// Query with large radius
	results = tree.QueryRadius(0, 0, 50)
	if len(results) != 4 {
		t.Errorf("Expected 4 results within radius 50, got %d", len(results))
	}
}

func TestNearestNeighbor(t *testing.T) {
	tree := New()

	tree.InsertXY(0, 0, "origin")
	tree.InsertXY(10, 10, "ten")
	tree.InsertXY(100, 100, "far")

	// Find nearest to (1, 1) - should be origin
	nearest := tree.NearestNeighbor(1, 1)
	if nearest != "origin" {
		t.Errorf("Expected 'origin' as nearest to (1,1), got %v", nearest)
	}

	// Find nearest to (12, 12) - should be ten
	nearest = tree.NearestNeighbor(12, 12)
	if nearest != "ten" {
		t.Errorf("Expected 'ten' as nearest to (12,12), got %v", nearest)
	}

	// Find nearest to (90, 90) - should be far
	nearest = tree.NearestNeighbor(90, 90)
	if nearest != "far" {
		t.Errorf("Expected 'far' as nearest to (90,90), got %v", nearest)
	}
}

func TestNearestNeighborCoord(t *testing.T) {
	tree := New()

	tree.InsertXY(0, 0, "origin")
	tree.InsertXY(10, 10, "ten")

	query := geom.NewCoordinate(5, 5)
	nearest := tree.NearestNeighborCoord(query)

	// Both points are equidistant (sqrt(50)), but we should get one
	if nearest == nil {
		t.Error("Expected non-nil nearest neighbor")
	}
}

func TestNearestK(t *testing.T) {
	tree := New()

	// Insert points in a grid
	tree.InsertXY(0, 0, "p0")
	tree.InsertXY(10, 0, "p1")
	tree.InsertXY(0, 10, "p2")
	tree.InsertXY(10, 10, "p3")
	tree.InsertXY(5, 5, "center")

	// Find 3 nearest to center
	results := tree.NearestK(5, 5, 3)
	if len(results) != 3 {
		t.Errorf("Expected 3 nearest neighbors, got %d", len(results))
	}

	// First should be center itself
	if results[0] != "center" {
		t.Errorf("Expected 'center' as first result, got %v", results[0])
	}

	// Find all 5 neighbors
	results = tree.NearestK(5, 5, 5)
	if len(results) != 5 {
		t.Errorf("Expected 5 nearest neighbors, got %d", len(results))
	}

	// Request more than available
	results = tree.NearestK(5, 5, 10)
	if len(results) != 5 {
		t.Errorf("Expected 5 results (all available), got %d", len(results))
	}
}

func TestNearestKOrdering(t *testing.T) {
	tree := New()

	// Insert points at known distances from origin
	tree.InsertXY(1, 0, "dist1")   // distance 1
	tree.InsertXY(2, 0, "dist2")   // distance 2
	tree.InsertXY(3, 0, "dist3")   // distance 3
	tree.InsertXY(0, 4, "dist4")   // distance 4
	tree.InsertXY(0, 5, "dist5")   // distance 5

	results := tree.NearestK(0, 0, 3)
	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	// Results should be ordered by distance
	expected := []string{"dist1", "dist2", "dist3"}
	for i, exp := range expected {
		if results[i] != exp {
			t.Errorf("Expected results[%d] = %s, got %v", i, exp, results[i])
		}
	}
}

func TestInsertGeometry(t *testing.T) {
	tree := New()
	factory := geom.DefaultFactory

	// Create geometries
	pt1 := factory.CreatePoint(5, 5)
	pt2 := factory.CreatePoint(10, 10)

	tree.InsertGeometry(pt1)
	tree.InsertGeometry(pt2)

	if tree.Size() != 2 {
		t.Errorf("Expected size 2, got %d", tree.Size())
	}

	// Query should find the points
	results := tree.Query(geom.NewEnvelope(0, 0, 10, 10))
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

func TestQueryGeometry(t *testing.T) {
	tree := New()
	factory := geom.DefaultFactory

	tree.InsertXY(5, 5, "point1")
	tree.InsertXY(15, 15, "point2")

	// Create a query geometry
	queryPt := factory.CreatePoint(5, 5)
	results := tree.QueryGeometry(queryPt)

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

func TestLargeDataset(t *testing.T) {
	tree := New()

	// Insert many points in a grid
	n := 100
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			tree.InsertXY(float64(i), float64(j), i*n+j)
		}
	}

	if tree.Size() != n*n {
		t.Errorf("Expected size %d, got %d", n*n, tree.Size())
	}

	// Query a region
	env := geom.NewEnvelope(10, 10, 20, 20)
	results := tree.Query(env)

	// Should find 11x11 = 121 points (inclusive boundaries)
	if len(results) != 121 {
		t.Errorf("Expected 121 results, got %d", len(results))
	}

	// Test nearest neighbor performance
	nearest := tree.NearestNeighbor(50, 50)
	if nearest == nil {
		t.Error("Expected to find nearest neighbor in large dataset")
	}
}

func TestDepth(t *testing.T) {
	tree := New()

	// Insert points to create some depth
	for i := 0; i < 10; i++ {
		tree.InsertXY(float64(i), float64(i), i)
	}

	depth := tree.Depth()
	if depth < 2 {
		t.Errorf("Expected depth >= 2 with 10 items, got %d", depth)
	}
	if depth > 10 {
		t.Errorf("Expected depth <= 10 with 10 items, got %d", depth)
	}
}

func TestEnvelope(t *testing.T) {
	tree := New()

	tree.InsertXY(0, 0, "p1")
	tree.InsertXY(10, 5, "p2")
	tree.InsertXY(5, 10, "p3")

	env := tree.Envelope()

	if env.MinX != 0 || env.MinY != 0 {
		t.Errorf("Expected min (0, 0), got (%v, %v)", env.MinX, env.MinY)
	}

	if env.MaxX != 10 || env.MaxY != 10 {
		t.Errorf("Expected max (10, 10), got (%v, %v)", env.MaxX, env.MaxY)
	}
}

func TestItems(t *testing.T) {
	tree := New()

	tree.InsertXY(0, 0, "item1")
	tree.InsertXY(10, 10, "item2")
	tree.InsertXY(5, 5, "item3")

	items := tree.Items()
	if len(items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(items))
	}
}

func TestVisit(t *testing.T) {
	tree := New()

	tree.InsertXY(0, 0, "p1")
	tree.InsertXY(10, 10, "p2")
	tree.InsertXY(5, 5, "p3")

	count := 0
	tree.Visit(func(coord geom.Coordinate, data interface{}) bool {
		count++
		return true
	})

	if count != 3 {
		t.Errorf("Expected to visit 3 items, visited %d", count)
	}

	// Test early termination
	count = 0
	tree.Visit(func(coord geom.Coordinate, data interface{}) bool {
		count++
		return count < 2 // Stop after 2 visits
	})

	if count != 2 {
		t.Errorf("Expected to visit 2 items before stopping, visited %d", count)
	}
}

func TestRemove(t *testing.T) {
	tree := New()

	coord1 := geom.NewCoordinate(0, 0)
	coord2 := geom.NewCoordinate(10, 10)

	tree.Insert(coord1, "item1")
	tree.Insert(coord2, "item2")

	if tree.Size() != 2 {
		t.Errorf("Expected size 2, got %d", tree.Size())
	}

	// Remove item1
	removed := tree.Remove(coord1, "item1")
	if !removed {
		t.Error("Expected successful removal")
	}

	if tree.Size() != 1 {
		t.Errorf("Expected size 1 after removal, got %d", tree.Size())
	}

	// Verify item1 is gone
	results := tree.QueryPoint(0, 0)
	if len(results) != 0 {
		t.Error("Expected removed item to not be found")
	}

	// Verify item2 is still there
	results = tree.QueryPoint(10, 10)
	if len(results) != 1 {
		t.Errorf("Expected remaining item to be found, got %d results", len(results))
	}

	// Try to remove non-existent item
	removed = tree.Remove(coord1, "nonexistent")
	if removed {
		t.Error("Expected removal to fail for non-existent item")
	}
}

func TestClear(t *testing.T) {
	tree := New()

	tree.InsertXY(0, 0, "p1")
	tree.InsertXY(10, 10, "p2")
	tree.InsertXY(5, 5, "p3")

	tree.Clear()

	if !tree.IsEmpty() {
		t.Error("Expected empty tree after Clear")
	}

	if tree.Size() != 0 {
		t.Errorf("Expected size 0 after Clear, got %d", tree.Size())
	}

	// Verify envelope is empty
	env := tree.Envelope()
	if !env.IsNull() {
		t.Error("Expected null envelope after Clear")
	}
}

func TestEmptyTree(t *testing.T) {
	tree := New()

	// Query empty tree
	results := tree.Query(geom.NewEnvelope(0, 0, 10, 10))
	if len(results) != 0 {
		t.Error("Expected nil or empty results from empty tree")
	}

	// NearestNeighbor on empty tree
	nearest := tree.NearestNeighbor(0, 0)
	if nearest != nil {
		t.Error("Expected nil nearest neighbor from empty tree")
	}

	// NearestK on empty tree
	results = tree.NearestK(0, 0, 5)
	if len(results) != 0 {
		t.Error("Expected nil or empty results from NearestK on empty tree")
	}

	// Items on empty tree
	items := tree.Items()
	if len(items) != 0 {
		t.Error("Expected nil or empty items from empty tree")
	}
}

func TestNaNCoordinate(t *testing.T) {
	tree := New()

	// Insert with NaN coordinate should be ignored
	nanCoord := geom.NewCoordinateNaN()
	tree.Insert(nanCoord, "data")

	if tree.Size() != 0 {
		t.Error("Expected NaN coordinate insert to be ignored")
	}
}

func TestNullEnvelopeQuery(t *testing.T) {
	tree := New()
	tree.InsertXY(5, 5, "point")

	// Query with nil envelope
	results := tree.Query(nil)
	if len(results) != 0 {
		t.Error("Expected nil or empty results for nil envelope query")
	}

	// Query with empty envelope
	results = tree.Query(geom.NewEnvelopeEmpty())
	if len(results) != 0 {
		t.Error("Expected nil or empty results for empty envelope query")
	}
}

func TestDuplicatePoints(t *testing.T) {
	tree := New()

	// Insert duplicate points
	tree.InsertXY(5, 5, "first")
	tree.InsertXY(5, 5, "second")
	tree.InsertXY(5, 5, "third")

	if tree.Size() != 3 {
		t.Errorf("Expected size 3 with duplicates, got %d", tree.Size())
	}

	// All should be found in query
	results := tree.QueryPoint(5, 5)
	if len(results) != 3 {
		t.Errorf("Expected 3 results for duplicate points, got %d", len(results))
	}
}

func TestAxisSplitting(t *testing.T) {
	tree := New()

	// Insert points to verify correct axis splitting
	// Root should split on X (axis 0)
	tree.InsertXY(5, 5, "root")
	// Left child should split on Y (axis 1)
	tree.InsertXY(3, 3, "left")
	// Right child should split on Y (axis 1)
	tree.InsertXY(7, 7, "right")

	// Verify tree structure by checking queries
	results := tree.Query(geom.NewEnvelope(0, 0, 4, 10))
	if len(results) != 1 {
		t.Errorf("Expected 1 result in left region, got %d", len(results))
	}

	results = tree.Query(geom.NewEnvelope(6, 0, 10, 10))
	if len(results) != 1 {
		t.Errorf("Expected 1 result in right region, got %d", len(results))
	}
}

func TestNearestNeighborSamePoint(t *testing.T) {
	tree := New()

	tree.InsertXY(5, 5, "point")

	// Query the exact same point
	nearest := tree.NearestNeighbor(5, 5)
	if nearest != "point" {
		t.Errorf("Expected 'point' as nearest, got %v", nearest)
	}
}

func TestNearestKWithZeroK(t *testing.T) {
	tree := New()
	tree.InsertXY(5, 5, "point")

	results := tree.NearestK(0, 0, 0)
	if len(results) != 0 {
		t.Error("Expected nil or empty results for k=0")
	}
}

func TestNearestKWithNegativeK(t *testing.T) {
	tree := New()
	tree.InsertXY(5, 5, "point")

	results := tree.NearestK(0, 0, -1)
	if results != nil && len(results) != 0 {
		t.Error("Expected nil or empty results for negative k")
	}
}

func TestBuild(t *testing.T) {
	tree := New()
	tree.InsertXY(5, 5, "point")

	// Build should be a no-op but not panic
	tree.Build()

	// Tree should still work
	if tree.Size() != 1 {
		t.Error("Expected tree to work after Build()")
	}
}

// Benchmark tests

func BenchmarkInsert(b *testing.B) {
	tree := New()
	coord := geom.NewCoordinate(0, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.Insert(coord, i)
	}
}

func BenchmarkQuerySmallRegion(b *testing.B) {
	tree := New()

	// Insert 10000 random-ish points
	for i := 0; i < 10000; i++ {
		x := float64(i % 100)
		y := float64(i / 100)
		tree.InsertXY(x, y, i)
	}

	queryEnv := geom.NewEnvelope(40, 40, 60, 60)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.Query(queryEnv)
	}
}

func BenchmarkNearestNeighbor(b *testing.B) {
	tree := New()

	// Insert 10000 points in a grid
	for i := 0; i < 100; i++ {
		for j := 0; j < 100; j++ {
			tree.InsertXY(float64(i), float64(j), i*100+j)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.NearestNeighbor(50.5, 50.5)
	}
}

func BenchmarkNearestK(b *testing.B) {
	tree := New()

	// Insert 10000 points
	for i := 0; i < 100; i++ {
		for j := 0; j < 100; j++ {
			tree.InsertXY(float64(i), float64(j), i*100+j)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.NearestK(50.5, 50.5, 10)
	}
}

func BenchmarkQueryRadius(b *testing.B) {
	tree := New()

	// Insert points
	for i := 0; i < 1000; i++ {
		x := float64(i % 100)
		y := float64(i / 100)
		tree.InsertXY(x, y, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.QueryRadius(50, 50, 10)
	}
}

func TestQueryRadiusCircularFilter(t *testing.T) {
	tree := New()

	// Insert points at exact positions
	tree.InsertXY(0, 0, "origin")
	tree.InsertXY(3, 4, "at5")     // Distance = 5
	tree.InsertXY(6, 0, "at6")     // Distance = 6
	tree.InsertXY(0, 10, "at10")   // Distance = 10

	// Query with radius 5 - should include origin and at5 only
	results := tree.QueryRadius(0, 0, 5)

	found := make(map[string]bool)
	for _, r := range results {
		if s, ok := r.(string); ok {
			found[s] = true
		}
	}

	if !found["origin"] {
		t.Error("Expected to find 'origin'")
	}
	if !found["at5"] {
		t.Error("Expected to find 'at5' (distance 5)")
	}
	if found["at6"] {
		t.Error("Should not find 'at6' (distance 6 > radius 5)")
	}
	if found["at10"] {
		t.Error("Should not find 'at10' (distance 10 > radius 5)")
	}
}

func TestNearestNeighborTieBreaking(t *testing.T) {
	tree := New()

	// Insert two points equidistant from query point
	tree.InsertXY(0, 5, "north")
	tree.InsertXY(5, 0, "east")

	// Query from origin (0,0) - both are distance 5 away
	nearest := tree.NearestNeighbor(0, 0)

	// Should get one of them (implementation dependent which one)
	if nearest != "north" && nearest != "east" {
		t.Errorf("Expected 'north' or 'east', got %v", nearest)
	}
}

func TestNearestKAllPoints(t *testing.T) {
	tree := New()

	// Insert 5 points
	for i := 0; i < 5; i++ {
		tree.InsertXY(float64(i), 0, i)
	}

	// Request more neighbors than exist
	results := tree.NearestK(0, 0, 100)

	// Should return all 5 points
	if len(results) != 5 {
		t.Errorf("Expected 5 results (all points), got %d", len(results))
	}
}

func TestEnvelopeEmptyTree(t *testing.T) {
	tree := New()

	env := tree.Envelope()
	if !env.IsNull() {
		t.Error("Expected null envelope for empty tree")
	}
}

func TestQueryOptimization(t *testing.T) {
	tree := New()

	// Insert points along X axis
	for i := 0; i < 100; i++ {
		tree.InsertXY(float64(i), 0, i)
	}

	// Query should efficiently prune branches
	// Query far right - should not search left subtrees
	env := geom.NewEnvelope(90, -1, 100, 1)
	results := tree.Query(env)

	// Should find only points 90-99
	if len(results) != 10 {
		t.Errorf("Expected 10 results, got %d", len(results))
	}
}

func TestInsertWithZCoordinate(t *testing.T) {
	tree := New()

	// KD-tree only indexes X,Y but should handle Z coordinates
	coord := geom.NewCoordinateZ(5, 5, 10)
	tree.Insert(coord, "point3d")

	if tree.Size() != 1 {
		t.Error("Expected to insert coordinate with Z value")
	}

	// Should be queryable by X,Y
	results := tree.QueryPoint(5, 5)
	if len(results) != 1 {
		t.Error("Expected to find point by X,Y coordinates")
	}
}

func TestNearestNeighborPruning(t *testing.T) {
	tree := New()

	// Create a scenario where pruning is important
	// Insert points clustered on the left
	for i := 0; i < 50; i++ {
		tree.InsertXY(float64(i), 0, i)
	}
	// And one point far on the right
	tree.InsertXY(1000, 0, "far")

	// Query from left side - should not need to search right subtree
	nearest := tree.NearestNeighbor(0, 0)
	if nearest != 0 {
		t.Errorf("Expected point 0 as nearest, got %v", nearest)
	}
}

func TestQueryBoundaryPoints(t *testing.T) {
	tree := New()

	tree.InsertXY(0, 0, "corner")
	tree.InsertXY(10, 0, "edge")
	tree.InsertXY(10, 10, "opposite")

	// Query with envelope boundaries exactly on points
	env := geom.NewEnvelope(0, 0, 10, 10)
	results := tree.Query(env)

	// All three points should be included (boundary is inclusive)
	if len(results) != 3 {
		t.Errorf("Expected 3 results including boundary points, got %d", len(results))
	}
}

func TestLargeNearestK(t *testing.T) {
	tree := New()

	// Insert 1000 points
	for i := 0; i < 1000; i++ {
		x := float64(i % 100)
		y := float64(i / 100)
		tree.InsertXY(x, y, i)
	}

	// Find 100 nearest neighbors
	results := tree.NearestK(50, 5, 100)

	if len(results) != 100 {
		t.Errorf("Expected 100 nearest neighbors, got %d", len(results))
	}

	// Verify ordering - distances should be non-decreasing
	query := geom.NewCoordinate(50, 5)
	prevDist := 0.0
	for i, item := range results {
		// Extract coordinate from data (this is implementation-dependent)
		// For this test, we know items are integers, so we reconstruct coordinates
		if idx, ok := item.(int); ok {
			x := float64(idx % 100)
			y := float64(idx / 100)
			coord := geom.NewCoordinate(x, y)
			dist := query.Distance(coord)

			if i > 0 && dist < prevDist {
				t.Errorf("Results not ordered by distance: result[%d] dist=%f < result[%d] dist=%f",
					i, dist, i-1, prevDist)
			}
			prevDist = dist
		}
	}
}

func TestNearestNeighborSinglePoint(t *testing.T) {
	tree := New()
	tree.InsertXY(5, 5, "only")

	nearest := tree.NearestNeighbor(100, 100)
	if nearest != "only" {
		t.Error("Expected to find the only point in tree")
	}
}

func TestQueryRadiusWithCoordinateData(t *testing.T) {
	tree := New()

	// Insert coordinates as data
	c1 := geom.NewCoordinate(0, 0)
	c2 := geom.NewCoordinate(3, 0)
	c3 := geom.NewCoordinate(10, 0)

	tree.Insert(c1, c1)
	tree.Insert(c2, c2)
	tree.Insert(c3, c3)

	// Query with radius that should include c1 and c2 but not c3
	results := tree.QueryRadius(0, 0, 5)

	if len(results) != 2 {
		t.Errorf("Expected 2 results within radius, got %d", len(results))
	}
}

func TestDeepTree(t *testing.T) {
	tree := New()

	// Insert points in a way that creates a deep tree (sorted order)
	for i := 0; i < 100; i++ {
		tree.InsertXY(float64(i), float64(i), i)
	}

	depth := tree.Depth()

	// Depth should be reasonable (log2(100) ≈ 6-7, but could be deeper due to sorted insertion)
	if depth < 6 {
		t.Errorf("Expected depth >= 6, got %d", depth)
	}

	// Verify tree still works correctly
	nearest := tree.NearestNeighbor(50, 50)
	if nearest == nil {
		t.Error("Expected to find nearest neighbor in deep tree")
	}
}

func TestBalancedVsUnbalanced(t *testing.T) {
	balanced := New()
	unbalanced := New()

	// Insert in random-ish order for balanced tree
	points := []struct{ x, y float64 }{
		{50, 50}, {25, 25}, {75, 75}, {12, 12}, {37, 37}, {62, 62}, {87, 87},
	}
	for i, p := range points {
		balanced.InsertXY(p.x, p.y, i)
	}

	// Insert in sorted order for unbalanced tree
	for i := 0; i < len(points); i++ {
		unbalanced.InsertXY(float64(i*10), float64(i*10), i)
	}

	// Both trees should have the same number of items
	if balanced.Size() != unbalanced.Size() {
		t.Error("Trees should have same size")
	}

	// Both should correctly answer queries
	balancedResults := balanced.Query(geom.NewEnvelope(0, 0, 100, 100))
	unbalancedResults := unbalanced.Query(geom.NewEnvelope(0, 0, 100, 100))

	if len(balancedResults) == 0 || len(unbalancedResults) == 0 {
		t.Error("Both trees should return results")
	}

	// Balanced tree will typically have smaller depth
	balDepth := balanced.Depth()
	unbalDepth := unbalanced.Depth()

	// Just verify both have reasonable depths
	if balDepth == 0 || unbalDepth == 0 {
		t.Error("Both trees should have non-zero depth")
	}
}
