package buffer

import (
	"errors"
	"fmt"
	"math"

	"github.com/terra-geo/terra"
	"github.com/terra-geo/terra/crs"
	"github.com/terra-geo/terra/geom"
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
			parts = append(parts, poly)
		}
		return geom.NewMultiPolygon(v.CRS(), parts...), nil

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

// segment is one edge of a polyline plus its precomputed unit
// left-perpendicular normal (rotation of (dx,dy)/L by +90° CCW).
type segment struct {
	a, b   geom.XY
	nx, ny float64 // unit left normal: (-dy, dx)/L
}

// forward returns the unit forward direction (a → b) of s.
// forward = rotate(normal, -90°) = (ny, -nx).
func (s segment) forward() (float64, float64) { return s.ny, -s.nx }

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

	segs := make([]segment, 0, len(pts)-1)
	for i := 0; i+1 < len(pts); i++ {
		a, b := pts[i], pts[i+1]
		dx, dy := b.X-a.X, b.Y-a.Y
		L := math.Hypot(dx, dy)
		if L == 0 {
			continue
		}
		segs = append(segs, segment{a: a, b: b, nx: -dy / L, ny: dx / L})
	}
	if len(segs) == 0 {
		return bufferPoint(ls.CRS(), pts[0], distance, cfg), nil
	}

	d := distance

	// --- Forward (left) offset chain ---
	left := make([]geom.XY, 0, 2*len(segs)+8)
	s0 := segs[0]
	left = append(left, geom.XY{X: s0.a.X + d*s0.nx, Y: s0.a.Y + d*s0.ny})
	for i := 0; i+1 < len(segs); i++ {
		curr := segs[i]
		next := segs[i+1]
		pCurrEnd := geom.XY{X: curr.b.X + d*curr.nx, Y: curr.b.Y + d*curr.ny}
		pNextStart := geom.XY{X: next.a.X + d*next.nx, Y: next.a.Y + d*next.ny}
		// Cross of curr.dir and next.dir; positive ⇒ left turn ⇒ convex
		// on the LEFT side, requires a join arc.
		// curr.dir = (curr.ny, -curr.nx); next.dir = (next.ny, -next.nx).
		cross := curr.ny*(-next.nx) - (-curr.nx)*next.ny
		if cross > 0 {
			arc := buildJoinArc(curr.b, pCurrEnd, pNextStart, curr, next, d, cfg)
			left = append(left, pCurrEnd)
			left = append(left, arc...)
		} else {
			left = append(left, pCurrEnd, pNextStart)
		}
	}
	sN := segs[len(segs)-1]
	left = append(left, geom.XY{X: sN.b.X + d*sN.nx, Y: sN.b.Y + d*sN.ny})

	// --- Backward (right) offset chain ---
	// Walking segs from N-1 down to 0 in reversed direction, the right
	// offset of each segment forms the second half of the ring. A "convex"
	// corner during reverse traversal of the right side corresponds to an
	// original RIGHT turn (cross < 0).
	right := make([]geom.XY, 0, 2*len(segs)+8)
	last := segs[len(segs)-1]
	right = append(right, geom.XY{X: last.b.X - d*last.nx, Y: last.b.Y - d*last.ny})
	for i := len(segs) - 1; i > 0; i-- {
		curr := segs[i]
		prev := segs[i-1]
		pCurrStart := geom.XY{X: curr.a.X - d*curr.nx, Y: curr.a.Y - d*curr.ny}
		pPrevEnd := geom.XY{X: prev.b.X - d*prev.nx, Y: prev.b.Y - d*prev.ny}
		// Original cross at vertex i (between prev.dir and curr.dir):
		//   c = prev.ny*(-curr.nx) - (-prev.nx)*curr.ny
		//     = -prev.ny*curr.nx + prev.nx*curr.ny
		c := -prev.ny*curr.nx + prev.nx*curr.ny
		if c < 0 {
			// Original right turn ⇒ convex on the right side during reverse
			// traversal. Build the join with reversed segments so the
			// "forward" direction of the join matches the traversal.
			revCurr := reverseSegment(curr)
			revPrev := reverseSegment(prev)
			arc := buildJoinArc(curr.a, pCurrStart, pPrevEnd, revCurr, revPrev, d, cfg)
			right = append(right, pCurrStart)
			right = append(right, arc...)
		} else {
			right = append(right, pCurrStart, pPrevEnd)
		}
	}
	first := segs[0]
	right = append(right, geom.XY{X: first.a.X - d*first.nx, Y: first.a.Y - d*first.ny})

	// --- End and start caps ---
	endCap := buildEndCap(sN.b, sN, d, cfg)
	startCap := buildStartCap(first.a, first, d, cfg)

	// Assemble closed ring.
	ring := make([]geom.XY, 0, len(left)+len(endCap)+len(right)+len(startCap)+1)
	ring = append(ring, left...)
	ring = append(ring, endCap...)
	ring = append(ring, right...)
	ring = append(ring, startCap...)
	ring = append(ring, ring[0])

	return geom.NewPolygon(ls.CRS(), ring), nil
}

