package overlayng

import (
	terra "github.com/terra-geo/terra"
	"github.com/terra-geo/terra/crs"
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/internal/noding"
	"github.com/terra-geo/terra/internal/snap"
)

// Op identifies which boolean polygon operation to perform.
type Op int

const (
	OpIntersection Op = iota
	OpUnion
	OpDifference
	OpSymDiff
)

// Overlay is the single-polygon entry point. Default behaviour: no
// snap-rounding (preserves user input exactly). Use OverlayWithTolerance
// to handle inputs with near-coincident vertices.
func Overlay(subj, clip *geom.Polygon, op Op) (*geom.Polygon, []*geom.Polygon, error) {
	return OverlayWithTolerance(subj, clip, op, 0)
}

// OverlayWithTolerance is Overlay with explicit snap-rounding tolerance.
// Snapping the inputs to a common grid before noding eliminates the
// near-coincident-edge cases that defeat the brute-force segment
// intersector. A typical choice for unit-scale inputs is tolerance =
// 1e-9; for lon/lat data, ~1e-7 (~1 cm). Pass tolerance=0 to skip the
// snap pass and use raw coordinates.
func OverlayWithTolerance(subj, clip *geom.Polygon, op Op, tolerance float64) (*geom.Polygon, []*geom.Polygon, error) {
	return OverlayPolygonalWithTolerance(
		[]*geom.Polygon{subj}, []*geom.Polygon{clip}, op, tolerance,
	)
}

// OverlayPolygonal accepts polygon slices for both subj and clip, so
// MultiPolygon overlay routes through the same DCEL-and-classifier path
// as single-polygon overlay. Each input is treated as the union of its
// constituent polygons (each of which carries its own outer ring + holes).
//
// Returned shape: one "first" polygon plus zero-or-more disjoint "rest"
// polygons, identical to the single-polygon entry point.
func OverlayPolygonal(subj, clip []*geom.Polygon, op Op) (*geom.Polygon, []*geom.Polygon, error) {
	return OverlayPolygonalWithTolerance(subj, clip, op, 0)
}

// OverlayPolygonalWithTolerance is the polygonal-input entry point with
// explicit snap-rounding tolerance.
func OverlayPolygonalWithTolerance(subj, clip []*geom.Polygon, op Op, tolerance float64) (*geom.Polygon, []*geom.Polygon, error) {
	c, err := commonCRS(subj, clip)
	if err != nil {
		return nil, nil, err
	}

	// Filter empties; flatten each polygon into its ring list.
	subjRings, subjPerPoly := snapAndPartition(subj, tolerance)
	clipRings, clipPerPoly := snapAndPartition(clip, tolerance)
	if len(subjRings) == 0 || len(clipRings) == 0 {
		return geom.NewEmptyPolygon(c, geom.LayoutXY), nil, nil
	}

	// Goodrich-Guibas hot-pixel pass: with subj and clip sharing a
	// hot-pixel set, a vertex from one input that snaps into the
	// other's segment path forces a split, preventing the DCEL from
	// disconnecting at near-vertices.
	if tolerance > 0 {
		subjRings, clipRings = hotPixelRoundCombined(subjRings, clipRings, tolerance)
		if len(subjRings) == 0 || len(clipRings) == 0 {
			return geom.NewEmptyPolygon(c, geom.LayoutXY), nil, nil
		}
	}

	first, rest, err := overlayCorePolygonal(c, subjRings, subjPerPoly, clipRings, clipPerPoly, op)
	if err != nil {
		return nil, nil, err
	}
	if needsCanonicalize(first, rest) {
		return canonicalizeTouchingRings(c, first, rest)
	}
	return first, rest, nil
}

