// Package quadtree provides a quadtree spatial index implementation.
// A quadtree recursively subdivides a 2D space into four quadrants.
package quadtree

import (
	"sync"

	"github.com/robert-malhotra/go-topology-suite/geom"
)

// Quadtree is a spatial index that subdivides space into four quadrants.
// Quadtree is safe for concurrent use: multiple goroutines may call
// read methods (Query, Size, etc.) concurrently, but write methods
// (Insert, Remove, Clear) require exclusive access.
type Quadtree struct {
	mu       sync.RWMutex
	root     *node
	envelope *geom.Envelope
	size     int
	maxDepth int
	maxItems int
}

type node struct {
	envelope *geom.Envelope
	items    []*item
	children [4]*node // NW, NE, SW, SE
	depth    int
}

type item struct {
	envelope *geom.Envelope
	data     interface{}
}

// Quadrant indices
const (
	NW = 0
	NE = 1
	SW = 2
	SE = 3
)

// New creates a new Quadtree with automatic bounds.
func New() *Quadtree {
	return &Quadtree{
		envelope: nil,
		maxDepth: 20,
		maxItems: 8,
	}
}

// NewWithBounds creates a new Quadtree with specified bounds.
func NewWithBounds(envelope *geom.Envelope) *Quadtree {
	return &Quadtree{
		root:     newNode(envelope, 0),
		envelope: envelope.Clone(),
		maxDepth: 20,
		maxItems: 8,
	}
}

// NewWithOptions creates a Quadtree with custom settings.
func NewWithOptions(maxDepth, maxItems int) *Quadtree {
	return &Quadtree{
		maxDepth: maxDepth,
		maxItems: maxItems,
	}
}

func newNode(envelope *geom.Envelope, depth int) *node {
	return &node{
		envelope: envelope.Clone(),
		items:    make([]*item, 0),
		depth:    depth,
	}
}

// Insert adds an item to the quadtree with the given envelope.
func (q *Quadtree) Insert(envelope *geom.Envelope, data interface{}) {
	if envelope == nil || envelope.IsNull() {
		return
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	// Auto-expand bounds if needed
	if q.envelope == nil {
		q.envelope = envelope.Clone()
		q.root = newNode(q.envelope, 0)
	} else if !q.envelope.ContainsEnvelope(envelope) {
		q.expandBounds(envelope)
	}

	q.insertIntoNode(q.root, &item{envelope: envelope.Clone(), data: data})
	q.size++
}

// InsertGeometry adds a geometry using its envelope.
func (q *Quadtree) InsertGeometry(g geom.Geometry) {
	q.Insert(g.Envelope(), g)
}

func (q *Quadtree) expandBounds(envelope *geom.Envelope) {
	// Expand the tree bounds
	newEnv := q.envelope.Clone()
	newEnv.ExpandToInclude(envelope)

	// Round to power of 2 for clean subdivision
	width := newEnv.Width()
	height := newEnv.Height()
	maxDim := width
	if height > maxDim {
		maxDim = height
	}

	// Make square
	cx := newEnv.Centre().X
	cy := newEnv.Centre().Y
	halfDim := maxDim / 2

	q.envelope = geom.NewEnvelope(cx-halfDim, cy-halfDim, cx+halfDim, cy+halfDim)

	// Rebuild tree with new bounds
	oldItems := q.collectAllItems(q.root)
	q.root = newNode(q.envelope, 0)
	q.size = 0

	for _, it := range oldItems {
		q.insertIntoNode(q.root, it)
		q.size++
	}
}

func (q *Quadtree) collectAllItems(n *node) []*item {
	if n == nil {
		return nil
	}

	result := make([]*item, 0, len(n.items))
	result = append(result, n.items...)

	for _, child := range n.children {
		if child != nil {
			result = append(result, q.collectAllItems(child)...)
		}
	}

	return result
}

func (q *Quadtree) insertIntoNode(n *node, it *item) {
	// If this is a leaf node with space, add here
	if n.children[0] == nil && len(n.items) < q.maxItems {
		n.items = append(n.items, it)
		return
	}

	// If at max depth, add here regardless
	if n.depth >= q.maxDepth {
		n.items = append(n.items, it)
		return
	}

	// Subdivide if not already
	if n.children[0] == nil {
		q.subdivide(n)
	}

	// Find which quadrant(s) the item belongs to
	quadrant := q.findQuadrant(n, it.envelope)
	if quadrant >= 0 {
		// Item fits entirely in one quadrant
		q.insertIntoNode(n.children[quadrant], it)
	} else {
		// Item spans multiple quadrants, keep at this level
		n.items = append(n.items, it)
	}
}

func (q *Quadtree) subdivide(n *node) {
	cx := (n.envelope.MinX + n.envelope.MaxX) / 2
	cy := (n.envelope.MinY + n.envelope.MaxY) / 2

	// NW
	n.children[NW] = newNode(geom.NewEnvelope(n.envelope.MinX, cy, cx, n.envelope.MaxY), n.depth+1)
	// NE
	n.children[NE] = newNode(geom.NewEnvelope(cx, cy, n.envelope.MaxX, n.envelope.MaxY), n.depth+1)
	// SW
	n.children[SW] = newNode(geom.NewEnvelope(n.envelope.MinX, n.envelope.MinY, cx, cy), n.depth+1)
	// SE
	n.children[SE] = newNode(geom.NewEnvelope(cx, n.envelope.MinY, n.envelope.MaxX, cy), n.depth+1)

	// Redistribute existing items
	oldItems := n.items
	n.items = make([]*item, 0)
	for _, it := range oldItems {
		quadrant := q.findQuadrant(n, it.envelope)
		if quadrant >= 0 {
			q.insertIntoNode(n.children[quadrant], it)
		} else {
			n.items = append(n.items, it)
		}
	}
}

func (q *Quadtree) findQuadrant(n *node, envelope *geom.Envelope) int {
	cx := (n.envelope.MinX + n.envelope.MaxX) / 2
	cy := (n.envelope.MinY + n.envelope.MaxY) / 2

	// Check if envelope fits entirely in one quadrant
	inNorth := envelope.MinY >= cy
	inSouth := envelope.MaxY <= cy
	inWest := envelope.MaxX <= cx
	inEast := envelope.MinX >= cx

	if inNorth && inWest {
		return NW
	}
	if inNorth && inEast {
		return NE
	}
	if inSouth && inWest {
		return SW
	}
	if inSouth && inEast {
		return SE
	}

	return -1 // Spans multiple quadrants
}

// Query returns all items whose envelopes intersect the given envelope.
func (q *Quadtree) Query(envelope *geom.Envelope) []interface{} {
	q.mu.RLock()
	defer q.mu.RUnlock()
	if q.root == nil || envelope == nil || envelope.IsNull() {
		return nil
	}

	var results []interface{}
	q.queryNode(q.root, envelope, &results)
	return results
}

func (q *Quadtree) queryNode(n *node, envelope *geom.Envelope, results *[]interface{}) {
	if !n.envelope.Intersects(envelope) {
		return
	}

	// Check items at this node
	for _, it := range n.items {
		if it.envelope.Intersects(envelope) {
			*results = append(*results, it.data)
		}
	}

	// Check children
	for _, child := range n.children {
		if child != nil {
			q.queryNode(child, envelope, results)
		}
	}
}

// QueryGeometry returns items intersecting the geometry's envelope.
func (q *Quadtree) QueryGeometry(g geom.Geometry) []interface{} {
	return q.Query(g.Envelope())
}

// QueryPoint returns items containing the given point.
func (q *Quadtree) QueryPoint(x, y float64) []interface{} {
	return q.Query(geom.NewEnvelope(x, y, x, y))
}

// QueryAll returns all items in the tree.
func (q *Quadtree) QueryAll() []interface{} {
	if q.root == nil {
		return nil
	}
	return q.Query(q.envelope)
}

// Remove removes an item from the quadtree.
func (q *Quadtree) Remove(envelope *geom.Envelope, data interface{}) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.root == nil || envelope == nil || envelope.IsNull() {
		return false
	}
	removed := q.removeFromNode(q.root, envelope, data)
	if removed {
		q.size--
	}
	return removed
}

