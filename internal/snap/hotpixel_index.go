package snap

import (
	"math"
	"math/rand"

	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/index"
)

// HotPixelIndex is a KdTree-backed index of HotPixels keyed by rounded
// coordinate. It is a port of
// org.locationtech.jts.noding.snapround.HotPixelIndex and exists as a
// faster alternative to HotPixelSet's R-tree backing for the snap-
// rounding pipeline (which only needs point lookups by rounded
// coordinate plus segment-overlap range queries).
//
// Points passed to Add need NOT be pre-rounded; the index applies the
// rounder internally so callers can feed raw input vertices directly.
//
// Thread safety: this type is not safe for concurrent writes. Reads are
// safe after the last write has returned.
type HotPixelIndex struct {
	rounder *Rounder
	half    float64
	tree    *index.KdTree[*HotPixelEntry]
}

// HotPixelEntry is the per-pixel record stored in HotPixelIndex. It
// carries the rounded centre and a "node" flag indicating whether the
// pixel was added more than once (and so must split any segment that
// passes through).
type HotPixelEntry struct {
	Centre geom.XY
	IsNode bool
}

// NewHotPixelIndex returns an empty index keyed to the given snap
// tolerance.
func NewHotPixelIndex(tolerance float64) *HotPixelIndex {
	return &HotPixelIndex{
		rounder: New(tolerance),
		half:    tolerance / 2,
		// Use a tolerance of 0 on the KdTree itself: we explicitly
		// round each input through the Rounder before insertion, so
		// duplicate detection is bit-exact at the snapped coordinate.
		tree: index.NewKdTree[*HotPixelEntry](0),
	}
}

// Tolerance returns the configured grid spacing.
func (idx *HotPixelIndex) Tolerance() float64 { return idx.rounder.tolerance }

// Len returns the number of distinct pixels.
func (idx *HotPixelIndex) Len() int { return idx.tree.Len() }

// Add inserts pt as a hot pixel. If a pixel at the same rounded
// coordinate already exists it is marked as a node (segments crossing
// it must split). Returns the resulting entry.
func (idx *HotPixelIndex) Add(pt geom.XY) *HotPixelEntry {
	if !isFinite(pt.X) || !isFinite(pt.Y) {
		return nil
	}
	rounded := idx.rounder.SnapVertex(pt)
	if existing := idx.find(rounded); existing != nil {
		// Adding twice at the same pixel implies a node.
		existing.IsNode = true
		return existing
	}
	entry := &HotPixelEntry{Centre: rounded}
	idx.tree.Insert(rounded, entry)
	return entry
}

// AddShuffled mirrors JTS HotPixelIndex.add(Coordinate[]) — randomising
// insertion order with Fisher-Yates to avoid coherent monotonic input
// runs degrading the KdTree into a near-list. Use this when feeding a
// large batch of input vertices.
func (idx *HotPixelIndex) AddShuffled(pts []geom.XY) {
	if len(pts) == 0 {
		return
	}
	// Match JTS by using a fixed seed so output is deterministic.
	rng := rand.New(rand.NewSource(13))
	indices := make([]int, len(pts))
	for i := range indices {
		indices[i] = i
	}
	for i := len(indices) - 1; i >= 0; i-- {
		j := rng.Intn(i + 1)
		idx.Add(pts[indices[j]])
		indices[j] = indices[i]
	}
}

// AddNodes adds each point as a hot pixel and marks it as a node
// (since intersection points must always split crossing segments).
func (idx *HotPixelIndex) AddNodes(pts []geom.XY) {
	for _, p := range pts {
		entry := idx.Add(p)
		if entry != nil {
			entry.IsNode = true
		}
	}
}

// find returns the entry at the (already-rounded) pixelPt, or nil.
func (idx *HotPixelIndex) find(pixelPt geom.XY) *HotPixelEntry {
	n := idx.tree.QueryPoint(pixelPt)
	if n == nil {
		return nil
	}
	return n.Value
}

// QuerySegment visits every hot pixel whose centre lies within an
// envelope expanded by one pixel width around segment p0-p1. The visit
// function MUST itself decide which candidates actually intersect the
// segment (callers can reuse HotPixelSet.SegmentSplitsAt's per-pixel
// test).
func (idx *HotPixelIndex) QuerySegment(p0, p1 geom.XY, visit func(*HotPixelEntry)) {
	if !isFinite(p0.X) || !isFinite(p0.Y) || !isFinite(p1.X) || !isFinite(p1.Y) {
		return
	}
	queryEnv := geom.SegmentEnvelope(p0, p1).ExpandBy(idx.expandRadius())
	idx.tree.Query(queryEnv, func(n *index.KdNode[*HotPixelEntry]) {
		visit(n.Value)
	})
}

// expandRadius is the safety margin added to a segment's envelope on
// every side before querying the index. JTS uses one pixel width
// (1/scaleFactor); we use the same to remain bit-compatible with the
// JTS tolerance semantics (the surrounding cell of every pixel touched
// by the segment is included in the candidate set).
func (idx *HotPixelIndex) expandRadius() float64 {
	if idx.rounder.tolerance > 0 {
		return idx.rounder.tolerance
	}
	return 0
}

// Pixels returns every entry in deterministic envelope-traversal order.
// Primarily useful for tests; production callers should query by
// segment instead.
func (idx *HotPixelIndex) Pixels() []*HotPixelEntry {
	all := idx.tree.QueryAll(geom.Envelope{
		MinX: math.Inf(-1), MinY: math.Inf(-1),
		MaxX: math.Inf(+1), MaxY: math.Inf(+1),
	})
	out := make([]*HotPixelEntry, 0, len(all))
	for _, n := range all {
		out = append(out, n.Value)
	}
	return out
}
