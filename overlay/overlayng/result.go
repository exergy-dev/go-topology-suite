package overlayng

import "github.com/terra-geo/terra/geom"

// extractResultRings walks the boundary of the kept region — the union
// of all faces marked keep=true. A boundary half-edge e satisfies:
//
//	e.face.keep && !e.twin.face.keep
//
// At any vertex on the boundary, the boundary "enters" along one edge
// (twin of an arriving boundary half-edge) and "exits" along another
// outgoing boundary half-edge. The exit is the FIRST outgoing edge in
// CCW-around-the-vertex order, starting from twin(e), whose face is
// kept (and twin not kept).
//
// This is the standard "boundary of a face union" trace; it correctly
// handles the case where multiple kept faces meet at a vertex (the
// interior-of-kept-region edges between them are simply skipped).
func extractResultRings(d *dcel) [][]geom.XY {
	isBoundary := func(e *halfEdge) bool {
		if e.face == nil || e.twin == nil || e.twin.face == nil {
			return false
		}
		return e.face.keep && !e.twin.face.keep
	}

	var allBoundary []*halfEdge
	for _, e := range d.edges {
		if isBoundary(e) {
			allBoundary = append(allBoundary, e)
		}
	}
	if len(allBoundary) == 0 {
		return nil
	}

	visited := map[*halfEdge]bool{}
	var rings [][]geom.XY
	for _, start := range allBoundary {
		if visited[start] {
			continue
		}
		var ring []geom.XY
		cur := start
		const maxSteps = 1 << 20
		for steps := 0; steps < maxSteps; steps++ {
			if visited[cur] {
				break
			}
			visited[cur] = true
			ring = append(ring, cur.origin.p)
			next := nextBoundaryAtVertex(cur, isBoundary)
			if next == nil || next == start {
				break
			}
			cur = next
		}
		if len(ring) >= 3 {
			ring = append(ring, ring[0])
			rings = append(rings, ring)
		}
	}
	return rings
}

// nextBoundaryAtVertex returns the next outgoing boundary edge in CCW
// order around e.target, starting after twin(e). Returns nil if none.
func nextBoundaryAtVertex(e *halfEdge, isBoundary func(*halfEdge) bool) *halfEdge {
	v := e.target
	twin := e.twin
	idx := -1
	for i, oe := range v.out {
		if oe == twin {
			idx = i
			break
		}
	}
	if idx < 0 {
		return nil
	}
	n := len(v.out)
	for step := 1; step < n; step++ {
		j := (idx + step) % n
		candidate := v.out[j]
		if isBoundary(candidate) {
			return candidate
		}
	}
	return nil
}

// applyOp tags each face with its keep flag for the given operation.
// Every face — including the CW "outer" cycles — is classified by its
// own representative interior point, so multi-component DCELs correctly
// track that the CW twin of an inner component is geometrically inside
// the surrounding outer face. The single TRUE outer face (covering the
// unbounded universe) is naturally not kept because its interior point
// lies outside both inputs.
func applyOp(d *dcel, op Op) {
	for _, f := range d.faces {
		switch op {
		case OpIntersection:
			f.keep = f.inSubj && f.inClip
		case OpUnion:
			f.keep = f.inSubj || f.inClip
		case OpDifference:
			f.keep = f.inSubj && !f.inClip
		case OpSymDiff:
			f.keep = f.inSubj != f.inClip
		}
	}
}
