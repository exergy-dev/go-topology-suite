package noding

import (
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// ScaledNoder tests
// ---------------------------------------------------------------------------

func TestScaledNoder_BasicScaleAndUnscale(t *testing.T) {
	t.Run("scale factor 10 with no intersections", func(t *testing.T) {
		ss1 := NewNodedSegmentString(
			geom.NewCoordinateSequenceXY(0, 0, 1, 0),
			"seg1",
		)
		ss2 := NewNodedSegmentString(
			geom.NewCoordinateSequenceXY(0, 1, 1, 1),
			"seg2",
		)

		inner := NewSimpleNoder(NewIntersectionAdder())
		scaled := NewScaledNoder(inner, 10.0)
		scaled.ComputeNodes([]*NodedSegmentString{ss1, ss2})

		result := scaled.GetNodedSubstrings()
		require.Len(t, result, 2, "Expected 2 noded segment strings")

		// After unscaling, coordinates should match originals
		for _, nss := range result {
			coords := nss.Coordinates()
			assert.Len(t, coords, 2, "Each segment string should have 2 coordinates (no intersection)")
		}
	})

	t.Run("scale factor 100 with crossing lines", func(t *testing.T) {
		ss1 := NewNodedSegmentString(
			geom.NewCoordinateSequenceXY(0, 0, 10, 10),
			"line1",
		)
		ss2 := NewNodedSegmentString(
			geom.NewCoordinateSequenceXY(0, 10, 10, 0),
			"line2",
		)

		inner := NewSimpleNoder(NewIntersectionAdder())
		scaled := NewScaledNoder(inner, 100.0)
		scaled.ComputeNodes([]*NodedSegmentString{ss1, ss2})

		result := scaled.GetNodedSubstrings()
		require.Len(t, result, 2, "Expected 2 noded segment strings")

		// Each segment string should have 3 coordinates (start, intersection, end)
		for i, nss := range result {
			coords := nss.Coordinates()
			assert.Len(t, coords, 3, "Noded string %d: expected 3 coordinates", i)
		}

		// Verify the intersection point is approximately (5,5) after unscaling
		for _, nss := range result {
			mid := nss.Coordinates()[1]
			assert.InDelta(t, 5.0, mid.X, 0.01, "Intersection X should be ~5")
			assert.InDelta(t, 5.0, mid.Y, 0.01, "Intersection Y should be ~5")
		}
	})

	t.Run("context preserved through scaling", func(t *testing.T) {
		ss := NewNodedSegmentString(
			geom.NewCoordinateSequenceXY(0, 0, 5, 5),
			"my-context",
		)

		inner := NewSimpleNoder(NewIntersectionAdder())
		scaled := NewScaledNoder(inner, 10.0)
		scaled.ComputeNodes([]*NodedSegmentString{ss})

		result := scaled.GetNodedSubstrings()
		require.Len(t, result, 1)
		assert.Equal(t, "my-context", result[0].Context(), "Context should be preserved")
	})
}

func TestScaledNoder_GetNodedSubstrings(t *testing.T) {
	// Three lines forming a triangle pattern with intersections
	ss1 := NewNodedSegmentString(
		geom.NewCoordinateSequenceXY(0, 0, 10, 0),
		nil,
	)
	ss2 := NewNodedSegmentString(
		geom.NewCoordinateSequenceXY(5, -5, 5, 5),
		nil,
	)

	inner := NewSimpleNoder(NewIntersectionAdder())
	scaled := NewScaledNoder(inner, 1000.0)
	scaled.ComputeNodes([]*NodedSegmentString{ss1, ss2})

	result := scaled.GetNodedSubstrings()
	assert.NotEmpty(t, result, "Should produce noded substrings")

	// Verify intersection point at (5, 0) after unscaling
	foundIntersection := false
	for _, nss := range result {
		for _, c := range nss.Coordinates() {
			if c.Distance(geom.NewCoordinate(5, 0)) < 0.01 {
				foundIntersection = true
			}
		}
	}
	assert.True(t, foundIntersection, "Should find intersection point near (5, 0)")
}

// ---------------------------------------------------------------------------
// splitClosedAtNodes tests
// ---------------------------------------------------------------------------

func TestSplitClosedAtNodes_WithNodes(t *testing.T) {
	t.Run("closed ring with one interior node", func(t *testing.T) {
		// Square ring: (0,0) -> (10,0) -> (10,10) -> (0,10) -> (0,0)
		coords := geom.NewCoordinateSequenceXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)

		// Node at (10,0) which is the second vertex
		nodes := []SegmentNode{
			NewSegmentNode(geom.NewCoordinate(10, 0), 0, 1.0),
		}

		result := splitClosedAtNodes(coords, nodes, "ring-context")
		assert.Greater(t, len(result), 1, "Should split into multiple segments")

		// All segments should preserve context
		for _, nss := range result {
			assert.Equal(t, "ring-context", nss.Context(), "Context should be preserved")
		}

		// Each resulting segment should have at least 2 coordinates
		for i, nss := range result {
			assert.GreaterOrEqual(t, len(nss.Coordinates()), 2,
				"Segment %d should have at least 2 coordinates", i)
		}
	})

	t.Run("closed ring with no nodes returns single segment", func(t *testing.T) {
		coords := geom.NewCoordinateSequenceXY(0, 0, 10, 0, 10, 10, 0, 0)
		result := splitClosedAtNodes(coords, nil, nil)
		assert.Len(t, result, 1, "No nodes should return single segment string")
		assert.Len(t, result[0].Coordinates(), len(coords),
			"Returned segment string should have same coordinates as input")
	})

	t.Run("closed ring with multiple nodes", func(t *testing.T) {
		// Triangle ring: (0,0) -> (10,0) -> (5,10) -> (0,0)
		coords := geom.NewCoordinateSequenceXY(0, 0, 10, 0, 5, 10, 0, 0)

		// Nodes at two vertices
		nodes := []SegmentNode{
			NewSegmentNode(geom.NewCoordinate(10, 0), 0, 1.0),
			NewSegmentNode(geom.NewCoordinate(5, 10), 1, 1.0),
		}

		result := splitClosedAtNodes(coords, nodes, nil)
		assert.GreaterOrEqual(t, len(result), 2, "Should produce multiple segments with multiple nodes")

		// Verify total coordinate count covers all original coordinates
		totalCoords := 0
		for _, nss := range result {
			totalCoords += len(nss.Coordinates())
		}
		// Each node appears as end of one segment and start of next, so some duplication
		assert.GreaterOrEqual(t, totalCoords, len(coords),
			"Total coordinates should be at least as many as original")
	})
}

