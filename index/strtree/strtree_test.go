package strtree

import (
	"testing"

	"github.com/go-topology-suite/gts/geom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSTRtree(t *testing.T) {
	tree := New()
	require.NotNil(t, tree, "Expected non-nil tree")
	assert.Equal(t, 0, tree.Size(), "Expected size 0")
	assert.True(t, tree.IsEmpty(), "Expected empty tree")
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

	assert.Equal(t, 3, tree.Size(), "Expected size 3")

	// Query overlapping region
	queryEnv := geom.NewEnvelope(0, 0, 5, 5)
	results := tree.Query(queryEnv)
	assert.Equal(t, 2, len(results), "Expected 2 results")

	// Query non-overlapping region
	queryEnv = geom.NewEnvelope(100, 100, 110, 110)
	results = tree.Query(queryEnv)
	assert.Equal(t, 0, len(results), "Expected 0 results")
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
	assert.Equal(t, 1, len(results), "Expected 1 result")

	// Point in overlap region
	results = tree.QueryPoint(7, 7)
	assert.Equal(t, 2, len(results), "Expected 2 results for overlap point")

	// Point inside env3 only
	results = tree.QueryPoint(25, 25)
	assert.Equal(t, 1, len(results), "Expected 1 result")
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

	assert.Equal(t, n, tree.Size(), "Expected size %d", n)

	// Query specific region
	queryEnv := geom.NewEnvelope(0, 0, 100, 100)
	results := tree.Query(queryEnv)

	// Should find items with x,y from 0 to ~95
	assert.NotEmpty(t, results, "Expected some results from large dataset query")
}

func TestNearestNeighbor(t *testing.T) {
	tree := New()

	// Insert points at known locations
	tree.Insert(geom.NewEnvelope(0, 0, 0, 0), "origin")
	tree.Insert(geom.NewEnvelope(10, 10, 10, 10), "ten")
	tree.Insert(geom.NewEnvelope(100, 100, 100, 100), "far")

	// Find nearest to (1, 1)
	nearest := tree.NearestNeighbor(geom.NewEnvelope(1, 1, 1, 1))
	assert.Equal(t, "origin", nearest, "Expected 'origin' as nearest")

	// Find nearest to (15, 15)
	nearest = tree.NearestNeighbor(geom.NewEnvelope(15, 15, 15, 15))
	assert.Equal(t, "ten", nearest, "Expected 'ten' as nearest")

	// Find nearest to (90, 90)
	nearest = tree.NearestNeighbor(geom.NewEnvelope(90, 90, 90, 90))
	assert.Equal(t, "far", nearest, "Expected 'far' as nearest")
}

func TestRemove(t *testing.T) {
	tree := New()

	env1 := geom.NewEnvelope(0, 0, 10, 10)
	env2 := geom.NewEnvelope(20, 20, 30, 30)

	tree.Insert(env1, "item1")
	tree.Insert(env2, "item2")

	assert.Equal(t, 2, tree.Size(), "Expected size 2")

	// Remove item1
	removed := tree.Remove(env1, "item1")
	assert.True(t, removed, "Expected successful removal")
	assert.Equal(t, 1, tree.Size(), "Expected size 1 after removal")

	// Try to remove non-existent item
	removed = tree.Remove(env1, "nonexistent")
	assert.False(t, removed, "Expected removal to fail for non-existent item")
}

func TestClear(t *testing.T) {
	tree := New()

	tree.Insert(geom.NewEnvelope(0, 0, 10, 10), "item1")
	tree.Insert(geom.NewEnvelope(20, 20, 30, 30), "item2")

	tree.Clear()

	assert.True(t, tree.IsEmpty(), "Expected empty tree after Clear")
	assert.Equal(t, 0, tree.Size(), "Expected size 0 after Clear")
}

func TestItems(t *testing.T) {
	tree := New()

	tree.Insert(geom.NewEnvelope(0, 0, 10, 10), "item1")
	tree.Insert(geom.NewEnvelope(20, 20, 30, 30), "item2")
	tree.Insert(geom.NewEnvelope(40, 40, 50, 50), "item3")

	items := tree.Items()
	assert.Equal(t, 3, len(items), "Expected 3 items")
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
	assert.GreaterOrEqual(t, depth, 2, "Expected depth >= 2 with 20 items")
}

func TestEnvelope(t *testing.T) {
	tree := New()

	tree.Insert(geom.NewEnvelope(0, 0, 10, 10), "item1")
	tree.Insert(geom.NewEnvelope(20, 20, 30, 30), "item2")

	env := tree.Envelope()

	assert.Equal(t, float64(0), env.MinX, "Expected min X 0")
	assert.Equal(t, float64(0), env.MinY, "Expected min Y 0")
	assert.Equal(t, float64(30), env.MaxX, "Expected max X 30")
	assert.Equal(t, float64(30), env.MaxY, "Expected max Y 30")
}

func TestEmptyQuery(t *testing.T) {
	tree := New()

	// Query empty tree
	results := tree.Query(geom.NewEnvelope(0, 0, 10, 10))
	assert.True(t, results == nil || len(results) == 0, "Expected nil or empty results from empty tree")

	// Query with nil envelope
	results = tree.Query(nil)
	assert.True(t, results == nil || len(results) == 0, "Expected nil or empty results for nil envelope query")
}

func TestNullEnvelopeInsert(t *testing.T) {
	tree := New()

	// Insert with nil envelope should be ignored
	tree.Insert(nil, "data")
	assert.Equal(t, 0, tree.Size(), "Expected nil envelope insert to be ignored")

	// Insert with empty envelope should be ignored
	tree.Insert(geom.NewEnvelopeEmpty(), "data")
	assert.Equal(t, 0, tree.Size(), "Expected empty envelope insert to be ignored")
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

	assert.Equal(t, 1, len(results), "Expected 1 result for point query")
}

func TestAutoBuilding(t *testing.T) {
	tree := New()

	tree.Insert(geom.NewEnvelope(0, 0, 10, 10), "item1")
	tree.Insert(geom.NewEnvelope(20, 20, 30, 30), "item2")

	// Query should auto-build the tree
	results := tree.Query(geom.NewEnvelope(5, 5, 15, 15))

	assert.Equal(t, 1, len(results), "Expected 1 result")
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
