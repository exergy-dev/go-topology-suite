package linemerge

import (
	"fmt"
	"testing"

	"github.com/go-topology-suite/gts/geom"
)

// TestLineMergerEmpty tests merging with no input
func TestLineMergerEmpty(t *testing.T) {
	merger := NewLineMerger()
	result := merger.GetMergedLineStrings()

	if len(result) != 0 {
		t.Errorf("Expected 0 merged lines, got %d", len(result))
	}
}

// TestLineMergerSingleLine tests merging a single line
func TestLineMergerSingleLine(t *testing.T) {
	line := geom.NewLineStringXY(0, 0, 1, 1, 2, 2)
	merger := NewLineMerger()
	merger.Add(line)
	result := merger.GetMergedLineStrings()

	if len(result) != 1 {
		t.Fatalf("Expected 1 merged line, got %d", len(result))
	}

	if !result[0].EqualsExact(line, 1e-10) {
		t.Errorf("Expected line to be unchanged")
	}
}

// TestLineMergerTwoConnectedLines tests merging two lines that share an endpoint
func TestLineMergerTwoConnectedLines(t *testing.T) {
	line1 := geom.NewLineStringXY(0, 0, 1, 1)
	line2 := geom.NewLineStringXY(1, 1, 2, 2)

	merger := NewLineMerger()
	merger.Add(line1)
	merger.Add(line2)
	result := merger.GetMergedLineStrings()

	if len(result) != 1 {
		t.Fatalf("Expected 1 merged line, got %d", len(result))
	}

	expected := geom.NewLineStringXY(0, 0, 1, 1, 2, 2)
	if !result[0].EqualsExact(expected, 1e-10) {
		t.Errorf("Expected merged line to be %v, got %v", expected, result[0])
	}
}

// TestLineMergerThreeConnectedLines tests merging three lines in sequence
func TestLineMergerThreeConnectedLines(t *testing.T) {
	line1 := geom.NewLineStringXY(0, 0, 1, 0)
	line2 := geom.NewLineStringXY(1, 0, 2, 0)
	line3 := geom.NewLineStringXY(2, 0, 3, 0)

	merger := NewLineMerger()
	merger.Add(line1)
	merger.Add(line2)
	merger.Add(line3)
	result := merger.GetMergedLineStrings()

	if len(result) != 1 {
		t.Fatalf("Expected 1 merged line, got %d", len(result))
	}

	expected := geom.NewLineStringXY(0, 0, 1, 0, 2, 0, 3, 0)
	if !result[0].EqualsExact(expected, 1e-10) {
		t.Errorf("Expected merged line to be %v, got %v", expected, result[0])
	}
}

// TestLineMergerDisconnectedLines tests lines that don't connect
func TestLineMergerDisconnectedLines(t *testing.T) {
	line1 := geom.NewLineStringXY(0, 0, 1, 0)
	line2 := geom.NewLineStringXY(2, 0, 3, 0)

	merger := NewLineMerger()
	merger.Add(line1)
	merger.Add(line2)
	result := merger.GetMergedLineStrings()

	if len(result) != 2 {
		t.Fatalf("Expected 2 separate lines, got %d", len(result))
	}
}

// TestLineMergerReversedLines tests merging lines that need to be reversed
func TestLineMergerReversedLines(t *testing.T) {
	line1 := geom.NewLineStringXY(0, 0, 1, 1)
	line2 := geom.NewLineStringXY(2, 2, 1, 1) // Reversed

	merger := NewLineMerger()
	merger.Add(line1)
	merger.Add(line2)
	result := merger.GetMergedLineStrings()

	if len(result) != 1 {
		t.Fatalf("Expected 1 merged line, got %d", len(result))
	}

	expected := geom.NewLineStringXY(0, 0, 1, 1, 2, 2)
	if !result[0].EqualsExact(expected, 1e-10) {
		t.Errorf("Expected merged line to be %v, got %v", expected, result[0])
	}
}

