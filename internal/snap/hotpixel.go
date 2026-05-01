package snap

import (
	"math"

	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/index"
	"github.com/terra-geo/terra/kernel/planar"
)

// HotPixel is a grid cell that contains at least one vertex. After
// snap-rounding, every snapped vertex lies at a hot pixel's centre;
// the cell extends ±tolerance/2 from the centre on each axis.
//
// In Goodrich-Guibas snap rounding, hot pixels are the points at
// which segments must be split if they pass through the cell. Without
// such splitting the noder produces a planar subdivision in which
// some vertices are not incident to all segments that pass over them
// — overlay topology then disconnects at those vertices.
type HotPixel struct {
	// Centre is the grid-snapped vertex coordinate.
	Centre geom.XY
}

// HotPixelSet is a deduplicated, R-tree-indexed collection of hot
// pixels.
//
// "Deduplicated" means: inserting the same grid coordinate twice is a
// no-op. The set is keyed by integer grid index, so equality is
// exact — two snapped vertices that map to the same grid cell are
// the same hot pixel.
type HotPixelSet struct {
	tolerance float64
	half      float64 // tolerance/2; cached for envelope construction.
	keys      map[gridKey]struct{}
	tree      *index.RTree[HotPixel]
}

// gridKey is the integer-coordinate identity of a hot pixel cell.
// Using integer coordinates rather than the float Centre avoids any
// floating-point ambiguity in deduplication.
type gridKey struct {
	ix, iy int64
}

// NewHotPixelSet returns an empty set with the given snap tolerance.
// The tolerance must match the Rounder used to produce the input
// vertices; otherwise grid coordinates won't align.
func NewHotPixelSet(tolerance float64) *HotPixelSet {
	return &HotPixelSet{
		tolerance: tolerance,
		half:      tolerance / 2,
		keys:      make(map[gridKey]struct{}),
		tree:      index.New[HotPixel](),
	}
}

// Add records v as a hot pixel. v must already be grid-snapped (i.e.
// produced by Rounder.SnapVertex with the matching tolerance). Adding
// a duplicate vertex is a no-op.
func (s *HotPixelSet) Add(v geom.XY) {
	if !isFinite(v.X) || !isFinite(v.Y) {
		return
	}
	k := s.keyFor(v)
	if _, exists := s.keys[k]; exists {
		return
	}
	s.keys[k] = struct{}{}
	s.tree.Insert(s.envelopeFor(v), HotPixel{Centre: v})
}

// keyFor returns the integer-grid key for v. Assumes v is already
// snapped, so v / tolerance rounds to an integer; we still apply
// math.Round for robustness against minor float drift.
func (s *HotPixelSet) keyFor(v geom.XY) gridKey {
	return gridKey{
		ix: int64(math.Round(v.X / s.tolerance)),
		iy: int64(math.Round(v.Y / s.tolerance)),
	}
}

// envelopeFor returns the bounding envelope of the hot pixel cell
// centred at v.
func (s *HotPixelSet) envelopeFor(v geom.XY) geom.Envelope {
	return geom.Envelope{
		MinX: v.X - s.half,
		MinY: v.Y - s.half,
		MaxX: v.X + s.half,
		MaxY: v.Y + s.half,
	}
}

// Has reports whether v is a hot pixel in the set. v must be grid-
// snapped.
func (s *HotPixelSet) Has(v geom.XY) bool {
	if !isFinite(v.X) || !isFinite(v.Y) {
		return false
	}
	_, ok := s.keys[s.keyFor(v)]
	return ok
}

// Len returns the number of distinct hot pixels in the set.
func (s *HotPixelSet) Len() int { return len(s.keys) }

// QuerySegment returns every hot pixel whose cell envelope intersects
// the bounding box of the segment [a, b]. The caller must apply the
// finer "segment passes through cell" test on the candidates.
func (s *HotPixelSet) QuerySegment(a, b geom.XY) []HotPixel {
	env := geom.SegmentEnvelope(a, b)
	var out []HotPixel
	s.tree.Search(env, func(it index.Item[HotPixel]) bool {
		out = append(out, it.Value)
		return true
	})
	return out
}

