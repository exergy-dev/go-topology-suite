package shape

import (
	"testing"

	"github.com/terra-geo/terra/geom"
)

func TestMortonCurveCount(t *testing.T) {
	for order := 0; order <= 5; order++ {
		ls := MortonCurve(order, geom.Envelope{MinX: 0, MinY: 0, MaxX: 1, MaxY: 1})
		want := 1 << (2 * order)
		if ls.NumPoints() != want {
			t.Fatalf("order=%d points=%d want=%d", order, ls.NumPoints(), want)
		}
	}
}

func TestMortonCurveContainment(t *testing.T) {
	env := geom.Envelope{MinX: -2, MinY: 3, MaxX: 8, MaxY: 13}
	ls := MortonCurve(4, env)
	got := ls.Envelope()
	if got.MinX < env.MinX-1e-9 || got.MaxX > env.MaxX+1e-9 ||
		got.MinY < env.MinY-1e-9 || got.MaxY > env.MaxY+1e-9 {
		t.Fatalf("curve env %v escapes target %v", got, env)
	}
}

func TestMortonEncodeDecodeRoundtrip(t *testing.T) {
	// For all integer points in a 16x16 grid, encode then decode
	// must round-trip.
	for x := 0; x < 16; x++ {
		for y := 0; y < 16; y++ {
			idx := MortonEncode(x, y)
			dx, dy := MortonDecode(idx)
			if dx != x || dy != y {
				t.Fatalf("(%d,%d) -> %d -> (%d,%d)", x, y, idx, dx, dy)
			}
		}
	}
}

func TestMortonLevel(t *testing.T) {
	// MortonLevel returns the smallest level whose curve fits at
	// least n points. Curves at level k have 2^(2k) points: 1, 4, 16, 64, ...
	cases := []struct{ n, want int }{
		{1, 0},
		{2, 1}, {4, 1},
		{5, 2}, {16, 2},
		{17, 3}, {64, 3},
	}
	for _, c := range cases {
		if got := MortonLevel(c.n); got != c.want {
			t.Fatalf("MortonLevel(%d)=%d, want %d", c.n, got, c.want)
		}
	}
}

func TestMortonCurveEmptyEnvelope(t *testing.T) {
	ls := MortonCurve(2, geom.EmptyEnvelope())
	if ls.NumPoints() != 16 {
		t.Fatalf("got %d points", ls.NumPoints())
	}
	// First point at integer origin (0,0).
	if p := ls.PointAt(0); p.X != 0 || p.Y != 0 {
		t.Fatalf("first point: %v, want (0,0)", p)
	}
}

func TestMortonCurveOrderZero(t *testing.T) {
	ls := MortonCurve(0, geom.Envelope{MinX: 0, MinY: 0, MaxX: 1, MaxY: 1})
	if ls.NumPoints() != 1 {
		t.Fatalf("order=0: want 1 point, got %d", ls.NumPoints())
	}
}
