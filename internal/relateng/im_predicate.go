package relateng

// IMPredicate is an intermediate base struct for predicates whose
// outcome is determined by a DE-9IM intersection matrix. Port of
// org.locationtech.jts.operation.relateng.IMPredicate.
//
// IMPredicate maintains the running matrix and provides the standard
// UpdateDimension that records cell increases (and triggers an
// early-finish via the embedded ValueIM hook when the predicate
// reports IsDetermined).
//
// Concrete IM predicates implement two strategy methods via the
// IMStrategy interface: IsDetermined (can the value be settled now?)
// and ValueIM (compute the value from the current matrix). The
// IMPredicate uses these via interface dispatch from the embedded
// owner, so subclasses must call BindOwner once at construction.
type IMPredicate struct {
	BasicPredicate

	DimA, DimB int
	IM         *IntersectionMatrix
	owner      IMStrategy
}

// IMStrategy is the per-predicate decision logic embedded by an
// IMPredicate. The owner is set via BindOwner.
type IMStrategy interface {
	// IsDetermined reports whether the predicate's value can be
	// settled given the current state of the matrix.
	IsDetermined() bool
	// ValueIM computes the predicate's boolean value from the
	// current matrix (called when IsDetermined is true, or in
	// Finish).
	ValueIM() bool
}

// IsDimsCompatibleWithCovers mirrors the JTS helper of the same
// name. The bigger geometry's dim must be >= the smaller's, with one
// special case: a Point can be covered by a zero-length Line (both
// are topologically a point).
func IsDimsCompatibleWithCovers(dim0, dim1 int) bool {
	if dim0 == DimP && dim1 == DimL {
		return true
	}
	return dim0 >= dim1
}

// NewIMPredicate constructs a fresh IM predicate. The owner is
// bound via BindOwner to enable interface dispatch from
// UpdateDimension/Finish.
func NewIMPredicate() *IMPredicate {
	im := NewIntersectionMatrix()
	// E/E is always 2 (the matching exterior is the whole plane).
	im.Set(LocExterior, LocExterior, DimA)
	return &IMPredicate{IM: im}
}

// BindOwner records the embedding implementation so the base struct
// can dispatch IsDetermined/ValueIM via the interface. Must be
// called after construction.
func (p *IMPredicate) BindOwner(o IMStrategy) { p.owner = o }

// InitDim records dimensions and is a hook for subclass overrides.
func (p *IMPredicate) InitDim(dimA, dimB int) {
	p.DimA = dimA
	p.DimB = dimB
}

// UpdateDimension records a cell increase and triggers an
// early-finish if the strategy reports the predicate is determined.
func (p *IMPredicate) UpdateDimension(locA, locB, dim int) {
	if dim > p.IM.Get(locA, locB) {
		p.IM.Set(locA, locB, dim)
		if p.owner != nil && p.owner.IsDetermined() {
			p.SetValue(p.owner.ValueIM())
		}
	}
}

// Finish settles the value from the matrix's current state.
func (p *IMPredicate) Finish() {
	if p.owner != nil {
		p.SetValue(p.owner.ValueIM())
	}
}

// IsIntersects reports whether the (locA, locB) cell is non-empty.
func (p *IMPredicate) IsIntersects(locA, locB int) bool {
	return p.IM.Get(locA, locB) >= DimP
}

// IsDimension reports whether the cell value equals dim exactly.
func (p *IMPredicate) IsDimension(locA, locB, dim int) bool {
	return p.IM.Get(locA, locB) == dim
}

// IntersectsExteriorOf reports whether the exterior of the named
// geometry is intersected by any part of the other geometry.
func (p *IMPredicate) IntersectsExteriorOf(isA bool) bool {
	if isA {
		return p.IsIntersects(LocExterior, LocInterior) ||
			p.IsIntersects(LocExterior, LocBoundary)
	}
	return p.IsIntersects(LocInterior, LocExterior) ||
		p.IsIntersects(LocBoundary, LocExterior)
}
