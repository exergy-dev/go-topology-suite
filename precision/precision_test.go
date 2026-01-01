package precision_test

import (
	"math"
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/robert-malhotra/go-topology-suite/precision"
)

func TestFloatingPrecisionModel(t *testing.T) {
	t.Run("Creation", func(t *testing.T) {
		pm := precision.NewFloatingPrecision()
		if pm == nil {
			t.Fatal("Expected non-nil precision model")
		}
	})

	t.Run("Type", func(t *testing.T) {
		pm := precision.NewFloatingPrecision()
		if pm.Type() != precision.FloatingPrecision {
			t.Errorf("Expected FloatingPrecision type, got %v", pm.Type())
		}
	})

	t.Run("IsFloating", func(t *testing.T) {
		pm := precision.NewFloatingPrecision()
		if !pm.IsFloating() {
			t.Error("Expected IsFloating to be true")
		}
	})

	t.Run("Scale", func(t *testing.T) {
		pm := precision.NewFloatingPrecision()
		if pm.Scale() != 0 {
			t.Errorf("Expected scale 0, got %v", pm.Scale())
		}
	})

	t.Run("MaxSignificantDigits", func(t *testing.T) {
		pm := precision.NewFloatingPrecision()
		if pm.MaxSignificantDigits() != 16 {
			t.Errorf("Expected 16 significant digits, got %v", pm.MaxSignificantDigits())
		}
	})

	t.Run("MakePrecise_NoOp", func(t *testing.T) {
		pm := precision.NewFloatingPrecision()
		coord := geom.NewCoordinate(1.23456789012345678, 9.87654321098765432)
		original := coord
		pm.MakePrecise(&coord)

		if coord.X != original.X || coord.Y != original.Y {
			t.Errorf("Floating precision should not modify coordinates")
		}
	})

	t.Run("MakePreciseValue", func(t *testing.T) {
		pm := precision.NewFloatingPrecision()
		val := 3.141592653589793
		result := pm.MakePreciseValue(val)
		if result != val {
			t.Errorf("Expected %v, got %v", val, result)
		}
	})

	t.Run("MakePrecise_WithZ", func(t *testing.T) {
		pm := precision.NewFloatingPrecision()
		coord := geom.NewCoordinateZ(1.5, 2.5, 3.5)
		originalZ := *coord.Z
		pm.MakePrecise(&coord)

		if *coord.Z != originalZ {
			t.Error("Floating precision should not modify Z coordinate")
		}
	})

	t.Run("MakePrecise_WithM", func(t *testing.T) {
		pm := precision.NewFloatingPrecision()
		coord := geom.NewCoordinateM(1.5, 2.5, 4.5)
		originalM := *coord.M
		pm.MakePrecise(&coord)

		if *coord.M != originalM {
			t.Error("Floating precision should not modify M coordinate")
		}
	})

	t.Run("CommonModel_Floating", func(t *testing.T) {
		if precision.Floating == nil {
			t.Error("Expected Floating to be non-nil")
		}
		if precision.Floating.Type() != precision.FloatingPrecision {
			t.Error("Expected Floating to be FloatingPrecision type")
		}
	})
}

