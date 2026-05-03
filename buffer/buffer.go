package buffer

import (
	"errors"
	"fmt"
	"math"

	"github.com/terra-geo/terra"
	"github.com/terra-geo/terra/crs"
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/kernel/planar"
	"github.com/terra-geo/terra/overlay"
)

// errGeometryCollectionNotImplemented is returned for GeometryCollection
// inputs. Polygon and MultiPolygon are now supported (see polygon.go); a
// general collection buffer requires per-member dispatch + union, which is
// still pending.
var errGeometryCollectionNotImplemented = errors.New("buffer.Buffer: GeometryCollection input not yet supported")

// Buffer returns the planar buffer of g at the given distance.
//
// See package documentation for the supported geometry types and the
// limitations of v0.1 (notably: polygon inputs are rejected, and the
// result is not unioned across multi-geometry members).
//
// Behavior for special distance values:
//
//   - distance == 0 returns g unchanged.
//   - distance < 0 is only meaningful for polygon inputs (inset buffer);
//     it is rejected with terra.ErrInvalidGeometry for points and lines.
func Buffer(g geom.Geometry, distance float64, opts ...Option) (geom.Geometry, error) {
	if g == nil {
		return nil, terra.ErrInvalidGeometry
	}
	if math.IsNaN(distance) || math.IsInf(distance, 0) {
		return nil, fmt.Errorf("buffer.Buffer: distance must be finite: %w", terra.ErrInvalidGeometry)
	}

	cfg := defaultConfig()
	for _, o := range opts {
		o(&cfg)
	}

	// Treat a LinearRing as a LineString for buffering purposes: the
	// distinct type exists only for OGC validity semantics.
	if lr, ok := g.(*geom.LinearRing); ok {
		g = lr.AsLineString()
	}

	// distance ≤ 0 on Point/Line geometries collapses the geometry to
	// nothing (JTS semantics: buffer of a 0/1-dim with non-positive
	// distance is POLYGON EMPTY). Polygon inputs handle distance == 0
	// as identity in their per-type branches below.
	switch g.(type) {
	case *geom.Point, *geom.LineString,
		*geom.MultiPoint, *geom.MultiLineString:
		if distance <= 0 {
			return geom.NewEmptyPolygon(g.CRS(), geom.LayoutXY), nil
		}
	}
	if distance == 0 {
		// For polygon inputs, buffer(g, 0) is JTS's "polygonal cleanup"
		// — degenerate / zero-area rings collapse to POLYGON EMPTY,
		// otherwise the polygon is returned unchanged.
		switch v := g.(type) {
		case *geom.Polygon:
			if isDegenerateAreal(v) {
				return geom.NewEmptyPolygon(v.CRS(), v.Layout()), nil
			}
		case *geom.MultiPolygon:
			parts := make([]*geom.Polygon, 0, v.NumGeometries())
			for i := 0; i < v.NumGeometries(); i++ {
				pp := v.PolygonAt(i)
				if !isDegenerateAreal(pp) {
					parts = append(parts, pp)
				}
			}
			if len(parts) == 0 {
				return geom.NewEmptyPolygon(v.CRS(), v.Layout()), nil
			}
			return geom.NewMultiPolygon(v.CRS(), parts...), nil
		}
		return g, nil
	}

	switch v := g.(type) {
	case *geom.Point:
		if v.IsEmpty() {
			return geom.NewEmptyPolygon(v.CRS(), v.Layout()), nil
		}
		return bufferPoint(v.CRS(), v.XY(), distance, cfg), nil

	case *geom.LineString:
		if v.IsEmpty() {
			return geom.NewEmptyPolygon(v.CRS(), v.Layout()), nil
		}
		// A closed LineString (LinearRing-style) at positive distance is
		// buffered as an annulus: dilate the enclosed polygon by d and
		// subtract its inset by d. This matches JTS's
		// CLOSED_LINEAR_RING handling.
		if isClosedLine(v) {
			if poly, ok := bufferClosedLineAnnulus(v, distance, cfg); ok {
				return poly, nil
			}
		}
		return bufferLineString(v, distance, cfg)

	case *geom.MultiPoint:
		if v.IsEmpty() {
			return geom.NewMultiPolygon(v.CRS()), nil
		}
		parts := make([]*geom.Polygon, 0, v.NumGeometries())
		for i := 0; i < v.NumGeometries(); i++ {
			parts = append(parts, bufferPoint(v.CRS(), v.PointAt(i), distance, cfg))
		}
		return geom.NewMultiPolygon(v.CRS(), parts...), nil

	case *geom.MultiLineString:
		if v.IsEmpty() {
			return geom.NewMultiPolygon(v.CRS()), nil
		}
		// Buffer each line, then union. Use unionMultiBufferParts which
		// is robust to overlay.Union returning empty/spurious results
		// (a known fragile area when buffer inputs sit at large
		// coordinate magnitudes).
		parts := make([]*geom.Polygon, 0, v.NumGeometries())
		for i := 0; i < v.NumGeometries(); i++ {
			ls := v.LineStringAt(i)
			if ls.IsEmpty() {
				continue
			}
			poly, err := bufferLineString(ls, distance, cfg)
			if err != nil {
				return nil, err
			}
			if poly == nil || poly.IsEmpty() {
				continue
			}
			parts = append(parts, poly)
		}
		if len(parts) == 0 {
			return geom.NewEmptyPolygon(v.CRS(), v.Layout()), nil
		}
		return unionMultiBufferParts(v.CRS(), parts), nil

	case *geom.Polygon:
		if v.IsEmpty() {
			return geom.NewEmptyPolygon(v.CRS(), v.Layout()), nil
		}
		return bufferPolygon(v, distance, cfg)

	case *geom.MultiPolygon:
		if v.IsEmpty() {
			return geom.NewEmptyPolygon(v.CRS(), v.Layout()), nil
		}
		return bufferMultiPolygon(v, distance, cfg)

	case *geom.GeometryCollection:
		return nil, errGeometryCollectionNotImplemented
	}

	return nil, fmt.Errorf("buffer.Buffer: unsupported geometry type %T: %w", g, terra.ErrInvalidGeometry)
}

