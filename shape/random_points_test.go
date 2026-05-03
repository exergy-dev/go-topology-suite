package shape

import (
	"testing"

	"github.com/terra-geo/terra/geom"
)

func TestRandomPointsCount(t *testing.T) {
	env := geom.Envelope{MinX: 0, MinY: 0, MaxX: 10, MaxY: 10}
	pts := RandomPoints(50, env, WithSeed(1))
	if len(pts) != 50 {
		t.Fatalf("count = %d, want 50", len(pts))
	}
}

func TestRandomPointsEnvelopeContainment(t *testing.T) {
	env := geom.Envelope{MinX: -3, MinY: 5, MaxX: 7, MaxY: 12}
	pts := RandomPoints(100, env, WithSeed(42))
	for i, p := range pts {
		if !env.ContainsXY(p) {
			t.Fatalf("pt[%d]=%v not in env", i, p)
		}
	}
}

func TestRandomPointsDeterministic(t *testing.T) {
	env := geom.Envelope{MinX: 0, MinY: 0, MaxX: 1, MaxY: 1}
	a := RandomPoints(20, env, WithSeed(7))
	b := RandomPoints(20, env, WithSeed(7))
	if len(a) != len(b) {
		t.Fatalf("len mismatch")
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("seed not deterministic at i=%d: %v vs %v", i, a[i], b[i])
		}
	}
}

func TestRandomPointsInPolygon(t *testing.T) {
	// Square hole-less polygon.
	shell := []geom.XY{
		{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}, {X: 0, Y: 10}, {X: 0, Y: 0},
	}
	p := geom.NewPolygon(nil, shell)
	pts := RandomPointsInPolygon(40, p, WithSeed(99))
	if len(pts) != 40 {
		t.Fatalf("count = %d, want 40", len(pts))
	}
	for i, q := range pts {
		if !pointInRing(q, shell) {
			t.Fatalf("pt[%d]=%v not in shell", i, q)
		}
	}
}

func TestRandomPointsInPolygonWithHole(t *testing.T) {
	shell := []geom.XY{
		{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}, {X: 0, Y: 10}, {X: 0, Y: 0},
	}
	hole := []geom.XY{
		{X: 3, Y: 3}, {X: 7, Y: 3}, {X: 7, Y: 7}, {X: 3, Y: 7}, {X: 3, Y: 3},
	}
	p := geom.NewPolygon(nil, shell, hole)
	pts := RandomPointsInPolygon(30, p, WithSeed(123))
	if len(pts) != 30 {
		t.Fatalf("count=%d", len(pts))
	}
	for i, q := range pts {
		if !pointInRing(q, shell) {
			t.Fatalf("pt[%d]=%v not in shell", i, q)
		}
		if pointInRing(q, hole) {
			t.Fatalf("pt[%d]=%v inside hole", i, q)
		}
	}
}

func TestRandomPointsZeroN(t *testing.T) {
	if RandomPoints(0, geom.Envelope{MinX: 0, MaxX: 1, MinY: 0, MaxY: 1}) != nil {
		t.Fatal("expected nil")
	}
}
