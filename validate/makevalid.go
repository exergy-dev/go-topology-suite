package validate

import (
	terra "github.com/terra-geo/terra"
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/kernel/planar"
	"github.com/terra-geo/terra/overlay"
)

// MakeValid returns a topologically valid geometry that approximates g.
// Uses overlay.Union(g, g) to snap-round and clean rings; ensures ring
// closure and consistent CCW outer / CW hole orientation.
//
// v0.1 limitations: holes are dropped from polygons (overlay-NG port
// restores them), and self-intersecting input may simplify in
// unexpected ways. Documented per-call.
func MakeValid(g geom.Geometry) (geom.Geometry, error) {
	if g == nil {
		return nil, terra.ErrEmpty
	}
	if g.IsEmpty() {
		return nil, terra.ErrEmpty
	}
	switch x := g.(type) {
	case *geom.Point:
		// Always valid (empty handled above).
		return x, nil
	case *geom.LineString:
		return makeValidLineString(x), nil
	case *geom.Polygon:
		return makeValidPolygon(x), nil
	case *geom.MultiPoint:
		// MultiPoint is always structurally valid (members are coordinates).
		return x, nil
	case *geom.MultiLineString:
		return makeValidMultiLineString(x), nil
	case *geom.MultiPolygon:
		return makeValidMultiPolygon(x), nil
	case *geom.GeometryCollection:
		return makeValidCollection(x), nil
	default:
		// Unknown concrete type: pass through.
		return g, nil
	}
}

// makeValidLineString ensures the line has at least two distinct vertices.
// Adjacent duplicates are collapsed; if fewer than two unique vertices
// remain, the result degrades to a Point. An originally well-formed line
// is returned with duplicates removed (which is still valid, never empty).
func makeValidLineString(ls *geom.LineString) geom.Geometry {
	pts := collectPoints(ls)
	dedup := collapseAdjacentDuplicates(pts)
	if len(dedup) < 2 {
		// Degrade to a Point at the only remaining vertex.
		if len(dedup) == 1 {
			return geom.NewPoint(ls.CRS(), dedup[0])
		}
		// All vertices were duplicates of nothing — construct from first input
		// vertex if present (we guaranteed non-empty at entry).
		return geom.NewPoint(ls.CRS(), pts[0])
	}
	return geom.NewLineString(ls.CRS(), dedup)
}

func collectPoints(ls *geom.LineString) []geom.XY {
	out := make([]geom.XY, 0, ls.NumPoints())
	for p := range ls.CoordsXY() {
		out = append(out, p)
	}
	return out
}

func collapseAdjacentDuplicates(pts []geom.XY) []geom.XY {
	if len(pts) == 0 {
		return pts
	}
	out := make([]geom.XY, 0, len(pts))
	out = append(out, pts[0])
	for i := 1; i < len(pts); i++ {
		if pts[i] != out[len(out)-1] {
			out = append(out, pts[i])
		}
	}
	return out
}

// makeValidPolygon corrects ring closure, orientation, and (if possible)
// self-intersection in the outer ring. Holes are dropped per v0.1
// limitations.
func makeValidPolygon(p *geom.Polygon) geom.Geometry {
	if p.NumRings() == 0 {
		return geom.NewEmptyPolygon(p.CRS(), geom.LayoutXY)
	}
	outer := closeRing(p.ExteriorRing())
	if len(outer) < 4 {
		return geom.NewEmptyPolygon(p.CRS(), geom.LayoutXY)
	}
	outer = orientCCW(outer)

	// Try snap-rounding via Union(g, g) when the outer ring self-intersects.
	if _, hit := ringSelfIntersection(outer); hit {
		clean := geom.NewPolygon(p.CRS(), outer)
		if cleaned, err := overlay.Union(clean, clean); err == nil && cleaned != nil && !cleaned.IsEmpty() {
			// Ensure orientation on returned polygon(s).
			return reorientResult(cleaned)
		}
		// Fall through with the structurally-corrected (closed, CCW) polygon
		// but no intersection cleaning.
	}
	return geom.NewPolygon(p.CRS(), outer)
}

