package predicate

import (
	terra "github.com/terra-geo/terra"
	"github.com/terra-geo/terra/crs"
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/kernel"
)

// DE9IM is the dimensionally-extended 9-intersection model relationship
// between two geometries. It is a 9-character string ordered:
//
//	II IB IE BI BB BE EI EB EE
//
// where I = interior, B = boundary, E = exterior, and each character is
// 'F' (intersection is empty) or '0'/'1'/'2' (dimension of the
// intersection: point / curve / area).
type DE9IM string

// Matrix indices for clarity.
const (
	mII = 0
	mIB = 1
	mIE = 2
	mBI = 3
	mBB = 4
	mBE = 5
	mEI = 6
	mEB = 7
	mEE = 8
)

// matrix is the working DE-9IM cell array with cells in [-1, 2]:
// -1 = empty intersection (F), 0/1/2 = dimension.
type matrix [9]int8

func newMatrix() matrix {
	return matrix{-1, -1, -1, -1, -1, -1, -1, -1, -1}
}

// raise raises cell i to at least dim.
func (m *matrix) raise(i int, dim int8) {
	if m[i] < dim {
		m[i] = dim
	}
}

// merge takes the cell-wise maximum of m and o. Used to combine per-member
// results for Multi-geometries. (See note in Relate about the limitations
// of this combiner for the *E columns.)
func (m *matrix) merge(o matrix) {
	for i, v := range o {
		if m[i] < v {
			m[i] = v
		}
	}
}

func (m matrix) toDE9IM() DE9IM {
	out := make([]byte, 9)
	for i, v := range m {
		if v < 0 {
			out[i] = 'F'
		} else {
			out[i] = '0' + byte(v)
		}
	}
	return DE9IM(out)
}

// Relate returns the DE-9IM matrix for (a, b).
//
// The matrix is computed from the topology of the input geometries:
// interior, boundary, and exterior intersections are derived using
// primitive operations (point-on-segment, segment-segment intersection,
// point-in-polygon, ring overlap), and for polygon-polygon pairs the
// overlay-NG DCEL provides edge-level boundary classification.
//
// Coverage:
//   - Single-type pairs (Point, LineString, Polygon and their cross
//     pairings) produce the exact matrix.
//   - Multi-geometries are handled by per-member combination. For
//     non-overlapping members (the typical case) the combined matrix is
//     exact; for Multi*s with overlapping members the *E columns may be
//     conservative (over-reporting interior-vs-exterior intersection).
//   - GeometryCollection inputs delegate to per-member combination.
func Relate(a, b geom.Geometry, opts ...Option) (DE9IM, error) {
	if !crs.Equal(a.CRS(), b.CRS()) {
		return "", terra.ErrCRSMismatch
	}
	a = unwrapLinearRing(a)
	b = unwrapLinearRing(b)
	cfg := resolve(a, opts)
	bnr := cfg.bnr
	if !cfg.bnrSet {
		bnr = Mod2BoundaryNodeRule
	}
	if cfg.useRelateNG {
		if im, ok := relateViaNG(a, b, bnr); ok {
			return im, nil
		}
		// Fall through to legacy path on inputs the new driver can't
		// definitively answer yet (edge-segment crossings).
	}
	m := computeMatrixWithRule(a, b, cfg.kernel, bnr)
	return m.toDE9IM(), nil
}

// computeMatrixWithRule wraps computeMatrix, applying the custom
// boundary-node rule by post-processing the boundary rows/columns.
//
// We call multiLineStringBoundaryRule with the chosen rule and use
// its emptiness to decide whether to suppress the *B/B* rows/cols.
// This is the narrowest correct integration: predicates that only
// look at "is boundary empty?" (boundaryDim) get the rule's answer.
func computeMatrixWithRule(a, b geom.Geometry, k kernel.Kernel, rule BoundaryNodeRule) matrix {
	m := computeMatrix(a, b, k)
	// If a non-default rule says the boundary IS empty, blank the
	// boundary rows. If it says boundary is NOT empty (when default
	// would say empty), promote the boundary rows to match the
	// dimension of the lineal geometry.
	applyBoundaryRule(&m, a, rule, true)
	applyBoundaryRule(&m, b, rule, false)
	return m
}

