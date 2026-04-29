package simplify

import (
	"math"

	"github.com/terra-geo/terra/geom"
)

// TopologyPreserving returns a simplified copy of g that is guaranteed
// not to introduce self-intersections (the simplified geometry remains
// simple if the input was simple).
//
// Vertices are removed in increasing order of their Visvalingam-Whyatt
// effective triangle area. A vertex is removed only if (a) its area is
// below tolerance² AND (b) the replacement segment does not cross any
// other live segment in the chain or in a sibling ring of the same
// polygon. Vertices that fail the safety check are kept.
//
// A tolerance ≤ 0 returns g unchanged.
func TopologyPreserving(g geom.Geometry, tolerance float64) geom.Geometry {
	if tolerance <= 0 || g.IsEmpty() {
		return g
	}
	switch v := g.(type) {
	case *geom.Point:
		return v
	case *geom.LineString:
		return tpsLineString(v, tolerance)
	case *geom.Polygon:
		return tpsPolygon(v, tolerance)
	case *geom.MultiPoint:
		return v
	case *geom.MultiLineString:
		parts := make([]*geom.LineString, 0, v.NumGeometries())
		for i := 0; i < v.NumGeometries(); i++ {
			parts = append(parts, tpsLineString(v.LineStringAt(i), tolerance))
		}
		return geom.NewMultiLineString(v.CRS(), parts...)
	case *geom.MultiPolygon:
		parts := make([]*geom.Polygon, 0, v.NumGeometries())
		for i := 0; i < v.NumGeometries(); i++ {
			parts = append(parts, tpsPolygon(v.PolygonAt(i), tolerance))
		}
		return geom.NewMultiPolygon(v.CRS(), parts...)
	case *geom.GeometryCollection:
		parts := make([]geom.Geometry, 0, v.NumGeometries())
		for i := 0; i < v.NumGeometries(); i++ {
			parts = append(parts, TopologyPreserving(v.GeometryAt(i), tolerance))
		}
		return geom.NewGeometryCollection(v.CRS(), parts...)
	}
	return g
}

func tpsLineString(ls *geom.LineString, tol float64) *geom.LineString {
	pts := lineToXY(ls)
	out := visvalingamSimplify(pts, tol*tol, false /*closed*/, nil)
	return geom.NewLineString(ls.CRS(), out)
}

func tpsPolygon(p *geom.Polygon, tol float64) *geom.Polygon {
	threshold := tol * tol
	rings := make([][]geom.XY, 0, p.NumRings())
	for r := 0; r < p.NumRings(); r++ {
		var constraints [][]geom.XY
		for s := 0; s < p.NumRings(); s++ {
			if s == r {
				continue
			}
			constraints = append(constraints, p.Ring(s))
		}
		simplified := visvalingamSimplify(p.Ring(r), threshold, true, constraints)
		if len(simplified) >= 4 {
			rings = append(rings, simplified)
		} else if r == 0 {
			return p // refuse to over-simplify the outer ring
		}
	}
	return geom.NewPolygon(p.CRS(), rings...)
}

