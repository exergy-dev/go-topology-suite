package bench

import "testing"

// BenchmarkPairwiseIntersection clips PairwiseIntersectCount small polygons
// against the fixed ~50-vertex reference polygon via overlay.Intersection.
// Workload size is scaled 100x down from the 10k reference (see doc.go).
func BenchmarkPairwiseIntersection(b *testing.B) {
	b.ReportAllocs()
	PairwiseIntersectionWorkload(b)
}
