package simplify

import (
	"github.com/exergy-dev/go-topology-suite/geom"
	"github.com/exergy-dev/go-topology-suite/internal/geomath"
)

// indexThreshold is the total input vertex count at or above which the
// spatial indexes are used; below it the brute-force scans win because
// building and querying the trees costs more than it saves.
//
// Measured on Apple M3 (BenchmarkTopologyPreserving on benchfix.Star,
// ms/op, brute force vs indexed): n=512 0.7 vs 1.0; n=1024 2.7 vs 3.7;
// n=2048 21 vs 8.0; n=4096 83 vs 15; n=8192 300 vs 27. The crossover
// lies between 1024 and 2048; 1536 leans toward brute force.
var indexThreshold = 1536

const blockFanout = 16

// blockTree is a flat, append-only packed bounding tree over a numbered
// item sequence: level 0 holds one envelope per blockFanout consecutive
// items, level l+1 one per blockFanout level-l nodes. It stores only
// node bounds (about 2 bytes per item); item envelopes are re-derived
// by the caller, so the memory cost is a few KB, not a node-per-item
// tree. Bounds only ever grow.
type blockTree struct {
	levels [][]geom.Envelope
	n      int
}

// add appends an item with envelope e.
func (t *blockTree) add(e geom.Envelope) {
	i := t.n
	t.n++
	t.grow(i, e)
}

// grow extends the bounds covering item i to include e.
func (t *blockTree) grow(i int, e geom.Envelope) {
	for l := 0; ; l++ {
		i /= blockFanout
		if l == len(t.levels) {
			if l == 0 {
				t.levels = append(t.levels, nil)
			} else {
				// New root: seed it with the union of the level below.
				u := t.levels[l-1][0]
				for _, c := range t.levels[l-1][1:] {
					u = u.ExpandToInclude(c)
				}
				t.levels = append(t.levels, []geom.Envelope{u})
				return
			}
		}
		if i == len(t.levels[l]) {
			t.levels[l] = append(t.levels[l], e)
		} else {
			t.levels[l][i] = t.levels[l][i].ExpandToInclude(e)
		}
		if l == len(t.levels)-1 && len(t.levels[l]) == 1 {
			return
		}
	}
}

// query calls fn(i) for every item i in a block whose bounds intersect q
// (a superset of the items intersecting q); fn returning false stops.
func (t *blockTree) query(q geom.Envelope, fn func(i int) bool) bool {
	if t.n == 0 {
		return true
	}
	return t.queryNode(q, len(t.levels)-1, 0, fn)
}

func (t *blockTree) queryNode(q geom.Envelope, l, i int, fn func(int) bool) bool {
	if !t.levels[l][i].Intersects(q) {
		return true
	}
	lo := i * blockFanout
	if l == 0 {
		hi := min(lo+blockFanout, t.n)
		for k := lo; k < hi; k++ {
			if !fn(k) {
				return false
			}
		}
		return true
	}
	hi := min(lo+blockFanout, len(t.levels[l-1]))
	for c := lo; c < hi; c++ {
		if !t.queryNode(q, l-1, c, fn) {
			return false
		}
	}
	return true
}

// topoIndex holds the spatial indexes that prune the topology checks to
// envelope-overlapping candidates. Results are identical to the
// brute-force scans: pairs with disjoint envelopes can neither touch nor
// have a vertex inside the loop.
type topoIndex struct {
	all     []*taggedLine
	indexed bool        // false: brute-force scans, no indexes maintained
	verts   []blockTree // per line: vertices (built lazily)
	input   []blockTree // per line: input segments (built lazily; alive[] filters)
	output  []blockTree // per line: result segments, appended as emitted
}

func newTopoIndex(all []*taggedLine) *topoIndex {
	x := &topoIndex{all: all}
	n := 0
	for _, m := range all {
		n += len(m.pts)
	}
	x.indexed = n >= indexThreshold
	if x.indexed {
		x.output = make([]blockTree, len(all))
	}
	return x
}

