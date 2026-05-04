package relateng

import "github.com/exergy-dev/go-topology-suite/geom"

// LinearBoundary determines the boundary points of a linear geometry,
// using a BoundaryNodeRule. Port of
// org.locationtech.jts.operation.relateng.LinearBoundary.
//
// A "boundary point" of a MultiLineString under a given rule is a
// vertex whose endpoint-incidence count satisfies the rule's
// IsInBoundary test (e.g. odd under the Mod2 rule).
type LinearBoundary struct {
	vertexDegree map[geom.XY]int
	hasBoundary  bool
	rule         BoundaryNodeRule
}

// NewLinearBoundary builds the boundary for the supplied LineStrings
// (which collectively are taken to be the linear part of a parent
// geometry). The supplied rule determines the boundary semantics.
func NewLinearBoundary(lines []*geom.LineString, rule BoundaryNodeRule) *LinearBoundary {
	if rule == nil {
		rule = OGCSFSBoundaryRule
	}
	lb := &LinearBoundary{rule: rule}
	lb.vertexDegree = computeBoundaryPoints(lines)
	for _, deg := range lb.vertexDegree {
		if rule.IsInBoundary(deg) {
			lb.hasBoundary = true
			break
		}
	}
	return lb
}

// HasBoundary reports whether at least one vertex satisfies the rule.
func (b *LinearBoundary) HasBoundary() bool {
	if b == nil {
		return false
	}
	return b.hasBoundary
}

// IsBoundary reports whether p is a boundary point under the rule.
func (b *LinearBoundary) IsBoundary(p geom.XY) bool {
	if b == nil {
		return false
	}
	deg, ok := b.vertexDegree[p]
	if !ok {
		return false
	}
	return b.rule.IsInBoundary(deg)
}

// computeBoundaryPoints accumulates endpoint incidences across all
// supplied lines. Each non-empty LineString contributes one count to
// its first and last vertex (which may coincide for closed rings).
func computeBoundaryPoints(lines []*geom.LineString) map[geom.XY]int {
	deg := make(map[geom.XY]int)
	for _, line := range lines {
		if line == nil || line.IsEmpty() || line.NumPoints() < 1 {
			continue
		}
		first := line.PointAt(0)
		last := line.PointAt(line.NumPoints() - 1)
		deg[first]++
		// Mirror JTS: a closed line still increments both endpoints,
		// resulting in degree 2 (even, hence non-boundary under Mod2).
		// For the OGC default this gives the correct closed-ring →
		// empty boundary semantics.
		deg[last]++
	}
	return deg
}
