package bench

import (
	"testing"

	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/overlay"
	"github.com/terra-geo/terra/predicate"
	"github.com/terra-geo/terra/prepare"
	"github.com/terra-geo/terra/wkb"
)

// Workload is one iteration's worth of work for a benchmark scenario,
// shaped to plug straight into testing.B / testing.Benchmark.
type Workload func(b *testing.B)

// Workloads is the canonical ordered list of harness workloads. Both the
// _test.go benchmarks and bench/cmd/terra-bench iterate this list so the
// two entry points stay in sync.
func Workloads() []NamedWorkload {
	return []NamedWorkload{
		{Name: "IngestWKB", Fn: IngestWorkload},
		{Name: "PairwiseIntersection", Fn: PairwiseIntersectionWorkload},
		{Name: "PointInPolygon", Fn: PointInPolygonWorkload},
		{Name: "PointInPolygonPrepared", Fn: PointInPolygonPreparedWorkload},
	}
}

// NamedWorkload pairs a workload with its display name.
type NamedWorkload struct {
	Name string
	Fn   Workload
}

// IngestWorkload decodes IngestPolygonCount WKB-encoded polygons.
func IngestWorkload(b *testing.B) {
	blobs := IngestBlobs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, blob := range blobs {
			g, err := wkb.Unmarshal(blob)
			if err != nil {
				b.Fatalf("wkb.Unmarshal: %v", err)
			}
			_ = g
		}
	}
}

// PairwiseIntersectionWorkload clips PairwiseIntersectCount small polygons
// against the reference polygon.
func PairwiseIntersectionWorkload(b *testing.B) {
	ref := ReferencePolygon()
	smalls := SmallPolygons()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, s := range smalls {
			out, err := overlay.Intersection(s, ref)
			if err != nil {
				b.Fatalf("overlay.Intersection: %v", err)
			}
			_ = out
		}
	}
}

// PointInPolygonWorkload runs PointInPolygonCount queries with no prepared
// acceleration structure.
func PointInPolygonWorkload(b *testing.B) {
	ref := ReferencePolygon()
	pts := QueryPoints()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hits := 0
		for j := range pts {
			pt := geom.NewPoint(nil, pts[j])
			ok, err := predicate.Intersects(ref, pt)
			if err != nil {
				b.Fatalf("predicate.Intersects: %v", err)
			}
			if ok {
				hits++
			}
		}
		if hits < 0 {
			b.Fatal("impossible")
		}
	}
}

// PointInPolygonPreparedWorkload runs PointInPolygonCount queries with a
// prepare.Polygon-backed acceleration structure built once outside the timed
// region.
func PointInPolygonPreparedWorkload(b *testing.B) {
	ref := ReferencePolygon()
	pp := prepare.Polygon(ref)
	pts := QueryPoints()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hits := 0
		for j := range pts {
			pt := geom.NewPoint(nil, pts[j])
			ok, err := predicate.Intersects(ref, pt, predicate.WithPrepared(pp))
			if err != nil {
				b.Fatalf("predicate.Intersects (prepared): %v", err)
			}
			if ok {
				hits++
			}
		}
		if hits < 0 {
			b.Fatal("impossible")
		}
	}
}
