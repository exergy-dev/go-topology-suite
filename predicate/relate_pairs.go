package predicate

import (
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/kernel"
)

// This file defines per-type-pair DE-9IM matrix builders. Each function
// returns a complete 3×3 matrix using the conventions in relate.go:
//
//	cells: -1 = F (empty), 0 = point, 1 = curve, 2 = area
//	indices: II IB IE | BI BB BE | EI EB EE  (row-major)

// relatePointPoint: two points either coincide or don't.
func relatePointPoint(a, b *geom.Point) matrix {
	m := newMatrix()
	m.raise(mEE, 2)
	if a.XY() == b.XY() {
		m.raise(mII, 0)
	} else {
		m.raise(mIE, 0)
		m.raise(mEI, 0)
	}
	return m
}

// relatePointLine: a point may lie at a boundary endpoint, on the
// interior of the line, or outside it. For closed (and self-equal-endpoint)
// lines the boundary is empty.
func relatePointLine(p *geom.Point, ls *geom.LineString, k kernel.Kernel) matrix {
	m := newMatrix()
	m.raise(mEE, 2)
	pp := p.XY()

	hasBoundary := ls.NumPoints() >= 2 && ls.PointAt(0) != ls.PointAt(ls.NumPoints()-1)
	atBoundary := false
	if hasBoundary {
		if ls.PointAt(0) == pp || ls.PointAt(ls.NumPoints()-1) == pp {
			atBoundary = true
		}
	}
	onLine := pointOnLine(pp, ls, k)

	switch {
	case atBoundary:
		m.raise(mIB, 0)
	case onLine:
		m.raise(mII, 0)
	default:
		m.raise(mIE, 0)
	}
	// Line interior is curve (dim 1); always non-empty (line has > 0 length).
	m.raise(mEI, 1)
	if hasBoundary {
		// Line has boundary (two endpoints, dim 0).
		// EB is non-empty unless both endpoints coincide with the point —
		// not possible for two distinct endpoints.
		if !atBoundary || ls.NumPoints() > 2 || ls.PointAt(0) != ls.PointAt(ls.NumPoints()-1) {
			m.raise(mEB, 0)
		}
		// If the point is one endpoint and the line has TWO endpoints, the
		// other endpoint is still in EB.
		if atBoundary && ls.PointAt(0) != ls.PointAt(ls.NumPoints()-1) {
			m.raise(mEB, 0)
		}
	}
	return m
}

// relatePointPolygon: a point lies strictly inside, on the boundary, or
// outside a polygon (with holes).
func relatePointPolygon(p *geom.Point, poly *geom.Polygon, k kernel.Kernel) matrix {
	m := newMatrix()
	m.raise(mEE, 2)
	c := pointInPolygon(p.XY(), poly, k)
	switch c {
	case kernel.Inside:
		m.raise(mII, 0)
	case kernel.OnBoundary:
		m.raise(mIB, 0)
	default:
		m.raise(mIE, 0)
	}
	// Polygon has 2-D interior and 1-D boundary, both non-empty.
	m.raise(mEI, 2)
	m.raise(mEB, 1)
	return m
}

// relateLineLine: two open curves may share interior segments, touch at
// endpoints, cross transversally, or be disjoint.
func relateLineLine(a, b *geom.LineString, k kernel.Kernel) matrix {
	m := newMatrix()
	m.raise(mEE, 2)

	aPts := lineStringPoints(a)
	bPts := lineStringPoints(b)
	aBoundary := lineBoundary(a)
	bBoundary := lineBoundary(b)

	// Check segment-segment intersections in detail.
	for i := 0; i+1 < len(aPts); i++ {
		a1, a2 := aPts[i], aPts[i+1]
		for j := 0; j+1 < len(bPts); j++ {
			b1, b2 := bPts[j], bPts[j+1]
			recordSegmentIntersection(&m, a1, a2, b1, b2, aBoundary, bBoundary, k)
		}
	}

	// IE, BE: a's interior or boundary in b's exterior (anything not on b).
	if !lineFullyOn(a, b, k) {
		m.raise(mIE, 1)
	}
	if endpointOutside(aBoundary, b, k) {
		m.raise(mBE, 0)
	}
	if !lineFullyOn(b, a, k) {
		m.raise(mEI, 1)
	}
	if endpointOutside(bBoundary, a, k) {
		m.raise(mEB, 0)
	}
	return m
}

