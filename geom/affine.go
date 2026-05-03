// Port of org.locationtech.jts.geom.util.AffineTransformation.
//
// Represents a 2D affine transformation as a 3x3 matrix:
//
//	T = | m00 m01 m02 |
//	    | m10 m11 m12 |
//	    |  0   0   1  |
//
// Affine transformations preserve straightness and parallelism but not
// distance or shape in general. Composition of transformations is
// computed via matrix multiplication and is generally non-commutative.
//
// The Compose method follows the JTS convention:
//
//	A.Compose(B) = T_B x T_A
//
// i.e. the resulting transform applies A first, then B.

package geom

import (
	"errors"
	"fmt"
	"math"
)

// AffineTransformation is a 2D affine transformation represented as a
// 3x3 matrix (the bottom row is implicitly [0 0 1]).
//
// A zero-value AffineTransformation is NOT the identity; use
// NewAffineTransformation() to obtain an identity transform.
type AffineTransformation struct {
	m00, m01, m02 float64
	m10, m11, m12 float64
}

// ErrNoninvertibleTransformation is returned by Inverse when the
// transformation matrix has determinant zero.
var ErrNoninvertibleTransformation = errors.New("geom: affine transformation is non-invertible")

// NewAffineTransformation returns a new identity transformation.
func NewAffineTransformation() *AffineTransformation {
	t := &AffineTransformation{}
	t.SetToIdentity()
	return t
}

// NewAffineTransformationMatrix returns a transformation with the given
// matrix entries.
func NewAffineTransformationMatrix(m00, m01, m02, m10, m11, m12 float64) *AffineTransformation {
	return &AffineTransformation{m00: m00, m01: m01, m02: m02, m10: m10, m11: m11, m12: m12}
}

// AffineTranslation returns a transformation for a translation by (dx, dy).
func AffineTranslation(dx, dy float64) *AffineTransformation {
	t := NewAffineTransformation()
	t.SetToTranslation(dx, dy)
	return t
}

// AffineRotation returns a transformation for a counter-clockwise
// rotation about the origin by theta radians.
func AffineRotation(theta float64) *AffineTransformation {
	t := NewAffineTransformation()
	t.SetToRotation(theta)
	return t
}

// AffineRotationAround returns a transformation for a counter-clockwise
// rotation about (x, y) by theta radians.
func AffineRotationAround(theta, x, y float64) *AffineTransformation {
	t := NewAffineTransformation()
	t.SetToRotationAround(theta, x, y)
	return t
}

// AffineScale returns a transformation for a scaling relative to the origin.
func AffineScale(sx, sy float64) *AffineTransformation {
	t := NewAffineTransformation()
	t.SetToScale(sx, sy)
	return t
}

// AffineShear returns a transformation for a shear by (xShear, yShear).
func AffineShear(xShear, yShear float64) *AffineTransformation {
	t := NewAffineTransformation()
	t.SetToShear(xShear, yShear)
	return t
}

// AffineReflection returns a transformation for a reflection about the
// line through (x0,y0) and (x1,y1).
func AffineReflection(x0, y0, x1, y1 float64) *AffineTransformation {
	t := NewAffineTransformation()
	t.SetToReflection(x0, y0, x1, y1)
	return t
}

// AffineReflectionVector returns a transformation for a reflection
// about the line through the origin in the direction of (x, y).
func AffineReflectionVector(x, y float64) *AffineTransformation {
	t := NewAffineTransformation()
	t.SetToReflectionVector(x, y)
	return t
}

// SetToIdentity sets t to the identity transformation.
func (t *AffineTransformation) SetToIdentity() *AffineTransformation {
	t.m00, t.m01, t.m02 = 1, 0, 0
	t.m10, t.m11, t.m12 = 0, 1, 0
	return t
}

// SetTransformation sets the matrix entries directly.
func (t *AffineTransformation) SetTransformation(m00, m01, m02, m10, m11, m12 float64) *AffineTransformation {
	t.m00, t.m01, t.m02 = m00, m01, m02
	t.m10, t.m11, t.m12 = m10, m11, m12
	return t
}

// SetTransformationFrom sets t to be a copy of other.
func (t *AffineTransformation) SetTransformationFrom(other *AffineTransformation) *AffineTransformation {
	*t = *other
	return t
}

// MatrixEntries returns a slice {m00, m01, m02, m10, m11, m12}.
func (t *AffineTransformation) MatrixEntries() []float64 {
	return []float64{t.m00, t.m01, t.m02, t.m10, t.m11, t.m12}
}

// Determinant returns the determinant of the 2x2 linear part of the
// matrix: m00*m11 - m01*m10. The full matrix is invertible iff this is
// non-zero.
func (t *AffineTransformation) Determinant() float64 {
	return t.m00*t.m11 - t.m01*t.m10
}