// applyBoundaryRule rewrites the matrix's *B / B* rows for a lineal
// operand using the supplied BoundaryNodeRule. forA selects which
// operand to act on (true => a's B-row, false => b's B-column).
func applyBoundaryRule(m *matrix, g geom.Geometry, rule BoundaryNodeRule, forA bool) {
	ml, ok := asMLSWrapped(g)
	if !ok {
		return
	}
	ruleEmpty := len(multiLineStringBoundaryRule(ml, rule)) == 0
	defaultEmpty := len(multiLineStringBoundary(ml)) == 0
	if ruleEmpty == defaultEmpty {
		return
	}
	if ruleEmpty {
		if forA {
			m[mBI] = -1
			m[mBB] = -1
			m[mBE] = -1
		} else {
			m[mIB] = -1
			m[mBB] = -1
			m[mEB] = -1
		}
		return
	}
	// Rule says boundary non-empty but default said empty: at least
	// the BE / EB cell should reflect dimension 0 (boundary is a
	// 0-dimensional point set).
	if forA {
		m.raise(mBE, 0)
	} else {
		m.raise(mEB, 0)
	}
}

// unwrapLinearRing routes a LinearRing through the LineString code paths.
// LinearRing exists primarily for OGC validity (rejecting self-intersecting
// closed rings); operationally relate/intersect/etc treat it as a 1-D curve.
func unwrapLinearRing(g geom.Geometry) geom.Geometry {
	if lr, ok := g.(*geom.LinearRing); ok {
		return lr.AsLineString()
	}
	return g
}

// computeMatrix dispatches on the geometry types and returns the raw
// matrix. Empty inputs are handled here; non-empty inputs are routed to
// per-pair builders.
func computeMatrix(a, b geom.Geometry, k kernel.Kernel) matrix {
	if a.IsEmpty() && b.IsEmpty() {
		// Both empty: only EE is non-empty (the whole plane).
		m := newMatrix()
		m.raise(mEE, 2)
		return m
	}
	if a.IsEmpty() {
		// int(a)=∂(a)=∅. ext(a) intersects b's interior/boundary at their
		// own dimensions.
		m := newMatrix()
		m.raise(mEE, 2)
		if d := dimensionOf(b); d >= 0 {
			m.raise(mEI, int8(d))
		}
		if d := boundaryDim(b); d >= 0 {
			m.raise(mEB, int8(d))
		}
		return m
	}
	if b.IsEmpty() {
		m := newMatrix()
		m.raise(mEE, 2)
		if d := dimensionOf(a); d >= 0 {
			m.raise(mIE, int8(d))
		}
		if d := boundaryDim(a); d >= 0 {
			m.raise(mBE, int8(d))
		}
		return m
	}

	// GC-aware dispatch: simplify GeometryCollection operands via
	// UnaryUnion (registered by the overlay package) before
	// classification. This collapses shared edges between adjacent
	// polygon members into the union's interior (per OGC).
	if m, ok := dispatchGCPair(a, b, k); ok {
		return m
	}

	// MLS-* dedicated multi-level paths that respect the mod-2 boundary
	// rule on intersection-point classification. The pairs not handled
	// here fall through to the per-member merge in relateMulti.
	if m, ok := dispatchMLSPair(a, b, k); ok {
		if boundaryDim(a) < 0 {
			m[mBI] = -1
			m[mBB] = -1
			m[mBE] = -1
		}
		if boundaryDim(b) < 0 {
			m[mIB] = -1
			m[mBB] = -1
			m[mEB] = -1
		}
		return m
	}

	// Multi/collection on either side: iterate members and merge.
	if isMulti(a) || isMulti(b) {
		m := relateMulti(a, b, k)
		// Aggregate-boundary post-processing: when A or B has empty
		// boundary at the multi-level (mod-2 for MLS, closed-line
		// detection), clear the row/column cells that the per-member
		// merge spuriously raised.
		if boundaryDim(a) < 0 {
			m[mBI] = -1
			m[mBB] = -1
			m[mBE] = -1
		}
		if boundaryDim(b) < 0 {
			m[mIB] = -1
			m[mBB] = -1
			m[mEB] = -1
		}
		return m
	}

	// Single-type dispatch.
	m := relatePair(a, b, k)
	// Closed-line boundary post-processing: if a is a closed LineString
	// (LinearRing) its boundary is empty; same for b.
	if boundaryDim(a) < 0 {
		m[mBI] = -1
		m[mBB] = -1
		m[mBE] = -1
	}
	if boundaryDim(b) < 0 {
		m[mIB] = -1
		m[mBB] = -1
		m[mEB] = -1
	}
	return m
}

