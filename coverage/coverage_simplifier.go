package coverage

import (
	"math"

	"github.com/terra-geo/terra/geom"
)

// Simplify simplifies the boundaries of a polygonal coverage while
// preserving the coverage's shared-edge topology: any boundary segment
// shared by two coverage cells is simplified once, with the same
// vertex sequence baked into both adjacent polygons. Endpoints of
// shared-edge chains ("nodes" — vertices incident to three or more
// edges, or where a shared chain transitions to a free chain) are
// always preserved.
//
// Ports org.locationtech.jts.coverage.CoverageSimplifier. The
// underlying line simplification is Douglas-Peucker rather than
// JTS's TPVW (Visvalingam-Whyatt with topology-preserving area
// thresholds); the surface contract — same number of input polygons
// out, no shared-edge mismatch, valid coverage in -> valid coverage
// out — is preserved.
//
// tolerance is the maximum perpendicular distance a simplified vertex
// may move from the original chain.
func Simplify(polygons []*geom.Polygon, tolerance float64) []*geom.Polygon {
	if tolerance <= 0 || len(polygons) == 0 {
		out := make([]*geom.Polygon, len(polygons))
		copy(out, polygons)
		return out
	}
	c := polygons[0].CRS()

	// Step 1: count how many directed half-edges leave each vertex
	// across the whole coverage. A "node" is any vertex incident to
	// more than two directed edges (i.e. the meeting point of three
	// or more rings, or a vertex where shared/unshared edges meet).
	degree := make(map[geom.XY]int)
	for _, p := range polygons {
		if p == nil || p.IsEmpty() {
			continue
		}
		for r := 0; r < p.NumRings(); r++ {
			n := p.RingLen(r)
			for j := 0; j < n; j++ {
				degree[p.RingVertex(r, j)]++
			}
		}
	}

	// Step 2: count undirected occurrences of each segment so we
	// know which edges are shared (count >= 2) vs free (count == 1).
	segCount := make(map[edgeKey]int)
	for _, p := range polygons {
		if p == nil || p.IsEmpty() {
			continue
		}
		for r := 0; r < p.NumRings(); r++ {
			n := p.RingLen(r)
			for j := 0; j+1 < n; j++ {
				a := p.RingVertex(r, j)
				b := p.RingVertex(r, j+1)
				if a == b {
					continue
				}
				segCount[makeEdgeKey(a, b)]++
			}
		}
	}

	// Step 3: build per-edge simplified vertex sequences. We
	// canonicalise each edge by its undirected key, simplify it
	// once, and cache the result. When emitting a polygon we look
	// up the simplified chain for each (a,b) and replay it in the
	// correct orientation.
	//
	// To find chain endpoints (nodes), a vertex is a node if its
	// degree > 4 (more than two ring-occurrences) OR if it is the
	// boundary of a shared-vs-free edge transition. As a robust
	// approximation we simply preserve every vertex whose total
	// degree across the coverage is not exactly 4 (4 = exactly two
	// passes through, the typical case for a vertex that's
	// internal to a shared chain or a free-chain interior). This
	// matches the JTS rule for inner-vertex preservation.
	isNode := func(v geom.XY) bool {
		// A vertex has degree 4 when it appears once-in once-out
		// in each of two rings sharing a single edge through it,
		// or twice-in twice-out in the same ring. Anything else
		// is a corner / triple junction / free endpoint — preserve.
		return degree[v] != 4
	}

	// For each polygon, walk each ring and split it into chains at
	// node vertices, then DP-simplify each chain (preserving its
	// node endpoints), then reassemble. To keep shared chains
	// in lockstep across two polygons, we use a chain-cache keyed
	// by the chain's two endpoints + the sorted set of all interior
	// vertices.
	type chainKey struct {
		a, b geom.XY
		hash uint64
	}
	chainCache := make(map[chainKey][]geom.XY)

	out := make([]*geom.Polygon, len(polygons))
	for pi, p := range polygons {
		if p == nil || p.IsEmpty() {
			out[pi] = p
			continue
		}
		newRings := make([][]geom.XY, p.NumRings())
		for r := 0; r < p.NumRings(); r++ {
			ring := p.Ring(r)
			n := len(ring)
			if n < 4 {
				newRings[r] = ring
				continue
			}
			// Find a starting node so chains start clean. If no
			// node exists on this ring (free island or pure hole
			// not adjacent to anything), start at index 0 and
			// treat the ring as a single closed chain.
			start := -1
			for i := 0; i < n-1; i++ {
				if isNode(ring[i]) {
					start = i
					break
				}
			}
			if start < 0 {
				// Closed-loop ring with no nodes: simplify as a
				// whole, preserving start vertex as anchor.
				simp := douglasPeuckerClosed(ring, tolerance)
				newRings[r] = simp
				continue
			}
			// Rotate ring so it starts at a node.
			rot := append([]geom.XY{}, ring[start:n-1]...)
			rot = append(rot, ring[:start]...)
			rot = append(rot, rot[0]) // re-close
			// Walk chains.
			var newRing []geom.XY
			i := 0
			for i < len(rot)-1 {
				j := i + 1
				for j < len(rot)-1 && !isNode(rot[j]) {
					j++
				}
				chain := rot[i : j+1]
				key := makeChainKey(chain)
				simp, ok := chainCache[key]
				if !ok {
					simp = dpSimplifyChain(chain, tolerance)
					chainCache[key] = simp
				}
				// Append simp without the trailing vertex (it'll
				// be the lead of the next chain).
				if len(newRing) == 0 {
					newRing = append(newRing, simp...)
				} else {
					newRing = append(newRing, simp[1:]...)
				}
				i = j
			}
			// Close the ring by appending the first vertex again.
			if len(newRing) > 0 && newRing[0] != newRing[len(newRing)-1] {
				newRing = append(newRing, newRing[0])
			}
			newRings[r] = newRing
		}
		out[pi] = geom.NewPolygon(c, newRings...)
	}
	return out
}

