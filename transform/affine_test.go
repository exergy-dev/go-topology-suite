package transform

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAffineIdentity(t *testing.T) {
	affine := NewAffineIdentity()

	assert.True(t, affine.IsIdentity(), "NewAffineIdentity() should return identity matrix")

	x, y := 123.456, 789.012
	xOut, yOut, err := affine.Forward(x, y)
	require.NoError(t, err, "Forward() error")

	assert.InDelta(t, x, xOut, epsilon)
	assert.InDelta(t, y, yOut, epsilon)
}

func TestAffineTranslation(t *testing.T) {
	affine := NewAffineTranslation(10, 20)

	tests := []struct {
		name             string
		x, y             float64
		expectX, expectY float64
	}{
		{"origin", 0, 0, 10, 20},
		{"positive", 5, 15, 15, 35},
		{"negative", -5, -15, 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xOut, yOut, err := affine.Forward(tt.x, tt.y)
			require.NoError(t, err, "Forward() error")

			assert.InDelta(t, tt.expectX, xOut, epsilon)
			assert.InDelta(t, tt.expectY, yOut, epsilon)

			// Test inverse
			xInv, yInv, err := affine.Inverse(xOut, yOut)
			require.NoError(t, err, "Inverse() error")

			assert.InDelta(t, tt.x, xInv, epsilon, "Round trip X failed")
			assert.InDelta(t, tt.y, yInv, epsilon, "Round trip Y failed")
		})
	}
}

func TestAffineScale(t *testing.T) {
	affine := NewAffineScale(2, 3)

	tests := []struct {
		name             string
		x, y             float64
		expectX, expectY float64
	}{
		{"origin", 0, 0, 0, 0},
		{"positive", 5, 10, 10, 30},
		{"negative", -5, -10, -10, -30},
		{"unit", 1, 1, 2, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xOut, yOut, err := affine.Forward(tt.x, tt.y)
			require.NoError(t, err, "Forward() error")

			assert.InDelta(t, tt.expectX, xOut, epsilon)
			assert.InDelta(t, tt.expectY, yOut, epsilon)

			// Test inverse
			xInv, yInv, err := affine.Inverse(xOut, yOut)
			require.NoError(t, err, "Inverse() error")

			assert.InDelta(t, tt.x, xInv, epsilon, "Round trip X failed")
			assert.InDelta(t, tt.y, yInv, epsilon, "Round trip Y failed")
		})
	}
}

func TestAffineScaleOrigin(t *testing.T) {
	// Scale by 2x about the point (10, 20)
	affine := NewAffineScaleOrigin(2, 2, 10, 20)

	// Point at the origin should stay at the origin
	x, y := 10.0, 20.0
	xOut, yOut, err := affine.Forward(x, y)
	require.NoError(t, err, "Forward() error")

	assert.InDelta(t, x, xOut, epsilon, "Origin point X should stay fixed")
	assert.InDelta(t, y, yOut, epsilon, "Origin point Y should stay fixed")

	// Point 5 units away should be 10 units away after scaling by 2
	x, y = 15.0, 20.0 // 5 units to the right of origin
	xOut, yOut, err = affine.Forward(x, y)
	require.NoError(t, err, "Forward() error")

	assert.InDelta(t, 20.0, xOut, epsilon, "Scaled point X")
	assert.InDelta(t, 20.0, yOut, epsilon, "Scaled point Y")
}

func TestAffineRotation(t *testing.T) {
	// Rotate 90 degrees counter-clockwise
	affine := NewAffineRotation(math.Pi / 2)

	tests := []struct {
		name             string
		x, y             float64
		expectX, expectY float64
	}{
		{"origin", 0, 0, 0, 0},
		{"x-axis", 1, 0, 0, 1},
		{"y-axis", 0, 1, -1, 0},
		{"diagonal", 1, 1, -1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xOut, yOut, err := affine.Forward(tt.x, tt.y)
			require.NoError(t, err, "Forward() error")

			assert.InDelta(t, tt.expectX, xOut, epsilon, "Forward X")
			assert.InDelta(t, tt.expectY, yOut, epsilon, "Forward Y")

			// Test inverse
			xInv, yInv, err := affine.Inverse(xOut, yOut)
			require.NoError(t, err, "Inverse() error")

			assert.InDelta(t, tt.x, xInv, epsilon, "Round trip X failed")
			assert.InDelta(t, tt.y, yInv, epsilon, "Round trip Y failed")
		})
	}
}

func TestAffineRotationOrigin(t *testing.T) {
	// Rotate 90 degrees counter-clockwise about (10, 10)
	affine := NewAffineRotationOrigin(math.Pi/2, 10, 10)

	// Center should stay fixed
	x, y := 10.0, 10.0
	xOut, yOut, err := affine.Forward(x, y)
	require.NoError(t, err, "Forward() error")

	assert.InDelta(t, x, xOut, epsilon, "Center X should stay fixed")
	assert.InDelta(t, y, yOut, epsilon, "Center Y should stay fixed")

	// Point to the right of center should rotate to above center
	x, y = 15.0, 10.0 // 5 units to the right
	xOut, yOut, err = affine.Forward(x, y)
	require.NoError(t, err, "Forward() error")

	assert.InDelta(t, 10.0, xOut, epsilon, "Rotated point X")
	assert.InDelta(t, 15.0, yOut, epsilon, "Rotated point Y")
}

