package simplify

import (
	"math"
	"math/rand"
	"testing"

	"github.com/exergy-dev/go-topology-suite/geom"
)

// BenchmarkPolygonHull_ConcaveStar_N1000 builds a 1000-vertex star ring
// (alternating long/short radii) and runs an outer hull at fraction 0.1.
// This exercises the corner-removal loop heavily, where the per-ring
// VertexSequencePackedRtree replaces a linear scan of the ring's
// vertices on every triangle test.
func BenchmarkPolygonHull_ConcaveStar_N1000(b *testing.B) {
	poly := makeConcaveStar(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = PolygonHull(poly, true, 0.1)
	}
}

// BenchmarkPolygonHull_ConcaveStar_N4000 — the same workload at four
// times the vertex count to expose the asymptotic difference.
func BenchmarkPolygonHull_ConcaveStar_N4000(b *testing.B) {
	poly := makeConcaveStar(4000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = PolygonHull(poly, true, 0.1)
	}
}

func makeConcaveStar(n int) *geom.Polygon {
	rng := rand.New(rand.NewSource(7))
	ring := make([]geom.XY, 0, n+1)
	for i := 0; i < n; i++ {
		theta := 2 * math.Pi * float64(i) / float64(n)
		var r float64
		if i%2 == 0 {
			r = 100 + 5*rng.Float64()
		} else {
			r = 60 + 5*rng.Float64()
		}
		ring = append(ring, geom.XY{X: r * math.Cos(theta), Y: r * math.Sin(theta)})
	}
	ring = append(ring, ring[0])
	return geom.NewPolygon(nil, ring)
}
