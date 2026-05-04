package index

import (
	"sync"

	"github.com/exergy-dev/go-topology-suite/geom"
)

// HPRtree is a Hilbert-Packed R-tree generic over its payload type. It is
// a port of JTS's org.locationtech.jts.index.hprtree.HPRtree.
//
// Items are inserted with their envelope; on first query the tree is
// packed: items are sorted by the Hilbert code of their envelope midpoint,
// and a series of internal layers is built bottom-up where each node
// covers `nodeCapacity` children. The internal layers and item arrays use
// flat float64 storage (4 floats per envelope) — no per-node object — so
// HPRtree is typically ~30% faster than STRtree for typical workloads.
//
// Because the index is static, insertion after the first query panics
// (matches JTS's IllegalStateException).
//
// Concurrency: build is guarded by an internal mutex; subsequent reads are
// safe for concurrent use.
type HPRtree[T any] struct {
	mu              sync.Mutex
	itemsToLoad     []hprItem[T]
	nodeCapacity    int
	numItems        int
	totalExtent     geom.Envelope
	layerStartIndex []int
	nodeBounds      []float64
	itemBounds      []float64
	itemValues      []T
	isBuilt         bool
}

type hprItem[T any] struct {
	env   geom.Envelope
	value T
}

const (
	hpEnvSize           = 4
	hpHilbertLevel      = 12
	hpDefaultNodeCap    = 16
	hilbertCodeMaxLevel = 16
)

// NewHPRtree returns a new HPRtree with the default node capacity (16).
func NewHPRtree[T any]() *HPRtree[T] {
	return NewHPRtreeWithCapacity[T](hpDefaultNodeCap)
}

// NewHPRtreeWithCapacity returns a new HPRtree with the given node capacity.
func NewHPRtreeWithCapacity[T any](nodeCapacity int) *HPRtree[T] {
	return &HPRtree[T]{
		nodeCapacity: nodeCapacity,
		totalExtent:  geom.EmptyEnvelope(),
	}
}

// Len returns the number of items in the index.
func (t *HPRtree[T]) Len() int { return t.numItems }

// Insert adds (env, value) to the index.
//
// Insert is a build-time operation. Once the tree has been queried (via
// Query, QueryVisit, or any other read operation that triggers the
// internal build), further Insert calls panic — the index is build-once
// by design, mirroring JTS's HPRtree IllegalStateException semantics.
// Bulk-load all items, then start querying.
func (t *HPRtree[T]) Insert(env geom.Envelope, value T) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.isBuilt {
		panic("HPRtree: cannot insert items after tree is built")
	}
	t.numItems++
	t.itemsToLoad = append(t.itemsToLoad, hprItem[T]{env: env, value: value})
	t.totalExtent = t.totalExtent.ExpandToInclude(env)
}

// Query collects every item whose envelope intersects search.
func (t *HPRtree[T]) Query(search geom.Envelope) []Item[T] {
	var out []Item[T]
	t.QueryVisit(search, func(it Item[T]) bool {
		out = append(out, it)
		return true
	})
	return out
}

// QueryVisit invokes visit for every candidate item. Returning false from
// visit aborts traversal early.
func (t *HPRtree[T]) QueryVisit(search geom.Envelope, visit func(Item[T]) bool) {
	t.build()
	if !t.totalExtent.Intersects(search) {
		return
	}
	if t.layerStartIndex == nil {
		// Small tree — items live directly in itemBounds/itemValues.
		t.queryItems(0, search, visit)
		return
	}
	t.queryTopLayer(search, visit)
}

// Build forces the index to pack now. Subsequent inserts will panic.
func (t *HPRtree[T]) Build() { t.build() }

func (t *HPRtree[T]) build() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.isBuilt {
		return
	}
	t.prepareIndex()
	t.prepareItems()
	t.isBuilt = true
}

func (t *HPRtree[T]) prepareIndex() {
	if len(t.itemsToLoad) <= t.nodeCapacity {
		return
	}
	t.sortItems()
	t.layerStartIndex = computeLayerIndices(t.numItems, t.nodeCapacity)
	nodeCount := t.layerStartIndex[len(t.layerStartIndex)-1] / 4
	t.nodeBounds = createBoundsArray(nodeCount)
	t.computeLeafNodes(t.layerStartIndex[1])
	for i := 1; i < len(t.layerStartIndex)-1; i++ {
		t.computeLayerNodes(i)
	}
}