// relatePair computes the matrix for a pair of single-typed (Point,
// LineString, Polygon) geometries.
func relatePair(a, b geom.Geometry, k kernel.Kernel) matrix {
	swap := false
	if typeRank(a) > typeRank(b) {
		a, b = b, a
		swap = true
	}
	var m matrix
	switch va := a.(type) {
	case *geom.Point:
		switch vb := b.(type) {
		case *geom.Point:
			m = relatePointPoint(va, vb)
		case *geom.LineString:
			m = relatePointLine(va, vb, k)
		case *geom.Polygon:
			m = relatePointPolygon(va, vb, k)
		default:
			m = newMatrix()
		}
	case *geom.LineString:
		switch vb := b.(type) {
		case *geom.LineString:
			m = relateLineLine(va, vb, k)
		case *geom.Polygon:
			m = relateLinePolygon(va, vb, k)
		default:
			m = newMatrix()
		}
	case *geom.Polygon:
		if vb, ok := b.(*geom.Polygon); ok {
			m = relatePolygonPolygon(va, vb, k)
		} else {
			m = newMatrix()
		}
	default:
		m = newMatrix()
	}
	if swap {
		m = transposeMatrix(m)
	}
	return m
}

// transposeMatrix swaps rows ↔ columns: II↔II, IB↔BI, IE↔EI, BB↔BB,
// BE↔EB, EE↔EE.
func transposeMatrix(m matrix) matrix {
	return matrix{
		m[mII], m[mBI], m[mEI],
		m[mIB], m[mBB], m[mEB],
		m[mIE], m[mBE], m[mEE],
	}
}

// boundaryDim returns the topological dimension of g's boundary, or -1 if
// the boundary is empty (Point, MultiPoint, closed LineString, or a
// MultiLineString whose mod-2 endpoint set is empty).
func boundaryDim(g geom.Geometry) int {
	switch v := g.(type) {
	case *geom.Point, *geom.MultiPoint:
		return -1
	case *geom.LineString:
		if v.NumPoints() < 2 {
			return -1
		}
		if v.PointAt(0) == v.PointAt(v.NumPoints()-1) {
			// Closed line — empty boundary.
			return -1
		}
		return 0
	case *geom.MultiLineString:
		// Mod-2 rule: an endpoint shared by an even number of members
		// is not a boundary point. If the aggregate set is empty,
		// boundary is empty.
		if len(multiLineStringBoundary(v)) == 0 {
			return -1
		}
		return 0
	case *geom.Polygon, *geom.MultiPolygon:
		if g.IsEmpty() {
			return -1
		}
		return 1
	case *geom.GeometryCollection:
		bd := -1
		for i := 0; i < v.NumGeometries(); i++ {
			d := boundaryDim(v.GeometryAt(i))
			if d > bd {
				bd = d
			}
		}
		return bd
	}
	return -1
}

