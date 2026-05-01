package validate

import (
	"fmt"
	"math"
	"strings"

	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/kernel"
	"github.com/terra-geo/terra/kernel/planar"
)

// DefectKind classifies a structural defect.
type DefectKind string

const (
	DefectRingNotClosed     DefectKind = "ring-not-closed"
	DefectRingTooFewPoints  DefectKind = "ring-too-few-points"
	DefectLineTooFewPoints  DefectKind = "line-too-few-points"
	DefectSelfIntersection  DefectKind = "self-intersection"
	DefectHoleOutsideShell  DefectKind = "hole-outside-shell"
	DefectInvalidLayout        DefectKind = "invalid-layout"
	DefectInvalidCoordinate    DefectKind = "invalid-coordinate"
	DefectDisconnectedInterior DefectKind = "disconnected-interior"
)

// Defect describes one specific failure.
type Defect struct {
	Kind     DefectKind
	Message  string
	Location geom.XY // approximate location, zero if not applicable
}

// ValidationError aggregates all defects found in a single Validate call.
type ValidationError struct {
	Defects []Defect
}

func (e *ValidationError) Error() string {
	var b strings.Builder
	b.WriteString("terra: invalid geometry: ")
	for i, d := range e.Defects {
		if i > 0 {
			b.WriteString("; ")
		}
		fmt.Fprintf(&b, "%s: %s", d.Kind, d.Message)
	}
	return b.String()
}

// Validate returns nil if g is a valid OGC geometry, or *ValidationError
// listing every defect detected.
func Validate(g geom.Geometry) error {
	v := &validator{}
	v.check(g)
	if len(v.defects) == 0 {
		return nil
	}
	return &ValidationError{Defects: v.defects}
}

type validator struct {
	defects []Defect
}

func (v *validator) add(kind DefectKind, msg string, loc geom.XY) {
	v.defects = append(v.defects, Defect{Kind: kind, Message: msg, Location: loc})
}

func (v *validator) check(g geom.Geometry) {
	if g.Layout() == geom.NoLayout && !g.IsEmpty() {
		v.add(DefectInvalidLayout, "geometry has NoLayout but is not empty", geom.XY{})
	}
	v.checkCoordinates(g)
	switch x := g.(type) {
	case *geom.Point:
		// Empty or single coordinate; nothing further to validate.
	case *geom.LineString:
		v.checkLineString(x)
	case *geom.Polygon:
		v.checkPolygon(x)
	case *geom.MultiPoint:
		// Each member is a single coordinate; no structural rule beyond layout.
	case *geom.MultiLineString:
		for i := 0; i < x.NumGeometries(); i++ {
			v.checkLineString(x.LineStringAt(i))
		}
	case *geom.MultiPolygon:
		for i := 0; i < x.NumGeometries(); i++ {
			v.checkPolygon(x.PolygonAt(i))
		}
		v.checkMultiPolygon(x)
	case *geom.GeometryCollection:
		for i := 0; i < x.NumGeometries(); i++ {
			v.check(x.GeometryAt(i))
		}
	}
}

// checkCoordinates flags any non-finite ordinate (NaN or ±Inf) found
// anywhere in g. JTS treats such inputs as invalid geometries.
func (v *validator) checkCoordinates(g geom.Geometry) {
	bad := func(p geom.XY) bool {
		return math.IsNaN(p.X) || math.IsNaN(p.Y) ||
			math.IsInf(p.X, 0) || math.IsInf(p.Y, 0)
	}
	report := func(p geom.XY) {
		v.add(DefectInvalidCoordinate,
			fmt.Sprintf("non-finite coordinate: %v", p), p)
	}
	switch x := g.(type) {
	case *geom.Point:
		if !x.IsEmpty() && bad(x.XY()) {
			report(x.XY())
		}
	case *geom.LineString:
		for i := 0; i < x.NumPoints(); i++ {
			if p := x.PointAt(i); bad(p) {
				report(p)
				return
			}
		}
	case *geom.Polygon:
		for r := 0; r < x.NumRings(); r++ {
			for _, p := range x.Ring(r) {
				if bad(p) {
					report(p)
					return
				}
			}
		}
	case *geom.MultiPoint:
		for i := 0; i < x.NumGeometries(); i++ {
			if p := x.PointAt(i); bad(p) {
				report(p)
				return
			}
		}
	case *geom.MultiLineString:
		for i := 0; i < x.NumGeometries(); i++ {
			v.checkCoordinates(x.LineStringAt(i))
		}
	case *geom.MultiPolygon:
		for i := 0; i < x.NumGeometries(); i++ {
			v.checkCoordinates(x.PolygonAt(i))
		}
	case *geom.GeometryCollection:
		for i := 0; i < x.NumGeometries(); i++ {
			v.checkCoordinates(x.GeometryAt(i))
		}
	}
}