func (t *HPRtree[T]) prepareItems() {
	n := len(t.itemsToLoad)
	t.itemBounds = make([]float64, n*4)
	t.itemValues = make([]T, n)
	for i, it := range t.itemsToLoad {
		t.itemBounds[i*4+0] = it.env.MinX
		t.itemBounds[i*4+1] = it.env.MinY
		t.itemBounds[i*4+2] = it.env.MaxX
		t.itemBounds[i*4+3] = it.env.MaxY
		t.itemValues[i] = it.value
	}
	t.itemsToLoad = nil
}

func createBoundsArray(size int) []float64 {
	a := make([]float64, 4*size)
	const inf = 1.0e308
	for i := 0; i < size; i++ {
		a[4*i+0] = inf
		a[4*i+1] = inf
		a[4*i+2] = -inf
		a[4*i+3] = -inf
	}
	return a
}

func (t *HPRtree[T]) computeLeafNodes(layerSize int) {
	for i := 0; i < layerSize; i += hpEnvSize {
		t.computeLeafNodeBounds(i, t.nodeCapacity*i/4)
	}
}

func (t *HPRtree[T]) computeLeafNodeBounds(nodeIndex, blockStart int) {
	for i := 0; i <= t.nodeCapacity; i++ {
		itemIdx := blockStart + i
		if itemIdx >= len(t.itemsToLoad) {
			break
		}
		env := t.itemsToLoad[itemIdx].env
		t.updateNodeBounds(nodeIndex, env.MinX, env.MinY, env.MaxX, env.MaxY)
	}
}

func (t *HPRtree[T]) computeLayerNodes(layerIndex int) {
	layerStart := t.layerStartIndex[layerIndex]
	childLayerStart := t.layerStartIndex[layerIndex-1]
	layerSize := t.layerStartIndex[layerIndex+1] - layerStart
	childLayerEnd := layerStart
	for i := 0; i < layerSize; i += hpEnvSize {
		childStart := childLayerStart + t.nodeCapacity*i
		t.computeNodeBounds(layerStart+i, childStart, childLayerEnd)
	}
}

func (t *HPRtree[T]) computeNodeBounds(nodeIndex, blockStart, nodeMaxIndex int) {
	for i := 0; i <= t.nodeCapacity; i++ {
		idx := blockStart + 4*i
		if idx >= nodeMaxIndex {
			break
		}
		t.updateNodeBounds(nodeIndex, t.nodeBounds[idx], t.nodeBounds[idx+1],
			t.nodeBounds[idx+2], t.nodeBounds[idx+3])
	}
}

func (t *HPRtree[T]) updateNodeBounds(nodeIndex int, minX, minY, maxX, maxY float64) {
	if minX < t.nodeBounds[nodeIndex] {
		t.nodeBounds[nodeIndex] = minX
	}
	if minY < t.nodeBounds[nodeIndex+1] {
		t.nodeBounds[nodeIndex+1] = minY
	}
	if maxX > t.nodeBounds[nodeIndex+2] {
		t.nodeBounds[nodeIndex+2] = maxX
	}
	if maxY > t.nodeBounds[nodeIndex+3] {
		t.nodeBounds[nodeIndex+3] = maxY
	}
}

func computeLayerIndices(itemSize, nodeCapacity int) []int {
	var out []int
	layerSize := itemSize
	index := 0
	for {
		out = append(out, index)
		layerSize = numNodesToCover(layerSize, nodeCapacity)
		index += hpEnvSize * layerSize
		if layerSize <= 1 {
			break
		}
	}
	return out
}

func numNodesToCover(nChild, nodeCapacity int) int {
	mult := nChild / nodeCapacity
	total := mult * nodeCapacity
	if total == nChild {
		return mult
	}
	return mult + 1
}

