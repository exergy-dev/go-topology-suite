package buffer

import (
	"cmp"
	"math"
	"slices"

	"github.com/terra-geo/terra/crs"
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/internal/noding"
	"github.com/terra-geo/terra/internal/snaprounding"
	"github.com/terra-geo/terra/kernel/planar"
)

// offsetSegment is one edge of a parallel-offset curve, oriented so the
// buffer-INTERIOR is on the LEFT of the edge direction. depthDelta is
// the signed contribution to a face's depth when a horizontal ray cast
// from the face crosses this edge.
//
// The polygonizer's depth-from-input invariant is: at any face F,
//
//   depth(F) = winding-number sum over offset edges crossed by a ray
//              from F to +infinity, weighted by depthDelta.
//
// For a positive buffer of any input ring, every emitted segment has
// depthDelta = +1; faces with depth >= 1 are inside the buffer. For a
// negative buffer (inset), the offset is reoriented so depthDelta is
// still +1 with the inset interior on the LEFT side; nothing about the
// downstream pipeline changes.
type offsetSegment struct {
	p0, p1     geom.XY
	depthDelta int8
}

// emitPolygonOffsetSegments converts every ring of p into offset
// segments tagged with depthDelta=+1 (buffer interior on LEFT). The
// orientation is normalised so the polygonizer's depth invariant
// holds regardless of input ring direction.
//
// distance > 0 ("dilation"): every ring's offset goes OUTWARD from the
// ring's geometric interior — for the outer ring this is exterior of
// polygon; for a hole this is INTO the hole interior (which is outside
// the polygon body). The buffer-interior side of every offset segment
// is on the LEFT when walked in the ring's natural direction.
//
// distance < 0 ("inset"): every ring's offset goes INWARD into the
// ring's geometric interior — outer's offset moves INTO the polygon;
// hole's offset moves OUT of the hole into the polygon body. The
// inset-interior side of every offset segment is on the LEFT when
// walked in the ring's natural direction.
func emitPolygonOffsetSegments(p *geom.Polygon, distance float64, cfg config) []offsetSegment {
	if p == nil || p.IsEmpty() || p.NumRings() == 0 || distance == 0 {
		return nil
	}
	d := math.Abs(distance)
	var out []offsetSegment
	for r := 0; r < p.NumRings(); r++ {
		ring := p.Ring(r)
		if len(ring) < 4 {
			continue
		}
		ringCCW := planar.Default.RingArea(ring) > 0
		isHole := r > 0
		// Choose the offset side: which side of the RING is the buffer
		// expanding INTO?
		//
		//   positive buffer + outer ring: ring-exterior (away from
		//     polygon interior, growing outward).
		//   positive buffer + hole ring: ring-interior (into the hole,
		//     shrinking the hole).
		//   negative buffer + outer ring: ring-interior (into polygon,
		//     shrinking the outer).
		//   negative buffer + hole ring: ring-exterior (into polygon
		//     body, growing the hole).
		//
		// offsetClosedRing's `outward=true` puts the offset on the
		// RING-EXTERIOR side iff the ring is CCW (signed=-d, RIGHT
		// of direction = exterior of CCW). For CW rings the convention
		// flips: outward=true gives ring-INTERIOR. So:
		//   wantRingExterior == (outward == ringCCW)
		// Rearranging:
		//   outward = (wantRingExterior == ringCCW)
		wantRingExterior := (distance > 0) != isHole
		outward := wantRingExterior == ringCCW
		// Inversion guard for holes during positive buffer: when a
		// hole is too small for the offset distance, its mitre/round
		// corners overshoot beyond the original hole's extent and the
		// emitted "shrunk hole" ring is actually LARGER than the
		// original (e.g., a 1×1 hole offset by d=2 produces a 3×3
		// mitre square). Such an inverted offset would create a
		// spurious depth-deficit region inside what should be filled
		// buffer. Skip these — the polygonizer naturally fills the
		// hole because the outer offset's depth dominates with no
		// hole-offset contribution.
		if r > 0 && distance > 0 && holeIsConsumed(ring, d) {
			continue
		}
		offset, ok := offsetClosedRing(ring, d, outward, cfg)
		if !ok {
			continue
		}
		// Orient so buffer-interior is on the LEFT of every emitted
		// segment direction (depthDelta=+1 invariant of the
		// polygonizer). Working through the four cases of {ring
		// orientation × outer/hole} for positive buffer:
		//
		//   CCW outer: offset on ring-exterior, walked CCW. Buffer
		//     interior is between original and offset → on LEFT of
		//     offset direction. ✓ natural emission.
		//   CW outer:  offset on ring-exterior, walked CW. Buffer
		//     interior on RIGHT of offset direction. → reverse.
		//   CCW hole:  offset on ring-interior (inside hole), walked
		//     CCW. Buffer interior between original hole boundary and
		//     offset is on RIGHT of offset direction. → reverse.
		//   CW hole:   offset on ring-interior, walked CW. Buffer
		//     interior on LEFT. ✓ natural emission.
		//
		// Pattern: reverse iff ringCCW XNOR isHole (i.e., ringCCW ==
		// isHole). The same rule applies to negative buffer because
		// the inset interior is on the same relative side of its
		// natural offset direction as the dilation case.
		reverse := ringCCW == isHole
		for i := 0; i+1 < len(offset); i++ {
			a, b := offset[i], offset[i+1]
			if reverse {
				// Walk the offset ring backward, swapping each
				// segment's endpoints so the segment direction also
				// flips. For a closed ring of length N+1 (last == first),
				// segment i in reverse is offset[N-i] -> offset[N-1-i].
				n := len(offset)
				a, b = offset[n-1-i], offset[n-2-i]
			}
			if a == b {
				continue
			}
			// Skip near-zero segments — these arise from mitre joins
			// where a two adjacent corner vertices are within ULP
			// distance of each other due to floating-point noise. They
			// confuse the noder (the two endpoints become separate
			// vertices in the DCEL) and produce spurious zero-area
			// faces.
			dx, dy := b.X-a.X, b.Y-a.Y
			const minLen2 = 1e-20
			if dx*dx+dy*dy < minLen2 {
				continue
			}
			out = append(out, offsetSegment{p0: a, p1: b, depthDelta: 1})
		}
	}
	return out
}