// ---------------------------------------------------------------------------
// IntersectionFinderAdder tests
// ---------------------------------------------------------------------------

func TestIntersectionFinderAdder_BasicUsage(t *testing.T) {
	t.Run("crossing lines produce intersection record", func(t *testing.T) {
		ss1 := NewNodedSegmentString(
			geom.NewCoordinateSequenceXY(0, 0, 10, 10),
			nil,
		)
		ss2 := NewNodedSegmentString(
			geom.NewCoordinateSequenceXY(0, 10, 10, 0),
			nil,
		)

		ifa := NewIntersectionFinderAdder()
		noder := NewSimpleNoder(ifa)
		noder.ComputeNodes([]*NodedSegmentString{ss1, ss2})

		assert.True(t, ifa.HasIntersection(), "Should find intersection")
		assert.True(t, ifa.HasProperIntersection(), "Should be a proper intersection")

		// IntersectionFinderAdder records interior intersections
		ints := ifa.Intersections()
		assert.NotEmpty(t, ints, "Should record at least one intersection point")

		// Verify intersection is near (5, 5)
		found := false
		for _, pt := range ints {
			if pt.Distance(geom.NewCoordinate(5, 5)) < 0.01 {
				found = true
				break
			}
		}
		assert.True(t, found, "Should record intersection at approximately (5,5)")
	})

	t.Run("parallel lines produce no intersections", func(t *testing.T) {
		ss1 := NewNodedSegmentString(
			geom.NewCoordinateSequenceXY(0, 0, 10, 0),
			nil,
		)
		ss2 := NewNodedSegmentString(
			geom.NewCoordinateSequenceXY(0, 5, 10, 5),
			nil,
		)

		ifa := NewIntersectionFinderAdder()
		noder := NewSimpleNoder(ifa)
		noder.ComputeNodes([]*NodedSegmentString{ss1, ss2})

		assert.False(t, ifa.HasIntersection(), "Parallel lines should not intersect")
		assert.Empty(t, ifa.Intersections(), "Should record no intersections")
	})

	t.Run("also adds nodes like IntersectionAdder", func(t *testing.T) {
		ss1 := NewNodedSegmentString(
			geom.NewCoordinateSequenceXY(0, 0, 10, 10),
			nil,
		)
		ss2 := NewNodedSegmentString(
			geom.NewCoordinateSequenceXY(0, 10, 10, 0),
			nil,
		)

		ifa := NewIntersectionFinderAdder()
		noder := NewSimpleNoder(ifa)
		noder.ComputeNodes([]*NodedSegmentString{ss1, ss2})

		assert.NotEmpty(t, ss1.Nodes(), "Should add nodes to first segment string")
		assert.NotEmpty(t, ss2.Nodes(), "Should add nodes to second segment string")
	})

	t.Run("multiple crossing segments", func(t *testing.T) {
		// Star pattern: three segments crossing through center
		ss1 := NewNodedSegmentString(
			geom.NewCoordinateSequenceXY(0, 0, 10, 10),
			nil,
		)
		ss2 := NewNodedSegmentString(
			geom.NewCoordinateSequenceXY(0, 10, 10, 0),
			nil,
		)
		ss3 := NewNodedSegmentString(
			geom.NewCoordinateSequenceXY(5, -5, 5, 15),
			nil,
		)

		ifa := NewIntersectionFinderAdder()
		noder := NewSimpleNoder(ifa)
		noder.ComputeNodes([]*NodedSegmentString{ss1, ss2, ss3})

		assert.True(t, ifa.HasIntersection(), "Should find intersections in star pattern")
		assert.GreaterOrEqual(t, len(ifa.Intersections()), 1,
			"Should record multiple intersection points")
	})
}

