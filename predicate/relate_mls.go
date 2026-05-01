package predicate

import (
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/kernel"
)

// relateMLStoMLS computes the DE-9IM matrix for two MultiLineStrings using
// the OGC mod-2 boundary rule at the multi level.
//
// Per-member relate followed by max-merge mis-classifies intersection
// points that fall at a vertex of one component while ALSO being a mod-2
// boundary point of the parent multi: the per-pair pass records the
// vertex as "interior" (because it's in the middle of one line) while at
// the multi level it's a boundary point (because it appears as endpoint
// in an odd number of components). This routine drops that ambiguity by
// classifying every intersection point against the multi-level
// boundary sets directly.
func relateMLStoMLS(a, b *geom.MultiLineString, k kernel.Kernel) matrix {
	m := newMatrix()
	m.raise(mEE, 2)
	if a.IsEmpty() || b.IsEmpty() {
		return m
	}
	aBd := mlsBoundarySetMap(a)
	bBd := mlsBoundarySetMap(b)

	aLines := mlsActiveLines(a)
	bLines := mlsActiveLines(b)
	if len(aLines) == 0 || len(bLines) == 0 {
		return m
	}

	for _, aLs := range aLines {
		for i := 0; i+1 < aLs.NumPoints(); i++ {
			a1, a2 := aLs.PointAt(i), aLs.PointAt(i+1)
			if a1 == a2 {
				continue
			}
			for _, bLs := range bLines {
				for j := 0; j+1 < bLs.NumPoints(); j++ {
					b1, b2 := bLs.PointAt(j), bLs.PointAt(j+1)
					if b1 == b2 {
						continue
					}
					recordMLSSegmentPair(&m, a1, a2, b1, b2, aBd, bBd, k)
				}
			}
		}
	}

	if !mlsFullyOn(a, b, k) {
		m.raise(mIE, 1)
	}
	if !mlsFullyOn(b, a, k) {
		m.raise(mEI, 1)
	}
	for p := range aBd {
		if !pointOnAnyLineOfMLS(p, b, k) {
			m.raise(mBE, 0)
			break
		}
	}
	for p := range bBd {
		if !pointOnAnyLineOfMLS(p, a, k) {
			m.raise(mEB, 0)
			break
		}
	}
	return m
}

// recordMLSSegmentPair classifies the intersection of two segments using
// multi-level boundary sets.
func recordMLSSegmentPair(m *matrix, a1, a2, b1, b2 geom.XY, aBd, bBd map[geom.XY]bool, k kernel.Kernel) {
	if collinearOverlap(a1, a2, b1, b2) {
		m.raise(mII, 1)
		for _, p := range []geom.XY{a1, a2} {
			if k.SegmentDistance(p, b1, b2) <= 1e-12 {
				classifyMLSIntersection(m, p, aBd, bBd)
			}
		}
		for _, p := range []geom.XY{b1, b2} {
			if k.SegmentDistance(p, a1, a2) <= 1e-12 {
				classifyMLSIntersection(m, p, aBd, bBd)
			}
		}
		return
	}
	ip, ok := k.SegmentIntersection(a1, a2, b1, b2)
	if !ok {
		for _, p := range []geom.XY{a1, a2} {
			if k.SegmentDistance(p, b1, b2) <= 1e-12 {
				classifyMLSIntersection(m, p, aBd, bBd)
			}
		}
		for _, p := range []geom.XY{b1, b2} {
			if k.SegmentDistance(p, a1, a2) <= 1e-12 {
				classifyMLSIntersection(m, p, aBd, bBd)
			}
		}
		return
	}
	classifyMLSIntersection(m, ip, aBd, bBd)
}

// classifyMLSIntersection classifies a single intersection point at the
// multi-line level: a point in the mod-2 boundary set is boundary,
// otherwise it lies in the interior (since the point is on a member's
// closure by construction).
func classifyMLSIntersection(m *matrix, p geom.XY, aBd, bBd map[geom.XY]bool) {
	aIsB := aBd[p]
	bIsB := bBd[p]
	switch {
	case aIsB && bIsB:
		m.raise(mBB, 0)
	case aIsB:
		m.raise(mBI, 0)
	case bIsB:
		m.raise(mIB, 0)
	default:
		m.raise(mII, 0)
	}
}

