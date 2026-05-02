package overlayng

import (
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/index"
	"github.com/terra-geo/terra/internal/noding"
	"github.com/terra-geo/terra/internal/snaprounding"
)

// indexThreshold is the total-segment count at and above which we route
// the noding stage through the R-tree-backed IndexedNoder. Below it the
// brute-force SimpleNoder is competitive (no index build cost) and we
// stick with the simpler path.
//
// 64 was chosen empirically: at that point the indexed path's bulk-load
// cost has been amortised; below it SimpleNoder is faster. Adjust by
// re-running BenchmarkIndexedNoder vs BenchmarkSimpleNoder in
// index_bench_test.go.
const indexThreshold = 64

// segIdxItem is the payload stored in the per-overlay R-tree of segment
// envelopes. We keep it pointer-free so the leaves stay compact: an
// (input-string-index, segment-index) pair plus a one-byte source tag
// (subj=1, clip=2). The actual segment endpoints are recovered from the
// SegmentString slice by indexing back through stringIdx/segmentIdx.
type segIdxItem struct {
	stringIdx  int32
	segmentIdx int32
	tag        uint8
	_pad       [3]byte // explicit padding so the struct is 12 bytes flat
}

// segmentRTree wraps index.RTree[segIdxItem] so callers don't have to
// instantiate the generic at every use-site.
type segmentRTree = index.RTree[segIdxItem]

// buildSegmentIndex bulk-loads every segment of every input string into
// an R-tree. It is exported (within the package) for use by the overlay
// noding stage and by benchmarks measuring the index-build cost.
func buildSegmentIndex(strings []*noding.SegmentString) *segmentRTree {
	total := 0
	for _, ss := range strings {
		total += ss.NumSegments()
	}
	items := make([]index.Item[segIdxItem], 0, total)
	for i, ss := range strings {
		n := ss.NumSegments()
		for j := 0; j < n; j++ {
			a, b := ss.Segment(j)
			items = append(items, index.Item[segIdxItem]{
				Env: geom.SegmentEnvelope(a, b),
				Value: segIdxItem{
					stringIdx:  int32(i),
					segmentIdx: int32(j),
					tag:        uint8(ss.Tag),
				},
			})
		}
	}
	t := index.New[segIdxItem]()
	t.Bulk(items)
	return t
}

// totalSegments returns the sum of NumSegments across the input slice —
// used to decide whether the index-backed path is worth its build cost.
func totalSegments(strings []*noding.SegmentString) int {
	total := 0
	for _, ss := range strings {
		total += ss.NumSegments()
	}
	return total
}

// nodeAdaptive picks the best noder for the input size: the
// brute-force SimpleNoder for small inputs (where O(n^2) is dominated
// by constants) and the R-tree IndexedNoder once we cross the
// threshold. The selection is internal — callers see only a noded
// []*SegmentString.
func nodeAdaptive(strings []*noding.SegmentString) []*noding.SegmentString {
	if totalSegments(strings) < indexThreshold {
		return noding.SimpleNoder{}.Node(strings)
	}
	return noding.IndexedNoder{}.Node(strings)
}

// nodeAndSnap wraps nodeAdaptive with a snap-rounding post-pass.
// Tolerance > 0 routes through the snap-rounding noder, which iterates
// a noding/hot-pixel-insertion fixpoint until no segment passes through
// a hot pixel without sharing it as a vertex. Tolerance <= 0 short-
// circuits to plain noding.
//
// Non-convergence (the snap-rounding fixpoint failing to stabilise
// within MaxIter) is not propagated to the caller: the best-effort
// result is returned, matching the previous bounded-iteration
// behaviour. The harness will still surface any topological mismatch
// downstream as a divergence.
func nodeAndSnap(strings []*noding.SegmentString, tolerance float64) []*noding.SegmentString {
	if tolerance <= 0 {
		return nodeAdaptive(strings)
	}
	// Overlay-NG opts in to MergeNearCollinear: shifting result areas
	// at the tolerance level is the expected behaviour for snap-
	// rounded overlays. Buffer keeps the conservative default.
	out, _, _ := (&snaprounding.Noder{
		Tolerance:          tolerance,
		MergeNearCollinear: true,
	}).Node(strings)
	return out
}