func (v *validator) checkLineString(ls *geom.LineString) {
	if ls.IsEmpty() {
		return
	}
	if ls.NumPoints() < 2 {
		v.add(DefectLineTooFewPoints,
			fmt.Sprintf("line has %d points, need ≥2", ls.NumPoints()),
			ls.PointAt(0))
		return
	}
	// JTS validity requires ≥ 2 *distinct* vertices — a line with all
	// coincident points (LINESTRING(10 10, 10 10)) is invalid.
	first := ls.PointAt(0)
	distinct := false
	for i := 1; i < ls.NumPoints(); i++ {
		if ls.PointAt(i) != first {
			distinct = true
			break
		}
	}
	if !distinct {
		v.add(DefectLineTooFewPoints,
			"line has fewer than 2 distinct vertices",
			first)
	}
	// Note: a self-crossing (open OR closed) LineString is VALID per
	// OGC SFA — only its boundary endpoints are part of validity. WKT
	// LINEARRING is currently decoded as LineString, so this loses one JTS
	// distinction until geom grows a distinct LinearRing representation.
}

func (v *validator) checkPolygon(p *geom.Polygon) {
	if p.IsEmpty() {
		return
	}
	for r := 0; r < p.NumRings(); r++ {
		if !v.checkRing(p.Ring(r), r) {
			continue
		}
	}
	if p.NumRings() > 1 {
		k := planar.Default
		v.checkPolygonHoles(p, k)
		v.checkInteriorConnectivity(p, k)
	}
}

// checkInteriorConnectivity reports a disconnected-interior defect
// when the polygon's rings (shell + holes) form a graph cycle whose
// touch points are distinct enough to enclose a real sub-region of
// the interior.
//
// The touch graph has rings as nodes and an undirected edge for each
// pair of rings that share a point. A connected component contains a
// cycle iff its edge count exceeds (node count - 1). A cycle
// genuinely disconnects the interior iff the cycle traverses ≥2
// distinct touch points; multiple holes meeting at a single common
// point form a "spider" that's still topologically simply connected.
func (v *validator) checkInteriorConnectivity(p *geom.Polygon, k kernel.Kernel) {
	n := p.NumRings()
	if n < 3 {
		return
	}
	rings := make([][]geom.XY, n)
	for i := 0; i < n; i++ {
		rings[i] = p.Ring(i)
	}
	type edge struct {
		i, j  int
		point geom.XY
	}
	var edges []edge
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			if pt, ok := ringsTouchPoint(rings[i], rings[j], k); ok {
				edges = append(edges, edge{i, j, pt})
			}
		}
	}
	if len(edges) == 0 {
		return
	}
	// Group edges into connected components, then test each component
	// for cycles + distinct touch points.
	parent := make([]int, n)
	for i := range parent {
		parent[i] = i
	}
	find := func(x int) int {
		for parent[x] != x {
			parent[x] = parent[parent[x]]
			x = parent[x]
		}
		return x
	}
	for _, e := range edges {
		ri, rj := find(e.i), find(e.j)
		if ri != rj {
			parent[ri] = rj
		}
	}
	// For each component: count nodes, edges, and distinct points.
	comps := map[int]*componentStats{}
	for i := 0; i < n; i++ {
		root := find(i)
		c := comps[root]
		if c == nil {
			c = &componentStats{points: map[geom.XY]struct{}{}}
			comps[root] = c
		}
		c.nodes++
	}
	for _, e := range edges {
		root := find(e.i)
		c := comps[root]
		c.edges++
		c.points[e.point] = struct{}{}
	}
	for root, c := range comps {
		if c.edges > c.nodes-1 && len(c.points) >= 2 {
			v.add(DefectDisconnectedInterior,
				"rings form a cycle that disconnects the interior",
				rings[root][0])
			return
		}
	}
}

type componentStats struct {
	nodes  int
	edges  int
	points map[geom.XY]struct{}
}