func TestAffineShear(t *testing.T) {
	// Shear in x direction
	affine := NewAffineShear(0.5, 0)

	tests := []struct {
		name             string
		x, y             float64
		expectX, expectY float64
	}{
		{"origin", 0, 0, 0, 0},
		{"x-axis", 2, 0, 2, 0},   // No change along x-axis
		{"y-axis", 0, 2, 1, 2},   // x' = x + 0.5*y = 0 + 0.5*2 = 1
		{"diagonal", 2, 4, 4, 4}, // x' = 2 + 0.5*4 = 4
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xOut, yOut, err := affine.Forward(tt.x, tt.y)
			require.NoError(t, err, "Forward() error")

			assert.InDelta(t, tt.expectX, xOut, epsilon, "Forward X")
			assert.InDelta(t, tt.expectY, yOut, epsilon, "Forward Y")

			// Test inverse
			xInv, yInv, err := affine.Inverse(xOut, yOut)
			require.NoError(t, err, "Inverse() error")

			assert.InDelta(t, tt.x, xInv, epsilon, "Round trip X failed")
			assert.InDelta(t, tt.y, yInv, epsilon, "Round trip Y failed")
		})
	}
}

func TestAffineCompose(t *testing.T) {
	// Compose: translate by (10, 20), then scale by (2, 3)
	translate := NewAffineTranslation(10, 20)
	scale := NewAffineScale(2, 3)
	composed := translate.Compose(scale)

	x, y := 5.0, 8.0

	// Apply composed transformation
	xComp, yComp, err := composed.Forward(x, y)
	require.NoError(t, err, "Forward() error")

	// Apply manually: translate first, then scale
	xTrans, yTrans, _ := translate.Forward(x, y)
	xExpected, yExpected, _ := scale.Forward(xTrans, yTrans)

	assert.InDelta(t, xExpected, xComp, epsilon, "Composed Forward X")
	assert.InDelta(t, yExpected, yComp, epsilon, "Composed Forward Y")

	// Test that composed inverse works
	xInv, yInv, err := composed.Inverse(xComp, yComp)
	require.NoError(t, err, "Inverse() error")

	assert.InDelta(t, x, xInv, epsilon, "Composed round trip X failed")
	assert.InDelta(t, y, yInv, epsilon, "Composed round trip Y failed")
}

func TestAffineDeterminant(t *testing.T) {
	tests := []struct {
		name   string
		affine *Affine
		det    float64
	}{
		{"identity", NewAffineIdentity(), 1},
		{"scale 2x2", NewAffineScale(2, 2), 4},
		{"scale 2x3", NewAffineScale(2, 3), 6},
		{"rotation", NewAffineRotation(math.Pi / 4), 1},  // Rotation preserves area
		{"translation", NewAffineTranslation(10, 20), 1}, // Translation preserves area
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			det := tt.affine.Determinant()
			assert.InDelta(t, tt.det, det, epsilon, "Determinant")
		})
	}
}

func TestAffineSingular(t *testing.T) {
	// Create a singular matrix (determinant = 0)
	// Example: scale by 0 in one direction
	singular := NewAffineScale(2, 0)

	det := singular.Determinant()
	assert.InDelta(t, 0, det, epsilon, "Expected singular matrix determinant")

	// Inverse should fail for singular matrix
	_, _, err := singular.Inverse(10, 20)
	assert.Error(t, err, "Inverse() should return error for singular matrix")
}

func TestAffineClone(t *testing.T) {
	original := NewAffineTranslation(10, 20)
	clone := original.Clone()

	// Modify original
	original.C = 999

	// Clone should be unchanged
	assert.Equal(t, 10.0, clone.C, "Clone was modified")
}

func TestAffineComposeChain(t *testing.T) {
	// Create a transformation chain: translate, scale, rotate
	trans := NewAffineTranslation(5, 10)
	scale := NewAffineScale(2, 2)
	rotate := NewAffineRotation(math.Pi / 4) // 45 degrees

	// Compose: trans -> scale -> rotate
	composed := trans.Compose(scale).Compose(rotate)

	x, y := 3.0, 4.0

	// Apply composed
	xComp, yComp, err := composed.Forward(x, y)
	require.NoError(t, err, "Forward() error")

	// Apply manually
	x1, y1, _ := trans.Forward(x, y)
	x2, y2, _ := scale.Forward(x1, y1)
	x3, y3, _ := rotate.Forward(x2, y2)

	assert.InDelta(t, x3, xComp, epsilon, "Composed chain X")
	assert.InDelta(t, y3, yComp, epsilon, "Composed chain Y")

	// Test inverse
	xInv, yInv, err := composed.Inverse(xComp, yComp)
	require.NoError(t, err, "Inverse() error")

	assert.InDelta(t, x, xInv, epsilon, "Round trip X failed")
	assert.InDelta(t, y, yInv, epsilon, "Round trip Y failed")
}

func BenchmarkAffineTranslation(b *testing.B) {
	affine := NewAffineTranslation(10, 20)
	x, y := 123.456, 789.012

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = affine.Forward(x, y)
	}
}

func BenchmarkAffineScale(b *testing.B) {
	affine := NewAffineScale(2, 3)
	x, y := 123.456, 789.012

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = affine.Forward(x, y)
	}
}

func BenchmarkAffineRotation(b *testing.B) {
	affine := NewAffineRotation(math.Pi / 4)
	x, y := 123.456, 789.012

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = affine.Forward(x, y)
	}
}

func BenchmarkAffineCompose(b *testing.B) {
	trans := NewAffineTranslation(10, 20)
	scale := NewAffineScale(2, 3)
	rotate := NewAffineRotation(math.Pi / 4)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		trans.Compose(scale).Compose(rotate)
	}
}
