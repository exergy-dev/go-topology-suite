package measure

import (
	"math"

	"github.com/exergy-dev/go-topology-suite/geom"
	"github.com/exergy-dev/go-topology-suite/index"
	"github.com/exergy-dev/go-topology-suite/internal/xybuf"
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

// geometryDistancePlanar is the branch-and-bound form of
// geometryDistanceRef. It returns ok=false when the operands need the
// reference path. Both compute the same minimum over the same candidate
// set (order-independent); pruning only discards facet pairs whose
// envelope gap, padded by a rounding margin, exceeds the current best, so
// they provably cannot lower it or cross.
func geometryDistancePlanar(a, b geom.Geometry, k planar.Kernel) (float64, bool) {
	bufp := xybuf.Borrow()
	defer xybuf.Release(bufp)
	envA, envB := geom.EmptyEnvelope(), geom.EmptyEnvelope()
	fa, ok := collectFacets(a, nil, &envA, bufp)
	if !ok {
		return 0, false
	}
	fb, ok := collectFacets(b, nil, &envB, bufp)
	if !ok {
		return 0, false
	}
	if len(fa) == 0 || len(fb) == 0 {
		return 0, true
	}
	// Index the larger side.
	if len(fa) < len(fb) {
		fa, fb = fb, fa
		envA, envB = envB, envA
	}
	items := make([]index.Item[pfacet], len(fa))
	for i, f := range fa {
		items[i] = index.Item[pfacet]{Env: facetEnv(f), Value: f}
	}
	tree := index.New[pfacet]()
	tree.Bulk(items)

	scale := 0.0
	for _, e := range [2]geom.Envelope{envA, envB} {
		scale = math.Max(scale, math.Max(math.Max(math.Abs(e.MinX), math.Abs(e.MaxX)), math.Max(math.Abs(e.MinY), math.Abs(e.MaxY))))
	}
	slack := 1e-9 * scale

	min := math.Inf(+1)
	cross := false

	// Seed with a few nearest-neighbour pairs so the search envelopes
	// start tight. Seed values are genuine candidates (pairMin results).
	step := len(fb)/16 + 1
	for i := 0; i < len(fb); i += step {
		q := fb[i]
		it, found := tree.Nearest(facetEnv(q), index.ItemDistanceFunc[pfacet](func(_ geom.Envelope, it index.Item[pfacet]) float64 {
			return envGap(facetEnv(q), it.Env)
		}))
		if !found {
			continue
		}
		var c bool
		if min, c = pairMin(k, q, it.Value, min); c {
			return 0, true
		}
	}
	if min == 0 {
		return 0, true
	}

	for _, q := range fb {
		qe := facetEnv(q)
		if envGap(qe, envA) > min+slack {
			continue
		}
		pad := min + slack
		if math.IsInf(pad, +1) {
			pad = math.MaxFloat64
		}
		search := geom.Envelope{MinX: qe.MinX - pad, MinY: qe.MinY - pad, MaxX: qe.MaxX + pad, MaxY: qe.MaxY + pad}
		tree.Search(search, func(it index.Item[pfacet]) bool {
			if envGap(qe, it.Env) > min+slack {
				return true
			}
			var c bool
			if min, c = pairMin(k, q, it.Value, min); c {
				cross = true
				return false
			}
			return true
		})
		if cross || min == 0 {
			return 0, true
		}
	}
	if math.IsInf(min, +1) {
		return 0, true
	}
	return min, true
}
