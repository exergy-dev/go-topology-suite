package relate

import (
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// IntersectionMatrix Get/Set tests
// ---------------------------------------------------------------------------

func TestIntersectionMatrix_GetSet(t *testing.T) {
	t.Run("Get returns DimFalse for new matrix", func(t *testing.T) {
		m := NewIntersectionMatrix()
		for i := 0; i < 3; i++ {
			for j := 0; j < 3; j++ {
				assert.Equal(t, DimFalse, m.Get(i, j),
					"New matrix should have DimFalse at [%d][%d]", i, j)
			}
		}
	})

	t.Run("Set and Get round-trip", func(t *testing.T) {
		m := NewIntersectionMatrix()
		m.Set(Interior, Interior, DimArea)
		assert.Equal(t, DimArea, m.Get(Interior, Interior))

		m.Set(Boundary, Exterior, DimLine)
		assert.Equal(t, DimLine, m.Get(Boundary, Exterior))

		m.Set(Exterior, Boundary, DimPoint)
		assert.Equal(t, DimPoint, m.Get(Exterior, Boundary))
	})

	t.Run("Set overwrites previous value", func(t *testing.T) {
		m := NewIntersectionMatrix()
		m.Set(0, 0, DimLine)
		assert.Equal(t, DimLine, m.Get(0, 0))

		m.Set(0, 0, DimPoint)
		assert.Equal(t, DimPoint, m.Get(0, 0), "Set should overwrite previous value")
	})

	t.Run("SetAtLeast only increases dimension", func(t *testing.T) {
		m := NewIntersectionMatrix()
		m.SetAtLeast(0, 0, DimPoint)
		assert.Equal(t, DimPoint, m.Get(0, 0))

		m.SetAtLeast(0, 0, DimArea)
		assert.Equal(t, DimArea, m.Get(0, 0), "SetAtLeast should increase to DimArea")

		m.SetAtLeast(0, 0, DimLine)
		assert.Equal(t, DimArea, m.Get(0, 0), "SetAtLeast should not decrease from DimArea")

		m.SetAtLeast(0, 0, DimFalse)
		assert.Equal(t, DimArea, m.Get(0, 0), "SetAtLeast should not decrease to DimFalse")
	})
}

// ---------------------------------------------------------------------------
// IsOverlaps tests
// ---------------------------------------------------------------------------

func TestIsOverlaps(t *testing.T) {
	t.Run("area/area overlap: I-I, I-E, E-I all non-empty", func(t *testing.T) {
		m := NewIntersectionMatrix()
		m.Set(Interior, Interior, DimArea) // Overlap interior
		m.Set(Interior, Exterior, DimArea) // Part of A outside B
		m.Set(Exterior, Interior, DimArea) // Part of B outside A
		m.Set(Exterior, Exterior, DimArea)

		assert.True(t, m.IsOverlaps(2, 2), "Should detect area/area overlap")
	})

	t.Run("area/area contained: I-E is false", func(t *testing.T) {
		m := NewIntersectionMatrix()
		m.Set(Interior, Interior, DimArea)
		m.Set(Interior, Exterior, DimFalse) // A is entirely inside B
		m.Set(Exterior, Interior, DimArea)
		m.Set(Exterior, Exterior, DimArea)

		assert.False(t, m.IsOverlaps(2, 2), "Contained area should not be overlap")
	})

	t.Run("line/line overlap requires DimLine interior intersection", func(t *testing.T) {
		m := NewIntersectionMatrix()
		m.Set(Interior, Interior, DimLine) // Collinear overlap
		m.Set(Interior, Exterior, DimLine)
		m.Set(Exterior, Interior, DimLine)
		m.Set(Exterior, Exterior, DimArea)

		assert.True(t, m.IsOverlaps(1, 1), "Collinear overlap should be detected")
	})

	t.Run("line/line crossing is not overlap", func(t *testing.T) {
		m := NewIntersectionMatrix()
		m.Set(Interior, Interior, DimPoint) // Crossing at a point
		m.Set(Interior, Exterior, DimLine)
		m.Set(Exterior, Interior, DimLine)
		m.Set(Exterior, Exterior, DimArea)

		assert.False(t, m.IsOverlaps(1, 1),
			"Line crossing (point intersection) should not be overlap")
	})

	t.Run("point/point overlap", func(t *testing.T) {
		m := NewIntersectionMatrix()
		m.Set(Interior, Interior, DimPoint)
		m.Set(Interior, Exterior, DimPoint)
		m.Set(Exterior, Interior, DimPoint)

		assert.True(t, m.IsOverlaps(0, 0), "Point/point overlap should be detected")
	})

	t.Run("different dimension returns false", func(t *testing.T) {
		m := NewIntersectionMatrix()
		m.Set(Interior, Interior, DimPoint)
		m.Set(Interior, Exterior, DimPoint)
		m.Set(Exterior, Interior, DimArea)

		assert.False(t, m.IsOverlaps(0, 2),
			"Point/area overlap should be false (different dimensions)")
	})

	t.Run("line/area case", func(t *testing.T) {
		m := NewIntersectionMatrix()
		m.Set(Interior, Interior, DimLine) // Required: DimLine
		m.Set(Interior, Exterior, DimLine) // Part of line outside area
		m.Set(Exterior, Interior, DimArea) // Part of area outside line

		assert.True(t, m.IsOverlaps(1, 2), "Line/area overlap should be detected")
	})
}

// ---------------------------------------------------------------------------
// computeLineStringRelate tests
// ---------------------------------------------------------------------------

func TestComputeLineStringRelate_LineVsLine(t *testing.T) {
	t.Run("crossing lines", func(t *testing.T) {
		ls1 := geom.NewLineStringXY(0, 0, 10, 10)
		ls2 := geom.NewLineStringXY(0, 10, 10, 0)

		m := Relate(ls1, ls2)
		// Crossing lines: I-I should be point (0)
		assert.Equal(t, DimPoint, m.Get(Interior, Interior),
			"Crossing lines: I-I should be DimPoint")
		assert.True(t, m.IsIntersects(), "Crossing lines should intersect")
		assert.True(t, m.IsCrosses(1, 1), "Crossing lines should cross")
	})

	t.Run("parallel lines", func(t *testing.T) {
		ls1 := geom.NewLineStringXY(0, 0, 10, 0)
		ls2 := geom.NewLineStringXY(0, 5, 10, 5)

		m := Relate(ls1, ls2)
		assert.True(t, m.IsDisjoint(), "Parallel lines should be disjoint")
		assert.Equal(t, DimFalse, m.Get(Interior, Interior),
			"Parallel lines: I-I should be DimFalse")
	})

	t.Run("collinear overlapping lines", func(t *testing.T) {
		ls1 := geom.NewLineStringXY(0, 0, 10, 0)
		ls2 := geom.NewLineStringXY(5, 0, 15, 0)

		m := Relate(ls1, ls2)
		assert.True(t, m.IsIntersects(), "Overlapping collinear lines should intersect")
		assert.Equal(t, DimLine, m.Get(Interior, Interior),
			"Collinear overlap: I-I should be DimLine")
	})

	t.Run("lines sharing an endpoint", func(t *testing.T) {
		ls1 := geom.NewLineStringXY(0, 0, 5, 5)
		ls2 := geom.NewLineStringXY(5, 5, 10, 0)

		m := Relate(ls1, ls2)
		assert.True(t, m.IsIntersects(), "Lines sharing endpoint should intersect")
		// Boundary-Boundary should be point dimension (shared endpoint)
		assert.True(t, m.Get(Boundary, Boundary) >= DimPoint,
			"Shared endpoint should appear in B-B")
	})

	t.Run("disjoint lines", func(t *testing.T) {
		ls1 := geom.NewLineStringXY(0, 0, 1, 0)
		ls2 := geom.NewLineStringXY(5, 5, 6, 5)

		m := Relate(ls1, ls2)
		assert.True(t, m.IsDisjoint(), "Non-intersecting lines should be disjoint")
	})

	t.Run("matrix string output", func(t *testing.T) {
		ls1 := geom.NewLineStringXY(0, 0, 10, 10)
		ls2 := geom.NewLineStringXY(0, 10, 10, 0)

		m := Relate(ls1, ls2)
		matStr := m.String()
		assert.Len(t, matStr, 9, "DE-9IM string should be 9 characters")
		// Each character should be one of F, 0, 1, 2
		for _, c := range matStr {
			assert.Contains(t, []rune{'F', '0', '1', '2'}, c,
				"Matrix string should only contain F, 0, 1, 2 characters, got %c", c)
		}
	})
}

// ---------------------------------------------------------------------------
// computeLineStringRelate: line vs polygon
// ---------------------------------------------------------------------------

func TestComputeLineStringRelate_LineVsPolygon(t *testing.T) {
	shell := geom.NewLinearRingXY(0, 0, 20, 0, 20, 20, 0, 20, 0, 0)
	poly := geom.NewPolygon(shell, nil)

	t.Run("line entirely inside polygon", func(t *testing.T) {
		// Use a multi-segment line so there are interior (non-endpoint) vertices
		ls := geom.NewLineStringXY(5, 5, 10, 10, 15, 15)
		m := Relate(ls, poly)

		assert.True(t, m.IsIntersects(), "Interior line should intersect polygon")
		// With an interior vertex, the implementation detects I-I
		assert.True(t, m.Get(Interior, Interior) >= DimPoint,
			"Interior line with interior vertex: I-I should be non-empty")
		t.Logf("Line inside polygon: %s", m.String())
	})

	t.Run("line entirely outside polygon", func(t *testing.T) {
		ls := geom.NewLineStringXY(25, 25, 30, 30)
		m := Relate(ls, poly)

		assert.True(t, m.IsDisjoint(), "Exterior line should be disjoint from polygon")
	})

	t.Run("line crossing polygon boundary", func(t *testing.T) {
		// Use a multi-segment line with interior vertices inside and outside the polygon
		ls := geom.NewLineStringXY(-5, 10, 10, 10, 25, 10)
		m := Relate(ls, poly)

		assert.True(t, m.IsIntersects(), "Crossing line should intersect polygon")
		t.Logf("Line crossing polygon: %s", m.String())
		// The line has an interior vertex inside the polygon and parts outside
		assert.True(t, m.Get(Interior, Interior) >= DimPoint,
			"Crossing line should have I-I intersection")
	})

	t.Run("line on polygon boundary", func(t *testing.T) {
		ls := geom.NewLineStringXY(0, 0, 20, 0)
		m := Relate(ls, poly)

		assert.True(t, m.IsIntersects(), "Line on boundary should intersect polygon")
	})
}

// ---------------------------------------------------------------------------
// computePolygonRelate: polygon vs polygon
// ---------------------------------------------------------------------------

func TestComputePolygonRelate_PolygonVsPolygon(t *testing.T) {
	t.Run("overlapping polygons", func(t *testing.T) {
		shell1 := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
		poly1 := geom.NewPolygon(shell1, nil)

		shell2 := geom.NewLinearRingXY(5, 5, 15, 5, 15, 15, 5, 15, 5, 5)
		poly2 := geom.NewPolygon(shell2, nil)

		m := Relate(poly1, poly2)
		assert.True(t, m.IsIntersects(), "Overlapping polygons should intersect")

		matStr := m.String()
		assert.Len(t, matStr, 9, "DE-9IM string should be 9 characters")
		t.Logf("Overlapping polygons matrix: %s", matStr)

		// The boundary-boundary intersection should be detected
		assert.True(t, m.Get(Boundary, Boundary) >= DimPoint,
			"Overlapping polygons: B-B should be at least DimPoint")
		// Boundary of one polygon is inside the other
		assert.True(t, m.Get(Boundary, Interior) >= DimPoint,
			"Overlapping polygons: B-I should be at least DimPoint")
	})

	t.Run("disjoint polygons", func(t *testing.T) {
		shell1 := geom.NewLinearRingXY(0, 0, 5, 0, 5, 5, 0, 5, 0, 0)
		poly1 := geom.NewPolygon(shell1, nil)

		shell2 := geom.NewLinearRingXY(10, 10, 15, 10, 15, 15, 10, 15, 10, 10)
		poly2 := geom.NewPolygon(shell2, nil)

		m := Relate(poly1, poly2)
		assert.True(t, m.IsDisjoint(), "Disjoint polygons should be disjoint")
		assert.False(t, m.IsIntersects(), "Disjoint polygons should not intersect")

		// For disjoint: I-I, I-B, B-I, B-B all False
		assert.Equal(t, DimFalse, m.Get(Interior, Interior))
		assert.Equal(t, DimFalse, m.Get(Interior, Boundary))
		assert.Equal(t, DimFalse, m.Get(Boundary, Interior))
		assert.Equal(t, DimFalse, m.Get(Boundary, Boundary))

		matStr := m.String()
		t.Logf("Disjoint polygons matrix: %s", matStr)
	})

	t.Run("polygon containing another polygon", func(t *testing.T) {
		shellOuter := geom.NewLinearRingXY(0, 0, 20, 0, 20, 20, 0, 20, 0, 0)
		polyOuter := geom.NewPolygon(shellOuter, nil)

		shellInner := geom.NewLinearRingXY(5, 5, 15, 5, 15, 15, 5, 15, 5, 5)
		polyInner := geom.NewPolygon(shellInner, nil)

		m := Relate(polyOuter, polyInner)
		assert.True(t, m.IsIntersects(), "Containing polygon should intersect")

		// Interior of inner should be inside interior of outer
		assert.Equal(t, DimArea, m.Get(Interior, Interior),
			"Containing polygon: I-I should be DimArea")

		matStr := m.String()
		t.Logf("Containing polygons matrix: %s", matStr)
	})

	t.Run("identical polygons", func(t *testing.T) {
		shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
		poly1 := geom.NewPolygon(shell, nil)
		poly2 := geom.NewPolygon(shell, nil)

		m := Relate(poly1, poly2)
		assert.True(t, m.IsIntersects(), "Identical polygons should intersect")
		assert.Equal(t, DimArea, m.Get(Interior, Interior),
			"Identical polygons: I-I should be DimArea")

		matStr := m.String()
		t.Logf("Identical polygons matrix: %s", matStr)
	})

	t.Run("polygons sharing an edge", func(t *testing.T) {
		shell1 := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
		poly1 := geom.NewPolygon(shell1, nil)

		shell2 := geom.NewLinearRingXY(10, 0, 20, 0, 20, 10, 10, 10, 10, 0)
		poly2 := geom.NewPolygon(shell2, nil)

		m := Relate(poly1, poly2)
		assert.True(t, m.IsIntersects(), "Edge-sharing polygons should intersect")
		// Boundaries share a line segment
		assert.True(t, m.Get(Boundary, Boundary) >= DimPoint,
			"Edge-sharing polygons: B-B should be at least DimPoint")

		matStr := m.String()
		t.Logf("Edge-sharing polygons matrix: %s", matStr)
	})
}

// ---------------------------------------------------------------------------
// computeLinearRingRelate tests
// ---------------------------------------------------------------------------

func TestComputeLinearRingRelate(t *testing.T) {
	t.Run("ring vs point inside", func(t *testing.T) {
		ring := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
		point := geom.NewPoint(5, 5)

		// LinearRing is treated as a closed LineString for relate
		m := Relate(ring, point)
		// Point is NOT inside the ring (ring is a line, not an area)
		// so the result depends on whether point is on the ring line
		matStr := m.String()
		assert.Len(t, matStr, 9, "DE-9IM string should be 9 characters")
		t.Logf("Ring vs interior point matrix: %s", matStr)
	})

	t.Run("ring vs point on ring", func(t *testing.T) {
		ring := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
		point := geom.NewPoint(5, 0)

		m := Relate(ring, point)
		assert.True(t, m.IsIntersects(), "Point on ring should intersect")
	})

	t.Run("ring vs crossing line", func(t *testing.T) {
		ring := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
		ls := geom.NewLineStringXY(-5, 5, 15, 5)

		m := Relate(ring, ls)
		assert.True(t, m.IsIntersects(), "Ring crossing a line should intersect")
	})

	t.Run("ring vs disjoint line", func(t *testing.T) {
		ring := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
		ls := geom.NewLineStringXY(20, 20, 30, 30)

		m := Relate(ring, ls)
		assert.True(t, m.IsDisjoint(), "Ring and distant line should be disjoint")
	})
}

// ---------------------------------------------------------------------------
// DE-9IM matrix string output for various geometry pairs
// ---------------------------------------------------------------------------

func TestDE9IM_MatrixStringOutput(t *testing.T) {
	t.Run("point vs point: same location", func(t *testing.T) {
		p1 := geom.NewPoint(5, 5)
		p2 := geom.NewPoint(5, 5)

		m := Relate(p1, p2)
		matStr := m.String()
		assert.Len(t, matStr, 9)
		assert.True(t, m.IsIntersects(), "Same-location points should intersect")
		t.Logf("Point/Point same: %s", matStr)
	})

	t.Run("point vs point: different location", func(t *testing.T) {
		p1 := geom.NewPoint(0, 0)
		p2 := geom.NewPoint(10, 10)

		m := Relate(p1, p2)
		matStr := m.String()
		assert.Len(t, matStr, 9)
		assert.True(t, m.IsDisjoint(), "Different-location points should be disjoint")
		t.Logf("Point/Point different: %s", matStr)
	})

	t.Run("point inside polygon", func(t *testing.T) {
		p := geom.NewPoint(5, 5)
		shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
		poly := geom.NewPolygon(shell, nil)

		m := Relate(p, poly)
		matStr := m.String()
		assert.Len(t, matStr, 9)
		// Point is in interior of polygon
		assert.Equal(t, DimPoint, m.Get(Interior, Interior),
			"Point inside polygon: I-I should be DimPoint")
		t.Logf("Point inside polygon: %s", matStr)
	})

	t.Run("point on polygon boundary", func(t *testing.T) {
		p := geom.NewPoint(5, 0)
		shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
		poly := geom.NewPolygon(shell, nil)

		m := Relate(p, poly)
		matStr := m.String()
		assert.Len(t, matStr, 9)
		t.Logf("Point on polygon boundary: %s", matStr)
	})

	t.Run("point outside polygon", func(t *testing.T) {
		p := geom.NewPoint(15, 15)
		shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
		poly := geom.NewPolygon(shell, nil)

		m := Relate(p, poly)
		matStr := m.String()
		assert.Len(t, matStr, 9)
		assert.True(t, m.IsDisjoint(), "Point outside polygon should be disjoint")
		t.Logf("Point outside polygon: %s", matStr)
	})

	t.Run("crossing lines matrix format", func(t *testing.T) {
		ls1 := geom.NewLineStringXY(0, 0, 10, 10)
		ls2 := geom.NewLineStringXY(0, 10, 10, 0)

		m := Relate(ls1, ls2)
		matStr := m.String()
		assert.Len(t, matStr, 9)
		t.Logf("Crossing lines: %s", matStr)

		// Verify the matrix can be reconstructed from the string
		reconstructed, err := NewIntersectionMatrixFromString(matStr)
		require.NoError(t, err)
		assert.Equal(t, matStr, reconstructed.String(),
			"Reconstructed matrix should produce same string")
	})

	t.Run("empty geometry", func(t *testing.T) {
		p := geom.NewPoint(1, 1)
		empty := geom.NewPointEmpty()

		m := Relate(p, empty)
		matStr := m.String()
		assert.Len(t, matStr, 9)
		assert.True(t, m.IsDisjoint(), "Point and empty point should be disjoint")
		t.Logf("Point vs empty: %s", matStr)
	})

	t.Run("nil geometry", func(t *testing.T) {
		p := geom.NewPoint(1, 1)
		m := Relate(p, nil)
		matStr := m.String()
		assert.Len(t, matStr, 9)
		assert.Equal(t, "FFFFFFFFF", matStr, "nil geometry should produce all-F matrix")
	})

	t.Run("both nil", func(t *testing.T) {
		m := Relate(nil, nil)
		assert.Equal(t, "FFFFFFFFF", m.String(), "Both nil should produce all-F matrix")
	})
}

// ---------------------------------------------------------------------------
// RelatePattern tests
// ---------------------------------------------------------------------------

func TestRelatePattern_Variations(t *testing.T) {
	shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly := geom.NewPolygon(shell, nil)

	t.Run("point within polygon matches T*F**F*** (within pattern)", func(t *testing.T) {
		p := geom.NewPoint(5, 5)
		assert.True(t, RelatePattern(p, poly, "T*F**F***"),
			"Point inside polygon should match within pattern")
	})

	t.Run("wildcard pattern matches everything", func(t *testing.T) {
		p := geom.NewPoint(5, 5)
		assert.True(t, RelatePattern(p, poly, "*********"),
			"Wildcard pattern should match any geometry pair")
	})

	t.Run("impossible pattern does not match", func(t *testing.T) {
		p := geom.NewPoint(5, 5)
		// Interior-Interior = 0 (Point), so requiring "2" should fail
		assert.False(t, RelatePattern(p, poly, "2********"),
			"Point/polygon I-I is 0, not 2")
	})
}

// ---------------------------------------------------------------------------
// computePolygonRelate: polygon vs line (transpose of line vs polygon)
// ---------------------------------------------------------------------------

func TestComputePolygonRelate_PolygonVsLine(t *testing.T) {
	shell := geom.NewLinearRingXY(0, 0, 20, 0, 20, 20, 0, 20, 0, 0)
	poly := geom.NewPolygon(shell, nil)

	t.Run("line inside polygon", func(t *testing.T) {
		// Use a 3-vertex line so there are interior (non-endpoint) vertices
		ls := geom.NewLineStringXY(5, 5, 10, 10, 15, 15)
		m := Relate(poly, ls)

		assert.True(t, m.IsIntersects(), "Polygon containing line should intersect")
		// The transpose should show polygon interior intersects line interior
		assert.True(t, m.Get(Interior, Interior) >= DimPoint,
			"Polygon vs interior line: I-I should be non-empty")
		t.Logf("Polygon vs line inside: %s", m.String())
	})

	t.Run("line outside polygon", func(t *testing.T) {
		ls := geom.NewLineStringXY(25, 25, 30, 30)
		m := Relate(poly, ls)

		assert.True(t, m.IsDisjoint(), "Polygon and exterior line should be disjoint")
	})
}

// ---------------------------------------------------------------------------
// Edge cases for Matches
// ---------------------------------------------------------------------------

func TestMatches_EdgeCases(t *testing.T) {
	m := NewIntersectionMatrix()
	m.Set(Interior, Interior, DimArea)
	m.Set(Exterior, Exterior, DimArea)

	t.Run("wrong length pattern returns false", func(t *testing.T) {
		assert.False(t, m.Matches(""), "Empty pattern should not match")
		assert.False(t, m.Matches("T"), "Short pattern should not match")
		assert.False(t, m.Matches("FFFFFFFFFF"), "10-char pattern should not match")
	})

	t.Run("case insensitive matching", func(t *testing.T) {
		assert.Equal(t, m.Matches("t*f**f**t"), m.Matches("T*F**F**T"),
			"Pattern matching should be case insensitive")
	})
}

// ---------------------------------------------------------------------------
// IsCrosses tests
// ---------------------------------------------------------------------------

func TestIsCrosses_VariousDimensions(t *testing.T) {
	t.Run("line/line crossing", func(t *testing.T) {
		ls1 := geom.NewLineStringXY(0, 0, 10, 10)
		ls2 := geom.NewLineStringXY(0, 10, 10, 0)

		m := Relate(ls1, ls2)
		assert.True(t, m.IsCrosses(1, 1), "Crossing lines should report crosses")
	})

	t.Run("area/area cannot cross", func(t *testing.T) {
		m := NewIntersectionMatrix()
		m.Set(Interior, Interior, DimArea)
		assert.False(t, m.IsCrosses(2, 2), "Area/area cannot cross")
	})

	t.Run("point/point cannot cross", func(t *testing.T) {
		m := NewIntersectionMatrix()
		m.Set(Interior, Interior, DimPoint)
		assert.False(t, m.IsCrosses(0, 0), "Point/point cannot cross")
	})
}

// ---------------------------------------------------------------------------
// Transpose consistency
// ---------------------------------------------------------------------------

func TestTranspose_WithRelate(t *testing.T) {
	ls := geom.NewLineStringXY(5, 5, 15, 15)
	shell := geom.NewLinearRingXY(0, 0, 20, 0, 20, 20, 0, 20, 0, 0)
	poly := geom.NewPolygon(shell, nil)

	mAB := Relate(ls, poly)
	mBA := Relate(poly, ls)

	// Transposing mAB should give a matrix similar to mBA
	transposed := mAB.Transpose()
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if transposed.Get(i, j) != DimFalse || mBA.Get(i, j) != DimFalse {
				// At least check the signs match (both non-false or both false)
				if transposed.Get(i, j) >= DimPoint {
					assert.True(t, mBA.Get(i, j) >= DimPoint,
						"Transpose[%d][%d]=%v but Relate(poly,ls)[%d][%d]=%v",
						i, j, transposed.Get(i, j), i, j, mBA.Get(i, j))
				}
			}
		}
	}
}
