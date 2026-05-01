package predicate

import (
	"math"
	"sort"

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
		return collectionContains(multiPolygonAsCollection(va), b, c.kernel), nil
	case *geom.GeometryCollection:
		return collectionContains(va, b, c.kernel), nil
	}
	// Fallback for non-polygonal a (LineString, MultiLineString, Point,
	// MultiPoint, GeometryCollection): delegate to the DE-9IM matrix.
	// Contains pattern: T*****FF* (II=T, EI=F, EB=F).
	d, err := Relate(a, b, opts...)
	if err != nil {
		return false, err
	}
	return d.Matches("T*****FF*"), nil
}

func multiPolygonAsCollection(mp *geom.MultiPolygon) *geom.GeometryCollection {
	parts := make([]geom.Geometry, 0, mp.NumGeometries())
	for i := 0; i < mp.NumGeometries(); i++ {
		parts = append(parts, mp.PolygonAt(i))
	}
	return geom.NewGeometryCollection(mp.CRS(), parts...)
}

func collectionContains(a *geom.GeometryCollection, b geom.Geometry, k kernel.Kernel) bool {
	if a.IsEmpty() || b.IsEmpty() {
		return false
	}
	covered, interior := collectionCoversWithInteriorHit(a, b, k)
	return covered && interior
}

func collectionCoversWithInteriorHit(a *geom.GeometryCollection, b geom.Geometry, k kernel.Kernel) (covered bool, interior bool) {
	switch vb := b.(type) {
	case *geom.Point:
		return collectionPointCoveredInterior(a, vb.XY(), k)
	case *geom.MultiPoint:
		seen := false
		for i := 0; i < vb.NumGeometries(); i++ {
			covered, hit := collectionPointCoveredInterior(a, vb.PointAt(i), k)
			if !covered {
				return false, false
			}
			seen = true
			interior = interior || hit
		}
		return seen, interior
	case *geom.LineString:
		return collectionLineCoveredInterior(a, vb, k)
	case *geom.MultiLineString:
		seen := false
		for i := 0; i < vb.NumGeometries(); i++ {
			covered, hit := collectionLineCoveredInterior(a, vb.LineStringAt(i), k)
			if !covered {
				return false, false
			}
			seen = true
			interior = interior || hit
		}
		return seen, interior
	case *geom.Polygon:
		return collectionPolygonCoveredInterior(a, vb, k)
	case *geom.MultiPolygon:
		seen := false
		for i := 0; i < vb.NumGeometries(); i++ {
			covered, hit := collectionPolygonCoveredInterior(a, vb.PolygonAt(i), k)
			if !covered {
				return false, false
			}
			seen = true
			interior = interior || hit
		}
		return seen, interior
	case *geom.GeometryCollection:
		seen := false
		for i := 0; i < vb.NumGeometries(); i++ {
			g := vb.GeometryAt(i)
			if g.IsEmpty() {
				continue
			}
			covered, hit := collectionCoversWithInteriorHit(a, g, k)
			if !covered {
				return false, false
			}
			seen = true
			interior = interior || hit
		}
		return seen, interior
	}
	return false, false
}

func collectionLineCoveredInterior(a *geom.GeometryCollection, line *geom.LineString, k kernel.Kernel) (bool, bool) {
	if line.NumPoints() == 0 {
		return false, false
	}
	interior := false
	for i := 0; i < line.NumPoints(); i++ {
		covered, hit := collectionPointCoveredInterior(a, line.PointAt(i), k)
		if !covered {
			return false, false
		}
		interior = interior || hit
	}
	for i := 0; i+1 < line.NumPoints(); i++ {
		p, q := line.PointAt(i), line.PointAt(i+1)
		ts := []float64{0, 1}
		collectionBoundaryParams(a, p, q, k, &ts)
		sort.Float64s(ts)
		for j := 0; j+1 < len(ts); j++ {
			if ts[j+1]-ts[j] <= 1e-12 {
				continue
			}
			t := (ts[j] + ts[j+1]) / 2
			sample := interpolate(p, q, t)
			covered, hit := collectionPointCoveredInterior(a, sample, k)
			if !covered {
				return false, false
			}
			interior = interior || hit
		}
	}
	return true, interior
}

func collectionBoundaryParams(a *geom.GeometryCollection, p, q geom.XY, k kernel.Kernel, ts *[]float64) {
	for i := 0; i < a.NumGeometries(); i++ {
		geometryBoundaryParams(a.GeometryAt(i), p, q, k, ts)
	}
}