// visvalingamSimplify removes vertices in order of smallest effective
// area, skipping any whose removal would introduce a segment crossing
// in the chain or against `constraints`. The algorithm is greedy O(n²);
// adequate for v1.0 scope.
func visvalingamSimplify(pts []geom.XY, threshold float64, closed bool, constraints [][]geom.XY) []geom.XY {
	if len(pts) <= 2 {
		return append([]geom.XY(nil), pts...)
	}
	// Closed rings come in with first == last. Drop the closing
	// duplicate; we'll restore it at the end.
	closing := false
	if closed && pts[0] == pts[len(pts)-1] {
		pts = pts[:len(pts)-1]
		closing = true
	}

	n := len(pts)
	prev := make([]int, n)
	next := make([]int, n)
	alive := make([]bool, n)
	frozen := make([]bool, n)
	for i := 0; i < n; i++ {
		prev[i] = (i - 1 + n) % n
		next[i] = (i + 1) % n
		alive[i] = true
	}
	if !closed {
		prev[0] = -1
		next[n-1] = -1
	}
	count := n
	minLive := 3
	if !closed {
		minLive = 2
	}

	for count > minLive {
		minArea := math.Inf(1)
		minIdx := -1
		for i := 0; i < n; i++ {
			if !alive[i] || frozen[i] {
				continue
			}
			if !closed && (prev[i] < 0 || next[i] < 0) {
				continue // endpoints of open lines are pinned
			}
			a := pts[prev[i]]
			b := pts[i]
			c := pts[next[i]]
			area := triangleArea2(a, b, c)
			if area < minArea {
				minArea = area
				minIdx = i
			}
		}
		if minIdx < 0 || minArea > threshold {
			break
		}
		a := pts[prev[minIdx]]
		c := pts[next[minIdx]]
		if !safeReplace(a, c, prev[minIdx], next[minIdx], pts, prev, next, alive, closed, constraints) {
			frozen[minIdx] = true
			continue
		}
		alive[minIdx] = false
		count--
		next[prev[minIdx]] = next[minIdx]
		prev[next[minIdx]] = prev[minIdx]
	}

	out := make([]geom.XY, 0, count+1)
	if closed {
		start := -1
		for i := 0; i < n; i++ {
			if alive[i] {
				start = i
				break
			}
		}
		if start < 0 {
			return nil
		}
		i := start
		for {
			out = append(out, pts[i])
			i = next[i]
			if i == start {
				break
			}
		}
		if closing {
			out = append(out, out[0])
		}
	} else {
		i := 0
		for i >= 0 {
			if alive[i] {
				out = append(out, pts[i])
			}
			i = next[i]
		}
	}
	return out
}

// triangleArea2 returns 2× the absolute area of the triangle (a, b, c).
// Comparing 2A is sufficient since we only need ordering and a
// threshold; multiplying once at the call site avoids per-vertex
// halving.
func triangleArea2(a, b, c geom.XY) float64 {
	return math.Abs((b.X-a.X)*(c.Y-a.Y) - (b.Y-a.Y)*(c.X-a.X))
}

// safeReplace reports whether replacing the segments
// (a → pts[mid] → c) with the single segment (a → c) introduces any
// crossing with another live segment in the chain or any segment in
// the constraints rings.
//
// pIdx and nIdx are the prev/next indices of the vertex about to be
// removed; a == pts[pIdx], c == pts[nIdx]. We must skip those two
// adjacency segments when scanning the chain (they're the segments
// being replaced and the ones immediately preceding/following).
func safeReplace(a, c geom.XY, pIdx, nIdx int, pts []geom.XY, prev, next []int, alive []bool, closed bool, constraints [][]geom.XY) bool {
	// Test against every live chain segment except those incident to
	// pIdx or nIdx.
	n := len(pts)
	for i := 0; i < n; i++ {
		if !alive[i] {
			continue
		}
		j := next[i]
		if j < 0 || !alive[j] {
			continue
		}
		// Segment (pts[i] → pts[j]). Skip if it's the segment incident
		// to pIdx or nIdx.
		if i == pIdx || j == pIdx || i == nIdx || j == nIdx {
			continue
		}
		if segmentsProperlyCross(a, c, pts[i], pts[j]) {
			return false
		}
	}
	// Test against constraints.
	for _, ring := range constraints {
		for k := 0; k+1 < len(ring); k++ {
			if segmentsProperlyCross(a, c, ring[k], ring[k+1]) {
				return false
			}
		}
	}
	return true
}

// segmentsProperlyCross reports whether segments (a,b) and (c,d) cross
// in their interiors. Endpoint-touch is allowed (returns false).
func segmentsProperlyCross(a, b, c, d geom.XY) bool {
	// Standard orientation test.
	o1 := orient(a, b, c)
	o2 := orient(a, b, d)
	o3 := orient(c, d, a)
	o4 := orient(c, d, b)
	if o1 != o2 && o3 != o4 {
		// Endpoint-touch case: if any endpoint coincides, treat as
		// non-crossing (boundary contact is OK).
		if a == c || a == d || b == c || b == d {
			return false
		}
		// Strict inequalities: only "proper" sign disagreements count.
		if o1 == 0 || o2 == 0 || o3 == 0 || o4 == 0 {
			return false
		}
		return true
	}
	return false
}

// orient returns the sign of the cross product (b-a) × (c-a):
// +1 = CCW, -1 = CW, 0 = collinear.
func orient(a, b, c geom.XY) int {
	v := (b.X-a.X)*(c.Y-a.Y) - (b.Y-a.Y)*(c.X-a.X)
	switch {
	case v > 0:
		return 1
	case v < 0:
		return -1
	}
	return 0
}
