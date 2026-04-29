package geom

import (
	"math"
	"testing"
)

func TestXYEqual(t *testing.T) {
	a := XY{1, 2}
	b := XY{1, 2}
	c := XY{1, 3}
	if !a.Equal(b) {
		t.Errorf("a.Equal(b) = false")
	}
	if a.Equal(c) {
		t.Errorf("a.Equal(c) = true")
	}
}

func TestXYEqualOrBothNaN(t *testing.T) {
	nan := math.NaN()
	a := XY{nan, 2}
	b := XY{nan, 2}
	if !a.EqualOrBothNaN(b) {
		t.Errorf("EqualOrBothNaN should treat matching NaN as equal")
	}
	if a.Equal(b) {
		t.Errorf("plain Equal should not treat NaN==NaN")
	}
}

func TestLayoutStride(t *testing.T) {
	if LayoutXY.Stride() != 2 {
		t.Errorf("XY stride = %d, want 2", LayoutXY.Stride())
	}
	if LayoutXYZ.Stride() != 3 || LayoutXYM.Stride() != 3 {
		t.Errorf("XYZ/XYM stride should be 3")
	}
	if LayoutXYZM.Stride() != 4 {
		t.Errorf("XYZM stride = %d, want 4", LayoutXYZM.Stride())
	}
	if NoLayout.Stride() != 0 {
		t.Errorf("NoLayout stride should be 0")
	}
}

func TestLayoutHasZHasM(t *testing.T) {
	if LayoutXY.HasZ() || LayoutXY.HasM() {
		t.Errorf("XY has neither Z nor M")
	}
	if !LayoutXYZ.HasZ() || LayoutXYZ.HasM() {
		t.Errorf("XYZ.HasZ should be true, HasM false")
	}
	if LayoutXYM.HasZ() || !LayoutXYM.HasM() {
		t.Errorf("XYM.HasM should be true")
	}
	if !LayoutXYZM.HasZ() || !LayoutXYZM.HasM() {
		t.Errorf("XYZM should have both Z and M")
	}
}
