package geom

import "testing"

func TestEmptyEnvelopeIsEmpty(t *testing.T) {
	e := EmptyEnvelope()
	if !e.IsEmpty() {
		t.Fatalf("EmptyEnvelope().IsEmpty() = false, want true")
	}
	if e.Width() != 0 || e.Height() != 0 || e.Area() != 0 {
		t.Fatalf("empty envelope should have zero dimensions, got w=%v h=%v area=%v",
			e.Width(), e.Height(), e.Area())
	}
}

func TestEnvelopeExpandToIncludeXY(t *testing.T) {
	e := EmptyEnvelope()
	e = e.ExpandToIncludeXY(XY{1, 2})
	if e.IsEmpty() {
		t.Fatalf("envelope still empty after first expand")
	}
	if e.MinX != 1 || e.MinY != 2 || e.MaxX != 1 || e.MaxY != 2 {
		t.Fatalf("expand-to-first-point bounds wrong: %+v", e)
	}
	e = e.ExpandToIncludeXY(XY{-1, 5})
	if e.MinX != -1 || e.MinY != 2 || e.MaxX != 1 || e.MaxY != 5 {
		t.Fatalf("expand bounds wrong: %+v", e)
	}
}

func TestEnvelopeIntersects(t *testing.T) {
	a := Envelope{0, 0, 10, 10}
	b := Envelope{5, 5, 15, 15}
	c := Envelope{20, 20, 30, 30}

	if !a.Intersects(b) {
		t.Errorf("a should intersect b")
	}
	if a.Intersects(c) {
		t.Errorf("a should not intersect c")
	}
	if a.Intersects(EmptyEnvelope()) {
		t.Errorf("nothing should intersect empty")
	}
}

func TestEnvelopeContainsXY(t *testing.T) {
	e := Envelope{0, 0, 10, 10}
	cases := []struct {
		p    XY
		want bool
	}{
		{XY{5, 5}, true},
		{XY{0, 0}, true},   // boundary
		{XY{10, 10}, true}, // boundary
		{XY{-1, 5}, false},
		{XY{5, 11}, false},
	}
	for _, c := range cases {
		if got := e.ContainsXY(c.p); got != c.want {
			t.Errorf("Contains(%v) = %v, want %v", c.p, got, c.want)
		}
	}
}

func TestEnvelopeOfFlat(t *testing.T) {
	flat := []float64{1, 2, 3, 4, -1, 5}
	e := envelopeOfFlat(flat, 2)
	if e.MinX != -1 || e.MinY != 2 || e.MaxX != 3 || e.MaxY != 5 {
		t.Fatalf("envelopeOfFlat got %+v", e)
	}

	flat3 := []float64{1, 2, 99, 3, 4, 99, -1, 5, 99}
	e3 := envelopeOfFlat(flat3, 3)
	if e3.MinX != -1 || e3.MinY != 2 || e3.MaxX != 3 || e3.MaxY != 5 {
		t.Fatalf("envelopeOfFlat XYZ got %+v", e3)
	}
}
