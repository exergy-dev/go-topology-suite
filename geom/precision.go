// Port of org.locationtech.jts.geom.PrecisionModel.
//
// PrecisionModel describes the grid of allowable coordinate values for a
// geometry. It comes in three flavours: full IEEE-754 double precision
// (Floating, the default), single-precision floating (FloatingSingle, ~6
// significant digits), and Fixed precision with a configurable scale factor
// or grid size.
//
// This package only adds the type and the MakePrecise helpers; plumbing
// the model through the buffer / overlay / WKT writer APIs is a follow-up.

package geom

import (
	"fmt"
	"math"
)

// PrecisionType identifies one of the three supported precision regimes.
type PrecisionType int

const (
	// PrecisionFloating is the standard IEEE-754 double precision model.
	// No rounding is performed; this is the default for newly constructed
	// geometries.
	PrecisionFloating PrecisionType = iota

	// PrecisionFloatingSingle truncates to single-precision (~6 significant
	// digits) by round-tripping through float32.
	PrecisionFloatingSingle

	// PrecisionFixed snaps coordinates to a regular grid defined by a scale
	// factor (gridSize == 1/scale).
	PrecisionFixed
)

// MaxPreciseValue is the largest integer representable exactly in IEEE-754
// double precision: 2^53. Mirrors PrecisionModel.maximumPreciseValue.
const MaxPreciseValue = 9007199254740992.0

// PrecisionModel specifies the coordinate grid of a Geometry. The zero
// value is a Floating model (full double precision, no rounding).
//
// Values are immutable; methods take a PrecisionModel by value. Construct
// via NewFloatingPrecision, NewFloatingSinglePrecision or NewFixedPrecision.
type PrecisionModel struct {
	modelType PrecisionType
	// scale is the multiplicative factor used by the Fixed model
	// (the snap grid has spacing 1/scale).
	scale float64
	// gridSize is non-zero when the model was constructed from an explicit
	// negative scale (i.e. a grid size). When zero, gridSize is derived
	// from scale on demand.
	gridSize float64
}

// NewFloatingPrecision returns the default full-precision model.
func NewFloatingPrecision() PrecisionModel {
	return PrecisionModel{modelType: PrecisionFloating}
}

// NewFloatingSinglePrecision returns a single-precision (~6 sig-fig) model.
func NewFloatingSinglePrecision() PrecisionModel {
	return PrecisionModel{modelType: PrecisionFloatingSingle}
}

// NewFixedPrecision returns a fixed-precision model with the given scale.
//
// A positive scale is interpreted as a multiplicative factor: coordinates
// are snapped to a grid of spacing 1/scale. A negative scale is interpreted
// as an explicit grid size: |scale| is the snap spacing, and the scale
// itself is its reciprocal.
//
// Mirrors the JTS PrecisionModel(double scale) constructor.
func NewFixedPrecision(scale float64) PrecisionModel {
	pm := PrecisionModel{modelType: PrecisionFixed}
	if scale < 0 {
		pm.gridSize = math.Abs(scale)
		pm.scale = 1.0 / pm.gridSize
	} else {
		pm.scale = math.Abs(scale)
	}
	return pm
}

// Type returns the model type.
func (pm PrecisionModel) Type() PrecisionType { return pm.modelType }

// Scale returns the fixed-model scale factor (zero for floating models).
func (pm PrecisionModel) Scale() float64 { return pm.scale }

// IsFloating reports whether the model is one of the floating types
// (no rounding required).
func (pm PrecisionModel) IsFloating() bool {
	return pm.modelType == PrecisionFloating || pm.modelType == PrecisionFloatingSingle
}

// GridSize returns the snap grid spacing for a fixed model, or NaN for
// floating models. If an explicit grid size was set via a negative scale,
// it is returned directly; otherwise it is computed as 1/scale.
func (pm PrecisionModel) GridSize() float64 {
	if pm.IsFloating() {
		return math.NaN()
	}
	if pm.gridSize != 0 {
		return pm.gridSize
	}
	return 1.0 / pm.scale
}

// MaximumSignificantDigits approximates the number of decimal digits the
// model preserves. For Fixed models the value is 1 + ceil(log10(scale)).
//
// Mirrors PrecisionModel.getMaximumSignificantDigits.
func (pm PrecisionModel) MaximumSignificantDigits() int {
	switch pm.modelType {
	case PrecisionFloatingSingle:
		return 6
	case PrecisionFixed:
		return 1 + int(math.Ceil(math.Log10(pm.scale)))
	default:
		return 16
	}
}

// MakePreciseValue snaps a single ordinate. NaN and infinite values are
// returned unchanged. For Floating, the value is returned as-is.
//
// Mirrors PrecisionModel.makePrecise(double).
func (pm PrecisionModel) MakePreciseValue(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return v
	}
	switch pm.modelType {
	case PrecisionFloatingSingle:
		return float64(float32(v))
	case PrecisionFixed:
		if pm.gridSize > 0 {
			return math.Round(v/pm.gridSize) * pm.gridSize
		}
		return math.Round(v*pm.scale) / pm.scale
	default:
		return v
	}
}

// MakePrecise snaps an XY coordinate. The Floating model returns p
// unchanged.
func (pm PrecisionModel) MakePrecise(p XY) XY {
	if pm.modelType == PrecisionFloating {
		return p
	}
	return XY{X: pm.MakePreciseValue(p.X), Y: pm.MakePreciseValue(p.Y)}
}

// String returns a JTS-style description.
func (pm PrecisionModel) String() string {
	switch pm.modelType {
	case PrecisionFloating:
		return "Floating"
	case PrecisionFloatingSingle:
		return "Floating-Single"
	case PrecisionFixed:
		return fmt.Sprintf("Fixed (Scale=%g)", pm.scale)
	}
	return "UNKNOWN"
}

// Equal reports whether two precision models snap to the same grid.
func (pm PrecisionModel) Equal(other PrecisionModel) bool {
	return pm.modelType == other.modelType && pm.scale == other.scale
}

// Compare returns -1/0/+1 by ordering models from least to most precise.
//
// Mirrors PrecisionModel.compareTo: the ordering is by
// MaximumSignificantDigits.
func (pm PrecisionModel) Compare(other PrecisionModel) int {
	a, b := pm.MaximumSignificantDigits(), other.MaximumSignificantDigits()
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}
