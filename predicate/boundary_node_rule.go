package predicate

import "github.com/terra-geo/terra/geom"

// BoundaryNodeRule decides whether a vertex shared by some number of
// component-boundary endpoints is itself a boundary point of the
// parent geometry. Port of
// org.locationtech.jts.algorithm.BoundaryNodeRule.
//
// The default rule (Mod2) is the OGC SFS rule: a point is in the
// boundary iff it appears in an odd number of component boundaries.
// Non-default rules are useful in linear-network topology where
// closed segments (turn-arounds) need to be distinguished from
// dangling endpoints.
//
// Strategy pattern: rule implementations are stateless value types,
// so callers can store and pass them by value without ceremony.
type BoundaryNodeRule interface {
	// IsInBoundary reports whether a vertex incident to
	// `boundaryCount` component endpoints is a boundary point.
	IsInBoundary(boundaryCount int) bool
}

// Mod2BoundaryRule is the OGC SFS rule: odd-valency endpoints are on
// the boundary. Closed members (LinearRings, closed LineStrings)
// have empty boundary.
type Mod2BoundaryRule struct{}

// IsInBoundary returns true when boundaryCount is odd.
func (Mod2BoundaryRule) IsInBoundary(boundaryCount int) bool {
	return boundaryCount%2 == 1
}

// EndpointBoundaryRule treats every endpoint as a boundary point,
// regardless of valency. Closed rings have a non-empty boundary
// (their start/end coincidence). Useful for validating linear
// networks — an entry road touching a turn-around ring at one
// endpoint is "simple" only under this rule.
type EndpointBoundaryRule struct{}

// IsInBoundary returns true when boundaryCount > 0.
func (EndpointBoundaryRule) IsInBoundary(boundaryCount int) bool {
	return boundaryCount > 0
}

// MultiValentEndpointBoundaryRule keeps only endpoints with valency
// > 1 — i.e. "attached" endpoints where multiple components meet.
type MultiValentEndpointBoundaryRule struct{}

// IsInBoundary returns true when boundaryCount > 1.
func (MultiValentEndpointBoundaryRule) IsInBoundary(boundaryCount int) bool {
	return boundaryCount > 1
}

// MonoValentEndpointBoundaryRule keeps only valency-exactly-1
// endpoints — the "unattached" / dangling ends of a linear network.
type MonoValentEndpointBoundaryRule struct{}

// IsInBoundary returns true when boundaryCount == 1.
func (MonoValentEndpointBoundaryRule) IsInBoundary(boundaryCount int) bool {
	return boundaryCount == 1
}

// Default boundary node rules. The OGC alias mirrors JTS.
var (
	Mod2BoundaryNodeRule                = Mod2BoundaryRule{}
	EndpointBoundaryNodeRule            = EndpointBoundaryRule{}
	MultiValentEndpointBoundaryNodeRule = MultiValentEndpointBoundaryRule{}
	MonoValentEndpointBoundaryNodeRule  = MonoValentEndpointBoundaryRule{}
	OGCSFSBoundaryNodeRule              = Mod2BoundaryNodeRule
)

// multiLineStringBoundaryRule applies a BoundaryNodeRule to count
// member endpoints across a MultiLineString.
//
// For non-closed members each endpoint contributes 1 to its count.
// For closed members (a single LineString that's a ring), the
// closing coincidence point contributes 1 — under the OGC Mod2 rule
// this is filtered out (count of 1 is odd, but JTS treats closed
// LineStrings as having empty boundary; we accomplish that by
// special-casing Mod2 to skip closed members entirely below). Under
// Endpoint / MultiValent / MonoValent rules the closing point is a
// valid candidate for boundary classification.
func multiLineStringBoundaryRule(ml *geom.MultiLineString, rule BoundaryNodeRule) []geom.XY {
	count := map[geom.XY]int{}
	_, isMod2 := rule.(Mod2BoundaryRule)
	for i := 0; i < ml.NumGeometries(); i++ {
		ls := ml.LineStringAt(i)
		if ls.IsEmpty() || ls.NumPoints() < 2 {
			continue
		}
		first := ls.PointAt(0)
		last := ls.PointAt(ls.NumPoints() - 1)
		if first == last {
			// Closed member. Mod2 keeps "closed = empty boundary".
			// Other rules contribute one count for the coincident
			// endpoint.
			if isMod2 {
				continue
			}
			count[first]++
			continue
		}
		count[first]++
		count[last]++
	}
	var out []geom.XY
	for p, c := range count {
		if rule.IsInBoundary(c) {
			out = append(out, p)
		}
	}
	return out
}