func geometryBoundaryParams(g geom.Geometry, p, q geom.XY, k kernel.Kernel, ts *[]float64) {
	switch v := g.(type) {
	case *geom.Polygon:
		polygonBoundaryParams(v, p, q, k, ts)
	case *geom.MultiPolygon:
		for i := 0; i < v.NumGeometries(); i++ {
			polygonBoundaryParams(v.PolygonAt(i), p, q, k, ts)
		}
	case *geom.GeometryCollection:
		for i := 0; i < v.NumGeometries(); i++ {
			geometryBoundaryParams(v.GeometryAt(i), p, q, k, ts)
		}
	}
}

func polygonBoundaryParams(poly *geom.Polygon, p, q geom.XY, k kernel.Kernel, ts *[]float64) {
	for r := 0; r < poly.NumRings(); r++ {
		ring := poly.Ring(r)
		for i := 0; i+1 < len(ring); i++ {
			if ip, ok := k.SegmentIntersection(p, q, ring[i], ring[i+1]); ok {
				addParam(ts, segmentParam(p, q, ip))
			}
			if pointOnSegmentXY(ring[i], p, q, k) {
				addParam(ts, segmentParam(p, q, ring[i]))
			}
			if pointOnSegmentXY(ring[i+1], p, q, k) {
				addParam(ts, segmentParam(p, q, ring[i+1]))
			}
		}
	}
}

func pointOnSegmentXY(p, a, b geom.XY, k kernel.Kernel) bool {
	return k.SegmentDistance(p, a, b) == 0
}

func interpolate(a, b geom.XY, t float64) geom.XY {
	return geom.XY{X: a.X + (b.X-a.X)*t, Y: a.Y + (b.Y-a.Y)*t}
}

func segmentParam(a, b, p geom.XY) float64 {
	dx, dy := b.X-a.X, b.Y-a.Y
	if dx*dx >= dy*dy {
		if dx == 0 {
			return 0
		}
		return (p.X - a.X) / dx
	}
	if dy == 0 {
		return 0
	}
	return (p.Y - a.Y) / dy
}

func addParam(ts *[]float64, t float64) {
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	for _, existing := range *ts {
		if math.Abs(existing-t) <= 1e-12 {
			return
		}
	}
	*ts = append(*ts, t)
}

func collectionPolygonCoveredInterior(a *geom.GeometryCollection, poly *geom.Polygon, k kernel.Kernel) (bool, bool) {
	if poly.IsEmpty() || poly.NumRings() == 0 {
		return false, false
	}
	interior := false
	for r := 0; r < poly.NumRings(); r++ {
		ring := poly.Ring(r)
		for i := 0; i+1 < len(ring); i++ {
			covered, hit := collectionPointCoveredInterior(a, ring[i], k)
			if !covered {
				return false, false
			}
			interior = interior || hit
			mid := geom.XY{X: (ring[i].X + ring[i+1].X) / 2, Y: (ring[i].Y + ring[i+1].Y) / 2}
			covered, hit = collectionPointCoveredInterior(a, mid, k)
			if !covered {
				return false, false
			}
			interior = interior || hit
		}
	}
	if len(poly.Ring(0)) > 0 {
		covered, hit := collectionPointCoveredInterior(a, poly.Ring(0)[0], k)
		interior = interior || (covered && hit)
	}
	// Interior representative: a point strictly inside poly. The collection
	// must cover this point too — otherwise poly's interior contains a
	// region in the collection's exterior. Without this check, a polygon
	// that "wraps around" gaps between collection members would falsely be
	// reported as contained.
	rep := samplePoint(poly)
	if rep != (geom.XY{}) {
		covered, hit := collectionPointCoveredInterior(a, rep, k)
		if !covered {
			return false, false
		}
		interior = interior || hit
	}
	return true, interior
}

func collectionPointCoveredInterior(a *geom.GeometryCollection, p geom.XY, k kernel.Kernel) (covered bool, interior bool) {
	polygonBoundaryHits := 0
	for i := 0; i < a.NumGeometries(); i++ {
		g := a.GeometryAt(i)
		if g.IsEmpty() {
			continue
		}
		c, hit, boundary := pointCoveredInteriorByGeometry(g, p, k)
		covered = covered || c
		interior = interior || hit
		if boundary {
			polygonBoundaryHits++
		}
	}
	if polygonBoundaryHits >= 2 && collectionPointNeighborhoodCovered(a, p, k) {
		interior = true
	}
	return covered, interior
}

