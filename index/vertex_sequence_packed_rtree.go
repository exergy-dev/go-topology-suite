// Port of org.locationtech.jts.index.VertexSequencePackedRtree.
//
// A static R-tree packed in coordinate-sequence order. The layout
// exploits the spatial coherence of consecutive vertices in a
// LineString or Polygon ring: every NODE_CAPACITY consecutive vertices
// form a leaf node, and every NODE_CAPACITY consecutive leaves form a
// parent node, and so on up to a single root.
//
// Removal is supported (remove(i) marks the vertex inactive); the
// underlying coordinate slice is never mutated. Coordinates are not
// re-balanced after removal.
//
// Use case: simplification algorithms that need fast "is any other
// vertex of this ring inside this triangle?" tests during a corner-
// removal loop. JTS uses it inside RingHull / TPVWSimplifier / the
// PolygonEarClipper.

package index

import "github.com/exergy-dev/go-topology-suite/geom"

// vsprNodeCapacity is the fanout of every node in a
// VertexSequencePackedRtree. JTS uses 16 and notes the index is "not
// too sensitive" to this value.
const vsprNodeCapacity = 16

// VertexSequencePackedRtree is a semi-static spatial index over a
// coordinate sequence. Construction is O(N) and queries are O(log N +
// hits). Vertices may be marked removed via Remove(i); the underlying
// slice is never modified.
type VertexSequencePackedRtree struct {
	items       []geom.XY
	levelOffset []int
	bounds      []geom.Envelope
	// boundsValid is a parallel slice of validity flags so that fully
	// emptied nodes can be pruned without sentinel envelopes.
	boundsValid []bool
	isRemoved   []bool
}

// NewVertexSequencePackedRtree constructs the index over a (typically
// spatially coherent) point sequence. The slice is retained by
// reference; callers must not mutate pts after construction.
//
// Mirrors JTS VertexSequencePackedRtree(Coordinate[]).
func NewVertexSequencePackedRtree(pts []geom.XY) *VertexSequencePackedRtree {
	t := &VertexSequencePackedRtree{
		items:     pts,
		isRemoved: make([]bool, len(pts)),
	}
	t.build()
	return t
}

func (t *VertexSequencePackedRtree) build() {
	t.levelOffset = t.computeLevelOffsets()
	t.bounds, t.boundsValid = t.createBounds()
}

// computeLevelOffsets returns the prefix-sum array of per-level node
// counts. levelOffset[0]=0, levelOffset[1]=numLeaves,
// levelOffset[2]=numLeaves+numLevel1Nodes, etc. The last element is the
// position of the root.
func (t *VertexSequencePackedRtree) computeLevelOffsets() []int {
	offsets := []int{0}
	levelSize := len(t.items)
	if levelSize == 0 {
		return offsets
	}
	curr := 0
	for {
		levelSize = vsprLevelNodeCount(levelSize)
		curr += levelSize
		offsets = append(offsets, curr)
		if levelSize <= 1 {
			break
		}
	}
	return offsets
}

func vsprLevelNodeCount(numNodes int) int {
	// ceil(numNodes / nodeCapacity)
	return (numNodes + vsprNodeCapacity - 1) / vsprNodeCapacity
}

// createBounds populates the per-node envelope array. The bounds array
// is laid out level-by-level; level 0 holds the leaf-level envelopes
// (one per chunk of vsprNodeCapacity items), then level 1, etc.
func (t *VertexSequencePackedRtree) createBounds() ([]geom.Envelope, []bool) {
	if len(t.items) == 0 {
		return nil, nil
	}
	size := t.levelOffset[len(t.levelOffset)-1] + 1
	bounds := make([]geom.Envelope, size)
	valid := make([]bool, size)
	t.fillItemBounds(bounds, valid)
	for lvl := 1; lvl < len(t.levelOffset); lvl++ {
		t.fillLevelBounds(lvl, bounds, valid)
	}
	return bounds, valid
}

func (t *VertexSequencePackedRtree) fillItemBounds(bounds []geom.Envelope, valid []bool) {
	nodeStart := 0
	boundIndex := 0
	for nodeStart < len(t.items) {
		nodeEnd := min(nodeStart+vsprNodeCapacity, len(t.items))
		bounds[boundIndex] = computeItemEnvelope(t.items, nodeStart, nodeEnd)
		valid[boundIndex] = true
		boundIndex++
		nodeStart = nodeEnd
	}
}

func (t *VertexSequencePackedRtree) fillLevelBounds(lvl int, bounds []geom.Envelope, valid []bool) {
	levelStart := t.levelOffset[lvl-1]
	levelEnd := t.levelOffset[lvl]
	nodeStart := levelStart
	levelBoundIndex := t.levelOffset[lvl]
	for nodeStart < levelEnd {
		nodeEnd := min(nodeStart+vsprNodeCapacity, levelEnd)
		bounds[levelBoundIndex] = computeNodeEnvelope(bounds, valid, nodeStart, nodeEnd)
		valid[levelBoundIndex] = true
		levelBoundIndex++
		nodeStart = nodeEnd
	}
}

