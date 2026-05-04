package predicate

import (
	"errors"

	terra "github.com/terra-geo/terra"
	"github.com/terra-geo/terra/crs"
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/kernel"
)

// Intersects reports whether a and b share at least one point.
//
// Returns ErrCRSMismatch if the geometries' CRS differ. The empty case is
// well-defined: any geometry with an empty operand is Disjoint, so
// Intersects returns false (not an error).
func Intersects(a, b geom.Geometry, opts ...Option) (bool, error) {
	if !crs.Equal(a.CRS(), b.CRS()) {
		return false, terra.ErrCRSMismatch
	}
	a = unwrapLinearRing(a)
	b = unwrapLinearRing(b)
	c := resolve(a, opts)
	// Envelope-first short-circuit is only sound when the kernel agrees
	// with lon/lat-space rectangles — i.e. for planar. Geographic
	// envelopes that span the antimeridian look disjoint to the planar
	// envelope test even when the underlying spherical geometries
	// intersect. Until terra/index grows a spherical-cap variant we
	// simply skip the short-circuit for non-planar kernels.
	//
	// Routed through the RelateNG short-circuit layer
	// (relate_short_circuit.go) so all predicates share consistent
	// envelope/dim fast paths.
	if sc := scIntersects(a, b, c.kernel.Name() == "planar"); sc.resolved {
		return sc.get(), nil
	}
	// Prepared fast-path. The Polygon-vs-Point case is checked FIRST
	// (before the preparedIntersector tier) because ContainsPoint goes
	// straight to the segment R-tree, while preparedIntersector.Intersects
	// routes through walkVertices + closure dispatch — measurably slower
	// for the very common single-point query.
	if c.prepared != nil {
		if pb, ok := b.(*geom.Point); ok {
			return c.prepared.ContainsPoint(pb.XY()) != kernel.Outside, nil
		}
		if pi, ok := c.prepared.(preparedIntersector); ok {
			return pi.Intersects(b), nil
		}
	}
	return intersectsDispatch(a, b, c.kernel)
}

// Disjoint is the complement of Intersects.
func Disjoint(a, b geom.Geometry, opts ...Option) (bool, error) {
	x, err := Intersects(a, b, opts...)
	if err != nil {
		return false, err
	}
	return !x, nil
}

func intersectsDispatch(a, b geom.Geometry, k kernel.Kernel) (bool, error) {
	// Order operands so that the type code of a <= b. Symmetric handling
	// per pair-of-types.
	if typeRank(a) > typeRank(b) {
		a, b = b, a
	}
	switch va := a.(type) {
	case *geom.Point:
		return pointIntersectsAny(va, b, k)
	case *geom.LineString:
		switch vb := b.(type) {
		case *geom.LineString:
			return lineLineIntersects(va, vb, k), nil
		case *geom.Polygon:
			return lineRingsIntersect(va, vb, k), nil
		default:
			return collectionFanout(a, b, k)
		}
	case *geom.Polygon:
		switch vb := b.(type) {
		case *geom.Polygon:
			return polygonPolygonIntersects(va, vb, k), nil
		default:
			return collectionFanout(a, b, k)
		}
	default:
		return collectionFanout(a, b, k)
	}
}

// typeRank ranks the seven geometry types so that ordering gives stable
// dispatch. Lower rank = simpler shape.
func typeRank(g geom.Geometry) int {
	switch g.(type) {
	case *geom.Point:
		return 0
	case *geom.LineString:
		return 1
	case *geom.Polygon:
		return 2
	case *geom.MultiPoint:
		return 3
	case *geom.MultiLineString:
		return 4
	case *geom.MultiPolygon:
		return 5
	case *geom.GeometryCollection:
		return 6
	default:
		return 99
	}
}

// pointIntersectsAny checks a point against any other geometry.
func pointIntersectsAny(p *geom.Point, other geom.Geometry, k kernel.Kernel) (bool, error) {
	pp := p.XY()
	switch v := other.(type) {
	case *geom.Point:
		return pp == v.XY(), nil
	case *geom.LineString:
		return pointOnLine(pp, v, k), nil
	case *geom.Polygon:
		return pointInPolygon(pp, v, k) != kernel.Outside, nil
	case *geom.MultiPoint:
		for i := 0; i < v.NumGeometries(); i++ {
			if v.PointAt(i) == pp {
				return true, nil
			}
		}
		return false, nil
	case *geom.MultiLineString:
		for i := 0; i < v.NumGeometries(); i++ {
			if pointOnLine(pp, v.LineStringAt(i), k) {
				return true, nil
			}
		}
		return false, nil
	case *geom.MultiPolygon:
		for i := 0; i < v.NumGeometries(); i++ {
			if pointInPolygon(pp, v.PolygonAt(i), k) != kernel.Outside {
				return true, nil
			}
		}
		return false, nil
	case *geom.GeometryCollection:
		for i := 0; i < v.NumGeometries(); i++ {
			ok, err := Intersects(p, v.GeometryAt(i))
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
		}
		return false, nil
	}
	return false, errors.New("predicate: unhandled geometry type")
}

func pointOnLine(p geom.XY, ls *geom.LineString, k kernel.Kernel) bool {
	n := ls.NumPoints()
	for i := 0; i+1 < n; i++ {
		a, b := ls.PointAt(i), ls.PointAt(i+1)
		if pointOnSegmentRobust(p, a, b, k) {
			return true
		}
	}
	return false
}

func pointOnSegmentRobust(p, a, b geom.XY, k kernel.Kernel) bool {
	if a == b {
		return p == a
	}
	if k.Orient(a, b, p) != kernel.Collinear {
		return false
	}
	const eps = 1e-12
	minX, maxX := a.X, b.X
	if minX > maxX {
		minX, maxX = maxX, minX
	}
	minY, maxY := a.Y, b.Y
	if minY > maxY {
		minY, maxY = maxY, minY
	}
	return p.X >= minX-eps && p.X <= maxX+eps &&
		p.Y >= minY-eps && p.Y <= maxY+eps
}

