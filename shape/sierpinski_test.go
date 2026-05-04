package shape

import (
	"testing"

	"github.com/exergy-dev/go-topology-suite/geom"
)

func TestSierpinskiCarpetHoleCount(t *testing.T) {
	// Total holes for level n is sum_{k=0..n} 8^k.
	cases := []struct {
		level int
		want  int
	}{
		{0, 1},
		{1, 9},
		{2, 73},
		{3, 585},
	}
	env := geom.Envelope{MinX: 0, MinY: 0, MaxX: 9, MaxY: 9}
	for _, c := range cases {
		mp := SierpinskiCarpet(c.level, env)
		if mp.NumGeometries() != 1 {
			t.Fatalf("level=%d: expected 1 polygon, got %d", c.level, mp.NumGeometries())
		}
		p := mp.PolygonAt(0)
		holes := p.NumRings() - 1 // subtract shell
		if holes != c.want {
			t.Fatalf("level=%d holes=%d want=%d", c.level, holes, c.want)
		}
	}
}

func TestSierpinskiCarpetEnvelope(t *testing.T) {
	env := geom.Envelope{MinX: 0, MinY: 0, MaxX: 9, MaxY: 9}
	mp := SierpinskiCarpet(2, env)
	got := mp.Envelope()
	if got.MinX < env.MinX-1e-9 || got.MaxX > env.MaxX+1e-9 ||
		got.MinY < env.MinY-1e-9 || got.MaxY > env.MaxY+1e-9 {
		t.Fatalf("carpet env %v escapes target %v", got, env)
	}
}

func TestSierpinskiCarpetEmpty(t *testing.T) {
	mp := SierpinskiCarpet(2, geom.EmptyEnvelope())
	if mp.NumGeometries() != 0 {
		t.Fatalf("expected empty MultiPolygon")
	}
}

func TestSierpinskiCarpetDeterministic(t *testing.T) {
	env := geom.Envelope{MinX: 0, MinY: 0, MaxX: 9, MaxY: 9}
	a := SierpinskiCarpet(2, env)
	b := SierpinskiCarpet(2, env)
	pa := a.PolygonAt(0)
	pb := b.PolygonAt(0)
	if pa.NumRings() != pb.NumRings() {
		t.Fatal("ring count mismatch")
	}
	for i := 0; i < pa.NumRings(); i++ {
		ra := pa.Ring(i)
		rb := pb.Ring(i)
		if len(ra) != len(rb) {
			t.Fatalf("ring %d len mismatch", i)
		}
		for j := range ra {
			if ra[j] != rb[j] {
				t.Fatalf("ring %d vertex %d differs", i, j)
			}
		}
	}
}
