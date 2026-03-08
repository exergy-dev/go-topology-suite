package geom_test

import (
	"math"
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Coordinate constructors and methods
// ---------------------------------------------------------------------------

func TestNewCoordinateZM(t *testing.T) {
	t.Run("basic values", func(t *testing.T) {
		c := geom.NewCoordinateZM(1, 2, 3, 4)
		assert.Equal(t, 1.0, c.X)
		assert.Equal(t, 2.0, c.Y)
		assert.Equal(t, 3.0, c.Z)
		assert.Equal(t, 4.0, c.M)
		assert.True(t, c.HasZ())
		assert.True(t, c.HasM())
	})

	t.Run("zero values", func(t *testing.T) {
		c := geom.NewCoordinateZM(0, 0, 0, 0)
		assert.True(t, c.HasZ(), "Z=0 is a valid Z value, not absent")
		assert.True(t, c.HasM(), "M=0 is a valid M value, not absent")
		assert.Equal(t, 0.0, c.Z)
		assert.Equal(t, 0.0, c.M)
	})

	t.Run("negative values", func(t *testing.T) {
		c := geom.NewCoordinateZM(-10, -20, -30, -40)
		assert.Equal(t, -10.0, c.X)
		assert.Equal(t, -20.0, c.Y)
		assert.Equal(t, -30.0, c.Z)
		assert.Equal(t, -40.0, c.M)
	})
}

func TestNewCoordinateM(t *testing.T) {
	t.Run("basic values", func(t *testing.T) {
		c := geom.NewCoordinateM(5, 6, 7)
		assert.Equal(t, 5.0, c.X)
		assert.Equal(t, 6.0, c.Y)
		assert.False(t, c.HasZ(), "M-only coordinate should not have Z")
		assert.True(t, c.HasM())
		assert.Equal(t, 7.0, c.M)
	})

	t.Run("zero M value", func(t *testing.T) {
		c := geom.NewCoordinateM(1, 2, 0)
		assert.True(t, c.HasM(), "M=0 is a valid M value")
		assert.Equal(t, 0.0, c.M)
	})
}

func TestNewCoordinateNaN(t *testing.T) {
	t.Run("all NaN", func(t *testing.T) {
		c := geom.NewCoordinateNaN()
		assert.True(t, math.IsNaN(c.X))
		assert.True(t, math.IsNaN(c.Y))
		assert.False(t, c.HasZ(), "NaN Z should be absent")
		assert.False(t, c.HasM(), "NaN M should be absent")
	})

	t.Run("IsNaN returns true", func(t *testing.T) {
		c := geom.NewCoordinateNaN()
		assert.True(t, c.IsNaN())
	})
}

func TestCoordinate_String(t *testing.T) {
	t.Run("2D", func(t *testing.T) {
		c := geom.NewCoordinate(1, 2)
		assert.Equal(t, "(1, 2)", c.String())
	})

	t.Run("3D with Z", func(t *testing.T) {
		c := geom.NewCoordinateZ(1, 2, 3)
		assert.Equal(t, "(1, 2, 3)", c.String())
	})

	t.Run("with M only", func(t *testing.T) {
		c := geom.NewCoordinateM(1, 2, 4)
		assert.Equal(t, "(1, 2, M=4)", c.String())
	})

	t.Run("with Z and M", func(t *testing.T) {
		c := geom.NewCoordinateZM(1, 2, 3, 4)
		assert.Equal(t, "(1, 2, 3, 4)", c.String())
	})

	t.Run("fractional values", func(t *testing.T) {
		c := geom.NewCoordinate(1.5, 2.5)
		assert.Equal(t, "(1.5, 2.5)", c.String())
	})
}

func TestCoordinate_Equals3D(t *testing.T) {
	eps := 1e-10

	t.Run("both have Z, same", func(t *testing.T) {
		c1 := geom.NewCoordinateZ(1, 2, 3)
		c2 := geom.NewCoordinateZ(1, 2, 3)
		assert.True(t, c1.Equals(c2, eps))
	})

	t.Run("both have Z, different Z", func(t *testing.T) {
		c1 := geom.NewCoordinateZ(1, 2, 3)
		c2 := geom.NewCoordinateZ(1, 2, 5)
		assert.False(t, c1.Equals(c2, eps))
	})

	t.Run("one has Z, other does not", func(t *testing.T) {
		c1 := geom.NewCoordinateZ(1, 2, 3)
		c2 := geom.NewCoordinate(1, 2)
		assert.False(t, c1.Equals(c2, eps))
	})

	t.Run("neither has Z", func(t *testing.T) {
		c1 := geom.NewCoordinate(1, 2)
		c2 := geom.NewCoordinate(1, 2)
		assert.True(t, c1.Equals(c2, eps))
	})

	t.Run("both have M, same", func(t *testing.T) {
		c1 := geom.NewCoordinateM(1, 2, 10)
		c2 := geom.NewCoordinateM(1, 2, 10)
		assert.True(t, c1.Equals(c2, eps))
	})

	t.Run("both have M, different M", func(t *testing.T) {
		c1 := geom.NewCoordinateM(1, 2, 10)
		c2 := geom.NewCoordinateM(1, 2, 20)
		assert.False(t, c1.Equals(c2, eps))
	})

	t.Run("one has M, other does not", func(t *testing.T) {
		c1 := geom.NewCoordinateM(1, 2, 10)
		c2 := geom.NewCoordinate(1, 2)
		assert.False(t, c1.Equals(c2, eps))
	})

	t.Run("full ZM equals", func(t *testing.T) {
		c1 := geom.NewCoordinateZM(1, 2, 3, 4)
		c2 := geom.NewCoordinateZM(1, 2, 3, 4)
		assert.True(t, c1.Equals(c2, eps))
	})

	t.Run("full ZM differs in Z", func(t *testing.T) {
		c1 := geom.NewCoordinateZM(1, 2, 3, 4)
		c2 := geom.NewCoordinateZM(1, 2, 99, 4)
		assert.False(t, c1.Equals(c2, eps))
	})

	t.Run("full ZM differs in M", func(t *testing.T) {
		c1 := geom.NewCoordinateZM(1, 2, 3, 4)
		c2 := geom.NewCoordinateZM(1, 2, 3, 99)
		assert.False(t, c1.Equals(c2, eps))
	})

	t.Run("different XY fails even with matching Z", func(t *testing.T) {
		c1 := geom.NewCoordinateZ(1, 2, 3)
		c2 := geom.NewCoordinateZ(9, 9, 3)
		assert.False(t, c1.Equals(c2, eps))
	})

	t.Run("within epsilon", func(t *testing.T) {
		c1 := geom.NewCoordinateZ(1, 2, 3)
		c2 := geom.NewCoordinateZ(1, 2, 3.0000000001)
		assert.True(t, c1.Equals(c2, 1e-9))
	})
}

func TestCoordinate_Distance3D(t *testing.T) {
	t.Run("both have Z", func(t *testing.T) {
		c1 := geom.NewCoordinateZ(0, 0, 0)
		c2 := geom.NewCoordinateZ(1, 2, 2)
		// sqrt(1+4+4) = 3
		assert.InDelta(t, 3.0, c1.Distance3D(c2), 1e-10)
	})

	t.Run("first lacks Z falls back to 2D", func(t *testing.T) {
		c1 := geom.NewCoordinate(0, 0)
		c2 := geom.NewCoordinateZ(3, 4, 100)
		// Should use 2D distance: sqrt(9+16) = 5
		assert.InDelta(t, 5.0, c1.Distance3D(c2), 1e-10)
	})

	t.Run("second lacks Z falls back to 2D", func(t *testing.T) {
		c1 := geom.NewCoordinateZ(3, 4, 100)
		c2 := geom.NewCoordinate(0, 0)
		assert.InDelta(t, 5.0, c1.Distance3D(c2), 1e-10)
	})

	t.Run("neither has Z equals 2D distance", func(t *testing.T) {
		c1 := geom.NewCoordinate(0, 0)
		c2 := geom.NewCoordinate(3, 4)
		assert.InDelta(t, 5.0, c1.Distance3D(c2), 1e-10)
	})

	t.Run("same point 3D distance is zero", func(t *testing.T) {
		c := geom.NewCoordinateZ(7, 8, 9)
		assert.Equal(t, 0.0, c.Distance3D(c))
	})
}

func TestCoordinate_XY(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		c := geom.NewCoordinate(3.5, 4.5)
		xy := c.XY()
		assert.Equal(t, 3.5, xy.X)
		assert.Equal(t, 4.5, xy.Y)
	})

	t.Run("usable as map key", func(t *testing.T) {
		m := make(map[geom.CoordinateXY]int)
		c1 := geom.NewCoordinate(1, 2)
		c2 := geom.NewCoordinate(1, 2)
		c3 := geom.NewCoordinate(3, 4)

		m[c1.XY()] = 10
		m[c2.XY()] = 20 // overwrites c1 entry

		assert.Equal(t, 20, m[c1.XY()])
		_, ok := m[c3.XY()]
		assert.False(t, ok)
	})

	t.Run("strips Z and M", func(t *testing.T) {
		c := geom.NewCoordinateZM(1, 2, 3, 4)
		xy := c.XY()
		assert.Equal(t, 1.0, xy.X)
		assert.Equal(t, 2.0, xy.Y)
	})
}