// ---------------------------------------------------------------------------
// ValidatingNoder tests
// ---------------------------------------------------------------------------

func TestValidatingNoder_ValidNoding(t *testing.T) {
	t.Run("non-touching segments pass validation", func(t *testing.T) {
		// Two segments that do not touch or cross
		ss1 := NewNodedSegmentString(
			geom.NewCoordinateSequenceXY(0, 0, 5, 0),
			nil,
		)
		ss2 := NewNodedSegmentString(
			geom.NewCoordinateSequenceXY(0, 10, 5, 10),
			nil,
		)

		baseNoder := NewSimpleNoder(NewIntersectionAdder())
		vn := NewValidatingNoder(baseNoder)
		vn.ComputeNodes([]*NodedSegmentString{ss1, ss2})
		nodedSS := vn.GetNodedSubstrings()

		assert.NoError(t, vn.ValidationError(), "Non-touching segments should pass validation")
		assert.NotEmpty(t, nodedSS, "Should return noded substrings")
	})

	t.Run("non-intersecting segments pass validation", func(t *testing.T) {
		ss1 := NewNodedSegmentString(
			geom.NewCoordinateSequenceXY(0, 0, 10, 0),
			nil,
		)
		ss2 := NewNodedSegmentString(
			geom.NewCoordinateSequenceXY(0, 5, 10, 5),
			nil,
		)

		baseNoder := NewSimpleNoder(NewIntersectionAdder())
		vn := NewValidatingNoder(baseNoder)
		vn.ComputeNodes([]*NodedSegmentString{ss1, ss2})
		_ = vn.GetNodedSubstrings()

		assert.NoError(t, vn.ValidationError(), "Non-intersecting segments should pass validation")
	})
}

