package precision

import (
	"testing"
)

// Benchmarks for evaluating a small-N fast path in MinimumClearance.
// Cases at N=4,16,64,256,1024 vertices on a circular polygon, comparing
// the tree-backed implementation against the brute-force reference.

func benchTreeAt(b *testing.B, n int) {
	poly := largeBenchPolygon(n)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MinimumClearance(poly)
	}
}

func benchSimpleAt(b *testing.B, n int) {
	poly := largeBenchPolygon(n)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		smc := NewSimpleMinimumClearance(poly)
		smc.Distance()
	}
}

func BenchmarkMinimumClearance_Tree_N4(b *testing.B)    { benchTreeAt(b, 4) }
func BenchmarkMinimumClearance_Simple_N4(b *testing.B)  { benchSimpleAt(b, 4) }
func BenchmarkMinimumClearance_Tree_N16(b *testing.B)   { benchTreeAt(b, 16) }
func BenchmarkMinimumClearance_Simple_N16(b *testing.B) { benchSimpleAt(b, 16) }
func BenchmarkMinimumClearance_Tree_N64(b *testing.B)   { benchTreeAt(b, 64) }
func BenchmarkMinimumClearance_Simple_N64(b *testing.B) { benchSimpleAt(b, 64) }
func BenchmarkMinimumClearance_Tree_N256(b *testing.B)  { benchTreeAt(b, 256) }
func BenchmarkMinimumClearance_Simple_N256(b *testing.B) {
	benchSimpleAt(b, 256)
}
