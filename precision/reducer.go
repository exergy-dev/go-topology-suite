// Port of org.locationtech.jts.precision.GeometryPrecisionReducer.
//
// Reduces a geometry's coordinates to the grid defined by a PrecisionModel.
// Coordinates are snapped via PrecisionModel.MakePrecise; collapsed rings
// and degenerate lines may be dropped from the output depending on
// configuration.

package precision

import (
	"github.com/exergy-dev/go-topology-suite/geom"
)

// PrecisionReducer reduces the precision of a geometry to a target
// PrecisionModel. By default it removes collapsed rings/lines and keeps
// the resulting structural type — a Polygon stays a Polygon (possibly
// with fewer holes), a LineString stays a LineString unless it collapses
// to a single point in which case the polygon/component is dropped.
//
// Mirrors org.locationtech.jts.precision.GeometryPrecisionReducer.
type PrecisionReducer struct {
	pm geom.PrecisionModel
	// RemoveCollapsed controls whether collapsed components (rings with
	// fewer than 4 distinct vertices, lines collapsing to a single
	// point) are removed from the output. JTS default is true.
	RemoveCollapsed bool
	// Pointwise, when true, applies the precision model to each
	// coordinate independently without performing topology cleanup
	// (i.e. the output may be invalid). JTS default is false.
	Pointwise bool
}

// NewPrecisionReducer returns a reducer for the given precision model
// with JTS defaults: RemoveCollapsed=true, Pointwise=false.
func NewPrecisionReducer(pm geom.PrecisionModel) *PrecisionReducer {
	return &PrecisionReducer{pm: pm, RemoveCollapsed: true}
}

// Reduce snaps every coordinate of g to pm and returns the resulting
// geometry. This is the convenience entry point matching JTS's static
// GeometryPrecisionReducer.reduce(Geometry, PrecisionModel).
func Reduce(g geom.Geometry, pm geom.PrecisionModel) geom.Geometry {
	return NewPrecisionReducer(pm).Reduce(g)
}

// ReducePointwise applies the precision model to every coordinate
// without collapse removal or topology cleanup. Mirrors
// GeometryPrecisionReducer.reducePointwise.
func ReducePointwise(g geom.Geometry, pm geom.PrecisionModel) geom.Geometry {
	r := NewPrecisionReducer(pm)
	r.Pointwise = true
	r.RemoveCollapsed = false
	return r.Reduce(g)
}

// Reduce returns a new geometry with coordinates snapped to the
// configured PrecisionModel. The original geometry is not modified.
//
// Behaviour for Pointwise=false (default):
//
//   - Polygon rings that collapse below 4 distinct vertices are
//     dropped (holes only) or, if the shell collapses, the entire
//     polygon is replaced with an empty geometry of the same type.
//   - LineStrings that collapse to a single distinct vertex are
//     dropped from a parent MultiLineString / GeometryCollection, or
//     replaced with an empty LineString at the top level.
//
// Behaviour for Pointwise=true: every coordinate is snapped, but no
// collapse removal is performed. The result may contain degenerate
// components.
func (r *PrecisionReducer) Reduce(g geom.Geometry) geom.Geometry {
	if g == nil {
		return nil
	}
	if r.pm.IsFloating() {
		// No rounding required; return a coordinate-cloned copy via
		// Edit (so callers can rely on the result being independent).
		return geom.Edit(g, func(p geom.XY) geom.XY { return p })
	}
	if r.Pointwise {
		return geom.Edit(g, func(p geom.XY) geom.XY {
			return r.pm.MakePrecise(p)
		})
	}
	return r.reduce(g)
}

// reduce implements the structure-aware variant: snap every vertex,
// then dedupe immediate duplicates (so rings/lines that fold flat
// expose their collapse), then drop sub-components that no longer have
// enough distinct vertices to be valid.
func (r *PrecisionReducer) reduce(g geom.Geometry) geom.Geometry {
	switch v := g.(type) {
	case *geom.Point:
		return r.reducePoint(v)
	case *geom.LineString:
		return r.reduceLineString(v)
	case *geom.LinearRing:
		ls := r.reduceLineString(v.AsLineString())
		// Re-promote to LinearRing if it remained closed and valid.
		if ls.IsEmpty() || ls.NumPoints() < 4 {
			return geom.NewLineString(v.CRS(), nil)
		}
		return ls
	case *geom.Polygon:
		return r.reducePolygon(v)
	case *geom.MultiPoint:
		return r.reduceMultiPoint(v)
	case *geom.MultiLineString:
		return r.reduceMultiLineString(v)
	case *geom.MultiPolygon:
		return r.reduceMultiPolygon(v)
	case *geom.GeometryCollection:
		return r.reduceCollection(v)
	}
	return g
}

func (r *PrecisionReducer) reducePoint(p *geom.Point) *geom.Point {
	if p.IsEmpty() {
		return p
	}
	xy := r.pm.MakePrecise(p.XY())
	return geom.NewPoint(p.CRS(), xy)
}