func TestCoordinate_GetZ(t *testing.T) {
	t.Run("has Z returns value", func(t *testing.T) {
		c := geom.NewCoordinateZ(1, 2, 42)
		assert.Equal(t, 42.0, c.GetZ())
	})

	t.Run("no Z returns zero", func(t *testing.T) {
		c := geom.NewCoordinate(1, 2)
		assert.Equal(t, 0.0, c.GetZ())
	})
}

func TestCoordinate_GetM(t *testing.T) {
	t.Run("has M returns value", func(t *testing.T) {
		c := geom.NewCoordinateM(1, 2, 99)
		assert.Equal(t, 99.0, c.GetM())
	})

	t.Run("no M returns zero", func(t *testing.T) {
		c := geom.NewCoordinate(1, 2)
		assert.Equal(t, 0.0, c.GetM())
	})
}

// ---------------------------------------------------------------------------
// CoordinateSequence
// ---------------------------------------------------------------------------

func TestNewCoordinateSequence(t *testing.T) {
	t.Run("from multiple coordinates", func(t *testing.T) {
		c1 := geom.NewCoordinate(1, 2)
		c2 := geom.NewCoordinate(3, 4)
		seq := geom.NewCoordinateSequence(c1, c2)
		require.Equal(t, 2, seq.Len())
		assert.Equal(t, 1.0, seq.Get(0).X)
		assert.Equal(t, 3.0, seq.Get(1).X)
	})

	t.Run("empty variadic creates empty sequence", func(t *testing.T) {
		seq := geom.NewCoordinateSequence()
		assert.Equal(t, 0, seq.Len())
		assert.True(t, seq.IsEmpty())
	})

	t.Run("single coordinate", func(t *testing.T) {
		c := geom.NewCoordinate(5, 6)
		seq := geom.NewCoordinateSequence(c)
		assert.Equal(t, 1, seq.Len())
		assert.Equal(t, 5.0, seq.Get(0).X)
	})

	t.Run("creates a defensive copy", func(t *testing.T) {
		c := geom.NewCoordinate(1, 2)
		seq := geom.NewCoordinateSequence(c)
		c.X = 999
		assert.Equal(t, 1.0, seq.Get(0).X, "mutating the original should not affect the sequence")
	})
}

func TestCoordinateSequence_Len(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		seq := geom.NewCoordinateSequence()
		assert.Equal(t, 0, seq.Len())
	})

	t.Run("non-empty", func(t *testing.T) {
		seq := mustCoordsXY(1, 2, 3, 4, 5, 6)
		assert.Equal(t, 3, seq.Len())
	})
}

func TestCoordinateSequence_IsEmpty(t *testing.T) {
	t.Run("empty sequence", func(t *testing.T) {
		seq := geom.NewCoordinateSequence()
		assert.True(t, seq.IsEmpty())
	})

	t.Run("non-empty sequence", func(t *testing.T) {
		seq := mustCoordsXY(0, 0)
		assert.False(t, seq.IsEmpty())
	})

	t.Run("nil sequence", func(t *testing.T) {
		var seq geom.CoordinateSequence
		assert.True(t, seq.IsEmpty())
	})
}

func TestCoordinateSequence_Get(t *testing.T) {
	seq := mustCoordsXY(10, 20, 30, 40, 50, 60)

	t.Run("first element", func(t *testing.T) {
		c := seq.Get(0)
		assert.Equal(t, 10.0, c.X)
		assert.Equal(t, 20.0, c.Y)
	})

	t.Run("middle element", func(t *testing.T) {
		c := seq.Get(1)
		assert.Equal(t, 30.0, c.X)
		assert.Equal(t, 40.0, c.Y)
	})

	t.Run("last element", func(t *testing.T) {
		c := seq.Get(2)
		assert.Equal(t, 50.0, c.X)
		assert.Equal(t, 60.0, c.Y)
	})
}

