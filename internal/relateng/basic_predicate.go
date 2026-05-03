package relateng

import "github.com/terra-geo/terra/geom"

// triState models the TRUE / FALSE / UNKNOWN state of a predicate
// during incremental evaluation. Mirrors JTS BasicPredicate's
// integer-based UNKNOWN/FALSE/TRUE.
//
// We choose 0 as UNKNOWN so the zero value of a freshly-constructed
// BasicPredicate is "unknown" — no explicit initializer needed.
type triState int8

const (
	triUnknown triState = 0
	triFalse   triState = 1
	triTrue    triState = 2
)

// BasicPredicate is a base struct for boolean-valued
// TopologyPredicate implementations. It wires the tri-state value,
// the SetValue/SetValueIf/Require helpers, and the Done() check.
//
// Port of org.locationtech.jts.operation.relateng.BasicPredicate.
//
// Concrete predicates embed BasicPredicate and override the
// strategy methods (UpdateDimension, Finish, ...).
type BasicPredicate struct {
	value triState
}

// IsIntersection reports whether locA and locB combine to a
// non-empty intersection (neither is exterior).
func IsIntersection(locA, locB int) bool {
	return locA != LocExterior && locB != LocExterior
}

// IsKnown reports whether the predicate has been resolved.
func (p *BasicPredicate) IsKnown() bool { return p.value != triUnknown }

// Value returns the final value (only meaningful once IsKnown is
// true).
func (p *BasicPredicate) Value() bool { return p.value == triTrue }

// SetValue updates the predicate iff still unknown.
func (p *BasicPredicate) SetValue(v bool) {
	if p.value != triUnknown {
		return
	}
	if v {
		p.value = triTrue
	} else {
		p.value = triFalse
	}
}

// SetValueIf updates the predicate to v iff cond is true.
func (p *BasicPredicate) SetValueIf(v bool, cond bool) {
	if cond {
		p.SetValue(v)
	}
}

// Require sets the value to false unless cond is true. Used to
// thread short-circuit pre-conditions: `Require(envA.Intersects(envB))`
// resolves the predicate to false when envelopes are disjoint.
func (p *BasicPredicate) Require(cond bool) {
	if !cond {
		p.SetValue(false)
	}
}

// RequireEnvCovers sets the value to false when envA does not
// cover envB. Convenience wrapper used by the cover/contains family.
//
// Renamed from JTS's `requireCovers(Envelope,Envelope)` to avoid a
// name clash with the TopologyPredicate.RequireCovers(bool) bool
// strategy method.
func (p *BasicPredicate) RequireEnvCovers(envA, envB geom.Envelope) {
	p.Require(envA.Contains(envB))
}

// RequireCovers is the default for the TopologyPredicate strategy
// method: most predicates do not require covers semantics.
func (p *BasicPredicate) RequireCovers(isSourceA bool) bool { return false }

// Default RequireSelfNoding/RequireInteraction/RequireCovers/
// RequireExteriorCheck for embedded use; concrete predicates may
// override individually.

// RequireSelfNoding returns true (the conservative default).
func (p *BasicPredicate) RequireSelfNoding() bool { return true }

// RequireInteraction returns true (most predicates do require
// interaction; envelope-disjoint inputs can be answered without it
// only for disjoint() which overrides this).
func (p *BasicPredicate) RequireInteraction() bool { return true }

// RequireExteriorCheck returns true (default: predicates inspect
// the *E columns).
func (p *BasicPredicate) RequireExteriorCheck(isSourceA bool) bool { return true }

// InitDim is a no-op default. Concrete predicates override.
func (p *BasicPredicate) InitDim(dimA, dimB int) {}

// InitEnv is a no-op default.
func (p *BasicPredicate) InitEnv(envA, envB geom.Envelope) {}

// Finish is a no-op default. IM-based predicates override.
func (p *BasicPredicate) Finish() {}
