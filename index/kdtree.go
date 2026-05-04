package index

import (
	"math"
	"sync"

	"github.com/terra-geo/terra/geom"
)

// KdTree is a 2D KD-tree spatial index supporting tolerance-based
// deduplication of inserted points. It is a port of
// org.locationtech.jts.index.kdtree.KdTree.
//
// The index supports point insertion, envelope (range) queries, and
// nearest-neighbour search. Insertion alternates between X- and Y-axis
// splits depending on tree depth (root splits on X, its children on Y,
// etc.).
//
// When a positive snapTolerance is configured, an Insert call that
// finds an existing node within snapTolerance of the input point
// increments that node's count instead of allocating a new node. The
// returned Item.New flag distinguishes inserts that created a new node
// from those that snapped to an existing one. With a tolerance of 0 no
// snapping is performed and every Insert allocates a new node (unless
// the coordinate is bit-exact equal at the insertion path's leaf).
//
// All read methods (Query, NearestNeighbor, Len) are safe for concurrent
// use after the last write. Concurrent writes require external
// synchronisation; the internal mutex serialises Insert against itself
// but does not protect callers that retain references to KdNode values.
type KdTree[T any] struct {
	mu              sync.RWMutex
	root            *KdNode[T]
	count           int
	snapTolerance   float64
	snapToleranceSq float64
}

// KdNode is one node of a KdTree. A node carries one inserted point
// (its Coordinate), a payload, a Count of duplicate inserts that
// snapped to this node, and pointers to its two children.
//
// AxisX reports whether this node splits the plane on the X axis.
// Children of an X-splitting node split on Y, and vice versa.
type KdNode[T any] struct {
	Coordinate geom.XY
	Value      T
	Count      int
	AxisX      bool
	Left       *KdNode[T]
	Right      *KdNode[T]
}

// NewKdTree returns a new KdTree with the given snap tolerance. A
// tolerance of 0 disables snap-based deduplication: every distinct
// (bit-equal) coordinate gets its own node.
func NewKdTree[T any](snapTolerance float64) *KdTree[T] {
	if snapTolerance < 0 {
		snapTolerance = 0
	}
	return &KdTree[T]{
		snapTolerance:   snapTolerance,
		snapToleranceSq: snapTolerance * snapTolerance,
	}
}

// Len returns the number of distinct nodes stored in the tree.
func (t *KdTree[T]) Len() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.count
}

// IsEmpty reports whether the tree has any nodes.
func (t *KdTree[T]) IsEmpty() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.root == nil
}

// Insert adds the point p with the given value to the tree. If the
// tree's snapTolerance is positive and an existing node lies within
// that tolerance of p, the existing node's Count is incremented and
// returned with isNew=false; otherwise a new node is allocated and
// returned with isNew=true.
func (t *KdTree[T]) Insert(p geom.XY, value T) (node *KdNode[T], isNew bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.root == nil {
		t.root = &KdNode[T]{Coordinate: p, Value: value, Count: 1, AxisX: true}
		t.count = 1
		return t.root, true
	}

	// Tolerance-based dedup: if any existing node is within
	// snapTolerance of p, snap to the closest (deterministic on ties).
	if t.snapTolerance > 0 {
		if match := t.findBestMatchNode(p); match != nil {
			match.Count++
			return match, false
		}
	}
	return t.insertExact(p, value)
}

// findBestMatchNode returns the closest existing node to p within the
// configured snap tolerance, breaking ties by lexicographic coordinate
// ordering for determinism. Returns nil if no node lies within tolerance.
func (t *KdTree[T]) findBestMatchNode(p geom.XY) *KdNode[T] {
	queryEnv := geom.Envelope{
		MinX: p.X - t.snapTolerance, MinY: p.Y - t.snapTolerance,
		MaxX: p.X + t.snapTolerance, MaxY: p.Y + t.snapTolerance,
	}
	var best *KdNode[T]
	bestDist := math.Inf(+1)
	t.queryEnvelope(queryEnv, func(n *KdNode[T]) {
		dx := p.X - n.Coordinate.X
		dy := p.Y - n.Coordinate.Y
		d := math.Hypot(dx, dy)
		if d > t.snapTolerance {
			return
		}
		if best == nil || d < bestDist ||
			(d == bestDist && n.Coordinate.Compare(best.Coordinate) < 0) {
			best = n
			bestDist = d
		}
	})
	return best
}