func TestCoordinateSequence_RemoveRepeatedPoints(t *testing.T) {
	eps := 1e-10

	t.Run("no duplicates", func(t *testing.T) {
		seq := mustCoordsXY(0, 0, 1, 1, 2, 2)
		result := seq.RemoveRepeatedPoints(eps)
		assert.Equal(t, 3, result.Len())
	})

	t.Run("consecutive duplicates", func(t *testing.T) {
		seq := mustCoordsXY(0, 0, 0, 0, 1, 1, 1, 1, 2, 2)
		result := seq.RemoveRepeatedPoints(eps)
		assert.Equal(t, 3, result.Len())
		assert.Equal(t, 0.0, result.Get(0).X)
		assert.Equal(t, 1.0, result.Get(1).X)
		assert.Equal(t, 2.0, result.Get(2).X)
	})

	t.Run("all same point", func(t *testing.T) {
		seq := mustCoordsXY(5, 5, 5, 5, 5, 5)
		result := seq.RemoveRepeatedPoints(eps)
		assert.Equal(t, 1, result.Len())
	})

	t.Run("single point", func(t *testing.T) {
		seq := mustCoordsXY(1, 2)
		result := seq.RemoveRepeatedPoints(eps)
		assert.Equal(t, 1, result.Len())
	})

	t.Run("empty sequence", func(t *testing.T) {
		seq := geom.NewCoordinateSequence()
		result := seq.RemoveRepeatedPoints(eps)
		assert.Equal(t, 0, result.Len())
	})

	t.Run("near-duplicate within epsilon", func(t *testing.T) {
		c1 := geom.NewCoordinate(0, 0)
		c2 := geom.NewCoordinate(1e-12, 1e-12)
		c3 := geom.NewCoordinate(1, 1)
		seq := geom.NewCoordinateSequence(c1, c2, c3)
		result := seq.RemoveRepeatedPoints(1e-10)
		assert.Equal(t, 2, result.Len(), "near-duplicate should be removed")
	})

	t.Run("non-consecutive duplicates are kept", func(t *testing.T) {
		seq := mustCoordsXY(0, 0, 1, 1, 0, 0)
		result := seq.RemoveRepeatedPoints(eps)
		assert.Equal(t, 3, result.Len(), "non-consecutive duplicates should remain")
	})

	t.Run("result is a deep copy", func(t *testing.T) {
		seq := mustCoordsXY(0, 0, 1, 1)
		result := seq.RemoveRepeatedPoints(eps)
		result[0].X = 999
		assert.Equal(t, 0.0, seq.Get(0).X, "original should not be mutated")
	})
}

func TestCoordinateSequence_SubSequence(t *testing.T) {
	seq := mustCoordsXY(0, 0, 1, 1, 2, 2, 3, 3, 4, 4)

	t.Run("normal range", func(t *testing.T) {
		sub := seq.SubSequence(1, 3)
		require.Equal(t, 2, sub.Len())
		assert.Equal(t, 1.0, sub.Get(0).X)
		assert.Equal(t, 2.0, sub.Get(1).X)
	})

	t.Run("full range", func(t *testing.T) {
		sub := seq.SubSequence(0, 5)
		assert.Equal(t, 5, sub.Len())
	})

	t.Run("start equals end returns empty", func(t *testing.T) {
		sub := seq.SubSequence(2, 2)
		assert.Equal(t, 0, sub.Len())
	})

	t.Run("start greater than end returns empty", func(t *testing.T) {
		sub := seq.SubSequence(3, 1)
		assert.Equal(t, 0, sub.Len())
	})

	t.Run("negative start clamped to 0", func(t *testing.T) {
		sub := seq.SubSequence(-5, 2)
		assert.Equal(t, 2, sub.Len())
		assert.Equal(t, 0.0, sub.Get(0).X)
	})

	t.Run("end beyond length clamped", func(t *testing.T) {
		sub := seq.SubSequence(3, 100)
		assert.Equal(t, 2, sub.Len())
		assert.Equal(t, 3.0, sub.Get(0).X)
		assert.Equal(t, 4.0, sub.Get(1).X)
	})

	t.Run("result is a deep copy", func(t *testing.T) {
		sub := seq.SubSequence(0, 2)
		sub[0].X = 999
		assert.Equal(t, 0.0, seq.Get(0).X, "original should not be mutated")
	})

	t.Run("empty sequence", func(t *testing.T) {
		empty := geom.NewCoordinateSequence()
		sub := empty.SubSequence(0, 0)
		assert.Equal(t, 0, sub.Len())
	})
}

// ---------------------------------------------------------------------------
// Envelope
// ---------------------------------------------------------------------------

func TestNewEnvelopeFromCoords(t *testing.T) {
	t.Run("normal order", func(t *testing.T) {
		c1 := geom.NewCoordinate(1, 2)
		c2 := geom.NewCoordinate(10, 20)
		env := geom.NewEnvelopeFromCoords(c1, c2)
		assert.Equal(t, 1.0, env.MinX)
		assert.Equal(t, 2.0, env.MinY)
		assert.Equal(t, 10.0, env.MaxX)
		assert.Equal(t, 20.0, env.MaxY)
	})

	t.Run("reversed order auto-swaps", func(t *testing.T) {
		c1 := geom.NewCoordinate(10, 20)
		c2 := geom.NewCoordinate(1, 2)
		env := geom.NewEnvelopeFromCoords(c1, c2)
		assert.Equal(t, 1.0, env.MinX)
		assert.Equal(t, 2.0, env.MinY)
		assert.Equal(t, 10.0, env.MaxX)
		assert.Equal(t, 20.0, env.MaxY)
	})

	t.Run("same coordinate produces point envelope", func(t *testing.T) {
		c := geom.NewCoordinate(5, 5)
		env := geom.NewEnvelopeFromCoords(c, c)
		assert.Equal(t, 0.0, env.Width())
		assert.Equal(t, 0.0, env.Height())
		assert.False(t, env.IsNull())
	})
}

func TestEnvelope_CloneEmpty(t *testing.T) {
	t.Run("clone of empty envelope is also empty", func(t *testing.T) {
		env := geom.NewEnvelopeEmpty()
		clone := env.Clone()
		require.NotNil(t, clone)
		assert.True(t, clone.IsNull())
	})

	t.Run("clone of nil returns nil", func(t *testing.T) {
		var env *geom.Envelope
		clone := env.Clone()
		assert.Nil(t, clone)
	})
}

