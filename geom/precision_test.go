package geom

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrecisionModel_FloatingNoRounding(t *testing.T) {
	pm := NewFloatingPrecision()
	in := XY{1.234567890123, 9.876543210987}
	assert.Equal(t, in, pm.MakePrecise(in), "floating should not round")
	assert.True(t, pm.IsFloating(), "expected IsFloating true")
	assert.True(t, math.IsNaN(pm.GridSize()), "expected NaN grid size for floating, got %v", pm.GridSize())
}

func TestPrecisionModel_FixedScale(t *testing.T) {
	pm := NewFixedPrecision(1000) // 3 decimal places
	got := pm.MakePrecise(XY{1.23456, 9.87654})
	want := XY{1.235, 9.877}
	assert.Equal(t, want, got)
	assert.Equal(t, float64(1000), pm.Scale())
}

func TestPrecisionModel_FixedGridSize(t *testing.T) {
	pm := NewFixedPrecision(-1000) // grid size 1000 -> rounds to nearest 1000
	got := pm.MakePrecise(XY{1234, 5678})
	want := XY{1000, 6000}
	assert.Equal(t, want, got)
	assert.Equal(t, float64(1000), pm.GridSize())
}

func TestPrecisionModel_FloatingSingle(t *testing.T) {
	pm := NewFloatingSinglePrecision()
	got := pm.MakePreciseValue(1.0 / 3.0)
	assert.NotEqual(t, 1.0/3.0, got, "floating-single should round through float32")
}

func TestPrecisionModel_NaNInfPassthrough(t *testing.T) {
	pm := NewFixedPrecision(100)
	assert.True(t, math.IsNaN(pm.MakePreciseValue(math.NaN())), "NaN should pass through")
	assert.True(t, math.IsInf(pm.MakePreciseValue(math.Inf(1)), 1), "Inf should pass through")
}

func TestPrecisionModel_Compare(t *testing.T) {
	floating := NewFloatingPrecision()
	single := NewFloatingSinglePrecision()
	fixed3 := NewFixedPrecision(1000)
	assert.Greater(t, floating.Compare(single), 0, "floating should be more precise than single")
	// single = 6 digits, fixed3 = 1 + ceil(log10(1000)) = 4 digits
	assert.Greater(t, single.Compare(fixed3), 0, "single should be more precise than fixed-3")
}
