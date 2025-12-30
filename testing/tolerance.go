// Package testing provides test utilities and JTS-compatible tolerance constants
// for validating GTS geometric operations.
package testing

// JTS-compatible tolerance constants for geometric operations.
// These values are derived from JTS (Java Topology Suite) validation constants
// to ensure GTS achieves the same level of accuracy.
const (
	// BufferDistanceTolerance is the maximum allowable fraction of buffer distance
	// for validation. Matches JTS MAX_DISTANCE_DIFF_FRAC from BufferDistanceValidator.
	BufferDistanceTolerance = 0.012 // 1.2%

	// BufferEnvelopeTolerance is the maximum allowable fraction for envelope validation.
	// Matches JTS MAX_ENV_DIFF_FRAC from BufferDistanceValidator.
	BufferEnvelopeTolerance = 0.012 // 1.2%

	// AreaTolerance is the default relative tolerance for area comparisons.
	AreaTolerance = 0.01 // 1%

	// CoordinateTolerance is the absolute tolerance for coordinate comparisons.
	// This is the standard precision used in JTS for coordinate equality.
	CoordinateTolerance = 1e-10

	// DistanceTolerance is the absolute tolerance for distance calculations.
	DistanceTolerance = 1e-9

	// SnapPrecisionFactor is used for size-based snap tolerance calculation.
	// Matches JTS SNAP_PRECISION_FACTOR.
	SnapPrecisionFactor = 1e-9

	// OverlayAreaTolerance is the relative tolerance for overlay operation areas.
	OverlayAreaTolerance = 0.01 // 1%

	// PointBufferTolerance is the tolerance for point buffer (circle) area calculations.
	// With default 8 quadrant segments, the approximation error is bounded.
	PointBufferTolerance = 0.02 // 2% for default quality (8 segments)

	// HighQualityBufferTolerance is for high-resolution buffers (32+ quadrant segments).
	HighQualityBufferTolerance = 0.001 // 0.1%
)

// BufferQualityTolerance maps quadrant segment counts to expected area tolerances.
// Higher segment counts produce more accurate circular approximations.
var BufferQualityTolerance = map[int]float64{
	2:  0.30,   // Octagon approximation - very rough
	4:  0.10,   // 16-gon
	8:  0.02,   // JTS default - < 2% error
	12: 0.01,   // < 1% error
	16: 0.005,  // < 0.5% error
	18: 0.001,  // < 0.1% error
	32: 0.0005, // < 0.05% error
}

// ToleranceForQuadrantSegments returns the appropriate area tolerance
// for the given number of quadrant segments.
func ToleranceForQuadrantSegments(segments int) float64 {
	if tol, ok := BufferQualityTolerance[segments]; ok {
		return tol
	}
	// For unlisted segment counts, interpolate or use conservative estimate
	if segments < 8 {
		return 0.10 // Conservative for low segment counts
	}
	if segments < 16 {
		return 0.02
	}
	if segments < 32 {
		return 0.005
	}
	return 0.001 // High quality for 32+
}