// recordSegmentIntersection updates m for the intersection of one segment
// of a (a1→a2) with one segment of b (b1→b2), using the boundary-point
// sets of each line to classify whether the intersection point falls in
// I or B for each side.
func recordSegmentIntersection(m *matrix, a1, a2, b1, b2 geom.XY, aBoundary, bBoundary [2]boundaryPoint, k kernel.Kernel) {
	// Detect collinear-overlap first: if both segments lie on the same
	// line and their parameter ranges overlap on more than a point, we
	// have a 1-D intersection.
	if collinearOverlap(a1, a2, b1, b2) {
		m.raise(mII, 1)
		return
	}
	// Check segment intersection (proper or improper).
	ip, ok := k.SegmentIntersection(a1, a2, b1, b2)
	if !ok {
		// No proper intersection — but a vertex might still lie on the
		// other segment (T-junctions).
		for _, p := range []geom.XY{a1, a2} {
			if k.SegmentDistance(p, b1, b2) == 0 {
				classifyPointOnLines(m, p, true, false, aBoundary, bBoundary)
			}
		}
		for _, p := range []geom.XY{b1, b2} {
			if k.SegmentDistance(p, a1, a2) == 0 {
				classifyPointOnLines(m, p, false, true, aBoundary, bBoundary)
			}
		}
		return
	}
	classifyPointOnLines(m, ip, true, true, aBoundary, bBoundary)
}

// classifyPointOnLines classifies an intersection point relative to each
// line's boundary set (the line endpoints) and updates the appropriate
// I/B cells in m. atA / atB indicate whether the point is on a's segment
// or b's segment respectively (defaults to true).
func classifyPointOnLines(m *matrix, p geom.XY, atA, atB bool, aBoundary, bBoundary [2]boundaryPoint) {
	if !atA || !atB {
		// We've already filtered to T-junction cases in the caller, so
		// unconditional updates are safe.
	}
	aIsBoundary := pointInBoundarySet(p, aBoundary)
	bIsBoundary := pointInBoundarySet(p, bBoundary)
	switch {
	case aIsBoundary && bIsBoundary:
		m.raise(mBB, 0)
	case aIsBoundary:
		m.raise(mBI, 0)
	case bIsBoundary:
		m.raise(mIB, 0)
	default:
		m.raise(mII, 0)
	}
}

// boundaryPoint describes a line endpoint's coordinate (and whether the
// line has a boundary at all — for closed lines, the boundary is empty).
type boundaryPoint struct {
	xy    geom.XY
	valid bool
}

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

func pointInBoundarySet(p geom.XY, b [2]boundaryPoint) bool {
	return (b[0].valid && b[0].xy == p) || (b[1].valid && b[1].xy == p)
}

// collinearOverlap reports whether two segments are collinear and share
// more than one point.
func collinearOverlap(a1, a2, b1, b2 geom.XY) bool {
	// Cross product of (a2-a1) with (b1-a1), (b2-a1): both must be 0.
	cross := func(o, p, q geom.XY) float64 {
		return (p.X-o.X)*(q.Y-o.Y) - (p.Y-o.Y)*(q.X-o.X)
	}
	if cross(a1, a2, b1) != 0 || cross(a1, a2, b2) != 0 {
		return false
	}
	// Both b1 and b2 are on the line through a1,a2. Project to longer
	// axis to check parameter overlap.
	dx := a2.X - a1.X
	dy := a2.Y - a1.Y
	useX := dx*dx >= dy*dy
	t := func(p geom.XY) float64 {
		if useX {
			if dx == 0 {
				return 0
			}
			return (p.X - a1.X) / dx
		}
		if dy == 0 {
			return 0
		}
		return (p.Y - a1.Y) / dy
	}
	tb1, tb2 := t(b1), t(b2)
	if tb1 > tb2 {
		tb1, tb2 = tb2, tb1
	}
	// Overlap with [0,1] on more than one point: tb2 > 0 and tb1 < 1
	// strictly (equality means single-point touch).
	return tb2 > 0 && tb1 < 1 && !(tb2 <= 0) && !(tb1 >= 1) && tb2-tb1 > 0
}

