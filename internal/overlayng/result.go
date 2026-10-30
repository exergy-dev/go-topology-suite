package overlayng

import "github.com/exergy-dev/go-topology-suite/geom"

// extractResultRings walks the boundary of the kept region — the union
// of all faces marked keep=true. A boundary half-edge e satisfies:
//
//	e.face.keep && !e.twin.face.keep
//
// At any vertex on the boundary, the boundary "enters" along one edge
// (twin of an arriving boundary half-edge) and "exits" along another
// outgoing boundary half-edge. The exit is the FIRST outgoing edge in
// CCW-around-the-vertex order, starting from twin(e), whose face is
// kept (and twin not kept) AND in the same connected-component-of-kept
// faces as the incoming edge's face.
//
// The connected-component constraint matters at PINCH-POINT vertices,
// where two kept components touch only at a single vertex (no shared
// edge). Without the constraint, the trace's "next CCW after twin"
// rule can pick an outgoing edge whose face belongs to a DIFFERENT
// kept component, fusing what should be separate boundary loops into
// a single self-touching ring. TestOverlayAA case#9 (multipoly
// SymDifference with comb notches partially filled) exposes this
// directly.
//
// Connected components are computed by union-find over kept faces:
// two kept faces are in the same component iff they share at least
// one half-edge whose twin is also kept (i.e., a non-boundary
// interior edge of the union).
func extractResultRings(d *DCEL) [][]geom.XY {
	isBoundary := func(e *HalfEdge) bool {
		if e.Face == nil || e.Twin == nil || e.Twin.Face == nil {
			return false
		}
		return e.Face.Keep && !e.Twin.Face.Keep
	}

	// Union-find over kept faces. Faces are joined when they share an
	// interior edge (both halves kept). Pinch-point-only contact does
	// NOT join (no shared edge — only a shared vertex).
	// parent is dense, keyed by Face.index (set by traceFaceCycles);
	// nil means "not yet seen", matching the former map's zero value.
	parent := make([]*Face, len(d.Faces))
	var find func(f *Face) *Face
	find = func(f *Face) *Face {
		if parent[f.index] == nil {
			parent[f.index] = f
		}
		if parent[f.index] == f {
			return f
		}
		root := find(parent[f.index])
		parent[f.index] = root
		return root
	}
	union := func(a, b *Face) {
		ra, rb := find(a), find(b)
		if ra != rb {
			parent[ra.index] = rb
		}
	}
	for _, f := range d.Faces {
		if f.Keep {
			find(f)
		}
	}
	for _, e := range d.Edges {
		if e.Face == nil || e.Twin == nil || e.Twin.Face == nil {
			continue
		}
		if e.Face.Keep && e.Twin.Face.Keep {
			union(e.Face, e.Twin.Face)
		}
	}

	nBoundary := 0
	for _, e := range d.Edges {
		if isBoundary(e) {
			nBoundary++
		}
	}
	if nBoundary == 0 {
		return nil
	}
	allBoundary := make([]*HalfEdge, 0, nBoundary)
	for _, e := range d.Edges {
		if isBoundary(e) {
			allBoundary = append(allBoundary, e)
		}
	}

	// visited is dense over half-edges (HalfEdge.index == position in
	// d.Edges). Ring coordinates are gathered in a reused scratch
	// buffer and copied into a shared arena, each ring handed out as a
	// 3-index slice (cap == len). A ring has len(points)+1 entries and
	// at least 3 points, so 2*nBoundary bounds the arena; the make
	// fallback is unreachable but keeps the code obviously safe.
	visited := make([]bool, len(d.Edges))
	var rings [][]geom.XY
	arena := make([]geom.XY, 0, 2*nBoundary)
	var scratch []geom.XY
	for _, start := range allBoundary {
		if visited[start.index] {
			continue
		}
		scratch = scratch[:0]
		cur := start
		const maxSteps = 1 << 20
		for steps := 0; steps < maxSteps; steps++ {
			if visited[cur.index] {
				break
			}
			visited[cur.index] = true
			scratch = append(scratch, cur.Origin.P)
			next := nextBoundaryAtVertex(cur, isBoundary, find)
			if next == nil || next == start {
				break
			}
			cur = next
		}
		if len(scratch) >= 3 {
			n := len(scratch) + 1
			var ring []geom.XY
			if len(arena)+n <= cap(arena) {
				base := len(arena)
				arena = append(arena, scratch...)
				arena = append(arena, scratch[0])
				ring = arena[base : base+n : base+n]
			} else {
				ring = make([]geom.XY, 0, n)
				ring = append(ring, scratch...)
				ring = append(ring, scratch[0])
			}
			rings = append(rings, ring)
		}
	}
	return rings
}

// nextBoundaryAtVertex returns the next outgoing boundary edge in CCW
// order around e.target, starting after twin(e). The selected edge's
// face must be in the same kept-region connected component as e.face
// (per the find function); this prevents the trace from crossing
// pinch-point vertices into a different component. Returns nil if no
// same-component boundary edge exists.
//
// If find is nil, falls back to plain "first CCW boundary after twin"
// — preserves the original behaviour for callers that don't require
// component-aware tracing.
func nextBoundaryAtVertex(e *HalfEdge, isBoundary func(*HalfEdge) bool, find func(*Face) *Face) *HalfEdge {
	v := e.Target
	twin := e.Twin
	idx := -1
	for i, oe := range v.Out {
		if oe == twin {
			idx = i
			break
		}
	}
	if idx < 0 {
		return nil
	}
	var component *Face
	if find != nil && e.Face != nil {
		component = find(e.Face)
	}
	n := len(v.Out)
	for step := 1; step < n; step++ {
		j := (idx + step) % n
		candidate := v.Out[j]
		if !isBoundary(candidate) {
			continue
		}
		if component != nil && find(candidate.Face) != component {
			continue
		}
		return candidate
	}
	// Fallback: if no same-component candidate exists at this vertex,
	// any boundary edge will do (rare; happens in degenerate
	// disjoint-component traces). Try without the component filter.
	if component != nil {
		for step := 1; step < n; step++ {
			j := (idx + step) % n
			candidate := v.Out[j]
			if isBoundary(candidate) {
				return candidate
			}
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
func applyOp(d *DCEL, op Op) {
	for _, f := range d.Faces {
		switch op {
		case OpIntersection:
			f.Keep = f.inSubj && f.inClip
		case OpUnion:
			f.Keep = f.inSubj || f.inClip
		case OpDifference:
			f.Keep = f.inSubj && !f.inClip
		case OpSymDiff:
			f.Keep = f.inSubj != f.inClip
		}
	}
}