// dispatchMLSPair routes MLS-on-one-side or MLS-on-both-sides cases to
// dedicated multi-level relate functions. Returns (matrix, true) when a
// dispatch was made, otherwise (zero, false).
//
// LineString operands are also accepted: they're wrapped as a one-member
// MultiLineString. Mod-2 boundary of {start,end} is exactly the natural
// open-line boundary, so the dispatch is semantics-preserving.
func dispatchMLSPair(a, b geom.Geometry, k kernel.Kernel) (matrix, bool) {
	amls, aIsLineal := asMLSWrapped(a)
	bmls, bIsLineal := asMLSWrapped(b)
	if aIsLineal && bIsLineal {
		return relateMLStoMLS(amls, bmls, k), true
	}
	if aIsLineal {
		switch v := b.(type) {
		case *geom.Polygon:
			return relateMLStoPolygon(amls, v, k), true
		case *geom.MultiPolygon:
			return relateMLStoMultiPolygon(amls, v, k), true
		}
	}
	if bIsLineal {
		switch v := a.(type) {
		case *geom.Polygon:
			return transposeMatrix(relateMLStoPolygon(bmls, v, k)), true
		case *geom.MultiPolygon:
			return transposeMatrix(relateMLStoMultiPolygon(bmls, v, k)), true
		}
	}
	return matrix{}, false
}

// asMLSWrapped returns g as a MultiLineString if g is lineal: any
// MultiLineString is returned as-is; a LineString is wrapped as a
// single-member MLS. Other types return (nil, false).
func asMLSWrapped(g geom.Geometry) (*geom.MultiLineString, bool) {
	switch v := g.(type) {
	case *geom.MultiLineString:
		return v, true
	case *geom.LineString:
		ml := geom.NewMultiLineString(v.CRS(), v)
		return ml, true
	}
	return nil, false
}

// multiLineStringBoundary returns the OGC mod-2 boundary point set of
// a MultiLineString: endpoints appearing in an odd number of member
// boundary endpoints. Closed members contribute nothing.
//
// Delegates to multiLineStringBoundaryRule with Mod2BoundaryRule for
// the OGC default. Callers wanting a non-default rule should pass
// the rule explicitly via WithBoundaryNodeRule.
func multiLineStringBoundary(ml *geom.MultiLineString) []geom.XY {
	return multiLineStringBoundaryRule(ml, Mod2BoundaryNodeRule)
}

// isMulti reports whether g is a Multi* or GeometryCollection.
func isMulti(g geom.Geometry) bool {
	switch g.(type) {
	case *geom.MultiPoint, *geom.MultiLineString, *geom.MultiPolygon, *geom.GeometryCollection:
		return true
	}
	return false
}

// relateMulti combines per-member relate matrices.
//
// I/B-row × I/B-column cells: cell-wise max over (a_i, b_j) pairs (an
// intersection found in any member-pair contributes its dimension).
//
// *E and E* cells: also max-merged. This is exact when the multi
// geometry members don't overlap each other (the typical case), and
// conservative otherwise. The conservative case over-reports IE/EI/BE/EB
// for overlapping members — acceptable for v1.0 because the OGC spec
// explicitly defines Multi-geometries to have non-overlapping members.
func relateMulti(a, b geom.Geometry, k kernel.Kernel) matrix {
	m := newMatrix()
	m.raise(mEE, 2)

	aMembers := flatten(a)
	bMembers := flatten(b)
	if len(aMembers) == 0 || len(bMembers) == 0 {
		// One side is effectively empty.
		return m
	}

	// Track whether ANY member intersected — used to fix the IE/BE
	// columns: if some member of a has an interior region that doesn't
	// meet any b member's closure, IE picks up dim(a_member).
	for _, ai := range aMembers {
		mi := newMatrix()
		mi.raise(mEE, 2)
		// Default: a's interior/boundary is in b's exterior unless any
		// member intersects.
		di := int8(dimensionOf(ai))
		bdi := int8(boundaryDim(ai))
		if di >= 0 {
			mi.raise(mIE, di)
		}
		if bdi >= 0 {
			mi.raise(mBE, bdi)
		}
		for _, bj := range bMembers {
			pair := relatePair(ai, bj, k)
			// Bring across only the I/B-row × I/B-col cells. The
			// exterior cells (IE/BE on the I/B rows, EI/EB on the
			// E row) are about ext(ai) or ext(bj), not ext(A) or
			// ext(B); we compute those globally after this loop.
			for _, idx := range []int{mII, mIB, mBI, mBB} {
				if pair[idx] > mi[idx] {
					mi[idx] = pair[idx]
				}
			}
		}
		// After processing all bj, IE/BE for ai stays at its default
		// only if ai had no overlap at all; if ai is entirely covered
		// by ∪bj, IE/BE should be F. We approximate via the simple
		// "any member containment" check below.
		if di >= 0 {
			covered := geometryCoveredByAny(ai, bMembers, k)
			if covered {
				mi[mIE] = -1
				if bdi >= 0 {
					mi[mBE] = -1
				}
			} else if bdi >= 0 && boundaryCoveredByAny(ai, bMembers, k) {
				mi[mBE] = -1
			}
		}
		m.merge(mi)
	}

	// EI/EB: at least one b member sticks out of all a members.
	for _, bj := range bMembers {
		dj := int8(dimensionOf(bj))
		bdj := int8(boundaryDim(bj))
		if dj < 0 {
			continue
		}
		if !geometryCoveredByAny(bj, aMembers, k) {
			m.raise(mEI, dj)
			if bdj >= 0 && !boundaryCoveredByAny(bj, aMembers, k) {
				m.raise(mEB, bdj)
			}
		}
	}
	return m
}