func TestEnvelope_String(t *testing.T) {
	t.Run("non-empty", func(t *testing.T) {
		env := geom.NewEnvelope(0, 0, 10, 20)
		assert.Equal(t, "Envelope(0, 0, 10, 20)", env.String())
	})

	t.Run("empty", func(t *testing.T) {
		env := geom.NewEnvelopeEmpty()
		assert.Equal(t, "Envelope(EMPTY)", env.String())
	})

	t.Run("nil", func(t *testing.T) {
		var env *geom.Envelope
		assert.Equal(t, "Envelope(EMPTY)", env.String())
	})
}

func TestEnvelope_MinExtent(t *testing.T) {
	t.Run("wider than tall", func(t *testing.T) {
		env := geom.NewEnvelope(0, 0, 20, 5)
		assert.Equal(t, 5.0, env.MinExtent())
	})

	t.Run("taller than wide", func(t *testing.T) {
		env := geom.NewEnvelope(0, 0, 3, 10)
		assert.Equal(t, 3.0, env.MinExtent())
	})

	t.Run("square", func(t *testing.T) {
		env := geom.NewEnvelope(0, 0, 7, 7)
		assert.Equal(t, 7.0, env.MinExtent())
	})

	t.Run("empty returns 0", func(t *testing.T) {
		env := geom.NewEnvelopeEmpty()
		assert.Equal(t, 0.0, env.MinExtent())
	})
}

func TestEnvelope_MaxExtent(t *testing.T) {
	t.Run("wider than tall", func(t *testing.T) {
		env := geom.NewEnvelope(0, 0, 20, 5)
		assert.Equal(t, 20.0, env.MaxExtent())
	})

	t.Run("taller than wide", func(t *testing.T) {
		env := geom.NewEnvelope(0, 0, 3, 10)
		assert.Equal(t, 10.0, env.MaxExtent())
	})

	t.Run("square", func(t *testing.T) {
		env := geom.NewEnvelope(0, 0, 7, 7)
		assert.Equal(t, 7.0, env.MaxExtent())
	})

	t.Run("empty returns 0", func(t *testing.T) {
		env := geom.NewEnvelopeEmpty()
		assert.Equal(t, 0.0, env.MaxExtent())
	})
}

func TestEnvelope_ExpandToIncludeCoord(t *testing.T) {
	t.Run("expand non-empty", func(t *testing.T) {
		env := geom.NewEnvelope(0, 0, 10, 10)
		env.ExpandToIncludeCoord(geom.NewCoordinate(20, 30))
		assert.Equal(t, 0.0, env.MinX)
		assert.Equal(t, 0.0, env.MinY)
		assert.Equal(t, 20.0, env.MaxX)
		assert.Equal(t, 30.0, env.MaxY)
	})

	t.Run("expand with negative coordinates", func(t *testing.T) {
		env := geom.NewEnvelope(0, 0, 10, 10)
		env.ExpandToIncludeCoord(geom.NewCoordinate(-5, -3))
		assert.Equal(t, -5.0, env.MinX)
		assert.Equal(t, -3.0, env.MinY)
	})

	t.Run("expand empty envelope", func(t *testing.T) {
		env := geom.NewEnvelopeEmpty()
		env.ExpandToIncludeCoord(geom.NewCoordinate(5, 10))
		assert.False(t, env.IsNull())
		assert.Equal(t, 5.0, env.MinX)
		assert.Equal(t, 10.0, env.MinY)
		assert.Equal(t, 5.0, env.MaxX)
		assert.Equal(t, 10.0, env.MaxY)
	})

	t.Run("coordinate already inside no change", func(t *testing.T) {
		env := geom.NewEnvelope(0, 0, 10, 10)
		env.ExpandToIncludeCoord(geom.NewCoordinate(5, 5))
		assert.Equal(t, 0.0, env.MinX)
		assert.Equal(t, 10.0, env.MaxX)
	})
}

func TestEnvelope_ExpandBy(t *testing.T) {
	t.Run("positive distance", func(t *testing.T) {
		env := geom.NewEnvelope(5, 5, 15, 15)
		env.ExpandBy(3)
		assert.Equal(t, 2.0, env.MinX)
		assert.Equal(t, 2.0, env.MinY)
		assert.Equal(t, 18.0, env.MaxX)
		assert.Equal(t, 18.0, env.MaxY)
	})

	t.Run("negative distance collapses to empty", func(t *testing.T) {
		env := geom.NewEnvelope(0, 0, 4, 4)
		env.ExpandBy(-5) // shrinks more than half the width/height
		assert.True(t, env.IsNull(), "should collapse to empty when shrunk too much")
	})

	t.Run("negative distance just within bounds", func(t *testing.T) {
		env := geom.NewEnvelope(0, 0, 10, 10)
		env.ExpandBy(-2)
		assert.False(t, env.IsNull())
		assert.Equal(t, 2.0, env.MinX)
		assert.Equal(t, 2.0, env.MinY)
		assert.Equal(t, 8.0, env.MaxX)
		assert.Equal(t, 8.0, env.MaxY)
	})

	t.Run("empty envelope stays empty", func(t *testing.T) {
		env := geom.NewEnvelopeEmpty()
		env.ExpandBy(10)
		assert.True(t, env.IsNull())
	})
}

func TestEnvelope_ExpandByXY(t *testing.T) {
	t.Run("different X and Y deltas", func(t *testing.T) {
		env := geom.NewEnvelope(10, 10, 20, 20)
		env.ExpandByXY(5, 2)
		assert.Equal(t, 5.0, env.MinX)
		assert.Equal(t, 8.0, env.MinY)
		assert.Equal(t, 25.0, env.MaxX)
		assert.Equal(t, 22.0, env.MaxY)
	})

	t.Run("collapse X only", func(t *testing.T) {
		env := geom.NewEnvelope(0, 0, 2, 100)
		env.ExpandByXY(-5, 0)
		assert.True(t, env.IsNull())
	})

	t.Run("collapse Y only", func(t *testing.T) {
		env := geom.NewEnvelope(0, 0, 100, 2)
		env.ExpandByXY(0, -5)
		assert.True(t, env.IsNull())
	})

	t.Run("empty envelope stays empty", func(t *testing.T) {
		env := geom.NewEnvelopeEmpty()
		env.ExpandByXY(10, 20)
		assert.True(t, env.IsNull())
	})
}