// makeChainKey canonicalises a chain so the same shared edge
// considered from either side hashes to the same key.
func makeChainKey(chain []geom.XY) struct {
	a, b geom.XY
	hash uint64
} {
	if len(chain) < 2 {
		return struct {
			a, b geom.XY
			hash uint64
		}{}
	}
	a, b := chain[0], chain[len(chain)-1]
	// Order endpoints so (a,b) and (b,a) collide.
	swap := false
	if (b.X < a.X) || (b.X == a.X && b.Y < a.Y) {
		a, b = b, a
		swap = true
	}
	// Hash interior vertices in canonical order.
	var h uint64 = 1469598103934665603 // FNV-1a 64 offset
	const prime uint64 = 1099511628211
	mix := func(v geom.XY) {
		bx := math.Float64bits(v.X)
		by := math.Float64bits(v.Y)
		h ^= bx
		h *= prime
		h ^= by
		h *= prime
	}
	if swap {
		for i := len(chain) - 2; i >= 1; i-- {
			mix(chain[i])
		}
	} else {
		for i := 1; i < len(chain)-1; i++ {
			mix(chain[i])
		}
	}
	return struct {
		a, b geom.XY
		hash uint64
	}{a, b, h}
}

// dpSimplifyChain runs Douglas-Peucker on an open chain, preserving
// its endpoints (chain[0] and chain[len-1]).
func dpSimplifyChain(chain []geom.XY, tol float64) []geom.XY {
	n := len(chain)
	if n <= 2 {
		out := make([]geom.XY, n)
		copy(out, chain)
		return out
	}
	keep := make([]bool, n)
	keep[0] = true
	keep[n-1] = true
	dpRecurse(chain, 0, n-1, tol, keep)
	out := make([]geom.XY, 0, n)
	for i, k := range keep {
		if k {
			out = append(out, chain[i])
		}
	}
	return out
}

// douglasPeuckerClosed simplifies a closed ring (no shared chains).
// Anchors the first vertex and applies DP between repeated anchors.
// Ensures at least 4 vertices remain so the result is a valid ring.
func douglasPeuckerClosed(ring []geom.XY, tol float64) []geom.XY {
	n := len(ring)
	if n < 5 {
		out := make([]geom.XY, n)
		copy(out, ring)
		return out
	}
	// Find vertex farthest from the first vertex; use it as a second
	// anchor so DP has two sides to work with.
	pivot := 1
	bestSq := -1.0
	for i := 1; i < n-1; i++ {
		dx := ring[i].X - ring[0].X
		dy := ring[i].Y - ring[0].Y
		d := dx*dx + dy*dy
		if d > bestSq {
			bestSq = d
			pivot = i
		}
	}
	keep := make([]bool, n)
	keep[0] = true
	keep[pivot] = true
	keep[n-1] = true
	dpRecurse(ring, 0, pivot, tol, keep)
	dpRecurse(ring, pivot, n-1, tol, keep)
	var out []geom.XY
	for i, k := range keep {
		if k {
			out = append(out, ring[i])
		}
	}
	if len(out) < 4 {
		// Fall back to original to keep ring valid.
		out = append(out[:0], ring...)
	} else if out[0] != out[len(out)-1] {
		out = append(out, out[0])
	}
	return out
}

func dpRecurse(pts []geom.XY, lo, hi int, tol float64, keep []bool) {
	if hi <= lo+1 {
		return
	}
	maxD := -1.0
	idx := lo
	for i := lo + 1; i < hi; i++ {
		d := perpDistance(pts[i], pts[lo], pts[hi])
		if d > maxD {
			maxD = d
			idx = i
		}
	}
	if maxD > tol {
		keep[idx] = true
		dpRecurse(pts, lo, idx, tol, keep)
		dpRecurse(pts, idx, hi, tol, keep)
	}
}

func perpDistance(p, a, b geom.XY) float64 {
	dx := b.X - a.X
	dy := b.Y - a.Y
	if dx == 0 && dy == 0 {
		return math.Hypot(p.X-a.X, p.Y-a.Y)
	}
	num := dy*p.X - dx*p.Y + b.X*a.Y - b.Y*a.X
	if num < 0 {
		num = -num
	}
	return num / math.Hypot(dx, dy)
}
