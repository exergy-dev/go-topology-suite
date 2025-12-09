package transform

import (
	"math"
	"testing"
)

func TestAffineIdentity(t *testing.T) {
	affine := NewAffineIdentity()

	if !affine.IsIdentity() {
		t.Error("NewAffineIdentity() should return identity matrix")
	}

	x, y := 123.456, 789.012
	xOut, yOut, err := affine.Forward(x, y)
	if err != nil {
		t.Fatalf("Forward() error = %v", err)
	}

	if math.Abs(xOut-x) > epsilon || math.Abs(yOut-y) > epsilon {
		t.Errorf("Identity.Forward(%f, %f) = (%f, %f), want (%f, %f)",
			x, y, xOut, yOut, x, y)
	}
}

func TestAffineTranslation(t *testing.T) {
	affine := NewAffineTranslation(10, 20)

	tests := []struct {
		name       string
		x, y       float64
		expectX, expectY float64
	}{
		{"origin", 0, 0, 10, 20},
		{"positive", 5, 15, 15, 35},
		{"negative", -5, -15, 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xOut, yOut, err := affine.Forward(tt.x, tt.y)
			if err != nil {
				t.Fatalf("Forward() error = %v", err)
			}

			if math.Abs(xOut-tt.expectX) > epsilon || math.Abs(yOut-tt.expectY) > epsilon {
				t.Errorf("Forward(%f, %f) = (%f, %f), want (%f, %f)",
					tt.x, tt.y, xOut, yOut, tt.expectX, tt.expectY)
			}

			// Test inverse
			xInv, yInv, err := affine.Inverse(xOut, yOut)
			if err != nil {
				t.Fatalf("Inverse() error = %v", err)
			}

			if math.Abs(xInv-tt.x) > epsilon || math.Abs(yInv-tt.y) > epsilon {
				t.Errorf("Round trip failed: got (%f, %f), want (%f, %f)",
					xInv, yInv, tt.x, tt.y)
			}
		})
	}
}

func TestAffineScale(t *testing.T) {
	affine := NewAffineScale(2, 3)

	tests := []struct {
		name       string
		x, y       float64
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
			if err != nil {
				t.Fatalf("Forward() error = %v", err)
			}

			if math.Abs(xOut-tt.expectX) > epsilon || math.Abs(yOut-tt.expectY) > epsilon {
				t.Errorf("Forward(%f, %f) = (%f, %f), want (%f, %f)",
					tt.x, tt.y, xOut, yOut, tt.expectX, tt.expectY)
			}

			// Test inverse
			xInv, yInv, err := affine.Inverse(xOut, yOut)
			if err != nil {
				t.Fatalf("Inverse() error = %v", err)
			}

			if math.Abs(xInv-tt.x) > epsilon || math.Abs(yInv-tt.y) > epsilon {
				t.Errorf("Round trip failed: got (%f, %f), want (%f, %f)",
					xInv, yInv, tt.x, tt.y)
			}
		})
	}
}

func TestAffineScaleOrigin(t *testing.T) {
	// Scale by 2x about the point (10, 20)
	affine := NewAffineScaleOrigin(2, 2, 10, 20)

	// Point at the origin should stay at the origin
	x, y := 10.0, 20.0
	xOut, yOut, err := affine.Forward(x, y)
	if err != nil {
		t.Fatalf("Forward() error = %v", err)
	}

	if math.Abs(xOut-x) > epsilon || math.Abs(yOut-y) > epsilon {
		t.Errorf("Origin point should stay fixed: got (%f, %f), want (%f, %f)",
			xOut, yOut, x, y)
	}

	// Point 5 units away should be 10 units away after scaling by 2
	x, y = 15.0, 20.0 // 5 units to the right of origin
	xOut, yOut, err = affine.Forward(x, y)
	if err != nil {
		t.Fatalf("Forward() error = %v", err)
	}

	expectedX, expectedY := 20.0, 20.0 // 10 units to the right of origin
	if math.Abs(xOut-expectedX) > epsilon || math.Abs(yOut-expectedY) > epsilon {
		t.Errorf("Scaled point: got (%f, %f), want (%f, %f)",
			xOut, yOut, expectedX, expectedY)
	}
}

func TestAffineRotation(t *testing.T) {
	// Rotate 90 degrees counter-clockwise
	affine := NewAffineRotation(math.Pi / 2)

	tests := []struct {
		name       string
		x, y       float64
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
			if err != nil {
				t.Fatalf("Forward() error = %v", err)
			}

			if math.Abs(xOut-tt.expectX) > epsilon || math.Abs(yOut-tt.expectY) > epsilon {
				t.Errorf("Forward(%f, %f) = (%f, %f), want (%f, %f)",
					tt.x, tt.y, xOut, yOut, tt.expectX, tt.expectY)
			}

			// Test inverse
			xInv, yInv, err := affine.Inverse(xOut, yOut)
			if err != nil {
				t.Fatalf("Inverse() error = %v", err)
			}

			if math.Abs(xInv-tt.x) > epsilon || math.Abs(yInv-tt.y) > epsilon {
				t.Errorf("Round trip failed: got (%f, %f), want (%f, %f)",
					xInv, yInv, tt.x, tt.y)
			}
		})
	}
}