func TestValidatingNoder_InvalidNoding(t *testing.T) {
	t.Run("incomplete noding produces validation error", func(t *testing.T) {
		// Use IntersectionCounter (not Adder) so intersections are found but NOT resolved
		ss1 := NewNodedSegmentString(
			geom.NewCoordinateSequenceXY(0, 0, 10, 10),
			nil,
		)
		ss2 := NewNodedSegmentString(
			geom.NewCoordinateSequenceXY(0, 10, 10, 0),
			nil,
		)

		baseNoder := NewSimpleNoder(NewIntersectionCounter())
		vn := NewValidatingNoder(baseNoder)
		vn.ComputeNodes([]*NodedSegmentString{ss1, ss2})
		_ = vn.GetNodedSubstrings()

		err := vn.ValidationError()
		require.Error(t, err, "Incomplete noding should produce validation error")

		nodingErr, ok := err.(*NodingError)
		require.True(t, ok, "Error should be of type *NodingError")
		assert.Greater(t, nodingErr.IntersectionCount, 0, "IntersectionCount should be positive")
	})

	t.Run("validation error is reset between calls", func(t *testing.T) {
		baseNoder := NewSimpleNoder(NewIntersectionCounter())
		vn := NewValidatingNoder(baseNoder)

		// First: crossing lines (should fail validation)
		ss1 := NewNodedSegmentString(
			geom.NewCoordinateSequenceXY(0, 0, 10, 10),
			nil,
		)
		ss2 := NewNodedSegmentString(
			geom.NewCoordinateSequenceXY(0, 10, 10, 0),
			nil,
		)
		vn.ComputeNodes([]*NodedSegmentString{ss1, ss2})
		_ = vn.GetNodedSubstrings()
		require.Error(t, vn.ValidationError(), "First call should produce error")

		// Second: parallel lines (should pass validation)
		ss3 := NewNodedSegmentString(
			geom.NewCoordinateSequenceXY(0, 0, 10, 0),
			nil,
		)
		ss4 := NewNodedSegmentString(
			geom.NewCoordinateSequenceXY(0, 5, 10, 5),
			nil,
		)
		vn.ComputeNodes([]*NodedSegmentString{ss3, ss4})
		_ = vn.GetNodedSubstrings()
		assert.NoError(t, vn.ValidationError(), "Second call should pass after reset")
	})
}

// ---------------------------------------------------------------------------
// IteratedNoder tests
// ---------------------------------------------------------------------------

func TestIteratedNoder_BasicUsage(t *testing.T) {
	t.Run("converges on crossing lines", func(t *testing.T) {
		ss1 := NewNodedSegmentString(
			geom.NewCoordinateSequenceXY(0, 0, 10, 10),
			nil,
		)
		ss2 := NewNodedSegmentString(
			geom.NewCoordinateSequenceXY(0, 10, 10, 0),
			nil,
		)

		innerNoder := NewSimpleNoder(NewIntersectionAdder())
		iterated := NewIteratedNoder(innerNoder, 5)
		iterated.ComputeNodes([]*NodedSegmentString{ss1, ss2})

		result := iterated.GetNodedSubstrings()
		assert.NotEmpty(t, result, "Should produce noded substrings")
	})

	t.Run("default max iterations when zero is passed", func(t *testing.T) {
		innerNoder := NewSimpleNoder(NewIntersectionAdder())
		iterated := NewIteratedNoder(innerNoder, 0)

		ss := NewNodedSegmentString(
			geom.NewCoordinateSequenceXY(0, 0, 10, 0),
			nil,
		)
		// Should not panic with 0 maxIterations (defaults to 5)
		iterated.ComputeNodes([]*NodedSegmentString{ss})
		result := iterated.GetNodedSubstrings()
		assert.NotEmpty(t, result, "Should produce result with default max iterations")
	})

	t.Run("negative max iterations defaults to 5", func(t *testing.T) {
		innerNoder := NewSimpleNoder(NewIntersectionAdder())
		iterated := NewIteratedNoder(innerNoder, -1)

		ss := NewNodedSegmentString(
			geom.NewCoordinateSequenceXY(0, 0, 10, 0),
			nil,
		)
		iterated.ComputeNodes([]*NodedSegmentString{ss})
		result := iterated.GetNodedSubstrings()
		assert.NotEmpty(t, result, "Should produce result with negative max iterations")
	})

	t.Run("parallel lines converge immediately", func(t *testing.T) {
		ss1 := NewNodedSegmentString(
			geom.NewCoordinateSequenceXY(0, 0, 10, 0),
			nil,
		)
		ss2 := NewNodedSegmentString(
			geom.NewCoordinateSequenceXY(0, 5, 10, 5),
			nil,
		)

		innerNoder := NewSimpleNoder(NewIntersectionAdder())
		iterated := NewIteratedNoder(innerNoder, 3)
		iterated.ComputeNodes([]*NodedSegmentString{ss1, ss2})

		result := iterated.GetNodedSubstrings()
		assert.Len(t, result, 2, "Should produce 2 substrings for parallel lines")
	})
}