func collectionPointNeighborhoodCovered(a *geom.GeometryCollection, p geom.XY, k kernel.Kernel) bool {
	env := a.Envelope()
	scale := math.Max(env.Width(), env.Height())
	eps := scale * 1e-9
	if eps == 0 {
		eps = 1e-9
	}
	for _, d := range []geom.XY{
		{X: eps}, {X: -eps}, {Y: eps}, {Y: -eps},
		{X: eps, Y: eps}, {X: eps, Y: -eps}, {X: -eps, Y: eps}, {X: -eps, Y: -eps},
	} {
		if !collectionPointCoveredOnly(a, geom.XY{X: p.X + d.X, Y: p.Y + d.Y}, k) {
			return false
		}
	}
	return true
}

func collectionPointCoveredOnly(a *geom.GeometryCollection, p geom.XY, k kernel.Kernel) bool {
	for i := 0; i < a.NumGeometries(); i++ {
		covered, _, _ := pointCoveredInteriorByGeometry(a.GeometryAt(i), p, k)
		if covered {
			return true
		}
	}
	return false
}

func pointCoveredInteriorByGeometry(g geom.Geometry, p geom.XY, k kernel.Kernel) (covered bool, interior bool, polygonBoundary bool) {
	switch v := g.(type) {
	case *geom.Point:
		return !v.IsEmpty() && v.XY() == p, !v.IsEmpty() && v.XY() == p, false
	case *geom.MultiPoint:
		for i := 0; i < v.NumGeometries(); i++ {
			if v.PointAt(i) == p {
				return true, true, false
			}
		}
	case *geom.LineString:
		if pointOnLine(p, v, k) {
			b := lineBoundary(v)
			return true, !pointInBoundarySet(p, b), false
		}
	case *geom.MultiLineString:
		for i := 0; i < v.NumGeometries(); i++ {
			c, hit, _ := pointCoveredInteriorByGeometry(v.LineStringAt(i), p, k)
			if c {
				return true, hit, false
			}
		}
	case *geom.Polygon:
		c := pointInPolygon(p, v, k)
		return c != kernel.Outside, c == kernel.Inside, c == kernel.OnBoundary
	case *geom.MultiPolygon:
		boundaryHits := 0
		for i := 0; i < v.NumGeometries(); i++ {
			c := pointInPolygon(p, v.PolygonAt(i), k)
			if c == kernel.Inside {
				return true, true, false
			}
			if c == kernel.OnBoundary {
				covered = true
				boundaryHits++
			}
		}
		return covered, boundaryHits >= 2, boundaryHits > 0
	case *geom.GeometryCollection:
		c, hit := collectionPointCoveredInterior(v, p, k)
		return c, hit, false
	}
	return false, false, false
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
	case *geom.MultiPoint:
		// Contains(poly, MP): every point of MP must be in poly's
		// CLOSURE, AND at least one must be strictly inside.
		anyInside := false
		for i := 0; i < vb.NumGeometries(); i++ {
			c := pointInPolygon(vb.PointAt(i), a, k)
			if c == kernel.Outside {
				return false, nil
			}
			if c == kernel.Inside {
				anyInside = true
			}
		}
		return anyInside, nil
	case *geom.MultiLineString:
		// Every member must be covered by the polygon, and at least one
		// member interior must intersect the polygon interior.
		anyInterior := false
		for i := 0; i < vb.NumGeometries(); i++ {
			covered, interior := polygonLineInClosure(a, vb.LineStringAt(i), k)
			if !covered {
				return false, nil
			}
			anyInterior = anyInterior || interior
		}
		return anyInterior, nil
	case *geom.MultiPolygon:
		anyInterior := false
		for i := 0; i < vb.NumGeometries(); i++ {
			covered, interior := polygonCoversWithInteriorHit(a, vb.PolygonAt(i), k)
			if !covered {
				return false, nil
			}
			anyInterior = anyInterior || interior
		}
		return anyInterior, nil
	case *geom.GeometryCollection:
		anyInterior := false
		seen := false
		for i := 0; i < vb.NumGeometries(); i++ {
			g := vb.GeometryAt(i)
			if g.IsEmpty() {
				continue
			}
			seen = true
			covered, interior := polygonCoversWithInteriorHit(a, g, k)
			if !covered {
				return false, nil
			}
			anyInterior = anyInterior || interior
		}
		return seen && anyInterior, nil
	case *geom.LineString:
		covered, interior := polygonLineInClosure(a, vb, k)
		return covered && interior, nil
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
		for r := 1; r < a.NumRings(); r++ {
			hole := a.Ring(r)
			if polygonHasEquivalentHole(vb, hole) {
				continue
			}
			for i := 0; i+1 < len(hole); i++ {
				if pointInPolygon(hole[i], vb, k) != kernel.Outside {
					return false, nil
				}
			}
		}
		return true, nil
	}
	return false, nil
}

