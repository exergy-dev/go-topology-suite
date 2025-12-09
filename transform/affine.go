package transform

import (
	"fmt"
	"math"
)

// Affine represents a 2D affine transformation using a 3x3 matrix.
// The transformation is defined as:
//   x' = A*x + B*y + C
//   y' = D*x + E*y + F
//
// This can represent translation, scaling, rotation, shearing, and
// combinations of these operations.
type Affine struct {
	A, B, C float64 // First row: x' = Ax + By + C
	D, E, F float64 // Second row: y' = Dx + Ey + F
}

// NewAffineIdentity creates an identity affine transformation.
// The identity transformation leaves all coordinates unchanged.
func NewAffineIdentity() *Affine {
	return &Affine{
		A: 1, B: 0, C: 0,
		D: 0, E: 1, F: 0,
	}
}

// NewAffineTranslation creates a translation transformation.
// All points are shifted by (dx, dy).
func NewAffineTranslation(dx, dy float64) *Affine {
	return &Affine{
		A: 1, B: 0, C: dx,
		D: 0, E: 1, F: dy,
	}
}

// NewAffineScale creates a scaling transformation.
// Points are scaled by sx in the x-direction and sy in the y-direction.
// Scaling is performed about the origin (0, 0).
func NewAffineScale(sx, sy float64) *Affine {
	return &Affine{
		A: sx, B: 0, C: 0,
		D: 0, E: sy, F: 0,
	}
}

// NewAffineScaleOrigin creates a scaling transformation about a specified origin point.
// Points are scaled by sx and sy relative to the origin point (ox, oy).
func NewAffineScaleOrigin(sx, sy, ox, oy float64) *Affine {
	return &Affine{
		A: sx,
		B: 0,
		C: ox - sx*ox,
		D: 0,
		E: sy,
		F: oy - sy*oy,
	}
}

// NewAffineRotation creates a rotation transformation.
// Points are rotated counter-clockwise by the given angle (in radians)
// about the origin (0, 0).
func NewAffineRotation(angle float64) *Affine {
	cos := math.Cos(angle)
	sin := math.Sin(angle)
	return &Affine{
		A: cos, B: -sin, C: 0,
		D: sin, E: cos, F: 0,
	}
}

// NewAffineRotationOrigin creates a rotation transformation about a specified origin.
// Points are rotated counter-clockwise by the given angle (in radians)
// about the point (ox, oy).
func NewAffineRotationOrigin(angle, ox, oy float64) *Affine {
	cos := math.Cos(angle)
	sin := math.Sin(angle)
	return &Affine{
		A: cos,
		B: -sin,
		C: ox - cos*ox + sin*oy,
		D: sin,
		E: cos,
		F: oy - sin*ox - cos*oy,
	}
}

// NewAffineShear creates a shear transformation.
// shearX controls shearing in the x-direction (x' = x + shearX*y).
// shearY controls shearing in the y-direction (y' = y + shearY*x).
func NewAffineShear(shearX, shearY float64) *Affine {
	return &Affine{
		A: 1, B: shearX, C: 0,
		D: shearY, E: 1, F: 0,
	}
}

// NewAffine creates an affine transformation with the specified matrix values.
func NewAffine(a, b, c, d, e, f float64) *Affine {
	return &Affine{A: a, B: b, C: c, D: d, E: e, F: f}
}

// Forward applies the affine transformation to the given coordinates.
func (a *Affine) Forward(x, y float64) (float64, float64, error) {
	xNew := a.A*x + a.B*y + a.C
	yNew := a.D*x + a.E*y + a.F
	return xNew, yNew, nil
}

// Inverse applies the inverse affine transformation to the given coordinates.
// Returns an error if the transformation matrix is singular (non-invertible).
func (a *Affine) Inverse(x, y float64) (float64, float64, error) {
	// Compute the determinant
	det := a.A*a.E - a.B*a.D

	if math.Abs(det) < 1e-10 {
		return 0, 0, fmt.Errorf("affine transformation is singular (determinant = %g)", det)
	}

	// Apply the inverse transformation
	// For matrix [[A, B, C], [D, E, F], [0, 0, 1]]
	// The inverse is [[E/det, -B/det, (B*F-C*E)/det],
	//                 [-D/det, A/det, (C*D-A*F)/det],
	//                 [0, 0, 1]]

	xAdj := x - a.C
	yAdj := y - a.F

	xNew := (a.E*xAdj - a.B*yAdj) / det
	yNew := (-a.D*xAdj + a.A*yAdj) / det

	return xNew, yNew, nil
}

// Compose creates a new affine transformation that is the composition
// of this transformation followed by the other transformation.
// The result is equivalent to applying this transform first, then other.
func (a *Affine) Compose(other *Affine) *Affine {
	return &Affine{
		A: other.A*a.A + other.B*a.D,
		B: other.A*a.B + other.B*a.E,
		C: other.A*a.C + other.B*a.F + other.C,
		D: other.D*a.A + other.E*a.D,
		E: other.D*a.B + other.E*a.E,
		F: other.D*a.C + other.E*a.F + other.F,
	}
}

// Determinant returns the determinant of the transformation matrix.
// A determinant of 0 indicates a singular (non-invertible) transformation.
// The absolute value of the determinant represents the area scaling factor.
func (a *Affine) Determinant() float64 {
	return a.A*a.E - a.B*a.D
}

// IsIdentity returns true if this is the identity transformation
// (within floating-point tolerance).
func (a *Affine) IsIdentity() bool {
	const epsilon = 1e-10
	return math.Abs(a.A-1) < epsilon &&
		math.Abs(a.B) < epsilon &&
		math.Abs(a.C) < epsilon &&
		math.Abs(a.D) < epsilon &&
		math.Abs(a.E-1) < epsilon &&
		math.Abs(a.F) < epsilon
}

// Clone returns a copy of this affine transformation.
func (a *Affine) Clone() *Affine {
	return &Affine{
		A: a.A, B: a.B, C: a.C,
		D: a.D, E: a.E, F: a.F,
	}
}
