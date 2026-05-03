package planar

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSinSnap_AtAxes mirrors JTS Angle.sinSnap: snap |res|<5e-16 to 0.
// At multiples of π, math.Sin returns ~1.2e-16 noise — SinSnap should
// collapse it to exactly 0.
func TestSinSnap_AtAxes(t *testing.T) {
	cases := []struct {
		name string
		a    float64
		want float64
	}{
		{"sin(0)", 0, 0},
		{"sin(π/2)", math.Pi / 2, 1},
		{"sin(π)", math.Pi, 0}, // raw sin produces ~1.2e-16
		{"sin(3π/2)", 3 * math.Pi / 2, -1},
		{"sin(2π)", 2 * math.Pi, 0}, // raw sin produces ~-2.4e-16
		{"sin(-π)", -math.Pi, 0},
	}
	for _, tc := range cases {
		got := k.SinSnap(tc.a)
		assert.Equalf(t, tc.want, got, "%s = %v want %v", tc.name, got, tc.want)
	}
}

// TestCosSnap_AtAxes: cos(π/2) and cos(3π/2) produce ~6e-17 noise; CosSnap
// should collapse them to exactly 0.
func TestCosSnap_AtAxes(t *testing.T) {
	cases := []struct {
		name string
		a    float64
		want float64
	}{
		{"cos(0)", 0, 1},
		{"cos(π/2)", math.Pi / 2, 0}, // raw cos produces ~6e-17
		{"cos(π)", math.Pi, -1},
		{"cos(3π/2)", 3 * math.Pi / 2, 0}, // raw cos produces ~-1.8e-16
		{"cos(2π)", 2 * math.Pi, 1},
	}
	for _, tc := range cases {
		got := k.CosSnap(tc.a)
		assert.Equalf(t, tc.want, got, "%s = %v want %v", tc.name, got, tc.want)
	}
}

// Small perturbations larger than the snap threshold must NOT be snapped:
// the snap is only meant to clean up trig round-off, not user-supplied
// near-zero values.
func TestSnap_DoesNotSnapPerturbations(t *testing.T) {
	// 1e-10 is well above the 5e-16 threshold; must pass through.
	const eps = 1e-10
	got := k.SinSnap(eps)
	assert.InDelta(t, eps, got, 1e-20, "SinSnap should not snap 1e-10 to 0")
	got = k.CosSnap(math.Pi/2 + eps)
	// cos(π/2 + 1e-10) ≈ -1e-10 — should pass through unchanged (raw cos).
	assert.Equal(t, math.Cos(math.Pi/2+eps), got, "CosSnap should match raw cos for non-snap input")
}

// SinSnap/CosSnap must agree with raw sin/cos at all non-snap inputs.
func TestSnap_AgreesAwayFromAxes(t *testing.T) {
	for _, a := range []float64{0.1, 0.7, 1.3, 2.0, -0.5, 100.0} {
		assert.Equalf(t, math.Sin(a), k.SinSnap(a), "SinSnap(%v)", a)
		assert.Equalf(t, math.Cos(a), k.CosSnap(a), "CosSnap(%v)", a)
	}
}