// Inverse returns a new transformation that is the inverse of t, or
// ErrNoninvertibleTransformation if t is singular.
func (t *AffineTransformation) Inverse() (*AffineTransformation, error) {
	det := t.Determinant()
	if det == 0 {
		return nil, ErrNoninvertibleTransformation
	}
	im00 := t.m11 / det
	im10 := -t.m10 / det
	im01 := -t.m01 / det
	im11 := t.m00 / det
	im02 := (t.m01*t.m12 - t.m02*t.m11) / det
	im12 := (-t.m00*t.m12 + t.m10*t.m02) / det
	return NewAffineTransformationMatrix(im00, im01, im02, im10, im11, im12), nil
}

// SetToTranslation sets t to a pure translation by (dx, dy).
func (t *AffineTransformation) SetToTranslation(dx, dy float64) *AffineTransformation {
	t.m00, t.m01, t.m02 = 1, 0, dx
	t.m10, t.m11, t.m12 = 0, 1, dy
	return t
}

// SetToScale sets t to a pure scaling about the origin.
func (t *AffineTransformation) SetToScale(sx, sy float64) *AffineTransformation {
	t.m00, t.m01, t.m02 = sx, 0, 0
	t.m10, t.m11, t.m12 = 0, sy, 0
	return t
}

// SetToShear sets t to a pure shear.
func (t *AffineTransformation) SetToShear(xShear, yShear float64) *AffineTransformation {
	t.m00, t.m01, t.m02 = 1, xShear, 0
	t.m10, t.m11, t.m12 = yShear, 1, 0
	return t
}

// SetToRotation sets t to a CCW rotation about the origin by theta radians.
func (t *AffineTransformation) SetToRotation(theta float64) *AffineTransformation {
	return t.SetToRotationSinCos(math.Sin(theta), math.Cos(theta))
}

// SetToRotationSinCos sets t to a CCW rotation about the origin
// specified by sin(theta) and cos(theta) directly.
func (t *AffineTransformation) SetToRotationSinCos(sinTheta, cosTheta float64) *AffineTransformation {
	t.m00, t.m01, t.m02 = cosTheta, -sinTheta, 0
	t.m10, t.m11, t.m12 = sinTheta, cosTheta, 0
	return t
}

// SetToRotationAround sets t to a CCW rotation about (x,y) by theta radians.
func (t *AffineTransformation) SetToRotationAround(theta, x, y float64) *AffineTransformation {
	return t.SetToRotationAroundSinCos(math.Sin(theta), math.Cos(theta), x, y)
}

// SetToRotationAroundSinCos sets t to a CCW rotation about (x,y)
// specified by sin(theta) and cos(theta) directly.
func (t *AffineTransformation) SetToRotationAroundSinCos(sinTheta, cosTheta, x, y float64) *AffineTransformation {
	t.m00, t.m01, t.m02 = cosTheta, -sinTheta, x-x*cosTheta+y*sinTheta
	t.m10, t.m11, t.m12 = sinTheta, cosTheta, y-x*sinTheta-y*cosTheta
	return t
}

// SetToReflection sets t to a reflection about the line through
// (x0,y0) and (x1,y1).
func (t *AffineTransformation) SetToReflection(x0, y0, x1, y1 float64) *AffineTransformation {
	if x0 == x1 && y0 == y1 {
		// Degenerate; leave matrix as identity (mirrors JTS by panic;
		// we choose a non-panicking degenerate identity).
		t.SetToIdentity()
		return t
	}
	// translate line vector to origin
	t.SetToTranslation(-x0, -y0)
	dx := x1 - x0
	dy := y1 - y0
	d := math.Hypot(dx, dy)
	sin := dy / d
	cos := dx / d
	t.Rotate(-sin, cos)
	t.Scale(1, -1)
	t.Rotate(sin, cos)
	t.Translate(x0, y0)
	return t
}

// SetToReflectionVector sets t to a reflection about the line through
// the origin in the direction (x, y).
func (t *AffineTransformation) SetToReflectionVector(x, y float64) *AffineTransformation {
	if x == 0 && y == 0 {
		t.SetToIdentity()
		return t
	}
	if x == y {
		// Special case to avoid roundoff.
		t.m00, t.m01, t.m02 = 0, 1, 0
		t.m10, t.m11, t.m12 = 1, 0, 0
		return t
	}
	d := math.Hypot(x, y)
	sin := y / d
	cos := x / d
	t.SetToIdentity()
	t.Rotate(-sin, cos)
	t.Scale(1, -1)
	t.Rotate(sin, cos)
	return t
}

