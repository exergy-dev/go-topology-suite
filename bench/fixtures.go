package bench

import (
	"math"
	"math/rand"
	"sync"

	"github.com/exergy-dev/go-topology-suite/geom"
	"github.com/exergy-dev/go-topology-suite/wkb"
)

// Scaled-down workload sizes. See doc.go for rationale.
const (
	// IngestPolygonCount is the number of synthesised WKB polygons the ingest
	// benchmark decodes per iteration. Reference scenario is 1,000,000; we
	// run 1/100 of that to keep `go test -bench` runtime bounded.
	IngestPolygonCount = 10_000

	// PairwiseIntersectCount is the number of small polygons clipped against
	// the reference polygon per iteration. Reference scenario is 10,000.
	PairwiseIntersectCount = 100

	// PointInPolygonCount is the number of point-in-polygon queries per
	// iteration. The full 100k matches the reference scenario.
	PointInPolygonCount = 100_000

	// ReferenceVertexCount is the vertex count of the synthesised "country
	// boundary" reference polygon (a regular n-gon plus closing vertex).
	ReferenceVertexCount = 50
)

var (
	ingestOnce  sync.Once
	ingestBlobs [][]byte

	refPolyOnce sync.Once
	refPoly     *geom.Polygon

	smallPolysOnce sync.Once
	smallPolys     []*geom.Polygon

	pointsOnce sync.Once
	points     []geom.XY
)

// IngestBlobs returns a deterministic slice of IngestPolygonCount WKB-encoded
// quad polygons. The slice is built once and shared across iterations.
func IngestBlobs() [][]byte {
	ingestOnce.Do(func() {
		ingestBlobs = make([][]byte, IngestPolygonCount)
		rng := rand.New(rand.NewSource(1))
		for i := range ingestBlobs {
			cx := rng.Float64() * 1000
			cy := rng.Float64() * 1000
			r := 0.1 + rng.Float64()*2
			ring := []geom.XY{
				{X: cx - r, Y: cy - r},
				{X: cx + r, Y: cy - r},
				{X: cx + r, Y: cy + r},
				{X: cx - r, Y: cy + r},
				{X: cx - r, Y: cy - r},
			}
			poly := geom.NewPolygon(nil, ring)
			b, err := wkb.Marshal(poly)
			if err != nil {
				panic(err)
			}
			ingestBlobs[i] = b
		}
	})
	return ingestBlobs
}

// ReferencePolygon returns a fixed ~50-vertex regular polygon centred at the
// origin. Approximates a "country boundary" fixture for clipping benchmarks.
func ReferencePolygon() *geom.Polygon {
	refPolyOnce.Do(func() {
		const radius = 100.0
		n := ReferenceVertexCount
		ring := make([]geom.XY, 0, n+1)
		for i := 0; i < n; i++ {
			theta := 2 * math.Pi * float64(i) / float64(n)
			ring = append(ring, geom.XY{
				X: radius * math.Cos(theta),
				Y: radius * math.Sin(theta),
			})
		}
		ring = append(ring, ring[0]) // close
		refPoly = geom.NewPolygon(nil, ring)
	})
	return refPoly
}

// SmallPolygons returns PairwiseIntersectCount small quad polygons scattered
// across and around the reference polygon's envelope, deterministic across
// runs (seeded RNG).
func SmallPolygons() []*geom.Polygon {
	smallPolysOnce.Do(func() {
		smallPolys = make([]*geom.Polygon, PairwiseIntersectCount)
		rng := rand.New(rand.NewSource(2))
		for i := range smallPolys {
			// Spread across [-150, 150]^2 so ~half the polygons partially
			// overlap the radius-100 reference and exercise clipping.
			cx := -150 + rng.Float64()*300
			cy := -150 + rng.Float64()*300
			r := 1 + rng.Float64()*5
			ring := []geom.XY{
				{X: cx - r, Y: cy - r},
				{X: cx + r, Y: cy - r},
				{X: cx + r, Y: cy + r},
				{X: cx - r, Y: cy + r},
				{X: cx - r, Y: cy - r},
			}
			smallPolys[i] = geom.NewPolygon(nil, ring)
		}
	})
	return smallPolys
}

// QueryPoints returns PointInPolygonCount deterministic random points spread
// over a square that covers the reference polygon plus a margin, so that
// roughly π/4 of the points fall inside.
func QueryPoints() []geom.XY {
	pointsOnce.Do(func() {
		points = make([]geom.XY, PointInPolygonCount)
		rng := rand.New(rand.NewSource(3))
		for i := range points {
			points[i] = geom.XY{
				X: -100 + rng.Float64()*200,
				Y: -100 + rng.Float64()*200,
			}
		}
	})
	return points
}