// flatten returns the constituent single-typed members of g.
func flatten(g geom.Geometry) []geom.Geometry {
	switch v := g.(type) {
	case *geom.MultiPoint:
		out := make([]geom.Geometry, 0, v.NumGeometries())
		for i := 0; i < v.NumGeometries(); i++ {
			out = append(out, geom.NewPoint(v.CRS(), v.PointAt(i)))
		}
		return out
	case *geom.MultiLineString:
		out := make([]geom.Geometry, 0, v.NumGeometries())
		for i := 0; i < v.NumGeometries(); i++ {
			out = append(out, v.LineStringAt(i))
		}
		return out
	case *geom.MultiPolygon:
		out := make([]geom.Geometry, 0, v.NumGeometries())
		for i := 0; i < v.NumGeometries(); i++ {
			out = append(out, v.PolygonAt(i))
		}
		return out
	case *geom.GeometryCollection:
		out := make([]geom.Geometry, 0, v.NumGeometries())
		for i := 0; i < v.NumGeometries(); i++ {
			out = append(out, flatten(v.GeometryAt(i))...)
		}
		return out
	default:
		if g.IsEmpty() {
			return nil
		}
		return []geom.Geometry{g}
	}
}

// geometryCoveredByAny reports whether g is covered by the union of
// `members` (i.e., every point of g lies in some member's closure).
//
// Fast path: if any single member covers g, return true.
// Slow path (lineal/pointal g): split g into atomic pieces (vertices
// for points, segments for lines) and verify each piece is covered by
// at least one member. Areal g still uses the single-member fast path
// — the cross-member union case is rare for valid inputs.
func geometryCoveredByAny(g geom.Geometry, members []geom.Geometry, k kernel.Kernel) bool {
	for _, m := range members {
		ok, err := Covers(m, g, WithKernel(k))
		if err == nil && ok {
			return true
		}
	}
	// Fallback for lineal/pointal split-coverage cases.
	switch v := g.(type) {
	case *geom.Point:
		return pointCoveredByAny(v.XY(), members, k)
	case *geom.MultiPoint:
		for i := 0; i < v.NumGeometries(); i++ {
			if !pointCoveredByAny(v.PointAt(i), members, k) {
				return false
			}
		}
		return v.NumGeometries() > 0
	case *geom.LineString:
		return lineCoveredByAny(v, members, k)
	case *geom.MultiLineString:
		for i := 0; i < v.NumGeometries(); i++ {
			if !lineCoveredByAny(v.LineStringAt(i), members, k) {
				return false
			}
		}
		return v.NumGeometries() > 0
	}
	return false
}