// holeIsConsumed reports whether a hole ring is too small to survive
// a positive buffer of magnitude d. The simple bounding-box bound:
// if the smaller side of the hole's bbox is less than 2d, no point
// inside the hole is at distance > d from the hole boundary, so the
// hole is fully consumed by the dilation.
func holeIsConsumed(ring []geom.XY, d float64) bool {
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

// polygonizeBuffer is the JTS-style buffer pipeline:
//  1. Snap-round the input offset segments so every intersection is a
//     shared vertex.
//  2. Build a DCEL of the noded segments.
//  3. Compute each face's depth via ray-casting against the original
//     offset segments (winding-number sum weighted by depthDelta).
//  4. Mark faces with depth >= 1 as "inside the buffer".
//  5. Walk boundary half-edges (kept ↔ not-kept) to extract result rings.
//  6. Assemble rings into Polygons / MultiPolygon by containment.
//
// tolerance is the snap-rounding grid spacing. Pass tolerance = 0 to
// skip snap-rounding (the noder will still split segments at exact
// intersections via its initial non-rounded pass).
func polygonizeBuffer(c *crs.CRS, segs []offsetSegment, tolerance float64) (geom.Geometry, error) {
	if len(segs) == 0 {
		return geom.NewEmptyPolygon(c, geom.LayoutXY), nil
	}

	noded, err := snapRoundOffsets(segs, tolerance)
	if err != nil {
		return nil, err
	}
	if len(noded) == 0 {
		return geom.NewEmptyPolygon(c, geom.LayoutXY), nil
	}

	g := buildPolygonizeDCEL(noded)
	if g == nil || len(g.faces) == 0 {
		return geom.NewEmptyPolygon(c, geom.LayoutXY), nil
	}

	labelFaceDepths(g, noded)
	rings := extractKeptRings(g)
	if len(rings) == 0 {
		return geom.NewEmptyPolygon(c, geom.LayoutXY), nil
	}

	return assemblePolygonizeRings(c, rings), nil
}

// snapRoundOffsets feeds the offset segments through the
// snap-rounding noder. Tolerance == 0 keeps coordinates intact and
// only inserts intersection vertices.
//
// The noder emits SegmentStrings whose Tag carries depthDelta+128 so
// the [-127, +127] depth range fits in a uint8. Output segments
// inherit their parent string's Tag.
func snapRoundOffsets(segs []offsetSegment, tolerance float64) ([]offsetSegment, error) {
	// Group consecutive segments with the same depthDelta into chains
	// (common case: every segment of one offset ring has the same
	// depth, so the chain ends up being the entire ring).
	type chain struct {
		coords []geom.XY
		delta  int8
	}
	var chains []chain
	flush := func(c *chain) {
		if c == nil || len(c.coords) < 2 {
			return
		}
		chains = append(chains, *c)
	}

	var cur *chain
	for _, s := range segs {
		if cur == nil || cur.delta != s.depthDelta || cur.coords[len(cur.coords)-1] != s.p0 {
			flush(cur)
			cur = &chain{delta: s.depthDelta, coords: []geom.XY{s.p0, s.p1}}
			continue
		}
		cur.coords = append(cur.coords, s.p1)
	}
	flush(cur)

	if len(chains) == 0 {
		return nil, nil
	}

	strings := make([]*noding.SegmentString, 0, len(chains))
	for _, ch := range chains {
		strings = append(strings, &noding.SegmentString{
			Coords: append([]geom.XY(nil), ch.coords...),
			Tag:    int(ch.delta) + 128,
		})
	}

	if tolerance > 0 {
		out, _, err := (&snaprounding.Noder{Tolerance: tolerance}).Node(strings)
		if err != nil {
			// Best-effort: noder couldn't converge; use the un-rounded
			// chains directly. The DCEL build below will still attempt
			// to construct a valid subdivision.
			return flattenChains(strings), nil
		}
		return flattenChains(out), nil
	}

	out := noding.IndexedNoder{}.Node(strings)
	return flattenChains(out), nil
}

// flattenChains turns SegmentStrings back into individual offsetSegments,
// recovering depthDelta from the Tag (which was tag = depthDelta+128).
func flattenChains(strings []*noding.SegmentString) []offsetSegment {
	var out []offsetSegment
	for _, s := range strings {
		if len(s.Coords) < 2 {
			continue
		}
		delta := int8(s.Tag - 128)
		for i := 0; i+1 < len(s.Coords); i++ {
			a, b := s.Coords[i], s.Coords[i+1]
			if a == b {
				continue
			}
			out = append(out, offsetSegment{p0: a, p1: b, depthDelta: delta})
		}
	}
	return out
}

// pgVertex / pgHalfEdge / pgFace — planar-subdivision primitives for
// the polygonizer. Distinct from overlay/overlayng's DCEL because face
// classification is by signed depth (computed below), not tag-based.
type pgVertex struct {
	p   geom.XY
	out []*pgHalfEdge
}

type pgHalfEdge struct {
	origin, target *pgVertex
	twin           *pgHalfEdge
	next           *pgHalfEdge
	face           *pgFace
	angle          float64
	depthDelta     int8 // +1 if walking from origin→target crosses INTO buffer interior
}

type pgFace struct {
	edges []*pgHalfEdge
	depth int
	keep  bool
}

type pgGraph struct {
	vertices []*pgVertex
	edges    []*pgHalfEdge
	faces    []*pgFace
}

type pgVertexKey struct{ x, y uint64 }

func pgMakeKey(p geom.XY) pgVertexKey {
	return pgVertexKey{x: math.Float64bits(p.X), y: math.Float64bits(p.Y)}
}

// buildPolygonizeDCEL constructs a planar subdivision from the noded
// offset segments. Coincident edges (same endpoints, either direction)
// merge into a single half-edge pair whose depthDelta is the sum of
// contributions — so two oppositely-oriented offsets on the same edge
// cancel out (they share boundary; the boundary is "interior-to-both"
// and contributes nothing to either side's depth).
func buildPolygonizeDCEL(segs []offsetSegment) *pgGraph {
	g := &pgGraph{}
	vmap := map[pgVertexKey]*pgVertex{}
	getVertex := func(p geom.XY) *pgVertex {
		k := pgMakeKey(p)
		if v, ok := vmap[k]; ok {
			return v
		}
		v := &pgVertex{p: p}
		vmap[k] = v
		g.vertices = append(g.vertices, v)
		return v
	}

	type edgeKey struct{ a, b pgVertexKey }
	edgeMap := map[edgeKey]*pgHalfEdge{}

	for _, s := range segs {
		if s.p0 == s.p1 {
			continue
		}
		va := getVertex(s.p0)
		vb := getVertex(s.p1)
		ka := pgMakeKey(va.p)
		kb := pgMakeKey(vb.p)
		fk := edgeKey{ka, kb}
		bk := edgeKey{kb, ka}
		if e, exists := edgeMap[fk]; exists {
			// Same direction reappeared: depths add (the segment is
			// shared between two source curves on the same side).
			e.depthDelta += s.depthDelta
			continue
		}
		if e, exists := edgeMap[bk]; exists {
			// Opposite direction reappeared: walking origin→target on the
			// reverse swaps left and right. depthDelta on the existing
			// (reverse-direction) edge is decremented, twin incremented.
			e.depthDelta -= s.depthDelta
			e.twin.depthDelta += s.depthDelta
			continue
		}
		eFwd := &pgHalfEdge{origin: va, target: vb, depthDelta: s.depthDelta}
		eBack := &pgHalfEdge{origin: vb, target: va, depthDelta: -s.depthDelta}
		eFwd.twin = eBack
		eBack.twin = eFwd
		eFwd.angle = math.Atan2(vb.p.Y-va.p.Y, vb.p.X-va.p.X)
		eBack.angle = math.Atan2(va.p.Y-vb.p.Y, va.p.X-vb.p.X)
		va.out = append(va.out, eFwd)
		vb.out = append(vb.out, eBack)
		g.edges = append(g.edges, eFwd, eBack)
		edgeMap[fk] = eFwd
		edgeMap[bk] = eBack
	}

	for _, v := range g.vertices {
		slices.SortFunc(v.out, func(a, b *pgHalfEdge) int {
			return cmp.Compare(a.angle, b.angle)
		})
	}

	// Set next pointers (predecessor-of-twin rule, same as overlayng).
	for _, e := range g.edges {
		t := e.target
		twin := e.twin
		idx := -1
		for i, oe := range t.out {
			if oe == twin {
				idx = i
				break
			}
		}
		if idx < 0 {
			continue
		}
		nextIdx := (idx - 1 + len(t.out)) % len(t.out)
		e.next = t.out[nextIdx]
	}

	// Trace faces.
	for _, e := range g.edges {
		if e.face != nil {
			continue
		}
		f := &pgFace{}
		cur := e
		const maxSteps = 1 << 20
		for steps := 0; steps < maxSteps; steps++ {
			if cur == nil || cur.face != nil {
				break
			}
			cur.face = f
			f.edges = append(f.edges, cur)
			cur = cur.next
			if cur == e {
				break
			}
		}
		if len(f.edges) > 0 {
			g.faces = append(g.faces, f)
		}
	}

	return g
}

// labelFaceDepths computes each face's depth via ray-casting against
// the noded offset segments. The face's representative interior point
// is the midpoint of its longest edge nudged perpendicular to the LEFT
// (which is the face-interior side under the CCW convention).
func labelFaceDepths(g *pgGraph, segs []offsetSegment) {
	for _, f := range g.faces {
		ip, ok := faceRepresentativePoint(f)
		if !ok {
			continue
		}
		f.depth = rayCastDepth(ip, segs)
		f.keep = f.depth >= 1
	}
}

// faceRepresentativePoint returns the midpoint of the longest non-spur
// edge of f, nudged perpendicular into f's interior (LEFT of edge
// direction by DCEL convention). Returns ok=false if f has no usable
// edge (degenerate).
func faceRepresentativePoint(f *pgFace) (geom.XY, bool) {
	bestIdx := -1
	var bestLen2 float64
	for i, e := range f.edges {
		if e.twin != nil && e.twin.face == f {
			continue
		}
		dx := e.target.p.X - e.origin.p.X
		dy := e.target.p.Y - e.origin.p.Y
		l2 := dx*dx + dy*dy
		if bestIdx < 0 || l2 > bestLen2 {
			bestIdx = i
			bestLen2 = l2
		}
	}
	if bestIdx < 0 {
		// All edges are spurs; pick the first edge regardless.
		if len(f.edges) == 0 {
			return geom.XY{}, false
		}
		bestIdx = 0
		dx := f.edges[0].target.p.X - f.edges[0].origin.p.X
		dy := f.edges[0].target.p.Y - f.edges[0].origin.p.Y
		bestLen2 = dx*dx + dy*dy
		if bestLen2 == 0 {
			return geom.XY{}, false
		}
	}
	e := f.edges[bestIdx]
	mx, my := (e.origin.p.X+e.target.p.X)/2, (e.origin.p.Y+e.target.p.Y)/2
	dx, dy := e.target.p.X-e.origin.p.X, e.target.p.Y-e.origin.p.Y
	l := math.Sqrt(dx*dx + dy*dy)
	if l == 0 {
		return geom.XY{}, false
	}
	// Perpendicular LEFT unit vector: rotate (dx,dy)/l by +90° → (-dy/l, dx/l).
	const eps = 1e-9
	nx, ny := -dy/l, dx/l
	return geom.XY{X: mx + nx*eps, Y: my + ny*eps}, true
}

// rayCastDepth casts a horizontal ray from p to +∞ and sums depthDelta
// contributions of every offset segment crossed. Standard winding-rule
// half-open convention (a.Y > p.Y XOR b.Y > p.Y) so a vertex at p.Y is
// counted on at most one of its incident edges.
func rayCastDepth(p geom.XY, segs []offsetSegment) int {
	depth := 0
	for _, s := range segs {
		a, b := s.p0, s.p1
		if (a.Y > p.Y) == (b.Y > p.Y) {
			continue
		}
		// Compute X of the segment at y=p.Y.
		t := (p.Y - a.Y) / (b.Y - a.Y)
		xCross := a.X + t*(b.X-a.X)
		if xCross <= p.X {
			continue
		}
		// Determine sign: walking origin→target, when ray crosses the
		// segment from RIGHT (below) to LEFT (above) of the direction,
		// contributes +depthDelta. Equivalent winding-number rule:
		//   - segment goes upward (b.Y > a.Y): +depthDelta
		//   - segment goes downward (b.Y < a.Y): -depthDelta
		if b.Y > a.Y {
			depth += int(s.depthDelta)
		} else {
			depth -= int(s.depthDelta)
		}
	}
	return depth
}

// extractKeptRings walks every boundary half-edge (kept face on one
// side, non-kept on the other) into a closed ring.
func extractKeptRings(g *pgGraph) [][]geom.XY {
	isBoundary := func(e *pgHalfEdge) bool {
		if e.face == nil || e.twin == nil || e.twin.face == nil {
			return false
		}
		return e.face.keep && !e.twin.face.keep
	}
	var rings [][]geom.XY
	visited := map[*pgHalfEdge]bool{}
	for _, start := range g.edges {
		if !isBoundary(start) || visited[start] {
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
			next := nextBoundaryAtPGVertex(cur, isBoundary)
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

// nextBoundaryAtPGVertex returns the next outgoing boundary edge in CCW
// order around e.target, starting after twin(e). Returns nil if none.
func nextBoundaryAtPGVertex(e *pgHalfEdge, isBoundary func(*pgHalfEdge) bool) *pgHalfEdge {
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

// assemblePolygonizeRings nests extracted rings into Polygons /
// MultiPolygon by containment. Outer rings (depth-from-other-rings is
// even) get any inner rings (odd depth) directly contained as holes.
func assemblePolygonizeRings(c *crs.CRS, rings [][]geom.XY) geom.Geometry {
	if len(rings) == 0 {
		return geom.NewEmptyPolygon(c, geom.LayoutXY)
	}
	if len(rings) == 1 {
		return geom.NewPolygon(c, rings[0])
	}
	reps := make([]geom.XY, len(rings))
	for i, ring := range rings {
		reps[i] = ringRepPoint(ring)
	}
	depths := make([]int, len(rings))
	for i := range rings {
		for j := range rings {
			if i == j {
				continue
			}
			if pointInRingPG(reps[i], rings[j]) {
				depths[i]++
			}
		}
	}
	type group struct {
		outer int
		holes []int
	}
	var groups []group
	for i := range rings {
		if depths[i]%2 != 0 {
			continue
		}
		gr := group{outer: i}
		for j := range rings {
			if i == j || depths[j] != depths[i]+1 {
				continue
			}
			if !pointInRingPG(reps[j], rings[i]) {
				continue
			}
			deeper := false
			for k := range rings {
				if k == i || depths[k] >= depths[i]+1 {
					continue
				}
				if !pointInRingPG(reps[j], rings[k]) {
					continue
				}
				if depths[k] > depths[i] {
					deeper = true
					break
				}
			}
			if !deeper {
				gr.holes = append(gr.holes, j)
			}
		}
		groups = append(groups, gr)
	}
	if len(groups) == 0 {
		// Defensive: emit each ring as its own polygon.
		polys := make([]*geom.Polygon, 0, len(rings))
		for _, r := range rings {
			polys = append(polys, geom.NewPolygon(c, r))
		}
		if len(polys) == 1 {
			return polys[0]
		}
		return geom.NewMultiPolygon(c, polys...)
	}
	polys := make([]*geom.Polygon, 0, len(groups))
	for _, gr := range groups {
		all := make([][]geom.XY, 0, 1+len(gr.holes))
		all = append(all, rings[gr.outer])
		for _, h := range gr.holes {
			all = append(all, rings[h])
		}
		polys = append(polys, geom.NewPolygon(c, all...))
	}
	if len(polys) == 1 {
		return polys[0]
	}
	return geom.NewMultiPolygon(c, polys...)
}

// ringRepPoint returns a strictly-interior representative point of a
// ring (midpoint of longest segment, nudged into the interior).
func ringRepPoint(ring []geom.XY) geom.XY {
	if len(ring) < 4 {
		if len(ring) > 0 {
			return ring[0]
		}
		return geom.XY{}
	}
	bestIdx := 0
	var bestLen2 float64
	for i := 0; i+1 < len(ring); i++ {
		dx := ring[i+1].X - ring[i].X
		dy := ring[i+1].Y - ring[i].Y
		l2 := dx*dx + dy*dy
		if l2 > bestLen2 {
			bestLen2 = l2
			bestIdx = i
		}
	}
	a, b := ring[bestIdx], ring[bestIdx+1]
	mx, my := (a.X+b.X)/2, (a.Y+b.Y)/2
	dx, dy := b.X-a.X, b.Y-a.Y
	signedArea2 := 0.0
	for i := 0; i+1 < len(ring); i++ {
		signedArea2 += ring[i].X*ring[i+1].Y - ring[i+1].X*ring[i].Y
	}
	const eps = 1e-9
	nx, ny := -dy, dx
	if signedArea2 < 0 {
		nx, ny = dy, -dx
	}
	return geom.XY{X: mx + nx*eps, Y: my + ny*eps}
}

func pointInRingPG(p geom.XY, ring []geom.XY) bool {
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