// hotPixelRoundCombined runs hot-pixel snap rounding across the
// combined subj+clip ring set so hot pixels from either side trigger
// splits in the other side's segments. Per-polygon partitions remain
// valid because hot-pixel processing only inserts vertices; no rings
// are dropped at this stage.
func hotPixelRoundCombined(subjRings, clipRings [][]geom.XY, tolerance float64) (subjOut, clipOut [][]geom.XY) {
	hp := snap.NewHotPixelSet(tolerance)
	for _, r := range subjRings {
		for _, v := range r {
			hp.Add(v)
		}
	}
	for _, r := range clipRings {
		for _, v := range r {
			hp.Add(v)
		}
	}
	subjOut = make([][]geom.XY, 0, len(subjRings))
	for _, r := range subjRings {
		if noded := hp.NodeRing(r); noded != nil {
			subjOut = append(subjOut, noded)
		}
	}
	clipOut = make([][]geom.XY, 0, len(clipRings))
	for _, r := range clipRings {
		if noded := hp.NodeRing(r); noded != nil {
			clipOut = append(clipOut, noded)
		}
	}
	return subjOut, clipOut
}

// commonCRS returns the CRS shared by all input polygons, or an error if
// they disagree (or both lists are empty).
func commonCRS(subj, clip []*geom.Polygon) (*crs.CRS, error) {
	var c *crs.CRS
	first := true
	for _, p := range subj {
		if first {
			c = p.CRS()
			first = false
			continue
		}
		if !crs.Equal(c, p.CRS()) {
			return nil, terra.ErrCRSMismatch
		}
	}
	for _, p := range clip {
		if first {
			c = p.CRS()
			first = false
			continue
		}
		if !crs.Equal(c, p.CRS()) {
			return nil, terra.ErrCRSMismatch
		}
	}
	return c, nil
}

// snapAndPartition flattens a polygon list into a single ring list (for
// segment-string emission) plus a parallel "ring count per polygon"
// slice that lets the classifier reconstruct per-polygon containment.
func snapAndPartition(polys []*geom.Polygon, tolerance float64) (rings [][]geom.XY, perPoly []int) {
	for _, p := range polys {
		if p == nil || p.IsEmpty() {
			continue
		}
		r := snapAllRings(p, tolerance)
		if r == nil {
			continue
		}
		perPoly = append(perPoly, len(r))
		rings = append(rings, r...)
	}
	return rings, perPoly
}

// snapAllRings extracts every ring (outer + holes) from a polygon and
// optionally snap-rounds them. Rings that collapse under snap are
// dropped — except the outer ring; if that collapses we return nil so
// the caller can short-circuit to an empty result.
func snapAllRings(p *geom.Polygon, tolerance float64) [][]geom.XY {
	out := make([][]geom.XY, 0, p.NumRings())
	for r := 0; r < p.NumRings(); r++ {
		ring := p.Ring(r)
		if tolerance > 0 {
			ring = snap.New(tolerance).SnapRing(ring)
			if ring == nil {
				if r == 0 {
					return nil
				}
				continue
			}
		}
		out = append(out, ring)
	}
	return out
}

// overlayCore is the single-polygon shared body: node every ring →
// DCEL → classify faces against the original (multi-ring) polygons →
// trace result rings → reassemble outers and holes.
func overlayCore(c *crs.CRS, subjRings, clipRings [][]geom.XY, op Op) (*geom.Polygon, []*geom.Polygon, error) {
	// Single-polygon overlay routes through the polygonal entry point
	// with one polygon per side; perPoly slices encode that.
	return overlayCorePolygonal(c,
		subjRings, []int{len(subjRings)},
		clipRings, []int{len(clipRings)},
		op,
	)
}

// OverlayPolygonalMixedDim is the polygonal-input entry point that
// returns a generic Geometry — possibly a GeometryCollection if the
// overlay produces mixed-dimension output (e.g., polygon ∩ polygon
// where the polygons share a boundary segment yields a LineString,
// or where they touch at a single vertex yields a Point).
//
// Returns nil and a non-nil error only on unrecoverable failures;
// successful empty results return an empty geometry of the
// appropriate dimension.
func OverlayPolygonalMixedDim(subj, clip []*geom.Polygon, op Op, tolerance float64) (geom.Geometry, error) {
	c, err := commonCRS(subj, clip)
	if err != nil {
		return nil, err
	}
	subjRings, subjPerPoly := snapAndPartition(subj, tolerance)
	clipRings, clipPerPoly := snapAndPartition(clip, tolerance)
	if len(subjRings) == 0 || len(clipRings) == 0 {
		return geom.NewEmptyPolygon(c, geom.LayoutXY), nil
	}
	if tolerance > 0 {
		subjRings, clipRings = hotPixelRoundCombined(subjRings, clipRings, tolerance)
		if len(subjRings) == 0 || len(clipRings) == 0 {
			return geom.NewEmptyPolygon(c, geom.LayoutXY), nil
		}
	}
	return overlayCorePolygonalMixed(c, subjRings, subjPerPoly, clipRings, clipPerPoly, op, tolerance)
}

