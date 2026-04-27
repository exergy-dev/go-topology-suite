// Package strtree provides a Sort-Tile-Recursive (STR) tree implementation.
// STRtree is a spatial index that is built by sorting items along one axis,
// then partitioning into slices (tiles) and recursively building the tree.
package strtree

import (
	"errors"
	"math"
	"sort"
	"sync"

	"github.com/robert-malhotra/go-topology-suite/geom"
)

// ErrTreeBuilt is returned when inserting into an already-built tree.
var ErrTreeBuilt = errors.New("strtree: cannot insert into already-built tree")

// STRtree is a spatial index using the Sort-Tile-Recursive algorithm.
// It provides efficient spatial queries for large datasets.
// STRtree is safe for concurrent use: multiple goroutines may call
// read methods (Query, Size, etc.) concurrently, but write methods
// (Insert, Remove, Clear, Build) require exclusive access.
type STRtree struct {
	mu           sync.RWMutex
	root         *node
	nodeCapacity int
	items        []*item
	built        bool
}

type item struct {
	envelope *geom.Envelope
	data     interface{}
}

type node struct {
	envelope *geom.Envelope
	children []*node
	items    []*item
	level    int
}

// New creates a new STRtree with the default node capacity of 10.
func New() *STRtree {
	return NewWithCapacity(10)
}

// NewWithCapacity creates a new STRtree with the specified node capacity.
// Higher capacity means fewer nodes but potentially more items to check per node.
func NewWithCapacity(nodeCapacity int) *STRtree {
	if nodeCapacity < 2 {
		nodeCapacity = 2
	}
	return &STRtree{
		nodeCapacity: nodeCapacity,
		items:        make([]*item, 0),
		built:        false,
	}
}

// Insert adds an item to the tree with the given envelope.
// The tree must not have been built yet; returns ErrTreeBuilt otherwise.
func (t *STRtree) Insert(envelope *geom.Envelope, data interface{}) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.built {
		return ErrTreeBuilt
	}
	if envelope == nil || envelope.IsNull() {
		return nil
	}
	t.items = append(t.items, &item{
		envelope: envelope.Clone(),
		data:     data,
	})
	return nil
}

// InsertGeometry adds a geometry to the tree using its envelope.
func (t *STRtree) InsertGeometry(g geom.Geometry) error {
	return t.Insert(g.Envelope(), g)
}

// Build constructs the tree from inserted items.
// This must be called before querying.
func (t *STRtree) Build() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.buildLocked()
}

func (t *STRtree) buildLocked() {
	if t.built {
		return
	}
	t.built = true

	if len(t.items) == 0 {
		t.root = nil
		return
	}

	t.root = t.buildTree(t.items, 0)
}

func (t *STRtree) buildTree(items []*item, level int) *node {
	if len(items) == 0 {
		return nil
	}

	// If items fit in one node, create a leaf
	if len(items) <= t.nodeCapacity {
		n := &node{
			items: items,
			level: level,
		}
		n.computeEnvelope()
		return n
	}

	// Calculate number of slices
	numSlices := int(math.Ceil(float64(len(items)) / float64(t.nodeCapacity)))
	sliceSize := int(math.Ceil(float64(len(items)) / float64(numSlices)))

	// Sort by X coordinate of centroid
	sort.Slice(items, func(i, j int) bool {
		ci := items[i].envelope.Centre()
		cj := items[j].envelope.Centre()
		return ci.X < cj.X
	})

	// Create slices and sort each by Y
	var childNodes []*node
	for i := 0; i < len(items); i += sliceSize {
		end := i + sliceSize
		if end > len(items) {
			end = len(items)
		}
		slice := items[i:end]

		// Sort slice by Y
		sort.Slice(slice, func(a, b int) bool {
			ca := slice[a].envelope.Centre()
			cb := slice[b].envelope.Centre()
			return ca.Y < cb.Y
		})

		// Create nodes from slice
		for j := 0; j < len(slice); j += t.nodeCapacity {
			nodeEnd := j + t.nodeCapacity
			if nodeEnd > len(slice) {
				nodeEnd = len(slice)
			}
			nodeItems := slice[j:nodeEnd]
			childNode := &node{
				items: nodeItems,
				level: level,
			}
			childNode.computeEnvelope()
			childNodes = append(childNodes, childNode)
		}
	}

	// If we have few enough child nodes, we're done
	if len(childNodes) <= t.nodeCapacity {
		parent := &node{
			children: childNodes,
			level:    level + 1,
		}
		parent.computeEnvelopeFromChildren()
		return parent
	}

	// Otherwise, recursively build parent nodes
	return t.buildParentNodes(childNodes, level+1)
}

