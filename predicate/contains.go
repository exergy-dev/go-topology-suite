package predicate

import (
	terra "github.com/terra-geo/terra"
	"github.com/terra-geo/terra/crs"
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/kernel"
)

// Contains reports whether a contains b — every point of b lies in the
// interior or boundary of a, and the interiors intersect.
//
// Phase 1 supports the common cases: Polygon-contains-Point,
// Polygon-contains-LineString, Polygon-contains-Polygon. Other type pairs
// return false; full topological Contains is a Phase 3 deliverable that
// rides on the DE-9IM matrix.
func Contains(a, b geom.Geometry, opts ...Option) (bool, error) {
	if !crs.Equal(a.CRS(), b.CRS()) {
		return false, terra.ErrCRSMismatch
	}
	if a.IsEmpty() || b.IsEmpty() {
		return false, nil
	}
	c := resolve(a, opts)
	if c.kernel.Name() == "planar" && !a.Envelope().Contains(b.Envelope()) {
		return false, nil
	}
	// Prepared fast-path for Polygon-contains-Point — the most common
	// hot loop. Falls through to the generic path for other type pairs.
	if c.prepared != nil {
		if pb, ok := b.(*geom.Point); ok {
			cont := c.prepared.ContainsPoint(pb.XY())
			return cont == kernel.Inside, nil
		}
	}
	switch va := a.(type) {
	case *geom.Polygon:
		return polygonContains(va, b, c.kernel)
	case *geom.MultiPolygon:
		// b must be contained in some single member.
		for i := 0; i < va.NumGeometries(); i++ {
			ok, err := polygonContains(va.PolygonAt(i), b, c.kernel)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
		}
		return false, nil
	}
	return false, nil
}

// Within is Contains with the operands swapped.
func Within(a, b geom.Geometry, opts ...Option) (bool, error) {
	return Contains(b, a, opts...)
}

func polygonContains(a *geom.Polygon, b geom.Geometry, k kernel.Kernel) (bool, error) {
	switch vb := b.(type) {
	case *geom.Point:
		c := pointInPolygon(vb.XY(), a, k)
		return c == kernel.Inside, nil
	case *geom.LineString:
		// Every vertex inside or on boundary, and no edge crosses
		// outward through any ring of a.
		n := vb.NumPoints()
		for i := 0; i < n; i++ {
			if pointInPolygon(vb.PointAt(i), a, k) == kernel.Outside {
				return false, nil
			}
		}
		// Edges must not properly cross any ring.
		for r := 0; r < a.NumRings(); r++ {
			ring := a.Ring(r)
			for i := 0; i+1 < n; i++ {
				p1, p2 := vb.PointAt(i), vb.PointAt(i+1)
				for j := 0; j+1 < len(ring); j++ {
					ip, ok := k.SegmentIntersection(p1, p2, ring[j], ring[j+1])
					if !ok {
						continue
					}
					// Touching at a vertex is fine; only proper crossings disqualify.
					if ip != p1 && ip != p2 && ip != ring[j] && ip != ring[j+1] {
						return false, nil
					}
				}
			}
		}
		return true, nil
	case *geom.Polygon:
		// Every vertex of b's outer ring must lie inside or on boundary of a;
		// no edge of b may properly cross any edge of a.
		bOuter := vb.Ring(0)
		for _, p := range bOuter {
			if pointInPolygon(p, a, k) == kernel.Outside {
				return false, nil
			}
		}
		for ra := 0; ra < a.NumRings(); ra++ {
			ringA := a.Ring(ra)
			for i := 0; i+1 < len(bOuter); i++ {
				for j := 0; j+1 < len(ringA); j++ {
					ip, ok := k.SegmentIntersection(bOuter[i], bOuter[i+1], ringA[j], ringA[j+1])
					if !ok {
						continue
					}
					if ip != bOuter[i] && ip != bOuter[i+1] && ip != ringA[j] && ip != ringA[j+1] {
						return false, nil
					}
				}
			}
		}
		return true, nil
	}
	return false, nil
}