// reverseSegment returns the segment with a/b swapped. The left normal of
// the reversed segment is the negation of the original left normal (the
// "left" of a→b is the "right" of b→a).
func reverseSegment(s segment) segment {
	return segment{a: s.b, b: s.a, nx: -s.nx, ny: -s.ny}
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

// buildJoinArc builds the interior vertices of a left-side convex join.
// pCurrEnd and pNextStart are the offset endpoints of the two adjacent
// segments at vertex; the caller emits pCurrEnd before the returned slice
// and the slice ends with pNextStart.
func buildJoinArc(vertex, pCurrEnd, pNextStart geom.XY, curr, next segment, d float64, cfg config) []geom.XY {
	switch cfg.join {
	case JoinBevel:
		return []geom.XY{pNextStart}
	case JoinMitre:
		mp, ok := mitrePoint(vertex, pCurrEnd, pNextStart, curr, next, d, cfg.mitreLimit)
		if !ok {
			return []geom.XY{pNextStart}
		}
		return []geom.XY{mp, pNextStart}
	case JoinRound:
		return roundArc(vertex, pCurrEnd, pNextStart, d, cfg.quadSegments)
	}
	return []geom.XY{pNextStart}
}

// roundArc returns the arc points (excluding p0, including p1) sweeping
// the short way around vertex at radius d.
func roundArc(vertex, p0, p1 geom.XY, d float64, quad int) []geom.XY {
	a0 := math.Atan2(p0.Y-vertex.Y, p0.X-vertex.X)
	a1 := math.Atan2(p1.Y-vertex.Y, p1.X-vertex.X)
	delta := a1 - a0
	for delta > math.Pi {
		delta -= 2 * math.Pi
	}
	for delta <= -math.Pi {
		delta += 2 * math.Pi
	}
	steps := int(math.Ceil(math.Abs(delta) / (math.Pi / 2) * float64(quad)))
	if steps < 1 {
		steps = 1
	}
	out := make([]geom.XY, 0, steps)
	for i := 1; i <= steps; i++ {
		t := float64(i) / float64(steps)
		theta := a0 + delta*t
		out = append(out, geom.XY{
			X: vertex.X + d*math.Cos(theta),
			Y: vertex.Y + d*math.Sin(theta),
		})
	}
	if len(out) > 0 {
		out[len(out)-1] = p1
	}
	return out
}

// mitrePoint computes the intersection of the two offset half-lines.
// Returns ok=false when the resulting mitre extension exceeds limit*d (the
// caller should fall back to a bevel).
func mitrePoint(vertex, pA, pB geom.XY, ea, eb segment, d, limit float64) (geom.XY, bool) {
	dax, day := ea.b.X-ea.a.X, ea.b.Y-ea.a.Y
	dbx, dby := eb.b.X-eb.a.X, eb.b.Y-eb.a.Y
	denom := dax*dby - day*dbx
	if denom == 0 {
		return geom.XY{}, false
	}
	t := ((pB.X-pA.X)*dby - (pB.Y-pA.Y)*dbx) / denom
	mp := geom.XY{X: pA.X + t*dax, Y: pA.Y + t*day}
	if math.Hypot(mp.X-vertex.X, mp.Y-vertex.Y) > limit*d {
		return geom.XY{}, false
	}
	return mp, true
}

// buildEndCap returns the cap vertices going from the last segment's left
// offset endpoint to its right offset endpoint at vertex. The endpoints
// themselves are NOT included (caller emits them).
func buildEndCap(vertex geom.XY, last segment, d float64, cfg config) []geom.XY {
	leftEnd := geom.XY{X: vertex.X + d*last.nx, Y: vertex.Y + d*last.ny}
	rightEnd := geom.XY{X: vertex.X - d*last.nx, Y: vertex.Y - d*last.ny}
	fx, fy := last.forward()
	switch cfg.cap {
	case CapFlat:
		return nil
	case CapSquare:
		_ = leftEnd
		_ = rightEnd
		// Two corner points extending forward by d.
		p1 := geom.XY{X: vertex.X + d*last.nx + d*fx, Y: vertex.Y + d*last.ny + d*fy}
		p2 := geom.XY{X: vertex.X - d*last.nx + d*fx, Y: vertex.Y - d*last.ny + d*fy}
		return []geom.XY{p1, p2}
	case CapRound:
		return semicircle(vertex, leftEnd, rightEnd, fx, fy, d, cfg.quadSegments)
	}
	return nil
}

// buildStartCap mirrors buildEndCap for the start of the polyline. The
// vertices go from the first segment's right offset start to its left
// offset start.
func buildStartCap(vertex geom.XY, first segment, d float64, cfg config) []geom.XY {
	leftStart := geom.XY{X: vertex.X + d*first.nx, Y: vertex.Y + d*first.ny}
	rightStart := geom.XY{X: vertex.X - d*first.nx, Y: vertex.Y - d*first.ny}
	fx, fy := first.forward()
	bx, by := -fx, -fy
	switch cfg.cap {
	case CapFlat:
		return nil
	case CapSquare:
		_ = leftStart
		_ = rightStart
		p1 := geom.XY{X: vertex.X - d*first.nx + d*bx, Y: vertex.Y - d*first.ny + d*by}
		p2 := geom.XY{X: vertex.X + d*first.nx + d*bx, Y: vertex.Y + d*first.ny + d*by}
		return []geom.XY{p1, p2}
	case CapRound:
		return semicircle(vertex, rightStart, leftStart, bx, by, d, cfg.quadSegments)
	}
	return nil
}

// semicircle returns the interior vertices of a semicircular arc on the
// circle of radius d around vertex, sweeping from p0 through the
// (fx,fy)-pointing half-plane to p1. p0 and p1 are NOT included; exactly
// 2*quad - 1 interior vertices are returned.
func semicircle(vertex, p0, p1 geom.XY, fx, fy, d float64, quad int) []geom.XY {
	a0 := math.Atan2(p0.Y-vertex.Y, p0.X-vertex.X)
	a1 := math.Atan2(p1.Y-vertex.Y, p1.X-vertex.X)
	delta := a1 - a0
	// Choose sweep direction so the arc midpoint lies in the (fx,fy)
	// half-plane (the "outside" of the line at this endpoint).
	mid := func() (float64, float64) {
		th := a0 + delta*0.5
		return math.Cos(th), math.Sin(th)
	}
	mx, my := mid()
	if mx*fx+my*fy < 0 {
		if delta > 0 {
			delta -= 2 * math.Pi
		} else {
			delta += 2 * math.Pi
		}
	}
	steps := 2 * quad
	out := make([]geom.XY, 0, steps-1)
	for i := 1; i < steps; i++ {
		t := float64(i) / float64(steps)
		theta := a0 + delta*t
		out = append(out, geom.XY{
			X: vertex.X + d*math.Cos(theta),
			Y: vertex.Y + d*math.Sin(theta),
		})
	}
	return out
}
