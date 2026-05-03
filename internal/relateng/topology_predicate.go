package relateng

import "github.com/terra-geo/terra/geom"

// TopologyPredicate is the strategy interface implemented by all
// RelateNG predicates. Port of
// org.locationtech.jts.operation.relateng.TopologyPredicate.
//
// The TopologyComputer feeds the predicate via UpdateDimension as it
// discovers DE-9IM cells; the predicate is free to call SetValue at
// any point to short-circuit the computation. Once IsKnown returns
// true the computer stops feeding cells.
type TopologyPredicate interface {
	// Name returns the predicate's identifier (e.g. "intersects").
	Name() string

	// RequireSelfNoding reports whether the predicate needs the
	// inputs to be self-noded (lines pre-split at self-crossings).
	// Most predicates do; intersects/disjoint do not.
	RequireSelfNoding() bool

	// RequireInteraction reports whether non-empty interaction in
	// at least one of II/IB/BI/BB is needed. If true, disjoint
	// envelopes can short-circuit to false.
	RequireInteraction() bool

	// RequireCovers reports whether the predicate requires the
	// source to cover the target (i.e. exterior-of-source vs
	// interior/boundary-of-target cells must be empty). If true,
	// envelope-doesn't-cover short-circuits to false.
	RequireCovers(isSourceA bool) bool

	// RequireExteriorCheck reports whether the predicate cares
	// about source-vs-exterior-of-target cells. When false, the
	// computer can skip exterior-of-target work for that source.
	RequireExteriorCheck(isSourceA bool) bool

	// InitDim is called once with the dimensions of A and B before
	// any cells are reported. Predicates may set their value if
	// the result is determined by dim alone.
	InitDim(dimA, dimB int)

	// InitEnv is called once with the envelopes of A and B. Same
	// purpose as InitDim but envelope-driven.
	InitEnv(envA, envB geom.Envelope)

	// UpdateDimension reports a cell of the DE-9IM matrix. The
	// predicate must accumulate (via max or whatever rule it
	// chooses) and may set its value early.
	UpdateDimension(locA, locB, dim int)

	// Finish is called once after all cells are reported, allowing
	// the predicate to compute its value if not already known.
	Finish()

	// IsKnown reports whether the predicate's value is final.
	IsKnown() bool

	// Value returns the final boolean value (only valid when
	// IsKnown is true).
	Value() bool
}