func (t *STRtree) buildParentNodes(nodes []*node, level int) *node {
	if len(nodes) <= t.nodeCapacity {
		parent := &node{
			children: nodes,
			level:    level,
		}
		parent.computeEnvelopeFromChildren()
		return parent
	}

	// Calculate number of slices
	numSlices := int(math.Ceil(float64(len(nodes)) / float64(t.nodeCapacity)))
	sliceSize := int(math.Ceil(float64(len(nodes)) / float64(numSlices)))

	// Sort by X
	sort.Slice(nodes, func(i, j int) bool {
		ci := nodes[i].envelope.Centre()
		cj := nodes[j].envelope.Centre()
		return ci.X < cj.X
	})

	var parentNodes []*node
	for i := 0; i < len(nodes); i += sliceSize {
		end := i + sliceSize
		if end > len(nodes) {
			end = len(nodes)
		}
		slice := nodes[i:end]

		// Sort by Y
		sort.Slice(slice, func(a, b int) bool {
			ca := slice[a].envelope.Centre()
			cb := slice[b].envelope.Centre()
			return ca.Y < cb.Y
		})

		// Group into parent nodes
		for j := 0; j < len(slice); j += t.nodeCapacity {
			nodeEnd := j + t.nodeCapacity
			if nodeEnd > len(slice) {
				nodeEnd = len(slice)
			}
			parent := &node{
				children: slice[j:nodeEnd],
				level:    level,
			}
			parent.computeEnvelopeFromChildren()
			parentNodes = append(parentNodes, parent)
		}
	}

	return t.buildParentNodes(parentNodes, level+1)
}

func (n *node) computeEnvelope() {
	n.envelope = geom.NewEnvelopeEmpty()
	for _, item := range n.items {
		n.envelope.ExpandToInclude(item.envelope)
	}
}

func (n *node) computeEnvelopeFromChildren() {
	n.envelope = geom.NewEnvelopeEmpty()
	for _, child := range n.children {
		n.envelope.ExpandToInclude(child.envelope)
	}
}

// Query returns all items whose envelopes intersect the given envelope.
func (t *STRtree) Query(envelope *geom.Envelope) []interface{} {
	t.mu.RLock()
	if t.built {
		defer t.mu.RUnlock()
		return t.queryLocked(envelope)
	}
	t.mu.RUnlock()

	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.built {
		t.buildLocked()
	}
	return t.queryLocked(envelope)
}

func (t *STRtree) queryLocked(envelope *geom.Envelope) []interface{} {
	if t.root == nil || envelope == nil || envelope.IsNull() {
		return nil
	}

	var results []interface{}
	t.queryNode(t.root, envelope, &results)
	return results
}

func (t *STRtree) queryNode(n *node, envelope *geom.Envelope, results *[]interface{}) {
	if !n.envelope.Intersects(envelope) {
		return
	}

	// Check items (leaf node)
	for _, item := range n.items {
		if item.envelope.Intersects(envelope) {
			*results = append(*results, item.data)
		}
	}

	// Check children
	for _, child := range n.children {
		t.queryNode(child, envelope, results)
	}
}

// QueryGeometry returns all items whose envelopes intersect the geometry's envelope.
func (t *STRtree) QueryGeometry(g geom.Geometry) []interface{} {
	return t.Query(g.Envelope())
}

