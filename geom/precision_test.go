package geom

import (
	"math"
	"testing"
)

func TestPrecisionModel_FloatingNoRounding(t *testing.T) {
	pm := NewFloatingPrecision()
	in := XY{1.234567890123, 9.876543210987}
	if got := pm.MakePrecise(in); got != in {
		t.Errorf("floating should not round: got %v, want %v", got, in)
	}
	if !pm.IsFloating() {
		t.Errorf("expected IsFloating true")
	}
	if !math.IsNaN(pm.GridSize()) {
		t.Errorf("expected NaN grid size for floating, got %v", pm.GridSize())
	}
}

func TestPrecisionModel_FixedScale(t *testing.T) {
	pm := NewFixedPrecision(1000) // 3 decimal places
	got := pm.MakePrecise(XY{1.23456, 9.87654})
	want := XY{1.235, 9.877}
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	if pm.Scale() != 1000 {
		t.Errorf("scale = %v", pm.Scale())
	}
}

func TestPrecisionModel_FixedGridSize(t *testing.T) {
	pm := NewFixedPrecision(-1000) // grid size 1000 -> rounds to nearest 1000
	got := pm.MakePrecise(XY{1234, 5678})
	want := XY{1000, 6000}
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	if pm.GridSize() != 1000 {
		t.Errorf("grid size = %v", pm.GridSize())
	}
}

func TestPrecisionModel_FloatingSingle(t *testing.T) {
	pm := NewFloatingSinglePrecision()
	got := pm.MakePreciseValue(1.0 / 3.0)
	if got == 1.0/3.0 {
		t.Errorf("floating-single should round through float32")
	}
}

func TestPrecisionModel_NaNInfPassthrough(t *testing.T) {
	pm := NewFixedPrecision(100)
	if !math.IsNaN(pm.MakePreciseValue(math.NaN())) {
		t.Errorf("NaN should pass through")
	}
	if !math.IsInf(pm.MakePreciseValue(math.Inf(1)), 1) {
		t.Errorf("Inf should pass through")
	}
}

func TestPrecisionModel_Compare(t *testing.T) {
	floating := NewFloatingPrecision()
	single := NewFloatingSinglePrecision()
	fixed3 := NewFixedPrecision(1000)
	if floating.Compare(single) <= 0 {
		t.Errorf("floating should be more precise than single")
	}
	if single.Compare(fixed3) <= 0 {
		// single = 6 digits, fixed3 = 1 + ceil(log10(1000)) = 4 digits
		t.Errorf("single should be more precise than fixed-3")
	}
}
