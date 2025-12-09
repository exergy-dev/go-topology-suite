package quadtree

import (
	"testing"

	"github.com/go-topology-suite/gts/geom"
)

func TestNewQuadtree(t *testing.T) {
	qt := New()
	if qt == nil {
		t.Fatal("Expected non-nil quadtree")
	}

	if qt.Size() != 0 {
		t.Errorf("Expected size 0, got %d", qt.Size())
	}

	if !qt.IsEmpty() {
		t.Error("Expected empty quadtree")
	}
}

func TestNewWithBounds(t *testing.T) {
	bounds := geom.NewEnvelope(0, 0, 100, 100)
	qt := NewWithBounds(bounds)

	if qt == nil {
		t.Fatal("Expected non-nil quadtree")
	}

	env := qt.Envelope()
	if env.IsNull() {
		t.Error("Expected non-null envelope")
	}
}

func TestInsertAndQuery(t *testing.T) {
	qt := New()

	// Insert items
	env1 := geom.NewEnvelope(0, 0, 10, 10)
	env2 := geom.NewEnvelope(5, 5, 15, 15)
	env3 := geom.NewEnvelope(20, 20, 30, 30)

	qt.Insert(env1, "item1")
	qt.Insert(env2, "item2")
	qt.Insert(env3, "item3")

	if qt.Size() != 3 {
		t.Errorf("Expected size 3, got %d", qt.Size())
	}

	// Query overlapping region
	queryEnv := geom.NewEnvelope(0, 0, 5, 5)
	results := qt.Query(queryEnv)

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Query non-overlapping region
	queryEnv = geom.NewEnvelope(100, 100, 110, 110)
	results = qt.Query(queryEnv)

	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}

func TestQueryPoint(t *testing.T) {
	qt := New()

	env1 := geom.NewEnvelope(0, 0, 10, 10)
	env2 := geom.NewEnvelope(5, 5, 15, 15)
	env3 := geom.NewEnvelope(20, 20, 30, 30)

	qt.Insert(env1, "item1")
	qt.Insert(env2, "item2")
	qt.Insert(env3, "item3")

	// Point inside env1 only
	results := qt.QueryPoint(2, 2)
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	// Point in overlap region
	results = qt.QueryPoint(7, 7)
	if len(results) != 2 {
		t.Errorf("Expected 2 results for overlap point, got %d", len(results))
	}

	// Point inside env3 only
	results = qt.QueryPoint(25, 25)
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

func TestAutoExpandBounds(t *testing.T) {
	qt := New()

	// Insert first item - sets initial bounds
	qt.Insert(geom.NewEnvelope(0, 0, 10, 10), "item1")

	// Insert item outside bounds - should auto-expand
	qt.Insert(geom.NewEnvelope(100, 100, 110, 110), "item2")

	if qt.Size() != 2 {
		t.Errorf("Expected size 2, got %d", qt.Size())
	}

	// Both items should be queryable
	results := qt.Query(geom.NewEnvelope(5, 5, 105, 105))
	if len(results) != 2 {
		t.Errorf("Expected 2 results after expansion, got %d", len(results))
	}
}

func TestLargeDataset(t *testing.T) {
	qt := NewWithOptions(20, 8)

	// Insert many items
	n := 1000
	for i := 0; i < n; i++ {
		x := float64(i % 100)
		y := float64(i / 10)
		env := geom.NewEnvelope(x, y, x+5, y+5)
		qt.Insert(env, i)
	}

	if qt.Size() != n {
		t.Errorf("Expected size %d, got %d", n, qt.Size())
	}

	// Query specific region
	queryEnv := geom.NewEnvelope(0, 0, 20, 20)
	results := qt.Query(queryEnv)

	if len(results) == 0 {
		t.Error("Expected some results from large dataset query")
	}
}

func TestRemove(t *testing.T) {
	qt := New()

	env1 := geom.NewEnvelope(0, 0, 10, 10)
	env2 := geom.NewEnvelope(20, 20, 30, 30)

	qt.Insert(env1, "item1")
	qt.Insert(env2, "item2")

	if qt.Size() != 2 {
		t.Errorf("Expected size 2, got %d", qt.Size())
	}

	// Remove item1
	removed := qt.Remove(env1, "item1")
	if !removed {
		t.Error("Expected successful removal")
	}

	if qt.Size() != 1 {
		t.Errorf("Expected size 1 after removal, got %d", qt.Size())
	}

	// Try to remove non-existent item
	removed = qt.Remove(env1, "nonexistent")
	if removed {
		t.Error("Expected removal to fail for non-existent item")
	}
}

func TestClear(t *testing.T) {
	qt := New()

	qt.Insert(geom.NewEnvelope(0, 0, 10, 10), "item1")
	qt.Insert(geom.NewEnvelope(20, 20, 30, 30), "item2")

	qt.Clear()

	if !qt.IsEmpty() {
		t.Error("Expected empty quadtree after Clear")
	}

	if qt.Size() != 0 {
		t.Errorf("Expected size 0 after Clear, got %d", qt.Size())
	}
}

func TestQueryAll(t *testing.T) {
	qt := New()

	qt.Insert(geom.NewEnvelope(0, 0, 10, 10), "item1")
	qt.Insert(geom.NewEnvelope(20, 20, 30, 30), "item2")
	qt.Insert(geom.NewEnvelope(40, 40, 50, 50), "item3")

	items := qt.QueryAll()
	if len(items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(items))
	}
}