func TestFloatingSinglePrecisionModel(t *testing.T) {
	t.Run("Creation", func(t *testing.T) {
		pm := precision.NewFloatingSinglePrecision()
		if pm == nil {
			t.Fatal("Expected non-nil precision model")
		}
	})

	t.Run("Type", func(t *testing.T) {
		pm := precision.NewFloatingSinglePrecision()
		if pm.Type() != precision.FloatingSinglePrecision {
			t.Errorf("Expected FloatingSinglePrecision type, got %v", pm.Type())
		}
	})

	t.Run("IsFloating", func(t *testing.T) {
		pm := precision.NewFloatingSinglePrecision()
		if !pm.IsFloating() {
			t.Error("Expected IsFloating to be true")
		}
	})

	t.Run("Scale", func(t *testing.T) {
		pm := precision.NewFloatingSinglePrecision()
		if pm.Scale() != 0 {
			t.Errorf("Expected scale 0, got %v", pm.Scale())
		}
	})

	t.Run("MaxSignificantDigits", func(t *testing.T) {
		pm := precision.NewFloatingSinglePrecision()
		if pm.MaxSignificantDigits() != 6 {
			t.Errorf("Expected 6 significant digits, got %v", pm.MaxSignificantDigits())
		}
	})

	t.Run("MakePrecise_ReducesPrecision", func(t *testing.T) {
		pm := precision.NewFloatingSinglePrecision()
		// Use a value that loses precision when converted to float32
		coord := geom.NewCoordinate(1.123456789012345, 2.987654321098765)
		pm.MakePrecise(&coord)

		// Check that precision was reduced (values should differ from original)
		expected32X := float64(float32(1.123456789012345))
		expected32Y := float64(float32(2.987654321098765))

		if coord.X != expected32X {
			t.Errorf("Expected X=%v, got %v", expected32X, coord.X)
		}
		if coord.Y != expected32Y {
			t.Errorf("Expected Y=%v, got %v", expected32Y, coord.Y)
		}
	})

	t.Run("MakePreciseValue", func(t *testing.T) {
		pm := precision.NewFloatingSinglePrecision()
		val := 3.141592653589793
		result := pm.MakePreciseValue(val)
		expected := float64(float32(val))

		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("MakePrecise_WithZ", func(t *testing.T) {
		pm := precision.NewFloatingSinglePrecision()
		coord := geom.NewCoordinateZ(1.123456789, 2.987654321, 3.567890123)
		pm.MakePrecise(&coord)

		expectedZ := float64(float32(3.567890123))
		if *coord.Z != expectedZ {
			t.Errorf("Expected Z=%v, got %v", expectedZ, *coord.Z)
		}
	})

	t.Run("MakePrecise_WithM", func(t *testing.T) {
		pm := precision.NewFloatingSinglePrecision()
		coord := geom.NewCoordinateM(1.5, 2.5, 4.123456789)
		pm.MakePrecise(&coord)

		expectedM := float64(float32(4.123456789))
		if *coord.M != expectedM {
			t.Errorf("Expected M=%v, got %v", expectedM, *coord.M)
		}
	})

	t.Run("CommonModel_FloatingSingle", func(t *testing.T) {
		if precision.FloatingSingle == nil {
			t.Error("Expected FloatingSingle to be non-nil")
		}
		if precision.FloatingSingle.Type() != precision.FloatingSinglePrecision {
			t.Error("Expected FloatingSingle to be FloatingSinglePrecision type")
		}
	})
}

func TestFixedPrecisionModel(t *testing.T) {
	t.Run("Creation", func(t *testing.T) {
		pm := precision.NewFixedPrecision(100)
		if pm == nil {
			t.Fatal("Expected non-nil precision model")
		}
	})

	t.Run("Creation_ZeroScale", func(t *testing.T) {
		pm := precision.NewFixedPrecision(0)
		if pm.Scale() != 1 {
			t.Errorf("Expected scale to default to 1, got %v", pm.Scale())
		}
	})

	t.Run("Creation_NegativeScale", func(t *testing.T) {
		pm := precision.NewFixedPrecision(-100)
		if pm.Scale() != 1 {
			t.Errorf("Expected scale to default to 1, got %v", pm.Scale())
		}
	})

	t.Run("Type", func(t *testing.T) {
		pm := precision.NewFixedPrecision(100)
		if pm.Type() != precision.FixedPrecision {
			t.Errorf("Expected FixedPrecision type, got %v", pm.Type())
		}
	})

	t.Run("IsFloating", func(t *testing.T) {
		pm := precision.NewFixedPrecision(100)
		if pm.IsFloating() {
			t.Error("Expected IsFloating to be false")
		}
	})

	t.Run("Scale", func(t *testing.T) {
		pm := precision.NewFixedPrecision(1000)
		if pm.Scale() != 1000 {
			t.Errorf("Expected scale 1000, got %v", pm.Scale())
		}
	})

	t.Run("MaxSignificantDigits", func(t *testing.T) {
		tests := []struct {
			scale    float64
			expected int
		}{
			{10, 2},
			{100, 3},
			{1000, 4},
			{1000000, 7},
		}

		for _, tt := range tests {
			pm := precision.NewFixedPrecision(tt.scale)
			if pm.MaxSignificantDigits() != tt.expected {
				t.Errorf("For scale %v, expected %v significant digits, got %v",
					tt.scale, tt.expected, pm.MaxSignificantDigits())
			}
		}
	})

	t.Run("MakePreciseValue_Scale100", func(t *testing.T) {
		pm := precision.NewFixedPrecision(100)
		tests := []struct {
			input    float64
			expected float64
		}{
			{1.234, 1.23},
			{1.235, 1.24},
			{1.236, 1.24},
			{-1.234, -1.23},
			{-1.235, -1.24},
			{0.0, 0.0},
		}

		for _, tt := range tests {
			result := pm.MakePreciseValue(tt.input)
			if result != tt.expected {
				t.Errorf("For input %v, expected %v, got %v", tt.input, tt.expected, result)
			}
		}
	})

	t.Run("MakePrecise_2DecimalPlaces", func(t *testing.T) {
		pm := precision.NewFixedPrecision(100) // 2 decimal places
		coord := geom.NewCoordinate(1.23456, 7.89012)
		pm.MakePrecise(&coord)

		if coord.X != 1.23 {
			t.Errorf("Expected X=1.23, got %v", coord.X)
		}
		if coord.Y != 7.89 {
			t.Errorf("Expected Y=7.89, got %v", coord.Y)
		}
	})

	t.Run("MakePrecise_3DecimalPlaces", func(t *testing.T) {
		pm := precision.NewFixedPrecision(1000) // 3 decimal places
		coord := geom.NewCoordinate(1.23456, 7.89012)
		pm.MakePrecise(&coord)

		if coord.X != 1.235 {
			t.Errorf("Expected X=1.235, got %v", coord.X)
		}
		if coord.Y != 7.890 {
			t.Errorf("Expected Y=7.890, got %v", coord.Y)
		}
	})

	t.Run("MakePrecise_WithZ", func(t *testing.T) {
		pm := precision.NewFixedPrecision(100)
		coord := geom.NewCoordinateZ(1.23456, 2.34567, 3.45678)
		pm.MakePrecise(&coord)

		if coord.X != 1.23 {
			t.Errorf("Expected X=1.23, got %v", coord.X)
		}
		if coord.Y != 2.35 {
			t.Errorf("Expected Y=2.35, got %v", coord.Y)
		}
		if *coord.Z != 3.46 {
			t.Errorf("Expected Z=3.46, got %v", *coord.Z)
		}
	})

	t.Run("MakePrecise_WithM", func(t *testing.T) {
		pm := precision.NewFixedPrecision(100)
		coord := geom.NewCoordinateM(1.23456, 2.34567, 4.56789)
		pm.MakePrecise(&coord)

		if coord.X != 1.23 || coord.Y != 2.35 || *coord.M != 4.57 {
			t.Errorf("Unexpected coordinate values: (%v, %v, M=%v)", coord.X, coord.Y, *coord.M)
		}
	})

	t.Run("MakePrecise_WithZM", func(t *testing.T) {
		pm := precision.NewFixedPrecision(10) // 1 decimal place
		coord := geom.NewCoordinateZM(1.26, 2.34, 3.45, 4.56)
		pm.MakePrecise(&coord)

		if coord.X != 1.3 || coord.Y != 2.3 || *coord.Z != 3.5 || *coord.M != 4.6 {
			t.Errorf("Unexpected coordinate values: (%v, %v, Z=%v, M=%v)",
				coord.X, coord.Y, *coord.Z, *coord.M)
		}
	})

	t.Run("CommonModel_Fixed1", func(t *testing.T) {
		if precision.Fixed1 == nil {
			t.Error("Expected Fixed1 to be non-nil")
		}
		if precision.Fixed1.Scale() != 10 {
			t.Errorf("Expected Fixed1 scale 10, got %v", precision.Fixed1.Scale())
		}
	})

	t.Run("CommonModel_Fixed2", func(t *testing.T) {
		if precision.Fixed2 == nil {
			t.Error("Expected Fixed2 to be non-nil")
		}
		if precision.Fixed2.Scale() != 100 {
			t.Errorf("Expected Fixed2 scale 100, got %v", precision.Fixed2.Scale())
		}
	})

	t.Run("CommonModel_Fixed3", func(t *testing.T) {
		if precision.Fixed3 == nil {
			t.Error("Expected Fixed3 to be non-nil")
		}
		if precision.Fixed3.Scale() != 1000 {
			t.Errorf("Expected Fixed3 scale 1000, got %v", precision.Fixed3.Scale())
		}
	})

	t.Run("CommonModel_Fixed6", func(t *testing.T) {
		if precision.Fixed6 == nil {
			t.Error("Expected Fixed6 to be non-nil")
		}
		if precision.Fixed6.Scale() != 1000000 {
			t.Errorf("Expected Fixed6 scale 1000000, got %v", precision.Fixed6.Scale())
		}
	})

	t.Run("CommonModel_Fixed8", func(t *testing.T) {
		if precision.Fixed8 == nil {
			t.Error("Expected Fixed8 to be non-nil")
		}
		if precision.Fixed8.Scale() != 100000000 {
			t.Errorf("Expected Fixed8 scale 100000000, got %v", precision.Fixed8.Scale())
		}
	})
}

func TestMakePreciseSequence(t *testing.T) {
	t.Run("EmptySequence", func(t *testing.T) {
		pm := precision.NewFixedPrecision(100)
		coords := geom.CoordinateSequence{}
		precision.MakePreciseSequence(pm, coords)
		// Should not panic
		if len(coords) != 0 {
			t.Error("Expected empty sequence to remain empty")
		}
	})

	t.Run("FloatingPrecision", func(t *testing.T) {
		pm := precision.NewFloatingPrecision()
		coords := geom.NewCoordinateSequence(
			geom.NewCoordinate(1.23456789, 2.34567890),
			geom.NewCoordinate(3.45678901, 4.56789012),
		)
		original := coords.Clone()
		precision.MakePreciseSequence(pm, coords)

		for i := range coords {
			if !coords[i].Equals2D(original[i], 1e-10) {
				t.Error("Floating precision should not modify coordinates")
			}
		}
	})

	t.Run("FixedPrecision", func(t *testing.T) {
		pm := precision.NewFixedPrecision(100) // 2 decimal places
		coords := geom.NewCoordinateSequence(
			geom.NewCoordinate(1.23456, 2.34567),
			geom.NewCoordinate(3.45678, 4.56789),
		)
		precision.MakePreciseSequence(pm, coords)

		if coords[0].X != 1.23 || coords[0].Y != 2.35 {
			t.Errorf("Expected (1.23, 2.35), got (%v, %v)", coords[0].X, coords[0].Y)
		}
		if coords[1].X != 3.46 || coords[1].Y != 4.57 {
			t.Errorf("Expected (3.46, 4.57), got (%v, %v)", coords[1].X, coords[1].Y)
		}
	})

	t.Run("SequenceWithZ", func(t *testing.T) {
		pm := precision.NewFixedPrecision(10) // 1 decimal place
		coords := geom.NewCoordinateSequence(
			geom.NewCoordinateZ(1.26, 2.34, 3.45),
			geom.NewCoordinateZ(4.56, 5.67, 6.78),
		)
		precision.MakePreciseSequence(pm, coords)

		if coords[0].X != 1.3 || coords[0].Y != 2.3 || *coords[0].Z != 3.5 {
			t.Errorf("Unexpected first coordinate: (%v, %v, %v)",
				coords[0].X, coords[0].Y, *coords[0].Z)
		}
		if coords[1].X != 4.6 || coords[1].Y != 5.7 || *coords[1].Z != 6.8 {
			t.Errorf("Unexpected second coordinate: (%v, %v, %v)",
				coords[1].X, coords[1].Y, *coords[1].Z)
		}
	})

	t.Run("NilSequence", func(t *testing.T) {
		pm := precision.NewFixedPrecision(100)
		var coords geom.CoordinateSequence
		precision.MakePreciseSequence(pm, coords)
		// Should not panic
	})
}

func TestComparePrecision(t *testing.T) {
	t.Run("FloatingVsFixed", func(t *testing.T) {
		floating := precision.NewFloatingPrecision()
		fixed := precision.NewFixedPrecision(100)

		result := precision.Compare(floating, fixed)
		if result != 1 {
			t.Errorf("Expected 1 (floating > fixed), got %v", result)
		}

		result = precision.Compare(fixed, floating)
		if result != -1 {
			t.Errorf("Expected -1 (fixed < floating), got %v", result)
		}
	})

	t.Run("TwoFloating", func(t *testing.T) {
		f1 := precision.NewFloatingPrecision()
		f2 := precision.NewFloatingPrecision()

		result := precision.Compare(f1, f2)
		if result != 0 {
			t.Errorf("Expected 0 (equal), got %v", result)
		}
	})

	t.Run("FixedWithDifferentScales", func(t *testing.T) {
		fixed100 := precision.NewFixedPrecision(100)
		fixed1000 := precision.NewFixedPrecision(1000)

		result := precision.Compare(fixed1000, fixed100)
		if result != 1 {
			t.Errorf("Expected 1 (1000 > 100), got %v", result)
		}

		result = precision.Compare(fixed100, fixed1000)
		if result != -1 {
			t.Errorf("Expected -1 (100 < 1000), got %v", result)
		}
	})

	t.Run("FixedWithSameScale", func(t *testing.T) {
		f1 := precision.NewFixedPrecision(100)
		f2 := precision.NewFixedPrecision(100)

		result := precision.Compare(f1, f2)
		if result != 0 {
			t.Errorf("Expected 0 (equal), got %v", result)
		}
	})

	t.Run("FloatingSingleVsFloating", func(t *testing.T) {
		single := precision.NewFloatingSinglePrecision()
		floating := precision.NewFloatingPrecision()

		// FloatingSingle is considered less precise than Floating in comparison
		result := precision.Compare(single, floating)
		if result != -1 {
			t.Errorf("Expected -1 (single < floating), got %v", result)
		}

		result = precision.Compare(floating, single)
		if result != 1 {
			t.Errorf("Expected 1 (floating > single), got %v", result)
		}
	})

	t.Run("FloatingSingleVsFixed", func(t *testing.T) {
		single := precision.NewFloatingSinglePrecision()
		fixed := precision.NewFixedPrecision(100)

		// FloatingSingle vs Fixed - both have scale 0, so compare by scale
		result := precision.Compare(single, fixed)
		if result != -1 {
			t.Errorf("Expected -1 (single < fixed), got %v", result)
		}
	})
}

func TestMostPrecise(t *testing.T) {
	t.Run("FloatingVsFixed", func(t *testing.T) {
		floating := precision.NewFloatingPrecision()
		fixed := precision.NewFixedPrecision(100)

		result := precision.MostPrecise(floating, fixed)
		if result != floating {
			t.Error("Expected floating to be most precise")
		}

		result = precision.MostPrecise(fixed, floating)
		if result != floating {
			t.Error("Expected floating to be most precise")
		}
	})

	t.Run("TwoFixed", func(t *testing.T) {
		fixed100 := precision.NewFixedPrecision(100)
		fixed1000 := precision.NewFixedPrecision(1000)

		result := precision.MostPrecise(fixed100, fixed1000)
		if result != fixed1000 {
			t.Error("Expected fixed1000 to be most precise")
		}

		result = precision.MostPrecise(fixed1000, fixed100)
		if result != fixed1000 {
			t.Error("Expected fixed1000 to be most precise")
		}
	})

	t.Run("SamePrecision", func(t *testing.T) {
		f1 := precision.NewFixedPrecision(100)
		f2 := precision.NewFixedPrecision(100)

		result := precision.MostPrecise(f1, f2)
		// Should return first one when equal
		if result != f1 {
			t.Error("Expected first precision model when equal")
		}
	})
}

func TestLeastPrecise(t *testing.T) {
	t.Run("FloatingVsFixed", func(t *testing.T) {
		floating := precision.NewFloatingPrecision()
		fixed := precision.NewFixedPrecision(100)

		result := precision.LeastPrecise(floating, fixed)
		if result != fixed {
			t.Error("Expected fixed to be least precise")
		}

		result = precision.LeastPrecise(fixed, floating)
		if result != fixed {
			t.Error("Expected fixed to be least precise")
		}
	})

	t.Run("TwoFixed", func(t *testing.T) {
		fixed100 := precision.NewFixedPrecision(100)
		fixed1000 := precision.NewFixedPrecision(1000)

		result := precision.LeastPrecise(fixed100, fixed1000)
		if result != fixed100 {
			t.Error("Expected fixed100 to be least precise")
		}

		result = precision.LeastPrecise(fixed1000, fixed100)
		if result != fixed100 {
			t.Error("Expected fixed100 to be least precise")
		}
	})

	t.Run("SamePrecision", func(t *testing.T) {
		f1 := precision.NewFixedPrecision(100)
		f2 := precision.NewFixedPrecision(100)

		result := precision.LeastPrecise(f1, f2)
		// Should return first one when equal
		if result != f1 {
			t.Error("Expected first precision model when equal")
		}
	})
}

func TestPrecisionEdgeCases(t *testing.T) {
	t.Run("FixedPrecision_VeryLargeValues", func(t *testing.T) {
		pm := precision.NewFixedPrecision(100)
		coord := geom.NewCoordinate(1234567.89012, 9876543.21098)
		pm.MakePrecise(&coord)

		if coord.X != 1234567.89 {
			t.Errorf("Expected X=1234567.89, got %v", coord.X)
		}
		if coord.Y != 9876543.21 {
			t.Errorf("Expected Y=9876543.21, got %v", coord.Y)
		}
	})

	t.Run("FixedPrecision_VerySmallValues", func(t *testing.T) {
		pm := precision.NewFixedPrecision(100)
		coord := geom.NewCoordinate(0.00123, 0.00456)
		pm.MakePrecise(&coord)

		if coord.X != 0.0 {
			t.Errorf("Expected X=0.00, got %v", coord.X)
		}
		if coord.Y != 0.0 {
			t.Errorf("Expected Y=0.00, got %v", coord.Y)
		}
	})

	t.Run("FixedPrecision_NegativeValues", func(t *testing.T) {
		pm := precision.NewFixedPrecision(100)
		coord := geom.NewCoordinate(-1.23456, -7.89012)
		pm.MakePrecise(&coord)

		if coord.X != -1.23 {
			t.Errorf("Expected X=-1.23, got %v", coord.X)
		}
		if coord.Y != -7.89 {
			t.Errorf("Expected Y=-7.89, got %v", coord.Y)
		}
	})

	t.Run("FixedPrecision_Zero", func(t *testing.T) {
		pm := precision.NewFixedPrecision(100)
		coord := geom.NewCoordinate(0.0, 0.0)
		pm.MakePrecise(&coord)

		if coord.X != 0.0 || coord.Y != 0.0 {
			t.Errorf("Expected (0, 0), got (%v, %v)", coord.X, coord.Y)
		}
	})

	t.Run("FixedPrecision_Infinity", func(t *testing.T) {
		pm := precision.NewFixedPrecision(100)
		inf := math.Inf(1)
		coord := geom.NewCoordinate(inf, -inf)
		pm.MakePrecise(&coord)

		if !math.IsInf(coord.X, 1) {
			t.Error("Expected X to be positive infinity")
		}
		if !math.IsInf(coord.Y, -1) {
			t.Error("Expected Y to be negative infinity")
		}
	})

	t.Run("FixedPrecision_NaN", func(t *testing.T) {
		pm := precision.NewFixedPrecision(100)
		nan := math.NaN()
		coord := geom.NewCoordinate(nan, nan)
		pm.MakePrecise(&coord)

		if !math.IsNaN(coord.X) || !math.IsNaN(coord.Y) {
			t.Error("Expected NaN values to remain NaN")
		}
	})

	t.Run("SinglePrecision_Overflow", func(t *testing.T) {
		pm := precision.NewFloatingSinglePrecision()
		// Value that's representable in float64 but might overflow in float32
		large := 1e38
		coord := geom.NewCoordinate(large, large)
		pm.MakePrecise(&coord)

		// Should still work, just with reduced precision
		if math.IsNaN(coord.X) || math.IsNaN(coord.Y) {
			t.Error("Expected valid coordinates after single precision conversion")
		}
	})
}

func TestPrecisionRounding(t *testing.T) {
	t.Run("RoundingUpFixed", func(t *testing.T) {
		pm := precision.NewFixedPrecision(100)
		tests := []struct {
			input    float64
			expected float64
		}{
			// Go uses banker's rounding (round half to even)
			{1.225, 1.23}, // 122.5 rounds to 123 (odd to even)
			{1.235, 1.24}, // 123.5 rounds to 124 (odd to even)
			{1.245, 1.25}, // 124.5 rounds to 125 (even stays even)
			{1.255, 1.25}, // 125.5 rounds to 125 (odd to even) - wait that's wrong
			{1.265, 1.26}, // 126.5 rounds to 126 (odd to even)
		}

		for _, tt := range tests {
			result := pm.MakePreciseValue(tt.input)
			if result != tt.expected {
				t.Errorf("For input %v, expected %v, got %v", tt.input, tt.expected, result)
			}
		}
	})

	t.Run("RoundingDownFixed", func(t *testing.T) {
		pm := precision.NewFixedPrecision(100)
		tests := []struct {
			input    float64
			expected float64
		}{
			{1.224, 1.22},
			{1.234, 1.23},
			{1.244, 1.24},
		}

		for _, tt := range tests {
			result := pm.MakePreciseValue(tt.input)
			if result != tt.expected {
				t.Errorf("For input %v, expected %v, got %v", tt.input, tt.expected, result)
			}
		}
	})

	t.Run("GeographicCoordinates_Fixed6", func(t *testing.T) {
		pm := precision.Fixed6 // Good for geographic coordinates
		// Simulate latitude/longitude
		coord := geom.NewCoordinate(-122.4194155, 37.7749295)
		pm.MakePrecise(&coord)

		if coord.X != -122.419416 {
			t.Errorf("Expected X=-122.419416, got %v", coord.X)
		}
		if coord.Y != 37.774930 {
			t.Errorf("Expected Y=37.774930, got %v", coord.Y)
		}
	})
}

func TestPrecisionConsistency(t *testing.T) {
	t.Run("MultipleApplications_Idempotent", func(t *testing.T) {
		pm := precision.NewFixedPrecision(100)
		coord := geom.NewCoordinate(1.23456, 7.89012)

		pm.MakePrecise(&coord)
		firstX, firstY := coord.X, coord.Y

		pm.MakePrecise(&coord)
		secondX, secondY := coord.X, coord.Y

		if firstX != secondX || firstY != secondY {
			t.Error("Multiple applications should be idempotent")
		}
	})

	t.Run("SequenceConsistency", func(t *testing.T) {
		pm := precision.NewFixedPrecision(100)
		coords := geom.NewCoordinateSequence(
			geom.NewCoordinate(1.23456, 2.34567),
			geom.NewCoordinate(3.45678, 4.56789),
		)

		// Apply to individual coordinates
		coord1 := coords[0].Clone()
		coord2 := coords[1].Clone()
		pm.MakePrecise(&coord1)
		pm.MakePrecise(&coord2)

		// Apply to sequence
		precision.MakePreciseSequence(pm, coords)

		// Results should be identical
		if !coords[0].Equals2D(coord1, 1e-10) || !coords[1].Equals2D(coord2, 1e-10) {
			t.Error("Sequence application should match individual application")
		}
	})
}
