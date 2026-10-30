package prepare

import "github.com/exergy-dev/go-topology-suite/geom"

const segTreeCap = 16

// seg is one directed boundary edge a -> b.
type seg struct{ a, b geom.XY }

// segTree is a static, sort-free packed R-tree over an edge sequence.
// Consecutive edges (which are spatially coherent within a ring) form
// leaves of segTreeCap edges; segTreeCap consecutive nodes form a parent,
// up to a single root. Construction is O(n); no sorting is performed.
// Queries visit exactly those edges whose own envelope intersects the
// query envelope (closed), so candidate sets match an edge-level R-tree.
type segTree struct {
	segs    []seg
	offsets []int           // offsets[l] = start of level l in bounds; level 0 = leaves
	bounds  []geom.Envelope // all node envelopes, level by level
}

func newSegTree(segs []seg) *segTree {
	t := &segTree{segs: segs}
	n := len(segs)
	if n == 0 {
		return t
	}
	t.offsets = []int{0}
	size, total := n, 0
	for {
		size = (size + segTreeCap - 1) / segTreeCap
		total += size
		t.offsets = append(t.offsets, total)
		if size <= 1 {
			break
		}
	}
	t.bounds = make([]geom.Envelope, total)
	bi := 0
	for s := 0; s < n; s += segTreeCap {
		e := min(s+segTreeCap, n)
		env := geom.SegmentEnvelope(segs[s].a, segs[s].b)
		for i := s + 1; i < e; i++ {
			env = env.ExpandToInclude(geom.SegmentEnvelope(segs[i].a, segs[i].b))
		}
		t.bounds[bi] = env
		bi++
	}
	for lvl := 1; lvl < len(t.offsets)-1; lvl++ {
		start, end := t.offsets[lvl-1], t.offsets[lvl]
		bi = t.offsets[lvl]
		for s := start; s < end; s += segTreeCap {
			e := min(s+segTreeCap, end)
			env := t.bounds[s]
			for i := s + 1; i < e; i++ {
				env = env.ExpandToInclude(t.bounds[i])
			}
			t.bounds[bi] = env
			bi++
		}
	}
	return t
}

// query calls fn for every edge whose envelope intersects q. fn returns
// false to stop; query reports whether traversal ran to completion.
func (t *segTree) query(q geom.Envelope, fn func(s *seg) bool) {
	if len(t.segs) == 0 {
		return
	}
	top := len(t.offsets) - 2
	t.node(q, top, 0, fn)
}

func (t *segTree) node(q geom.Envelope, level, idx int, fn func(*seg) bool) bool {
	if !q.Intersects(t.bounds[t.offsets[level]+idx]) {
		return true
	}
	cs := idx * segTreeCap
	if level == 0 {
		ce := min(cs+segTreeCap, len(t.segs))
		for i := cs; i < ce; i++ {
			s := &t.segs[i]
			if q.Intersects(geom.SegmentEnvelope(s.a, s.b)) {
				if !fn(s) {
					return false
				}
			}
		}
		return true
	}
	ce := min(cs+segTreeCap, t.offsets[level]-t.offsets[level-1])
	for i := cs; i < ce; i++ {
		if !t.node(q, level-1, i, fn) {
			return false
		}
	}
	return true
}
