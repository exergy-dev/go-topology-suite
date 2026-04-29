package index

import (
	"sort"

	"github.com/terra-geo/terra/geom"
)

// R*-tree split (Beckmann et al., 1990).
//
// The algorithm is in two phases:
//
//  1. ChooseSplitAxis. For each axis a in {X, Y}:
//     - sort entries by min(a) and again by max(a);
//     - for each of those two orderings, consider every distribution of the
//     entries into a "left" group of size minEntries+k and a "right" group
//     of size n-(minEntries+k), for k in [0, maxEntries-2*minEntries+1];
//     - sum the perimeters (the paper calls this "margin") of the bounding
//     boxes of the two groups across all those distributions.
//     The axis with the smaller total margin wins.
//
//  2. ChooseSplitIndex. On the winning axis, scan the same distributions and
//     pick the one with minimum overlap (area of intersection of the two
//     groups' bounding boxes). Ties are broken by minimum total area.
//
// The classic linear split (sort by MinX, halve) has been observed to visit
// 30-50% more nodes per query on real-world point/polygon data; the R*-style
// split closes most of that gap at very modest CPU cost during inserts.

// splitEntries is a uniform view over either leaf items or interior children.
// Indexing returns the entry's envelope; we only need envelopes for the
// split decision.
type splitEntries[T any] struct {
	leaf     bool
	items    []Item[T]
	children []*node[T]
}

func (s splitEntries[T]) Len() int {
	if s.leaf {
		return len(s.items)
	}
	return len(s.children)
}

func (s splitEntries[T]) env(i int) geom.Envelope {
	if s.leaf {
		return s.items[i].Env
	}
	return s.children[i].env
}

// rstarSplit chooses the best axis and split index per the R*-tree paper and
// returns the two resulting nodes. The input node n is consumed; callers
// should not reuse it.
func rstarSplit[T any](n *node[T], minEntries int) (*node[T], *node[T]) {
	se := splitEntries[T]{leaf: n.leaf, items: n.items, children: n.children}
	total := se.Len()
	// Defensive fallback: degenerate sizes can't be split sensibly.
	if total < 2*minEntries {
		// Fall back to a simple halving that still respects minEntries.
		mid := total / 2
		if mid < minEntries {
			mid = minEntries
		}
		return buildSplitNodes(n, mid)
	}

	axis := chooseSplitAxis(se, minEntries)
	// Sort entries on the winning axis by min(axis); then choose the split
	// index that minimises overlap (ties broken by area). The paper sorts
	// by both min and max and considers both — we replicate that.
	bestOverlap := -1.0
	bestArea := -1.0
	var bestSort int // 0 = by min(axis), 1 = by max(axis)
	var bestK int    // distribution: left has minEntries+k entries
	distCount := total - 2*minEntries + 1

	for sortKind := 0; sortKind < 2; sortKind++ {
		sortByAxis(se, axis, sortKind == 1)
		for k := 0; k < distCount; k++ {
			cut := minEntries + k
			leftEnv, rightEnv := boundingPair(se, cut)
			ov := overlapArea(leftEnv, rightEnv)
			ar := leftEnv.Area() + rightEnv.Area()
			if bestOverlap < 0 || ov < bestOverlap ||
				(ov == bestOverlap && ar < bestArea) {
				bestOverlap = ov
				bestArea = ar
				bestSort = sortKind
				bestK = k
			}
		}
	}

	// Re-apply the winning order, then materialise the two nodes.
	sortByAxis(se, axis, bestSort == 1)
	cut := minEntries + bestK
	return buildSplitNodes(n, cut)
}

// chooseSplitAxis returns 0 for X, 1 for Y by minimising the total margin
// across all distributions on each axis.
func chooseSplitAxis[T any](se splitEntries[T], minEntries int) int {
	mX := totalMargin(se, 0, minEntries)
	mY := totalMargin(se, 1, minEntries)
	if mY < mX {
		return 1
	}
	return 0
}

// totalMargin sums the perimeters of the two-group bounding boxes across
// every valid distribution, summed over both the min-sort and max-sort
// orderings on the given axis.
func totalMargin[T any](se splitEntries[T], axis, minEntries int) float64 {
	total := se.Len()
	distCount := total - 2*minEntries + 1
	var sum float64
	for sortKind := 0; sortKind < 2; sortKind++ {
		sortByAxis(se, axis, sortKind == 1)
		for k := 0; k < distCount; k++ {
			cut := minEntries + k
			leftEnv, rightEnv := boundingPair(se, cut)
			sum += margin(leftEnv) + margin(rightEnv)
		}
	}
	return sum
}

// boundingPair computes the envelopes of [0:cut) and [cut:n) of se.
// O(n) per call; we accept the cost — split is rare relative to query.
func boundingPair[T any](se splitEntries[T], cut int) (geom.Envelope, geom.Envelope) {
	left := geom.EmptyEnvelope()
	for i := 0; i < cut; i++ {
		left = left.ExpandToInclude(se.env(i))
	}
	right := geom.EmptyEnvelope()
	for i := cut; i < se.Len(); i++ {
		right = right.ExpandToInclude(se.env(i))
	}
	return left, right
}

func margin(e geom.Envelope) float64 {
	if e.IsEmpty() {
		return 0
	}
	return 2 * (e.Width() + e.Height())
}

// overlapArea returns the area of the intersection of a and b, or 0 if they
// don't overlap.
func overlapArea(a, b geom.Envelope) float64 {
	if a.IsEmpty() || b.IsEmpty() {
		return 0
	}
	minX := maxFloat(a.MinX, b.MinX)
	minY := maxFloat(a.MinY, b.MinY)
	maxX := minFloat(a.MaxX, b.MaxX)
	maxY := minFloat(a.MaxY, b.MaxY)
	if minX > maxX || minY > maxY {
		return 0
	}
	return (maxX - minX) * (maxY - minY)
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// sortByAxis orders the entries by min(axis) when useMax is false, else by
// max(axis). axis: 0 = X, 1 = Y.
func sortByAxis[T any](se splitEntries[T], axis int, useMax bool) {
	key := func(e geom.Envelope) float64 {
		switch {
		case axis == 0 && !useMax:
			return e.MinX
		case axis == 0 && useMax:
			return e.MaxX
		case axis == 1 && !useMax:
			return e.MinY
		default:
			return e.MaxY
		}
	}
	if se.leaf {
		sort.SliceStable(se.items, func(i, j int) bool {
			return key(se.items[i].Env) < key(se.items[j].Env)
		})
		return
	}
	sort.SliceStable(se.children, func(i, j int) bool {
		return key(se.children[i].env) < key(se.children[j].env)
	})
}

// buildSplitNodes materialises the two nodes from the (now-ordered) input
// using the cut index. The original node's slices are split in place.
func buildSplitNodes[T any](n *node[T], cut int) (*node[T], *node[T]) {
	left := &node[T]{leaf: n.leaf}
	right := &node[T]{leaf: n.leaf}
	if n.leaf {
		left.items = append(left.items, n.items[:cut]...)
		right.items = append(right.items, n.items[cut:]...)
	} else {
		left.children = append(left.children, n.children[:cut]...)
		right.children = append(right.children, n.children[cut:]...)
	}
	recomputeEnvelope(left)
	recomputeEnvelope(right)
	return left, right
}