// mlsBoundarySetMap returns the OGC mod-2 boundary set as a map for fast
// membership queries.
func mlsBoundarySetMap(ml *geom.MultiLineString) map[geom.XY]bool {
	count := map[geom.XY]int{}
	for i := 0; i < ml.NumGeometries(); i++ {
		ls := ml.LineStringAt(i)
		if ls.IsEmpty() || ls.NumPoints() < 2 {
			continue
		}
		first := ls.PointAt(0)
		last := ls.PointAt(ls.NumPoints() - 1)
		if first == last {
			continue
		}
		count[first]++
		count[last]++
	}
	out := make(map[geom.XY]bool, len(count))
	for p, c := range count {
		if c%2 == 1 {
			out[p] = true
		}
	}
	return out
}

func mlsActiveLines(ml *geom.MultiLineString) []*geom.LineString {
	out := make([]*geom.LineString, 0, ml.NumGeometries())
	for i := 0; i < ml.NumGeometries(); i++ {
		ls := ml.LineStringAt(i)
		if ls.IsEmpty() || ls.NumPoints() < 2 {
			continue
		}
		out = append(out, ls)
	}
	return out
}

// mlsFullyOn reports whether every component-segment of `inner` lies
// entirely on the closure of any component of `outer` (sampled at
// endpoints + midpoint).
func mlsFullyOn(inner, outer *geom.MultiLineString, k kernel.Kernel) bool {
	for i := 0; i < inner.NumGeometries(); i++ {
		ls := inner.LineStringAt(i)
		if ls.IsEmpty() || ls.NumPoints() < 2 {
			continue
		}
		for j := 0; j+1 < ls.NumPoints(); j++ {
			a, b := ls.PointAt(j), ls.PointAt(j+1)
			if a == b {
				continue
			}
			mid := geom.XY{X: (a.X + b.X) / 2, Y: (a.Y + b.Y) / 2}
			if !pointOnAnyLineOfMLS(a, outer, k) ||
				!pointOnAnyLineOfMLS(b, outer, k) ||
				!pointOnAnyLineOfMLS(mid, outer, k) {
				return false
			}
		}
	}
	return true
}

// pointOnAnyLineOfMLS reports whether the point lies on the closure of
// any component of the MultiLineString.
func pointOnAnyLineOfMLS(p geom.XY, ml *geom.MultiLineString, k kernel.Kernel) bool {
	for i := 0; i < ml.NumGeometries(); i++ {
		ls := ml.LineStringAt(i)
		if ls.IsEmpty() {
			continue
		}
		if pointOnLine(p, ls, k) {
			return true
		}
	}
	return false
}