func (t *HPRtree[T]) queryTopLayer(search geom.Envelope, visit func(Item[T]) bool) bool {
	layerIndex := len(t.layerStartIndex) - 2
	layerSize := t.layerStartIndex[layerIndex+1] - t.layerStartIndex[layerIndex]
	for i := 0; i < layerSize; i += hpEnvSize {
		if !t.queryNode(layerIndex, i, search, visit) {
			return false
		}
	}
	return true
}

func (t *HPRtree[T]) queryNode(layerIndex, nodeOffset int, search geom.Envelope, visit func(Item[T]) bool) bool {
	layerStart := t.layerStartIndex[layerIndex]
	nodeIndex := layerStart + nodeOffset
	if !boundsIntersect(t.nodeBounds, nodeIndex, search) {
		return true
	}
	if layerIndex == 0 {
		childOffset := nodeOffset / hpEnvSize * t.nodeCapacity
		return t.queryItems(childOffset, search, visit)
	}
	childOffset := nodeOffset * t.nodeCapacity
	return t.queryNodeChildren(layerIndex-1, childOffset, search, visit)
}

func (t *HPRtree[T]) queryNodeChildren(layerIndex, blockOffset int, search geom.Envelope, visit func(Item[T]) bool) bool {
	layerStart := t.layerStartIndex[layerIndex]
	layerEnd := t.layerStartIndex[layerIndex+1]
	for i := 0; i < t.nodeCapacity; i++ {
		nodeOffset := blockOffset + hpEnvSize*i
		if layerStart+nodeOffset >= layerEnd {
			break
		}
		if !t.queryNode(layerIndex, nodeOffset, search, visit) {
			return false
		}
	}
	return true
}

func (t *HPRtree[T]) queryItems(blockStart int, search geom.Envelope, visit func(Item[T]) bool) bool {
	for i := 0; i < t.nodeCapacity; i++ {
		itemIndex := blockStart + i
		if itemIndex >= t.numItems {
			break
		}
		if boundsIntersect(t.itemBounds, itemIndex*hpEnvSize, search) {
			env := geom.Envelope{
				MinX: t.itemBounds[itemIndex*hpEnvSize+0],
				MinY: t.itemBounds[itemIndex*hpEnvSize+1],
				MaxX: t.itemBounds[itemIndex*hpEnvSize+2],
				MaxY: t.itemBounds[itemIndex*hpEnvSize+3],
			}
			if !visit(Item[T]{Env: env, Value: t.itemValues[itemIndex]}) {
				return false
			}
		}
	}
	return true
}

func boundsIntersect(bounds []float64, idx int, env geom.Envelope) bool {
	beyond := env.MaxX < bounds[idx] ||
		env.MaxY < bounds[idx+1] ||
		env.MinX > bounds[idx+2] ||
		env.MinY > bounds[idx+3]
	return !beyond
}

// ---------------------------------------------------------------------------
// Hilbert sort
// ---------------------------------------------------------------------------

func (t *HPRtree[T]) sortItems() {
	enc := newHilbertEncoder(hpHilbertLevel, t.totalExtent)
	values := make([]int, len(t.itemsToLoad))
	for i, it := range t.itemsToLoad {
		values[i] = enc.encode(it.env)
	}
	t.quickSortItemsIntoNodes(values, 0, len(t.itemsToLoad)-1)
}

// quickSortItemsIntoNodes uses Hoare-partitioned quicksort but stops
// partitioning once the lo/hi pair lives within the same leaf block —
// queryItems performs a linear scan there anyway, so further sorting buys
// nothing.
func (t *HPRtree[T]) quickSortItemsIntoNodes(values []int, lo, hi int) {
	if lo/t.nodeCapacity >= hi/t.nodeCapacity {
		return
	}
	pivot := t.hoarePartition(values, lo, hi)
	t.quickSortItemsIntoNodes(values, lo, pivot)
	t.quickSortItemsIntoNodes(values, pivot+1, hi)
}

func (t *HPRtree[T]) hoarePartition(values []int, lo, hi int) int {
	pivot := values[(lo+hi)>>1]
	i := lo - 1
	j := hi + 1
	for {
		for {
			i++
			if values[i] >= pivot {
				break
			}
		}
		for {
			j--
			if values[j] <= pivot {
				break
			}
		}
		if i >= j {
			return j
		}
		t.itemsToLoad[i], t.itemsToLoad[j] = t.itemsToLoad[j], t.itemsToLoad[i]
		values[i], values[j] = values[j], values[i]
	}
}

