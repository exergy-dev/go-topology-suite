package index

import (
	"container/heap"
	"math"

	"github.com/terra-geo/terra/geom"
)

// ItemDistance computes the distance between a query envelope and a
// stored item. The implementation defines what "distance" means for the
// item type — point-to-point, point-to-segment, point-to-polygon, etc.
//
// The function MUST satisfy:
//
//   - Non-negative: the returned value is >= 0.
//   - Bound-consistent with the item's stored envelope: the returned
//     distance is >= the envelope-to-envelope distance between query
//     and item.Env. The Nearest traversal relies on this monotonicity
//     to safely prune branches; if Distance can return a value LESS
//     than the envelope distance, results may be wrong.
//
// Anti-reflexivity (returning +Inf when query and item are the "same"
// thing) is the caller's responsibility — Nearest itself does not
// special-case identity.
//
// Port of org.locationtech.jts.index.ItemDistance.
type ItemDistance[T any] interface {
	Distance(query geom.Envelope, item Item[T]) float64
}

// ItemDistanceFunc adapts a plain function into an ItemDistance.
type ItemDistanceFunc[T any] func(query geom.Envelope, item Item[T]) float64

// Distance dispatches to the underlying function.
func (f ItemDistanceFunc[T]) Distance(query geom.Envelope, item Item[T]) float64 {
	return f(query, item)
}

// Nearest returns the item closest to query under the supplied
// distance metric. It performs a best-first branch-and-bound traversal
// using a min-priority queue keyed on the lower-bound envelope-to-
// envelope distance from each candidate to query.
//
// The second return value is false iff the tree is empty.
//
// Port of org.locationtech.jts.index.strtree.STRtree.nearestNeighbour
// for the single-query, single-item-result variant.
func (t *RTree[T]) Nearest(query geom.Envelope, dist ItemDistance[T]) (Item[T], bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.root == nil || t.count == 0 {
		return Item[T]{}, false
	}

	// Best-first PQ. Each entry holds either a node or a leaf item;
	// the priority is the envelope-to-envelope distance from query
	// (a sound lower bound for any descendant of a node).
	pq := &nearestQueue[T]{}
	heap.Init(pq)
	heap.Push(pq, nearestEntry[T]{
		bound: envelopeDistance(query, t.root.env),
		node:  t.root,
	})

	bestDist := math.Inf(+1)
	var bestItem Item[T]
	hasBest := false

	for pq.Len() > 0 {
		e := heap.Pop(pq).(nearestEntry[T])
		if e.bound >= bestDist {
			// Every remaining candidate's bound is >= e.bound (heap
			// property), so none can beat the current best.
			break
		}
		if e.isItem {
			d := dist.Distance(query, e.item)
			if d < bestDist {
				bestDist = d
				bestItem = e.item
				hasBest = true
			}
			continue
		}
		// Expand a node: push children (or leaf items).
		n := e.node
		if n.leaf {
			for _, it := range n.items {
				bound := envelopeDistance(query, it.Env)
				if bound >= bestDist {
					continue
				}
				heap.Push(pq, nearestEntry[T]{
					bound:  bound,
					item:   it,
					isItem: true,
				})
			}
		} else {
			for _, c := range n.children {
				bound := envelopeDistance(query, c.env)
				if bound >= bestDist {
					continue
				}
				heap.Push(pq, nearestEntry[T]{
					bound: bound,
					node:  c,
				})
			}
		}
	}
	if !hasBest {
		return Item[T]{}, false
	}
	return bestItem, true
}

// envelopeDistance returns the minimum Euclidean distance between
// envelopes a and b. If they intersect the distance is 0. Returns
// +Inf when either envelope is empty (no extent — the empty envelope
// is infinitely far from anything in this metric).
func envelopeDistance(a, b geom.Envelope) float64 {
	if a.IsEmpty() || b.IsEmpty() {
		return math.Inf(+1)
	}
	dx := 0.0
	if a.MaxX < b.MinX {
		dx = b.MinX - a.MaxX
	} else if b.MaxX < a.MinX {
		dx = a.MinX - b.MaxX
	}
	dy := 0.0
	if a.MaxY < b.MinY {
		dy = b.MinY - a.MaxY
	} else if b.MaxY < a.MinY {
		dy = a.MinY - b.MaxY
	}
	if dx == 0 && dy == 0 {
		return 0
	}
	return math.Sqrt(dx*dx + dy*dy)
}

// nearestEntry is a heap element: either a tree node (when isItem is
// false) or a leaf item (when true). bound is the lower-bound distance
// from the query envelope to anything inside this entry.
type nearestEntry[T any] struct {
	bound  float64
	node   *node[T]
	item   Item[T]
	isItem bool
}

// nearestQueue is a min-heap of nearestEntry by bound. heap.Interface
// has its usual sort.Interface bones plus Push/Pop on the slice end.
type nearestQueue[T any] []nearestEntry[T]

func (q nearestQueue[T]) Len() int           { return len(q) }
func (q nearestQueue[T]) Less(i, j int) bool { return q[i].bound < q[j].bound }
func (q nearestQueue[T]) Swap(i, j int)      { q[i], q[j] = q[j], q[i] }

func (q *nearestQueue[T]) Push(x any) { *q = append(*q, x.(nearestEntry[T])) }
func (q *nearestQueue[T]) Pop() any {
	old := *q
	n := len(old)
	x := old[n-1]
	*q = old[:n-1]
	return x
}
