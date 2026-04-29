package buffer

import (
	"fmt"
	"math"

	"github.com/terra-geo/terra/crs"
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/kernel/planar"
	"github.com/terra-geo/terra/overlay"
)

// bufferPolygon implements positive/negative buffering of a single Polygon
// (with optional holes) on top of the overlay-NG path. Contract:
//
//   - distance > 0 ("dilation"): the polygon's solid material grows. The
//     outer ring is offset to its exterior and unioned with the original
//     outer; each hole is offset toward its own interior (the hole shrinks)
//     and subtracted from the dilated outer. Holes that collapse under the
//     offset are dropped.
//   - distance < 0 ("inset"): the polygon's solid material shrinks. The
//     outer ring is offset to its interior; each hole is offset to its
//     exterior (the hole grows into the polygon body) and subtracted from
//     the shrunk outer. If the outer collapses the result is empty.
//   - distance == 0: handled by the top-level Buffer; not reached here.
//
// Holes are now plumbed end-to-end (Pillar A5).
func bufferPolygon(p *geom.Polygon, distance float64, cfg config) (geom.Geometry, error) {
	if p.IsEmpty() || p.NumRings() == 0 {
		return geom.NewEmptyPolygon(p.CRS(), p.Layout()), nil
	}
	outer := p.Ring(0)
	if len(outer) < 4 {
		return geom.NewEmptyPolygon(p.CRS(), p.Layout()), nil
	}

	outerSigned := planar.Default.RingArea(outer)
	outerCCW := outerSigned > 0

	switch {
	case distance > 0:
		// 1. Build the dilated outer: union of the original outer with the
		//    exterior-offset of the outer. For convex shapes the offset
		//    already contains the original; for concave shapes the union
		//    fills in the reflex gaps.
		offsetOuter, ok := offsetClosedRing(outer, distance, outerCCW /*exterior*/, cfg)
		if !ok {
			// Offsetting failed; preserve the original polygon (with holes)
			// as the safest no-growth answer.
			return geom.NewPolygon(p.CRS(), allRings(p)...), nil
		}
		dilated, err := overlay.Union(
			geom.NewPolygon(p.CRS(), outer),
			geom.NewPolygon(p.CRS(), offsetOuter),
		)
		if err != nil {
			return nil, fmt.Errorf("buffer: union outer and offset: %w", err)
		}
		// 2. Subtract each shrunk hole (offset toward the hole's interior).
		for r := 1; r < p.NumRings(); r++ {
			hole := p.Ring(r)
			holeSigned := planar.Default.RingArea(hole)
			holeCCW := holeSigned > 0
			shrunk, ok := offsetClosedRing(hole, distance, !holeCCW /*interior*/, cfg)
			if !ok {
				continue
			}
			if ringDegenerate(shrunk) {
				continue
			}
			shrunkSigned := planar.Default.RingArea(shrunk)
			if (holeSigned > 0) != (shrunkSigned > 0) {
				// Hole collapsed past zero — fully erased by dilation.
				continue
			}
			if math.Abs(shrunkSigned) >= math.Abs(holeSigned) {
				// Inset failed to shrink the ring — mitre corners pushed
				// the offset outside the original. The hole is effectively
				// erased.
				continue
			}
			dilated, err = overlay.Difference(dilated, geom.NewPolygon(p.CRS(), shrunk))
			if err != nil {
				return nil, fmt.Errorf("buffer: subtract shrunk hole %d: %w", r-1, err)
			}
			if dilated.IsEmpty() {
				return dilated, nil
			}
		}
		return dilated, nil

	case distance < 0:
		// 1. Inset the outer ring toward its interior.
		d := -distance
		shrunkOuter, ok := offsetClosedRing(outer, d, !outerCCW /*interior*/, cfg)
		if !ok {
			return geom.NewEmptyPolygon(p.CRS(), p.Layout()), nil
		}
		if ringDegenerate(shrunkOuter) {
			return geom.NewEmptyPolygon(p.CRS(), p.Layout()), nil
		}
		shrunkSigned := planar.Default.RingArea(shrunkOuter)
		if (outerSigned > 0) != (shrunkSigned > 0) {
			return geom.NewEmptyPolygon(p.CRS(), p.Layout()), nil
		}
		var result geom.Geometry = geom.NewPolygon(p.CRS(), shrunkOuter)
		// 2. Grow each hole and subtract from the shrunk outer.
		for r := 1; r < p.NumRings(); r++ {
			hole := p.Ring(r)
			holeSigned := planar.Default.RingArea(hole)
			holeCCW := holeSigned > 0
			grown, ok := offsetClosedRing(hole, d, holeCCW /*exterior*/, cfg)
			if !ok {
				continue
			}
			if ringDegenerate(grown) {
				continue
			}
			grownSigned := planar.Default.RingArea(grown)
			if (holeSigned > 0) != (grownSigned > 0) {
				// Grown hole inverted — pathological; skip.
				continue
			}
			next, err := overlay.Difference(result, geom.NewPolygon(p.CRS(), grown))
			if err != nil {
				return nil, fmt.Errorf("buffer: subtract grown hole %d: %w", r-1, err)
			}
			result = next
			if result.IsEmpty() {
				return result, nil
			}
		}
		return result, nil
	}

	// distance == 0 unreachable; Buffer short-circuits earlier.
	return p, nil
}