// ringsTouchPoint reports whether ring A and ring B share at least
// one point and returns one such point (vertex-on-segment or
// vertex-on-vertex). Multiple touch points for the same pair are
// already detected by `ringTouchPointCount > 1` upstream as a
// self-intersection.
func ringsTouchPoint(a, b []geom.XY, k kernel.Kernel) (geom.XY, bool) {
	for i := 0; i+1 < len(a); i++ {
		if pointOnRingSegments(a[i], b, k) {
			return a[i], true
		}
	}
	for i := 0; i+1 < len(b); i++ {
		if pointOnRingSegments(b[i], a, k) {
			return b[i], true
		}
	}
	return geom.XY{}, false
}

func pointOnRingSegments(p geom.XY, ring []geom.XY, k kernel.Kernel) bool {
	for i := 0; i+1 < len(ring); i++ {
		if k.SegmentDistance(p, ring[i], ring[i+1]) <= 1e-12 {
			return true
		}
	}
	return false
}

func (v *validator) checkRing(ring []geom.XY, index int) bool {
	if len(ring) < 4 {
		loc := geom.XY{}
		if len(ring) > 0 {
			loc = ring[0]
		}
		v.add(DefectRingTooFewPoints,
			fmt.Sprintf("ring %d has %d vertices, need ≥4", index, len(ring)),
			loc)
		return false
	}
	distinctVerts := map[geom.XY]struct{}{}
	for i := 0; i+1 < len(ring); i++ {
		distinctVerts[ring[i]] = struct{}{}
	}
	if len(distinctVerts) < 3 {
		v.add(DefectRingTooFewPoints,
			fmt.Sprintf("ring %d has only %d distinct vertices, need ≥3",
				index, len(distinctVerts)),
			ring[0])
		return false
	}
	if ringSignedArea(ring) == 0 {
		v.add(DefectSelfIntersection,
			fmt.Sprintf("ring %d has zero area", index), ring[0])
		return false
	}
	if ring[0] != ring[len(ring)-1] {
		v.add(DefectRingNotClosed,
			fmt.Sprintf("ring %d not closed: first=%v last=%v", index, ring[0], ring[len(ring)-1]),
			ring[0])
	}
	if loc, ok := ringSelfIntersection(ring); ok {
		v.add(DefectSelfIntersection,
			fmt.Sprintf("ring %d self-intersects", index), loc)
	}
	return true
}

func (v *validator) checkPolygonHoles(p *geom.Polygon, k kernel.Kernel) {
	outer := p.Ring(0)
	for r := 1; r < p.NumRings(); r++ {
		v.checkHoleAgainstShell(r-1, p.Ring(r), outer, k)
	}
	for i := 1; i < p.NumRings(); i++ {
		for j := i + 1; j < p.NumRings(); j++ {
			v.checkHolePair(i-1, j-1, p.Ring(i), p.Ring(j), k)
		}
	}
}

func (v *validator) checkHoleAgainstShell(index int, hole, shell []geom.XY, k kernel.Kernel) {
	outsideHits := 0
	for _, vert := range hole {
		if k.PointInRing(vert, shell) == kernel.Outside {
			outsideHits++
		}
	}
	if outsideHits > 0 {
		v.add(DefectHoleOutsideShell,
			fmt.Sprintf("hole %d has %d vertices outside shell", index, outsideHits),
			hole[0])
		return
	}
	if ringEquivalent(hole, shell) {
		v.add(DefectSelfIntersection,
			fmt.Sprintf("hole %d equals shell", index), hole[0])
	}
	if ringsShareCurve(hole, shell) {
		v.add(DefectSelfIntersection,
			fmt.Sprintf("hole %d shares shell boundary", index), hole[0])
	} else if ringTouchPointCount(hole, shell) > 1 {
		v.add(DefectSelfIntersection,
			fmt.Sprintf("hole %d touches shell multiple times", index), hole[0])
	}
}

func (v *validator) checkHolePair(i, j int, a, b []geom.XY, k kernel.Kernel) {
	if ringEquivalent(a, b) {
		v.add(DefectSelfIntersection,
			fmt.Sprintf("holes %d and %d are equal", i, j),
			a[0])
		return
	}
	if ringContainsRingVerts(a, b, k) || ringContainsRingVerts(b, a, k) {
		v.add(DefectSelfIntersection,
			fmt.Sprintf("hole %d nested inside hole %d", i, j),
			a[0])
		return
	}
	if ringsShareCurve(a, b) {
		v.add(DefectSelfIntersection,
			fmt.Sprintf("holes %d and %d share a curve", i, j),
			a[0])
		return
	}
	if ringTouchPointCount(a, b) > 1 {
		v.add(DefectSelfIntersection,
			fmt.Sprintf("holes %d and %d touch multiple times", i, j),
			a[0])
	}
}