// QueryPoint returns all items whose envelopes contain the given point.
func (t *STRtree) QueryPoint(x, y float64) []interface{} {
	env := geom.NewEnvelope(x, y, x, y)
	return t.Query(env)
}

// NearestNeighbor returns the nearest item to the given envelope.
func (t *STRtree) NearestNeighbor(envelope *geom.Envelope) interface{} {
	t.mu.RLock()
	if t.built {
		defer t.mu.RUnlock()
		return t.nearestNeighborLocked(envelope)
	}
	t.mu.RUnlock()

	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.built {
		t.buildLocked()
	}
	return t.nearestNeighborLocked(envelope)
}

func (t *STRtree) nearestNeighborLocked(envelope *geom.Envelope) interface{} {
	if t.root == nil || envelope == nil {
		return nil
	}
	nearest, _ := t.nearestNeighborNode(t.root, envelope, nil, math.MaxFloat64)
	return nearest
}

func (t *STRtree) nearestNeighborNode(n *node, target *geom.Envelope, nearest interface{}, minDist float64) (interface{}, float64) {
	if n.envelope.Distance(target) > minDist {
		return nearest, minDist
	}

	// Check items
	for _, item := range n.items {
		dist := item.envelope.Distance(target)
		if dist < minDist {
			minDist = dist
			nearest = item.data
		}
	}

	// Sort children by distance and check closest first
	type childDist struct {
		child *node
		dist  float64
	}
	var children []childDist
	for _, child := range n.children {
		children = append(children, childDist{child, child.envelope.Distance(target)})
	}
	sort.Slice(children, func(i, j int) bool {
		return children[i].dist < children[j].dist
	})

	for _, cd := range children {
		if cd.dist > minDist {
			break
		}
		nearest, minDist = t.nearestNeighborNode(cd.child, target, nearest, minDist)
	}

	return nearest, minDist
}

// Size returns the number of items in the tree.
func (t *STRtree) Size() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.items)
}

// IsEmpty returns true if the tree has no items.
func (t *STRtree) IsEmpty() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.items) == 0
}

// Depth returns the depth of the tree.
func (t *STRtree) Depth() int {
	t.mu.RLock()
	if t.built {
		defer t.mu.RUnlock()
		return t.depthLocked()
	}
	t.mu.RUnlock()

	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.built {
		t.buildLocked()
	}
	return t.depthLocked()
}

func (t *STRtree) depthLocked() int {
	if t.root == nil {
		return 0
	}
	return t.nodeDepth(t.root)
}

func (t *STRtree) nodeDepth(n *node) int {
	if len(n.children) == 0 {
		return 1
	}
	maxDepth := 0
	for _, child := range n.children {
		d := t.nodeDepth(child)
		if d > maxDepth {
			maxDepth = d
		}
	}
	return maxDepth + 1
}

// Envelope returns the bounding envelope of all items.
func (t *STRtree) Envelope() *geom.Envelope {
	t.mu.RLock()
	if t.built {
		defer t.mu.RUnlock()
		return t.envelopeLocked()
	}
	t.mu.RUnlock()

	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.built {
		t.buildLocked()
	}
	return t.envelopeLocked()
}

func (t *STRtree) envelopeLocked() *geom.Envelope {
	if t.root == nil {
		return geom.NewEnvelopeEmpty()
	}
	return t.root.envelope.Clone()
}

// Items returns all items in the tree.
func (t *STRtree) Items() []interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()
	result := make([]interface{}, len(t.items))
	for i, item := range t.items {
		result[i] = item.data
	}
	return result
}

// Remove removes an item from the tree.
// Note: This is O(n) and requires rebuilding the tree.
func (t *STRtree) Remove(envelope *geom.Envelope, data interface{}) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	for i, item := range t.items {
		if item.data == data && item.envelope.Equals(envelope, geom.DefaultEpsilon) {
			t.items = append(t.items[:i], t.items[i+1:]...)
			t.built = false
			t.root = nil
			return true
		}
	}
	return false
}

// Clear removes all items from the tree.
func (t *STRtree) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.items = make([]*item, 0)
	t.root = nil
	t.built = false
}
