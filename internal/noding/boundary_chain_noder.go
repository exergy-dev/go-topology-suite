package noding

import (
	"github.com/terra-geo/terra/geom"
)

// BoundaryChainNoder is a specialised noder for polygonal coverages.
// It extracts chains of boundary segments — segments that appear in
// exactly one polygon and therefore lie on the coverage boundary —
// from the input set, dropping every interior segment that appears in
// two adjacent polygons. The output is the minimum set of edges
// describing the coverage boundary, partitioned into long chains
// rather than individual segments.
//
// Mirrors org.locationtech.jts.noding.BoundaryChainNoder. Used by the
// CoverageUnion fast path: when callers know the input is a clean
// polygonal coverage (no overlaps, no gaps), this noder is faster than
// SegmentExtractingNoder or BoundarySegmentNoder because it eliminates
// the O(boundaryEdges) noded-output overhead and produces fewer
// downstream graph nodes.
//
// No precision reduction is performed; the input must already be
// noded at the desired precision (e.g. via SnappingNoder or a
// snap-rounding pass).
type BoundaryChainNoder struct {
	input []*SegmentString
}

// NewBoundaryChainNoder builds a noder over the given segment strings.
// The slice is retained for the duration of the noder's life; do not
// mutate it concurrently.
func NewBoundaryChainNoder(strings []*SegmentString) *BoundaryChainNoder {
	return &BoundaryChainNoder{input: strings}
}

// Node satisfies the Noder interface. The input slice argument is
// ignored — the noder always operates on the segment strings supplied
// at construction. (This matches the JTS API where Node has a void
// return and getNodedSubstrings retrieves the result.)
func (n *BoundaryChainNoder) Node(_ []*SegmentString) []*SegmentString {
	return n.NodedSubstrings()
}

// NodedSubstrings returns the boundary chains. Each output chain is a
// maximal run of consecutive boundary segments from a single input
// string, optionally split at points where multiple chains meet
// (so downstream graph builders see the topological nodes).
func (n *BoundaryChainNoder) NodedSubstrings() []*SegmentString {
	if len(n.input) == 0 {
		return nil
	}

	// Step 1: hash every segment by its (sorted-endpoint, endpoint)
	// pair. A segment that appears exactly once is a boundary
	// segment; one that appears more than once is an interior edge
	// shared by two polygons and should be dropped.
	type segOwner struct {
		stringIdx int
		segIdx    int
		count     int
	}
	segMap := make(map[[2]geom.XY]*segOwner, totalSegments(n.input))
	for i, ss := range n.input {
		ns := ss.NumSegments()
		for j := 0; j < ns; j++ {
			a, b := ss.Segment(j)
			key := canonicalSegKey(a, b)
			if existing, ok := segMap[key]; ok {
				existing.count++
			} else {
				segMap[key] = &segOwner{stringIdx: i, segIdx: j, count: 1}
			}
		}
	}

	// Step 2: build per-string boundary masks and extract maximal
	// runs of consecutive boundary segments as chains.
	var chains []*SegmentString
	for i, ss := range n.input {
		ns := ss.NumSegments()
		if ns == 0 {
			continue
		}
		mask := make([]bool, ns)
		for j := 0; j < ns; j++ {
			a, b := ss.Segment(j)
			key := canonicalSegKey(a, b)
			if owner, ok := segMap[key]; ok && owner.count == 1 {
				mask[j] = true
			}
		}
		chains = append(chains, extractChains(ss, mask)...)
		_ = i
	}

	// Step 3: detect self-touching nodes — coordinates that appear
	// as interior vertices of more than one chain — and split chains
	// at those nodes so the topological graph has explicit junctions.
	nodes := findNodePoints(chains)
	if len(nodes) > 0 {
		chains = splitAtNodes(chains, nodes)
	}
	return chains
}

func totalSegments(input []*SegmentString) int {
	n := 0
	for _, ss := range input {
		n += ss.NumSegments()
	}
	return n
}

// canonicalSegKey returns a direction-independent key for segment
// (a,b): the lexicographically smaller endpoint first.
func canonicalSegKey(a, b geom.XY) [2]geom.XY {
	if a.Compare(b) <= 0 {
		return [2]geom.XY{a, b}
	}
	return [2]geom.XY{b, a}
}

// extractChains walks ss and emits one SegmentString per maximal run
// of true entries in mask.
func extractChains(ss *SegmentString, mask []bool) []*SegmentString {
	var out []*SegmentString
	i := 0
	for i < len(mask) {
		if !mask[i] {
			i++
			continue
		}
		start := i
		for i < len(mask) && mask[i] {
			i++
		}
		// Chain covers segments [start, i) → vertices [start, i].
		coords := make([]geom.XY, i-start+1)
		copy(coords, ss.Coords[start:i+1])
		out = append(out, &SegmentString{Coords: coords, Tag: ss.Tag})
	}
	return out
}

// findNodePoints returns the set of XY values that act as topological
// nodes: chain endpoints and any interior vertex shared between two
// or more distinct chains.
func findNodePoints(chains []*SegmentString) map[geom.XY]struct{} {
	interior := make(map[geom.XY]struct{})
	nodes := make(map[geom.XY]struct{})
	for _, ss := range chains {
		if len(ss.Coords) == 0 {
			continue
		}
		nodes[ss.Coords[0]] = struct{}{}
		nodes[ss.Coords[len(ss.Coords)-1]] = struct{}{}
		for j := 1; j+1 < len(ss.Coords); j++ {
			p := ss.Coords[j]
			if _, dup := interior[p]; dup {
				nodes[p] = struct{}{}
			}
			interior[p] = struct{}{}
		}
	}
	return nodes
}

// splitAtNodes splits each chain at every interior vertex that
// belongs to nodes.
func splitAtNodes(chains []*SegmentString, nodes map[geom.XY]struct{}) []*SegmentString {
	var out []*SegmentString
	for _, ss := range chains {
		out = append(out, splitChainAtNodes(ss, nodes)...)
	}
	return out
}

func splitChainAtNodes(ss *SegmentString, nodes map[geom.XY]struct{}) []*SegmentString {
	if len(ss.Coords) < 2 {
		return []*SegmentString{ss}
	}
	var out []*SegmentString
	start := 0
	for {
		end := findNodeIndexAfter(ss, start, nodes)
		if start == 0 && end == len(ss.Coords)-1 {
			// No interior split needed — keep the chain whole.
			return []*SegmentString{ss}
		}
		piece := make([]geom.XY, end-start+1)
		copy(piece, ss.Coords[start:end+1])
		out = append(out, &SegmentString{Coords: piece, Tag: ss.Tag})
		if end >= len(ss.Coords)-1 {
			break
		}
		start = end
	}
	return out
}

func findNodeIndexAfter(ss *SegmentString, start int, nodes map[geom.XY]struct{}) int {
	for i := start + 1; i < len(ss.Coords); i++ {
		if _, ok := nodes[ss.Coords[i]]; ok {
			return i
		}
	}
	return len(ss.Coords) - 1
}
