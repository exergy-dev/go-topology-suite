package wkb

import (
	"testing"

	"github.com/terra-geo/terra/geom"
)

func TestRoundTripPointXYM(t *testing.T) {
	src := geom.NewPointXYM(nil, geom.XYM{X: 1, Y: 2, M: 99})
	data, err := Marshal(src)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Unmarshal(data)
	if err != nil {
		t.Fatal(err)
	}
	if got.Layout() != geom.LayoutXYM {
		t.Errorf("layout = %v", got.Layout())
	}
	pp := got.(*geom.Point)
	if pp.M() != 99 {
		t.Errorf("M lost: got %v", pp.M())
	}
}

func TestRoundTripPointXYZM(t *testing.T) {
	src := geom.NewPointXYZM(nil, geom.XYZM{X: 1, Y: 2, Z: 3, M: 4})
	data, err := Marshal(src)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Unmarshal(data)
	if err != nil {
		t.Fatal(err)
	}
	pp := got.(*geom.Point)
	if pp.Z() != 3 || pp.M() != 4 {
		t.Errorf("Z=%v M=%v", pp.Z(), pp.M())
	}
}
