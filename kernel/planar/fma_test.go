package planar

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/exergy-dev/go-topology-suite/geom"
	"github.com/exergy-dev/go-topology-suite/kernel"
)

// These regressions pin results that differed between arm64 and amd64
// because the arm64 compiler fused x*y - z into one FMA instruction
// (see crossDiff). The inputs are package-level variables so the
// arithmetic is not constant-folded at compile time.

// fmaReversalA/B are the endpoints of an input segment that the
// GEOSBuffer.xml #605 line traverses forward and then exactly back.
var (
	fmaReversalA = geom.XY{X: 373718.87248, Y: 6137668.4963}
	fmaReversalB = geom.XY{X: 375295.49858, Y: 6138886.9262}
)

func TestCrossDiff_AntiparallelIsExactlyZero(t *testing.T) {
	rx, ry := fmaReversalB.X-fmaReversalA.X, fmaReversalB.Y-fmaReversalA.Y
	// Fused, rx*(-ry) - ry*(-rx) evaluates to the rounding error of
	// rx*ry (about -1.7e-11), not 0.
	assert.Equal(t, 0.0, crossDiff(rx, -ry, ry, -rx))
}

func TestSegmentIntersect_ExactReversalIsCollinearOverlap(t *testing.T) {
	// a→b followed by b→a: the buffer offset generator relies on this
	// being CollinearOverlap to emit the 180° round join; a spurious
	// PointIntersection dropped the join and lost ~0.7% of the buffer
	// area on arm64.
	res := k.SegmentIntersect(fmaReversalA, fmaReversalB, fmaReversalB, fmaReversalA)
	assert.Equal(t, kernel.CollinearOverlap, res.Kind)

	_, ok := k.SegmentIntersection(fmaReversalA, fmaReversalB, fmaReversalB, fmaReversalA)
	assert.False(t, ok, "antiparallel overlapping segments have no unique intersection point")
}

var (
	fmaOnSegP = geom.XY{X: 4, Y: 2}
	fmaOnSegA = geom.XY{X: 4, Y: 6}
	fmaOnSegB = geom.XY{X: 4, Y: 1}
)

func TestSegmentDistance_PointOnSegmentIsExactlyZero(t *testing.T) {
	// t = 20/25 = 0.8 is inexact; fused, 6 + 0.8*-5 gives
	// 1.9999999999999998 instead of 2 and a non-zero distance.
	assert.Equal(t, 0.0, k.SegmentDistance(fmaOnSegP, fmaOnSegA, fmaOnSegB))
	assert.Equal(t, 0.0, k.SegmentDistanceSq(fmaOnSegP, fmaOnSegA, fmaOnSegB))
	assert.True(t, onSegment(fmaOnSegP, fmaOnSegA, fmaOnSegB))
}
