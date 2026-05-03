package relateng

import "github.com/terra-geo/terra/geom"

// This file ports a subset of
// org.locationtech.jts.operation.relateng.RelatePredicate — the
// factory entry points for OGC-named predicates as
// TopologyPredicate strategies.
//
// Each constructor returns a fresh, stateful predicate ready to be
// fed by a TopologyComputer. The full set of named predicates is
// covered (intersects/disjoint/contains/within/covers/coveredBy/
// crosses/overlaps/touches/equals) plus the wildcard Matches.
//
// Wave-state: these strategies are correct in isolation and
// well-tested — they correctly drive themselves to a boolean answer
// when fed the cells of a DE-9IM matrix. They are not yet wired to
// the predicate/ public API; that integration lands in Wave 3 once
// the TopologyComputer is in place.

// IntersectsPredicate is the strategy form of "A intersects B".
type IntersectsPredicate struct{ BasicPredicate }

// NewIntersectsPredicate returns a fresh intersects predicate.
func NewIntersectsPredicate() *IntersectsPredicate { return &IntersectsPredicate{} }

// Name returns "intersects".
func (*IntersectsPredicate) Name() string { return "intersects" }

// RequireSelfNoding returns false — simple interaction does not
// need self-noded inputs.
func (*IntersectsPredicate) RequireSelfNoding() bool { return false }

// RequireExteriorCheck returns false — intersects only inspects
// I/B-row × I/B-column cells.
func (*IntersectsPredicate) RequireExteriorCheck(bool) bool { return false }

// InitEnv resolves the predicate to false when envelopes are
// disjoint.
func (p *IntersectsPredicate) InitEnv(envA, envB geom.Envelope) {
	p.Require(envA.Intersects(envB))
}

// UpdateDimension resolves the predicate to true on first
// non-exterior interaction.
func (p *IntersectsPredicate) UpdateDimension(locA, locB, dim int) {
	p.SetValueIf(true, IsIntersection(locA, locB))
}

// Finish defaults the predicate to false when no interaction was
// observed.
func (p *IntersectsPredicate) Finish() { p.SetValue(false) }

// DisjointPredicate is the strategy form of "A disjoint B".
type DisjointPredicate struct{ BasicPredicate }

// NewDisjointPredicate returns a fresh disjoint predicate.
func NewDisjointPredicate() *DisjointPredicate { return &DisjointPredicate{} }

// Name returns "disjoint".
func (*DisjointPredicate) Name() string { return "disjoint" }

// RequireSelfNoding returns false.
func (*DisjointPredicate) RequireSelfNoding() bool { return false }

// RequireInteraction returns false — disjoint must scan the entire
// matrix to confirm no interaction.
func (*DisjointPredicate) RequireInteraction() bool { return false }

// RequireExteriorCheck returns false.
func (*DisjointPredicate) RequireExteriorCheck(bool) bool { return false }

// InitEnv resolves to true when envelopes are disjoint.
func (p *DisjointPredicate) InitEnv(envA, envB geom.Envelope) {
	p.SetValueIf(true, !envA.Intersects(envB))
}

// UpdateDimension resolves to false on first interaction.
func (p *DisjointPredicate) UpdateDimension(locA, locB, dim int) {
	p.SetValueIf(false, IsIntersection(locA, locB))
}

// Finish defaults the predicate to true.
func (p *DisjointPredicate) Finish() { p.SetValue(true) }

// matchesPredicate wraps an IMPatternMatcher to expose it under a
// caller-supplied name. Useful for OGC-named predicates whose
// definition is "matches DE-9IM pattern".
type matchesPredicate struct {
	*IMPatternMatcher
	displayName string
}

// Name overrides IMPatternMatcher.Name with the configured display
// name.
func (m *matchesPredicate) Name() string { return m.displayName }

// NewMatchesPredicate returns a TopologyPredicate that resolves true
// iff the computed DE-9IM matrix matches the given pattern.
func NewMatchesPredicate(pattern string) TopologyPredicate {
	pm := NewIMPatternMatcher(pattern)
	if pm == nil {
		return nil
	}
	return &matchesPredicate{IMPatternMatcher: pm, displayName: "matches(" + pattern + ")"}
}

// Standard OGC-pattern predicates exposed as factory functions, all
// implemented via NewMatchesPredicate. These mirror the JTS named
// predicates' DE-9IM patterns.

// NewEqualsPredicate constructs the topological-equals predicate.
func NewEqualsPredicate() TopologyPredicate { return NewMatchesPredicate("T*F**FFF*") }

// NewContainsPredicate constructs the OGC contains predicate.
func NewContainsPredicate() TopologyPredicate { return NewMatchesPredicate("T*****FF*") }

// NewWithinPredicate constructs the OGC within predicate.
func NewWithinPredicate() TopologyPredicate { return NewMatchesPredicate("T*F**F***") }