// isDegenerateAreal reports whether a polygon's outer ring is too
// degenerate to represent any positive-area region: empty, fewer than
// 4 vertices (no closed ring), or zero signed area (all vertices
// collinear or coincident).
func isDegenerateAreal(p *geom.Polygon) bool {
	if p == nil || p.IsEmpty() || p.NumRings() == 0 {
		return true
	}
	outer := p.Ring(0)
	if len(outer) < 4 {
		return true
	}
	return planar.Default.RingArea(outer) == 0
}

// bufferPoint produces a regular polygon approximating a circle of radius
// distance around center. The polygon has exactly 4*quadSegments+1
// vertices (closed ring, last == first).
func bufferPoint(c *crs.CRS, center geom.XY, distance float64, cfg config) *geom.Polygon {
	n := 4 * cfg.quadSegments
	ring := make([]geom.XY, 0, n+1)
	step := 2 * math.Pi / float64(n)
	for i := 0; i < n; i++ {
		theta := float64(i) * step
		ring = append(ring, geom.XY{
			X: center.X + distance*math.Cos(theta),
			Y: center.Y + distance*math.Sin(theta),
		})
	}
	ring = append(ring, ring[0]) // close
	return geom.NewPolygon(c, ring)
}

// bufferLineString produces the offset polygon of ls at distance using cfg.
//
// Algorithm (textbook "thicken"):
//
//  1. Walk forward emitting the LEFT-side parallel offset of each segment,
//     joining at interior vertices per cfg.join.
//  2. Apply the END cap (transition from left side to right side at the
//     final vertex).
//  3. Walk backward emitting the RIGHT-side parallel offset, joining at
//     interior vertices.
//  4. Apply the START cap.
//  5. Close the ring.
//
// For non-self-intersecting input this produces a simple polygon. Self
// intersections at concave corners (where the two offsets overlap) are
// left in place; cleaning them requires the union operation, scheduled
// for Phase 3.
func bufferLineString(ls *geom.LineString, distance float64, cfg config) (*geom.Polygon, error) {
	pts := dedupedPoints(ls)
	if len(pts) == 0 {
		return geom.NewEmptyPolygon(ls.CRS(), ls.Layout()), nil
	}
	if len(pts) == 1 {
		return bufferPoint(ls.CRS(), pts[0], distance, cfg), nil
	}

	// Drop consecutive duplicates more aggressively: dedupedPoints
	// strips exact duplicates, but JTS-style buffering also rejects
	// zero-length input segments before the offset generator sees them.
	clean := make([]geom.XY, 0, len(pts))
	clean = append(clean, pts[0])
	for i := 1; i < len(pts); i++ {
		if pts[i] == clean[len(clean)-1] {
			continue
		}
		clean = append(clean, pts[i])
	}
	if len(clean) == 1 {
		return bufferPoint(ls.CRS(), clean[0], distance, cfg), nil
	}

	d := distance
	g := newOffsetSegmentGenerator(cfg, d)

	// LEFT-side offset traversal: forward through clean[0..n].
	n := len(clean) - 1
	g.initSideSegments(clean[0], clean[1], positionLeft)
	for i := 2; i <= n; i++ {
		g.addNextSegment(clean[i], true)
	}
	g.addLastSegment()
	// End cap: from second-to-last vertex toward the last.
	g.addLineEndCap(clean[n-1], clean[n])

	// RIGHT-side offset traversal: walk clean[n..0] in reverse, with
	// side still LEFT (we're walking the reversed line, so its LEFT is
	// the original's RIGHT).
	g.initSideSegments(clean[n], clean[n-1], positionLeft)
	for i := n - 2; i >= 0; i-- {
		g.addNextSegment(clean[i], true)
	}
	g.addLastSegment()
	// Start cap: from second vertex toward the first.
	g.addLineEndCap(clean[1], clean[0])

	g.closeRing()

	ring := g.coordinates()
	if len(ring) < 4 {
		return geom.NewEmptyPolygon(ls.CRS(), ls.Layout()), nil
	}
	raw := geom.NewPolygon(ls.CRS(), ring)
	return cleanOffsetPolygon(raw)
}