// boundaryCoveredByAny reports whether the topological boundary of g is
// fully covered by the union of `members`' closures. This is a weaker
// condition than full coverage (which also requires the interior to be
// covered); used to clear the BE cell independently of IE when a polygon's
// boundary fits inside the union of an MP without the polygon being fully
// contained.
func boundaryCoveredByAny(g geom.Geometry, members []geom.Geometry, k kernel.Kernel) bool {
	switch v := g.(type) {
	case *geom.Polygon:
		for r := 0; r < v.NumRings(); r++ {
			ring := v.Ring(r)
			if !ringCoveredByAny(ring, members, k) {
				return false
			}
		}
		return true
	case *geom.MultiPolygon:
		for i := 0; i < v.NumGeometries(); i++ {
			if !boundaryCoveredByAny(v.PolygonAt(i), members, k) {
				return false
			}
		}
		return true
	case *geom.LineString:
		if v.NumPoints() < 2 || v.PointAt(0) == v.PointAt(v.NumPoints()-1) {
			return true
		}
		return pointCoveredByAny(v.PointAt(0), members, k) &&
			pointCoveredByAny(v.PointAt(v.NumPoints()-1), members, k)
	case *geom.MultiLineString:
		bd := multiLineStringBoundary(v)
		for _, p := range bd {
			if !pointCoveredByAny(p, members, k) {
				return false
			}
		}
		return true
	case *geom.Point, *geom.MultiPoint:
		return true
	}
	return false
}

func ringCoveredByAny(ring []geom.XY, members []geom.Geometry, k kernel.Kernel) bool {
	for i := 0; i+1 < len(ring); i++ {
		a, b := ring[i], ring[i+1]
		if a == b {
			continue
		}
		mid := geom.XY{X: (a.X + b.X) / 2, Y: (a.Y + b.Y) / 2}
		if !segmentCoveredByAny(a, b, mid, members, k) {
			return false
		}
	}
	return true
}

func pointCoveredByAny(p geom.XY, members []geom.Geometry, k kernel.Kernel) bool {
	for _, m := range members {
		pt := geom.NewPoint(m.CRS(), p)
		if ok, err := Covers(m, pt, WithKernel(k)); err == nil && ok {
			return true
		}
	}
	return false
}

// lineCoveredByAny reports whether every segment of ls is covered by
// at least one member's closure. Segments are sampled at endpoints +
// midpoint for the test; a segment is "covered" iff all three samples
// are covered by the same single member.
func lineCoveredByAny(ls *geom.LineString, members []geom.Geometry, k kernel.Kernel) bool {
	if ls.NumPoints() < 2 {
		return false
	}
	for i := 0; i+1 < ls.NumPoints(); i++ {
		a := ls.PointAt(i)
		b := ls.PointAt(i + 1)
		mid := geom.XY{X: (a.X + b.X) / 2, Y: (a.Y + b.Y) / 2}
		if !segmentCoveredByAny(a, b, mid, members, k) {
			return false
		}
	}
	return true
}

func segmentCoveredByAny(a, b, mid geom.XY, members []geom.Geometry, k kernel.Kernel) bool {
	// Same-member fast path: many cases are covered by a single member.
	for _, m := range members {
		if pointCoveredByOne(a, m, k) &&
			pointCoveredByOne(b, m, k) &&
			pointCoveredByOne(mid, m, k) {
			return true
		}
	}
	// Union-coverage fallback: a segment may straddle two adjacent
	// members (sharing a boundary). Check that every sample lies in the
	// closure of at least one member, possibly different members per
	// sample. This is a 3-sample approximation that misses true
	// "dipping into the gap" cases for non-adjacent members; valid
	// MultiPolygons have non-overlapping members but they may touch
	// along boundaries, so the 3-sample check is reliable for those
	// touching configurations.
	return pointCoveredByAny(a, members, k) &&
		pointCoveredByAny(b, members, k) &&
		pointCoveredByAny(mid, members, k)
}

func pointCoveredByOne(p geom.XY, m geom.Geometry, k kernel.Kernel) bool {
	pt := geom.NewPoint(m.CRS(), p)
	ok, err := Covers(m, pt, WithKernel(k))
	return err == nil && ok
}

// JTS-compatible DE-9IM matrix pattern constants (mirrors
// org.locationtech.jts.operation.relateng.IntersectionMatrixPattern).
//
//   - PatternAdjacent matches polygonal geometries that share an edge but
//     do not overlap.
//   - PatternContainsProperly matches a geometry whose interior strictly
//     contains another (no boundary contact).
//   - PatternInteriorIntersects matches any pair whose interiors meet.
const (
	PatternAdjacent           = "F***1****"
	PatternContainsProperly   = "T**FF*FF*"
	PatternInteriorIntersects = "T********"
)

