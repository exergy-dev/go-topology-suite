package strtree

import (
	"testing"

	"github.com/go-topology-suite/gts/geom"
)

func TestNewSTRtree(t *testing.T) {
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

	// Insert items
	env1 := geom.NewEnvelope(0, 0, 10, 10)
	env2 := geom.NewEnvelope(5, 5, 15, 15)
	env3 := geom.NewEnvelope(20, 20, 30, 30)

	tree.Insert(env1, "item1")
	tree.Insert(env2, "item2")
	tree.Insert(env3, "item3")

	if tree.Size() != 3 {
		t.Errorf("Expected size 3, got %d", tree.Size())
	}

	// Query overlapping region
	queryEnv := geom.NewEnvelope(0, 0, 5, 5)
	results := tree.Query(queryEnv)

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Query non-overlapping region
	queryEnv = geom.NewEnvelope(100, 100, 110, 110)
	results = tree.Query(queryEnv)

	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}

func TestQueryPoint(t *testing.T) {
	tree := New()

	env1 := geom.NewEnvelope(0, 0, 10, 10)
	env2 := geom.NewEnvelope(5, 5, 15, 15)
	env3 := geom.NewEnvelope(20, 20, 30, 30)

	tree.Insert(env1, "item1")
	tree.Insert(env2, "item2")
	tree.Insert(env3, "item3")

	// Point inside env1 only
	results := tree.QueryPoint(2, 2)
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	// Point in overlap region
	results = tree.QueryPoint(7, 7)
	if len(results) != 2 {
		t.Errorf("Expected 2 results for overlap point, got %d", len(results))
	}

	// Point inside env3 only
	results = tree.QueryPoint(25, 25)
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

func TestLargeDataset(t *testing.T) {
	tree := NewWithCapacity(10)

	// Insert many items
	n := 1000
	for i := 0; i < n; i++ {
		x := float64(i * 10)
		y := float64(i * 10)
		env := geom.NewEnvelope(x, y, x+5, y+5)
		tree.Insert(env, i)
	}

	if tree.Size() != n {
		t.Errorf("Expected size %d, got %d", n, tree.Size())
	}

	// Query specific region
	queryEnv := geom.NewEnvelope(0, 0, 100, 100)
	results := tree.Query(queryEnv)

	// Should find items with x,y from 0 to ~95
	if len(results) == 0 {
		t.Error("Expected some results from large dataset query")
	}
}

func TestNearestNeighbor(t *testing.T) {
	tree := New()

	// Insert points at known locations
	tree.Insert(geom.NewEnvelope(0, 0, 0, 0), "origin")
	tree.Insert(geom.NewEnvelope(10, 10, 10, 10), "ten")
	tree.Insert(geom.NewEnvelope(100, 100, 100, 100), "far")

	// Find nearest to (1, 1)
	nearest := tree.NearestNeighbor(geom.NewEnvelope(1, 1, 1, 1))
	if nearest != "origin" {
		t.Errorf("Expected 'origin' as nearest, got %v", nearest)
	}

	// Find nearest to (15, 15)
	nearest = tree.NearestNeighbor(geom.NewEnvelope(15, 15, 15, 15))
	if nearest != "ten" {
		t.Errorf("Expected 'ten' as nearest, got %v", nearest)
	}

	// Find nearest to (90, 90)
	nearest = tree.NearestNeighbor(geom.NewEnvelope(90, 90, 90, 90))
	if nearest != "far" {
		t.Errorf("Expected 'far' as nearest, got %v", nearest)
	}
}

func TestRemove(t *testing.T) {
	tree := New()

	env1 := geom.NewEnvelope(0, 0, 10, 10)
	env2 := geom.NewEnvelope(20, 20, 30, 30)

	tree.Insert(env1, "item1")
	tree.Insert(env2, "item2")

	if tree.Size() != 2 {
		t.Errorf("Expected size 2, got %d", tree.Size())
	}

	// Remove item1
	removed := tree.Remove(env1, "item1")
	if !removed {
		t.Error("Expected successful removal")
	}

	if tree.Size() != 1 {
		t.Errorf("Expected size 1 after removal, got %d", tree.Size())
	}

	// Try to remove non-existent item
	removed = tree.Remove(env1, "nonexistent")
	if removed {
		t.Error("Expected removal to fail for non-existent item")
	}
}

func TestClear(t *testing.T) {
	tree := New()

	tree.Insert(geom.NewEnvelope(0, 0, 10, 10), "item1")
	tree.Insert(geom.NewEnvelope(20, 20, 30, 30), "item2")

	tree.Clear()

	if !tree.IsEmpty() {
		t.Error("Expected empty tree after Clear")
	}

	if tree.Size() != 0 {
		t.Errorf("Expected size 0 after Clear, got %d", tree.Size())
	}
}

func TestItems(t *testing.T) {
	tree := New()

	tree.Insert(geom.NewEnvelope(0, 0, 10, 10), "item1")
	tree.Insert(geom.NewEnvelope(20, 20, 30, 30), "item2")
	tree.Insert(geom.NewEnvelope(40, 40, 50, 50), "item3")

	items := tree.Items()
	if len(items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(items))
	}
}

