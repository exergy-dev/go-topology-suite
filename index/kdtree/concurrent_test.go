package kdtree_test

import (
	"sync"
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/robert-malhotra/go-topology-suite/index/kdtree"
)

// TestConcurrentQuery verifies that concurrent queries on a populated tree
// do not race. Run with -race to verify.
func TestConcurrentQuery(t *testing.T) {
	tree := kdtree.New()
	for i := 0; i < 100; i++ {
		tree.InsertXY(float64(i), float64(i), i)
	}

	var wg sync.WaitGroup
	n := 100
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(x float64) {
			defer wg.Done()
			_ = tree.Query(geom.NewEnvelope(x, x, x+10, x+10))
		}(float64(i % 90))
	}
	wg.Wait()
}

// TestConcurrentReadMethods verifies concurrent access to read methods.
func TestConcurrentReadMethods(t *testing.T) {
	tree := kdtree.New()
	for i := 0; i < 50; i++ {
		tree.InsertXY(float64(i), float64(i), i)
	}

	var wg sync.WaitGroup
	n := 100
	wg.Add(n * 4)
	for i := 0; i < n; i++ {
		go func() { defer wg.Done(); _ = tree.Size() }()
		go func() { defer wg.Done(); _ = tree.IsEmpty() }()
		go func() { defer wg.Done(); _ = tree.Envelope() }()
		go func() { defer wg.Done(); _ = tree.NearestNeighbor(25, 25) }()
	}
	wg.Wait()
}