// allRings returns every ring of p as [][]XY (outer first).
func allRings(p *geom.Polygon) [][]geom.XY {
	out := make([][]geom.XY, p.NumRings())
	for i := 0; i < p.NumRings(); i++ {
		out[i] = p.Ring(i)
	}
	return out
}

// bufferMultiPolygon buffers each member polygon and unions the results.
//
// For non-overlapping members the union is essentially a concatenation; for
// members whose buffers overlap (touching or near-touching parts) the union
// merges them into a single polygon, eliminating internal seams.
func bufferMultiPolygon(mp *geom.MultiPolygon, distance float64, cfg config) (geom.Geometry, error) {
	if mp.IsEmpty() {
		return geom.NewEmptyPolygon(mp.CRS(), mp.Layout()), nil
	}
	var acc geom.Geometry
	for i := 0; i < mp.NumGeometries(); i++ {
		part := mp.PolygonAt(i)
		buf, err := bufferPolygon(part, distance, cfg)
		if err != nil {
			return nil, err
		}
		if buf == nil || buf.IsEmpty() {
			continue
		}
		if acc == nil {
			acc = buf
			continue
		}
		acc, err = unionGeometries(mp.CRS(), acc, buf)
		if err != nil {
			return nil, err
		}
	}
	if acc == nil {
		return geom.NewEmptyPolygon(mp.CRS(), mp.Layout()), nil
	}
	return acc, nil
}

// unionGeometries unions two buffer results, each of which is either a
// Polygon or a MultiPolygon. It explodes both into Polygon parts and
// pairwise-unions them via overlay.Union, accumulating into a list. Disjoint
// pieces are kept as a MultiPolygon at the end.
//
// This is a v0.1 implementation: pairwise Union without a sweepline. For
// small multi-polygons (a handful of members) it is adequate.
func unionGeometries(c *crs.CRS, a, b geom.Geometry) (geom.Geometry, error) {
	parts := append(explodePolygons(a), explodePolygons(b)...)
	if len(parts) == 0 {
		return geom.NewEmptyPolygon(c, geom.LayoutXY), nil
	}
	// Repeatedly fuse any pair that overlap until no more fusions occur.
	merged := true
	for merged {
		merged = false
		for i := 0; i < len(parts); i++ {
			for j := i + 1; j < len(parts); j++ {
				u, err := overlay.Union(parts[i], parts[j])
				if err != nil {
					return nil, err
				}
				switch v := u.(type) {
				case *geom.Polygon:
					// They overlapped and merged into one polygon.
					parts[i] = v
					parts = append(parts[:j], parts[j+1:]...)
					merged = true
				case *geom.MultiPolygon:
					// Disjoint: leave them separate. (Union returns
					// MultiPolygon when the inputs don't intersect.)
					_ = v
				}
				if merged {
					break
				}
			}
			if merged {
				break
			}
		}
	}
	if len(parts) == 1 {
		return parts[0], nil
	}
	return geom.NewMultiPolygon(c, parts...), nil
}

// explodePolygons flattens g into a slice of individual *geom.Polygon
// parts (skipping empty ones).
func explodePolygons(g geom.Geometry) []*geom.Polygon {
	switch v := g.(type) {
	case *geom.Polygon:
		if v.IsEmpty() {
			return nil
		}
		return []*geom.Polygon{v}
	case *geom.MultiPolygon:
		out := make([]*geom.Polygon, 0, v.NumGeometries())
		for i := 0; i < v.NumGeometries(); i++ {
			pp := v.PolygonAt(i)
			if !pp.IsEmpty() {
				out = append(out, pp)
			}
		}
		return out
	}
	return nil
}