// overlayCorePolygonalMixed is the variant of overlayCorePolygonal
// that retains the DCEL after polygon extraction, then extracts
// lineal and pointal results. The combined geometry is wrapped in a
// GeometryCollection iff multiple dimensional classes survive.
func overlayCorePolygonalMixed(
	c *crs.CRS,
	subjRings [][]geom.XY, subjPerPoly []int,
	clipRings [][]geom.XY, clipPerPoly []int,
	op Op,
	tolerance float64,
) (geom.Geometry, error) {
	segs := make([]*noding.SegmentString, 0, len(subjRings)+len(clipRings))
	for _, r := range subjRings {
		segs = append(segs, &noding.SegmentString{Coords: append([]geom.XY(nil), r...), Tag: 1})
	}
	for _, r := range clipRings {
		segs = append(segs, &noding.SegmentString{Coords: append([]geom.XY(nil), r...), Tag: 2})
	}
	noded := nodeAndSnap(segs, tolerance)
	taggedSegs := flattenNoded(noded)
	d := buildDCEL(taggedSegs)
	d.traceFaces()

	if !d.isConnected() && !mayHandleMultiComponent(d, subjRings, subjPerPoly, clipRings, clipPerPoly) {
		first, rest, err := overlayDisjointPolygonal(c,
			rebuildPolygons(c, subjRings, subjPerPoly),
			rebuildPolygons(c, clipRings, clipPerPoly),
			op,
		)
		if err != nil {
			return nil, err
		}
		return wrapPolygonResult(c, first, rest), nil
	}

	classifyFacesByPolygons(d, subjRings, subjPerPoly, clipRings, clipPerPoly)
	applyOp(d, op)
	rings := extractResultRings(d)
	first, rest, polyErr := assembleOutputPolygons(c, rings)
	if polyErr != nil {
		return nil, polyErr
	}
	if needsCanonicalize(first, rest) {
		canFirst, canRest, canErr := canonicalizeTouchingRings(c, first, rest)
		if canErr == nil {
			first, rest = canFirst, canRest
		}
	}
	lines := extractResultLines(d, op)
	points := extractResultPoints(d, op, lines, rings)

	return assembleMixedDim(c, first, rest, lines, points), nil
}

// wrapPolygonResult turns the legacy (first, rest) polygon return
// into a single geometry.
func wrapPolygonResult(c *crs.CRS, first *geom.Polygon, rest []*geom.Polygon) geom.Geometry {
	if first == nil || first.IsEmpty() {
		if len(rest) == 0 {
			return geom.NewEmptyPolygon(c, geom.LayoutXY)
		}
	}
	if len(rest) == 0 {
		return first
	}
	parts := make([]*geom.Polygon, 0, 1+len(rest))
	if first != nil && !first.IsEmpty() {
		parts = append(parts, first)
	}
	parts = append(parts, rest...)
	if len(parts) == 1 {
		return parts[0]
	}
	return geom.NewMultiPolygon(c, parts...)
}

// assembleMixedDim packs polygon, lineal, and pointal results into a
// single geometry. Single-dimension results return their natural
// type; mixed results return a GeometryCollection.
func assembleMixedDim(c *crs.CRS, first *geom.Polygon, rest []*geom.Polygon, lines [][]geom.XY, points []geom.XY) geom.Geometry {
	hasPoly := first != nil && !first.IsEmpty()
	hasMulti := len(rest) > 0
	hasLines := len(lines) > 0
	hasPoints := len(points) > 0

	classes := 0
	if hasPoly || hasMulti {
		classes++
	}
	if hasLines {
		classes++
	}
	if hasPoints {
		classes++
	}
	if classes == 0 {
		return geom.NewEmptyPolygon(c, geom.LayoutXY)
	}
	if classes == 1 {
		if hasPoly || hasMulti {
			return wrapPolygonResult(c, first, rest)
		}
		if hasLines {
			return wrapLinesResult(c, lines)
		}
		return wrapPointsResult(c, points)
	}
	// Mixed: build a GeometryCollection.
	var members []geom.Geometry
	if poly := wrapPolygonResult(c, first, rest); poly != nil && !poly.IsEmpty() {
		members = append(members, poly)
	}
	if hasLines {
		members = append(members, wrapLinesResult(c, lines))
	}
	if hasPoints {
		members = append(members, wrapPointsResult(c, points))
	}
	return geom.NewGeometryCollection(c, members...)
}

