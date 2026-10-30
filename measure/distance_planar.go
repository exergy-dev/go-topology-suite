package measure

import (
	"cmp"
	"math"
	"slices"
	"sync"

	"github.com/exergy-dev/go-topology-suite/geom"
	"github.com/exergy-dev/go-topology-suite/kernel/planar"
)

// pfacet is one distance facet: a segment (a != b allowed equal) or an
// isolated vertex (seg=false, a == b).
type pfacet struct {
	a, b geom.XY
	seg  bool
}

// collectFacets appends every facet of g to out and widens env. It returns
// false when g has a degenerate linear component (fewer than two vertices)
// or non-finite coordinates, in which case the caller must use the
// reference path so vertex coverage stays identical.
func collectFacets(g geom.Geometry, out []pfacet, env *geom.Envelope, buf *[]geom.XY) ([]pfacet, bool) {
	ok := true
	addPath := func(pts []geom.XY) {
		if len(pts) < 2 {
			ok = false
			return
		}
		for i := 0; i+1 < len(pts); i++ {
			out = append(out, pfacet{a: pts[i], b: pts[i+1], seg: true})
		}
	}
	addPoint := func(p geom.XY) {
		out = append(out, pfacet{a: p, b: p})
	}
	var walk func(g geom.Geometry)
	walk = func(g geom.Geometry) {
		if !ok {
			return
		}
		switch v := g.(type) {
		case *geom.Point:
			if !v.IsEmpty() {
				addPoint(v.XY())
			}
		case *geom.MultiPoint:
			for i := 0; i < v.NumGeometries(); i++ {
				addPoint(v.PointAt(i))
			}
		case *geom.LineString:
			if v.IsEmpty() {
				return
			}
			pts := (*buf)[:0]
			for i := 0; i < v.NumPoints(); i++ {
				pts = append(pts, v.PointAt(i))
			}
			*buf = pts
			addPath(pts)
		case *geom.LinearRing:
			walk(v.AsLineString())
		case *geom.Polygon:
			for r := 0; r < v.NumRings(); r++ {
				*buf = v.RingInto((*buf)[:0], r)
				addPath(*buf)
			}
		case *geom.MultiLineString:
			for i := 0; i < v.NumGeometries(); i++ {
				walk(v.LineStringAt(i))
			}
		case *geom.MultiPolygon:
			for i := 0; i < v.NumGeometries(); i++ {
				walk(v.PolygonAt(i))
			}
		case *geom.GeometryCollection:
			for i := 0; i < v.NumGeometries(); i++ {
				walk(v.GeometryAt(i))
			}
		}
	}
	walk(g)
	if !ok {
		return nil, false
	}
	for _, f := range out {
		for _, p := range [2]geom.XY{f.a, f.b} {
			if math.IsNaN(p.X) || math.IsNaN(p.Y) || math.IsInf(p.X, 0) || math.IsInf(p.Y, 0) {
				return nil, false
			}
			*env = env.ExpandToInclude(geom.Envelope{MinX: p.X, MinY: p.Y, MaxX: p.X, MaxY: p.Y})
		}
	}
	return out, true
}

func facetEnv(f pfacet) geom.Envelope {
	return geom.SegmentEnvelope(f.a, f.b)
}

func envGap(a, b geom.Envelope) float64 {
	dx := math.Max(0, math.Max(a.MinX-b.MaxX, b.MinX-a.MaxX))
	dy := math.Max(0, math.Max(a.MinY-b.MaxY, b.MinY-a.MaxY))
	return math.Hypot(dx, dy)
}

// pairMin evaluates exactly the quantities the reference algorithm takes a
// minimum over for one facet pair: endpoint-to-segment distances in both
// directions, endpoint-to-endpoint distances, and (for two segments) the
// crossing test. cross reports a crossing.
func pairMin(k planar.Kernel, f, g pfacet, min float64) (m float64, cross bool) {
	if g.seg {
		if d := k.SegmentDistance(f.a, g.a, g.b); d < min {
			min = d
		}
		if f.seg {
			if d := k.SegmentDistance(f.b, g.a, g.b); d < min {
				min = d
			}
		}
	}
	if f.seg {
		if d := k.SegmentDistance(g.a, f.a, f.b); d < min {
			min = d
		}
		if g.seg {
			if d := k.SegmentDistance(g.b, f.a, f.b); d < min {
				min = d
			}
		}
	}
	if d := k.Distance(f.a, g.a); d < min {
		min = d
	}
	if g.seg {
		if d := k.Distance(f.a, g.b); d < min {
			min = d
		}
	}
	if f.seg {
		if d := k.Distance(f.b, g.a); d < min {
			min = d
		}
		if g.seg {
			if d := k.Distance(f.b, g.b); d < min {
				min = d
			}
		}
	}
	if f.seg && g.seg {
		if _, ok := k.SegmentIntersection(f.a, f.b, g.a, g.b); ok {
			return min, true
		}
	}
	return min, false
}