// Translate composes t with a translation transformation.
func (t *AffineTransformation) Translate(dx, dy float64) *AffineTransformation {
	return t.Compose(AffineTranslation(dx, dy))
}

// Scale composes t with a scaling transformation.
func (t *AffineTransformation) Scale(sx, sy float64) *AffineTransformation {
	return t.Compose(AffineScale(sx, sy))
}

// Shear composes t with a shear transformation.
func (t *AffineTransformation) Shear(xShear, yShear float64) *AffineTransformation {
	return t.Compose(AffineShear(xShear, yShear))
}

// Rotate composes t with a rotation about the origin specified by
// sin(theta) and cos(theta).
func (t *AffineTransformation) Rotate(sinTheta, cosTheta float64) *AffineTransformation {
	r := NewAffineTransformation()
	r.SetToRotationSinCos(sinTheta, cosTheta)
	return t.Compose(r)
}

// RotateAngle composes t with a CCW rotation about the origin by theta radians.
func (t *AffineTransformation) RotateAngle(theta float64) *AffineTransformation {
	return t.Compose(AffineRotation(theta))
}

// Reflect composes t with a reflection about the line through (x0,y0)
// and (x1,y1).
func (t *AffineTransformation) Reflect(x0, y0, x1, y1 float64) *AffineTransformation {
	return t.Compose(AffineReflection(x0, y0, x1, y1))
}

// Compose updates t to the composition (other ∘ t), so that the result
// applies t first, then other:
//
//	A.Compose(B) = T_B x T_A
func (t *AffineTransformation) Compose(other *AffineTransformation) *AffineTransformation {
	mp00 := other.m00*t.m00 + other.m01*t.m10
	mp01 := other.m00*t.m01 + other.m01*t.m11
	mp02 := other.m00*t.m02 + other.m01*t.m12 + other.m02
	mp10 := other.m10*t.m00 + other.m11*t.m10
	mp11 := other.m10*t.m01 + other.m11*t.m11
	mp12 := other.m10*t.m02 + other.m11*t.m12 + other.m12
	t.m00, t.m01, t.m02 = mp00, mp01, mp02
	t.m10, t.m11, t.m12 = mp10, mp11, mp12
	return t
}

// ComposeBefore updates t to the composition (t ∘ other), so that the
// result applies other first, then t:
//
//	A.ComposeBefore(B) = T_A x T_B
func (t *AffineTransformation) ComposeBefore(other *AffineTransformation) *AffineTransformation {
	mp00 := t.m00*other.m00 + t.m01*other.m10
	mp01 := t.m00*other.m01 + t.m01*other.m11
	mp02 := t.m00*other.m02 + t.m01*other.m12 + t.m02
	mp10 := t.m10*other.m00 + t.m11*other.m10
	mp11 := t.m10*other.m01 + t.m11*other.m11
	mp12 := t.m10*other.m02 + t.m11*other.m12 + t.m12
	t.m00, t.m01, t.m02 = mp00, mp01, mp02
	t.m10, t.m11, t.m12 = mp10, mp11, mp12
	return t
}

// TransformXY applies t to the given coordinate, returning a new XY.
func (t *AffineTransformation) TransformXY(p XY) XY {
	return XY{
		X: t.m00*p.X + t.m01*p.Y + t.m02,
		Y: t.m10*p.X + t.m11*p.Y + t.m12,
	}
}

// Transform applies t to every coordinate of g, returning a new
// Geometry of the same shape. Z and M ordinates are preserved
// unchanged. The returned Geometry shares the input's CRS.
func (t *AffineTransformation) Transform(g Geometry) Geometry {
	if g == nil {
		return nil
	}
	return Edit(g, t.TransformXY)
}

// IsIdentity reports whether t is the identity transformation.
func (t *AffineTransformation) IsIdentity() bool {
	return t.m00 == 1 && t.m01 == 0 && t.m02 == 0 &&
		t.m10 == 0 && t.m11 == 1 && t.m12 == 0
}

// Equals reports whether t and other have identical matrix entries.
func (t *AffineTransformation) Equals(other *AffineTransformation) bool {
	if other == nil {
		return false
	}
	return t.m00 == other.m00 && t.m01 == other.m01 && t.m02 == other.m02 &&
		t.m10 == other.m10 && t.m11 == other.m11 && t.m12 == other.m12
}

// Clone returns a deep copy of t.
func (t *AffineTransformation) Clone() *AffineTransformation {
	c := *t
	return &c
}

// String returns a textual representation of t.
func (t *AffineTransformation) String() string {
	return fmt.Sprintf("AffineTransformation[[%v, %v, %v], [%v, %v, %v]]",
		t.m00, t.m01, t.m02, t.m10, t.m11, t.m12)
}
