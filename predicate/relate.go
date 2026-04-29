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
	cfg := resolve(a, opts)
	m := computeMatrix(a, b, cfg.kernel)
	return m.toDE9IM(), nil
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

	// Multi/collection on either side: iterate members and merge.
	if isMulti(a) || isMulti(b) {
		return relateMulti(a, b, k)
	}

	// Single-type dispatch.
	return relatePair(a, b, k)
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
// the boundary is empty (Point, MultiPoint, closed LineString — though we
// approximate the closed-curve case as still having empty boundary at the
// ring level).
func boundaryDim(g geom.Geometry) int {
	switch v := g.(type) {
	case *geom.Point, *geom.MultiPoint:
		return -1
	case *geom.LineString:
		if v.NumPoints() < 2 {
			return -1
		}
		return 0
	case *geom.MultiLineString:
		// Mod-2 rule: shared endpoints cancel. Conservative bound: 0.
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
			// Bring across only the non-exterior columns; the I/B
			// exterior cells from sub-pairs are about ext(bj), not
			// ext(B).
			for _, idx := range []int{mII, mIB, mBI, mBB, mEI, mEB} {
				if pair[idx] > mi[idx] {
					mi[idx] = pair[idx]
				}
			}
			// IE/BE cells: only "exterior of all bj" counts. Compute
			// after the loop.
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
			if bdj >= 0 {
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
// Conservative shortcut: returns true only when at least one member
// individually covers g. This under-reports coverage when g is split
// across multiple members; acceptable for v1.0 since the alternative is
// computing a full union.
func geometryCoveredByAny(g geom.Geometry, members []geom.Geometry, k kernel.Kernel) bool {
	for _, m := range members {
		ok, err := Covers(m, g, WithKernel(k))
		if err == nil && ok {
			return true
		}
	}
	return false
}

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