func (x *topoIndex) vertIndex() []blockTree {
	if x.verts == nil {
		x.verts = make([]blockTree, len(x.all))
		for li, m := range x.all {
			t := &x.verts[li]
			for _, p := range m.pts {
				t.add(geom.Envelope{MinX: p.X, MinY: p.Y, MaxX: p.X, MaxY: p.Y})
			}
		}
	}
	return x.verts
}

func (x *topoIndex) inputIndex() []blockTree {
	if x.input == nil {
		x.input = make([]blockTree, len(x.all))
		for li, m := range x.all {
			t := &x.input[li]
			for k := 0; k+1 < len(m.pts); k++ {
				t.add(geom.SegmentEnvelope(m.pts[k], m.pts[k+1]))
			}
		}
	}
	return x.input
}

// addOutput records result segment i of line li (new, or replacing the
// segment at i).
func (x *topoIndex) addOutput(li, i int, a, b geom.XY) {
	if !x.indexed {
		return
	}
	t := &x.output[li]
	e := geom.SegmentEnvelope(a, b)
	if i == t.n {
		t.add(e)
	} else {
		t.grow(i, e)
	}
}

func hasOutputIntersection(x *topoIndex, a, b geom.XY) bool {
	if !x.indexed {
		for _, m := range x.all {
			for _, s := range m.result {
				if geomath.SegmentsCrossOrTouch(a, b, s[0], s[1]) {
					return true
				}
			}
		}
		return false
	}
	q := geom.SegmentEnvelope(a, b)
	for li, m := range x.all {
		if !m.hasResult() {
			continue
		}
		found := false
		x.output[li].query(q, func(i int) bool {
			if i >= len(m.result) {
				return true
			}
			s := m.result[i]
			if geomath.SegmentsCrossOrTouch(a, b, s[0], s[1]) {
				found = true
				return false
			}
			return true
		})
		if found {
			return true
		}
	}
	return false
}

// hasInputIntersection scans still-alive input segments; segments of ln
// inside [lo, hi) are the section being replaced and are excluded.
// Pass ln == nil to scan every alive segment (ring-endpoint variant).
func hasInputIntersection(x *topoIndex, ln *taggedLine, lo, hi int, a, b geom.XY) bool {
	if !x.indexed {
		for _, m := range x.all {
			for k := 0; k+1 < len(m.pts); k++ {
				if !m.alive[k] || (m == ln && k >= lo && k < hi) {
					continue
				}
				if geomath.SegmentsCrossOrTouch(a, b, m.pts[k], m.pts[k+1]) {
					return true
				}
			}
		}
		return false
	}
	q := geom.SegmentEnvelope(a, b)
	trees := x.inputIndex()
	for li, m := range x.all {
		found := false
		trees[li].query(q, func(k int) bool {
			if !m.alive[k] || (m == ln && k >= lo && k < hi) {
				return true
			}
			if geomath.SegmentsCrossOrTouch(a, b, m.pts[k], m.pts[k+1]) {
				found = true
				return false
			}
			return true
		})
		if found {
			return true
		}
	}
	return false
}

// hasJump reports whether any vertex of any line (outside the section
// being replaced) lies strictly inside the loop formed by the section
// path plus the flattening chord — flattening would flip that vertex's
// sidedness. Plays the role of JTS's ComponentJumpChecker. Only vertices
// inside the loop's envelope can be strictly inside the loop.
func hasJump(x *topoIndex, ln *taggedLine, lo, hi int, loop []geom.XY) bool {
	if !x.indexed {
		for _, m := range x.all {
			for k, p := range m.pts {
				if m == ln && k >= lo && k <= hi {
					continue
				}
				if pointStrictlyInLoop(p, loop) {
					return true
				}
			}
		}
		return false
	}
	q := geom.EnvelopeOfXY(loop)
	trees := x.vertIndex()
	for li, m := range x.all {
		found := false
		trees[li].query(q, func(k int) bool {
			if m == ln && k >= lo && k <= hi {
				return true
			}
			if p := m.pts[k]; q.ContainsXY(p) && pointStrictlyInLoop(p, loop) {
				found = true
				return false
			}
			return true
		})
		if found {
			return true
		}
	}
	return false
}