type planarScratch struct {
	fa, fb []pfacet
	pts    []geom.XY
}

const maxPooledFacets = 1 << 15

var planarScratchPool = sync.Pool{New: func() any { return new(planarScratch) }}

func facetMinX(f *pfacet) float64 { return math.Min(f.a.X, f.b.X) }
func facetMaxX(f *pfacet) float64 { return math.Max(f.a.X, f.b.X) }

// geometryDistancePlanar is the branch-and-bound form of
// distanceBruteForce. It returns ok=false when the operands need the
// reference path. Both compute the same minimum over the same candidate
// set (order-independent); pruning only discards facet pairs whose
// envelope gap, padded by a rounding margin, exceeds the current best, so
// they provably cannot lower it or cross.
//
// The larger operand's facets are sorted by minX into pooled scratch; each
// facet of the other operand scans only the x-slab that can lie within the
// current best distance, found by binary search. The scan is seeded with
// the query facet nearest the indexed envelope. Scratch is pooled, so
// steady-state calls allocate nothing.
func geometryDistancePlanar(a, b geom.Geometry, k planar.Kernel) (float64, bool) {
	sc := planarScratchPool.Get().(*planarScratch)
	defer func() {
		if cap(sc.fa) <= maxPooledFacets && cap(sc.fb) <= maxPooledFacets {
			planarScratchPool.Put(sc)
		}
	}()
	envA, envB := geom.EmptyEnvelope(), geom.EmptyEnvelope()
	fa, ok := collectFacets(a, sc.fa[:0], &envA, &sc.pts)
	sc.fa = fa
	if !ok {
		return 0, false
	}
	fb, ok := collectFacets(b, sc.fb[:0], &envB, &sc.pts)
	sc.fb = fb
	if !ok {
		return 0, false
	}
	if len(fa) == 0 || len(fb) == 0 {
		return 0, true
	}
	if len(fa) < len(fb) {
		fa, fb = fb, fa
		envA, envB = envB, envA
	}
	slices.SortFunc(fa, func(x, y pfacet) int {
		return cmp.Compare(facetMinX(&x), facetMinX(&y))
	})
	maxW := 0.0
	for i := range fa {
		if w := facetMaxX(&fa[i]) - facetMinX(&fa[i]); w > maxW {
			maxW = w
		}
	}

	scale := 0.0
	for _, e := range [2]geom.Envelope{envA, envB} {
		scale = math.Max(scale, math.Max(math.Max(math.Abs(e.MinX), math.Abs(e.MaxX)), math.Max(math.Abs(e.MinY), math.Abs(e.MaxY))))
	}
	slack := 1e-9 * scale

	min := math.Inf(+1)

	// scan evaluates query facet q against every indexed facet that may be
	// within the current best. Returns true on a crossing.
	scan := func(q *pfacet) bool {
		qe := facetEnv(*q)
		pad := min + slack
		if math.IsInf(pad, +1) {
			pad = math.MaxFloat64
		}
		lo, hi := qe.MinX-pad-maxW, qe.MaxX+pad
		// first index with minX >= lo
		l, h := 0, len(fa)
		for l < h {
			m := int(uint(l+h) >> 1)
			if facetMinX(&fa[m]) < lo {
				l = m + 1
			} else {
				h = m
			}
		}
		for i := l; i < len(fa); i++ {
			f := &fa[i]
			if facetMinX(f) > hi {
				break
			}
			if envGap(qe, facetEnv(*f)) > min+slack {
				continue
			}
			var c bool
			if min, c = pairMin(k, *q, *f, min); c {
				return true
			}
		}
		return false
	}

	// Seed: the query facet nearest the indexed envelope.
	seed, seedGap := 0, math.Inf(+1)
	for i := range fb {
		if g := envGap(facetEnv(fb[i]), envA); g < seedGap {
			seed, seedGap = i, g
		}
	}
	if scan(&fb[seed]) {
		return 0, true
	}
	for i := range fb {
		if min == 0 {
			return 0, true
		}
		if i == seed || envGap(facetEnv(fb[i]), envA) > min+slack {
			continue
		}
		if scan(&fb[i]) {
			return 0, true
		}
	}
	if min == 0 || math.IsInf(min, +1) {
		return 0, true
	}
	return min, true
}
