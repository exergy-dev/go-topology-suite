package geom

import "github.com/terra-geo/terra/crs"

// WithCRS returns a copy of g whose CRS pointer is c. The coordinate data
// itself is shared with g (no deep copy); callers must not mutate the
// underlying buffers afterwards. Empty g returns g unchanged. Nil g
// returns nil.
//
// This helper exists so packages outside geom (notably crs.Transform) can
// rebrand a geometry's CRS after editing its coordinates, without taking
// on a public SetCRS mutator that would compromise the read-only-after-
// construction guarantee.
func WithCRS(g Geometry, c *crs.CRS) Geometry {
	if g == nil {
		return nil
	}
	switch v := g.(type) {
	case *Point:
		// Reconstruct baseGeom field-by-field to avoid copying the
		// sync/atomic.Pointer envelope cache (which carries noCopy).
		// Mirrors the LineString/LinearRing/MultiPoint arms below.
		return &Point{baseGeom: baseGeom{layout: v.layout, coords: v.coords, crs: c}}
	case *LineString:
		return &LineString{baseGeom{layout: v.layout, coords: v.coords, crs: c}}
	case *LinearRing:
		return &LinearRing{baseGeom{layout: v.layout, coords: v.coords, crs: c}}
	case *Polygon:
		starts := make([]int, len(v.ringStarts))
		copy(starts, v.ringStarts)
		return &Polygon{
			baseGeom:   baseGeom{layout: v.layout, coords: v.coords, crs: c},
			ringStarts: starts,
		}
	case *MultiPoint:
		return &MultiPoint{baseGeom{layout: v.layout, coords: v.coords, crs: c}}
	case *MultiLineString:
		parts := make([]*LineString, len(v.parts))
		for i, p := range v.parts {
			parts[i] = WithCRS(p, c).(*LineString)
		}
		return &MultiLineString{layout: v.layout, crs: c, parts: parts}
	case *MultiPolygon:
		parts := make([]*Polygon, len(v.parts))
		for i, p := range v.parts {
			parts[i] = WithCRS(p, c).(*Polygon)
		}
		return &MultiPolygon{layout: v.layout, crs: c, parts: parts}
	case *GeometryCollection:
		parts := make([]Geometry, len(v.parts))
		for i, child := range v.parts {
			parts[i] = WithCRS(child, c)
		}
		return &GeometryCollection{layout: v.layout, crs: c, parts: parts}
	}
	return g
}
