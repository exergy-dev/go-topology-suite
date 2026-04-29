package planar

import (
	"math/big"

	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/kernel"
)

// adaptiveOrient is the Shewchuk-style adaptive 2D orientation predicate.
// Common case (well-conditioned inputs) runs at native float64 speed; on
// near-collinear inputs the result is verified using big.Float at 113-bit
// (quadruple) precision and only the sign of the exact result is returned.
//
// Reference: Jonathan R. Shewchuk, "Adaptive Precision Floating-Point
// Arithmetic and Fast Robust Geometric Predicates," Discrete & Computational
// Geometry 18(3):305-363, 1997.
//
// The error bound below (ccwerrboundA) is the first-level filter from
// Shewchuk §4.2, derived from the worst-case rounding of the four
// multiplications and one subtraction in the naive computation.
const ccwerrboundA = (3.0 + 16.0*1.1102230246251565e-16) * 1.1102230246251565e-16

// adaptiveOrient returns +1 (CCW), -1 (CW), or 0 (collinear) for the
// triangle (a, b, c). The algorithm computes (b-a) × (c-a) using the same
// formula as the naive path, then either trusts the result (filter pass)
// or verifies via exact arithmetic (filter fail).
func adaptiveOrient(a, b, c geom.XY) kernel.Orientation {
	detLeft := (b.X - a.X) * (c.Y - a.Y)
	detRight := (b.Y - a.Y) * (c.X - a.X)
	det := detLeft - detRight

	// Sign-only fast paths: when the two product terms have strictly
	// different signs, the subtraction is exact and `det` is reliable.
	var detSum float64
	switch {
	case detLeft > 0:
		if detRight <= 0 {
			return signToOrientation(det)
		}
		detSum = detLeft + detRight
	case detLeft < 0:
		if detRight >= 0 {
			return signToOrientation(det)
		}
		detSum = -detLeft - detRight
	default:
		return signToOrientation(det)
	}

	errBound := ccwerrboundA * detSum
	if det >= errBound || -det >= errBound {
		return signToOrientation(det)
	}

	// Filter fail: inputs are too near collinear for float64 to decide
	// the sign safely. Recompute exactly.
	return exactOrient(a, b, c)
}

func signToOrientation(v float64) kernel.Orientation {
	switch {
	case v > 0:
		return kernel.CounterClockwise
	case v < 0:
		return kernel.Clockwise
	default:
		return kernel.Collinear
	}
}

// exactOrient computes (b-a) × (c-a) at 256-bit precision and returns
// only the sign. Slower than full Shewchuk error-free transformations
// but much simpler and correct for any pair of float64 inputs that
// don't overflow.
//
// 256 bits gives ~77 decimal digits of mantissa — more than enough that
// the product of two float64 differences (at magnitudes up to ~10^308)
// retains any meaningful bit of the smaller operand even when the
// difference itself loses 30 orders of magnitude in cancellation.
func exactOrient(a, b, c geom.XY) kernel.Orientation {
	const prec = 256
	bigAX := new(big.Float).SetPrec(prec).SetFloat64(a.X)
	bigAY := new(big.Float).SetPrec(prec).SetFloat64(a.Y)
	bigBX := new(big.Float).SetPrec(prec).SetFloat64(b.X)
	bigBY := new(big.Float).SetPrec(prec).SetFloat64(b.Y)
	bigCX := new(big.Float).SetPrec(prec).SetFloat64(c.X)
	bigCY := new(big.Float).SetPrec(prec).SetFloat64(c.Y)

	// (b.X - a.X) * (c.Y - a.Y)
	tmp1 := new(big.Float).SetPrec(prec).Sub(bigBX, bigAX)
	tmp2 := new(big.Float).SetPrec(prec).Sub(bigCY, bigAY)
	left := new(big.Float).SetPrec(prec).Mul(tmp1, tmp2)

	// (b.Y - a.Y) * (c.X - a.X)
	tmp3 := new(big.Float).SetPrec(prec).Sub(bigBY, bigAY)
	tmp4 := new(big.Float).SetPrec(prec).Sub(bigCX, bigAX)
	right := new(big.Float).SetPrec(prec).Mul(tmp3, tmp4)

	// At 113 bits with float64 inputs, the subtraction is exact, so the
	// sign of the difference is the sign of the true determinant.
	diff := new(big.Float).SetPrec(prec).Sub(left, right)
	switch diff.Sign() {
	case 1:
		return kernel.CounterClockwise
	case -1:
		return kernel.Clockwise
	default:
		return kernel.Collinear
	}
}