// TestLineMergerClosedRing tests a closed ring
func TestLineMergerClosedRing(t *testing.T) {
	line1 := geom.NewLineStringXY(0, 0, 1, 0)
	line2 := geom.NewLineStringXY(1, 0, 1, 1)
	line3 := geom.NewLineStringXY(1, 1, 0, 1)
	line4 := geom.NewLineStringXY(0, 1, 0, 0)

	merger := NewLineMerger()
	merger.Add(line1)
	merger.Add(line2)
	merger.Add(line3)
	merger.Add(line4)
	result := merger.GetMergedLineStrings()

	if len(result) != 1 {
		t.Fatalf("Expected 1 merged line (closed ring), got %d", len(result))
	}

	// The result should be a closed ring
	if !result[0].IsClosed() {
		t.Errorf("Expected result to be a closed ring")
	}

	expected := geom.NewLineStringXY(0, 0, 1, 0, 1, 1, 0, 1, 0, 0)
	if !result[0].EqualsExact(expected, 1e-10) {
		t.Errorf("Expected merged line to be %v, got %v", expected, result[0])
	}
}

// TestLineMergerBranchingPoint tests a branching point where 3 lines meet
func TestLineMergerBranchingPoint(t *testing.T) {
	// Three lines meeting at (1, 1)
	line1 := geom.NewLineStringXY(0, 0, 1, 1)
	line2 := geom.NewLineStringXY(1, 1, 2, 2)
	line3 := geom.NewLineStringXY(1, 1, 2, 0)

	merger := NewLineMerger()
	merger.Add(line1)
	merger.Add(line2)
	merger.Add(line3)
	result := merger.GetMergedLineStrings()

	// Should have 3 separate lines because of the branch
	if len(result) != 3 {
		t.Fatalf("Expected 3 lines (branching prevents merge), got %d", len(result))
	}
}

// TestLineMergerMultipleSequences tests multiple separate sequences
func TestLineMergerMultipleSequences(t *testing.T) {
	// First sequence
	line1 := geom.NewLineStringXY(0, 0, 1, 0)
	line2 := geom.NewLineStringXY(1, 0, 2, 0)

	// Second sequence
	line3 := geom.NewLineStringXY(0, 5, 1, 5)
	line4 := geom.NewLineStringXY(1, 5, 2, 5)

	merger := NewLineMerger()
	merger.Add(line1)
	merger.Add(line2)
	merger.Add(line3)
	merger.Add(line4)
	result := merger.GetMergedLineStrings()

	if len(result) != 2 {
		t.Fatalf("Expected 2 merged sequences, got %d", len(result))
	}

	// Each result should be a line with 3 points
	for i, line := range result {
		if line.NumPoints() != 3 {
			t.Errorf("Sequence %d: expected 3 points, got %d", i, line.NumPoints())
		}
	}
}

// TestLineMergerMultiLineString tests merging from a MultiLineString
func TestLineMergerMultiLineString(t *testing.T) {
	line1 := geom.NewLineStringXY(0, 0, 1, 1)
	line2 := geom.NewLineStringXY(1, 1, 2, 2)
	mls := geom.NewMultiLineString([]*geom.LineString{line1, line2})

	merger := NewLineMerger()
	merger.AddMultiLineString(mls)
	result := merger.GetMergedLineStrings()

	if len(result) != 1 {
		t.Fatalf("Expected 1 merged line, got %d", len(result))
	}

	expected := geom.NewLineStringXY(0, 0, 1, 1, 2, 2)
	if !result[0].EqualsExact(expected, 1e-10) {
		t.Errorf("Expected merged line to be %v, got %v", expected, result[0])
	}
}

