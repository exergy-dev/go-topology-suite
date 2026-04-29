package bench

import "testing"

// BenchmarkPointInPolygon runs PointInPolygonCount point-in-polygon queries
// against the reference polygon using predicate.Intersects, with no prepared
// acceleration structure (cold path). Compare allocs/op against
// BenchmarkPointInPolygonPrepared.
func BenchmarkPointInPolygon(b *testing.B) {
	b.ReportAllocs()
	PointInPolygonWorkload(b)
}

// BenchmarkPointInPolygonPrepared runs the same workload but with a
// prepare.Polygon-backed acceleration structure passed via
// predicate.WithPrepared.
func BenchmarkPointInPolygonPrepared(b *testing.B) {
	b.ReportAllocs()
	PointInPolygonPreparedWorkload(b)
}