func TestEnvelope_ContainsXY_False(t *testing.T) {
	env := geom.NewEnvelope(0, 0, 10, 10)

	t.Run("outside right", func(t *testing.T) {
		assert.False(t, env.ContainsXY(11, 5))
	})

	t.Run("outside left", func(t *testing.T) {
		assert.False(t, env.ContainsXY(-1, 5))
	})

	t.Run("outside above", func(t *testing.T) {
		assert.False(t, env.ContainsXY(5, 11))
	})

	t.Run("outside below", func(t *testing.T) {
		assert.False(t, env.ContainsXY(5, -1))
	})

	t.Run("on boundary is contained", func(t *testing.T) {
		assert.True(t, env.ContainsXY(0, 0))
		assert.True(t, env.ContainsXY(10, 10))
		assert.True(t, env.ContainsXY(0, 10))
		assert.True(t, env.ContainsXY(10, 0))
	})

	t.Run("empty envelope contains nothing", func(t *testing.T) {
		empty := geom.NewEnvelopeEmpty()
		assert.False(t, empty.ContainsXY(0, 0))
	})
}

func TestEnvelope_ContainsEnvelope_False(t *testing.T) {
	env := geom.NewEnvelope(0, 0, 10, 10)

	t.Run("partially overlapping", func(t *testing.T) {
		other := geom.NewEnvelope(5, 5, 15, 15)
		assert.False(t, env.ContainsEnvelope(other))
	})

	t.Run("completely outside", func(t *testing.T) {
		other := geom.NewEnvelope(20, 20, 30, 30)
		assert.False(t, env.ContainsEnvelope(other))
	})

	t.Run("containing envelope does not fit in smaller", func(t *testing.T) {
		big := geom.NewEnvelope(-10, -10, 100, 100)
		assert.False(t, env.ContainsEnvelope(big))
	})

	t.Run("contained envelope true", func(t *testing.T) {
		inner := geom.NewEnvelope(2, 2, 8, 8)
		assert.True(t, env.ContainsEnvelope(inner))
	})

	t.Run("empty envelope cannot contain", func(t *testing.T) {
		empty := geom.NewEnvelopeEmpty()
		assert.False(t, empty.ContainsEnvelope(env))
	})

	t.Run("cannot contain empty", func(t *testing.T) {
		empty := geom.NewEnvelopeEmpty()
		assert.False(t, env.ContainsEnvelope(empty))
	})
}

func TestEnvelope_Covers(t *testing.T) {
	env := geom.NewEnvelope(0, 0, 10, 10)

	t.Run("covers interior point", func(t *testing.T) {
		assert.True(t, env.Covers(geom.NewCoordinate(5, 5)))
	})

	t.Run("covers boundary point", func(t *testing.T) {
		assert.True(t, env.Covers(geom.NewCoordinate(0, 0)))
		assert.True(t, env.Covers(geom.NewCoordinate(10, 10)))
	})

	t.Run("does not cover exterior point", func(t *testing.T) {
		assert.False(t, env.Covers(geom.NewCoordinate(15, 5)))
	})
}

func TestEnvelope_CoversXY(t *testing.T) {
	env := geom.NewEnvelope(0, 0, 10, 10)

	t.Run("covers interior", func(t *testing.T) {
		assert.True(t, env.CoversXY(5, 5))
	})

	t.Run("does not cover exterior", func(t *testing.T) {
		assert.False(t, env.CoversXY(15, 15))
	})

	t.Run("covers boundary", func(t *testing.T) {
		assert.True(t, env.CoversXY(0, 0))
	})
}

func TestEnvelope_CoversEnvelope(t *testing.T) {
	env := geom.NewEnvelope(0, 0, 10, 10)

	t.Run("covers smaller", func(t *testing.T) {
		inner := geom.NewEnvelope(2, 2, 8, 8)
		assert.True(t, env.CoversEnvelope(inner))
	})

	t.Run("covers itself", func(t *testing.T) {
		same := geom.NewEnvelope(0, 0, 10, 10)
		assert.True(t, env.CoversEnvelope(same))
	})

	t.Run("does not cover larger", func(t *testing.T) {
		bigger := geom.NewEnvelope(-1, -1, 11, 11)
		assert.False(t, env.CoversEnvelope(bigger))
	})

	t.Run("does not cover partial overlap", func(t *testing.T) {
		partial := geom.NewEnvelope(5, 5, 15, 15)
		assert.False(t, env.CoversEnvelope(partial))
	})
}

func TestEnvelope_IntersectsCoord(t *testing.T) {
	env := geom.NewEnvelope(0, 0, 10, 10)

	t.Run("interior point intersects", func(t *testing.T) {
		assert.True(t, env.IntersectsCoord(geom.NewCoordinate(5, 5)))
	})

	t.Run("boundary point intersects", func(t *testing.T) {
		assert.True(t, env.IntersectsCoord(geom.NewCoordinate(0, 0)))
	})

	t.Run("exterior point does not intersect", func(t *testing.T) {
		assert.False(t, env.IntersectsCoord(geom.NewCoordinate(15, 15)))
	})
}

func TestEnvelope_IntersectsXY(t *testing.T) {
	env := geom.NewEnvelope(0, 0, 10, 10)

	t.Run("interior", func(t *testing.T) {
		assert.True(t, env.IntersectsXY(5, 5))
	})

	t.Run("exterior", func(t *testing.T) {
		assert.False(t, env.IntersectsXY(20, 20))
	})
}

func TestEnvelope_Disjoint(t *testing.T) {
	t.Run("overlapping are not disjoint", func(t *testing.T) {
		e1 := geom.NewEnvelope(0, 0, 10, 10)
		e2 := geom.NewEnvelope(5, 5, 15, 15)
		assert.False(t, e1.Disjoint(e2))
	})

	t.Run("separated are disjoint", func(t *testing.T) {
		e1 := geom.NewEnvelope(0, 0, 10, 10)
		e2 := geom.NewEnvelope(20, 20, 30, 30)
		assert.True(t, e1.Disjoint(e2))
	})

	t.Run("touching at corner are not disjoint", func(t *testing.T) {
		e1 := geom.NewEnvelope(0, 0, 10, 10)
		e2 := geom.NewEnvelope(10, 10, 20, 20)
		assert.False(t, e1.Disjoint(e2))
	})

	t.Run("touching on edge are not disjoint", func(t *testing.T) {
		e1 := geom.NewEnvelope(0, 0, 10, 10)
		e2 := geom.NewEnvelope(10, 0, 20, 10)
		assert.False(t, e1.Disjoint(e2))
	})

	t.Run("empty envelopes are disjoint", func(t *testing.T) {
		e1 := geom.NewEnvelopeEmpty()
		e2 := geom.NewEnvelope(0, 0, 10, 10)
		assert.True(t, e1.Disjoint(e2))
	})

	t.Run("both empty are disjoint", func(t *testing.T) {
		e1 := geom.NewEnvelopeEmpty()
		e2 := geom.NewEnvelopeEmpty()
		assert.True(t, e1.Disjoint(e2))
	})
}

