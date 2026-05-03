package relateng

// BoundaryNodeRule decides whether a vertex shared by some number of
// component-boundary endpoints is itself a boundary point of the
// parent geometry. Mirrors org.locationtech.jts.algorithm.BoundaryNodeRule
// in shape, but kept local to the relateng package to avoid an import
// cycle with predicate (which has its own copy of the rule).
//
// Stateless value types — pass by value, no allocation.
type BoundaryNodeRule interface {
	// IsInBoundary reports whether a vertex incident to
	// `boundaryCount` component endpoints is a boundary point.
	IsInBoundary(boundaryCount int) bool
}

// Mod2BoundaryRule is the OGC SFS rule: odd-valency endpoints are on
// the boundary. Closed rings have empty boundary.
type Mod2BoundaryRule struct{}

// IsInBoundary returns true when boundaryCount is odd.
func (Mod2BoundaryRule) IsInBoundary(boundaryCount int) bool {
	return boundaryCount%2 == 1
}

// EndpointBoundaryRule treats every endpoint as a boundary point.
type EndpointBoundaryRule struct{}

// IsInBoundary returns true when boundaryCount > 0.
func (EndpointBoundaryRule) IsInBoundary(boundaryCount int) bool {
	return boundaryCount > 0
}

// MultiValentEndpointBoundaryRule keeps only valency > 1.
type MultiValentEndpointBoundaryRule struct{}

// IsInBoundary returns true when boundaryCount > 1.
func (MultiValentEndpointBoundaryRule) IsInBoundary(boundaryCount int) bool {
	return boundaryCount > 1
}

// MonoValentEndpointBoundaryRule keeps only valency == 1.
type MonoValentEndpointBoundaryRule struct{}

// IsInBoundary returns true when boundaryCount == 1.
func (MonoValentEndpointBoundaryRule) IsInBoundary(boundaryCount int) bool {
	return boundaryCount == 1
}

// OGCSFSBoundaryRule is the default Mod2 rule.
var OGCSFSBoundaryRule BoundaryNodeRule = Mod2BoundaryRule{}
