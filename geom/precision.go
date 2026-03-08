package geom

import (
	"math"
)

// PrecisionModel defines how coordinates are made precise.
type PrecisionModel interface {
	// MakePrecise adjusts a coordinate to the model's precision.
	MakePrecise(coord *Coordinate)

	// MakePreciseValue adjusts a single value to the model's precision.
	MakePreciseValue(val float64) float64

	// Scale returns the scale factor (0 for floating precision).
	Scale() float64

	// Type returns the type of precision model.
	Type() PrecisionType

	// IsFloating returns true if this is a floating precision model.
	IsFloating() bool

	// MaxSignificantDigits returns the maximum precision digits.
	MaxSignificantDigits() int
}

// PrecisionType represents the type of precision model.
type PrecisionType int

const (
	// FloatingPrecision uses full double precision (default).
	FloatingPrecision PrecisionType = iota
	// FloatingSinglePrecision uses single (float32) precision.
	FloatingSinglePrecision
	// FixedPrecision uses a fixed scale factor.
	FixedPrecision
)

// floatingPrecisionModel implements full double precision.
type floatingPrecisionModel struct{}

// NewFloatingPrecision creates a new floating precision model.
func NewFloatingPrecision() PrecisionModel {
	return &floatingPrecisionModel{}
}

func (f *floatingPrecisionModel) MakePrecise(coord *Coordinate) {
	// No-op for floating precision
}

func (f *floatingPrecisionModel) MakePreciseValue(val float64) float64 {
	return val
}

func (f *floatingPrecisionModel) Scale() float64 {
	return 0 // Indicates no scaling
}

func (f *floatingPrecisionModel) Type() PrecisionType {
	return FloatingPrecision
}

func (f *floatingPrecisionModel) IsFloating() bool {
	return true
}

func (f *floatingPrecisionModel) MaxSignificantDigits() int {
	return 16 // Approximate double precision
}

// floatingSinglePrecisionModel implements single (float32) precision.
type floatingSinglePrecisionModel struct{}

// NewFloatingSinglePrecision creates a new single precision model.
func NewFloatingSinglePrecision() PrecisionModel {
	return &floatingSinglePrecisionModel{}
}

func (f *floatingSinglePrecisionModel) MakePrecise(coord *Coordinate) {
	coord.X = float64(float32(coord.X))
	coord.Y = float64(float32(coord.Y))
	if coord.HasZ() {
		coord.Z = float64(float32(coord.Z))
	}
	if coord.HasM() {
		coord.M = float64(float32(coord.M))
	}
}

func (f *floatingSinglePrecisionModel) MakePreciseValue(val float64) float64 {
	return float64(float32(val))
}

func (f *floatingSinglePrecisionModel) Scale() float64 {
	return 0 // Indicates no scaling
}

func (f *floatingSinglePrecisionModel) Type() PrecisionType {
	return FloatingSinglePrecision
}

func (f *floatingSinglePrecisionModel) IsFloating() bool {
	return true
}

func (f *floatingSinglePrecisionModel) MaxSignificantDigits() int {
	return 6 // Approximate single precision
}

// fixedPrecisionModel implements fixed-scale precision.
type fixedPrecisionModel struct {
	scale float64
}

// NewFixedPrecision creates a new fixed precision model with the given scale.
// The scale determines the number of decimal places (e.g., 1000 = 3 decimal places).
func NewFixedPrecision(scale float64) PrecisionModel {
	if scale <= 0 {
		scale = 1
	}
	return &fixedPrecisionModel{scale: scale}
}

func (f *fixedPrecisionModel) MakePrecise(coord *Coordinate) {
	coord.X = f.MakePreciseValue(coord.X)
	coord.Y = f.MakePreciseValue(coord.Y)
	if coord.HasZ() {
		coord.Z = f.MakePreciseValue(coord.Z)
	}
	if coord.HasM() {
		coord.M = f.MakePreciseValue(coord.M)
	}
}

func (f *fixedPrecisionModel) MakePreciseValue(val float64) float64 {
	return math.Round(val*f.scale) / f.scale
}

func (f *fixedPrecisionModel) Scale() float64 {
	return f.scale
}

func (f *fixedPrecisionModel) Type() PrecisionType {
	return FixedPrecision
}

func (f *fixedPrecisionModel) IsFloating() bool {
	return false
}

func (f *fixedPrecisionModel) MaxSignificantDigits() int {
	return int(math.Log10(f.scale)) + 1
}

// Common precision models.
var (
	// Floating is the default floating point precision.
	Floating = NewFloatingPrecision()

	// FloatingSingle is single precision (float32).
	FloatingSingle = NewFloatingSinglePrecision()

	// Fixed1 has 1 decimal place (scale 10).
	Fixed1 = NewFixedPrecision(10)

	// Fixed2 has 2 decimal places (scale 100).
	Fixed2 = NewFixedPrecision(100)

	// Fixed3 has 3 decimal places (scale 1000).
	Fixed3 = NewFixedPrecision(1000)

	// Fixed6 has 6 decimal places (scale 1,000,000) - good for geographic coords.
	Fixed6 = NewFixedPrecision(1000000)

	// Fixed8 has 8 decimal places (scale 100,000,000) - high precision geographic.
	Fixed8 = NewFixedPrecision(100000000)
)

// MakePreciseSequence applies a precision model to a coordinate sequence.
func MakePreciseSequence(pm PrecisionModel, coords CoordinateSequence) {
	for i := range coords {
		pm.MakePrecise(&coords[i])
	}
}

// ComparePrecision compares two precision models.
// Returns:
//
//	-1 if pm1 is less precise than pm2
//	 0 if pm1 and pm2 are equally precise
//	 1 if pm1 is more precise than pm2
func ComparePrecision(pm1, pm2 PrecisionModel) int {
	// Floating precision is always most precise
	if pm1.Type() == FloatingPrecision && pm2.Type() != FloatingPrecision {
		return 1
	}
	if pm1.Type() != FloatingPrecision && pm2.Type() == FloatingPrecision {
		return -1
	}
	if pm1.Type() == FloatingPrecision && pm2.Type() == FloatingPrecision {
		return 0
	}

	// Compare scales for fixed precision
	scale1 := pm1.Scale()
	scale2 := pm2.Scale()
	if scale1 > scale2 {
		return 1
	}
	if scale1 < scale2 {
		return -1
	}
	return 0
}

// MostPrecise returns the more precise of two precision models.
func MostPrecise(pm1, pm2 PrecisionModel) PrecisionModel {
	if ComparePrecision(pm1, pm2) >= 0 {
		return pm1
	}
	return pm2
}

// LeastPrecise returns the less precise of two precision models.
func LeastPrecise(pm1, pm2 PrecisionModel) PrecisionModel {
	if ComparePrecision(pm1, pm2) <= 0 {
		return pm1
	}
	return pm2
}