// ---------------------------------------------------------------------------
// HilbertEncoder + HilbertCode (port of JTS shape.fractal.HilbertCode)
// ---------------------------------------------------------------------------

type hilbertEncoder struct {
	level                  int
	minx, miny             float64
	strideX, strideY       float64
	hasStrideX, hasStrideY bool
}

func newHilbertEncoder(level int, extent geom.Envelope) *hilbertEncoder {
	hside := (1 << level) - 1
	enc := &hilbertEncoder{
		level: level,
		minx:  extent.MinX,
		miny:  extent.MinY,
	}
	w := extent.Width()
	h := extent.Height()
	if w > 0 {
		enc.strideX = w / float64(hside)
		enc.hasStrideX = true
	}
	if h > 0 {
		enc.strideY = h / float64(hside)
		enc.hasStrideY = true
	}
	return enc
}

func (e *hilbertEncoder) encode(env geom.Envelope) int {
	midx := env.Width()/2 + env.MinX
	midy := env.Height()/2 + env.MinY
	var x, y int
	if e.hasStrideX {
		x = int((midx - e.minx) / e.strideX)
	}
	if e.hasStrideY {
		y = int((midy - e.miny) / e.strideY)
	}
	return hilbertCodeEncode(e.level, x, y)
}

// hilbertCodeEncode is the public-domain branchless Hilbert encoder ported
// from JTS shape.fractal.HilbertCode (originally from
// http://threadlocalmutex.com/ via github.com/rawrunprotected/hilbert_curves).
func hilbertCodeEncode(level, x, y int) int {
	lvl := level
	if lvl < 1 {
		lvl = 1
	}
	if lvl > hilbertCodeMaxLevel {
		lvl = hilbertCodeMaxLevel
	}

	x <<= 16 - lvl
	y <<= 16 - lvl

	a := uint64(x ^ y)
	b := uint64(0xFFFF) ^ a
	c := uint64(0xFFFF) ^ (uint64(x) | uint64(y))
	d := uint64(x) & (uint64(y) ^ uint64(0xFFFF))

	A := a | (b >> 1)
	B := (a >> 1) ^ a
	C := ((c >> 1) ^ (b & (d >> 1))) ^ c
	D := ((a & (c >> 1)) ^ (d >> 1)) ^ d

	a, b, c, d = A, B, C, D
	A = (a & (a >> 2)) ^ (b & (b >> 2))
	B = (a & (b >> 2)) ^ (b & ((a ^ b) >> 2))
	C ^= (a & (c >> 2)) ^ (b & (d >> 2))
	D ^= (b & (c >> 2)) ^ ((a ^ b) & (d >> 2))

	a, b, c, d = A, B, C, D
	A = (a & (a >> 4)) ^ (b & (b >> 4))
	B = (a & (b >> 4)) ^ (b & ((a ^ b) >> 4))
	C ^= (a & (c >> 4)) ^ (b & (d >> 4))
	D ^= (b & (c >> 4)) ^ ((a ^ b) & (d >> 4))

	a, b, c, d = A, B, C, D
	C ^= (a & (c >> 8)) ^ (b & (d >> 8))
	D ^= (b & (c >> 8)) ^ ((a ^ b) & (d >> 8))

	a = C ^ (C >> 1)
	b = D ^ (D >> 1)

	i0 := uint64(x) ^ uint64(y)
	i1 := b | (uint64(0xFFFF) ^ (i0 | a))

	i0 = (i0 | (i0 << 8)) & 0x00FF00FF
	i0 = (i0 | (i0 << 4)) & 0x0F0F0F0F
	i0 = (i0 | (i0 << 2)) & 0x33333333
	i0 = (i0 | (i0 << 1)) & 0x55555555

	i1 = (i1 | (i1 << 8)) & 0x00FF00FF
	i1 = (i1 | (i1 << 4)) & 0x0F0F0F0F
	i1 = (i1 | (i1 << 2)) & 0x33333333
	i1 = (i1 | (i1 << 1)) & 0x55555555

	index := ((i1 << 1) | i0) >> (32 - 2*lvl)
	return int(index)
}
