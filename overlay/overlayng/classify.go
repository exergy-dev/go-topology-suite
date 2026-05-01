package overlayng

import "github.com/terra-geo/terra/geom"

// classifyFacesPolygons tags each non-outer face with whether its
// interior lies inside subj (resp. clip), accounting for HOLES. A point
// is "inside" a multi-ring polygon iff it's inside the outer ring AND
// not inside any interior ring.
func classifyFacesPolygons(d *dcel, subjRings, clipRings [][]geom.XY) {
	for _, f := range d.faces {
		if f.isOuter {
			continue
		}
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
// non-outer face with whether its interior lies inside any subj polygon
// (resp. any clip polygon). subjPerPoly partitions subjRings into
// per-polygon ring lists ([outer, holes...]); same for clip.
func classifyFacesByPolygons(d *dcel,
	subjRings [][]geom.XY, subjPerPoly []int,
	clipRings [][]geom.XY, clipPerPoly []int,
) {
	for _, f := range d.faces {
		if f.isOuter {
			continue
		}
		ip := interiorPoint(f)
		f.inSubj = pointInAnyPolygon(ip, subjRings, subjPerPoly)
		f.inClip = pointInAnyPolygon(ip, clipRings, clipPerPoly)
	}
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
// face f. Technique: midpoint of the first edge nudged perpendicular
// to the LEFT (where the face lives, by our DCEL convention). Reliable
// for any simple polygonal face including non-convex ones.
func interiorPoint(f *face) geom.XY {
	if len(f.edges) == 0 {
		return geom.XY{}
	}
	e := f.edges[0]
	x0, y0 := e.origin.p.X, e.origin.p.Y
	x1, y1 := e.target.p.X, e.target.p.Y
	mx, my := (x0+x1)/2, (y0+y1)/2
	dx, dy := x1-x0, y1-y0
	const eps = 1e-9
	return geom.XY{
		X: mx + -dy*eps,
		Y: my + dx*eps,
	}
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