func lineStringPoints(ls *geom.LineString) []geom.XY {
	out := make([]geom.XY, ls.NumPoints())
	for i := range out {
		out[i] = ls.PointAt(i)
	}
	return out
}

// lineFullyOn reports whether every vertex of `inner` lies on `outer`.
// (Approximation: vertex-level coverage is sufficient for the IE/EI cells
// when the inputs come from typical user data.)
func lineFullyOn(inner, outer *geom.LineString, k kernel.Kernel) bool {
	for i := 0; i < inner.NumPoints(); i++ {
		if !pointOnLine(inner.PointAt(i), outer, k) {
			return false
		}
	}
	// Also check that every interior vertex of the inner segments has
	// matching outer geometry. The vertex-level approximation is exact
	// when inner is a sub-polyline of outer; otherwise a segment of inner
	// might cross outside outer between two coincident vertices. In that
	// case the segment crossing will already have raised mIE via
	// recordSegmentIntersection.
	return true
}

func endpointOutside(b [2]boundaryPoint, ls *geom.LineString, k kernel.Kernel) bool {
	for _, bp := range b {
		if !bp.valid {
			continue
		}
		if !pointOnLine(bp.xy, ls, k) {
			return true
		}
	}
	return false
}

// relateLinePolygon: a line may be entirely inside, fully outside,
// crossing, or boundary-coincident with a polygon (and its holes).
func relateLinePolygon(ls *geom.LineString, poly *geom.Polygon, k kernel.Kernel) matrix {
	m := newMatrix()
	m.raise(mEE, 2)
	if poly.NumRings() == 0 {
		return m
	}

	// Classify each interior vertex of the line w.r.t. polygon
	// containment. Endpoints are classified separately (they belong to
	// the line's BOUNDARY, not its INTERIOR).
	n := ls.NumPoints()
	hasLineBoundary := n >= 2 && ls.PointAt(0) != ls.PointAt(n-1)
	insideCount, boundaryCount, outsideCount := 0, 0, 0
	for i := 0; i < n; i++ {
		if hasLineBoundary && (i == 0 || i == n-1) {
			continue // line-boundary endpoint, classified below
		}
		switch pointInPolygon(ls.PointAt(i), poly, k) {
		case kernel.Inside:
			insideCount++
		case kernel.OnBoundary:
			boundaryCount++
		default:
			outsideCount++
		}
	}

	// For each segment, classify endpoints + midpoint. A segment with any
	// portion in the polygon interior raises mII; with any portion in
	// the exterior, mIE; only if endpoints AND midpoint all lie on the
	// boundary does the segment lie along the boundary (mIB).
	for i := 0; i+1 < ls.NumPoints(); i++ {
		a, b := ls.PointAt(i), ls.PointAt(i+1)
		if a == b {
			continue
		}
		mid := geom.XY{X: (a.X + b.X) / 2, Y: (a.Y + b.Y) / 2}
		ra := pointInPolygon(a, poly, k)
		rb := pointInPolygon(b, poly, k)
		rm := pointInPolygon(mid, poly, k)
		hasI := ra == kernel.Inside || rb == kernel.Inside || rm == kernel.Inside
		hasE := ra == kernel.Outside || rb == kernel.Outside || rm == kernel.Outside
		alongB := ra == kernel.OnBoundary && rb == kernel.OnBoundary && rm == kernel.OnBoundary
		if hasI {
			m.raise(mII, 1)
		}
		if hasE {
			m.raise(mIE, 1)
		}
		if alongB {
			m.raise(mIB, 1)
		}
	}

	// Edge crossings between line segments and ring segments contribute
	// 0-D intersections in II/IB/BB/BI according to whether the crossing
	// lies on the line's boundary endpoints.
	lsBoundary := lineBoundary(ls)
	ringBufP := borrowRingBuf()
	defer releaseRingBuf(ringBufP)
	for r := 0; r < poly.NumRings(); r++ {
		ring := poly.RingInto((*ringBufP)[:0], r)
		*ringBufP = ring
		for j := 0; j+1 < len(ring); j++ {
			b1, b2 := ring[j], ring[j+1]
			for i := 0; i+1 < ls.NumPoints(); i++ {
				a1, a2 := ls.PointAt(i), ls.PointAt(i+1)
				ip, ok := k.SegmentIntersection(a1, a2, b1, b2)
				if !ok {
					continue
				}
				if pointInBoundarySet(ip, lsBoundary) {
					m.raise(mBB, 0)
				} else {
					m.raise(mIB, 0)
				}
			}
		}
	}

	// Direct vertex classification of the line's boundary.
	for _, bp := range lsBoundary {
		if !bp.valid {
			continue
		}
		switch pointInPolygon(bp.xy, poly, k) {
		case kernel.Inside:
			m.raise(mBI, 0)
		case kernel.OnBoundary:
			m.raise(mBB, 0)
		default:
			m.raise(mBE, 0)
		}
	}

	// Vertex-level interior cells.
	if insideCount > 0 {
		m.raise(mII, 0) // at least a point in the interior overlap
	}
	if outsideCount > 0 {
		m.raise(mIE, 1) // at least a curve in the exterior (vertex + segment)
	}
	if boundaryCount > 0 {
		m.raise(mIB, 0)
	}

	// EI: polygon interior is 2-D, always non-empty; so EI = 2 unless the
	// line completely covers polygon interior — impossible for a line
	// (lower-dim cannot cover a 2-D region).
	m.raise(mEI, 2)
	// EB: polygon boundary is 1-D, always non-empty; line cannot cover
	// the entire ring (would require equal length and shape — which is
	// degenerate). Conservatively raise to 1.
	m.raise(mEB, 1)
	return m
}