func (r *PrecisionReducer) reduceLineString(ls *geom.LineString) *geom.LineString {
	pts := snapAndDedup(linePoints(ls), r.pm)
	if r.RemoveCollapsed && len(pts) < 2 {
		return geom.NewLineString(ls.CRS(), nil)
	}
	if len(pts) < 2 {
		// Pad a duplicate to avoid panicking constructors that
		// expect ≥ 2 vertices when not removing collapses.
		if len(pts) == 1 {
			pts = append(pts, pts[0])
		} else {
			return geom.NewLineString(ls.CRS(), nil)
		}
	}
	return geom.NewLineString(ls.CRS(), pts)
}

func (r *PrecisionReducer) reducePolygon(p *geom.Polygon) *geom.Polygon {
	if p.IsEmpty() {
		return p
	}
	shell := snapAndDedup(p.Ring(0), r.pm)
	if !ringIsValid(shell) {
		return geom.NewEmptyPolygon(p.CRS(), p.Layout())
	}
	rings := [][]geom.XY{shell}
	for i := 1; i < p.NumRings(); i++ {
		hole := snapAndDedup(p.Ring(i), r.pm)
		if !ringIsValid(hole) {
			if r.RemoveCollapsed {
				continue
			}
		}
		rings = append(rings, hole)
	}
	return geom.NewPolygon(p.CRS(), rings...)
}

func (r *PrecisionReducer) reduceMultiPoint(mp *geom.MultiPoint) *geom.MultiPoint {
	pts := make([]geom.XY, 0, mp.NumGeometries())
	for i := 0; i < mp.NumGeometries(); i++ {
		pts = append(pts, r.pm.MakePrecise(mp.PointAt(i)))
	}
	return geom.NewMultiPoint(mp.CRS(), pts)
}

func (r *PrecisionReducer) reduceMultiLineString(m *geom.MultiLineString) *geom.MultiLineString {
	parts := make([]*geom.LineString, 0, m.NumGeometries())
	for i := 0; i < m.NumGeometries(); i++ {
		ls := r.reduceLineString(m.LineStringAt(i))
		if ls.IsEmpty() {
			continue
		}
		parts = append(parts, ls)
	}
	return geom.NewMultiLineString(m.CRS(), parts...)
}

func (r *PrecisionReducer) reduceMultiPolygon(m *geom.MultiPolygon) *geom.MultiPolygon {
	parts := make([]*geom.Polygon, 0, m.NumGeometries())
	for i := 0; i < m.NumGeometries(); i++ {
		p := r.reducePolygon(m.PolygonAt(i))
		if p.IsEmpty() {
			continue
		}
		parts = append(parts, p)
	}
	return geom.NewMultiPolygon(m.CRS(), parts...)
}

func (r *PrecisionReducer) reduceCollection(gc *geom.GeometryCollection) *geom.GeometryCollection {
	parts := make([]geom.Geometry, 0, gc.NumGeometries())
	for i := 0; i < gc.NumGeometries(); i++ {
		child := r.reduce(gc.GeometryAt(i))
		if child == nil || child.IsEmpty() {
			continue
		}
		parts = append(parts, child)
	}
	return geom.NewGeometryCollection(gc.CRS(), parts...)
}

// linePoints extracts the XY vertices of a LineString.
func linePoints(ls *geom.LineString) []geom.XY {
	n := ls.NumPoints()
	out := make([]geom.XY, n)
	for i := 0; i < n; i++ {
		out[i] = ls.PointAt(i)
	}
	return out
}

// snapAndDedup snaps every coordinate to pm and removes immediate
// duplicates introduced by the snapping (or already present in the
// input). The first and last vertices are preserved exactly so callers
// can detect collapsed rings (first==last with fewer than 4 entries).
func snapAndDedup(pts []geom.XY, pm geom.PrecisionModel) []geom.XY {
	if len(pts) == 0 {
		return nil
	}
	out := make([]geom.XY, 0, len(pts))
	for _, p := range pts {
		q := pm.MakePrecise(p)
		if len(out) > 0 && out[len(out)-1].EqualBitwise(q) {
			continue
		}
		out = append(out, q)
	}
	return out
}

// ringIsValid reports whether a snapped ring still has at least four
// distinct-once-deduped vertices (the minimum for a closed simple ring).
// JTS uses CoordinateArrays.hasRepeatedPoints + min-vertex-count
// equivalents; this combined check captures both.
func ringIsValid(ring []geom.XY) bool {
	if len(ring) < 4 {
		return false
	}
	// Distinct-vertex count: first == last (closed), so subtract 1.
	distinct := 0
	if len(ring) > 0 {
		distinct = 1
		for i := 1; i < len(ring); i++ {
			if !ring[i].EqualBitwise(ring[i-1]) {
				distinct++
			}
		}
	}
	return distinct >= 4
}
