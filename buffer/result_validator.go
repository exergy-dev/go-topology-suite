package buffer

import (
	"fmt"
	"math"

	"github.com/exergy-dev/go-topology-suite/geom"
	"github.com/exergy-dev/go-topology-suite/measure"
)

// maxBufferEnvelopeDiffFrac is the allowable fractional difference
// between the expected and actual buffer envelope. JTS uses 1.2% (1%
// caused occasional false positives).
const maxBufferEnvelopeDiffFrac = 0.012

// ValidationErrorKind classifies the failure mode reported by
// ValidateBufferResult.
type ValidationErrorKind int

const (
	// ValidationErrorPolygonal — the buffer result is not a Polygon or
	// MultiPolygon.
	ValidationErrorPolygonal ValidationErrorKind = iota
	// ValidationErrorExpectedEmpty — the result should have been empty
	// (negative buffer of a non-areal input).
	ValidationErrorExpectedEmpty
	// ValidationErrorEnvelope — the result envelope does not approximately
	// match the expected (input expanded by distance).
	ValidationErrorEnvelope
	// ValidationErrorArea — the result area is incompatible with the sign
	// of the buffer distance (positive buffer cannot shrink area; negative
	// cannot grow it).
	ValidationErrorArea
	// ValidationErrorDistance — the distance between input and buffer
	// boundary deviates from the requested distance by more than tolerance.
	ValidationErrorDistance
)

// String returns a short identifier for kind.
func (k ValidationErrorKind) String() string {
	switch k {
	case ValidationErrorPolygonal:
		return "Polygonal"
	case ValidationErrorExpectedEmpty:
		return "ExpectedEmpty"
	case ValidationErrorEnvelope:
		return "Envelope"
	case ValidationErrorArea:
		return "Area"
	case ValidationErrorDistance:
		return "Distance"
	}
	return fmt.Sprintf("Unknown(%d)", int(k))
}

// ValidationError describes a single buffer validation failure. Multiple
// errors may be returned by ValidateBufferResult — though typically the
// first failure is the most informative.
type ValidationError struct {
	Kind     ValidationErrorKind
	Message  string
	Location geom.XY // populated for Distance errors
}

// Error implements the error interface.
func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Kind, e.Message)
}

// ValidateBufferResult verifies that output is geometrically a valid
// buffer of input at the requested distance. Returns a slice of validation
// errors, or an empty/nil slice on success.
//
// This is a heuristic test: it should never report a valid buffer as
// invalid (no false negatives), but may miss subtle errors (false
// positives are possible). It can be much more expensive than the buffer
// itself; intended for CI rails and debugging, not the hot path.
//
// Port of org.locationtech.jts.operation.buffer.validate.BufferResultValidator.
func ValidateBufferResult(input, output geom.Geometry, distance float64) []ValidationError {
	v := newBufferResultValidator(input, output, distance)
	v.runChecks()
	return v.errs
}

type bufferResultValidator struct {
	input    geom.Geometry
	output   geom.Geometry
	distance float64
	errs     []ValidationError
}

func newBufferResultValidator(input, output geom.Geometry, distance float64) *bufferResultValidator {
	return &bufferResultValidator{input: input, output: output, distance: distance}
}

// runChecks executes the validation pipeline. Each stage may abort early
// to avoid producing cascading false-positive errors.
func (v *bufferResultValidator) runChecks() {
	if v.checkPolygonal(); len(v.errs) > 0 {
		return
	}
	if v.checkExpectedEmpty(); len(v.errs) > 0 {
		return
	}
	if v.checkEnvelope(); len(v.errs) > 0 {
		return
	}
	if v.checkArea(); len(v.errs) > 0 {
		return
	}
	v.checkDistance()
}

