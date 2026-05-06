package shape

import (
	"testing"

	"github.com/exergy-dev/go-topology-suite/geom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGridPointsCount(t *testing.T) {
	env := geom.Envelope{MinX: 0, MinY: 0, MaxX: 10, MaxY: 10}
	// n=10 -> grid 4x4 = 16 points (next perfect square >= 10).
	pts := GridPoints(10, env, 1.0, WithSeed(1))
	require.Equal(t, 16, len(pts))
}

func TestGridPointsExactSquare(t *testing.T) {
	env := geom.Envelope{MinX: 0, MinY: 0, MaxX: 8, MaxY: 8}
	pts := GridPoints(25, env, 0.5, WithSeed(1))
	require.Equal(t, 25, len(pts))
}

func TestGridPointsContained(t *testing.T) {
	env := geom.Envelope{MinX: -5, MinY: -5, MaxX: 5, MaxY: 5}
	pts := GridPoints(36, env, 1.0, WithSeed(99))
	for i, p := range pts {
		require.Truef(t, env.ContainsXY(p), "pt[%d]=%v not in env", i, p)
	}
}

func TestGridPointsZeroJitter(t *testing.T) {
	// jitterFraction=0 means each point sits at a deterministic cell origin
	// (the centre when the gutter consumes the whole cell).
	env := geom.Envelope{MinX: 0, MinY: 0, MaxX: 10, MaxY: 10}
	a := GridPoints(9, env, 0)
	b := GridPoints(9, env, 0)
	require.Equal(t, len(a), len(b), "len mismatch")
	for i := range a {
		assert.Equalf(t, a[i], b[i], "zero jitter not deterministic at i=%d", i)
	}
}

func TestGridPointsDeterministicWithSeed(t *testing.T) {
	env := geom.Envelope{MinX: 0, MinY: 0, MaxX: 1, MaxY: 1}
	a := GridPoints(16, env, 0.5, WithSeed(42))
	b := GridPoints(16, env, 0.5, WithSeed(42))
	for i := range a {
		assert.Equalf(t, a[i], b[i], "not deterministic at i=%d", i)
	}
}