func lineLineIntersects(a, b *geom.LineString, k kernel.Kernel) bool {
	na, nb := a.NumPoints(), b.NumPoints()
	for i := 0; i+1 < na; i++ {
		a1, a2 := a.PointAt(i), a.PointAt(i+1)
		for j := 0; j+1 < nb; j++ {
			b1, b2 := b.PointAt(j), b.PointAt(j+1)
			if _, ok := k.SegmentIntersection(a1, a2, b1, b2); ok {
				return true
			}
			// Handle collinear-touch cases that SegmentIntersection misses.
			if k.SegmentDistance(b1, a1, a2) == 0 || k.SegmentDistance(b2, a1, a2) == 0 {
				return true
			}
			if k.SegmentDistance(a1, b1, b2) == 0 || k.SegmentDistance(a2, b1, b2) == 0 {
				return true
			}
		}
	}
	return false
}

func lineRingsIntersect(ls *geom.LineString, p *geom.Polygon, k kernel.Kernel) bool {
	// Any vertex inside the polygon, or any edge crossing any ring.
	n := ls.NumPoints()
	for i := 0; i < n; i++ {
		if pointInPolygon(ls.PointAt(i), p, k) != kernel.Outside {
			return true
		}
	}
	bufp := borrowRingBuf()
	defer releaseRingBuf(bufp)
	for r := 0; r < p.NumRings(); r++ {
		ring := p.RingInto((*bufp)[:0], r)
		*bufp = ring
		for i := 0; i+1 < n; i++ {
			a1, a2 := ls.PointAt(i), ls.PointAt(i+1)
			for j := 0; j+1 < len(ring); j++ {
				if _, ok := k.SegmentIntersection(a1, a2, ring[j], ring[j+1]); ok {
					return true
				}
			}
		}
	}
	return false
}

func polygonPolygonIntersects(a, b *geom.Polygon, k kernel.Kernel) bool {
	// Quick: any vertex of a inside b, or vice versa, or any edge crossing.
	bufA := borrowRingBuf()
	defer releaseRingBuf(bufA)
	bufB := borrowRingBuf()
	defer releaseRingBuf(bufB)

	for r := 0; r < a.NumRings(); r++ {
		ring := a.RingInto((*bufA)[:0], r)
		*bufA = ring
		for _, v := range ring {
			if pointInPolygon(v, b, k) != kernel.Outside {
				return true
			}
		}
	}
	for r := 0; r < b.NumRings(); r++ {
		ring := b.RingInto((*bufB)[:0], r)
		*bufB = ring
		for _, v := range ring {
			if pointInPolygon(v, a, k) != kernel.Outside {
				return true
			}
		}
	}
	for ra := 0; ra < a.NumRings(); ra++ {
		ringA := a.RingInto((*bufA)[:0], ra)
		*bufA = ringA
		for rb := 0; rb < b.NumRings(); rb++ {
			ringB := b.RingInto((*bufB)[:0], rb)
			*bufB = ringB
			for i := 0; i+1 < len(ringA); i++ {
				for j := 0; j+1 < len(ringB); j++ {
					if _, ok := k.SegmentIntersection(ringA[i], ringA[i+1], ringB[j], ringB[j+1]); ok {
						return true
					}
				}
			}
		}
	}
	return false
}

// pointInPolygon: outer ring contains, then no hole strictly contains.
// Borrows a pooled scratch buffer for ring snapshots so the hot
// PIP-many-points loop stays alloc-free.
func pointInPolygon(p geom.XY, poly *geom.Polygon, k kernel.Kernel) kernel.Containment {
	if poly.NumRings() == 0 {
		return kernel.Outside
	}
	bufp := borrowRingBuf()
	defer releaseRingBuf(bufp)
	outer := poly.RingInto((*bufp)[:0], 0)
	*bufp = outer
	c := k.PointInRing(p, outer)
	if c == kernel.Outside {
		return kernel.Outside
	}
	for r := 1; r < poly.NumRings(); r++ {
		ring := poly.RingInto((*bufp)[:0], r)
		*bufp = ring
		hc := k.PointInRing(p, ring)
		if hc == kernel.Inside {
			return kernel.Outside
		}
		if hc == kernel.OnBoundary {
			return kernel.OnBoundary
		}
	}
	return c
}

// collectionFanout handles MultiX and GeometryCollection by checking each
// member of the collection against the other operand.
func collectionFanout(a, b geom.Geometry, k kernel.Kernel) (bool, error) {
	for i := 0; i < a.NumGeometries(); i++ {
		ai := childOf(a, i)
		if ai == nil {
			continue
		}
		for j := 0; j < b.NumGeometries(); j++ {
			bj := childOf(b, j)
			if bj == nil {
				continue
			}
			ok, err := Intersects(ai, bj, WithKernel(k))
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
		}
	}
	return false, nil
}

// childOf returns the i-th sub-geometry, treating singleton types as a
// 1-element collection.
func childOf(g geom.Geometry, i int) geom.Geometry {
	switch v := g.(type) {
	case *geom.MultiPoint:
		// MultiPoint members aren't separately addressable as *Point here;
		// produce a fresh Point for the dispatch.
		return geom.NewPoint(v.CRS(), v.PointAt(i))
	case *geom.MultiLineString:
		return v.LineStringAt(i)
	case *geom.MultiPolygon:
		return v.PolygonAt(i)
	case *geom.GeometryCollection:
		return v.GeometryAt(i)
	default:
		if i == 0 {
			return g
		}
		return nil
	}
}