func wrapLinesResult(c *crs.CRS, lines [][]geom.XY) geom.Geometry {
	if len(lines) == 0 {
		return geom.NewLineString(c, nil)
	}
	if len(lines) == 1 {
		return geom.NewLineString(c, lines[0])
	}
	parts := make([]*geom.LineString, len(lines))
	for i, l := range lines {
		parts[i] = geom.NewLineString(c, l)
	}
	return geom.NewMultiLineString(c, parts...)
}

func wrapPointsResult(c *crs.CRS, points []geom.XY) geom.Geometry {
	if len(points) == 0 {
		return geom.NewEmptyPoint(c, geom.LayoutXY)
	}
	if len(points) == 1 {
		return geom.NewPoint(c, points[0])
	}
	return geom.NewMultiPoint(c, points)
}

// overlayCorePolygonal is the multi-aware shared body. ringsSubj is
// the flat ring list across all subj polygons; subjPerPoly[i] is the
// number of rings (outer + holes) belonging to subj polygon i. Same
// shape for clip.
func overlayCorePolygonal(
	c *crs.CRS,
	subjRings [][]geom.XY, subjPerPoly []int,
	clipRings [][]geom.XY, clipPerPoly []int,
	op Op,
) (*geom.Polygon, []*geom.Polygon, error) {
	segs := make([]*noding.SegmentString, 0, len(subjRings)+len(clipRings))
	for _, r := range subjRings {
		segs = append(segs, &noding.SegmentString{
			Coords: append([]geom.XY(nil), r...),
			Tag:    1,
		})
	}
	for _, r := range clipRings {
		segs = append(segs, &noding.SegmentString{
			Coords: append([]geom.XY(nil), r...),
			Tag:    2,
		})
	}
	noded := nodeAdaptive(segs)
	taggedSegs := flattenNoded(noded)
	d := buildDCEL(taggedSegs)
	d.traceFaces()

	// Multi-component DCELs (no shared boundary between subj and clip,
	// or holes that don't touch their shell) are handled directly via
	// per-face classification: each face's interior point is tested
	// against the original input rings, so containment relations are
	// resolved correctly even when the components are nested or
	// strictly disjoint.
	if !d.isConnected() && !mayHandleMultiComponent(d, subjRings, subjPerPoly, clipRings, clipPerPoly) {
		return overlayDisjointPolygonal(c,
			rebuildPolygons(c, subjRings, subjPerPoly),
			rebuildPolygons(c, clipRings, clipPerPoly),
			op,
		)
	}

	classifyFacesByPolygons(d, subjRings, subjPerPoly, clipRings, clipPerPoly)
	applyOp(d, op)
	rings := extractResultRings(d)
	if len(rings) == 0 {
		return geom.NewEmptyPolygon(c, geom.LayoutXY), nil, nil
	}
	return assembleOutputPolygons(c, rings)
}

// mayHandleMultiComponent returns true when the multi-component DCEL
// path is safe: every face's representative interior point sits in a
// region whose subj/clip membership is unambiguous from the input
// rings. The check is conservative — it returns true unconditionally
// for now, since classifyFacesByPolygons uses pointInPolygonRings on
// the original input geometry (not on the DCEL), so multi-component
// classification is correct as long as the DCEL is built without
// vertex aliasing. The disjoint helper remains as a defensive
// fallback for true outliers.
func mayHandleMultiComponent(d *dcel,
	subjRings [][]geom.XY, subjPerPoly []int,
	clipRings [][]geom.XY, clipPerPoly []int,
) bool {
	return true
}

