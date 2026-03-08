// Package kdtree provides a K-dimensional tree spatial index implementation.
// KD-trees are optimized for point data and support efficient nearest neighbor
// and range queries in k-dimensional space.
package kdtree

import (
	"math"
	"sync"

	"github.com/robert-malhotra/go-topology-suite/geom"
)

// KDTree is a binary space partitioning tree for organizing points in k-dimensional space.
// This implementation is optimized for 2D point data.
// KDTree is safe for concurrent use: multiple goroutines may call
// read methods (Query, Size, etc.) concurrently, but write methods
// (Insert, Remove, Clear) require exclusive access.
type KDTree struct {
	mu   sync.RWMutex
	root *node
	size int
}

type node struct {
	coord    geom.Coordinate
	data     interface{}
	left     *node
	right    *node
	axis     int  // 0 for X, 1 for Y
	envelope *geom.Envelope
}

// New creates a new empty KD-tree.
func New() *KDTree {
	return &KDTree{}
}

// Insert adds a point to the tree with associated data.
// Points are indexed by their X,Y coordinates.
func (t *KDTree) Insert(coord geom.Coordinate, data interface{}) {
	if coord.IsNaN() {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.root = t.insertNode(t.root, coord, data, 0)
	t.size++
}

// InsertXY adds a point to the tree given X,Y coordinates.
func (t *KDTree) InsertXY(x, y float64, data interface{}) {
	t.Insert(geom.NewCoordinate(x, y), data)
}

// InsertGeometry adds a geometry to the tree using its centroid.
// Only the centroid point is indexed, not the entire geometry.
func (t *KDTree) InsertGeometry(g geom.Geometry) {
	env := g.Envelope()
	if env.IsNull() {
		return
	}
	centroid := env.Centre()
	t.Insert(centroid, g)
}

func (t *KDTree) insertNode(n *node, coord geom.Coordinate, data interface{}, depth int) *node {
	if n == nil {
		axis := depth % 2
		return &node{
			coord:    coord,
			data:     data,
			axis:     axis,
			envelope: geom.NewEnvelopeFromCoord(coord),
		}
	}

	// Update envelope
	n.envelope.ExpandToIncludeCoord(coord)

	// Choose subtree based on current axis
	if n.axis == 0 {
		// Split on X axis
		if coord.X < n.coord.X {
			n.left = t.insertNode(n.left, coord, data, depth+1)
		} else {
			n.right = t.insertNode(n.right, coord, data, depth+1)
		}
	} else {
		// Split on Y axis
		if coord.Y < n.coord.Y {
			n.left = t.insertNode(n.left, coord, data, depth+1)
		} else {
			n.right = t.insertNode(n.right, coord, data, depth+1)
		}
	}

	return n
}

// Query returns all items whose points fall within the given envelope.
func (t *KDTree) Query(envelope *geom.Envelope) []interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.root == nil || envelope == nil || envelope.IsNull() {
		return nil
	}

	var results []interface{}
	t.queryNode(t.root, envelope, &results)
	return results
}

func (t *KDTree) queryNode(n *node, envelope *geom.Envelope, results *[]interface{}) {
	if n == nil {
		return
	}

	// Quick rejection test
	if !n.envelope.Intersects(envelope) {
		return
	}

	// Check if this point is within the query envelope
	if envelope.Contains(n.coord) {
		*results = append(*results, n.data)
	}

	// Recursively search subtrees
	// Optimize by only searching subtrees that can contain results
	if n.axis == 0 {
		// Split on X
		if envelope.MinX <= n.coord.X {
			t.queryNode(n.left, envelope, results)
		}
		if envelope.MaxX >= n.coord.X {
			t.queryNode(n.right, envelope, results)
		}
	} else {
		// Split on Y
		if envelope.MinY <= n.coord.Y {
			t.queryNode(n.left, envelope, results)
		}
		if envelope.MaxY >= n.coord.Y {
			t.queryNode(n.right, envelope, results)
		}
	}
}

// QueryPoint returns all items at the exact point location.
// Use QueryRadius for finding nearby points.
func (t *KDTree) QueryPoint(x, y float64) []interface{} {
	env := geom.NewEnvelope(x, y, x, y)
	return t.Query(env)
}

