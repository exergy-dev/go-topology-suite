package geom

import (
	"sync/atomic"

	"github.com/terra-geo/terra/crs"
)

// MultiPoint is an unordered collection of points. It uses the same flat
// storage as LineString since every member is a single coordinate.
type MultiPoint struct {
	baseGeom
}

// NewMultiPoint constructs a MultiPoint from XY coordinates.
func NewMultiPoint(c *crs.CRS, pts []XY) *MultiPoint {
	flat := make([]float64, 0, 2*len(pts))
	for _, p := range pts {
		flat = append(flat, p.X, p.Y)
	}
	return &MultiPoint{baseGeom{layout: LayoutXY, coords: flat, crs: c}}
}

func (mp *MultiPoint) isGeometry()       {}
func (mp *MultiPoint) Type() Type        { return MultiPointType }
func (mp *MultiPoint) Envelope() Envelope  { return mp.envelope() }
func (mp *MultiPoint) IsEmpty() bool       { return len(mp.coords) == 0 }
func (mp *MultiPoint) NumGeometries() int  { return mp.numCoords() }

// PointAt returns the i-th point projected to XY.
func (mp *MultiPoint) PointAt(i int) XY {
	stride := mp.stride()
	off := i * stride
	return XY{mp.coords[off], mp.coords[off+1]}
}

// MultiLineString is a collection of LineStrings.
type MultiLineString struct {
	layout Layout
	crs    *crs.CRS
	parts  []*LineString
	env    atomic.Pointer[Envelope]
}

// NewMultiLineString constructs from a slice of LineStrings. CRS and layout
// are taken from the first member; mismatched layouts/CRSes among members
// are not checked at construction time (callers requiring strict input
// should validate).
func NewMultiLineString(c *crs.CRS, parts ...*LineString) *MultiLineString {
	layout := LayoutXY
	if len(parts) > 0 {
		layout = parts[0].Layout()
	}
	return &MultiLineString{layout: layout, crs: c, parts: parts}
}

func (m *MultiLineString) isGeometry()      {}
func (m *MultiLineString) Type() Type       { return MultiLineStringType }
func (m *MultiLineString) Layout() Layout   { return m.layout }
func (m *MultiLineString) CRS() *crs.CRS    { return m.crs }
func (m *MultiLineString) IsEmpty() bool    { return len(m.parts) == 0 }
func (m *MultiLineString) NumGeometries() int { return len(m.parts) }

// LineStringAt returns the i-th member.
func (m *MultiLineString) LineStringAt(i int) *LineString { return m.parts[i] }

// Envelope returns the union of member envelopes (cached).
func (m *MultiLineString) Envelope() Envelope {
	return cachedUnionEnvelope(&m.env, func(yield func(Envelope) bool) {
		for _, p := range m.parts {
			if !yield(p.Envelope()) {
				return
			}
		}
	})
}

// MultiPolygon is a collection of Polygons.
type MultiPolygon struct {
	layout Layout
	crs    *crs.CRS
	parts  []*Polygon
	env    atomic.Pointer[Envelope]
}

// NewMultiPolygon constructs from a slice of Polygons.
func NewMultiPolygon(c *crs.CRS, parts ...*Polygon) *MultiPolygon {
	layout := LayoutXY
	if len(parts) > 0 {
		layout = parts[0].Layout()
	}
	return &MultiPolygon{layout: layout, crs: c, parts: parts}
}

func (m *MultiPolygon) isGeometry()      {}
func (m *MultiPolygon) Type() Type       { return MultiPolygonType }
func (m *MultiPolygon) Layout() Layout   { return m.layout }
func (m *MultiPolygon) CRS() *crs.CRS    { return m.crs }
func (m *MultiPolygon) IsEmpty() bool    { return len(m.parts) == 0 }
func (m *MultiPolygon) NumGeometries() int { return len(m.parts) }
func (m *MultiPolygon) PolygonAt(i int) *Polygon { return m.parts[i] }

func (m *MultiPolygon) Envelope() Envelope {
	return cachedUnionEnvelope(&m.env, func(yield func(Envelope) bool) {
		for _, p := range m.parts {
			if !yield(p.Envelope()) {
				return
			}
		}
	})
}

// GeometryCollection is a heterogeneous collection of geometries.
type GeometryCollection struct {
	layout Layout
	crs    *crs.CRS
	parts  []Geometry
	env    atomic.Pointer[Envelope]
}

// NewGeometryCollection constructs from a slice of arbitrary geometries.
func NewGeometryCollection(c *crs.CRS, parts ...Geometry) *GeometryCollection {
	layout := LayoutXY
	if len(parts) > 0 {
		layout = parts[0].Layout()
	}
	return &GeometryCollection{layout: layout, crs: c, parts: parts}
}

func (g *GeometryCollection) isGeometry()      {}
func (g *GeometryCollection) Type() Type       { return GeometryCollectionType }
func (g *GeometryCollection) Layout() Layout   { return g.layout }
func (g *GeometryCollection) CRS() *crs.CRS    { return g.crs }
func (g *GeometryCollection) IsEmpty() bool    { return len(g.parts) == 0 }
func (g *GeometryCollection) NumGeometries() int { return len(g.parts) }
func (g *GeometryCollection) GeometryAt(i int) Geometry { return g.parts[i] }

func (g *GeometryCollection) Envelope() Envelope {
	return cachedUnionEnvelope(&g.env, func(yield func(Envelope) bool) {
		for _, p := range g.parts {
			if !yield(p.Envelope()) {
				return
			}
		}
	})
}

// cachedUnionEnvelope is the shared lazy-init helper for collection types.
// It mirrors baseGeom.envelope() but accepts an iterator over child
// envelopes so we don't duplicate the CAS dance per concrete type.
func cachedUnionEnvelope(slot *atomic.Pointer[Envelope], children func(yield func(Envelope) bool)) Envelope {
	if e := slot.Load(); e != nil {
		return *e
	}
	out := EmptyEnvelope()
	children(func(c Envelope) bool {
		out = out.ExpandToInclude(c)
		return true
	})
	slot.CompareAndSwap(nil, &out)
	if e := slot.Load(); e != nil {
		return *e
	}
	return out
}