// insertExact inserts a point known to be beyond the snap tolerance of
// any existing node, walking down the alternating-axis splitting path
// to a leaf position. The point is placed deterministically by axis
// comparison, so the tree shape is a function of insertion order.
func (t *KdTree[T]) insertExact(p geom.XY, value T) (*KdNode[T], bool) {
	var parent *KdNode[T]
	curr := t.root
	goLeft := true

	for curr != nil {
		// Bit-exact dedup at any level: even with tolerance==0, two
		// inserts at exactly the same coordinate share a node so the
		// caller can detect duplicates via Count.
		dx := p.X - curr.Coordinate.X
		dy := p.Y - curr.Coordinate.Y
		distSq := dx*dx + dy*dy
		if distSq <= t.snapToleranceSq {
			curr.Count++
			return curr, false
		}
		parent = curr
		if curr.AxisX {
			goLeft = p.X < curr.Coordinate.X
		} else {
			goLeft = p.Y < curr.Coordinate.Y
		}
		if goLeft {
			curr = curr.Left
		} else {
			curr = curr.Right
		}
	}

	leaf := &KdNode[T]{
		Coordinate: p,
		Value:      value,
		Count:      1,
		AxisX:      !parent.AxisX,
	}
	if goLeft {
		parent.Left = leaf
	} else {
		parent.Right = leaf
	}
	t.count++
	return leaf, true
}

// Query visits every node whose coordinate falls within env. Visit
// order is unspecified. Returning a non-nil error from visit aborts the
// traversal (the index treats it as opaque).
func (t *KdTree[T]) Query(env geom.Envelope, visit func(*KdNode[T])) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	t.queryEnvelope(env, visit)
}

// QueryAll returns every node whose coordinate falls within env.
func (t *KdTree[T]) QueryAll(env geom.Envelope) []*KdNode[T] {
	var out []*KdNode[T]
	t.Query(env, func(n *KdNode[T]) {
		out = append(out, n)
	})
	return out
}

// queryEnvelope is the read-locked helper shared by Query and
// findBestMatchNode (which already holds the write lock during Insert).
func (t *KdTree[T]) queryEnvelope(env geom.Envelope, visit func(*KdNode[T])) {
	if t.root == nil {
		return
	}
	stack := []*KdNode[T]{t.root}
	for len(stack) > 0 {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if n == nil {
			continue
		}
		if env.ContainsXY(n.Coordinate) {
			visit(n)
		}
		if n.AxisX {
			if env.MinX <= n.Coordinate.X && n.Left != nil {
				stack = append(stack, n.Left)
			}
			if env.MaxX >= n.Coordinate.X && n.Right != nil {
				stack = append(stack, n.Right)
			}
		} else {
			if env.MinY <= n.Coordinate.Y && n.Left != nil {
				stack = append(stack, n.Left)
			}
			if env.MaxY >= n.Coordinate.Y && n.Right != nil {
				stack = append(stack, n.Right)
			}
		}
	}
}

// QueryPoint returns the node at queryPt if one exists in the tree
// (bit-exact match on both ordinates), or nil otherwise.
func (t *KdTree[T]) QueryPoint(queryPt geom.XY) *KdNode[T] {
	t.mu.RLock()
	defer t.mu.RUnlock()

	curr := t.root
	for curr != nil {
		if curr.Coordinate.EqualBitwise(queryPt) {
			return curr
		}
		var goLeft bool
		if curr.AxisX {
			goLeft = queryPt.X < curr.Coordinate.X
		} else {
			goLeft = queryPt.Y < curr.Coordinate.Y
		}
		if goLeft {
			curr = curr.Left
		} else {
			curr = curr.Right
		}
	}
	return nil
}

// NearestNeighbor returns the node closest to query, plus a found flag.
// If the tree is empty, found=false.
func (t *KdTree[T]) NearestNeighbor(query geom.XY) (node *KdNode[T], found bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.root == nil {
		return nil, false
	}
	var best *KdNode[T]
	bestSq := math.Inf(+1)

	stack := []*KdNode[T]{t.root}
	for len(stack) > 0 {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if n == nil {
			continue
		}
		dx := query.X - n.Coordinate.X
		dy := query.Y - n.Coordinate.Y
		dSq := dx*dx + dy*dy
		if dSq < bestSq {
			bestSq = dSq
			best = n
			if dSq == 0 {
				return best, true
			}
		}
		var diff float64
		if n.AxisX {
			diff = query.X - n.Coordinate.X
		} else {
			diff = query.Y - n.Coordinate.Y
		}
		var nearChild, farChild *KdNode[T]
		if diff < 0 {
			nearChild, farChild = n.Left, n.Right
		} else {
			nearChild, farChild = n.Right, n.Left
		}
		// Push far first so near is explored first (LIFO).
		if farChild != nil && diff*diff < bestSq {
			stack = append(stack, farChild)
		}
		if nearChild != nil {
			stack = append(stack, nearChild)
		}
	}
	return best, best != nil
}

// Depth reports the depth of the tree (a single root has depth 1, an
// empty tree has depth 0). Useful for diagnosing balance.
func (t *KdTree[T]) Depth() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return depthNode(t.root)
}

func depthNode[T any](n *KdNode[T]) int {
	if n == nil {
		return 0
	}
	dL := depthNode(n.Left)
	dR := depthNode(n.Right)
	if dL > dR {
		return 1 + dL
	}
	return 1 + dR
}
