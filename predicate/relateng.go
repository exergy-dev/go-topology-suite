package predicate

import (
	"github.com/terra-geo/terra/geom"
)

// RelateNG is an Experimental skeleton port of JTS RelateNG
// (`org.locationtech.jts.operation.relateng.RelateNG`). It exposes the
// same predicate menu as RelateOp but routes each predicate through the
// short-circuit fast-path layer (relate_short_circuit.go) before
// falling back to the existing DE-9IM topology graph.
//
// The full RelateNG architecture in JTS — incremental TopologyComputer,
// per-predicate value tracking, lazy edge intersection — is a larger
// port (≈2000 LOC) that has not yet landed in Terra. This type is the
// public API hook that future work will fill out; today its predicate
// methods are simple wrappers over the package-level functions, which
// already incorporate the relevant envelope/dim short-circuits.
//
// The boundary node rule may be overridden via WithBoundaryNodeRule.
// The kernel is selected automatically from the geometry's CRS unless
// a non-default kernel is supplied via WithKernel.
//
// Experimental: API and behaviour may change. Prefer the package-level
// predicate functions for stable code.
type RelateNG struct {
	a    geom.Geometry
	opts []Option
}

// NewRelateNG constructs a RelateNG for geometry a. The other operand
// is supplied per call. This mirrors JTS's `RelateNG.relate(g)` /
// `RelateNG.evaluate(g, predicate)` pattern.
//
// When the same geometry will participate in many predicate calls
// (e.g. one polygon vs many points), prefer also passing
// `WithPrepared(prepare.Polygon(a))` via opts.
func NewRelateNG(a geom.Geometry, opts ...Option) *RelateNG {
	return &RelateNG{a: a, opts: opts}
}

// Intersects: A intersects B.
func (r *RelateNG) Intersects(b geom.Geometry) (bool, error) {
	return Intersects(r.a, b, r.opts...)
}

// Disjoint: A and B share no points.
func (r *RelateNG) Disjoint(b geom.Geometry) (bool, error) {
	return Disjoint(r.a, b, r.opts...)
}

// Contains: every point of B lies in A's interior or boundary, and
// interiors meet.
func (r *RelateNG) Contains(b geom.Geometry) (bool, error) {
	return Contains(r.a, b, r.opts...)
}

// Within: every point of A lies in B's interior or boundary, and
// interiors meet. (Converse of Contains.)
func (r *RelateNG) Within(b geom.Geometry) (bool, error) {
	return Within(r.a, b, r.opts...)
}

// Covers: every point of B lies in A's closure (interior + boundary).
func (r *RelateNG) Covers(b geom.Geometry) (bool, error) {
	return Covers(r.a, b, r.opts...)
}

// CoveredBy: every point of A lies in B's closure.
func (r *RelateNG) CoveredBy(b geom.Geometry) (bool, error) {
	return CoveredBy(r.a, b, r.opts...)
}

// Crosses: per OGC, dim-dependent crossing predicate.
func (r *RelateNG) Crosses(b geom.Geometry) (bool, error) {
	return Crosses(r.a, b, r.opts...)
}

// Overlaps: same-dimension partial intersection.
func (r *RelateNG) Overlaps(b geom.Geometry) (bool, error) {
	return Overlaps(r.a, b, r.opts...)
}

// Touches: shared boundary, no interior intersection.
func (r *RelateNG) Touches(b geom.Geometry) (bool, error) {
	return Touches(r.a, b, r.opts...)
}

// Equals: topological equality (DE-9IM T*F**FFF*).
func (r *RelateNG) Equals(b geom.Geometry) (bool, error) {
	return Equals(r.a, b, r.opts...)
}

// Relate computes the full DE-9IM intersection matrix for A vs B.
func (r *RelateNG) Relate(b geom.Geometry) (DE9IM, error) {
	return Relate(r.a, b, r.opts...)
}
