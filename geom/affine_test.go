package geom

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func approxEqualXY(a, b XY, tol float64) bool {
	return math.Abs(a.X-b.X) <= tol && math.Abs(a.Y-b.Y) <= tol
}

func TestAffineIdentity(t *testing.T) {
	id := NewAffineTransformation()
	assert.True(t, id.IsIdentity(), "NewAffineTransformation() must be identity, got %s", id)
	for _, p := range []XY{{1, 2}, {-3, 5}, {0, 0}, {1.25, 1e6}} {
		got := id.TransformXY(p)
		assert.Equal(t, p, got, "identity.Transform(%v)", p)
	}
}

func TestAffineTranslation(t *testing.T) {
	tr := AffineTranslation(10, -5)
	got := tr.TransformXY(XY{1, 2})
	want := XY{11, -3}
	assert.Equal(t, want, got, "translate")
}

func TestAffineRotationOrigin(t *testing.T) {
	// 90° CCW rotation about origin: (1,0) -> (0,1)
	r := AffineRotation(math.Pi / 2)
	got := r.TransformXY(XY{1, 0})
	want := XY{0, 1}
	assert.True(t, approxEqualXY(got, want, 1e-12), "rotation: got %v, want %v", got, want)
}

func TestAffineRotationAround(t *testing.T) {
	// 180° around (1,1): (2,2) -> (0,0)
	r := AffineRotationAround(math.Pi, 1, 1)
	got := r.TransformXY(XY{2, 2})
	want := XY{0, 0}
	assert.True(t, approxEqualXY(got, want, 1e-9), "rotation around point: got %v, want %v", got, want)
}

func TestAffineScale(t *testing.T) {
	s := AffineScale(2, 3)
	got := s.TransformXY(XY{1, 2})
	want := XY{2, 6}
	assert.Equal(t, want, got, "scale")
}

func TestAffineComposition(t *testing.T) {
	// Compose translate(10,0) THEN rotate(90°) about origin.
	// (0,0) -> translate -> (10,0) -> rotate -> (0,10).
	tr := AffineTranslation(10, 0)
	tr.Compose(AffineRotation(math.Pi / 2))
	got := tr.TransformXY(XY{0, 0})
	want := XY{0, 10}
	assert.True(t, approxEqualXY(got, want, 1e-12), "compose translate+rotate: got %v, want %v", got, want)
}

func TestAffineCompositionScale(t *testing.T) {
	// Translate(1,2) then scale(2,3): (0,0)->(1,2)->(2,6).
	tr := AffineTranslation(1, 2)
	tr.Compose(AffineScale(2, 3))
	got := tr.TransformXY(XY{0, 0})
	want := XY{2, 6}
	assert.True(t, approxEqualXY(got, want, 1e-12), "compose translate+scale: got %v, want %v", got, want)
}

func TestAffineInverseRoundTrip(t *testing.T) {
	tr := AffineTranslation(7, -3)
	tr.Compose(AffineRotation(0.4))
	tr.Compose(AffineScale(2, 0.5))
	inv, err := tr.Inverse()
	require.NoError(t, err)
	for _, p := range []XY{{1, 2}, {-3, 5}, {1e3, -1e2}} {
		mapped := tr.TransformXY(p)
		back := inv.TransformXY(mapped)
		assert.True(t, approxEqualXY(p, back, 1e-9), "round-trip: %v -> %v -> %v", p, mapped, back)
	}
}

func TestAffineInverseSingular(t *testing.T) {
	t1 := AffineScale(0, 1)
	_, err := t1.Inverse()
	assert.Error(t, err, "expected non-invertible error for scale(0,1)")
}

func TestAffineTransformGeometry(t *testing.T) {
	pts := []XY{{0, 0}, {1, 0}, {1, 1}, {0, 1}, {0, 0}}
	poly := NewPolygon(nil, pts)
	tr := AffineTranslation(10, 20)
	out := tr.Transform(poly)
	op, ok := out.(*Polygon)
	require.True(t, ok, "Transform returned %T, want *Polygon", out)
	ring := op.Ring(0)
	wantFirst := XY{10, 20}
	assert.True(t, approxEqualXY(ring[0], wantFirst, 1e-12), "ring[0] = %v, want %v", ring[0], wantFirst)
}

func TestAffineEqualsAndClone(t *testing.T) {
	a := AffineTranslation(1, 2)
	b := a.Clone()
	assert.True(t, a.Equals(b), "clone should equal original")
	b.Translate(1, 0)
	assert.False(t, a.Equals(b), "post-modification should not equal")
}

func TestAffineReflection(t *testing.T) {
	// Reflection about the x-axis: (x,y) -> (x,-y).
	r := AffineReflection(0, 0, 1, 0)
	got := r.TransformXY(XY{3, 4})
	want := XY{3, -4}
	assert.True(t, approxEqualXY(got, want, 1e-9), "reflection: got %v, want %v", got, want)
}

func TestAffineReflectionVectorXY(t *testing.T) {
	// Reflection about y=x line: (x,y) -> (y,x).
	r := AffineReflectionVector(1, 1)
	got := r.TransformXY(XY{3, 4})
	want := XY{4, 3}
	assert.True(t, approxEqualXY(got, want, 1e-12), "reflection y=x: got %v, want %v", got, want)
}
