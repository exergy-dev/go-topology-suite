package shape

import (
	"testing"

	"github.com/exergy-dev/go-topology-suite/geom"
)

func TestHilbertCurveCount(t *testing.T) {
	for order := 0; order <= 5; order++ {
		ls := HilbertCurve(order, geom.Envelope{MinX: 0, MinY: 0, MaxX: 1, MaxY: 1})
		want := 1 << (2 * order)
		if ls.NumPoints() != want {
			t.Fatalf("order=%d points=%d want=%d", order, ls.NumPoints(), want)
		}
	}
}

func TestHilbertCurveContainment(t *testing.T) {
	env := geom.Envelope{MinX: -2, MinY: 3, MaxX: 8, MaxY: 13}
	ls := HilbertCurve(4, env)
	got := ls.Envelope()
	// All vertices must be inside env (within FP slop).
	if got.MinX < env.MinX-1e-9 || got.MaxX > env.MaxX+1e-9 ||
		got.MinY < env.MinY-1e-9 || got.MaxY > env.MaxY+1e-9 {
		t.Fatalf("curve env %v escapes target %v", got, env)
	}
}

func TestHilbertCurveAdjacentDistance(t *testing.T) {
	// On the integer grid, consecutive Hilbert points must be distance 1 apart
	// (each step is a unit edge of the curve).
	level := 4
	for i := 1; i < hilbertSize(level); i++ {
		x0, y0 := hilbertDecode(level, i-1)
		x1, y1 := hilbertDecode(level, i)
		dx := x1 - x0
		dy := y1 - y0
		if dx*dx+dy*dy != 1 {
			t.Fatalf("step %d: (%d,%d)->(%d,%d) is not a unit move", i, x0, y0, x1, y1)
		}
	}
}

func TestHilbertCurveDeterministic(t *testing.T) {
	env := geom.Envelope{MinX: 0, MinY: 0, MaxX: 1, MaxY: 1}
	a := HilbertCurve(3, env)
	b := HilbertCurve(3, env)
	if a.NumPoints() != b.NumPoints() {
		t.Fatal("len mismatch")
	}
	for i := 0; i < a.NumPoints(); i++ {
		if a.PointAt(i) != b.PointAt(i) {
			t.Fatalf("differ at %d", i)
		}
	}
}

func TestHilbertCurveEmptyEnvelope(t *testing.T) {
	// With an empty envelope the curve falls back to integer-grid
	// coordinates and should still have the expected vertex count.
	ls := HilbertCurve(2, geom.EmptyEnvelope())
	if ls.NumPoints() != 16 {
		t.Fatalf("got %d points", ls.NumPoints())
	}
}