func (q *Quadtree) removeFromNode(n *node, envelope *geom.Envelope, data interface{}) bool {
	if !n.envelope.Intersects(envelope) {
		return false
	}

	// Check items at this node
	for i, it := range n.items {
		if it.data == data && it.envelope.Intersects(envelope) {
			n.items = append(n.items[:i], n.items[i+1:]...)
			return true
		}
	}

	// Check children
	for _, child := range n.children {
		if child != nil {
			if q.removeFromNode(child, envelope, data) {
				return true
			}
		}
	}

	return false
}

// Size returns the number of items in the tree.
func (q *Quadtree) Size() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.size
}

// IsEmpty returns true if the tree has no items.
func (q *Quadtree) IsEmpty() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.size == 0
}

// Depth returns the maximum depth of the tree.
func (q *Quadtree) Depth() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	if q.root == nil {
		return 0
	}
	return q.nodeDepth(q.root)
}

func (q *Quadtree) nodeDepth(n *node) int {
	maxDepth := 1
	for _, child := range n.children {
		if child != nil {
			d := q.nodeDepth(child) + 1
			if d > maxDepth {
				maxDepth = d
			}
		}
	}
	return maxDepth
}

// Envelope returns the bounds of the tree.
func (q *Quadtree) Envelope() *geom.Envelope {
	q.mu.RLock()
	defer q.mu.RUnlock()
	if q.envelope == nil {
		return geom.NewEnvelopeEmpty()
	}
	return q.envelope.Clone()
}

// Clear removes all items from the tree.
func (q *Quadtree) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.root = nil
	q.envelope = nil
	q.size = 0
}

// Visit traverses all items in the tree.
func (q *Quadtree) Visit(visitor func(envelope *geom.Envelope, data interface{}) bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	if q.root == nil {
		return
	}
	q.visitNode(q.root, visitor)
}

func (q *Quadtree) visitNode(n *node, visitor func(*geom.Envelope, interface{}) bool) bool {
	for _, it := range n.items {
		if !visitor(it.envelope, it.data) {
			return false
		}
	}

	for _, child := range n.children {
		if child != nil {
			if !q.visitNode(child, visitor) {
				return false
			}
		}
	}

	return true
}
