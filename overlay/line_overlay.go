package overlay

import (
	"cmp"
	"slices"

	terra "github.com/terra-geo/terra"
	"github.com/terra-geo/terra/crs"
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/internal/noding"
	"github.com/terra-geo/terra/kernel"
	"github.com/terra-geo/terra/kernel/planar"
)

type overlayOp string

const (
	opIntersection overlayOp = "intersection"
	opUnion        overlayOp = "union"
	opDifference   overlayOp = "difference"
	opSymDiff      overlayOp = "symdifference"
)

// isLineal reports whether g is a LineString or MultiLineString.
func isLineal(g geom.Geometry) bool {
	switch g.(type) {
	case *geom.LineString, *geom.MultiLineString:
		return true
	}
	return false
}

// linealSegments converts a LineString or MultiLineString into a list of
// SegmentStrings tagged with `tag`.
func linealSegments(g geom.Geometry, tag int) []*noding.SegmentString {
	switch v := g.(type) {
	case *geom.LineString:
		if v.IsEmpty() || v.NumPoints() < 2 {
			return nil
		}
		coords := make([]geom.XY, v.NumPoints())
		for i := range coords {
			coords[i] = v.PointAt(i)
		}
		return []*noding.SegmentString{{Coords: coords, Tag: tag}}
	case *geom.MultiLineString:
		var out []*noding.SegmentString
		for i := 0; i < v.NumGeometries(); i++ {
			out = append(out, linealSegments(v.LineStringAt(i), tag)...)
		}
		return out
	}
	return nil
}

// canonicalEdge returns the unordered endpoint pair of (a, b) — used as a
// stable map key for deduplication after noding.
type canonicalEdge struct {
	p1, p2 geom.XY
}

func canon(a, b geom.XY) canonicalEdge {
	if a.X < b.X || (a.X == b.X && a.Y < b.Y) {
		return canonicalEdge{a, b}
	}
	return canonicalEdge{b, a}
}

// lineLineOverlay runs a noded segment-set overlay between two lineal
// geometries (LineString or MultiLineString). Returns the result of the
// requested op as a LineString, MultiLineString, or empty geometry of
// the appropriate dimension.
//
// Supported ops: opIntersection, opUnion, opDifference, opSymDiff.
//
// Algorithm (textbook noded line-overlay):
//  1. Tag each input's segments (1 for A, 2 for B).
//  2. Node the union via the SimpleNoder — each output edge is a
//     non-self-intersecting sub-segment that retains its origin tag.
//  3. Group output segments by canonical endpoint pair; the union of
//     tags identifies which inputs contain each edge.
//  4. Filter by op rule.
//  5. Stitch filtered edges into LineStrings via greedy chain walk.
func lineLineOverlay(a, b geom.Geometry, op overlayOp) (geom.Geometry, error) {
	if !crs.Equal(a.CRS(), b.CRS()) {
		return nil, terra.ErrCRSMismatch
	}
	c := a.CRS()

	aSegs := linealSegments(a, 1)
	bSegs := linealSegments(b, 2)

	if len(aSegs) == 0 && len(bSegs) == 0 {
		return emptyOfDim(c, 1), nil
	}

	noded := noding.SimpleNoder{}.Node(append(append([]*noding.SegmentString{}, aSegs...), bSegs...))

	// Group by canonical edge — accumulate the union of source tags and
	// also the line-intersection-point dimension for each canonical edge.
	type edgeInfo struct {
		tags     int // bitmask: 1 = A, 2 = B
		pos      int // first index where seen, used for stable ordering
		inserted bool
	}
	edges := make(map[canonicalEdge]*edgeInfo)
	pointCounts := make(map[geom.XY]int) // tag bitmask per shared vertex (for point-only intersections)
	pointInsert := make(map[geom.XY]int)
	pointOrder := 0
	idx := 0
	for _, ss := range noded {
		for j := 0; j+1 < len(ss.Coords); j++ {
			p1, p2 := ss.Coords[j], ss.Coords[j+1]
			if p1 == p2 {
				continue
			}
			ce := canon(p1, p2)
			info, ok := edges[ce]
			if !ok {
				info = &edgeInfo{pos: idx}
				edges[ce] = info
				idx++
			}
			info.tags |= ss.Tag
		}
		// Track endpoints for point-only intersection cases (when A
		// and B touch at a single vertex with no shared edge).
		for _, p := range ss.Coords {
			if _, ok := pointInsert[p]; !ok {
				pointInsert[p] = pointOrder
				pointOrder++
			}
			pointCounts[p] |= ss.Tag
		}
	}

	// Apply op filter.
	var keep []canonicalEdge
	for ce, info := range edges {
		want := false
		switch op {
		case opIntersection:
			want = info.tags == 3
		case opUnion:
			want = info.tags != 0
		case opDifference:
			want = info.tags == 1
		case opSymDiff:
			want = info.tags == 1 || info.tags == 2
		}
		if want {
			keep = append(keep, ce)
		}
	}
	// Sort kept edges by first-seen index for stable output order.
	slices.SortFunc(keep, func(a, b canonicalEdge) int {
		return cmp.Compare(edges[a].pos, edges[b].pos)
	})

	// Stitch into LineStrings.
	lines := stitchEdges(keep, c)

	// For intersection / symdifference, also surface point-only
	// intersections (vertices shared by A and B that aren't endpoints
	// of any kept edge).
	var extraPoints []geom.XY
	if op == opIntersection || op == opUnion {
		// Collect endpoints already covered by kept edges.
		covered := map[geom.XY]struct{}{}
		for _, ce := range keep {
			covered[ce.p1] = struct{}{}
			covered[ce.p2] = struct{}{}
		}
		for p, tags := range pointCounts {
			if op == opIntersection && tags != 3 {
				continue
			}
			if op == opUnion && tags == 0 {
				continue
			}
			if _, ok := covered[p]; ok {
				continue
			}
			extraPoints = append(extraPoints, p)
		}
		slices.SortFunc(extraPoints, func(a, b geom.XY) int {
			return cmp.Compare(pointInsert[a], pointInsert[b])
		})
	}

	return assembleLinealResult(c, lines, extraPoints, op), nil
}