// TestLineMergerComplexSequence tests a more complex sequence
func TestLineMergerComplexSequence(t *testing.T) {
	// Create a zigzag path that should merge into one line
	line1 := geom.NewLineStringXY(0, 0, 1, 1, 2, 0)
	line2 := geom.NewLineStringXY(2, 0, 3, 1, 4, 0)
	line3 := geom.NewLineStringXY(4, 0, 5, 1)

	merger := NewLineMerger()
	merger.Add(line1)
	merger.Add(line2)
	merger.Add(line3)
	result := merger.GetMergedLineStrings()

	if len(result) != 1 {
		t.Fatalf("Expected 1 merged line, got %d", len(result))
	}

	expected := geom.NewLineStringXY(0, 0, 1, 1, 2, 0, 3, 1, 4, 0, 5, 1)
	if !result[0].EqualsExact(expected, 1e-10) {
		t.Errorf("Expected merged line to be %v, got %v", expected, result[0])
	}
}

// TestLineMergerOutOfOrder tests that lines can be added in any order
func TestLineMergerOutOfOrder(t *testing.T) {
	line1 := geom.NewLineStringXY(0, 0, 1, 0)
	line2 := geom.NewLineStringXY(1, 0, 2, 0)
	line3 := geom.NewLineStringXY(2, 0, 3, 0)

	// Add in reverse order
	merger := NewLineMerger()
	merger.Add(line3)
	merger.Add(line1)
	merger.Add(line2)
	result := merger.GetMergedLineStrings()

	if len(result) != 1 {
		t.Fatalf("Expected 1 merged line, got %d", len(result))
	}

	expected := geom.NewLineStringXY(0, 0, 1, 0, 2, 0, 3, 0)
	if !result[0].EqualsExact(expected, 1e-10) {
		t.Errorf("Expected merged line to be %v, got %v", expected, result[0])
	}
}

// TestLineMergerDuplicateLines tests handling of duplicate lines
func TestLineMergerDuplicateLines(t *testing.T) {
	line1 := geom.NewLineStringXY(0, 0, 1, 0)
	line2 := geom.NewLineStringXY(0, 0, 1, 0) // Duplicate

	merger := NewLineMerger()
	merger.Add(line1)
	merger.Add(line2)
	result := merger.GetMergedLineStrings()

	// Duplicates should be kept separate as they create a branching point
	if len(result) < 1 {
		t.Fatalf("Expected at least 1 line, got %d", len(result))
	}
}

// TestMergeLineStringsConvenience tests the convenience function
func TestMergeLineStringsConvenience(t *testing.T) {
	line1 := geom.NewLineStringXY(0, 0, 1, 1)
	line2 := geom.NewLineStringXY(1, 1, 2, 2)

	result := MergeLineStrings([]*geom.LineString{line1, line2})

	if len(result) != 1 {
		t.Fatalf("Expected 1 merged line, got %d", len(result))
	}

	expected := geom.NewLineStringXY(0, 0, 1, 1, 2, 2)
	if !result[0].EqualsExact(expected, 1e-10) {
		t.Errorf("Expected merged line to be %v, got %v", expected, result[0])
	}
}

// TestMergeMultiLineStringConvenience tests the MultiLineString convenience function
func TestMergeMultiLineStringConvenience(t *testing.T) {
	line1 := geom.NewLineStringXY(0, 0, 1, 1)
	line2 := geom.NewLineStringXY(1, 1, 2, 2)
	mls := geom.NewMultiLineString([]*geom.LineString{line1, line2})

	result := MergeMultiLineString(mls)

	if result.NumGeometries() != 1 {
		t.Fatalf("Expected 1 merged line, got %d", result.NumGeometries())
	}

	expected := geom.NewLineStringXY(0, 0, 1, 1, 2, 2)
	if !result.LineStringN(0).EqualsExact(expected, 1e-10) {
		t.Errorf("Expected merged line to be %v, got %v", expected, result.LineStringN(0))
	}
}

