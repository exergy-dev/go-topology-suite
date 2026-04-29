package wkt

import (
	"testing"

	"github.com/terra-geo/terra/geom"
)

func TestRoundTripPointXYM(t *testing.T) {
	in := "POINT M (1 2 99)"
	g, err := Unmarshal(in)
	if err != nil {
		t.Fatal(err)
	}
	if g.Layout() != geom.LayoutXYM {
		t.Errorf("layout = %v", g.Layout())
	}
	out, _ := Marshal(g)
	if out != in {
		t.Errorf("got %q, want %q", out, in)
	}
}

func TestRoundTripPointXYZM(t *testing.T) {
	in := "POINT ZM (1 2 3 4)"
	g, err := Unmarshal(in)
	if err != nil {
		t.Fatal(err)
	}
	if g.Layout() != geom.LayoutXYZM {
		t.Errorf("layout = %v", g.Layout())
	}
	pp := g.(*geom.Point)
	if pp.Z() != 3 || pp.M() != 4 {
		t.Errorf("Z=%v M=%v", pp.Z(), pp.M())
	}
	out, _ := Marshal(g)
	if out != in {
		t.Errorf("got %q, want %q", out, in)
	}
}