func TestEnvelope_Intersection(t *testing.T) {
	t.Run("overlapping envelopes", func(t *testing.T) {
		e1 := geom.NewEnvelope(0, 0, 10, 10)
		e2 := geom.NewEnvelope(5, 5, 15, 15)
		result := e1.Intersection(e2)
		require.False(t, result.IsNull())
		assert.Equal(t, 5.0, result.MinX)
		assert.Equal(t, 5.0, result.MinY)
		assert.Equal(t, 10.0, result.MaxX)
		assert.Equal(t, 10.0, result.MaxY)
	})

	t.Run("non-overlapping returns empty", func(t *testing.T) {
		e1 := geom.NewEnvelope(0, 0, 10, 10)
		e2 := geom.NewEnvelope(20, 20, 30, 30)
		result := e1.Intersection(e2)
		assert.True(t, result.IsNull())
	})

	t.Run("touching at corner", func(t *testing.T) {
		e1 := geom.NewEnvelope(0, 0, 10, 10)
		e2 := geom.NewEnvelope(10, 10, 20, 20)
		result := e1.Intersection(e2)
		require.False(t, result.IsNull())
		// The intersection is a single point (10,10)
		assert.Equal(t, 10.0, result.MinX)
		assert.Equal(t, 10.0, result.MinY)
		assert.Equal(t, 10.0, result.MaxX)
		assert.Equal(t, 10.0, result.MaxY)
	})

	t.Run("one inside the other", func(t *testing.T) {
		outer := geom.NewEnvelope(0, 0, 100, 100)
		inner := geom.NewEnvelope(10, 10, 20, 20)
		result := outer.Intersection(inner)
		require.False(t, result.IsNull())
		assert.Equal(t, 10.0, result.MinX)
		assert.Equal(t, 10.0, result.MinY)
		assert.Equal(t, 20.0, result.MaxX)
		assert.Equal(t, 20.0, result.MaxY)
	})

	t.Run("with empty envelope", func(t *testing.T) {
		e1 := geom.NewEnvelope(0, 0, 10, 10)
		empty := geom.NewEnvelopeEmpty()
		result := e1.Intersection(empty)
		assert.True(t, result.IsNull())
	})

	t.Run("intersection is commutative", func(t *testing.T) {
		e1 := geom.NewEnvelope(0, 0, 10, 10)
		e2 := geom.NewEnvelope(5, 3, 15, 8)
		r1 := e1.Intersection(e2)
		r2 := e2.Intersection(e1)
		assert.True(t, r1.Equals(r2, 1e-10))
	})
}

func TestEnvelope_Equals(t *testing.T) {
	eps := 1e-10

	t.Run("equal envelopes", func(t *testing.T) {
		e1 := geom.NewEnvelope(0, 0, 10, 10)
		e2 := geom.NewEnvelope(0, 0, 10, 10)
		assert.True(t, e1.Equals(e2, eps))
	})

	t.Run("different envelopes", func(t *testing.T) {
		e1 := geom.NewEnvelope(0, 0, 10, 10)
		e2 := geom.NewEnvelope(0, 0, 10, 11)
		assert.False(t, e1.Equals(e2, eps))
	})

	t.Run("both empty are equal", func(t *testing.T) {
		e1 := geom.NewEnvelopeEmpty()
		e2 := geom.NewEnvelopeEmpty()
		assert.True(t, e1.Equals(e2, eps))
	})

	t.Run("empty vs non-empty not equal", func(t *testing.T) {
		e1 := geom.NewEnvelopeEmpty()
		e2 := geom.NewEnvelope(0, 0, 10, 10)
		assert.False(t, e1.Equals(e2, eps))
	})

	t.Run("non-empty vs empty not equal", func(t *testing.T) {
		e1 := geom.NewEnvelope(0, 0, 10, 10)
		e2 := geom.NewEnvelopeEmpty()
		assert.False(t, e1.Equals(e2, eps))
	})

	t.Run("within epsilon are equal", func(t *testing.T) {
		e1 := geom.NewEnvelope(0, 0, 10, 10)
		e2 := geom.NewEnvelope(0, 0, 10.0000000001, 10)
		assert.True(t, e1.Equals(e2, 1e-9))
	})

	t.Run("outside epsilon are not equal", func(t *testing.T) {
		e1 := geom.NewEnvelope(0, 0, 10, 10)
		e2 := geom.NewEnvelope(0, 0, 10.01, 10)
		assert.False(t, e1.Equals(e2, 1e-3))
	})
}

// ---------------------------------------------------------------------------
// GeometryCollection: ApplyCoordinateFilter
// ---------------------------------------------------------------------------

// testTranslateFilter is a CoordinateFilter that shifts X by dx and Y by dy.
type testTranslateFilter struct {
	dx, dy float64
}

func (f *testTranslateFilter) Filter(c *geom.Coordinate) {
	c.X += f.dx
	c.Y += f.dy
}

// testCountFilter counts how many coordinates it visits.
type testCountFilter struct {
	count int
}

func (f *testCountFilter) Filter(c *geom.Coordinate) {
	f.count++
}

