package shape

import (
	"testing"

	"github.com/terra-geo/terra/geom"
)

func TestKochSnowflakeVertexCount(t *testing.T) {
	// At level n the boundary has 3 * 4^n segments + closing vertex.
	cases := []struct {
		level int
		want  int
	}{
		{0, 4},  // triangle
		{1, 13}, // 12 segments + closing vertex
		{2, 49},
		{3, 193},
	}
	for _, c := range cases {
		p := KochSnowflake(c.level, geom.XY{X: 0, Y: 0}, 10)
		ring := p.ExteriorRing()
		if len(ring) != c.want {
			t.Fatalf("level=%d vertices=%d want=%d", c.level, len(ring), c.want)
		}
	}
}

func TestKochSnowflakeClosed(t *testing.T) {
	p := KochSnowflake(2, geom.XY{X: 5, Y: 5}, 4)
	ring := p.ExteriorRing()
	if ring[0] != ring[len(ring)-1] {
		t.Fatalf("ring not closed: %v vs %v", ring[0], ring[len(ring)-1])
	}
}

func TestKochSnowflakeApproxCentred(t *testing.T) {
	c := geom.XY{X: 0, Y: 0}
	p := KochSnowflake(2, c, 6)
	env := p.Envelope()
	// The figure should contain c (within FP slop). The level>0 vertical
	// shift puts c roughly inside the snowflake.
	if !env.ContainsXY(c) {
		t.Fatalf("centre %v not contained in env %v", c, env)
	}
}

func TestKochSnowflakeDeterministic(t *testing.T) {
	c := geom.XY{X: 1, Y: 2}
	a := KochSnowflake(2, c, 5)
	b := KochSnowflake(2, c, 5)
	ra := a.ExteriorRing()
	rb := b.ExteriorRing()
	if len(ra) != len(rb) {
		t.Fatal("len mismatch")
	}
	for i := range ra {
		if ra[i] != rb[i] {
			t.Fatalf("differ at %d", i)
		}
	}
}

func TestKochSnowflakeZeroSize(t *testing.T) {
	p := KochSnowflake(2, geom.XY{X: 0, Y: 0}, 0)
	if !p.IsEmpty() {
		t.Fatal("expected empty polygon")
	}
}
