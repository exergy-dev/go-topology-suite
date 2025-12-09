package transform

import (
	"math"
	"testing"

	"github.com/go-topology-suite/gts/geom"
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
			if err != nil {
				t.Errorf("Forward() error = %v", err)
			}
			if x != tt.x || y != tt.y {
				t.Errorf("Forward(%f, %f) = (%f, %f), want (%f, %f)",
					tt.x, tt.y, x, y, tt.x, tt.y)
			}

			// Test inverse
			x, y, err = identity.Inverse(tt.x, tt.y)
			if err != nil {
				t.Errorf("Inverse() error = %v", err)
			}
			if x != tt.x || y != tt.y {
				t.Errorf("Inverse(%f, %f) = (%f, %f), want (%f, %f)",
					tt.x, tt.y, x, y, tt.x, tt.y)
			}
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
	if err != nil {
		t.Fatalf("Forward() error = %v", err)
	}

	// Should be equivalent to translation.Inverse
	xExpected, yExpected, _ := translation.Inverse(x, y)
	if math.Abs(xFwd-xExpected) > epsilon || math.Abs(yFwd-yExpected) > epsilon {
		t.Errorf("InverseTransform.Forward() = (%f, %f), want (%f, %f)",
			xFwd, yFwd, xExpected, yExpected)
	}

	// Inverse on inverse should apply translation's forward
	xInv, yInv, err := inverse.Inverse(xFwd, yFwd)
	if err != nil {
		t.Fatalf("Inverse() error = %v", err)
	}

	if math.Abs(xInv-x) > epsilon || math.Abs(yInv-y) > epsilon {
		t.Errorf("Round trip failed: got (%f, %f), want (%f, %f)",
			xInv, yInv, x, y)
	}
}

func TestCompositeTransform(t *testing.T) {
	// Create a composite: translate by (10, 20), then scale by (2, 3)
	translate := NewAffineTranslation(10, 20)
	scale := NewAffineScale(2, 3)
	composite := NewComposite(translate, scale)

	x, y := 5.0, 8.0

	// Apply composite forward
	xComp, yComp, err := composite.Forward(x, y)
	if err != nil {
		t.Fatalf("Forward() error = %v", err)
	}

	// Apply manually: translate first, then scale
	xTrans, yTrans, _ := translate.Forward(x, y)
	xExpected, yExpected, _ := scale.Forward(xTrans, yTrans)

	if math.Abs(xComp-xExpected) > epsilon || math.Abs(yComp-yExpected) > epsilon {
		t.Errorf("Composite.Forward() = (%f, %f), want (%f, %f)",
			xComp, yComp, xExpected, yExpected)
	}

	// Test inverse (should apply in reverse order)
	xInv, yInv, err := composite.Inverse(xComp, yComp)
	if err != nil {
		t.Fatalf("Inverse() error = %v", err)
	}

	if math.Abs(xInv-x) > epsilon || math.Abs(yInv-y) > epsilon {
		t.Errorf("Composite round trip failed: got (%f, %f), want (%f, %f)",
			xInv, yInv, x, y)
	}
}

func TestCompositeEmptyTransforms(t *testing.T) {
	composite := NewComposite()

	x, y := 5.0, 10.0
	xFwd, yFwd, err := composite.Forward(x, y)
	if err != nil {
		t.Fatalf("Forward() error = %v", err)
	}

	if xFwd != x || yFwd != y {
		t.Errorf("Empty composite should act as identity: got (%f, %f), want (%f, %f)",
			xFwd, yFwd, x, y)
	}
}

func TestTransformCoordinate(t *testing.T) {
	translation := NewAffineTranslation(10, 20)

	// Test 2D coordinate
	coord := geom.NewCoordinate(5, 15)
	result, err := TransformCoordinate(translation, coord)
	if err != nil {
		t.Fatalf("TransformCoordinate() error = %v", err)
	}

	if math.Abs(result.X-15) > epsilon || math.Abs(result.Y-35) > epsilon {
		t.Errorf("TransformCoordinate() = (%f, %f), want (15, 35)",
			result.X, result.Y)
	}

	// Test 3D coordinate (Z should be preserved)
	coord3D := geom.NewCoordinateZ(5, 15, 100)
	result3D, err := TransformCoordinate(translation, coord3D)
	if err != nil {
		t.Fatalf("TransformCoordinate() error = %v", err)
	}

	if math.Abs(result3D.X-15) > epsilon || math.Abs(result3D.Y-35) > epsilon {
		t.Errorf("TransformCoordinate() = (%f, %f), want (15, 35)",
			result3D.X, result3D.Y)
	}

	if result3D.Z == nil || math.Abs(*result3D.Z-100) > epsilon {
		t.Errorf("Z coordinate not preserved: got %v, want 100", result3D.Z)
	}

	// Test coordinate with M value
	m := 42.0
	coordM := geom.NewCoordinateM(5, 15, m)
	resultM, err := TransformCoordinate(translation, coordM)
	if err != nil {
		t.Fatalf("TransformCoordinate() error = %v", err)
	}

	if resultM.M == nil || math.Abs(*resultM.M-42) > epsilon {
		t.Errorf("M value not preserved: got %v, want 42", resultM.M)
	}
}

func TestTransformCoordinates(t *testing.T) {
	scale := NewAffineScale(2, 3)

	coords := geom.NewCoordinateSequence(
		geom.NewCoordinate(1, 2),
		geom.NewCoordinate(3, 4),
		geom.NewCoordinate(5, 6),
	)

	result, err := TransformCoordinates(scale, coords)
	if err != nil {
		t.Fatalf("TransformCoordinates() error = %v", err)
	}

	expected := []struct{ x, y float64 }{
		{2, 6},
		{6, 12},
		{10, 18},
	}

	if len(result) != len(expected) {
		t.Fatalf("Result length = %d, want %d", len(result), len(expected))
	}

	for i, exp := range expected {
		if math.Abs(result[i].X-exp.x) > epsilon || math.Abs(result[i].Y-exp.y) > epsilon {
			t.Errorf("Coordinate %d = (%f, %f), want (%f, %f)",
				i, result[i].X, result[i].Y, exp.x, exp.y)
		}
	}
}

func TestTransformEmptyCoordinates(t *testing.T) {
	scale := NewAffineScale(2, 3)

	coords := geom.CoordinateSequence{}
	result, err := TransformCoordinates(scale, coords)
	if err != nil {
		t.Fatalf("TransformCoordinates() error = %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Empty coordinate sequence should remain empty, got length %d", len(result))
	}
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