// QueryRadius returns all items within the given radius of a point.
func (t *KDTree) QueryRadius(x, y, radius float64) []interface{} {
	env := geom.NewEnvelope(x-radius, y-radius, x+radius, y+radius)
	candidates := t.Query(env)

	// Filter to only include points within the actual radius
	// (envelope query gives us a square, we need a circle)
	var results []interface{}
	query := geom.NewCoordinate(x, y)
	radiusSq := radius * radius

	for _, item := range candidates {
		// Try to get coordinate from different types
		var coord geom.Coordinate
		switch v := item.(type) {
		case geom.Coordinate:
			coord = v
		case *geom.Coordinate:
			coord = *v
		default:
			// If we can't determine the coordinate, include it
			// The caller will need to filter further if needed
			results = append(results, item)
			continue
		}

		distSq := (coord.X-query.X)*(coord.X-query.X) + (coord.Y-query.Y)*(coord.Y-query.Y)
		if distSq <= radiusSq {
			results = append(results, item)
		}
	}

	return results
}

// QueryGeometry returns items within the geometry's envelope.
func (t *KDTree) QueryGeometry(g geom.Geometry) []interface{} {
	return t.Query(g.Envelope())
}

// NearestNeighbor returns the nearest item to the given point.
// Returns nil if the tree is empty.
func (t *KDTree) NearestNeighbor(x, y float64) interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.root == nil {
		return nil
	}

	query := geom.NewCoordinate(x, y)
	best := &nearestResult{
		data: nil,
		dist: math.MaxFloat64,
	}

	t.nearestNode(t.root, query, best)
	return best.data
}

// NearestNeighborCoord is a convenience method that accepts a Coordinate.
func (t *KDTree) NearestNeighborCoord(coord geom.Coordinate) interface{} {
	return t.NearestNeighbor(coord.X, coord.Y)
}

type nearestResult struct {
	data interface{}
	dist float64
	node *node
}

func (t *KDTree) nearestNode(n *node, query geom.Coordinate, best *nearestResult) {
	if n == nil {
		return
	}

	// Calculate distance to this point
	dist := n.coord.Distance(query)
	if dist < best.dist {
		best.dist = dist
		best.data = n.data
		best.node = n
	}

	// Determine which side to search first
	var first, second *node
	var splitDist float64

	if n.axis == 0 {
		// Split on X
		splitDist = math.Abs(query.X - n.coord.X)
		if query.X < n.coord.X {
			first = n.left
			second = n.right
		} else {
			first = n.right
			second = n.left
		}
	} else {
		// Split on Y
		splitDist = math.Abs(query.Y - n.coord.Y)
		if query.Y < n.coord.Y {
			first = n.left
			second = n.right
		} else {
			first = n.right
			second = n.left
		}
	}

	// Search the near side
	t.nearestNode(first, query, best)

	// Only search the far side if it could contain a closer point
	if splitDist < best.dist {
		t.nearestNode(second, query, best)
	}
}

// NearestK returns the k nearest neighbors to the given point.
// Results are returned in order of increasing distance.
func (t *KDTree) NearestK(x, y float64, k int) []interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.root == nil || k <= 0 {
		return nil
	}

	query := geom.NewCoordinate(x, y)
	knn := &kNearestResult{
		k:       k,
		results: make([]*nearestResult, 0, k),
	}

	t.nearestKNode(t.root, query, knn)

	// Extract data from results
	data := make([]interface{}, len(knn.results))
	for i, r := range knn.results {
		data[i] = r.data
	}
	return data
}

type kNearestResult struct {
	k       int
	results []*nearestResult
}

func (knr *kNearestResult) add(result *nearestResult) {
	// Find insertion point to keep results sorted by distance
	idx := len(knr.results)
	for i, r := range knr.results {
		if result.dist < r.dist {
			idx = i
			break
		}
	}

	// Insert at the correct position
	if idx < len(knr.results) {
		// Shift and insert
		knr.results = append(knr.results[:idx+1], knr.results[idx:]...)
		knr.results[idx] = result
	} else {
		// Append at end
		knr.results = append(knr.results, result)
	}

	// Trim if we exceed k
	if len(knr.results) > knr.k {
		knr.results = knr.results[:knr.k]
	}
}

func (knr *kNearestResult) worstDist() float64 {
	if len(knr.results) < knr.k {
		return math.MaxFloat64
	}
	return knr.results[len(knr.results)-1].dist
}