// closeRing returns ring with its first vertex appended if not already
// closed. Adjacent duplicates within the ring are collapsed first.
func closeRing(ring []geom.XY) []geom.XY {
	r := collapseAdjacentDuplicates(ring)
	if len(r) == 0 {
		return r
	}
	if r[0] != r[len(r)-1] {
		r = append(r, r[0])
	}
	return r
}

// orientCCW returns ring as CCW (positive shoelace area). Hole orientation
// is the reverse — see orientCW.
func orientCCW(ring []geom.XY) []geom.XY {
	if planar.Default.RingArea(ring) < 0 {
		return reverseRing(ring)
	}
	return ring
}

func reverseRing(r []geom.XY) []geom.XY {
	out := make([]geom.XY, len(r))
	for i := range r {
		out[i] = r[len(r)-1-i]
	}
	return out
}

// reorientResult walks a Polygon/MultiPolygon result and forces every outer
// ring to CCW.
func reorientResult(g geom.Geometry) geom.Geometry {
	switch x := g.(type) {
	case *geom.Polygon:
		if x.IsEmpty() || x.NumRings() == 0 {
			return x
		}
		outer := orientCCW(closeRing(x.ExteriorRing()))
		if len(outer) < 4 {
			return geom.NewEmptyPolygon(x.CRS(), geom.LayoutXY)
		}
		return geom.NewPolygon(x.CRS(), outer)
	case *geom.MultiPolygon:
		parts := make([]*geom.Polygon, 0, x.NumGeometries())
		for i := 0; i < x.NumGeometries(); i++ {
			r := reorientResult(x.PolygonAt(i))
			if poly, ok := r.(*geom.Polygon); ok && !poly.IsEmpty() {
				parts = append(parts, poly)
			}
		}
		if len(parts) == 0 {
			return geom.NewEmptyPolygon(x.CRS(), geom.LayoutXY)
		}
		if len(parts) == 1 {
			return parts[0]
		}
		return geom.NewMultiPolygon(x.CRS(), parts...)
	}
	return g
}

func makeValidMultiLineString(m *geom.MultiLineString) geom.Geometry {
	parts := make([]*geom.LineString, 0, m.NumGeometries())
	for i := 0; i < m.NumGeometries(); i++ {
		ls := m.LineStringAt(i)
		if ls.IsEmpty() {
			continue
		}
		v := makeValidLineString(ls)
		if v == nil || v.IsEmpty() {
			continue
		}
		// Result may have degraded to a Point — drop those (caller can
		// inspect via GeometryCollection variant if they need them).
		if line, ok := v.(*geom.LineString); ok {
			parts = append(parts, line)
		}
	}
	if len(parts) == 0 {
		return geom.NewMultiLineString(m.CRS())
	}
	return geom.NewMultiLineString(m.CRS(), parts...)
}

func makeValidMultiPolygon(m *geom.MultiPolygon) geom.Geometry {
	parts := make([]*geom.Polygon, 0, m.NumGeometries())
	for i := 0; i < m.NumGeometries(); i++ {
		p := m.PolygonAt(i)
		if p.IsEmpty() {
			continue
		}
		v := makeValidPolygon(p)
		switch r := v.(type) {
		case *geom.Polygon:
			if !r.IsEmpty() {
				parts = append(parts, r)
			}
		case *geom.MultiPolygon:
			for j := 0; j < r.NumGeometries(); j++ {
				if q := r.PolygonAt(j); !q.IsEmpty() {
					parts = append(parts, q)
				}
			}
		}
	}
	if len(parts) == 0 {
		return geom.NewMultiPolygon(m.CRS())
	}
	return geom.NewMultiPolygon(m.CRS(), parts...)
}

func makeValidCollection(c *geom.GeometryCollection) geom.Geometry {
	parts := make([]geom.Geometry, 0, c.NumGeometries())
	for i := 0; i < c.NumGeometries(); i++ {
		child := c.GeometryAt(i)
		if child == nil || child.IsEmpty() {
			continue
		}
		v, err := MakeValid(child)
		if err != nil || v == nil || v.IsEmpty() {
			continue
		}
		parts = append(parts, v)
	}
	return geom.NewGeometryCollection(c.CRS(), parts...)
}