// SegmentSplitsAt returns the list of hot pixel centres at which the
// segment [a, b] should be split. A pixel triggers a split iff:
//
//   - its centre is neither a nor b, AND
//   - the perpendicular distance from the pixel centre to the segment
//     is strictly less than tolerance/2.
//
// The half-tolerance threshold is the standard JTS test: it means the
// segment's path enters the pixel cell strictly through its interior
// (modulo the corner cases JTS handles via a tie-break, which v1
// approximates with a strict inequality and treats grazing-edge cases
// as non-splits — documented as a known divergence).
//
// The returned list is sorted by parameter t ∈ [0, 1] along the
// segment, and consecutive duplicates (within a tolerance-relative eps)
// are removed.
func (s *HotPixelSet) SegmentSplitsAt(a, b geom.XY) []geom.XY {
	candidates := s.QuerySegment(a, b)
	if len(candidates) == 0 {
		return nil
	}

	var splits []hotPixelSplit
	for _, hp := range candidates {
		if hp.Centre.Equal(a) || hp.Centre.Equal(b) {
			continue
		}
		d := planar.Default.SegmentDistance(hp.Centre, a, b)
		if d >= s.half {
			continue
		}
		t := segmentParam(a, b, hp.Centre)
		// Only count splits strictly interior to the segment.
		if t <= 0 || t >= 1 {
			continue
		}
		splits = append(splits, hotPixelSplit{t: t, centre: hp.Centre})
	}
	if len(splits) == 0 {
		return nil
	}
	// Sort by parameter t.
	sortSplitsByT(splits)
	out := make([]geom.XY, 0, len(splits))
	const tEps = 1e-12
	for i, sp := range splits {
		if i > 0 && sp.t-splits[i-1].t < tEps {
			continue
		}
		out = append(out, sp.centre)
	}
	return out
}

// hotPixelSplit is a recorded segment split point: the centre of a
// hot pixel the segment passes through, plus the parameter t at which
// it enters the segment's path.
type hotPixelSplit struct {
	t      float64
	centre geom.XY
}

// segmentParam returns the parameter t in [0, 1] such that p ≈ a + t*(b-a).
// Picks the more numerically stable axis. (Mirrors the helper in
// internal/noding; copied to keep the snap package free of internal/noding
// imports.)
func segmentParam(a, b, p geom.XY) float64 {
	dx := b.X - a.X
	dy := b.Y - a.Y
	if math.Abs(dx) >= math.Abs(dy) {
		if dx == 0 {
			return 0
		}
		return (p.X - a.X) / dx
	}
	if dy == 0 {
		return 0
	}
	return (p.Y - a.Y) / dy
}

func sortSplitsByT(s []hotPixelSplit) {
	// Insertion sort: split lists are small (O(hot pixels per segment)),
	// usually <10. Avoids the import cost of sort.Slice on the hot path.
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j].t < s[j-1].t; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

// SnapRoundRings is the full Goodrich-Guibas pipeline: snap every
// vertex to the grid, build a hot-pixel set from the unique snapped
// vertices, then split each segment at every hot pixel its interior
// passes through.
//
// The output is a topologically-consistent ring set: every
// segment-segment intersection (after rounding) is a shared vertex,
// so downstream noding sees no segment passing through a vertex it
// doesn't share.
//
// Rings that collapse under snap (fewer than 4 distinct vertices) are
// dropped, except when collapse occurs in the input's first ring (the
// outer ring of a polygon); in that case all rings derived from that
// polygon should be dropped, which the caller is responsible for —
// this function operates on a flat ring list and has no per-polygon
// awareness.
func (r *Rounder) SnapRoundRings(rings [][]geom.XY) [][]geom.XY {
	// Pass 1: snap each vertex; collect snapped rings.
	snapped := make([][]geom.XY, 0, len(rings))
	for _, ring := range rings {
		s := r.SnapRing(ring)
		if s == nil {
			continue
		}
		snapped = append(snapped, s)
	}
	if len(snapped) == 0 {
		return nil
	}

	// Pass 2: build hot pixel set from every unique snapped vertex.
	hp := NewHotPixelSet(r.tolerance)
	for _, ring := range snapped {
		for _, v := range ring {
			hp.Add(v)
		}
	}

	// Pass 3: for each segment of each ring, find hot-pixel splits and
	// emit the noded ring.
	out := make([][]geom.XY, 0, len(snapped))
	for _, ring := range snapped {
		noded := hp.NodeRing(ring)
		if noded == nil {
			continue
		}
		out = append(out, noded)
	}
	return out
}

// NodeRing returns ring with any segment that passes through a hot
// pixel split at that pixel's centre. Used by callers that want to
// share a single HotPixelSet across multiple ring sources (e.g.
// OverlayNG snapping subj and clip together so cross-input hot pixels
// are detected).
//
// ring must already be grid-snapped at the same tolerance as the set.
func (s *HotPixelSet) NodeRing(ring []geom.XY) []geom.XY {
	if len(ring) < 2 {
		return ring
	}
	out := make([]geom.XY, 0, len(ring))
	out = append(out, ring[0])
	for i := 0; i+1 < len(ring); i++ {
		a, b := ring[i], ring[i+1]
		splits := s.SegmentSplitsAt(a, b)
		for _, p := range splits {
			if n := len(out); n > 0 && out[n-1].Equal(p) {
				continue
			}
			out = append(out, p)
		}
		// Append b unless it duplicates the previous output vertex.
		if n := len(out); n > 0 && out[n-1].Equal(b) {
			continue
		}
		out = append(out, b)
	}
	// A noded ring is still a ring; verify closure.
	if len(out) >= 2 && !out[0].Equal(out[len(out)-1]) {
		out = append(out, out[0])
	}
	if len(out) < 4 {
		return nil
	}
	return out
}
