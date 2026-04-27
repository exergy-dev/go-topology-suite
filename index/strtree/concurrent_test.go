package strtree_test

import (
	"sync"
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/robert-malhotra/go-topology-suite/index/strtree"
	"github.com/stretchr/testify/require"
)

// TestConcurrentQuery verifies that concurrent queries on a built tree
// do not race. Run with -race to verify.
func TestConcurrentQuery(t *testing.T) {
	tree := strtree.New()
	for i := 0; i < 100; i++ {
		x := float64(i)
		require.NoError(t, tree.Insert(geom.NewEnvelope(x, 0, x+1, 1), i))
	}
	tree.Build()

	var wg sync.WaitGroup
	n := 100
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(x float64) {
			defer wg.Done()
			results := tree.Query(geom.NewEnvelope(x, 0, x+10, 1))
			if len(results) == 0 {
				t.Error("Expected results from query")
			}
		}(float64(i % 90))
	}
	wg.Wait()
}

// TestConcurrentQueryWithLazyBuild verifies no race when Query triggers a lazy build.
func TestConcurrentQueryWithLazyBuild(t *testing.T) {
	tree := strtree.New()
	for i := 0; i < 100; i++ {
		x := float64(i)
		require.NoError(t, tree.Insert(geom.NewEnvelope(x, 0, x+1, 1), i))
	}
	// Don't call Build() — let Query trigger it lazily

	var wg sync.WaitGroup
	n := 50
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(x float64) {
			defer wg.Done()
			results := tree.Query(geom.NewEnvelope(x, 0, x+10, 1))
			if len(results) == 0 {
				t.Error("Expected results from lazy-build query")
			}
		}(float64(i % 90))
	}
	wg.Wait()
}

// TestConcurrentReadMethods verifies concurrent access to various read methods.
func TestConcurrentReadMethods(t *testing.T) {
	tree := strtree.New()
	for i := 0; i < 50; i++ {
		x := float64(i)
		require.NoError(t, tree.Insert(geom.NewEnvelope(x, 0, x+1, 1), i))
	}
	tree.Build()

	var wg sync.WaitGroup
	n := 100
	wg.Add(n * 4)
	for i := 0; i < n; i++ {
		go func() { defer wg.Done(); _ = tree.Size() }()
		go func() { defer wg.Done(); _ = tree.IsEmpty() }()
		go func() { defer wg.Done(); _ = tree.Envelope() }()
		go func() { defer wg.Done(); _ = tree.Items() }()
	}
	wg.Wait()
}