// checkPolygonal — the buffer result must be a Polygon or MultiPolygon
// (or empty). An empty result is permitted; per JTS, the polygonal type
// check accepts both Polygon and MultiPolygon.
func (v *bufferResultValidator) checkPolygonal() {
	if v.output == nil {
		return
	}
	if v.output.IsEmpty() {
		return
	}
	switch v.output.(type) {
	case *geom.Polygon, *geom.MultiPolygon:
		return
	}
	v.errs = append(v.errs, ValidationError{
		Kind:    ValidationErrorPolygonal,
		Message: fmt.Sprintf("Result is not polygonal (got %T)", v.output),
	})
}

// checkExpectedEmpty — a non-positive buffer of a non-areal input must
// produce an empty result.
func (v *bufferResultValidator) checkExpectedEmpty() {
	// Areal inputs can have non-empty negative buffers — skip.
	if isAreal(v.input) {
		return
	}
	if v.distance > 0 {
		return
	}
	if v.output == nil || v.output.IsEmpty() {
		return
	}
	v.errs = append(v.errs, ValidationError{
		Kind:    ValidationErrorExpectedEmpty,
		Message: "Result is non-empty for non-positive buffer of non-areal input",
	})
}

// checkEnvelope — for a positive buffer, the result envelope (slightly
// padded for floating-point slop) must contain the input envelope expanded
// by distance.
func (v *bufferResultValidator) checkEnvelope() {
	if v.distance < 0 {
		return
	}
	if v.input == nil || v.input.IsEmpty() {
		return
	}
	if v.output == nil || v.output.IsEmpty() {
		return
	}
	padding := v.distance * maxBufferEnvelopeDiffFrac
	if padding == 0 {
		padding = 0.001
	}
	expected := v.input.Envelope().ExpandBy(v.distance)
	actual := v.output.Envelope().ExpandBy(padding)
	if !actual.Contains(expected) {
		v.errs = append(v.errs, ValidationError{
			Kind: ValidationErrorEnvelope,
			Message: fmt.Sprintf(
				"Buffer envelope is incorrect (expected %v, got %v)",
				expected, v.output.Envelope()),
		})
	}
}

// checkArea — a positive buffer never shrinks area; a negative buffer
// never grows it. Only meaningful when the input has area.
func (v *bufferResultValidator) checkArea() {
	if v.input == nil || v.output == nil {
		return
	}
	inA := measure.Area(v.input)
	outA := measure.Area(v.output)
	if v.distance > 0 && inA > outA {
		v.errs = append(v.errs, ValidationError{
			Kind: ValidationErrorArea,
			Message: fmt.Sprintf(
				"Area of positive buffer (%g) is smaller than input (%g)",
				outA, inA),
		})
	}
	if v.distance < 0 && inA < outA {
		v.errs = append(v.errs, ValidationError{
			Kind: ValidationErrorArea,
			Message: fmt.Sprintf(
				"Area of negative buffer (%g) is larger than input (%g)",
				outA, inA),
		})
	}
}

// checkDistance delegates to ValidateBufferDistance.
func (v *bufferResultValidator) checkDistance() {
	if v.input == nil || v.output == nil {
		return
	}
	if v.input.IsEmpty() || v.output.IsEmpty() {
		return
	}
	if math.IsNaN(v.distance) {
		return
	}
	r := validateBufferDistance(v.input, v.output, v.distance)
	if r.ok {
		return
	}
	v.errs = append(v.errs, ValidationError{
		Kind:     ValidationErrorDistance,
		Message:  fmt.Sprintf("%s (signed err=%g)", r.errorMessage, r.maxError),
		Location: r.errorLocation,
	})
}

// isAreal reports whether g has area-bearing components.
func isAreal(g geom.Geometry) bool {
	if g == nil || g.IsEmpty() {
		return false
	}
	switch v := g.(type) {
	case *geom.Polygon, *geom.MultiPolygon:
		return true
	case *geom.GeometryCollection:
		for i := 0; i < v.NumGeometries(); i++ {
			if isAreal(v.GeometryAt(i)) {
				return true
			}
		}
	}
	return false
}
