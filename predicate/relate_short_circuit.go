package predicate

import (
	"github.com/exergy-dev/go-topology-suite/geom"
)

// This file ports the SHORT-CIRCUIT layer of JTS RelateNG (the
// `org.locationtech.jts.operation.relateng` package, classes
// RelatePredicate / BasicPredicate / IMPredicate). RelateNG's primary
// optimisation over the classic RelateOp is that each predicate signals
// the minimum information it needs from the topology graph; many
// queries can be answered purely from envelopes and dimensions without
// ever building a graph.
//
// We don't (yet) port the incremental TopologyComputer — that's a
// multi-day effort. Instead we centralise the envelope/dim fast paths
// here so every public predicate (Intersects, Disjoint, Contains,
// Within, Covers, CoveredBy, Overlaps, Touches, Crosses, Equals) gets
// consistent, well-documented short-circuiting before falling back to
// the full DE-9IM matrix in relate.go.
//
// JTS reference (per-predicate fast paths):
//
//   - intersects:  init(envA, envB) → require envA.intersects(envB)
//   - disjoint:    init(envA, envB) → setValueIf(true, envA.disjoint(envB))
//   - contains:    init(envA, envB) → requireCovers(envA, envB)
//                  init(dimA, dimB) → require(dimsCompatibleWithCovers(dimA, dimB))
//   - within:      init(envA, envB) → requireCovers(envB, envA)
//                  init(dimA, dimB) → require(dimsCompatibleWithCovers(dimB, dimA))
//   - covers:      init(envA, envB) → requireCovers(envA, envB)
//                  init(dimA, dimB) → require(dimsCompatibleWithCovers(dimA, dimB))
//   - coveredBy:   init(envA, envB) → requireCovers(envB, envA)
//                  init(dimA, dimB) → require(dimsCompatibleWithCovers(dimB, dimA))
//   - crosses:     init(dimA, dimB) → require(!isBothPointsOrAreas)
//                  init(envA, envB) → require(envA.intersects(envB))
//   - overlaps:    init(dimA, dimB) → require(dimA == dimB)
//                  init(envA, envB) → require(envA.intersects(envB))
//   - touches:     init(dimA, dimB) → require(!(dimA == 0 && dimB == 0))
//                  init(envA, envB) → require(envA.intersects(envB))
//   - equals:      init(envA, envB) → setValueIf(true, both null);
//                                     require(envA.equals(envB))
//
// `dimsCompatibleWithCovers(big, small)` ↔ `big >= small` in JTS
// IMPredicate: a 1-D geometry can never cover a 2-D one, etc.

// shortCircuit represents a tri-state outcome from the fast path:
// resolved=true means the value is final; resolved=false means the
// caller must run the full DE-9IM logic.
type shortCircuit struct {
	resolved bool
	value    bool
}

func known(v bool) shortCircuit  { return shortCircuit{resolved: true, value: v} }
func unresolved() shortCircuit   { return shortCircuit{} }
func (s shortCircuit) get() bool { return s.value }

// scIntersects handles the `intersects` fast path.
//
// JTS: `init(envA,envB)` requires envA.intersects(envB); on failure the
// predicate is set to false. If envelopes intersect, fall through.
func scIntersects(a, b geom.Geometry, planar bool) shortCircuit {
	if a.IsEmpty() || b.IsEmpty() {
		return known(false)
	}
	if planar && !a.Envelope().Intersects(b.Envelope()) {
		return known(false)
	}
	return unresolved()
}

// scCovers handles the covers/contains family fast paths. The dim and
// envelope checks are identical for both predicates (RelateNG's
// `requireCovers(envA, envB)` plus `isDimsCompatibleWithCovers(dimA,
// dimB)`).
//
// Covers requires the bigger geometry's envelope to cover the smaller's,
// and the bigger geometry's dimension to be >= the smaller's. (A line
// cannot cover an area; a point cannot cover a line.)
//
// The dim-mismatch short-circuit is only applied when both operands are
// "regular" (non-collection, non-empty, and lines have non-zero
// length). Collections may report a higher dimension via an empty
// member, and a zero-length line is topologically a point — JTS
// matches Point.contains(zero-length-line)=true. Restricting the
// short-circuit to regular shapes avoids those edge cases.
func scCovers(a, b geom.Geometry, planar bool) shortCircuit {
	if a.IsEmpty() || b.IsEmpty() {
		return known(false)
	}
	if isRegularShape(a) && isRegularShape(b) && dimensionOf(a) < dimensionOf(b) {
		return known(false)
	}
	if planar && !a.Envelope().Contains(b.Envelope()) {
		return known(false)
	}
	return unresolved()
}