func (v *validator) checkMultiPolygon(mp *geom.MultiPolygon) {
	k := planar.Default
	for i := 0; i < mp.NumGeometries(); i++ {
		a := mp.PolygonAt(i)
		if a.IsEmpty() || a.NumRings() == 0 {
			continue
		}
		for j := i + 1; j < mp.NumGeometries(); j++ {
			b := mp.PolygonAt(j)
			if b.IsEmpty() || b.NumRings() == 0 {
				continue
			}
			ar, br := a.Ring(0), b.Ring(0)
			if ringEquivalent(ar, br) {
				v.add(DefectSelfIntersection,
					fmt.Sprintf("multipolygon shells %d and %d are equal", i, j), ar[0])
				continue
			}
			if ringsShareCurve(ar, br) {
				v.add(DefectSelfIntersection,
					fmt.Sprintf("multipolygon shells %d and %d share boundary", i, j), ar[0])
				continue
			}
			if polygonHasPointInInterior(a, b, k) || polygonHasPointInInterior(b, a, k) {
				v.add(DefectSelfIntersection,
					fmt.Sprintf("multipolygon shells %d and %d overlap or nest", i, j), ar[0])
				continue
			}
		}
	}
}

func polygonHasPointInInterior(a, b *geom.Polygon, k kernel.Kernel) bool {
	for r := 0; r < a.NumRings(); r++ {
		ring := a.Ring(r)
		for i := 0; i+1 < len(ring); i++ {
			if polygonPointLocation(ring[i], b, k) == kernel.Inside {
				return true
			}
		}
	}
	return false
}

func polygonPointLocation(pt geom.XY, p *geom.Polygon, k kernel.Kernel) kernel.Containment {
	if p.IsEmpty() || p.NumRings() == 0 {
		return kernel.Outside
	}
	loc := k.PointInRing(pt, p.Ring(0))
	if loc != kernel.Inside {
		return loc
	}
	for r := 1; r < p.NumRings(); r++ {
		holeLoc := k.PointInRing(pt, p.Ring(r))
		if holeLoc == kernel.Inside {
			return kernel.Outside
		}
		if holeLoc == kernel.OnBoundary {
			return kernel.OnBoundary
		}
	}
	return kernel.Inside
}

