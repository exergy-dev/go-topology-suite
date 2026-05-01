package overlayng

import (
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/index"
	"github.com/terra-geo/terra/internal/noding"
	"github.com/terra-geo/terra/internal/snap"
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

// nodeAndSnap wraps nodeAdaptive with a snap-rounding post-pass:
// every output segment vertex is snapped to the precision grid, and
// any newly-introduced hot pixels (intersection points snapped onto
// passing segments) trigger another noding round.
//
// Iterates up to a small fixed bound; in practice JTS-style snap
// rounding converges in 1-2 passes once segments are pre-snapped.
// Pure no-tolerance callers route directly through nodeAdaptive.
func nodeAndSnap(strings []*noding.SegmentString, tolerance float64) []*noding.SegmentString {
	if tolerance <= 0 {
		return nodeAdaptive(strings)
	}
	rd := snap.New(tolerance)
	noded := nodeAdaptive(strings)
	const maxIter = 3
	for iter := 0; iter < maxIter; iter++ {
		// Snap every vertex to the grid.
		for _, s := range noded {
			for i, v := range s.Coords {
				s.Coords[i] = rd.SnapVertex(v)
			}
		}
		// Drop runs of consecutive duplicates created by snapping.
		for _, s := range noded {
			s.Coords = dedupeConsecutive(s.Coords)
		}
		// Build a hot-pixel set from the snapped vertices and re-node
		// any segment that now passes through a vertex hot pixel that
		// isn't already one of its endpoints.
		hp := snap.NewHotPixelSet(tolerance)
		for _, s := range noded {
			for _, v := range s.Coords {
				hp.Add(v)
			}
		}
		split := false
		next := make([]*noding.SegmentString, 0, len(noded))
		for _, s := range noded {
			if len(s.Coords) < 2 {
				continue
			}
			parts := splitSegmentsAtHotPixels(s.Coords, hp)
			if len(parts) > 1 {
				split = true
			}
			for _, p := range parts {
				if len(p) >= 2 {
					next = append(next, &noding.SegmentString{Coords: p, Tag: s.Tag})
				}
			}
		}
		if !split {
			return noded
		}
		// Run the noder again on the split result, then iterate.
		noded = nodeAdaptive(next)
	}
	return noded
}

// dedupeConsecutive removes runs of consecutive equal coordinates.
func dedupeConsecutive(pts []geom.XY) []geom.XY {
	if len(pts) <= 1 {
		return pts
	}
	out := pts[:1]
	for i := 1; i < len(pts); i++ {
		if pts[i] != pts[i-1] {
			out = append(out, pts[i])
		}
	}
	return out
}

// splitSegmentsAtHotPixels walks pts as a polyline and inserts hot-
// pixel centres at any segment that passes through one (other than
// at its endpoints). Returns one slice per resulting sub-string —
// typically just a single string, but multiple if any segment was
// split into more than two pieces.
func splitSegmentsAtHotPixels(pts []geom.XY, hp *snap.HotPixelSet) [][]geom.XY {
	if len(pts) < 2 {
		return [][]geom.XY{pts}
	}
	out := []geom.XY{pts[0]}
	for i := 0; i+1 < len(pts); i++ {
		a, b := pts[i], pts[i+1]
		splits := hp.SegmentSplitsAt(a, b)
		for _, sp := range splits {
			if sp != a && sp != b {
				out = append(out, sp)
			}
		}
		out = append(out, b)
	}
	return [][]geom.XY{out}
}
