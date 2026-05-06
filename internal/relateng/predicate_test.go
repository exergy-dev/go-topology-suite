package relateng

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/exergy-dev/go-topology-suite/geom"
)

// driveCells feeds a sequence of (locA, locB, dim) cells into a
// predicate, calling Finish at the end. Mirrors what the future
// TopologyComputer will do.
func driveCells(p TopologyPredicate, cells [][3]int) bool {
	for _, c := range cells {
		if p.IsKnown() {
			break
		}
		p.UpdateDimension(c[0], c[1], c[2])
	}
	if !p.IsKnown() {
		p.Finish()
	}
	return p.Value()
}

func TestIntersectsPredicate(t *testing.T) {
	// Interaction in II → true.
	p := NewIntersectsPredicate()
	got := driveCells(p, [][3]int{
		{LocInterior, LocInterior, DimP},
	})
	assert.True(t, got, "II hit should be true")
	// No interaction → false.
	p2 := NewIntersectsPredicate()
	got = driveCells(p2, [][3]int{
		{LocInterior, LocExterior, DimA},
		{LocExterior, LocInterior, DimA},
	})
	assert.False(t, got, "disjoint cells should be false")
}

func TestIntersectsPredicate_EnvelopeShortCircuit(t *testing.T) {
	p := NewIntersectsPredicate()
	envA := geom.Envelope{MinX: 0, MinY: 0, MaxX: 1, MaxY: 1}
	envB := geom.Envelope{MinX: 10, MinY: 10, MaxX: 11, MaxY: 11}
	p.InitEnv(envA, envB)
	require.True(t, p.IsKnown(), "envelope short-circuit: predicate should be known")
	assert.False(t, p.Value(), "envelope short-circuit: value should be false")
}

func TestDisjointPredicate(t *testing.T) {
	p := NewDisjointPredicate()
	got := driveCells(p, [][3]int{
		{LocInterior, LocInterior, DimP},
	})
	assert.False(t, got, "interaction → disjoint should be false")
	p2 := NewDisjointPredicate()
	envA := geom.Envelope{MinX: 0, MinY: 0, MaxX: 1, MaxY: 1}
	envB := geom.Envelope{MinX: 10, MinY: 10, MaxX: 11, MaxY: 11}
	p2.InitEnv(envA, envB)
	assert.True(t, p2.IsKnown(), "disjoint envelopes → predicate should be known")
	assert.True(t, p2.Value(), "disjoint envelopes → value should be true")
}

func TestIMPatternMatcher_Equals(t *testing.T) {
	// Pattern "T*F**FFF*" — topological equality.
	p := NewMatchesPredicate("T*F**FFF*")
	require.NotNil(t, p, "matcher construction failed")
	// Feed an exact-equality matrix:
	//   II=2, IB=F, IE=F, BI=F, BB=1, BE=F, EI=F, EB=F, EE=2
	cells := [][3]int{
		{LocInterior, LocInterior, DimA},
		{LocBoundary, LocBoundary, DimL},
		// E/E set in NewIMPredicate to 2.
	}
	got := driveCells(p, cells)
	assert.True(t, got, "equality should be true")
}

func TestIMPatternMatcher_RejectsViaShortCircuit(t *testing.T) {
	// Pattern requires II=F (disjoint interiors). Feeding II>=0
	// should resolve to false at first such cell.
	p := NewMatchesPredicate("FF*FF****")
	require.NotNil(t, p, "matcher construction failed")
	p.UpdateDimension(LocInterior, LocInterior, DimP)
	require.True(t, p.IsKnown(), "expected short-circuit when II exceeds pattern bound")
	assert.False(t, p.Value(), "II>=0 with pattern II=F should be false")
}

func TestIntersectionMatrix_PatternParseAndMatch(t *testing.T) {
	pm := NewPatternMatrix("T*F**FFF*")
	require.NotNil(t, pm, "pattern parse failed")
	// Set II=2, BB=1, leave others at default (DimFalse).
	im := NewIntersectionMatrix()
	im.Set(LocInterior, LocInterior, DimA)
	im.Set(LocBoundary, LocBoundary, DimL)
	im.Set(LocExterior, LocExterior, DimA)
	assert.True(t, im.Matches("T*F**FFF*"),
		"matrix should match equals pattern: %s", im)
}

func TestIsDimsCompatibleWithCovers(t *testing.T) {
	// Same dim: ok.
	assert.True(t, IsDimsCompatibleWithCovers(DimA, DimA), "A covers A")
	// Bigger covers smaller: ok.
	assert.True(t, IsDimsCompatibleWithCovers(DimA, DimL), "A covers L")
	// Smaller covers bigger: not ok.
	assert.False(t, IsDimsCompatibleWithCovers(DimL, DimA), "L cannot cover A")
	// Special case: P covers L (zero-length line).
	assert.True(t, IsDimsCompatibleWithCovers(DimP, DimL),
		"P covers L should be compatible (zero-length line case)")
}