// stitchEdges greedily concatenates a set of edges into LineStrings by
// walking through shared endpoints. Edges with degree-2 nodes form
// continuous chains; degree-1 nodes start/end chains; degree ≥3 nodes
// terminate the current chain.
func stitchEdges(edges []canonicalEdge, c *crs.CRS) []*geom.LineString {
	if len(edges) == 0 {
		return nil
	}
	// Build adjacency: vertex -> list of edge indices.
	adj := make(map[geom.XY][]int)
	for i, e := range edges {
		adj[e.p1] = append(adj[e.p1], i)
		adj[e.p2] = append(adj[e.p2], i)
	}
	used := make([]bool, len(edges))
	var lines []*geom.LineString
	// Greedy: start a chain from any unused edge, extend in both directions.
	for i := range edges {
		if used[i] {
			continue
		}
		chain := []geom.XY{edges[i].p1, edges[i].p2}
		used[i] = true
		// Extend forward (from chain[len-1]) until no degree-2 continuation.
		for {
			tail := chain[len(chain)-1]
			next := -1
			candidates := 0
			for _, ei := range adj[tail] {
				if used[ei] {
					continue
				}
				candidates++
				next = ei
			}
			if candidates != 1 {
				break
			}
			used[next] = true
			e := edges[next]
			if e.p1 == tail {
				chain = append(chain, e.p2)
			} else {
				chain = append(chain, e.p1)
			}
		}
		// Extend backward.
		for {
			head := chain[0]
			next := -1
			candidates := 0
			for _, ei := range adj[head] {
				if used[ei] {
					continue
				}
				candidates++
				next = ei
			}
			if candidates != 1 {
				break
			}
			used[next] = true
			e := edges[next]
			if e.p1 == head {
				chain = append([]geom.XY{e.p2}, chain...)
			} else {
				chain = append([]geom.XY{e.p1}, chain...)
			}
		}
		flat := make([]float64, 0, len(chain)*2)
		for _, p := range chain {
			flat = append(flat, p.X, p.Y)
		}
		lines = append(lines, geom.NewLineStringFlatNoClone(geom.LayoutXY, c, flat))
	}
	return lines
}

func assembleLinealResult(c *crs.CRS, lines []*geom.LineString, points []geom.XY, op overlayOp) geom.Geometry {
	if len(lines) == 0 && len(points) == 0 {
		return emptyOfDim(c, 1)
	}
	if len(points) == 0 {
		if len(lines) == 1 {
			return lines[0]
		}
		return geom.NewMultiLineString(c, lines...)
	}
	if len(lines) == 0 {
		switch len(points) {
		case 1:
			return geom.NewPoint(c, points[0])
		default:
			return geom.NewMultiPoint(c, points)
		}
	}
	// Mixed result: GeometryCollection.
	members := make([]geom.Geometry, 0, len(lines)+1)
	if len(points) == 1 {
		members = append(members, geom.NewPoint(c, points[0]))
	} else {
		members = append(members, geom.NewMultiPoint(c, points))
	}
	if len(lines) == 1 {
		members = append(members, lines[0])
	} else {
		members = append(members, geom.NewMultiLineString(c, lines...))
	}
	return geom.NewGeometryCollection(c, members...)
}