func TestGeometryCollection_ApplyCoordinateFilter(t *testing.T) {
	t.Run("translate points in collection", func(t *testing.T) {
		gc := geom.NewGeometryCollection([]geom.Geometry{
			geom.NewPoint(1, 2),
			geom.NewPoint(3, 4),
		})

		filter := &testTranslateFilter{dx: 10, dy: 20}
		gc.ApplyCoordinateFilter(filter)

		coords := gc.Coordinates()
		require.Equal(t, 2, len(coords))
		assert.InDelta(t, 11.0, coords[0].X, 1e-10)
		assert.InDelta(t, 22.0, coords[0].Y, 1e-10)
		assert.InDelta(t, 13.0, coords[1].X, 1e-10)
		assert.InDelta(t, 24.0, coords[1].Y, 1e-10)
	})

	t.Run("translate linestring in collection", func(t *testing.T) {
		gc := geom.NewGeometryCollection([]geom.Geometry{
			mustLineStringXY(0, 0, 10, 10),
		})

		filter := &testTranslateFilter{dx: 5, dy: 5}
		gc.ApplyCoordinateFilter(filter)

		coords := gc.Coordinates()
		require.Equal(t, 2, len(coords))
		assert.InDelta(t, 5.0, coords[0].X, 1e-10)
		assert.InDelta(t, 5.0, coords[0].Y, 1e-10)
		assert.InDelta(t, 15.0, coords[1].X, 1e-10)
		assert.InDelta(t, 15.0, coords[1].Y, 1e-10)
	})

	t.Run("nil filter is a no-op", func(t *testing.T) {
		gc := geom.NewGeometryCollection([]geom.Geometry{
			geom.NewPoint(1, 2),
		})
		gc.ApplyCoordinateFilter(nil) // should not panic
		coords := gc.Coordinates()
		assert.InDelta(t, 1.0, coords[0].X, 1e-10)
	})

	t.Run("empty collection with filter is a no-op", func(t *testing.T) {
		gc := geom.NewGeometryCollectionEmpty()
		counter := &testCountFilter{}
		gc.ApplyCoordinateFilter(counter)
		assert.Equal(t, 0, counter.count)
	})

	t.Run("filter visits all coordinates in mixed collection", func(t *testing.T) {
		gc := geom.NewGeometryCollection([]geom.Geometry{
			geom.NewPoint(0, 0),                       // 1 coordinate
			mustLineStringXY(1, 1, 2, 2, 3, 3),   // 3 coordinates
		})
		counter := &testCountFilter{}
		gc.ApplyCoordinateFilter(counter)
		assert.Equal(t, 4, counter.count)
	})

	t.Run("filter invalidates envelope cache", func(t *testing.T) {
		gc := geom.NewGeometryCollection([]geom.Geometry{
			geom.NewPoint(0, 0),
			geom.NewPoint(10, 10),
		})

		// Access envelope to populate cache.
		envBefore := gc.Envelope()
		assert.Equal(t, 0.0, envBefore.MinX)
		assert.Equal(t, 10.0, envBefore.MaxX)

		// Apply filter to shift everything.
		filter := &testTranslateFilter{dx: 100, dy: 100}
		gc.ApplyCoordinateFilter(filter)

		// Envelope should reflect the new coordinates.
		envAfter := gc.Envelope()
		assert.InDelta(t, 100.0, envAfter.MinX, 1e-10)
		assert.InDelta(t, 110.0, envAfter.MaxX, 1e-10)
	})
}

// ---------------------------------------------------------------------------
// Additional edge-case and degenerate-input tests
// ---------------------------------------------------------------------------

func TestCoordinate_IsNaN(t *testing.T) {
	t.Run("normal coordinate is not NaN", func(t *testing.T) {
		c := geom.NewCoordinate(1, 2)
		assert.False(t, c.IsNaN())
	})

	t.Run("NaN X", func(t *testing.T) {
		c := geom.Coordinate{X: math.NaN(), Y: 2}
		assert.True(t, c.IsNaN())
	})

	t.Run("NaN Y", func(t *testing.T) {
		c := geom.Coordinate{X: 1, Y: math.NaN()}
		assert.True(t, c.IsNaN())
	})
}

func TestCoordinateSequence_HasZ(t *testing.T) {
	t.Run("no Z in any coord", func(t *testing.T) {
		seq := mustCoordsXY(0, 0, 1, 1)
		assert.False(t, seq.HasZ())
	})

	t.Run("one coord with Z", func(t *testing.T) {
		seq := geom.NewCoordinateSequence(
			geom.NewCoordinate(0, 0),
			geom.NewCoordinateZ(1, 1, 5),
		)
		assert.True(t, seq.HasZ())
	})

	t.Run("empty sequence has no Z", func(t *testing.T) {
		seq := geom.NewCoordinateSequence()
		assert.False(t, seq.HasZ())
	})
}

func TestCoordinateSequence_HasM(t *testing.T) {
	t.Run("no M in any coord", func(t *testing.T) {
		seq := mustCoordsXY(0, 0, 1, 1)
		assert.False(t, seq.HasM())
	})

	t.Run("one coord with M", func(t *testing.T) {
		seq := geom.NewCoordinateSequence(
			geom.NewCoordinate(0, 0),
			geom.NewCoordinateM(1, 1, 10),
		)
		assert.True(t, seq.HasM())
	})

	t.Run("empty sequence has no M", func(t *testing.T) {
		seq := geom.NewCoordinateSequence()
		assert.False(t, seq.HasM())
	})
}

func TestCoordinateSequence_Reverse(t *testing.T) {
	t.Run("normal reverse", func(t *testing.T) {
		seq := mustCoordsXY(0, 0, 1, 1, 2, 2)
		rev := seq.Reverse()
		require.Equal(t, 3, rev.Len())
		assert.Equal(t, 2.0, rev.Get(0).X)
		assert.Equal(t, 1.0, rev.Get(1).X)
		assert.Equal(t, 0.0, rev.Get(2).X)
	})

	t.Run("reverse of nil is nil", func(t *testing.T) {
		var seq geom.CoordinateSequence
		assert.Nil(t, seq.Reverse())
	})

	t.Run("single element", func(t *testing.T) {
		seq := mustCoordsXY(5, 5)
		rev := seq.Reverse()
		require.Equal(t, 1, rev.Len())
		assert.Equal(t, 5.0, rev.Get(0).X)
	})
}

func TestCoordinateSequence_Clone_Nil(t *testing.T) {
	var seq geom.CoordinateSequence
	clone := seq.Clone()
	assert.Nil(t, clone)
}

func TestCoordinateSequence_Envelope(t *testing.T) {
	t.Run("non-empty", func(t *testing.T) {
		seq := mustCoordsXY(-5, -10, 20, 30, 0, 0)
		env := seq.Envelope()
		assert.Equal(t, -5.0, env.MinX)
		assert.Equal(t, -10.0, env.MinY)
		assert.Equal(t, 20.0, env.MaxX)
		assert.Equal(t, 30.0, env.MaxY)
	})

	t.Run("empty sequence returns null envelope", func(t *testing.T) {
		seq := geom.NewCoordinateSequence()
		env := seq.Envelope()
		assert.True(t, env.IsNull())
	})

	t.Run("single point", func(t *testing.T) {
		seq := mustCoordsXY(5, 10)
		env := seq.Envelope()
		assert.False(t, env.IsNull())
		assert.Equal(t, 5.0, env.MinX)
		assert.Equal(t, 10.0, env.MinY)
		assert.Equal(t, 5.0, env.MaxX)
		assert.Equal(t, 10.0, env.MaxY)
	})
}

