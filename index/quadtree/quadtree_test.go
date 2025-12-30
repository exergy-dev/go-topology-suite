package quadtree

import (
	"testing"

	"github.com/go-topology-suite/gts/geom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewQuadtree(t *testing.T) {
	qt := New()
	require.NotNil(t, qt, "Expected non-nil quadtree")
	assert.Equal(t, 0, qt.Size(), "Expected size 0")
	assert.True(t, qt.IsEmpty(), "Expected empty quadtree")
}

func TestNewWithBounds(t *testing.T) {
	bounds := geom.NewEnvelope(0, 0, 100, 100)
	qt := NewWithBounds(bounds)

	require.NotNil(t, qt, "Expected non-nil quadtree")

	env := qt.Envelope()
	assert.False(t, env.IsNull(), "Expected non-null envelope")
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

	assert.Equal(t, 3, qt.Size(), "Expected size 3")

	// Query overlapping region
	queryEnv := geom.NewEnvelope(0, 0, 5, 5)
	results := qt.Query(queryEnv)
	assert.Equal(t, 2, len(results), "Expected 2 results")

	// Query non-overlapping region
	queryEnv = geom.NewEnvelope(100, 100, 110, 110)
	results = qt.Query(queryEnv)
	assert.Equal(t, 0, len(results), "Expected 0 results")
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
	assert.Equal(t, 1, len(results), "Expected 1 result")

	// Point in overlap region
	results = qt.QueryPoint(7, 7)
	assert.Equal(t, 2, len(results), "Expected 2 results for overlap point")

	// Point inside env3 only
	results = qt.QueryPoint(25, 25)
	assert.Equal(t, 1, len(results), "Expected 1 result")
}

func TestAutoExpandBounds(t *testing.T) {
	qt := New()

	// Insert first item - sets initial bounds
	qt.Insert(geom.NewEnvelope(0, 0, 10, 10), "item1")

	// Insert item outside bounds - should auto-expand
	qt.Insert(geom.NewEnvelope(100, 100, 110, 110), "item2")

	assert.Equal(t, 2, qt.Size(), "Expected size 2")

	// Both items should be queryable
	results := qt.Query(geom.NewEnvelope(5, 5, 105, 105))
	assert.Equal(t, 2, len(results), "Expected 2 results after expansion")
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

	assert.Equal(t, n, qt.Size(), "Expected size %d", n)

	// Query specific region
	queryEnv := geom.NewEnvelope(0, 0, 20, 20)
	results := qt.Query(queryEnv)
	assert.NotEmpty(t, results, "Expected some results from large dataset query")
}

func TestRemove(t *testing.T) {
	qt := New()

	env1 := geom.NewEnvelope(0, 0, 10, 10)
	env2 := geom.NewEnvelope(20, 20, 30, 30)

	qt.Insert(env1, "item1")
	qt.Insert(env2, "item2")

	assert.Equal(t, 2, qt.Size(), "Expected size 2")

	// Remove item1
	removed := qt.Remove(env1, "item1")
	assert.True(t, removed, "Expected successful removal")
	assert.Equal(t, 1, qt.Size(), "Expected size 1 after removal")

	// Try to remove non-existent item
	removed = qt.Remove(env1, "nonexistent")
	assert.False(t, removed, "Expected removal to fail for non-existent item")
}

func TestClear(t *testing.T) {
	qt := New()

	qt.Insert(geom.NewEnvelope(0, 0, 10, 10), "item1")
	qt.Insert(geom.NewEnvelope(20, 20, 30, 30), "item2")

	qt.Clear()

	assert.True(t, qt.IsEmpty(), "Expected empty quadtree after Clear")
	assert.Equal(t, 0, qt.Size(), "Expected size 0 after Clear")
}

func TestQueryAll(t *testing.T) {
	qt := New()

	qt.Insert(geom.NewEnvelope(0, 0, 10, 10), "item1")
	qt.Insert(geom.NewEnvelope(20, 20, 30, 30), "item2")
	qt.Insert(geom.NewEnvelope(40, 40, 50, 50), "item3")

	items := qt.QueryAll()
	assert.Equal(t, 3, len(items), "Expected 3 items")
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
	assert.GreaterOrEqual(t, depth, 2, "Expected depth >= 2 with 100 items")
}

func TestEnvelope(t *testing.T) {
	qt := New()

	qt.Insert(geom.NewEnvelope(0, 0, 10, 10), "item1")
	qt.Insert(geom.NewEnvelope(20, 20, 30, 30), "item2")

	env := qt.Envelope()
	assert.False(t, env.IsNull(), "Expected non-null envelope")
}

func TestEmptyQuery(t *testing.T) {
	qt := New()

	// Query empty quadtree
	results := qt.Query(geom.NewEnvelope(0, 0, 10, 10))
	assert.True(t, results == nil || len(results) == 0, "Expected nil or empty results from empty quadtree")

	// Query with nil envelope
	results = qt.Query(nil)
	assert.True(t, results == nil || len(results) == 0, "Expected nil or empty results for nil envelope query")
}

func TestNullEnvelopeInsert(t *testing.T) {
	qt := New()

	// Insert with nil envelope should be ignored
	qt.Insert(nil, "data")
	assert.Equal(t, 0, qt.Size(), "Expected nil envelope insert to be ignored")

	// Insert with empty envelope should be ignored
	qt.Insert(geom.NewEnvelopeEmpty(), "data")
	assert.Equal(t, 0, qt.Size(), "Expected empty envelope insert to be ignored")
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

	assert.Equal(t, 1, len(results), "Expected 1 result for point query")
}

func TestInsertGeometry(t *testing.T) {
	qt := New()

	factory := geom.DefaultFactory
	point := factory.CreatePoint(5, 5)

	qt.InsertGeometry(point)

	assert.Equal(t, 1, qt.Size(), "Expected size 1")

	results := qt.QueryPoint(5, 5)
	assert.Equal(t, 1, len(results), "Expected 1 result")
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

	assert.Equal(t, 3, count, "Expected 3 items visited")
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

	assert.Equal(t, 2, count, "Expected 2 items visited before termination")
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

	assert.Equal(t, 5, qt.Size(), "Expected size 5")

	// Query NW quadrant only
	results := qt.Query(geom.NewEnvelope(0, 50, 50, 100))
	foundNW := false
	for _, r := range results {
		if r == "NW" {
			foundNW = true
			break
		}
	}
	assert.True(t, foundNW, "Expected to find NW item in NW quadrant query")
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
