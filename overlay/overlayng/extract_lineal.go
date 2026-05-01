package overlayng

import "github.com/terra-geo/terra/geom"

// extractResultLines walks the DCEL after applyOp and harvests lineal
// components of the overlay result: chains of half-edges that satisfy
// the "in result" predicate for the operation but whose adjacent
// faces are NOT kept (so they're not part of a polygon ring of the
// result).
//
// This catches cases like polygon ∩ polygon where the polygons share
// a boundary segment but no overlap area: the shared segment is the
// intersection, lineal-dimension.
//
// Each chain is returned as a list of vertices [p0, p1, ..., pn].
func extractResultLines(d *dcel, op Op) [][]geom.XY {
	if d == nil {
		return nil
	}
	inResult := lineEdgePredicate(op)
	visited := map[*halfEdge]bool{}
	var lines [][]geom.XY
	for _, e := range d.edges {
		if visited[e] || visited[e.twin] {
			continue
		}
		if !inResult(e) {
			continue
		}
		// Trace the chain forward (via target's out-list, picking the
		// next outgoing in-result edge) and backward.
		chain := traceChain(e, inResult, visited)
		if len(chain) >= 2 {
			lines = append(lines, chain)
		}
	}
	return lines
}

// lineEdgePredicate returns the per-op predicate for "this edge is
// part of the lineal result". An edge qualifies iff (a) the operation
// includes its tag combination AND (b) no adjacent face is in the
// polygonal result (otherwise the edge belongs to the polygon's
// boundary, not the lineal output).
//
// Currently only Intersection produces lineal results from
// polygon-polygon overlay (shared boundary segments where neither
// side is in both inputs). The other operations rarely produce
// genuine lineal output for polygonal inputs and the edge-level
// classification is brittle, so they're disabled to avoid
// regressions.
func lineEdgePredicate(op Op) func(*halfEdge) bool {
	if op != OpIntersection {
		return func(*halfEdge) bool { return false }
	}
	return func(e *halfEdge) bool {
		if e.face == nil || e.twin == nil || e.twin.face == nil {
			return false
		}
		if e.face.keep || e.twin.face.keep {
			return false
		}
		return e.tags&0b11 == 0b11
	}
}

// traceChain extends an edge into a maximal chain of in-result edges.
// At each end, the chain extends iff there's exactly one in-result
// out-edge that continues the line (degree-2 internal node). At
// endpoints (degree-1 in-result, or degree>=3) the chain stops.
func traceChain(start *halfEdge, inResult func(*halfEdge) bool, visited map[*halfEdge]bool) []geom.XY {
	visited[start] = true
	visited[start.twin] = true

	// Walk forward from start.target.
	forward := []geom.XY{start.origin.p, start.target.p}
	cur := start
	for {
		nxt := nextInResultAt(cur.target, cur.twin, inResult, visited)
		if nxt == nil {
			break
		}
		visited[nxt] = true
		visited[nxt.twin] = true
		forward = append(forward, nxt.target.p)
		cur = nxt
	}
	// Walk backward from start.origin.
	curB := start.twin
	var backward []geom.XY
	for {
		nxt := nextInResultAt(curB.target, curB.twin, inResult, visited)
		if nxt == nil {
			break
		}
		visited[nxt] = true
		visited[nxt.twin] = true
		backward = append(backward, nxt.target.p)
		curB = nxt
	}
	if len(backward) == 0 {
		return forward
	}
	// Build [reversed-backward, forward] (skip first of forward since
	// it's the same as the last backward step's start).
	out := make([]geom.XY, 0, len(backward)+len(forward))
	for i := len(backward) - 1; i >= 0; i-- {
		out = append(out, backward[i])
	}
	out = append(out, forward...)
	return out
}

// nextInResultAt returns the unique unvisited in-result outgoing edge
// at vertex v, excluding the edge whose twin is `incoming` (we just
// arrived from it). If 0 or >1 candidates exist, returns nil.
func nextInResultAt(v *vertex, incoming *halfEdge, inResult func(*halfEdge) bool, visited map[*halfEdge]bool) *halfEdge {
	var found *halfEdge
	for _, oe := range v.out {
		if oe == incoming {
			continue
		}
		if visited[oe] {
			continue
		}
		if !inResult(oe) {
			continue
		}
		if found != nil {
			// Branching node — stop the chain here.
			return nil
		}
		found = oe
	}
	return found
}

// extractResultPoints harvests vertices that are part of the result
// but aren't covered by any extracted line or polygon ring. For
// intersection, a vertex qualifies iff it's adjacent to at least one
// subj-tagged edge AND at least one clip-tagged edge.
func extractResultPoints(d *dcel, op Op, lineCoords [][]geom.XY, polygonCoords [][]geom.XY) []geom.XY {
	if d == nil || op != OpIntersection {
		return nil
	}
	used := map[geom.XY]bool{}
	for _, c := range lineCoords {
		for _, p := range c {
			used[p] = true
		}
	}
	for _, c := range polygonCoords {
		for _, p := range c {
			used[p] = true
		}
	}
	var points []geom.XY
	for _, v := range d.vertices {
		if used[v.p] {
			continue
		}
		hasSubj, hasClip := false, false
		for _, e := range v.out {
			if e.tags&0b01 != 0 {
				hasSubj = true
			}
			if e.tags&0b10 != 0 {
				hasClip = true
			}
		}
		if hasSubj && hasClip {
			points = append(points, v.p)
		}
	}
	return points
}