func TestCoordinateSequence_IsClosed(t *testing.T) {
	t.Run("fewer than 2 points is not closed", func(t *testing.T) {
		seq := mustCoordsXY(0, 0)
		assert.False(t, seq.IsClosed(1e-10))
	})

	t.Run("empty is not closed", func(t *testing.T) {
		seq := geom.NewCoordinateSequence()
		assert.False(t, seq.IsClosed(1e-10))
	})
}

func TestNewCoordinateSequenceXY_OddReturnsError(t *testing.T) {
	_, err := geom.NewCoordinateSequenceXY(1, 2, 3)
	assert.Error(t, err, "Expected error for odd number of values")
}

func TestEnvelope_ExpandToInclude_NilOther(t *testing.T) {
	env := geom.NewEnvelope(0, 0, 10, 10)
	env.ExpandToInclude(nil)
	// Should be unchanged.
	assert.Equal(t, 0.0, env.MinX)
	assert.Equal(t, 10.0, env.MaxX)
}

func TestEnvelope_ExpandToInclude_EmptyIntoEmpty(t *testing.T) {
	env := geom.NewEnvelopeEmpty()
	other := geom.NewEnvelopeEmpty()
	env.ExpandToInclude(other)
	assert.True(t, env.IsNull(), "expanding empty by empty should stay empty")
}

func TestEnvelope_ExpandToInclude_NonEmptyIntoEmpty(t *testing.T) {
	env := geom.NewEnvelopeEmpty()
	other := geom.NewEnvelope(5, 5, 15, 15)
	env.ExpandToInclude(other)
	assert.False(t, env.IsNull())
	assert.Equal(t, 5.0, env.MinX)
	assert.Equal(t, 15.0, env.MaxX)
}

func TestEnvelope_Width_Height_Null(t *testing.T) {
	env := geom.NewEnvelopeEmpty()
	assert.Equal(t, 0.0, env.Width())
	assert.Equal(t, 0.0, env.Height())
}

func TestEnvelope_Area_Zero(t *testing.T) {
	t.Run("point envelope", func(t *testing.T) {
		env := geom.NewEnvelopeFromCoord(geom.NewCoordinate(5, 5))
		assert.Equal(t, 0.0, env.Area())
	})

	t.Run("line envelope", func(t *testing.T) {
		env := geom.NewEnvelope(0, 0, 10, 0) // zero height
		assert.Equal(t, 0.0, env.Area())
	})
}

func TestEnvelope_Centre_Empty(t *testing.T) {
	env := geom.NewEnvelopeEmpty()
	c := env.Centre()
	assert.True(t, math.IsNaN(c.X))
	assert.True(t, math.IsNaN(c.Y))
}

func TestEnvelope_SetToNull(t *testing.T) {
	env := geom.NewEnvelope(0, 0, 10, 10)
	assert.False(t, env.IsNull())
	env.SetToNull()
	assert.True(t, env.IsNull())
}

func TestEnvelope_Init(t *testing.T) {
	env := geom.NewEnvelopeEmpty()
	env.Init(20, 30, 5, 10)
	assert.Equal(t, 5.0, env.MinX)
	assert.Equal(t, 10.0, env.MinY)
	assert.Equal(t, 20.0, env.MaxX)
	assert.Equal(t, 30.0, env.MaxY)
}

func TestEnvelope_Translate(t *testing.T) {
	t.Run("non-empty", func(t *testing.T) {
		env := geom.NewEnvelope(0, 0, 10, 10)
		env.Translate(5, -3)
		assert.Equal(t, 5.0, env.MinX)
		assert.Equal(t, -3.0, env.MinY)
		assert.Equal(t, 15.0, env.MaxX)
		assert.Equal(t, 7.0, env.MaxY)
	})

	t.Run("empty envelope stays empty", func(t *testing.T) {
		env := geom.NewEnvelopeEmpty()
		env.Translate(10, 10)
		assert.True(t, env.IsNull())
	})
}

func TestEnvelope_Distance_Intersecting(t *testing.T) {
	e1 := geom.NewEnvelope(0, 0, 10, 10)
	e2 := geom.NewEnvelope(5, 5, 15, 15)
	assert.Equal(t, 0.0, e1.Distance(e2))
}

func TestEnvelope_Distance_EdgeAligned(t *testing.T) {
	// Envelopes separated in X only.
	e1 := geom.NewEnvelope(0, 0, 10, 10)
	e2 := geom.NewEnvelope(15, 0, 20, 10) // Same Y range, separated by 5 in X
	assert.InDelta(t, 5.0, e1.Distance(e2), 1e-10)
}

func TestEnvelope_Distance_DiagonallySeparated(t *testing.T) {
	e1 := geom.NewEnvelope(0, 0, 10, 10)
	e2 := geom.NewEnvelope(13, 14, 20, 20)
	// Closest points: (10,10) and (13,14), dist = sqrt(9+16) = 5
	assert.InDelta(t, 5.0, e1.Distance(e2), 1e-10)
}

func TestEnvelope_NewEnvelopeFromCoord(t *testing.T) {
	c := geom.NewCoordinate(7, 13)
	env := geom.NewEnvelopeFromCoord(c)
	assert.Equal(t, 7.0, env.MinX)
	assert.Equal(t, 13.0, env.MinY)
	assert.Equal(t, 7.0, env.MaxX)
	assert.Equal(t, 13.0, env.MaxY)
	assert.False(t, env.IsNull())
	assert.Equal(t, 0.0, env.Area())
}

func TestEnvelope_Intersects_EmptyEnvelopes(t *testing.T) {
	t.Run("first empty", func(t *testing.T) {
		e1 := geom.NewEnvelopeEmpty()
		e2 := geom.NewEnvelope(0, 0, 10, 10)
		assert.False(t, e1.Intersects(e2))
	})

	t.Run("second empty", func(t *testing.T) {
		e1 := geom.NewEnvelope(0, 0, 10, 10)
		e2 := geom.NewEnvelopeEmpty()
		assert.False(t, e1.Intersects(e2))
	})

	t.Run("both empty", func(t *testing.T) {
		e1 := geom.NewEnvelopeEmpty()
		e2 := geom.NewEnvelopeEmpty()
		assert.False(t, e1.Intersects(e2))
	})
}
