package overlayng

import "github.com/terra-geo/terra/geom"

// classifyFacesPolygons tags each non-outer face with whether its
// interior lies inside subj (resp. clip), accounting for HOLES. A point
// is "inside" a multi-ring polygon iff it's inside the outer ring AND
// not inside any interior ring.
func classifyFacesPolygons(d *dcel, subjRings, clipRings [][]geom.XY) {
	for _, f := range d.faces {
		ip := interiorPoint(f)
		f.inSubj = pointInPolygonRings(ip, subjRings)
		f.inClip = pointInPolygonRings(ip, clipRings)
	}
}

// classifyFaces (kept for the simpler shell-only call site in the
// disjoint helper) routes through the multi-ring path with a single ring.
func classifyFaces(d *dcel, subjRing, clipRing []geom.XY) {
	classifyFacesPolygons(d, [][]geom.XY{subjRing}, [][]geom.XY{clipRing})
}

// classifyFacesByPolygons is the multi-aware classifier: it tags each
// face with whether its interior lies inside any subj polygon (resp.
// any clip polygon). subjPerPoly partitions subjRings into
// per-polygon ring lists ([outer, holes...]); same for clip.
//
// To make classification robust against narrow-sliver inputs that
// collapse under snap-rounding (causing one input's edge to appear
// inside another input's degenerate band when nudged
// perpendicularly), we use TWO separate sample points:
//
//   - inSubj is tested at a point nudged from a clip-only-tagged
//     edge (where available). This avoids the nudge landing on
//     subj's collapsed sliver.
//   - inClip is tested symmetrically at a point nudged from a
//     subj-only-tagged edge.
//
// When no single-source edge of the desired tag exists, we fall back
// to the standard interiorPoint which picks the longest non-spur
// edge regardless of tag.
func classifyFacesByPolygons(d *dcel,
	subjRings [][]geom.XY, subjPerPoly []int,
	clipRings [][]geom.XY, clipPerPoly []int,
) {
	for _, f := range d.faces {
		// Sample for inSubj: prefer an edge contributed by clip only.
		ipSubj := interiorPointPreferringTag(f, 2, 1)
		f.inSubj = pointInAnyPolygon(ipSubj, subjRings, subjPerPoly)
		// Sample for inClip: prefer an edge contributed by subj only.
		ipClip := interiorPointPreferringTag(f, 1, 2)
		f.inClip = pointInAnyPolygon(ipClip, clipRings, clipPerPoly)
	}
}

// interiorPointPreferringTag returns a face interior point computed
// from the longest non-spur edge whose tag has the `prefer` bit set
// AND not the `avoid` bit. If no such edge exists, falls back to
// any longest non-spur edge.
func interiorPointPreferringTag(f *face, prefer, avoid uint8) geom.XY {
	if len(f.edges) == 0 {
		return geom.XY{}
	}
	bestIdx := -1
	var bestLen2 float64
	// First pass: prefer edges with `prefer` tag set and `avoid` tag NOT set.
	for i, e := range f.edges {
		if e.twin != nil && e.twin.face == f {
			continue
		}
		if e.tags&prefer == 0 || e.tags&avoid != 0 {
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
	if bestIdx >= 0 {
		return edgeNudgePoint(f.edges[bestIdx])
	}
	// Fallback: longest non-spur edge regardless of tag.
	return interiorPoint(f)
}

// edgeNudgePoint returns the midpoint of edge e nudged perpendicular
// into the face on the LEFT (the face's interior by DCEL convention).
func edgeNudgePoint(e *halfEdge) geom.XY {
	x0, y0 := e.origin.p.X, e.origin.p.Y
	x1, y1 := e.target.p.X, e.target.p.Y
	mx, my := (x0+x1)/2, (y0+y1)/2
	dx, dy := x1-x0, y1-y0
	const eps = 1e-9
	return geom.XY{X: mx + -dy*eps, Y: my + dx*eps}
}

// pointInAnyPolygon iterates over the per-polygon partitions and
// returns true iff p is inside any of them. Each partition is one
// polygon's ring list ([outer, holes...]); pointInPolygonRings handles
// the per-polygon "inside outer AND not inside any hole" semantics.
func pointInAnyPolygon(p geom.XY, rings [][]geom.XY, perPoly []int) bool {
	off := 0
	for _, n := range perPoly {
		if n == 0 || off+n > len(rings) {
			off += n
			continue
		}
		if pointInPolygonRings(p, rings[off:off+n]) {
			return true
		}
		off += n
	}
	return false
}

func pointInPolygonRings(p geom.XY, rings [][]geom.XY) bool {
	if len(rings) == 0 {
		return false
	}
	if !pointInRing(p, rings[0]) {
		return false
	}
	for i := 1; i < len(rings); i++ {
		if pointInRing(p, rings[i]) {
			return false
		}
	}
	return true
}

// interiorPoint returns a point guaranteed to be strictly inside the
// face f: midpoint of the longest non-spur edge nudged perpendicular
// into the face (left of edge direction by DCEL convention).
//
// Spur edges (e.face == e.twin.face) are skipped because they're
// internal "spikes" within the face; perpendicular nudges from them
// can fall on either side and behave unpredictably.
func interiorPoint(f *face) geom.XY {
	if len(f.edges) == 0 {
		return geom.XY{}
	}
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
		bestIdx = 0
	}
	return edgeNudgePoint(f.edges[bestIdx])
}

// pointInRing is the standard ray-cast (crossing-number) test against
// a closed ring. Returns true iff p is strictly interior. Boundary
// classification is handled separately by callers that need it.
func pointInRing(p geom.XY, ring []geom.XY) bool {
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