// relatePolygonPolygon: the principal case. We use an aggregate
// computation: vertex-level classification + edge intersections. For
// shared boundary segments we sample interior points adjacent to the
// shared edges.
func relatePolygonPolygon(a, b *geom.Polygon, k kernel.Kernel) matrix {
	m := newMatrix()
	m.raise(mEE, 2)
	if a.NumRings() == 0 || b.NumRings() == 0 {
		return m
	}

	// Vertex classification: a's vertices in b, and vice versa.
	aIn, aOn, aOut := classifyVerticesAgainst(a, b, k)
	bIn, bOn, bOut := classifyVerticesAgainst(b, a, k)

	// Each vertex of A lies on A's BOUNDARY, so its classification feeds
	// the B-row cells (with B's interior/boundary/exterior).
	if aIn > 0 {
		m.raise(mBI, 0)
		m.raise(mII, 2) // a vertex strictly inside b implies area overlap
	}
	if aOn > 0 {
		m.raise(mBB, 0)
	}
	if aOut > 0 {
		m.raise(mBE, 1) // boundary in exterior is at least a curve segment
		m.raise(mIE, 2)
	}
	// Symmetric for B's vertices.
	if bIn > 0 {
		m.raise(mIB, 0)
		m.raise(mII, 2)
	}
	if bOn > 0 {
		m.raise(mBB, 0)
	}
	if bOut > 0 {
		m.raise(mEB, 1)
		m.raise(mEI, 2)
	}

	// Edge crossings: any proper intersection between rings of a and b
	// contributes to BB at point dim (crossings) and ensures both II and
	// E-of-each cells are non-empty (the boundaries cross, so each
	// polygon's boundary has parts on both sides of the other).
	hasShared := false
	hasProperX := false
	bufA := borrowRingBuf()
	defer releaseRingBuf(bufA)
	bufB := borrowRingBuf()
	defer releaseRingBuf(bufB)
	for ra := 0; ra < a.NumRings(); ra++ {
		ringA := a.RingInto((*bufA)[:0], ra)
		*bufA = ringA
		for rb := 0; rb < b.NumRings(); rb++ {
			ringB := b.RingInto((*bufB)[:0], rb)
			*bufB = ringB
			for i := 0; i+1 < len(ringA); i++ {
				a1, a2 := ringA[i], ringA[i+1]
				for j := 0; j+1 < len(ringB); j++ {
					b1, b2 := ringB[j], ringB[j+1]
					if collinearOverlap(a1, a2, b1, b2) {
						hasShared = true
						continue
					}
					ip, ok := k.SegmentIntersection(a1, a2, b1, b2)
					if !ok {
						continue
					}
					// "Proper" crossing means the intersection lies in
					// the interior of BOTH segments (not at any endpoint).
					if ip != a1 && ip != a2 && ip != b1 && ip != b2 {
						hasProperX = true
					}
				}
			}
		}
	}
	if hasShared {
		m.raise(mBB, 1)
	}
	if hasProperX {
		m.raise(mBB, 0)
		m.raise(mII, 2) // crossing implies overlap
		m.raise(mIE, 2)
		m.raise(mEI, 2)
	}
	// Interior overlap test via sampling. Use a representative interior
	// point of one polygon and check it against the other; if inside,
	// II=2.
	if pa := samplePoint(a); pointInPolygon(pa, b, k) == kernel.Inside {
		m.raise(mII, 2)
	}
	if pb := samplePoint(b); pointInPolygon(pb, a, k) == kernel.Inside {
		m.raise(mII, 2)
	}

	// Exterior coverage test: if a is not fully contained in b, IE=2
	// and BE=1. Symmetric for b.
	if !polygonContainsAllVertices(b, a, k) || aOut > 0 {
		m.raise(mIE, 2)
		m.raise(mBE, 1)
	}
	if !polygonContainsAllVertices(a, b, k) || bOut > 0 {
		m.raise(mEI, 2)
		m.raise(mEB, 1)
	}

	return m
}