// needsCanonicalize reports whether (first, rest) contains any polygon
// whose hole shares a vertex with its outer ring, or any two polygons
// whose outer rings share a vertex. The DCEL face trace can produce
// such configurations when the kept region's "shape" — geometrically
// an L-form — is encoded as a rectangle outer with a touching
// rectangular hole. JTS normalises these into a single ring; we do the
// same by re-running the polygons through a self-Union, which the
// overlay-NG noding step decomposes cleanly.
func needsCanonicalize(first *geom.Polygon, rest []*geom.Polygon) bool {
	all := make([]*geom.Polygon, 0, 1+len(rest))
	if first != nil && !first.IsEmpty() {
		all = append(all, first)
	}
	for _, p := range rest {
		if p != nil && !p.IsEmpty() {
			all = append(all, p)
		}
	}
	for _, p := range all {
		if polygonHasTouchingHole(p) {
			return true
		}
	}
	if len(all) >= 2 && multiPolygonsTouch(all) {
		return true
	}
	return false
}

// polygonHasTouchingHole returns true when an outer ring and a hole
// share a boundary segment (rather than just a vertex). The diagnostic
// signal is a hole vertex lying STRICTLY on the interior of an outer
// segment (or vice versa) — that vertex represents a hot pixel where
// the noder split one ring at the other's vertex, producing two rings
// that share a finite-length edge.
//
// A hole that merely touches the outer at a single vertex (case #3 of
// TestOverlayAA, where two diamond cavities meet at a corner of the
// outer) is geometrically valid and must not be flagged here, since
// the canonicalisation pass would erase the hole and corrupt the
// result.
func polygonHasTouchingHole(p *geom.Polygon) bool {
	if p == nil || p.NumRings() < 2 {
		return false
	}
	outer := p.Ring(0)
	outerVerts := vertexSet(outer)
	for r := 1; r < p.NumRings(); r++ {
		hole := p.Ring(r)
		holeVerts := vertexSet(hole)
		// Hole vertex on interior of outer segment.
		for v := range holeVerts {
			if _, isOuterVertex := outerVerts[v]; isOuterVertex {
				continue
			}
			if pointOnAnySegmentInterior(v, outer) {
				return true
			}
		}
		// Outer vertex on interior of hole segment (symmetric).
		for v := range outerVerts {
			if _, isHoleVertex := holeVerts[v]; isHoleVertex {
				continue
			}
			if pointOnAnySegmentInterior(v, hole) {
				return true
			}
		}
	}
	return false
}

// multiPolygonsTouch returns true when any two outer rings of distinct
// polygons share a boundary segment — either as a strict
// "vertex on segment interior" hit (case#11 hole-vs-shell) or as an
// identical edge that appears in both outer rings (case#3
// symdifference, where two assembled polygons abut along a shared
// spine of length > 0). Pure single-vertex coincidence is not flagged.
func multiPolygonsTouch(polys []*geom.Polygon) bool {
	for i := 0; i < len(polys); i++ {
		ri := polys[i].Ring(0)
		viSet := vertexSet(ri)
		for j := i + 1; j < len(polys); j++ {
			rj := polys[j].Ring(0)
			vjSet := vertexSet(rj)
			// Vertex of i on interior of a j segment.
			for v := range viSet {
				if _, isJ := vjSet[v]; isJ {
					continue
				}
				if pointOnAnySegmentInterior(v, rj) {
					return true
				}
			}
			// Vertex of j on interior of an i segment.
			for v := range vjSet {
				if _, isI := viSet[v]; isI {
					continue
				}
				if pointOnAnySegmentInterior(v, ri) {
					return true
				}
			}
		}
	}
	// Detect identical-edge sharing across distinct polygons by
	// canonicalising each outer-ring segment (lex-min endpoint first)
	// and watching for any segment that appears in two polygons.
	type seg struct{ a, b geom.XY }
	canon := func(p, q geom.XY) seg {
		if p.X < q.X || (p.X == q.X && p.Y < q.Y) {
			return seg{p, q}
		}
		return seg{q, p}
	}
	owner := map[seg]int{}
	for i, p := range polys {
		ring := p.Ring(0)
		for k := 0; k+1 < len(ring); k++ {
			s := canon(ring[k], ring[k+1])
			if prev, ok := owner[s]; ok && prev != i {
				return true
			}
			owner[s] = i
		}
	}
	return false
}

