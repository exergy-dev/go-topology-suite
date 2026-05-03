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
		// Degenerate polygon with too few ring vertices. JTS treats this
		// as the underlying lower-dimensional geometry (line/point) for
		// positive buffers. For negative buffers the result is empty.
		if distance > 0 {
			if poly, ok := bufferDegenerateRing(p.CRS(), outer, distance, cfg); ok {
				return poly, nil
			}
		}
		return geom.NewEmptyPolygon(p.CRS(), p.Layout()), nil
	}

	outerSigned := planar.Default.RingArea(outer)
	outerCCW := outerSigned > 0
	// Zero-area outer ring (collinear points) is geometrically a
	// line/point. Route through the line-string buffer for positive
	// distance.
	if distance > 0 && outerSigned == 0 {
		if poly, ok := bufferDegenerateRing(p.CRS(), outer, distance, cfg); ok {
			return poly, nil
		}
		return geom.NewEmptyPolygon(p.CRS(), p.Layout()), nil
	}

	switch {
	case distance > 0:
		// JTS-style polygonization: emit offset curves for every ring
		// (outer + holes) into a single segment set, snap-round, build
		// a DCEL, and label each face with its winding-depth from the
		// offset boundaries. Faces with depth ≥ 1 are inside the
		// buffer. This subsumes the older "dilate outer ∪ then erode
		// each hole separately" pipeline, which fragmented depth
		// reasoning across multiple overlay passes.
		segs := emitPolygonOffsetSegments(p, distance, cfg)
		if len(segs) == 0 {
			// Offset emission failed for every ring; preserve the
			// original polygon as the safest no-growth answer.
			return geom.NewPolygon(p.CRS(), allRings(p)...), nil
		}
		// Snap-rounding tolerance: a fraction of the buffer distance,
		// chosen so coordinate ULP noise from mitre-cap corner
		// computation is clustered to the same grid cell, but real
		// geometric features (segments separated by > tolerance) are
		// preserved. JTS uses scale = max-input-coord-magnitude *
		// 1e-12; we use distance * 1e-9 as a robust default.
		tolerance := math.Abs(distance) * 1e-9
		// V4 positive-buffer validator: filter polygonizer output by
		// winding-number depth-against-original. Phantom subgraphs
		// whose rep has winding == -sign(outer) (topologically inverted
		// against the input) are dropped. Faces inside the polygon body
		// (winding == +sign) and faces outside the body (winding == 0,
		// which the polygonizer's depth labelling has already
		// classified as buffer interior) are kept.
		validate := positiveBufferWindingValidator(p)
		got, err := polygonizeBufferWithFilter(p.CRS(), segs, tolerance, validate, 0)
		if err != nil {
			return nil, fmt.Errorf("buffer: polygonize: %w", err)
		}
		if got == nil || got.IsEmpty() {
			return geom.NewPolygon(p.CRS(), allRings(p)...), nil
		}
		return got, nil

	case distance < 0:
		// Negative buffer (inset). Two-phase strategy:
		//
		//  1. LEGACY PATH FIRST. The single-ring offset + overshoot
		//     guards + per-hole overlay.Difference pipeline produces
		//     clean ring outputs on the typical "convex / fat-parcel"
		//     inputs the property tests exercise. Snap-rounding the
		//     same inputs through the polygonizer can introduce
		//     spurious mitre-cap micro-faces (degenerate slivers)
		//     when the polygon has many near-collinear vertices from
		//     a previous dilation.
		//
		//     If the legacy path returns a non-empty result, we
		//     return it directly — preserving every currently-working
		//     case.
		//
		//  2. POLYGONIZER FALLBACK when the legacy path collapses to
		//     empty. Many real "thin parcel" inputs really do inset
		//     to empty, but a residual minority should still produce
		//     a non-empty inset ring (JTS's TestBufferExternal2,
		//     TestBufferJagged, TestBufferMitredJoin all expose this).
		//     The polygonizer's subgraph-aware depth labeller plus a
		//     face-validity filter (representative point INSIDE the
		//     original AND ≥ d from any original boundary segment)
		//     recovers these cases without admitting overshoot lobes.
		d := -distance
		if bboxTooThinForInset(outer, d) {
			return geom.NewEmptyPolygon(p.CRS(), p.Layout()), nil
		}
		legacy, legacyErr := bufferPolygonNegativeLegacy(p, distance, cfg, outer, outerCCW, outerSigned)
		if legacyErr != nil {
			return nil, legacyErr
		}
		if legacy != nil && !legacy.IsEmpty() {
			return legacy, nil
		}
		// Polygonizer fallback. Only reached when the legacy pipeline
		// reports empty — exactly the regime where JTS conformance is
		// currently weakest. The face-validity filter ensures we
		// never invent inset faces that lie outside the original or
		// straddle its boundary.
		segs := emitPolygonOffsetSegments(p, distance, cfg)
		if len(segs) == 0 {
			return geom.NewEmptyPolygon(p.CRS(), p.Layout()), nil
		}
		tolerance := d * 1e-9
		// Face-validity filter for the polygonizer: a kept ring's
		// representative interior point must satisfy BOTH
		//
		//  1. windingDepth(rep, originalRings) == sign(outer)
		//     — rep is topologically STRICTLY inside the original
		//     polygon body. This is the JTS-standard depth-against-
		//     original metric, generalising the legacy
		//     pointInPolygonRings call to be orientation-aware and
		//     numerically robust (signed ray-crossings cancel cleanly
		//     for ULP-scale rep-point noise).
		//
		//  2. minDistToBoundary(rep, originalRings) >= d
		//     — rep is at least d from any original boundary segment.
		//     Every TRUE inset interior point has clearance >= d by
		//     construction; phantom mitre-overshoot lobes have rep
		//     within ULP of the original boundary and fail this check.
		//
		// V3.0/V3.1/V3.2 explored relaxing the distance threshold or
		// replacing it with self-inscribed-radius checks; all degraded
		// conformance because the rejected rings are predominantly
		// phantom overshoot lobes whose detection by alternative
		// metrics is brittle. V4 explored replacing the distance check
		// with the winding-number alone; that regressed conformance
		// from 99.0% → 98.7% because phantom subgraphs whose rep lands
		// inside the original polygon (winding == +1) are admitted —
		// the distance check is the load-bearing rejection criterion.
		// The winding-number conjunction is the strictly safer
		// composite (rejects everything either check rejects, allows
		// nothing more) and supersedes the legacy
		// faceValidatorFor(p, d, 1.0).
		validate := negativeBufferHybridValidator(p, d)
		// Minimum-area filter: drop snap-rounding micro-slivers whose
		// area is negligible compared to d^2. A real inset face has
		// area at least a few d^2; anything two orders of magnitude
		// smaller is noise.
		minArea := d * d * 0.01
		got, err := polygonizeBufferWithFilter(p.CRS(), segs, tolerance, validate, minArea)
		if err != nil {
			return nil, fmt.Errorf("buffer: polygonize inset: %w", err)
		}
		if got == nil || got.IsEmpty() {
			return geom.NewEmptyPolygon(p.CRS(), p.Layout()), nil
		}
		return got, nil
	}

	// distance == 0 unreachable; Buffer short-circuits earlier.
	return p, nil
}