func computeItemEnvelope(items []geom.XY, start, end int) geom.Envelope {
	env := geom.EmptyEnvelope()
	for i := start; i < end; i++ {
		env = env.ExpandToIncludeXY(items[i])
	}
	return env
}

func computeNodeEnvelope(bounds []geom.Envelope, valid []bool, start, end int) geom.Envelope {
	env := geom.EmptyEnvelope()
	for i := start; i < end; i++ {
		if !valid[i] {
			continue
		}
		env = env.ExpandToInclude(bounds[i])
	}
	return env
}

// Query returns the indices of all (non-removed) input coordinates that
// lie within queryEnv (closed: boundary inclusive). Order is undefined.
//
// Query invokes fn for every index whose vertex lies in queryEnv
// (closed: boundary inclusive). fn returns false to stop traversal
// early. Allocation-free.
//
// Mirrors VertexSequencePackedRtree.query, returning results via
// callback (matches KdTree.Query and IntervalRtree.Query in this
// codebase).
func (t *VertexSequencePackedRtree) Query(queryEnv geom.Envelope, fn func(idx int) bool) {
	if len(t.items) == 0 {
		return
	}
	level := len(t.levelOffset) - 1
	t.queryNode(queryEnv, level, 0, fn)
}

func (t *VertexSequencePackedRtree) queryNode(queryEnv geom.Envelope, level, nodeIndex int, fn func(int) bool) bool {
	boundsIndex := t.levelOffset[level] + nodeIndex
	if !t.boundsValid[boundsIndex] {
		return true
	}
	if !queryEnv.Intersects(t.bounds[boundsIndex]) {
		return true
	}
	childStart := nodeIndex * vsprNodeCapacity
	if level == 0 {
		return t.queryItemRange(queryEnv, childStart, fn)
	}
	return t.queryNodeRange(queryEnv, level-1, childStart, fn)
}

func (t *VertexSequencePackedRtree) queryNodeRange(queryEnv geom.Envelope, level, nodeStartIndex int, fn func(int) bool) bool {
	levelMax := t.levelSize(level)
	for i := 0; i < vsprNodeCapacity; i++ {
		index := nodeStartIndex + i
		if index >= levelMax {
			return true
		}
		if !t.queryNode(queryEnv, level, index, fn) {
			return false
		}
	}
	return true
}

func (t *VertexSequencePackedRtree) levelSize(level int) int {
	return t.levelOffset[level+1] - t.levelOffset[level]
}

func (t *VertexSequencePackedRtree) queryItemRange(queryEnv geom.Envelope, itemIndex int, fn func(int) bool) bool {
	for i := 0; i < vsprNodeCapacity; i++ {
		idx := itemIndex + i
		if idx >= len(t.items) {
			return true
		}
		if t.isRemoved[idx] {
			continue
		}
		if queryEnv.ContainsXY(t.items[idx]) {
			if !fn(idx) {
				return false
			}
		}
	}
	return true
}

// Remove marks the input vertex at index inactive. The underlying
// coordinate slice is not modified. Subsequent Query calls will not
// return this index. Prunes the leaf node and its level-1 parent when
// emptied; deeper levels are left in place (matching JTS).
func (t *VertexSequencePackedRtree) Remove(index int) {
	t.isRemoved[index] = true
	nodeIndex := index / vsprNodeCapacity
	if !t.isItemsNodeEmpty(nodeIndex) {
		return
	}
	t.boundsValid[nodeIndex] = false
	if len(t.levelOffset) <= 2 {
		return
	}
	nodeLevelIndex := nodeIndex / vsprNodeCapacity
	if !t.isNodeEmpty(1, nodeLevelIndex) {
		return
	}
	t.boundsValid[t.levelOffset[1]+nodeLevelIndex] = false
}

func (t *VertexSequencePackedRtree) isItemsNodeEmpty(nodeIndex int) bool {
	start := nodeIndex * vsprNodeCapacity
	end := min(start+vsprNodeCapacity, len(t.items))
	for i := start; i < end; i++ {
		if !t.isRemoved[i] {
			return false
		}
	}
	return true
}

func (t *VertexSequencePackedRtree) isNodeEmpty(level, index int) bool {
	start := index * vsprNodeCapacity
	end := min(start+vsprNodeCapacity, t.levelOffset[level])
	for i := start; i < end; i++ {
		if t.boundsValid[i] {
			return false
		}
	}
	return true
}

// Bounds returns a copy of the bounds array (for diagnostic / test
// purposes). The result includes empty entries for removed nodes;
// callers should consult IsBoundsValid.
func (t *VertexSequencePackedRtree) Bounds() []geom.Envelope {
	out := make([]geom.Envelope, len(t.bounds))
	copy(out, t.bounds)
	return out
}
