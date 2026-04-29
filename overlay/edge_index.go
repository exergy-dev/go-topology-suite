package overlay

import (
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/index"
)

// edgeIdxItem is the payload type stored in the per-overlay R-tree.
// We index by integer position into the edge slice so the payload stays
// pointer-free (eight bytes per node) — the actual edge struct can be
// rebuilt from the index when needed.
type edgeIdxItem struct {
	idx int
}

// edgeIndex is a thin generic-instantiation wrapper around index.RTree
// keyed on edge envelopes.
type edgeIndex = index.RTree[edgeIdxItem]

func indexClipEdges(edges []ghEdge) *edgeIndex {
	t := index.New[edgeIdxItem]()
	items := make([]index.Item[edgeIdxItem], len(edges))
	for i, e := range edges {
		env := geom.Envelope{}
		env.MinX = min2(e.p1.X, e.p2.X)
		env.MaxX = max2(e.p1.X, e.p2.X)
		env.MinY = min2(e.p1.Y, e.p2.Y)
		env.MaxY = max2(e.p1.Y, e.p2.Y)
		items[i] = index.Item[edgeIdxItem]{Env: env, Value: edgeIdxItem{idx: i}}
	}
	t.Bulk(items)
	return t
}

func min2(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max2(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
