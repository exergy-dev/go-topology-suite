package relateng

import (
	"testing"

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
	if !got {
		t.Errorf("II hit: got false, want true")
	}
	// No interaction → false.
	p2 := NewIntersectsPredicate()
	got = driveCells(p2, [][3]int{
		{LocInterior, LocExterior, DimA},
		{LocExterior, LocInterior, DimA},
	})
	if got {
		t.Errorf("disjoint cells: got true, want false")
	}
}

func TestIntersectsPredicate_EnvelopeShortCircuit(t *testing.T) {
	p := NewIntersectsPredicate()
	envA := geom.Envelope{MinX: 0, MinY: 0, MaxX: 1, MaxY: 1}
	envB := geom.Envelope{MinX: 10, MinY: 10, MaxX: 11, MaxY: 11}
	p.InitEnv(envA, envB)
	if !p.IsKnown() {
		t.Fatal("envelope short-circuit: predicate should be known")
	}
	if p.Value() {
		t.Error("envelope short-circuit: value should be false")
	}
}

func TestDisjointPredicate(t *testing.T) {
	p := NewDisjointPredicate()
	got := driveCells(p, [][3]int{
		{LocInterior, LocInterior, DimP},
	})
	if got {
		t.Errorf("interaction → disjoint should be false")
	}
	p2 := NewDisjointPredicate()
	envA := geom.Envelope{MinX: 0, MinY: 0, MaxX: 1, MaxY: 1}
	envB := geom.Envelope{MinX: 10, MinY: 10, MaxX: 11, MaxY: 11}
	p2.InitEnv(envA, envB)
	if !p2.IsKnown() || !p2.Value() {
		t.Error("disjoint envelopes → predicate should resolve true")
	}
}

func TestIMPatternMatcher_Equals(t *testing.T) {
	// Pattern "T*F**FFF*" — topological equality.
	p := NewMatchesPredicate("T*F**FFF*")
	if p == nil {
		t.Fatal("matcher construction failed")
	}
	// Feed an exact-equality matrix:
	//   II=2, IB=F, IE=F, BI=F, BB=1, BE=F, EI=F, EB=F, EE=2
	cells := [][3]int{
		{LocInterior, LocInterior, DimA},
		{LocBoundary, LocBoundary, DimL},
		// E/E set in NewIMPredicate to 2.
	}
	got := driveCells(p, cells)
	if !got {
		t.Errorf("equality: got false, want true")
	}
}

func TestIMPatternMatcher_RejectsViaShortCircuit(t *testing.T) {
	// Pattern requires II=F (disjoint interiors). Feeding II>=0
	// should resolve to false at first such cell.
	p := NewMatchesPredicate("FF*FF****")
	if p == nil {
		t.Fatal("matcher construction failed")
	}
	p.UpdateDimension(LocInterior, LocInterior, DimP)
	if !p.IsKnown() {
		t.Fatal("expected short-circuit when II exceeds pattern bound")
	}
	if p.Value() {
		t.Error("II>=0 with pattern II=F should be false")
	}
}

func TestIntersectionMatrix_PatternParseAndMatch(t *testing.T) {
	pm := NewPatternMatrix("T*F**FFF*")
	if pm == nil {
		t.Fatal("pattern parse failed")
	}
	// Set II=2, BB=1, leave others at default (DimFalse).
	im := NewIntersectionMatrix()
	im.Set(LocInterior, LocInterior, DimA)
	im.Set(LocBoundary, LocBoundary, DimL)
	im.Set(LocExterior, LocExterior, DimA)
	if !im.Matches("T*F**FFF*") {
		t.Errorf("matrix should match equals pattern: %s", im)
	}
}

func TestIsDimsCompatibleWithCovers(t *testing.T) {
	// Same dim: ok.
	if !IsDimsCompatibleWithCovers(DimA, DimA) {
		t.Error("A covers A: should be compatible")
	}
	// Bigger covers smaller: ok.
	if !IsDimsCompatibleWithCovers(DimA, DimL) {
		t.Error("A covers L: should be compatible")
	}
	// Smaller covers bigger: not ok.
	if IsDimsCompatibleWithCovers(DimL, DimA) {
		t.Error("L cannot cover A")
	}
	// Special case: P covers L (zero-length line).
	if !IsDimsCompatibleWithCovers(DimP, DimL) {
		t.Error("P covers L should be compatible (zero-length line case)")
	}
}