// TestLineMergerEmptyLines tests handling of empty LineStrings
func TestLineMergerEmptyLines(t *testing.T) {
	line1 := geom.NewLineStringXY(0, 0, 1, 1)
	emptyLine := geom.NewLineStringEmpty()
	line2 := geom.NewLineStringXY(1, 1, 2, 2)

	merger := NewLineMerger()
	merger.Add(line1)
	merger.Add(emptyLine)
	merger.Add(line2)
	result := merger.GetMergedLineStrings()

	if len(result) != 1 {
		t.Fatalf("Expected 1 merged line (empty ignored), got %d", len(result))
	}

	expected := geom.NewLineStringXY(0, 0, 1, 1, 2, 2)
	if !result[0].EqualsExact(expected, 1e-10) {
		t.Errorf("Expected merged line to be %v, got %v", expected, result[0])
	}
}

// TestLineMergerNilLines tests handling of nil LineStrings
func TestLineMergerNilLines(t *testing.T) {
	line1 := geom.NewLineStringXY(0, 0, 1, 1)
	line2 := geom.NewLineStringXY(1, 1, 2, 2)

	merger := NewLineMerger()
	merger.Add(line1)
	merger.Add(nil)
	merger.Add(line2)
	result := merger.GetMergedLineStrings()

	if len(result) != 1 {
		t.Fatalf("Expected 1 merged line (nil ignored), got %d", len(result))
	}

	expected := geom.NewLineStringXY(0, 0, 1, 1, 2, 2)
	if !result[0].EqualsExact(expected, 1e-10) {
		t.Errorf("Expected merged line to be %v, got %v", expected, result[0])
	}
}

// TestLineMergerTShapeJunction tests a T-junction (3 lines meeting)
func TestLineMergerTShapeJunction(t *testing.T) {
	// Horizontal line split at middle
	line1 := geom.NewLineStringXY(0, 1, 1, 1)
	line2 := geom.NewLineStringXY(1, 1, 2, 1)
	// Vertical line connecting at middle
	line3 := geom.NewLineStringXY(1, 0, 1, 1)

	merger := NewLineMerger()
	merger.Add(line1)
	merger.Add(line2)
	merger.Add(line3)
	result := merger.GetMergedLineStrings()

	// The T-junction prevents full merging - should get 3 lines
	if len(result) != 3 {
		t.Fatalf("Expected 3 lines (T-junction prevents merge), got %d", len(result))
	}
}

// TestLineMergerCrossJunction tests a cross junction (4 lines meeting)
func TestLineMergerCrossJunction(t *testing.T) {
	// Four lines meeting at (1, 1)
	line1 := geom.NewLineStringXY(0, 1, 1, 1)
	line2 := geom.NewLineStringXY(1, 1, 2, 1)
	line3 := geom.NewLineStringXY(1, 0, 1, 1)
	line4 := geom.NewLineStringXY(1, 1, 1, 2)

	merger := NewLineMerger()
	merger.Add(line1)
	merger.Add(line2)
	merger.Add(line3)
	merger.Add(line4)
	result := merger.GetMergedLineStrings()

	// The cross junction prevents merging - should get 4 lines
	if len(result) != 4 {
		t.Fatalf("Expected 4 lines (cross junction prevents merge), got %d", len(result))
	}
}

// TestLineMergerLongChain tests a long chain of connected lines
func TestLineMergerLongChain(t *testing.T) {
	merger := NewLineMerger()

	// Create a chain of 10 connected line segments
	for i := 0; i < 10; i++ {
		line := geom.NewLineStringXY(float64(i), 0, float64(i+1), 0)
		merger.Add(line)
	}

	result := merger.GetMergedLineStrings()

	if len(result) != 1 {
		t.Fatalf("Expected 1 merged line, got %d", len(result))
	}

	if result[0].NumPoints() != 11 {
		t.Errorf("Expected 11 points in merged line, got %d", result[0].NumPoints())
	}

	// Check start and end points
	start := result[0].StartPoint()
	end := result[0].EndPoint()

	if start.X() != 0 || start.Y() != 0 {
		t.Errorf("Expected start point (0, 0), got (%g, %g)", start.X(), start.Y())
	}

	if end.X() != 10 || end.Y() != 0 {
		t.Errorf("Expected end point (10, 0), got (%g, %g)", end.X(), end.Y())
	}
}