// ringEquivalent reports whether two closed rings describe the same
// vertex sequence (allowing a different starting offset and either
// orientation).
func ringEquivalent(a, b []geom.XY) bool {
	if len(a) != len(b) || len(a) < 4 {
		return false
	}
	// Compare without the closing duplicate.
	na := len(a) - 1
	if na != len(b)-1 {
		return false
	}
	// Same orientation: try every rotation.
	for off := 0; off < na; off++ {
		match := true
		for i := 0; i < na; i++ {
			if a[i] != b[(i+off)%na] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	// Reversed orientation.
	for off := 0; off < na; off++ {
		match := true
		for i := 0; i < na; i++ {
			if a[i] != b[(off-i+na*na)%na] {
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

// ringContainsRingVerts reports whether every vertex of `inner` is
// inside `outer`'s closure (and at least one is strictly inside).
func ringContainsRingVerts(outer, inner []geom.XY, k kernel.Kernel) bool {
	insideCount := 0
	for i := 0; i+1 < len(inner); i++ { // skip closing dup
		c := k.PointInRing(inner[i], outer)
		if c == kernel.Outside {
			return false
		}
		if c == kernel.Inside {
			insideCount++
		}
	}
	return insideCount > 0
}

// ringsShareCurve reports whether two rings share a 1-D segment (any
// edge of A overlaps collinearly with any edge of B).
func ringsShareCurve(a, b []geom.XY) bool {
	for i := 0; i+1 < len(a); i++ {
		for j := 0; j+1 < len(b); j++ {
			if collinearShare(a[i], a[i+1], b[j], b[j+1]) {
				return true
			}
		}
	}
	return false
}

func ringTouchPointCount(a, b []geom.XY) int {
	points := map[geom.XY]struct{}{}
	k := planar.Default
	for i := 0; i+1 < len(a); i++ {
		for j := 0; j+1 < len(b); j++ {
			if collinearShare(a[i], a[i+1], b[j], b[j+1]) {
				continue
			}
			ip, ok := k.SegmentIntersection(a[i], a[i+1], b[j], b[j+1])
			if !ok {
				continue
			}
			points[ip] = struct{}{}
		}
	}
	return len(points)
}

func ringSignedArea(ring []geom.XY) float64 {
	if len(ring) < 4 {
		return 0
	}
	var sum float64
	for i := 0; i+1 < len(ring); i++ {
		sum += ring[i].X*ring[i+1].Y - ring[i+1].X*ring[i].Y
	}
	return sum / 2
}

func collinearShare(p1, p2, p3, p4 geom.XY) bool {
	cross := func(o, p, q geom.XY) float64 {
		return (p.X-o.X)*(q.Y-o.Y) - (p.Y-o.Y)*(q.X-o.X)
	}
	if cross(p1, p2, p3) != 0 || cross(p1, p2, p4) != 0 {
		return false
	}
	dx, dy := p2.X-p1.X, p2.Y-p1.Y
	useX := dx*dx >= dy*dy
	t := func(p geom.XY) float64 {
		if useX {
			if dx == 0 {
				return 0
			}
			return (p.X - p1.X) / dx
		}
		if dy == 0 {
			return 0
		}
		return (p.Y - p1.Y) / dy
	}
	tb1, tb2 := t(p3), t(p4)
	if tb1 > tb2 {
		tb1, tb2 = tb2, tb1
	}
	// Open-interval overlap on more than a single point.
	if tb2 <= 0 || tb1 >= 1 {
		return false
	}
	lo := tb1
	if lo < 0 {
		lo = 0
	}
	hi := tb2
	if hi > 1 {
		hi = 1
	}
	return hi-lo > 0
}

// ringSelfIntersection returns the first pair of non-adjacent ring edges
// that touch or cross. Per JTS validity, a ring is invalid if any two
// non-consecutive segments share ANY point (vertex or interior) — this
// catches bow-ties, self-touching shells, and spikes.
//
// Consecutive-duplicate vertices (zero-length edges) are collapsed before
// the check so polygons like POLYGON((0 0, 1 0, 1 1, 1 1, 0 0)) are
// treated as their simple form.
func ringSelfIntersection(ring []geom.XY) (geom.XY, bool) {
	collapsed := collapseConsecutiveDuplicates(ring)
	if len(collapsed) < 4 {
		return geom.XY{}, false
	}
	seen := map[geom.XY]int{}
	n := len(collapsed)
	for i := 0; i+1 < n; i++ {
		if prev, ok := seen[collapsed[i]]; ok {
			if i-prev > 1 {
				return collapsed[i], true
			}
		} else {
			seen[collapsed[i]] = i
		}
	}
	k := planar.Kernel{}
	for i := 0; i+1 < n; i++ {
		a1, a2 := collapsed[i], collapsed[i+1]
		for j := i + 2; j+1 < n; j++ {
			// Skip the wraparound consecutive pair: edge (n-2, n-1)
			// shares its endpoint ring[n-1]=ring[0] with edge (0, 1).
			if i == 0 && j+1 == n-1 {
				continue
			}
			b1, b2 := collapsed[j], collapsed[j+1]
			ix := k.SegmentIntersect(a1, a2, b1, b2)
			switch ix.Kind {
			case kernel.PointIntersection:
				return ix.P, true
			case kernel.CollinearOverlap:
				return ix.P, true
			default:
				continue
			}
		}
	}
	return geom.XY{}, false
}

func collapseConsecutiveDuplicates(ring []geom.XY) []geom.XY {
	if len(ring) == 0 {
		return ring
	}
	out := make([]geom.XY, 0, len(ring))
	out = append(out, ring[0])
	for i := 1; i < len(ring); i++ {
		if ring[i] != out[len(out)-1] {
			out = append(out, ring[i])
		}
	}
	// Re-close if necessary: if the original was closed but the
	// collapse removed the closing duplicate (which it does), append
	// the first point so callers can still rely on ring[n-1]==ring[0].
	if len(out) > 0 && out[0] != out[len(out)-1] && len(ring) > 0 && ring[0] == ring[len(ring)-1] {
		out = append(out, out[0])
	}
	return out
}