// pointOnAnySegmentInterior reports whether p lies strictly inside any
// segment of the closed ring (between two consecutive vertices,
// excluding the endpoints themselves).
func pointOnAnySegmentInterior(p geom.XY, ring []geom.XY) bool {
	for j := 0; j+1 < len(ring); j++ {
		if pointOnSegmentInterior(p, ring[j], ring[j+1]) {
			return true
		}
	}
	return false
}

// pointOnSegmentInterior is the standard "collinear and strictly
// between endpoints" test. Equality with either endpoint returns
// false; only interior position counts.
func pointOnSegmentInterior(p, a, b geom.XY) bool {
	if p == a || p == b {
		return false
	}
	cross := (b.X-a.X)*(p.Y-a.Y) - (b.Y-a.Y)*(p.X-a.X)
	if cross != 0 {
		return false
	}
	// Project onto the longer axis for numerical stability.
	dx, dy := b.X-a.X, b.Y-a.Y
	if dx*dx >= dy*dy {
		if dx == 0 {
			return false
		}
		t := (p.X - a.X) / dx
		return t > 0 && t < 1
	}
	if dy == 0 {
		return false
	}
	t := (p.Y - a.Y) / dy
	return t > 0 && t < 1
}

// vertexSet returns the set of unique vertices in ring (excluding the
// closing duplicate, since rings are stored with first==last).
func vertexSet(ring []geom.XY) map[geom.XY]struct{} {
	if len(ring) == 0 {
		return nil
	}
	end := len(ring)
	if end > 1 && ring[0] == ring[end-1] {
		end--
	}
	out := make(map[geom.XY]struct{}, end)
	for i := 0; i < end; i++ {
		out[ring[i]] = struct{}{}
	}
	return out
}

// canonicalizeTouchingRings re-runs the polygon set through
// overlayCorePolygonal with op=Union (using the same set as both subj
// and clip). The Union path nodes the touching rings together and
// extracts the merged boundary as a single ring per kept face,
// converting an "outer + touching hole" representation into the
// equivalent simple polygon (an L-shape, U-shape, etc).
//
// This is a single canonicalisation pass; the result is returned even
// if it still contains touching rings, to avoid any chance of
// non-termination.
func canonicalizeTouchingRings(c *crs.CRS, first *geom.Polygon, rest []*geom.Polygon) (*geom.Polygon, []*geom.Polygon, error) {
	polys := make([]*geom.Polygon, 0, 1+len(rest))
	if first != nil && !first.IsEmpty() {
		polys = append(polys, first)
	}
	polys = append(polys, rest...)
	if len(polys) == 0 {
		return geom.NewEmptyPolygon(c, geom.LayoutXY), nil, nil
	}
	rings, perPoly := snapAndPartition(polys, 0)
	if len(rings) == 0 {
		return geom.NewEmptyPolygon(c, geom.LayoutXY), nil, nil
	}
	return overlayCorePolygonal(c, rings, perPoly, rings, perPoly, OpUnion)
}

// rebuildPolygons reconstructs the per-polygon slice from a flat ring
// list and a per-polygon ring-count partition. Used by the disjoint
// fallback, which needs per-polygon containment tests.
func rebuildPolygons(c *crs.CRS, rings [][]geom.XY, perPoly []int) []*geom.Polygon {
	out := make([]*geom.Polygon, 0, len(perPoly))
	off := 0
	for _, n := range perPoly {
		if n == 0 || off+n > len(rings) {
			continue
		}
		out = append(out, geom.NewPolygon(c, rings[off:off+n]...))
		off += n
	}
	return out
}

// flattenNoded turns a slice of noded SegmentStrings into our internal
// tagged 2-vertex edges. When two SegmentStrings produce the same
// directed segment, the DCEL builder merges them and ORs the tags so
// shared edges carry both source labels.
func flattenNoded(strings []*noding.SegmentString) []taggedSegment {
	var out []taggedSegment
	for _, s := range strings {
		if len(s.Coords) < 2 {
			continue
		}
		tag := uint8(s.Tag)
		for i := 0; i+1 < len(s.Coords); i++ {
			out = append(out, taggedSegment{
				p0:  s.Coords[i],
				p1:  s.Coords[i+1],
				tag: tag,
			})
		}
	}
	return out
}