// relateMLStoPolygon computes the matrix for (MultiLineString, Polygon)
// using the mod-2 boundary rule for the lineal side. Mirror of
// relateMLStoMLS but with the polygon's ring-based boundary on the
// other side. The return matrix is in the canonical (a-row, b-col)
// order.
func relateMLStoPolygon(a *geom.MultiLineString, b *geom.Polygon, k kernel.Kernel) matrix {
	m := newMatrix()
	m.raise(mEE, 2)
	if a.IsEmpty() || b.IsEmpty() || b.NumRings() == 0 {
		return m
	}
	aBd := mlsBoundarySetMap(a)
	aLines := mlsActiveLines(a)
	if len(aLines) == 0 {
		return m
	}

	hasInside, hasExterior, hasBoundarySegment := false, false, false
	for _, ls := range aLines {
		for i := 0; i+1 < ls.NumPoints(); i++ {
			p1, p2 := ls.PointAt(i), ls.PointAt(i+1)
			if p1 == p2 {
				continue
			}
			cls := func(p geom.XY, c kernel.Containment) {
				switch c {
				case kernel.Inside:
					if aBd[p] {
						m.raise(mBI, 0)
					} else {
						m.raise(mII, 0)
					}
				case kernel.OnBoundary:
					if aBd[p] {
						m.raise(mBB, 0)
					} else {
						m.raise(mIB, 0)
					}
				default:
					if aBd[p] {
						m.raise(mBE, 0)
					} else {
						m.raise(mIE, 0)
					}
				}
			}
			c1 := pointInPolygon(p1, b, k)
			c2 := pointInPolygon(p2, b, k)
			cls(p1, c1)
			cls(p2, c2)
			// Collect intersection parameters along this A segment
			// against every ring segment of B, then sample sub-intervals.
			ts := []float64{0, 1}
			dx, dy := p2.X-p1.X, p2.Y-p1.Y
			denom2 := dx*dx + dy*dy
			for r := 0; r < b.NumRings(); r++ {
				ring := b.Ring(r)
				for j := 0; j+1 < len(ring); j++ {
					q1, q2 := ring[j], ring[j+1]
					ip, ok := k.SegmentIntersection(p1, p2, q1, q2)
					if !ok {
						continue
					}
					t := ((ip.X-p1.X)*dx + (ip.Y-p1.Y)*dy) / denom2
					if t > 1e-15 && t < 1-1e-15 {
						ts = append(ts, t)
					}
				}
			}
			sortFloats(ts)
			anyInsideSeg, anyOutsideSeg, allBoundarySeg := false, false, true
			for s := 0; s+1 < len(ts); s++ {
				ta, tb := ts[s], ts[s+1]
				if tb-ta < 1e-15 {
					continue
				}
				tm := (ta + tb) / 2
				sample := geom.XY{X: p1.X + tm*dx, Y: p1.Y + tm*dy}
				switch pointInPolygon(sample, b, k) {
				case kernel.Inside:
					anyInsideSeg = true
					allBoundarySeg = false
				case kernel.Outside:
					anyOutsideSeg = true
					allBoundarySeg = false
				}
			}
			if c1 == kernel.Inside || c2 == kernel.Inside {
				anyInsideSeg = true
			}
			if c1 == kernel.Outside || c2 == kernel.Outside {
				anyOutsideSeg = true
			}
			if c1 != kernel.OnBoundary || c2 != kernel.OnBoundary {
				allBoundarySeg = false
			}
			if anyInsideSeg {
				m.raise(mII, 1)
				hasInside = true
			}
			if anyOutsideSeg {
				m.raise(mIE, 1)
				hasExterior = true
			}
			if allBoundarySeg {
				m.raise(mIB, 1)
				hasBoundarySegment = true
			}
		}
	}

	// Edge crossings between A and B's rings: each crossing contributes
	// at the boundary classification of A (mod-2). Where the crossing
	// point is at A's mod-2 boundary, raise BB; otherwise IB.
	for _, ls := range aLines {
		for i := 0; i+1 < ls.NumPoints(); i++ {
			a1, a2 := ls.PointAt(i), ls.PointAt(i+1)
			if a1 == a2 {
				continue
			}
			for r := 0; r < b.NumRings(); r++ {
				ring := b.Ring(r)
				for j := 0; j+1 < len(ring); j++ {
					b1, b2 := ring[j], ring[j+1]
					ip, ok := k.SegmentIntersection(a1, a2, b1, b2)
					if !ok {
						continue
					}
					if aBd[ip] {
						m.raise(mBB, 0)
					} else {
						m.raise(mIB, 0)
					}
				}
			}
		}
	}

	// EI: polygon's interior is 2-D and non-empty.
	m.raise(mEI, 2)
	// EB: polygon's boundary is 1-D. EB = F when every point of B's
	// boundary lies on A's closure; otherwise EB=1.
	if !polygonBoundaryFullyOnMLS(b, a, k) {
		m.raise(mEB, 1)
	}
	_ = hasInside
	_ = hasExterior
	_ = hasBoundarySegment
	return m
}

