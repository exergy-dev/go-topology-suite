package bench

import "testing"

// BenchmarkIngestWKB measures the cost of decoding a stream of WKB-encoded
// polygons via wkb.Unmarshal. Workload size is IngestPolygonCount (10k,
// scaled down 100x from the 1M reference; see doc.go).
func BenchmarkIngestWKB(b *testing.B) {
	b.ReportAllocs()
	IngestWorkload(b)
}