// offsetClosedRing builds a parallel offset of a closed ring at perpendicular
// distance d (>= 0). When outward is true the offset is on the opposite side
// of the interior; when false it is on the interior side.
//
// The implementation walks each segment in the original order, emits the
// offset endpoint, and handles the corner with the next segment using the
// configured join style. The ring wraps: the last segment joins the first.
// Caps are not used.
//
// Returns (ring, true) on success, ([], false) when the ring is too
// degenerate to offset (fewer than 3 distinct vertices).
func offsetClosedRing(ring []geom.XY, d float64, outward bool, cfg config) ([]geom.XY, bool) {
	pts := dedupeRing(ring)
	if len(pts) < 3 {
		return nil, false
	}
	// Build segments around the ring.
	n := len(pts)
	segs := make([]segment, 0, n)
	for i := 0; i < n; i++ {
		a := pts[i]
		b := pts[(i+1)%n]
		dx, dy := b.X-a.X, b.Y-a.Y
		L := math.Hypot(dx, dy)
		if L == 0 {
			continue
		}
		segs = append(segs, segment{a: a, b: b, nx: -dy / L, ny: dx / L})
	}
	if len(segs) < 3 {
		return nil, false
	}

	// Sign: positive d on the LEFT side (default). For outward offset on a
	// CCW ring, the outside is the RIGHT side ⇒ negate. The caller passed
	// outward=true exactly when we should put the offset on the right side.
	signed := d
	if outward {
		signed = -d
	}

	// Per-corner topology depends on whether the two adjacent offset edges
	// DIVERGE (gap that needs filling with mitre/round/bevel) or CROSS
	// (overlap that needs truncating to the line-line intersection).
	//
	// With our sign convention (signed > 0 = LEFT offset, signed < 0 =
	// RIGHT offset) and an original-edge cross product `cx`:
	//
	//   - signed * cx < 0  ⇒  offsets DIVERGE — emit pCurrEnd, [join arc],
	//     pNextStart.
	//   - signed * cx > 0  ⇒  offsets CROSS — emit just the line-line
	//     intersection point (mitre truncation), skipping pCurrEnd /
	//     pNextStart.
	//
	// Concretely: outward offset on a convex original corner diverges; the
	// inward offset of the same corner crosses; concave (reflex) corners
	// flip both.
	out := make([]geom.XY, 0, 2*len(segs)+8)
	for i := 0; i < len(segs); i++ {
		curr := segs[i]
		next := segs[(i+1)%len(segs)]
		pCurrEnd := geom.XY{X: curr.b.X + signed*curr.nx, Y: curr.b.Y + signed*curr.ny}
		pNextStart := geom.XY{X: next.a.X + signed*next.nx, Y: next.a.Y + signed*next.ny}
		// curr.dir × next.dir
		cx := curr.ny*(-next.nx) - (-curr.nx)*next.ny
		s := signed * cx
		switch {
		case s < 0:
			// Diverge: gap to fill.
			out = append(out, pCurrEnd)
			arc := buildClosedJoin(curr.b, pCurrEnd, pNextStart, curr, next, signed, cfg)
			out = append(out, arc...)
		case s > 0:
			// Cross: emit the line-line intersection of the two offset
			// edges (mitre truncation). Falls back to pNextStart if the
			// lines are parallel.
			mp, ok := mitrePoint(curr.b, pCurrEnd, pNextStart, curr, next, math.Abs(signed), math.Inf(1))
			if ok {
				out = append(out, mp)
			} else {
				out = append(out, pNextStart)
			}
		default:
			// Collinear / zero turn — pCurrEnd ≈ pNextStart.
			out = append(out, pCurrEnd)
		}
	}
	if len(out) == 0 {
		return nil, false
	}
	out = append(out, out[0])
	return out, true
}

// buildClosedJoin is the closed-ring analogue of buildJoinArc: it returns
// the interior vertices of the convex corner (pNextStart included as the
// last vertex). signed carries the side: +d = left offset, -d = right
// offset. The geometry needs to flip when offsetting on the right side.
func buildClosedJoin(vertex, pCurrEnd, pNextStart geom.XY, curr, next segment, signed float64, cfg config) []geom.XY {
	switch cfg.join {
	case JoinBevel:
		return []geom.XY{pNextStart}
	case JoinMitre:
		mp, ok := mitrePoint(vertex, pCurrEnd, pNextStart, curr, next, math.Abs(signed), cfg.mitreLimit)
		if !ok {
			return []geom.XY{pNextStart}
		}
		return []geom.XY{mp, pNextStart}
	case JoinRound:
		return roundArc(vertex, pCurrEnd, pNextStart, math.Abs(signed), cfg.quadSegments)
	}
	return []geom.XY{pNextStart}
}

// dedupeRing returns the ring's distinct vertices in order, with the
// trailing closing duplicate removed.
func dedupeRing(ring []geom.XY) []geom.XY {
	if len(ring) == 0 {
		return nil
	}
	// Drop the closing duplicate if present.
	end := len(ring)
	if ring[0].Equal(ring[end-1]) {
		end--
	}
	out := make([]geom.XY, 0, end)
	for i := 0; i < end; i++ {
		p := ring[i]
		if len(out) > 0 && out[len(out)-1].Equal(p) {
			continue
		}
		out = append(out, p)
	}
	return out
}

// ringDegenerate reports whether ring has effectively zero area (bounding
// box smaller than a tiny epsilon).
func ringDegenerate(ring []geom.XY) bool {
	if len(ring) < 4 {
		return true
	}
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)
	for _, p := range ring {
		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	const eps = 1e-12
	return (maxX-minX) < eps || (maxY-minY) < eps
}

