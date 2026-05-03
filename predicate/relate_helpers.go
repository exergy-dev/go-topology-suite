package predicate

import (
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/kernel"
	"github.com/terra-geo/terra/kernel/planar"
)

// Geometric helpers shared by Contains / Covers / Equals / short-circuit
// fast paths to answer degenerate-shape questions without invoking RelateNG.

// isZeroLengthLine reports whether every vertex of ls coincides with
// the first — i.e. the LineString collapses to a single point.
func isZeroLengthLine(ls *geom.LineString) bool {
	if ls.NumPoints() == 0 {
		return false
	}
	first := ls.PointAt(0)
	for i := 1; i < ls.NumPoints(); i++ {
		if ls.PointAt(i) != first {
			return false
		}
	}
	return true
}

// boundaryPoint describes a line endpoint's coordinate (and whether the
// line has a boundary at all — for closed lines, the boundary is empty).
type boundaryPoint struct {
	xy    geom.XY
	valid bool
}

// lineBoundary returns the two-element OGC boundary set of a
// LineString: the start and end vertices for an open curve, an
// all-invalid pair for a closed line or a degenerate input.
func lineBoundary(ls *geom.LineString) [2]boundaryPoint {
	if ls.NumPoints() < 2 {
		return [2]boundaryPoint{}
	}
	first := ls.PointAt(0)
	last := ls.PointAt(ls.NumPoints() - 1)
	if first == last {
		return [2]boundaryPoint{}
	}
	return [2]boundaryPoint{
		{xy: first, valid: true},
		{xy: last, valid: true},
	}
}

// pointInBoundarySet reports whether p coincides with one of the two
// boundary points returned by lineBoundary.
func pointInBoundarySet(p geom.XY, b [2]boundaryPoint) bool {
	return (b[0].valid && b[0].xy == p) || (b[1].valid && b[1].xy == p)
}

// lineFullyOn reports whether every point of `inner` lies on `outer`.
// Each segment of inner is sampled at its endpoints AND midpoint —
// catching the case where two adjacent vertices of inner lie on outer
// but the segment between them detours outside outer (e.g., when the
// two lines overlap on partial segments only).
func lineFullyOn(inner, outer *geom.LineString, k kernel.Kernel) bool {
	for i := 0; i < inner.NumPoints(); i++ {
		if !pointOnLine(inner.PointAt(i), outer, k) {
			return false
		}
	}
	for i := 0; i+1 < inner.NumPoints(); i++ {
		a, b := inner.PointAt(i), inner.PointAt(i+1)
		if a == b {
			continue
		}
		mid := geom.XY{X: (a.X + b.X) / 2, Y: (a.Y + b.Y) / 2}
		if !pointOnLine(mid, outer, k) {
			return false
		}
	}
	return true
}

// samplePoint returns an interior representative for a polygon: the
// midpoint of the first edge of the outer ring nudged toward the
// polygon's interior. The nudge direction is chosen by signed-area
// orientation: CCW rings have interior on the LEFT of each edge;
// CW rings have interior on the RIGHT.
func samplePoint(p *geom.Polygon) geom.XY {
	if p.NumRings() == 0 {
		return geom.XY{}
	}
	if p.RingLen(0) < 2 {
		return geom.XY{}
	}
	bufp := borrowRingBuf()
	defer releaseRingBuf(bufp)
	ring := p.RingInto((*bufp)[:0], 0)
	*bufp = ring
	a, b := ring[0], ring[1]
	mx, my := (a.X+b.X)/2, (a.Y+b.Y)/2
	dx, dy := b.X-a.X, b.Y-a.Y
	const eps = 1e-9
	// Signed area > 0 ⇒ CCW ⇒ left normal (-dy, dx) points inward.
	// Signed area < 0 ⇒ CW  ⇒ right normal (dy, -dx) points inward.
	if (planar.Kernel{}).RingArea(ring) >= 0 {
		return geom.XY{X: mx - dy*eps, Y: my + dx*eps}
	}
	return geom.XY{X: mx + dy*eps, Y: my - dx*eps}
}