func TestDepth(t *testing.T) {
	tree := NewWithCapacity(2)

	// Insert enough items to create depth
	for i := 0; i < 20; i++ {
		x := float64(i * 10)
		env := geom.NewEnvelope(x, 0, x+5, 5)
		tree.Insert(env, i)
	}

	depth := tree.Depth()
	if depth < 2 {
		t.Errorf("Expected depth >= 2 with 20 items, got %d", depth)
	}
}

func TestEnvelope(t *testing.T) {
	tree := New()

	tree.Insert(geom.NewEnvelope(0, 0, 10, 10), "item1")
	tree.Insert(geom.NewEnvelope(20, 20, 30, 30), "item2")

	env := tree.Envelope()

	if env.MinX != 0 || env.MinY != 0 {
		t.Errorf("Expected min (0, 0), got (%v, %v)", env.MinX, env.MinY)
	}

	if env.MaxX != 30 || env.MaxY != 30 {
		t.Errorf("Expected max (30, 30), got (%v, %v)", env.MaxX, env.MaxY)
	}
}

func TestEmptyQuery(t *testing.T) {
	tree := New()

	// Query empty tree
	results := tree.Query(geom.NewEnvelope(0, 0, 10, 10))
	if results != nil && len(results) != 0 {
		t.Error("Expected nil or empty results from empty tree")
	}

	// Query with nil envelope
	results = tree.Query(nil)
	if results != nil && len(results) != 0 {
		t.Error("Expected nil or empty results for nil envelope query")
	}
}

func TestNullEnvelopeInsert(t *testing.T) {
	tree := New()

	// Insert with nil envelope should be ignored
	tree.Insert(nil, "data")
	if tree.Size() != 0 {
		t.Error("Expected nil envelope insert to be ignored")
	}

	// Insert with empty envelope should be ignored
	tree.Insert(geom.NewEnvelopeEmpty(), "data")
	if tree.Size() != 0 {
		t.Error("Expected empty envelope insert to be ignored")
	}
}

func TestQueryGeometry(t *testing.T) {
	tree := New()

	env1 := geom.NewEnvelope(0, 0, 10, 10)
	env2 := geom.NewEnvelope(20, 20, 30, 30)

	tree.Insert(env1, "item1")
	tree.Insert(env2, "item2")

	// Query using a point geometry
	factory := geom.DefaultFactory
	point := factory.CreatePoint(5, 5)
	results := tree.QueryGeometry(point)

	if len(results) != 1 {
		t.Errorf("Expected 1 result for point query, got %d", len(results))
	}
}

func TestAutoBuilding(t *testing.T) {
	tree := New()

	tree.Insert(geom.NewEnvelope(0, 0, 10, 10), "item1")
	tree.Insert(geom.NewEnvelope(20, 20, 30, 30), "item2")

	// Query should auto-build the tree
	results := tree.Query(geom.NewEnvelope(5, 5, 15, 15))

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

func BenchmarkInsert(b *testing.B) {
	tree := New()
	env := geom.NewEnvelope(0, 0, 10, 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.Insert(env, i)
	}
}

func BenchmarkQuery(b *testing.B) {
	tree := New()

	// Insert items
	for i := 0; i < 10000; i++ {
		x := float64(i % 100)
		y := float64(i / 100)
		env := geom.NewEnvelope(x, y, x+1, y+1)
		tree.Insert(env, i)
	}
	tree.Build()

	queryEnv := geom.NewEnvelope(50, 50, 60, 60)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.Query(queryEnv)
	}
}