func (t *KDTree) nearestKNode(n *node, query geom.Coordinate, knn *kNearestResult) {
	if n == nil {
		return
	}

	// Calculate distance to this point
	dist := n.coord.Distance(query)
	if len(knn.results) < knn.k || dist < knn.worstDist() {
		knn.add(&nearestResult{
			data: n.data,
			dist: dist,
			node: n,
		})
	}

	// Determine which side to search first
	var first, second *node
	var splitDist float64

	if n.axis == 0 {
		splitDist = math.Abs(query.X - n.coord.X)
		if query.X < n.coord.X {
			first = n.left
			second = n.right
		} else {
			first = n.right
			second = n.left
		}
	} else {
		splitDist = math.Abs(query.Y - n.coord.Y)
		if query.Y < n.coord.Y {
			first = n.left
			second = n.right
		} else {
			first = n.right
			second = n.left
		}
	}

	// Search the near side
	t.nearestKNode(first, query, knn)

	// Only search the far side if it could contain a closer point
	if splitDist < knn.worstDist() {
		t.nearestKNode(second, query, knn)
	}
}

// Size returns the number of items in the tree.
func (t *KDTree) Size() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.size
}

// IsEmpty returns true if the tree has no items.
func (t *KDTree) IsEmpty() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.size == 0
}

// Depth returns the depth of the tree.
func (t *KDTree) Depth() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.nodeDepth(t.root)
}

func (t *KDTree) nodeDepth(n *node) int {
	if n == nil {
		return 0
	}
	leftDepth := t.nodeDepth(n.left)
	rightDepth := t.nodeDepth(n.right)
	if leftDepth > rightDepth {
		return leftDepth + 1
	}
	return rightDepth + 1
}

// Envelope returns the bounding envelope of all items.
func (t *KDTree) Envelope() *geom.Envelope {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.root == nil {
		return geom.NewEnvelopeEmpty()
	}
	return t.root.envelope.Clone()
}

// Items returns all items in the tree.
func (t *KDTree) Items() []interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.root == nil {
		return nil
	}
	result := make([]interface{}, 0, t.size)
	t.collectItems(t.root, &result)
	return result
}

func (t *KDTree) collectItems(n *node, results *[]interface{}) {
	if n == nil {
		return
	}
	*results = append(*results, n.data)
	t.collectItems(n.left, results)
	t.collectItems(n.right, results)
}

// Visit traverses all items in the tree.
// The visitor function receives each coordinate and data.
// If the visitor returns false, traversal stops.
func (t *KDTree) Visit(visitor func(coord geom.Coordinate, data interface{}) bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	t.visitNode(t.root, visitor)
}

func (t *KDTree) visitNode(n *node, visitor func(geom.Coordinate, interface{}) bool) bool {
	if n == nil {
		return true
	}
	if !visitor(n.coord, n.data) {
		return false
	}
	if !t.visitNode(n.left, visitor) {
		return false
	}
	return t.visitNode(n.right, visitor)
}

// Clear removes all items from the tree.
func (t *KDTree) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.root = nil
	t.size = 0
}

// Remove removes an item from the tree at the given coordinate.
// Note: This operation rebuilds the tree and is O(n log n).
// For frequent deletions, consider rebuilding the tree from scratch.
func (t *KDTree) Remove(coord geom.Coordinate, data interface{}) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Collect all items except the one to remove
	var items []nodeData
	found := false

	t.visitNode(t.root, func(c geom.Coordinate, d interface{}) bool {
		if !found && c.Equals2D(coord, geom.DefaultEpsilon) && d == data {
			found = true
			return true
		}
		items = append(items, nodeData{coord: c, data: d})
		return true
	})

	if !found {
		return false
	}

	// Rebuild tree without the removed item
	t.root = nil
	t.size = 0
	for _, item := range items {
		t.root = t.insertNode(t.root, item.coord, item.data, 0)
		t.size++
	}

	return true
}

type nodeData struct {
	coord geom.Coordinate
	data  interface{}
}

// Build is a no-op for KD-trees since they are built incrementally.
// This method is provided for interface compatibility with other spatial indexes.
func (t *KDTree) Build() {
	// KD-trees are built incrementally during insertion
	// Nothing to do here
}