// subSegmentOnBoundary reports whether finer subdivision of segment p1-p2
// shows additional points lying on the polygon's boundary (i.e. the
// segment coincides with the boundary over a non-zero length).
func subSegmentOnBoundary(p1, p2 geom.XY, b *geom.Polygon, k kernel.Kernel) bool {
	q := geom.XY{X: (3*p1.X + p2.X) / 4, Y: (3*p1.Y + p2.Y) / 4}
	r := geom.XY{X: (p1.X + 3*p2.X) / 4, Y: (p1.Y + 3*p2.Y) / 4}
	cq := pointInPolygon(q, b, k)
	cr := pointInPolygon(r, b, k)
	// 1-D coincidence over half of segment requires both quarter and
	// midpoint (or three-quarter and midpoint) to be on the boundary.
	return cq == kernel.OnBoundary || cr == kernel.OnBoundary
}

// polygonBoundaryFullyOnMLS reports whether every segment of b's rings
// lies on the closure of a. Used to determine whether EB=F (i.e. no
// part of b's boundary is in a's exterior).
func polygonBoundaryFullyOnMLS(b *geom.Polygon, a *geom.MultiLineString, k kernel.Kernel) bool {
	for r := 0; r < b.NumRings(); r++ {
		ring := b.Ring(r)
		for j := 0; j+1 < len(ring); j++ {
			s1, s2 := ring[j], ring[j+1]
			if s1 == s2 {
				continue
			}
			mid := geom.XY{X: (s1.X + s2.X) / 2, Y: (s1.Y + s2.Y) / 2}
			if !pointOnAnyLineOfMLS(s1, a, k) ||
				!pointOnAnyLineOfMLS(s2, a, k) ||
				!pointOnAnyLineOfMLS(mid, a, k) {
				return false
			}
		}
	}
	return true
}