func polygonHasEquivalentHole(p *geom.Polygon, hole []geom.XY) bool {
	for r := 1; r < p.NumRings(); r++ {
		if ringEquivalentXY(hole, p.Ring(r)) {
			return true
		}
	}
	return false
}

func ringEquivalentXY(a, b []geom.XY) bool {
	if len(a) != len(b) || len(a) < 4 {
		return false
	}
	n := len(a) - 1
	if len(b)-1 != n {
		return false
	}
	for off := 0; off < n; off++ {
		match := true
		for i := 0; i < n; i++ {
			if a[i] != b[(i+off)%n] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	for off := 0; off < n; off++ {
		match := true
		for i := 0; i < n; i++ {
			if a[i] != b[(off-i+n*n)%n] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func polygonCoversWithInteriorHit(a *geom.Polygon, b geom.Geometry, k kernel.Kernel) (covered bool, interior bool) {
	switch vb := b.(type) {
	case *geom.Point:
		c := pointInPolygon(vb.XY(), a, k)
		return c != kernel.Outside, c == kernel.Inside
	case *geom.MultiPoint:
		seen := false
		for i := 0; i < vb.NumGeometries(); i++ {
			p := vb.PointAt(i)
			c := pointInPolygon(p, a, k)
			if c == kernel.Outside {
				return false, false
			}
			seen = true
			interior = interior || c == kernel.Inside
		}
		return seen, interior
	case *geom.LineString:
		return polygonLineInClosure(a, vb, k)
	case *geom.MultiLineString:
		seen := false
		for i := 0; i < vb.NumGeometries(); i++ {
			covered, hit := polygonLineInClosure(a, vb.LineStringAt(i), k)
			if !covered {
				return false, false
			}
			seen = true
			interior = interior || hit
		}
		return seen, interior
	case *geom.Polygon:
		ok, err := polygonContains(a, vb, k)
		if err != nil {
			return false, false
		}
		return ok, ok
	case *geom.MultiPolygon:
		seen := false
		for i := 0; i < vb.NumGeometries(); i++ {
			covered, hit := polygonCoversWithInteriorHit(a, vb.PolygonAt(i), k)
			if !covered {
				return false, false
			}
			seen = true
			interior = interior || hit
		}
		return seen, interior
	case *geom.GeometryCollection:
		seen := false
		for i := 0; i < vb.NumGeometries(); i++ {
			g := vb.GeometryAt(i)
			if g.IsEmpty() {
				continue
			}
			covered, hit := polygonCoversWithInteriorHit(a, g, k)
			if !covered {
				return false, false
			}
			seen = true
			interior = interior || hit
		}
		return seen, interior
	}
	return false, false
}

func polygonLineInClosure(a *geom.Polygon, line *geom.LineString, k kernel.Kernel) (covered bool, interior bool) {
	n := line.NumPoints()
	for i := 0; i < n; i++ {
		switch pointInPolygon(line.PointAt(i), a, k) {
		case kernel.Outside:
			return false, false
		case kernel.Inside:
			interior = true
		}
	}
	for r := 0; r < a.NumRings(); r++ {
		ring := a.Ring(r)
		for i := 0; i+1 < n; i++ {
			p1, p2 := line.PointAt(i), line.PointAt(i+1)
			for j := 0; j+1 < len(ring); j++ {
				ip, ok := k.SegmentIntersection(p1, p2, ring[j], ring[j+1])
				if !ok {
					continue
				}
				// Touching at a vertex is fine; only proper crossings disqualify.
				if ip != p1 && ip != p2 && ip != ring[j] && ip != ring[j+1] {
					return false, false
				}
			}
		}
	}
	return true, interior
}