// IsIntersects reports whether d corresponds to two geometries with a
// non-empty intersection (any of II/IB/BI/BB non-F).
func (d DE9IM) IsIntersects() bool {
	return d.Matches("T********") ||
		d.Matches("*T*******") ||
		d.Matches("***T*****") ||
		d.Matches("****T****")
}

// IsDisjoint reports whether d corresponds to two geometries with no
// shared points.
func (d DE9IM) IsDisjoint() bool { return !d.IsIntersects() }

// IsContains reports whether d satisfies the OGC contains pattern.
func (d DE9IM) IsContains() bool { return d.Matches("T*****FF*") }

// IsWithin reports whether d satisfies the OGC within pattern.
func (d DE9IM) IsWithin() bool { return d.Matches("T*F**F***") }

// IsCovers reports whether d satisfies any of the OGC covers patterns.
func (d DE9IM) IsCovers() bool {
	return d.Matches("T*****FF*") ||
		d.Matches("*T****FF*") ||
		d.Matches("***T**FF*") ||
		d.Matches("****T*FF*")
}

// IsCoveredBy reports whether d satisfies any of the OGC covered-by
// patterns (the transposes of IsCovers).
func (d DE9IM) IsCoveredBy() bool {
	return d.Matches("T*F**F***") ||
		d.Matches("*TF**F***") ||
		d.Matches("**FT*F***") ||
		d.Matches("**F*TF***")
}

// IsTouches reports whether d satisfies the OGC touches pattern. The
// dimension-pair filter (no Point/Point) is the caller's responsibility.
func (d DE9IM) IsTouches() bool {
	return d.Matches("FT*******") ||
		d.Matches("F**T*****") ||
		d.Matches("F***T****")
}

// IsCrosses reports whether d satisfies the OGC crosses pattern for
// geometries of dimensions dimA and dimB. Same-dim 0/0 and 2/2 are
// undefined and return false.
func (d DE9IM) IsCrosses(dimA, dimB int) bool {
	switch {
	case dimA == 1 && dimB == 1:
		return d.Matches("0********")
	case dimA < dimB:
		return d.Matches("T*T******")
	case dimA > dimB:
		return d.Matches("T*****T**")
	}
	return false
}

// IsOverlaps reports whether d satisfies the OGC overlaps pattern for
// equal-dimension dim. Mixed-dim returns false.
func (d DE9IM) IsOverlaps(dimA, dimB int) bool {
	if dimA != dimB {
		return false
	}
	if dimA == 1 {
		return d.Matches("1*T***T**")
	}
	return d.Matches("T*T***T**")
}

// IsEquals reports whether d satisfies the OGC topological equals
// pattern (II non-empty, no exclusive parts).
func (d DE9IM) IsEquals() bool { return d.Matches("T*F**FFF*") }

// IsContainsProperly reports whether d satisfies the JTS
// "contains properly" pattern (interior strictly contains the other,
// no boundary contact).
func (d DE9IM) IsContainsProperly() bool { return d.Matches(PatternContainsProperly) }

// Matches reports whether the DE-9IM matrix matches the given pattern.
// The pattern uses the same 9-char layout but with extra wildcards:
//
//	'*' — matches any character
//	'T' — matches any of '0','1','2' (i.e. non-empty intersection)
//	'F' — matches only 'F' (empty intersection)
//	'0','1','2' — exact dimension match
func (d DE9IM) Matches(pattern string) bool {
	if len(d) != 9 || len(pattern) != 9 {
		return false
	}
	for i := 0; i < 9; i++ {
		p := pattern[i]
		c := d[i]
		switch p {
		case '*':
			continue
		case 'T':
			if c == 'F' {
				return false
			}
		case 'F':
			if c != 'F' {
				return false
			}
		default:
			if c != p {
				return false
			}
		}
	}
	return true
}