// linePolygonOverlay handles overlay between a lineal A and a polygonal
// B (or vice versa). Splits A's segments at boundary crossings, then
// classifies each sub-segment's midpoint by point-in-polygon to decide
// inclusion per op rule.
func linePolygonOverlay(a, b geom.Geometry, op overlayOp) (geom.Geometry, error) {
	if !crs.Equal(a.CRS(), b.CRS()) {
		return nil, terra.ErrCRSMismatch
	}
	// Identify which side is lineal vs polygonal. The op semantics
	// depend on direction: for `intersection`/`union`/`symdifference`
	// the result is symmetric, but for `difference` we must respect
	// (line\poly) vs (poly\line).
	var lineSide, polySide geom.Geometry
	swapped := false
	if isLineal(a) {
		lineSide, polySide = a, b
	} else {
		lineSide, polySide = b, a
		swapped = true
	}
	c := a.CRS()
	k := planar.Default

	// Collect polygon-boundary segments for noding.
	lineSegs := linealSegments(lineSide, 1)
	polySegs := polygonalSegments(polySide, 2)
	combined := append(append([]*noding.SegmentString{}, lineSegs...), polySegs...)
	noded := noding.SimpleNoder{}.Node(combined)

	// Keep only segments that originated from the line side, classified
	// by midpoint vs the polygon.
	insideLines := []canonicalEdge{}
	outsideLines := []canonicalEdge{}
	boundaryLines := []canonicalEdge{}
	seen := map[canonicalEdge]struct{}{}
	// Per-vertex tag mask across ALL noded segments (1=line, 2=poly).
	// Used to detect isolated touch points (vertex carries both 1 and
	// 2 tags but no kept line edge incidence).
	vertexTags := make(map[geom.XY]int)
	for _, ss := range noded {
		for _, v := range ss.Coords {
			vertexTags[v] |= ss.Tag
		}
	}
	for _, ss := range noded {
		if ss.Tag != 1 {
			continue
		}
		for j := 0; j+1 < len(ss.Coords); j++ {
			p1, p2 := ss.Coords[j], ss.Coords[j+1]
			if p1 == p2 {
				continue
			}
			ce := canon(p1, p2)
			if _, dup := seen[ce]; dup {
				continue
			}
			seen[ce] = struct{}{}
			mid := geom.XY{X: (p1.X + p2.X) / 2, Y: (p1.Y + p2.Y) / 2}
			cont := classifyAgainstPolygonal(mid, polySide, k)
			switch cont {
			case kernel.Inside:
				insideLines = append(insideLines, ce)
			case kernel.OnBoundary:
				boundaryLines = append(boundaryLines, ce)
			default:
				outsideLines = append(outsideLines, ce)
			}
		}
	}

	// Apply op rule.
	var kept []canonicalEdge
	switch op {
	case opIntersection:
		// line ∩ poly: line segments inside or on the polygon boundary.
		kept = append(kept, insideLines...)
		kept = append(kept, boundaryLines...)
	case opDifference:
		// (line \ poly): segments outside the polygon.
		// For (poly \ line), the polygon is unchanged (lines have
		// dim 1 < 2, can't subtract).
		if !swapped {
			kept = append(kept, outsideLines...)
		} else {
			return polySide, nil
		}
	case opSymDiff:
		// SymDiff(line, poly) = (line\poly) ∪ (poly\line).
		// Lower-dim line subtracts nothing from poly; result is
		// poly + (line outside poly).
		outsidePart := stitchEdges(outsideLines, c)
		if len(outsidePart) == 0 {
			return polySide, nil
		}
		var members []geom.Geometry
		members = append(members, polySide)
		if len(outsidePart) == 1 {
			members = append(members, outsidePart[0])
		} else {
			members = append(members, geom.NewMultiLineString(c, outsidePart...))
		}
		return geom.NewGeometryCollection(c, members...), nil
	case opUnion:
		outsidePart := stitchEdges(outsideLines, c)
		if len(outsidePart) == 0 {
			return polySide, nil
		}
		var members []geom.Geometry
		members = append(members, polySide)
		if len(outsidePart) == 1 {
			members = append(members, outsidePart[0])
		} else {
			members = append(members, geom.NewMultiLineString(c, outsidePart...))
		}
		return geom.NewGeometryCollection(c, members...), nil
	}

	lines := stitchEdges(kept, c)

	// For intersection, surface isolated touch points: vertices where
	// the line meets the polygon boundary at a single point (line
	// vertex coinciding with polygon vertex/edge but no incident line
	// edge classified as inside/boundary).
	var extraPoints []geom.XY
	if op == opIntersection {
		covered := map[geom.XY]struct{}{}
		for _, ce := range kept {
			covered[ce.p1] = struct{}{}
			covered[ce.p2] = struct{}{}
		}
		// Stable order: track first-seen vertex order from line-tagged
		// segments, then deduplicate.
		order := make(map[geom.XY]int)
		nextOrder := 0
		for _, ss := range noded {
			if ss.Tag != 1 {
				continue
			}
			for _, v := range ss.Coords {
				if _, ok := order[v]; !ok {
					order[v] = nextOrder
					nextOrder++
				}
			}
		}
		emitted := map[geom.XY]struct{}{}
		for v, mask := range vertexTags {
			if mask != 3 { // Need both line and polygon-boundary incidence.
				continue
			}
			if _, on := covered[v]; on {
				continue
			}
			// Confirm via point-in-polygon that the touch point is on
			// the boundary (it must be — polygon-boundary tag implies
			// it's a vertex on a polygon ring or a noded crossing).
			cont := classifyAgainstPolygonal(v, polySide, k)
			if cont != kernel.OnBoundary {
				continue
			}
			if _, dup := emitted[v]; dup {
				continue
			}
			emitted[v] = struct{}{}
			extraPoints = append(extraPoints, v)
		}
		slices.SortFunc(extraPoints, func(a, b geom.XY) int {
			return cmp.Compare(order[a], order[b])
		})
	}

	return assembleLinealResult(c, lines, extraPoints, op), nil
}