// TestLineMergerMultipleCalls tests calling GetMergedLineStrings multiple times
func TestLineMergerMultipleCalls(t *testing.T) {
	line1 := geom.NewLineStringXY(0, 0, 1, 1)
	line2 := geom.NewLineStringXY(1, 1, 2, 2)

	merger := NewLineMerger()
	merger.Add(line1)
	merger.Add(line2)

	result1 := merger.GetMergedLineStrings()
	result2 := merger.GetMergedLineStrings()

	// Should return the same result
	if len(result1) != len(result2) {
		t.Errorf("Multiple calls returned different number of lines")
	}

	if len(result1) > 0 && len(result2) > 0 {
		if !result1[0].EqualsExact(result2[0], 1e-10) {
			t.Errorf("Multiple calls returned different results")
		}
	}
}

// TestLineMergerSelfLoop tests a line that forms a self-loop
func TestLineMergerSelfLoop(t *testing.T) {
	// A closed ring as a single line
	ring := geom.NewLineStringXY(0, 0, 1, 0, 1, 1, 0, 1, 0, 0)

	merger := NewLineMerger()
	merger.Add(ring)
	result := merger.GetMergedLineStrings()

	if len(result) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(result))
	}

	if !result[0].IsClosed() {
		t.Errorf("Expected result to be closed")
	}

	if !result[0].EqualsExact(ring, 1e-10) {
		t.Errorf("Expected ring to remain unchanged")
	}
}

// Example_basicMerge demonstrates basic line merging.
func Example_basicMerge() {
	// Create two connected line segments
	line1 := geom.NewLineStringXY(0, 0, 1, 1)
	line2 := geom.NewLineStringXY(1, 1, 2, 2)

	// Merge them
	merger := NewLineMerger()
	merger.Add(line1)
	merger.Add(line2)
	result := merger.GetMergedLineStrings()

	fmt.Println(result[0])
	// Output: LINESTRING (0 0, 1 1, 2 2)
}

// Example_multipleSequences demonstrates merging multiple disconnected sequences.
func Example_multipleSequences() {
	// Create two separate sequences
	line1 := geom.NewLineStringXY(0, 0, 1, 0)
	line2 := geom.NewLineStringXY(1, 0, 2, 0)
	line3 := geom.NewLineStringXY(0, 5, 1, 5)
	line4 := geom.NewLineStringXY(1, 5, 2, 5)

	// Merge them
	result := MergeLineStrings([]*geom.LineString{line1, line2, line3, line4})

	fmt.Printf("Number of merged sequences: %d\n", len(result))
	for i, line := range result {
		fmt.Printf("Sequence %d: %d points\n", i+1, line.NumPoints())
	}
	// Output:
	// Number of merged sequences: 2
	// Sequence 1: 3 points
	// Sequence 2: 3 points
}

// Example_closedRing demonstrates merging lines into a closed ring.
func Example_closedRing() {
	// Create four lines that form a square
	line1 := geom.NewLineStringXY(0, 0, 1, 0)
	line2 := geom.NewLineStringXY(1, 0, 1, 1)
	line3 := geom.NewLineStringXY(1, 1, 0, 1)
	line4 := geom.NewLineStringXY(0, 1, 0, 0)

	// Merge them
	result := MergeLineStrings([]*geom.LineString{line1, line2, line3, line4})

	fmt.Printf("Number of merged lines: %d\n", len(result))
	fmt.Printf("Is closed: %v\n", result[0].IsClosed())
	// Output:
	// Number of merged lines: 1
	// Is closed: true
}