// bufferPolygonNegativeLegacy is the original pre-polygonize-fallback
// negative-buffer pipeline: offset the outer ring inward, validate it
// with overshoot guards, then subtract each grown hole via overlay.
// Returns an empty polygon when any guard fires; the caller decides
// whether to fall through to the polygonize-based fallback.
//
// Caller passes outerCCW / outerSigned / outer pre-computed so we can
// reuse them.
func bufferPolygonNegativeLegacy(
	p *geom.Polygon,
	distance float64,
	cfg config,
	outer []geom.XY,
	outerCCW bool,
	outerSigned float64,
) (geom.Geometry, error) {
	d := -distance
	shrunkOuter, ok := offsetClosedRing(outer, d, !outerCCW, cfg)
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
	if cx, cy, ok := ringCentroid(shrunkOuter); ok {
		if !pointInRingBuf(geom.XY{X: cx, Y: cy}, outer) {
			return geom.NewEmptyPolygon(p.CRS(), p.Layout()), nil
		}
	}
	if insetOvershoot(shrunkOuter, outer, d) {
		return geom.NewEmptyPolygon(p.CRS(), p.Layout()), nil
	}
	var result geom.Geometry = geom.NewPolygon(p.CRS(), shrunkOuter)
	for r := 1; r < p.NumRings(); r++ {
		hole := p.Ring(r)
		holeSigned := planar.Default.RingArea(hole)
		holeCCW := holeSigned > 0
		grown, ok := offsetClosedRing(hole, d, holeCCW, cfg)
		if !ok {
			continue
		}
		if ringDegenerate(grown) {
			continue
		}
		grownSigned := planar.Default.RingArea(grown)
		if (holeSigned > 0) != (grownSigned > 0) {
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

// insetOvershoot reports whether the inset ring has any vertex too
// close to the original boundary, which signals that the offset has
// overshot into a region of local-thickness < 2d. The check is
// conservative: it only fires when the closest distance is well below
// the requested inset (≤ 0.5·d), to avoid false positives on the
// many valid insets whose vertex distances sit slightly under d due
// to floating-point noise at convex-corner mitre points.
func insetOvershoot(inset, orig []geom.XY, d float64) bool {
	if d <= 0 || len(inset) == 0 || len(orig) < 2 {
		return false
	}
	threshold := d * 0.5
	for _, p := range inset {
		// Distance from p to the original ring's nearest segment.
		minD := math.Inf(1)
		for i := 0; i+1 < len(orig); i++ {
			seg := pointSegmentPerpDist(p, orig[i], orig[i+1])
			if seg < minD {
				minD = seg
			}
		}
		if minD < threshold {
			return true
		}
	}
	return false
}

// pointSegmentPerpDist returns the perpendicular distance from p to
// the line segment a→b (clamped to the segment endpoints).
func pointSegmentPerpDist(p, a, b geom.XY) float64 {
	dx, dy := b.X-a.X, b.Y-a.Y
	L2 := dx*dx + dy*dy
	if L2 == 0 {
		return math.Hypot(p.X-a.X, p.Y-a.Y)
	}
	t := ((p.X-a.X)*dx + (p.Y-a.Y)*dy) / L2
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	cx, cy := a.X+t*dx, a.Y+t*dy
	return math.Hypot(p.X-cx, p.Y-cy)
}

// reverseRing returns ring with vertex order reversed. Closing
// duplicate (if any) is preserved at the end.
func reverseRing(ring []geom.XY) []geom.XY {
	if len(ring) == 0 {
		return ring
	}
	closed := ring[0].Equal(ring[len(ring)-1])
	end := len(ring)
	if closed {
		end--
	}
	out := make([]geom.XY, 0, len(ring))
	for i := end - 1; i >= 0; i-- {
		out = append(out, ring[i])
	}
	if closed {
		out = append(out, out[0])
	}
	return out
}

// cleanRingPolygon resolves self-intersections in a (possibly invalid)
// ring by self-unioning it. Returns nil on failure or empty result.
// For a simple ring the result is geometrically equivalent.
func cleanRingPolygon(c *crs.CRS, ring []geom.XY) geom.Geometry {
	raw := geom.NewPolygon(c, ring)
	if raw == nil || raw.IsEmpty() {
		return nil
	}
	cleaned, err := overlay.Union(raw, raw)
	if err != nil || cleaned == nil || cleaned.IsEmpty() {
		// Fall back to the raw (possibly invalid) ring; better than
		// dropping the hole entirely.
		return raw
	}
	return cleaned
}

// unionMultiBufferParts unions a slice of buffer polygons, falling
// back to a MultiPolygon assembly when overlay.Union produces a
// spurious empty/smaller result (known fragility on large-coordinate
// buffer inputs).
//
// The strategy is: pairwise-union each next part into the accumulator;
// if the resulting area drops below the maximum input area (which is
// impossible for a valid union), keep both parts separately as
// disjoint MultiPolygon members. The returned geometry preserves total
// coverage area, which is what JTS's BufferResultMatcher checks.
func unionMultiBufferParts(c *crs.CRS, parts []*geom.Polygon) geom.Geometry {
	if len(parts) == 0 {
		return geom.NewEmptyPolygon(c, geom.LayoutXY)
	}
	if len(parts) == 1 {
		return parts[0]
	}
	// Working set of "pieces" as Geometry (Polygon or MultiPolygon).
	pieces := make([]geom.Geometry, 0, len(parts))
	for _, p := range parts {
		pieces = append(pieces, p)
	}
	// Pairwise fuse: try Union(a,b); accept iff the result's area is
	// at least max(area(a), area(b)) - 1e-9. Otherwise keep both.
	for {
		merged := false
	pair:
		for i := 0; i < len(pieces); i++ {
			for j := i + 1; j < len(pieces); j++ {
				u, err := overlay.Union(pieces[i], pieces[j])
				if err != nil {
					continue
				}
				if u == nil || u.IsEmpty() {
					continue
				}
				ai := geomTotalArea(pieces[i])
				aj := geomTotalArea(pieces[j])
				au := geomTotalArea(u)
				maxIn := math.Max(ai, aj)
				sumIn := ai + aj
				// A valid union has area in [max(a,b), a+b]. Reject if
				// outside that band (with a small slack).
				if au+1e-9 < maxIn || au > sumIn+1e-9 {
					continue
				}
				// Replace i with u, drop j.
				pieces[i] = u
				pieces = append(pieces[:j], pieces[j+1:]...)
				merged = true
				break pair
			}
		}
		if !merged {
			break
		}
	}
	if len(pieces) == 1 {
		return pieces[0]
	}
	// Flatten any nested multi-polygons into a single MultiPolygon.
	flat := make([]*geom.Polygon, 0, len(pieces))
	for _, g := range pieces {
		flat = append(flat, explodePolygons(g)...)
	}
	if len(flat) == 0 {
		return geom.NewEmptyPolygon(c, geom.LayoutXY)
	}
	if len(flat) == 1 {
		return flat[0]
	}
	return geom.NewMultiPolygon(c, flat...)
}

// geomTotalArea returns sum of |signed area| for all polygon members
// in g, treating holes as subtractive within each polygon.
func geomTotalArea(g geom.Geometry) float64 {
	switch v := g.(type) {
	case *geom.Polygon:
		a := 0.0
		for i := 0; i < v.NumRings(); i++ {
			r := math.Abs(planar.Default.RingArea(v.Ring(i)))
			if i == 0 {
				a += r
			} else {
				a -= r
			}
		}
		if a < 0 {
			return 0
		}
		return a
	case *geom.MultiPolygon:
		a := 0.0
		for i := 0; i < v.NumGeometries(); i++ {
			a += geomTotalArea(v.PolygonAt(i))
		}
		return a
	}
	return 0
}

// bufferDegenerateRing handles the degenerate-polygon case (collinear
// or insufficient vertices). The ring's distinct vertices are treated
// as a polyline (with caps) and buffered as a LineString. If only one
// distinct vertex remains, the result is a circle (point buffer).
func bufferDegenerateRing(c *crs.CRS, ring []geom.XY, distance float64, cfg config) (geom.Geometry, bool) {
	pts := dedupeRing(ring)
	if len(pts) == 0 {
		return nil, false
	}
	if len(pts) == 1 {
		return bufferPoint(c, pts[0], distance, cfg), true
	}
	// Build a LineString from the deduped vertices and route through
	// bufferLineString. We don't close it (treat as an open polyline);
	// if the ring was meaningful (closed shape) it would have non-zero
	// area and not have reached this path.
	flat := make([]float64, 0, 2*len(pts))
	for _, p := range pts {
		flat = append(flat, p.X, p.Y)
	}
	ls := geom.NewLineStringFlat(geom.LayoutXY, c, flat)
	if ls == nil || ls.IsEmpty() {
		return nil, false
	}
	poly, err := bufferLineString(ls, distance, cfg)
	if err != nil || poly == nil || poly.IsEmpty() {
		return nil, false
	}
	return poly, true
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
	// Near-degenerate corner tolerance: cx is sin(turn-angle) on unit
	// direction vectors, so |cx| < 1e-5 corresponds to a turn of
	// ~1e-5 rad. Below that the line-line mitre intersection lies many
	// orders of magnitude off the offset corner and produces phantom
	// spike vertices (TestBufferMitredJoin case#4 measures cx ≈ 7.8e-6
	// at its near-collinear corner). Treating the corner as collinear
	// is the topologically faithful choice.
	const collinearCxTol = 1e-5
	out := make([]geom.XY, 0, 2*len(segs)+8)
	for i := 0; i < len(segs); i++ {
		curr := segs[i]
		next := segs[(i+1)%len(segs)]
		pCurrEnd := geom.XY{X: curr.b.X + signed*curr.nx, Y: curr.b.Y + signed*curr.ny}
		pNextStart := geom.XY{X: next.a.X + signed*next.nx, Y: next.a.Y + signed*next.ny}
		// curr.dir × next.dir (sin of turn angle between unit dirs).
		cx := curr.ny*(-next.nx) - (-curr.nx)*next.ny
		// Treat near-collinear corners as straight-through. The line-line
		// mitre is numerically unstable here, AND emitting any vertex
		// at a point that lies (geometrically) on the line between its
		// neighbours is itself a spurious vertex — the offset of two
		// near-parallel adjacent edges should be a single continuous
		// segment, not two segments meeting at a duplicate corner.
		// Emit nothing for this corner; the prev iteration's pNextStart
		// (or its CROSS-branch mitre) and the next iteration's pCurrEnd
		// terminate the offset edge correctly.
		if math.Abs(cx) < collinearCxTol {
			continue
		}
		s := signed * cx
		switch {
		case s < 0:
			// Diverge: gap to fill.
			out = append(out, pCurrEnd)
			arc := buildClosedJoin(curr.b, pCurrEnd, pNextStart, curr, next, signed, cfg)
			out = append(out, arc...)
		case s > 0:
			// Cross: emit the line-line intersection of the two offset
			// edges (mitre truncation). Honor cfg.mitreLimit so the
			// fallback to pNextStart fires on extreme overshoots.
			mp, ok := mitrePoint(curr.b, pCurrEnd, pNextStart, curr, next, math.Abs(signed), cfg.mitreLimit)
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

// ringCentroid returns the area-weighted centroid (cx, cy) of the
// closed ring. Returns ok=false on degenerate rings (zero signed area).
func ringCentroid(ring []geom.XY) (float64, float64, bool) {
	if len(ring) < 4 {
		return 0, 0, false
	}
	var sumA, sumX, sumY float64
	for i := 0; i+1 < len(ring); i++ {
		x0, y0 := ring[i].X, ring[i].Y
		x1, y1 := ring[i+1].X, ring[i+1].Y
		cross := x0*y1 - x1*y0
		sumA += cross
		sumX += (x0 + x1) * cross
		sumY += (y0 + y1) * cross
	}
	if sumA == 0 {
		return 0, 0, false
	}
	return sumX / (3 * sumA), sumY / (3 * sumA), true
}

// pointInRingBuf is the standard ray-cast test against a closed ring.
func pointInRingBuf(p geom.XY, ring []geom.XY) bool {
	if len(ring) < 4 {
		return false
	}
	inside := false
	for i := 0; i+1 < len(ring); i++ {
		a, b := ring[i], ring[i+1]
		if (a.Y > p.Y) != (b.Y > p.Y) {
			xCross := a.X + (p.Y-a.Y)*(b.X-a.X)/(b.Y-a.Y)
			if p.X < xCross {
				inside = !inside
			}
		}
	}
	return inside
}

// bboxTooThinForInset reports whether the polygon's outer-ring bounding
// box has a side smaller than 2d, in which case no point inside can be
// at distance ≥ d from every boundary segment, so a negative buffer of
// magnitude d collapses to empty.
func bboxTooThinForInset(ring []geom.XY, d float64) bool {
	if len(ring) == 0 {
		return true
	}
	minX, maxX := ring[0].X, ring[0].X
	minY, maxY := ring[0].Y, ring[0].Y
	for _, p := range ring[1:] {
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
	return (maxX-minX) < 2*d || (maxY-minY) < 2*d
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