// polygonalSegments converts a Polygon or MultiPolygon to a list of
// SegmentStrings (one per ring) tagged with `tag`.
func polygonalSegments(g geom.Geometry, tag int) []*noding.SegmentString {
	switch v := g.(type) {
	case *geom.Polygon:
		var out []*noding.SegmentString
		for r := 0; r < v.NumRings(); r++ {
			ring := v.Ring(r)
			coords := append([]geom.XY(nil), ring...)
			out = append(out, &noding.SegmentString{Coords: coords, Tag: tag})
		}
		return out
	case *geom.MultiPolygon:
		var out []*noding.SegmentString
		for i := 0; i < v.NumGeometries(); i++ {
			out = append(out, polygonalSegments(v.PolygonAt(i), tag)...)
		}
		return out
	}
	return nil
}

// classifyAgainstPolygonal returns Inside/OnBoundary/Outside for p
// against a Polygon or MultiPolygon (the latter via member union: a
// point Inside any polygon → Inside; OnBoundary of any without being
// Inside another → OnBoundary).
func classifyAgainstPolygonal(p geom.XY, g geom.Geometry, k kernel.Kernel) kernel.Containment {
	switch v := g.(type) {
	case *geom.Polygon:
		return classifyAgainstPolygon(p, v, k)
	case *geom.MultiPolygon:
		best := kernel.Outside
		for i := 0; i < v.NumGeometries(); i++ {
			c := classifyAgainstPolygon(p, v.PolygonAt(i), k)
			if c == kernel.Inside {
				return kernel.Inside
			}
			if c == kernel.OnBoundary {
				best = kernel.OnBoundary
			}
		}
		return best
	}
	return kernel.Outside
}

func classifyAgainstPolygon(p geom.XY, poly *geom.Polygon, k kernel.Kernel) kernel.Containment {
	if poly.NumRings() == 0 {
		return kernel.Outside
	}
	c := k.PointInRing(p, poly.Ring(0))
	if c == kernel.Outside {
		return kernel.Outside
	}
	for r := 1; r < poly.NumRings(); r++ {
		hc := k.PointInRing(p, poly.Ring(r))
		if hc == kernel.Inside {
			return kernel.Outside
		}
		if hc == kernel.OnBoundary {
			return kernel.OnBoundary
		}
	}
	return c
}
