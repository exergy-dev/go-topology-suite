package relateng

// RelateMatrixPredicate is the TopologyPredicate used to drive a full
// DE-9IM matrix computation. Port of
// org.locationtech.jts.operation.relateng.RelateMatrixPredicate.
//
// Unlike pattern-matching predicates this one never short-circuits —
// every reported cell is recorded, and Finish leaves the result in
// the embedded IntersectionMatrix for callers to inspect.
type RelateMatrixPredicate struct {
	IMPredicate
}

// NewRelateMatrixPredicate constructs a predicate that accumulates
// every reported DE-9IM cell. Bind owner is wired in the constructor.
func NewRelateMatrixPredicate() *RelateMatrixPredicate {
	p := &RelateMatrixPredicate{IMPredicate: *NewIMPredicate()}
	p.BindOwner(p)
	return p
}

// Name returns "relateMatrix".
func (*RelateMatrixPredicate) Name() string { return "relateMatrix" }

// RequireSelfNoding returns true — a faithful matrix computation
// requires self-noded inputs whenever lines may self-cross.
func (*RelateMatrixPredicate) RequireSelfNoding() bool { return true }

// RequireInteraction returns false — even disjoint envelopes need
// the per-cell exterior fill so the matrix is complete.
func (*RelateMatrixPredicate) RequireInteraction() bool { return false }

// RequireExteriorCheck returns true — the matrix must record E-row
// and E-column cells.
func (*RelateMatrixPredicate) RequireExteriorCheck(bool) bool { return true }

// IsDetermined always returns false; this predicate runs to
// completion and only resolves its boolean value via Finish.
func (*RelateMatrixPredicate) IsDetermined() bool { return false }

// ValueIM always returns true: the matrix is the result, the boolean
// channel is unused. (Mirrors JTS, where the matrix is consumed via
// GetIM after evaluation.)
func (*RelateMatrixPredicate) ValueIM() bool { return true }

// Matrix returns the running intersection matrix. Callers should only
// inspect after the topology computation has finished.
func (p *RelateMatrixPredicate) Matrix() *IntersectionMatrix { return p.IM }
