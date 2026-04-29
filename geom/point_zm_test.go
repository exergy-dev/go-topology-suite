package geom

import (
	"math"
	"testing"
)

func TestPointXYMConstructor(t *testing.T) {
	p := NewPointXYM(nil, XYM{X: 1, Y: 2, M: 3.14})
	if p.Layout() != LayoutXYM {
		t.Errorf("layout = %v", p.Layout())
	}
	if p.M() != 3.14 {
		t.Errorf("M = %v", p.M())
	}
	if !math.IsNaN(p.Z()) {
		t.Errorf("Z on XYM should be NaN, got %v", p.Z())
	}
}

func TestPointXYZMConstructor(t *testing.T) {
	p := NewPointXYZM(nil, XYZM{X: 1, Y: 2, Z: 3, M: 4})
	if p.Layout() != LayoutXYZM {
		t.Errorf("layout = %v", p.Layout())
	}
	if p.Z() != 3 || p.M() != 4 {
		t.Errorf("Z=%v M=%v", p.Z(), p.M())
	}
}

func TestPointXYZMZAndM(t *testing.T) {
	xy := NewPoint(nil, XY{X: 1, Y: 2})
	if !math.IsNaN(xy.Z()) || !math.IsNaN(xy.M()) {
		t.Errorf("XY point Z/M should be NaN")
	}
}