// isRegularShape reports whether g's reported dimensionOf() faithfully
// reflects its topological dimension. Excludes:
//   - Heterogeneous GeometryCollections (may contain empty higher-dim
//     members that inflate the reported dim).
//   - Zero-length LineStrings (topologically a point, dim 0, but
//     dimensionOf reports 1).
//
// Multi* (MultiPoint / MultiLineString / MultiPolygon) are homogeneous,
// so their dimension is unambiguous and they ARE regular. Zero-area
// degenerate polygons remain topologically 2-D for relate purposes,
// so they are also regular.
func isRegularShape(g geom.Geometry) bool {
	switch v := g.(type) {
	case *geom.GeometryCollection:
		// A heterogeneous collection or a collection with empty
		// higher-dim members could mis-report dimensionOf. Be
		// conservative.
		_ = v
		return false
	case *geom.LineString:
		if isZeroLengthLine(v) {
			return false
		}
	}
	return true
}

// scContains is identical to scCovers from the short-circuit layer's
// perspective; the difference between contains/covers (boundary
// handling) only shows up in the matrix.
func scContains(a, b geom.Geometry, planar bool) shortCircuit {
	return scCovers(a, b, planar)
}

// scOverlaps handles dim-mismatch + envelope short-circuits for the
// overlaps predicate. JTS requires dimA == dimB; mismatched dimensions
// return false without inspecting the graph.
//
// Unlike the contains/covers family, the dim-mismatch check is safe to
// apply unconditionally here: an empty higher-dim member in a
// GeometryCollection cannot create an overlap that wasn't there before
// (the empty contributes no point set), and Overlaps is the original
// pre-short-circuit predicate's documented behaviour anyway.
func scOverlaps(a, b geom.Geometry, planar bool) shortCircuit {
	if a.IsEmpty() || b.IsEmpty() {
		return known(false)
	}
	if dimensionOf(a) != dimensionOf(b) {
		return known(false)
	}
	if planar && !a.Envelope().Intersects(b.Envelope()) {
		return known(false)
	}
	return unresolved()
}

// scCrosses handles the JTS crosses init checks: P/P and A/A always
// return false; envelopes must intersect. The dim-based short-circuit
// is only applied when both operands are regular shapes.
func scCrosses(a, b geom.Geometry, planar bool) shortCircuit {
	if a.IsEmpty() || b.IsEmpty() {
		return known(false)
	}
	if isRegularShape(a) && isRegularShape(b) {
		dA, dB := dimensionOf(a), dimensionOf(b)
		bothPoints := dA == 0 && dB == 0
		bothAreas := dA == 2 && dB == 2
		if bothPoints || bothAreas {
			return known(false)
		}
	}
	if planar && !a.Envelope().Intersects(b.Envelope()) {
		return known(false)
	}
	return unresolved()
}

// scTouches: P/P always false; envelopes must at least touch.
func scTouches(a, b geom.Geometry, planar bool) shortCircuit {
	if a.IsEmpty() || b.IsEmpty() {
		return known(false)
	}
	// Two pure points share no boundary, so cannot touch. JTS Multi
	// points are excluded here: a MultiPoint can touch via shared
	// points being interior to one and boundary to nothing — but JTS
	// itself permits the touches relation only when at least one
	// geometry has a boundary. Mirror our existing touches.go behaviour
	// (MultiPoint kept on the slow path). Restrict to regular shapes
	// to avoid empty-collection edge cases.
	if isRegularShape(a) && isRegularShape(b) &&
		dimensionOf(a) == 0 && dimensionOf(b) == 0 &&
		!isMulti(a) && !isMulti(b) {
		return known(false)
	}
	if planar && !a.Envelope().Intersects(b.Envelope()) {
		return known(false)
	}
	return unresolved()
}

// scEquals: empty=empty true; envelope mismatch false.
func scEquals(a, b geom.Geometry, planar bool) shortCircuit {
	if a.IsEmpty() && b.IsEmpty() {
		return known(true)
	}
	if a.IsEmpty() != b.IsEmpty() {
		return known(false)
	}
	if planar {
		ea, eb := a.Envelope(), b.Envelope()
		if ea != eb {
			return known(false)
		}
	}
	return unresolved()
}
