package simplify

import (
	"math"

	"github.com/terra-geo/terra/geom"
)

// Simplify returns a Douglas-Peucker simplification of g with the given
// tolerance (perpendicular distance in the geometry's coordinate units).
// A tolerance ≤ 0 returns g unchanged.
func Simplify(g geom.Geometry, tolerance float64) geom.Geometry {
	if tolerance <= 0 || g.IsEmpty() {
		return g
	}
	switch v := g.(type) {
	case *geom.Point:
		return v
	case *geom.LineString:
		return simplifyLineString(v, tolerance)
	case *geom.Polygon:
		return simplifyPolygon(v, tolerance)
	case *geom.MultiPoint:
		return v
	case *geom.MultiLineString:
		parts := make([]*geom.LineString, 0, v.NumGeometries())
		for i := 0; i < v.NumGeometries(); i++ {
			part := simplifyLineString(v.LineStringAt(i), tolerance)
			if !part.IsEmpty() {
				parts = append(parts, part)
			}
		}
		return geom.NewMultiLineString(v.CRS(), parts...)
	case *geom.MultiPolygon:
		parts := make([]*geom.Polygon, 0, v.NumGeometries())
		for i := 0; i < v.NumGeometries(); i++ {
			part := simplifyPolygon(v.PolygonAt(i), tolerance)
			if !part.IsEmpty() {
				parts = append(parts, part)
			}
		}
		if len(parts) == 1 {
			return parts[0]
		}
		return geom.NewMultiPolygon(v.CRS(), parts...)
	case *geom.GeometryCollection:
		parts := make([]geom.Geometry, 0, v.NumGeometries())
		for i := 0; i < v.NumGeometries(); i++ {
			parts = append(parts, Simplify(v.GeometryAt(i), tolerance))
		}
		return geom.NewGeometryCollection(v.CRS(), parts...)
	}
	return g
}

func simplifyLineString(ls *geom.LineString, tol float64) *geom.LineString {
	pts := lineToXY(ls)
	out := douglasPeucker(pts, tol)
	return geom.NewLineString(ls.CRS(), out)
}

func simplifyPolygon(p *geom.Polygon, tol float64) *geom.Polygon {
	rings := make([][]geom.XY, 0, p.NumRings())
	for r := 0; r < p.NumRings(); r++ {
		ring := p.Ring(r)
		simplified := douglasPeucker(ring, tol)
		// A polygon ring needs at least 4 distinct vertices (closed). If
		// simplification collapses below that, drop the ring.
		if len(simplified) >= 4 && math.Abs(ringArea2(simplified)) > 0 {
			rings = append(rings, simplified)
		} else if r == 0 {
			return geom.NewEmptyPolygon(p.CRS(), p.Layout())
		}
	}
	return geom.NewPolygon(p.CRS(), rings...)
}

func ringArea2(ring []geom.XY) float64 {
	var a float64
	for i := 0; i+1 < len(ring); i++ {
		a += ring[i].X*ring[i+1].Y - ring[i+1].X*ring[i].Y
	}
	return a
}

func lineToXY(ls *geom.LineString) []geom.XY {
	out := make([]geom.XY, ls.NumPoints())
	for i := range out {
		out[i] = ls.PointAt(i)
	}
	return out
}

// douglasPeucker is the classic recursive simplification.
func douglasPeucker(pts []geom.XY, tol float64) []geom.XY {
	if len(pts) <= 2 {
		return append([]geom.XY(nil), pts...)
	}
	keep := make([]bool, len(pts))
	keep[0] = true
	keep[len(pts)-1] = true
	dpRecurse(pts, 0, len(pts)-1, tol, keep)

	out := make([]geom.XY, 0, len(pts))
	for i, p := range pts {
		if keep[i] {
			out = append(out, p)
		}
	}
	return out
}

func dpRecurse(pts []geom.XY, lo, hi int, tol float64, keep []bool) {
	if hi-lo < 2 {
		return
	}
	a, b := pts[lo], pts[hi]
	maxD := -1.0
	maxI := -1
	for i := lo + 1; i < hi; i++ {
		d := perpDistance(pts[i], a, b)
		if d > maxD {
			maxD = d
			maxI = i
		}
	}
	if maxD > tol {
		keep[maxI] = true
		dpRecurse(pts, lo, maxI, tol, keep)
		dpRecurse(pts, maxI, hi, tol, keep)
	}
}

func perpDistance(p, a, b geom.XY) float64 {
	dx := b.X - a.X
	dy := b.Y - a.Y
	if dx == 0 && dy == 0 {
		return math.Hypot(p.X-a.X, p.Y-a.Y)
	}
	num := math.Abs(dy*p.X - dx*p.Y + b.X*a.Y - b.Y*a.X)
	den := math.Hypot(dx, dy)
	return num / den
}
