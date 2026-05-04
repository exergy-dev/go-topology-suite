package index

import (
	"slices"
	"sync"
)

// IntervalRTree is a static 1-dimensional R-tree generic over its payload
// type. It is a port of JTS's
// org.locationtech.jts.index.intervalrtree.SortedPackedIntervalRTree.
//
// Items are inserted as [min,max] intervals; once the first query is issued
// the tree is built (sorted by interval midpoint, then packed bottom-up in
// pairs) and further insertions panic. The build is done once, lazily, under
// an internal mutex.
//
// Typical use is to index 1-D projections of 2-D objects (for example, the
// y-extents of monotone chain segments tested against a vertical query
// line).
type IntervalRTree[T any] struct {
	mu     sync.Mutex
	leaves []*intervalNode[T]
	root   *intervalNode[T]
	built  bool
}

// IntervalItem pairs an [Min,Max] interval with a payload, returned by
// IntervalRTree.Query.
type IntervalItem[T any] struct {
	Min, Max float64
	Value    T
}

// NewIntervalRTree returns an empty IntervalRTree.
func NewIntervalRTree[T any]() *IntervalRTree[T] {
	return &IntervalRTree[T]{}
}

// Insert adds (min,max,value) to the index.
//
// Insert is a build-time operation. Once the tree has been queried (via
// Query, QueryVisit, or any other read operation that triggers the
// internal build), further Insert calls panic — the index is build-once
// by design, mirroring JTS's IntervalRTree IllegalStateException
// semantics. Bulk-load all items, then start querying.
func (t *IntervalRTree[T]) Insert(min, max float64, value T) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.built {
		panic("IntervalRTree: cannot insert after first query")
	}
	t.leaves = append(t.leaves, &intervalNode[T]{
		min:    min,
		max:    max,
		isLeaf: true,
		value:  value,
	})
}

// Query invokes visit for every item whose interval intersects [min,max].
// Returning false from visit aborts traversal early.
func (t *IntervalRTree[T]) Query(min, max float64, visit func(IntervalItem[T]) bool) {
	t.build()
	if t.root == nil {
		return
	}
	t.root.query(min, max, visit)
}

// build packs the tree on first query. Idempotent; safe under concurrent
// readers because subsequent queries do not mutate.
func (t *IntervalRTree[T]) build() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.built {
		return
	}
	t.built = true
	if len(t.leaves) == 0 {
		return
	}
	// Sort leaves by midpoint.
	slices.SortFunc(t.leaves, func(a, b *intervalNode[T]) int {
		ma := (a.min + a.max) / 2
		mb := (b.min + b.max) / 2
		switch {
		case ma < mb:
			return -1
		case ma > mb:
			return 1
		default:
			return 0
		}
	})

	// Pack bottom-up two at a time.
	src := t.leaves
	for len(src) > 1 {
		dest := make([]*intervalNode[T], 0, (len(src)+1)/2)
		for i := 0; i < len(src); i += 2 {
			n1 := src[i]
			if i+1 >= len(src) {
				dest = append(dest, n1)
				continue
			}
			n2 := src[i+1]
			min := n1.min
			if n2.min < min {
				min = n2.min
			}
			max := n1.max
			if n2.max > max {
				max = n2.max
			}
			dest = append(dest, &intervalNode[T]{
				min:   min,
				max:   max,
				left:  n1,
				right: n2,
			})
		}
		src = dest
	}
	t.root = src[0]
}

// intervalNode is both a leaf and an internal node — leaves carry value,
// internals carry left/right children. This collapses the JTS abstract /
// branch / leaf hierarchy into a single struct.
type intervalNode[T any] struct {
	min, max    float64
	isLeaf      bool
	value       T
	left, right *intervalNode[T]
}

func (n *intervalNode[T]) intersects(min, max float64) bool {
	return !(n.min > max || n.max < min)
}

func (n *intervalNode[T]) query(min, max float64, visit func(IntervalItem[T]) bool) bool {
	if !n.intersects(min, max) {
		return true
	}
	if n.isLeaf {
		return visit(IntervalItem[T]{Min: n.min, Max: n.max, Value: n.value})
	}
	if n.left != nil {
		if !n.left.query(min, max, visit) {
			return false
		}
	}
	if n.right != nil {
		if !n.right.query(min, max, visit) {
			return false
		}
	}
	return true
}