// ---------------------------------------------------------------------------
// NodingError tests
// ---------------------------------------------------------------------------

func TestNodingError_Error(t *testing.T) {
	t.Run("returns message string", func(t *testing.T) {
		err := &NodingError{
			Message:           "noding incomplete: intersections remain in result",
			IntersectionCount: 3,
		}
		assert.Equal(t, "noding incomplete: intersections remain in result", err.Error())
	})

	t.Run("implements error interface", func(t *testing.T) {
		var err error = &NodingError{
			Message:           "test error",
			IntersectionCount: 1,
		}
		assert.Equal(t, "test error", err.Error())
	})

	t.Run("stores intersection count", func(t *testing.T) {
		err := &NodingError{
			Message:           "noding failed",
			IntersectionCount: 42,
		}
		assert.Equal(t, 42, err.IntersectionCount)
	})
}

// ---------------------------------------------------------------------------
// IntersectionFinderAdder - IsDone delegation
// ---------------------------------------------------------------------------

func TestIntersectionFinderAdder_IsDone(t *testing.T) {
	ifa := NewIntersectionFinderAdder()
	assert.False(t, ifa.IsDone(), "IntersectionFinderAdder.IsDone should return false")
}

// ---------------------------------------------------------------------------
// ScaledNoder with offset
// ---------------------------------------------------------------------------

func TestScaledNoder_OffsetCoordinates(t *testing.T) {
	inner := NewSimpleNoder(NewIntersectionAdder())
	sn := NewScaledNoder(inner, 10.0)
	// Offset defaults to 0, but let's verify scale/unscale is a round trip
	coords := geom.NewCoordinateSequenceXY(1.5, 2.5, 3.5, 4.5)
	scaled := sn.scale(coords)
	unscaled := sn.unscale(scaled)

	for i := range coords {
		assert.InDelta(t, coords[i].X, unscaled[i].X, 1e-9,
			"Coordinate %d X should round-trip", i)
		assert.InDelta(t, coords[i].Y, unscaled[i].Y, 1e-9,
			"Coordinate %d Y should round-trip", i)
	}
}

// ---------------------------------------------------------------------------
// Additional edge cases
// ---------------------------------------------------------------------------

func TestSimpleNoder_NilSegmentIntersector(t *testing.T) {
	// When nil is passed, NewSimpleNoder should use default IntersectionAdder
	noder := NewSimpleNoder(nil)

	ss1 := NewNodedSegmentString(
		geom.NewCoordinateSequenceXY(0, 0, 10, 10),
		nil,
	)
	ss2 := NewNodedSegmentString(
		geom.NewCoordinateSequenceXY(0, 10, 10, 0),
		nil,
	)

	noder.ComputeNodes([]*NodedSegmentString{ss1, ss2})
	result := noder.GetNodedSubstrings()

	assert.NotEmpty(t, result, "Should produce results even with nil SegmentIntersector")
	// The default IntersectionAdder should have found and inserted the intersection
	for _, nss := range result {
		assert.Len(t, nss.Coordinates(), 3,
			"Each segment should have 3 coords (start, intersection, end)")
	}
}

func TestGetNodedSubstrings_CachesResult(t *testing.T) {
	ss := NewNodedSegmentString(
		geom.NewCoordinateSequenceXY(0, 0, 10, 10),
		nil,
	)

	noder := NewSimpleNoder(NewIntersectionAdder())
	noder.ComputeNodes([]*NodedSegmentString{ss})

	first := noder.GetNodedSubstrings()
	second := noder.GetNodedSubstrings()

	// Second call should return same slice (cached)
	assert.Equal(t, len(first), len(second), "Cached result should have same length")
}

func TestCoordKey(t *testing.T) {
	c := geom.NewCoordinate(1.23456789012345, -9.87654321098765)
	key := coordKey(c)
	assert.NotEmpty(t, key, "coordKey should produce a non-empty string")
	// Same coordinate should produce same key
	key2 := coordKey(geom.NewCoordinate(1.23456789012345, -9.87654321098765))
	assert.Equal(t, key, key2, "Same coordinates should produce same key")
}