// relateMLStoMultiPolygon computes the matrix for (MLS, MultiPolygon).
// MultiPolygon members are non-overlapping (valid input), so a sample's
// point-in-polygon result against the multi is the disjunction of
// per-member results. We classify each A-segment sample against the
// MultiPolygon as a whole.
func relateMLStoMultiPolygon(a *geom.MultiLineString, b *geom.MultiPolygon, k kernel.Kernel) matrix {
	m := newMatrix()
	m.raise(mEE, 2)
	if a.IsEmpty() || b.IsEmpty() {
		return m
	}
	aBd := mlsBoundarySetMap(a)
	aLines := mlsActiveLines(a)
	if len(aLines) == 0 {
		return m
	}
	classify := func(p geom.XY) kernel.Containment {
		any := kernel.Outside
		for i := 0; i < b.NumGeometries(); i++ {
			pg := b.PolygonAt(i)
			if pg.IsEmpty() {
				continue
			}
			c := pointInPolygon(p, pg, k)
			if c == kernel.Inside {
				return kernel.Inside
			}
			if c == kernel.OnBoundary {
				any = kernel.OnBoundary
			}
		}
		return any
	}

	for _, ls := range aLines {
		for i := 0; i+1 < ls.NumPoints(); i++ {
			p1, p2 := ls.PointAt(i), ls.PointAt(i+1)
			if p1 == p2 {
				continue
			}
			cls := func(p geom.XY, c kernel.Containment) {
				switch c {
				case kernel.Inside:
					if aBd[p] {
						m.raise(mBI, 0)
					} else {
						m.raise(mII, 0)
					}
				case kernel.OnBoundary:
					if aBd[p] {
						m.raise(mBB, 0)
					} else {
						m.raise(mIB, 0)
					}
				default:
					if aBd[p] {
						m.raise(mBE, 0)
					} else {
						m.raise(mIE, 0)
					}
				}
			}
			c1 := classify(p1)
			c2 := classify(p2)
			cls(p1, c1)
			cls(p2, c2)
			ts := []float64{0, 1}
			dx, dy := p2.X-p1.X, p2.Y-p1.Y
			denom2 := dx*dx + dy*dy
			for kk := 0; kk < b.NumGeometries(); kk++ {
				pg := b.PolygonAt(kk)
				for r := 0; r < pg.NumRings(); r++ {
					ring := pg.Ring(r)
					for j := 0; j+1 < len(ring); j++ {
						q1, q2 := ring[j], ring[j+1]
						ip, ok := k.SegmentIntersection(p1, p2, q1, q2)
						if !ok {
							continue
						}
						t := ((ip.X-p1.X)*dx + (ip.Y-p1.Y)*dy) / denom2
						if t > 1e-15 && t < 1-1e-15 {
							ts = append(ts, t)
						}
					}
				}
			}
			sortFloats(ts)
			anyInsideSeg, anyOutsideSeg, allBoundarySeg := false, false, true
			for s := 0; s+1 < len(ts); s++ {
				ta, tb := ts[s], ts[s+1]
				if tb-ta < 1e-15 {
					continue
				}
				tm := (ta + tb) / 2
				sample := geom.XY{X: p1.X + tm*dx, Y: p1.Y + tm*dy}
				switch classify(sample) {
				case kernel.Inside:
					anyInsideSeg = true
					allBoundarySeg = false
				case kernel.Outside:
					anyOutsideSeg = true
					allBoundarySeg = false
				}
			}
			if c1 == kernel.Inside || c2 == kernel.Inside {
				anyInsideSeg = true
			}
			if c1 == kernel.Outside || c2 == kernel.Outside {
				anyOutsideSeg = true
			}
			if c1 != kernel.OnBoundary || c2 != kernel.OnBoundary {
				allBoundarySeg = false
			}
			if anyInsideSeg {
				m.raise(mII, 1)
			}
			if anyOutsideSeg {
				m.raise(mIE, 1)
			}
			if allBoundarySeg {
				m.raise(mIB, 1)
			}
		}
	}

	// Edge crossings vs each polygon's rings.
	for _, ls := range aLines {
		for i := 0; i+1 < ls.NumPoints(); i++ {
			a1, a2 := ls.PointAt(i), ls.PointAt(i+1)
			if a1 == a2 {
				continue
			}
			for k2 := 0; k2 < b.NumGeometries(); k2++ {
				pg := b.PolygonAt(k2)
				for r := 0; r < pg.NumRings(); r++ {
					ring := pg.Ring(r)
					for j := 0; j+1 < len(ring); j++ {
						b1, b2 := ring[j], ring[j+1]
						ip, ok := k.SegmentIntersection(a1, a2, b1, b2)
						if !ok {
							continue
						}
						if aBd[ip] {
							m.raise(mBB, 0)
						} else {
							m.raise(mIB, 0)
						}
					}
				}
			}
		}
	}

	m.raise(mEI, 2)
	if !mpBoundaryFullyOnMLS(b, a, k) {
		m.raise(mEB, 1)
	}
	return m
}

func subSegmentOnBoundaryMP(p1, p2 geom.XY, b *geom.MultiPolygon, k kernel.Kernel) bool {
	q := geom.XY{X: (3*p1.X + p2.X) / 4, Y: (3*p1.Y + p2.Y) / 4}
	r := geom.XY{X: (p1.X + 3*p2.X) / 4, Y: (p1.Y + 3*p2.Y) / 4}
	cls := func(p geom.XY) kernel.Containment {
		any := kernel.Outside
		for i := 0; i < b.NumGeometries(); i++ {
			pg := b.PolygonAt(i)
			if pg.IsEmpty() {
				continue
			}
			c := pointInPolygon(p, pg, k)
			if c == kernel.Inside {
				return kernel.Inside
			}
			if c == kernel.OnBoundary {
				any = kernel.OnBoundary
			}
		}
		return any
	}
	return cls(q) == kernel.OnBoundary || cls(r) == kernel.OnBoundary
}

func mpBoundaryFullyOnMLS(b *geom.MultiPolygon, a *geom.MultiLineString, k kernel.Kernel) bool {
	for i := 0; i < b.NumGeometries(); i++ {
		pg := b.PolygonAt(i)
		if pg.IsEmpty() {
			continue
		}
		if !polygonBoundaryFullyOnMLS(pg, a, k) {
			return false
		}
	}
	return true
}