// cleanOffsetPolygon resolves the self-intersections produced by the
// offset-curve generator at concave corners by unioning the polygon
// with itself. Overlay-NG nodes the self-intersection points (now via
// the snap-rounding noder) and emits the simply-connected outer
// boundary. For a non-self-intersecting input the result is identical
// to the input (modulo coordinate canonicalisation).
//
// On overlay failure the raw polygon is returned, preserving the
// pre-cleanup behaviour as a safe fallback.
func cleanOffsetPolygon(raw *geom.Polygon) (*geom.Polygon, error) {
	cleaned, err := overlay.Union(raw, raw)
	if err != nil {
		return raw, nil
	}
	switch v := cleaned.(type) {
	case *geom.Polygon:
		if v.IsEmpty() {
			return raw, nil
		}
		return v, nil
	case *geom.MultiPolygon:
		// A self-intersecting offset that splits into multiple disjoint
		// polygons under union: return the largest by area as the
		// representative buffer body. Rare in practice — generally
		// concave corners produce a single outer boundary plus one or
		// more inner loops which Union absorbs as holes.
		var best *geom.Polygon
		bestArea := math.Inf(-1)
		for i := 0; i < v.NumGeometries(); i++ {
			p := v.PolygonAt(i)
			a := math.Abs(planar.Default.RingArea(p.Ring(0)))
			if a > bestArea {
				bestArea = a
				best = p
			}
		}
		if best != nil {
			return best, nil
		}
	}
	return raw, nil
}

// isClosedLine reports whether ls is a closed polyline (first vertex
// equals last) with at least 4 vertices.
func isClosedLine(ls *geom.LineString) bool {
	n := ls.NumPoints()
	if n < 4 {
		return false
	}
	return ls.PointAt(0) == ls.PointAt(n-1)
}

// bufferClosedLineAnnulus returns the buffer of a closed LineString
// at positive distance: an annulus equal to (interior dilated by d) ∖
// (interior eroded by d). Returns ok=false if the wrapped polygon is
// degenerate.
func bufferClosedLineAnnulus(ls *geom.LineString, distance float64, cfg config) (geom.Geometry, bool) {
	if distance <= 0 {
		return nil, false
	}
	pts := dedupedPoints(ls)
	if len(pts) < 4 {
		return nil, false
	}
	// Wrap the closed line as a polygon; orient it CCW so bufferPolygon's
	// dilation/inset logic applies correctly.
	poly := geom.NewPolygon(ls.CRS(), pts)
	if planar.Default.RingArea(poly.Ring(0)) < 0 {
		// Reverse to CCW.
		reversed := make([]geom.XY, len(pts))
		for i, p := range pts {
			reversed[len(pts)-1-i] = p
		}
		poly = geom.NewPolygon(ls.CRS(), reversed)
	}
	dilated, err := bufferPolygon(poly, distance, cfg)
	if err != nil {
		return nil, false
	}
	eroded, err := bufferPolygon(poly, -distance, cfg)
	if err != nil {
		return nil, false
	}
	if eroded == nil || eroded.IsEmpty() {
		return dilated, true
	}
	annulus, err := overlay.Difference(dilated, eroded)
	if err != nil {
		return dilated, true
	}
	return annulus, true
}

// dedupedPoints extracts XY vertices from ls, dropping consecutive
// duplicates.
func dedupedPoints(ls *geom.LineString) []geom.XY {
	n := ls.NumPoints()
	out := make([]geom.XY, 0, n)
	for i := 0; i < n; i++ {
		p := ls.PointAt(i)
		if len(out) > 0 && out[len(out)-1].Equal(p) {
			continue
		}
		out = append(out, p)
	}
	return out
}