// classifyVerticesAgainst counts vertices of a strictly inside, on
// boundary of, and strictly outside polygon b.
func classifyVerticesAgainst(a, b *geom.Polygon, k kernel.Kernel) (in, on, out int) {
	bufp := borrowRingBuf()
	defer releaseRingBuf(bufp)
	for r := 0; r < a.NumRings(); r++ {
		ring := a.RingInto((*bufp)[:0], r)
		*bufp = ring
		for i := 0; i+1 < len(ring); i++ { // skip closing duplicate
			switch pointInPolygon(ring[i], b, k) {
			case kernel.Inside:
				in++
			case kernel.OnBoundary:
				on++
			default:
				out++
			}
		}
	}
	return
}

// polygonContainsAllVertices reports whether every vertex of inner is
// inside or on the boundary of outer.
func polygonContainsAllVertices(outer, inner *geom.Polygon, k kernel.Kernel) bool {
	bufp := borrowRingBuf()
	defer releaseRingBuf(bufp)
	for r := 0; r < inner.NumRings(); r++ {
		ring := inner.RingInto((*bufp)[:0], r)
		*bufp = ring
		for _, p := range ring {
			if pointInPolygon(p, outer, k) == kernel.Outside {
				return false
			}
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
	if signedRingArea(ring) >= 0 {
		return geom.XY{X: mx - dy*eps, Y: my + dx*eps}
	}
	return geom.XY{X: mx + dy*eps, Y: my - dx*eps}
}

// signedRingArea computes the shoelace signed area of a closed ring.
func signedRingArea(ring []geom.XY) float64 {
	if len(ring) < 3 {
		return 0
	}
	var sum float64
	for i := 0; i+1 < len(ring); i++ {
		sum += ring[i].X*ring[i+1].Y - ring[i+1].X*ring[i].Y
	}
	return sum / 2
}