func TestDepth(t *testing.T) {
	qt := NewWithOptions(10, 2) // Low capacity to force depth

	// Insert enough items to create depth
	for i := 0; i < 100; i++ {
		x := float64(i % 10)
		y := float64(i / 10)
		env := geom.NewEnvelope(x, y, x+0.5, y+0.5)
		qt.Insert(env, i)
	}

	depth := qt.Depth()
	if depth < 2 {
		t.Errorf("Expected depth >= 2 with 100 items, got %d", depth)
	}
}

func TestEnvelope(t *testing.T) {
	qt := New()

	qt.Insert(geom.NewEnvelope(0, 0, 10, 10), "item1")
	qt.Insert(geom.NewEnvelope(20, 20, 30, 30), "item2")

	env := qt.Envelope()

	if env.IsNull() {
		t.Error("Expected non-null envelope")
	}
}

func TestEmptyQuery(t *testing.T) {
	qt := New()

	// Query empty quadtree
	results := qt.Query(geom.NewEnvelope(0, 0, 10, 10))
	if results != nil && len(results) != 0 {
		t.Error("Expected nil or empty results from empty quadtree")
	}

	// Query with nil envelope
	results = qt.Query(nil)
	if results != nil && len(results) != 0 {
		t.Error("Expected nil or empty results for nil envelope query")
	}
}

func TestNullEnvelopeInsert(t *testing.T) {
	qt := New()

	// Insert with nil envelope should be ignored
	qt.Insert(nil, "data")
	if qt.Size() != 0 {
		t.Error("Expected nil envelope insert to be ignored")
	}

	// Insert with empty envelope should be ignored
	qt.Insert(geom.NewEnvelopeEmpty(), "data")
	if qt.Size() != 0 {
		t.Error("Expected empty envelope insert to be ignored")
	}
}

func TestQueryGeometry(t *testing.T) {
	qt := New()

	env1 := geom.NewEnvelope(0, 0, 10, 10)
	env2 := geom.NewEnvelope(20, 20, 30, 30)

	qt.Insert(env1, "item1")
	qt.Insert(env2, "item2")

	// Query using a point geometry
	factory := geom.DefaultFactory
	point := factory.CreatePoint(5, 5)
	results := qt.QueryGeometry(point)

	if len(results) != 1 {
		t.Errorf("Expected 1 result for point query, got %d", len(results))
	}
}

func TestInsertGeometry(t *testing.T) {
	qt := New()

	factory := geom.DefaultFactory
	point := factory.CreatePoint(5, 5)

	qt.InsertGeometry(point)

	if qt.Size() != 1 {
		t.Errorf("Expected size 1, got %d", qt.Size())
	}

	results := qt.QueryPoint(5, 5)
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

func TestVisit(t *testing.T) {
	qt := New()

	qt.Insert(geom.NewEnvelope(0, 0, 10, 10), "item1")
	qt.Insert(geom.NewEnvelope(20, 20, 30, 30), "item2")
	qt.Insert(geom.NewEnvelope(40, 40, 50, 50), "item3")

	count := 0
	qt.Visit(func(env *geom.Envelope, data interface{}) bool {
		count++
		return true // continue
	})

	if count != 3 {
		t.Errorf("Expected 3 items visited, got %d", count)
	}
}

func TestVisitEarlyTermination(t *testing.T) {
	qt := New()

	qt.Insert(geom.NewEnvelope(0, 0, 10, 10), "item1")
	qt.Insert(geom.NewEnvelope(20, 20, 30, 30), "item2")
	qt.Insert(geom.NewEnvelope(40, 40, 50, 50), "item3")

	count := 0
	qt.Visit(func(env *geom.Envelope, data interface{}) bool {
		count++
		return count < 2 // stop after 2
	})

	if count != 2 {
		t.Errorf("Expected 2 items visited before termination, got %d", count)
	}
}

func TestQuadrantAssignment(t *testing.T) {
	// Create quadtree with known bounds
	bounds := geom.NewEnvelope(0, 0, 100, 100)
	qt := NewWithBounds(bounds)

	// Insert items in specific quadrants
	// Center is at (50, 50)
	// NW: x < 50, y >= 50
	// NE: x >= 50, y >= 50
	// SW: x < 50, y < 50
	// SE: x >= 50, y < 50

	qt.Insert(geom.NewEnvelope(10, 60, 20, 70), "NW")   // NW quadrant
	qt.Insert(geom.NewEnvelope(60, 60, 70, 70), "NE")   // NE quadrant
	qt.Insert(geom.NewEnvelope(10, 10, 20, 20), "SW")   // SW quadrant
	qt.Insert(geom.NewEnvelope(60, 10, 70, 20), "SE")   // SE quadrant
	qt.Insert(geom.NewEnvelope(40, 40, 60, 60), "span") // Spans center

	if qt.Size() != 5 {
		t.Errorf("Expected size 5, got %d", qt.Size())
	}

	// Query NW quadrant only
	results := qt.Query(geom.NewEnvelope(0, 50, 50, 100))
	foundNW := false
	for _, r := range results {
		if r == "NW" {
			foundNW = true
			break
		}
	}
	if !foundNW {
		t.Error("Expected to find NW item in NW quadrant query")
	}
}

func BenchmarkInsert(b *testing.B) {
	qt := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := float64(i % 1000)
		y := float64(i / 1000)
		env := geom.NewEnvelope(x, y, x+1, y+1)
		qt.Insert(env, i)
	}
}

func BenchmarkQuery(b *testing.B) {
	qt := New()

	// Insert items
	for i := 0; i < 10000; i++ {
		x := float64(i % 100)
		y := float64(i / 100)
		env := geom.NewEnvelope(x, y, x+1, y+1)
		qt.Insert(env, i)
	}

	queryEnv := geom.NewEnvelope(50, 50, 60, 60)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		qt.Query(queryEnv)
	}
}
