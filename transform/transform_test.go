package transform

import (
	"math"
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const epsilon = 1e-9

func TestIdentityTransform(t *testing.T) {
	identity := NewIdentity()

	tests := []struct {
		name string
		x, y float64
	}{
		{"origin", 0, 0},
		{"positive", 100, 200},
		{"negative", -50, -75},
		{"mixed", -123.45, 678.90},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test forward
			x, y, err := identity.Forward(tt.x, tt.y)
			require.NoError(t, err, "Forward() error")
			assert.Equal(t, tt.x, x)
			assert.Equal(t, tt.y, y)

			// Test inverse
			x, y, err = identity.Inverse(tt.x, tt.y)
			require.NoError(t, err, "Inverse() error")
			assert.Equal(t, tt.x, x)
			assert.Equal(t, tt.y, y)
		})
	}
}

func TestInverseTransform(t *testing.T) {
	// Create a translation transform
	translation := NewAffineTranslation(10, 20)
	inverse := NewInverse(translation)

	x, y := 5.0, 15.0

	// Forward on inverse should apply translation's inverse
	xFwd, yFwd, err := inverse.Forward(x, y)
	require.NoError(t, err, "Forward() error")

	// Should be equivalent to translation.Inverse
	xExpected, yExpected, _ := translation.Inverse(x, y)
	assert.InDelta(t, xExpected, xFwd, epsilon)
	assert.InDelta(t, yExpected, yFwd, epsilon)

	// Inverse on inverse should apply translation's forward
	xInv, yInv, err := inverse.Inverse(xFwd, yFwd)
	require.NoError(t, err, "Inverse() error")

	assert.InDelta(t, x, xInv, epsilon, "Round trip X failed")
	assert.InDelta(t, y, yInv, epsilon, "Round trip Y failed")
}

func TestCompositeTransform(t *testing.T) {
	// Create a composite: translate by (10, 20), then scale by (2, 3)
	translate := NewAffineTranslation(10, 20)
	scale := NewAffineScale(2, 3)
	composite := NewComposite(translate, scale)

	x, y := 5.0, 8.0

	// Apply composite forward
	xComp, yComp, err := composite.Forward(x, y)
	require.NoError(t, err, "Forward() error")

	// Apply manually: translate first, then scale
	xTrans, yTrans, _ := translate.Forward(x, y)
	xExpected, yExpected, _ := scale.Forward(xTrans, yTrans)

	assert.InDelta(t, xExpected, xComp, epsilon)
	assert.InDelta(t, yExpected, yComp, epsilon)

	// Test inverse (should apply in reverse order)
	xInv, yInv, err := composite.Inverse(xComp, yComp)
	require.NoError(t, err, "Inverse() error")

	assert.InDelta(t, x, xInv, epsilon, "Composite round trip X failed")
	assert.InDelta(t, y, yInv, epsilon, "Composite round trip Y failed")
}

func TestCompositeEmptyTransforms(t *testing.T) {
	composite := NewComposite()

	x, y := 5.0, 10.0
	xFwd, yFwd, err := composite.Forward(x, y)
	require.NoError(t, err, "Forward() error")

	assert.Equal(t, x, xFwd, "Empty composite should act as identity")
	assert.Equal(t, y, yFwd, "Empty composite should act as identity")
}

func TestTransformCoordinate(t *testing.T) {
	translation := NewAffineTranslation(10, 20)

	// Test 2D coordinate
	coord := geom.NewCoordinate(5, 15)
	result, err := TransformCoordinate(translation, coord)
	require.NoError(t, err, "TransformCoordinate() error")

	assert.InDelta(t, 15.0, result.X, epsilon)
	assert.InDelta(t, 35.0, result.Y, epsilon)

	// Test 3D coordinate (Z should be preserved)
	coord3D := geom.NewCoordinateZ(5, 15, 100)
	result3D, err := TransformCoordinate(translation, coord3D)
	require.NoError(t, err, "TransformCoordinate() error")

	assert.InDelta(t, 15.0, result3D.X, epsilon)
	assert.InDelta(t, 35.0, result3D.Y, epsilon)

	require.True(t, result3D.HasZ(), "Z coordinate should be preserved")
	assert.InDelta(t, 100.0, result3D.Z, epsilon, "Z coordinate value")

	// Test coordinate with M value
	m := 42.0
	coordM := geom.NewCoordinateM(5, 15, m)
	resultM, err := TransformCoordinate(translation, coordM)
	require.NoError(t, err, "TransformCoordinate() error")

	require.True(t, resultM.HasM(), "M value should be preserved")
	assert.InDelta(t, 42.0, resultM.M, epsilon, "M value")
}

func TestTransformCoordinates(t *testing.T) {
	scale := NewAffineScale(2, 3)

	coords := geom.NewCoordinateSequence(
		geom.NewCoordinate(1, 2),
		geom.NewCoordinate(3, 4),
		geom.NewCoordinate(5, 6),
	)

	result, err := TransformCoordinates(scale, coords)
	require.NoError(t, err, "TransformCoordinates() error")

	expected := []struct{ x, y float64 }{
		{2, 6},
		{6, 12},
		{10, 18},
	}

	require.Len(t, result, len(expected), "Result length mismatch")

	for i, exp := range expected {
		assert.InDelta(t, exp.x, result[i].X, epsilon, "Coordinate %d X", i)
		assert.InDelta(t, exp.y, result[i].Y, epsilon, "Coordinate %d Y", i)
	}
}

func TestTransformEmptyCoordinates(t *testing.T) {
	scale := NewAffineScale(2, 3)

	coords := geom.CoordinateSequence{}
	result, err := TransformCoordinates(scale, coords)
	require.NoError(t, err, "TransformCoordinates() error")

	assert.Empty(t, result, "Empty coordinate sequence should remain empty")
}

func BenchmarkIdentityTransform(b *testing.B) {
	identity := NewIdentity()
	x, y := 123.456, 789.012

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		identity.Forward(x, y)
	}
}

func BenchmarkAffineTransform(b *testing.B) {
	affine := NewAffineRotation(math.Pi / 4)
	x, y := 123.456, 789.012

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		affine.Forward(x, y)
	}
}

func BenchmarkCompositeTransform(b *testing.B) {
	composite := NewComposite(
		NewAffineTranslation(10, 20),
		NewAffineScale(2, 3),
		NewAffineRotation(math.Pi / 6),
	)
	x, y := 123.456, 789.012

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		composite.Forward(x, y)
	}
}

func BenchmarkTransformCoordinates(b *testing.B) {
	transform := NewAffineRotation(math.Pi / 4)

	// Create a coordinate sequence with 100 points
	coords := make(geom.CoordinateSequence, 100)
	for i := 0; i < 100; i++ {
		coords[i] = geom.NewCoordinate(float64(i), float64(i)*2)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TransformCoordinates(transform, coords)
	}
}