func TestAffineRotationOrigin(t *testing.T) {
	// Rotate 90 degrees counter-clockwise about (10, 10)
	affine := NewAffineRotationOrigin(math.Pi/2, 10, 10)

	// Center should stay fixed
	x, y := 10.0, 10.0
	xOut, yOut, err := affine.Forward(x, y)
	if err != nil {
		t.Fatalf("Forward() error = %v", err)
	}

	if math.Abs(xOut-x) > epsilon || math.Abs(yOut-y) > epsilon {
		t.Errorf("Center should stay fixed: got (%f, %f), want (%f, %f)",
			xOut, yOut, x, y)
	}

	// Point to the right of center should rotate to above center
	x, y = 15.0, 10.0 // 5 units to the right
	xOut, yOut, err = affine.Forward(x, y)
	if err != nil {
		t.Fatalf("Forward() error = %v", err)
	}

	expectedX, expectedY := 10.0, 15.0 // 5 units above
	if math.Abs(xOut-expectedX) > epsilon || math.Abs(yOut-expectedY) > epsilon {
		t.Errorf("Rotated point: got (%f, %f), want (%f, %f)",
			xOut, yOut, expectedX, expectedY)
	}
}

func TestAffineShear(t *testing.T) {
	// Shear in x direction
	affine := NewAffineShear(0.5, 0)

	tests := []struct {
		name       string
		x, y       float64
		expectX, expectY float64
	}{
		{"origin", 0, 0, 0, 0},
		{"x-axis", 2, 0, 2, 0}, // No change along x-axis
		{"y-axis", 0, 2, 1, 2}, // x' = x + 0.5*y = 0 + 0.5*2 = 1
		{"diagonal", 2, 4, 4, 4}, // x' = 2 + 0.5*4 = 4
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xOut, yOut, err := affine.Forward(tt.x, tt.y)
			if err != nil {
				t.Fatalf("Forward() error = %v", err)
			}

			if math.Abs(xOut-tt.expectX) > epsilon || math.Abs(yOut-tt.expectY) > epsilon {
				t.Errorf("Forward(%f, %f) = (%f, %f), want (%f, %f)",
					tt.x, tt.y, xOut, yOut, tt.expectX, tt.expectY)
			}

			// Test inverse
			xInv, yInv, err := affine.Inverse(xOut, yOut)
			if err != nil {
				t.Fatalf("Inverse() error = %v", err)
			}

			if math.Abs(xInv-tt.x) > epsilon || math.Abs(yInv-tt.y) > epsilon {
				t.Errorf("Round trip failed: got (%f, %f), want (%f, %f)",
					xInv, yInv, tt.x, tt.y)
			}
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
	if err != nil {
		t.Fatalf("Forward() error = %v", err)
	}

	// Apply manually: translate first, then scale
	xTrans, yTrans, _ := translate.Forward(x, y)
	xExpected, yExpected, _ := scale.Forward(xTrans, yTrans)

	if math.Abs(xComp-xExpected) > epsilon || math.Abs(yComp-yExpected) > epsilon {
		t.Errorf("Composed.Forward() = (%f, %f), want (%f, %f)",
			xComp, yComp, xExpected, yExpected)
	}

	// Test that composed inverse works
	xInv, yInv, err := composed.Inverse(xComp, yComp)
	if err != nil {
		t.Fatalf("Inverse() error = %v", err)
	}

	if math.Abs(xInv-x) > epsilon || math.Abs(yInv-y) > epsilon {
		t.Errorf("Composed round trip failed: got (%f, %f), want (%f, %f)",
			xInv, yInv, x, y)
	}
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
		{"rotation", NewAffineRotation(math.Pi / 4), 1}, // Rotation preserves area
		{"translation", NewAffineTranslation(10, 20), 1}, // Translation preserves area
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			det := tt.affine.Determinant()
			if math.Abs(det-tt.det) > epsilon {
				t.Errorf("Determinant() = %f, want %f", det, tt.det)
			}
		})
	}
}

func TestAffineSingular(t *testing.T) {
	// Create a singular matrix (determinant = 0)
	// Example: scale by 0 in one direction
	singular := NewAffineScale(2, 0)

	det := singular.Determinant()
	if math.Abs(det) > epsilon {
		t.Errorf("Expected singular matrix, got determinant %f", det)
	}

	// Inverse should fail for singular matrix
	_, _, err := singular.Inverse(10, 20)
	if err == nil {
		t.Error("Inverse() should return error for singular matrix")
	}
}

func TestAffineClone(t *testing.T) {
	original := NewAffineTranslation(10, 20)
	clone := original.Clone()

	// Modify original
	original.C = 999

	// Clone should be unchanged
	if clone.C != 10 {
		t.Errorf("Clone was modified: C = %f, want 10", clone.C)
	}
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
	if err != nil {
		t.Fatalf("Forward() error = %v", err)
	}

	// Apply manually
	x1, y1, _ := trans.Forward(x, y)
	x2, y2, _ := scale.Forward(x1, y1)
	x3, y3, _ := rotate.Forward(x2, y2)

	if math.Abs(xComp-x3) > epsilon || math.Abs(yComp-y3) > epsilon {
		t.Errorf("Composed chain: got (%f, %f), want (%f, %f)",
			xComp, yComp, x3, y3)
	}

	// Test inverse
	xInv, yInv, err := composed.Inverse(xComp, yComp)
	if err != nil {
		t.Fatalf("Inverse() error = %v", err)
	}

	if math.Abs(xInv-x) > epsilon || math.Abs(yInv-y) > epsilon {
		t.Errorf("Round trip failed: got (%f, %f), want (%f, %f)",
			xInv, yInv, x, y)
	}
}

func BenchmarkAffineTranslation(b *testing.B) {
	affine := NewAffineTranslation(10, 20)
	x, y := 123.456, 789.012

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		affine.Forward(x, y)
	}
}

func BenchmarkAffineScale(b *testing.B) {
	affine := NewAffineScale(2, 3)
	x, y := 123.456, 789.012

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		affine.Forward(x, y)
	}
}

func BenchmarkAffineRotation(b *testing.B) {
	affine := NewAffineRotation(math.Pi / 4)
	x, y := 123.456, 789.012

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		affine.Forward(x, y)
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
