package relateng

import "github.com/terra-geo/terra/geom"

// IMPatternMatcher is a TopologyPredicate that matches a DE-9IM
// pattern string. Port of
// org.locationtech.jts.operation.relateng.IMPatternMatcher.
//
// The matcher inspects the pattern up front to decide its
// short-circuit hooks (RequireInteraction). It uses the standard
// IMPredicate machinery for incremental short-circuiting once
// cells exceed the pattern bound.
type IMPatternMatcher struct {
	*IMPredicate
	pattern  string
	patternM *IntersectionMatrix
}

// NewIMPatternMatcher constructs a matcher for the given 9-char
// pattern. Returns nil when the pattern is malformed.
func NewIMPatternMatcher(pattern string) *IMPatternMatcher {
	pm := NewPatternMatrix(pattern)
	if pm == nil {
		return nil
	}
	m := &IMPatternMatcher{
		IMPredicate: NewIMPredicate(),
		pattern:     pattern,
		patternM:    pm,
	}
	m.BindOwner(m)
	return m
}

// Name returns "IMPattern" (mirroring JTS).
func (m *IMPatternMatcher) Name() string { return "IMPattern" }

// RequireInteraction reports whether the pattern requires non-empty
// interaction in any of II/IB/BI/BB.
func (m *IMPatternMatcher) RequireInteraction() bool {
	return patternRequiresInteraction(m.patternM)
}

// InitEnv may resolve the predicate to false when the pattern
// requires interaction but the envelopes are disjoint.
func (m *IMPatternMatcher) InitEnv(envA, envB geom.Envelope) {
	if patternRequiresInteraction(m.patternM) && !envA.Intersects(envB) {
		m.SetValue(false)
	}
}

// IsDetermined returns true once any computed cell has exceeded the
// allowed pattern bound (so the result must be false), or once
// every "T"-required cell has been observed (so the result must be
// true once all other cells settle).
func (m *IMPatternMatcher) IsDetermined() bool {
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			pe := m.patternM.Get(i, j)
			if pe == DimDontCare {
				continue
			}
			mv := m.IM.Get(i, j)
			if pe == DimTrue {
				// 'T' requires non-empty; not yet known until a
				// cell appears.
				if mv < 0 {
					return false
				}
				continue
			}
			// Concrete pattern entry (-1, 0, 1, 2). The matrix
			// monotonically increases, so if it has already exceeded
			// the pattern's allowed value the result is forced false.
			if mv > pe {
				return true
			}
		}
	}
	return false
}

// ValueIM matches the running matrix against the pattern.
func (m *IMPatternMatcher) ValueIM() bool {
	return m.IM.Matches(m.pattern)
}

// patternRequiresInteraction reports whether any of the four
// I/B-row × I/B-column cells expects a non-empty intersection.
func patternRequiresInteraction(pm *IntersectionMatrix) bool {
	cells := [...][2]int{
		{LocInterior, LocInterior},
		{LocInterior, LocBoundary},
		{LocBoundary, LocInterior},
		{LocBoundary, LocBoundary},
	}
	for _, c := range cells {
		v := pm.Get(c[0], c[1])
		if v == DimTrue || v >= DimP {
			return true
		}
	}
	return false
}
