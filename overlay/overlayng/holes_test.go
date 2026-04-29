package overlayng

import (
	"math"
	"testing"

	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/measure"
)

// TestSubjectWithHoleIntersection: a 10×10 square with a 4×4 hole at
// the center, intersected with a 6×6 shifted square. The intersection
// must respect the hole.
//
//	subj outer: (0,0)..(10,10)            area 100
//	subj hole : (3,3)..(7,7)              area 16
//	subj total: 84
//	clip      : (4,4)..(10,10)            area 36
//	subj∩clip : the 6×6 clip minus the part of the hole inside the clip.
//	            Hole inside clip is (4,4)..(7,7) → 9.
//	            Result area: 36 - 9 = 27.
func TestSubjectWithHoleIntersection(t *testing.T) {
	outer := []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}, {X: 0, Y: 10}, {X: 0, Y: 0}}
	hole := []geom.XY{{X: 3, Y: 3}, {X: 3, Y: 7}, {X: 7, Y: 7}, {X: 7, Y: 3}, {X: 3, Y: 3}}
	subj := geom.NewPolygon(nil, outer, hole)
	clip := geom.NewPolygon(nil, []geom.XY{
		{X: 4, Y: 4}, {X: 10, Y: 4}, {X: 10, Y: 10}, {X: 4, Y: 10}, {X: 4, Y: 4},
	})
	first, rest, err := Overlay(subj, clip, OpIntersection)
	if err != nil {
		t.Fatal(err)
	}
	total := measure.Area(first)
	for _, p := range rest {
		total += measure.Area(p)
	}
	if math.Abs(total-27) > 1e-9 {
		t.Errorf("intersection area = %v, want 27", total)
	}
}

// TestUnionWithHole: the same subject (with hole) unioned with a
// disjoint small square. Result area: 84 + small. We use a 1×1 small
// square outside the subject for area = 84 + 1 = 85.
func TestUnionWithHole(t *testing.T) {
	outer := []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}, {X: 0, Y: 10}, {X: 0, Y: 0}}
	hole := []geom.XY{{X: 3, Y: 3}, {X: 3, Y: 7}, {X: 7, Y: 7}, {X: 7, Y: 3}, {X: 3, Y: 3}}
	subj := geom.NewPolygon(nil, outer, hole)
	other := geom.NewPolygon(nil, []geom.XY{
		{X: 20, Y: 20}, {X: 21, Y: 20}, {X: 21, Y: 21}, {X: 20, Y: 21}, {X: 20, Y: 20},
	})
	first, rest, err := Overlay(subj, other, OpUnion)
	if err != nil {
		t.Fatal(err)
	}
	total := measure.Area(first)
	for _, p := range rest {
		total += measure.Area(p)
	}
	want := 84.0 + 1.0
	if math.Abs(total-want) > 1e-9 {
		t.Errorf("union area = %v, want %v", total, want)
	}
}

// TestDifferenceCreatesHole: a 10×10 square minus an interior 4×4 square.
// The expected result is the outer square WITH a hole. Verify that the
// returned polygon has 2 rings and the area equals 84.
func TestDifferenceCreatesHole(t *testing.T) {
	subj := geom.NewPolygon(nil, []geom.XY{
		{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}, {X: 0, Y: 10}, {X: 0, Y: 0},
	})
	hole := geom.NewPolygon(nil, []geom.XY{
		{X: 3, Y: 3}, {X: 7, Y: 3}, {X: 7, Y: 7}, {X: 3, Y: 7}, {X: 3, Y: 3},
	})
	first, _, err := Overlay(subj, hole, OpDifference)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(measure.Area(first)-84) > 1e-9 {
		t.Errorf("difference area = %v, want 84", measure.Area(first))
	}
	// The result should be a polygon with a hole — assemble.go classifies
	// the inner square (which is the smaller-area kept-region boundary)
	// as a hole.
	// Note: the difference is "subj minus clip" = subj outer with clip
	// boundary as the hole. Either path through the assembler must give
	// a polygon with NumRings() == 2, OR a polygon with NumRings() == 1
	// at area 100 (pure outer) plus an additional polygon-as-hole of 16
	// (which we'd then need to merge). The valid interpretation depends
	// on how `assemble.go` orders the rings; assert NumRings == 2 as the
	// production-quality outcome.
	if first.NumRings() != 2 {
		t.Errorf("difference result should have 2 rings (outer+hole), got %d", first.NumRings())
	}
}

// TestHoleInputProducingHoleOutput: subj is a 10x10 outer square with
// a 4x4 hole; clip is a 6x6 square overlapping the upper-right corner.
// A\B (subj minus clip) should: (a) preserve the hole that's outside
// clip, and (b) reshape the result around the carved-out region. The
// union of the inner+outer rings of the result should total ~84 - 27 = 57.
func TestHoleInputProducingHoleOutput(t *testing.T) {
	subj := geom.NewPolygon(nil,
		[]geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}, {X: 0, Y: 10}, {X: 0, Y: 0}},
		[]geom.XY{{X: 3, Y: 3}, {X: 3, Y: 7}, {X: 7, Y: 7}, {X: 7, Y: 3}, {X: 3, Y: 3}},
	)
	clip := geom.NewPolygon(nil, []geom.XY{
		{X: 4, Y: 4}, {X: 10, Y: 4}, {X: 10, Y: 10}, {X: 4, Y: 10}, {X: 4, Y: 4},
	})
	first, rest, err := Overlay(subj, clip, OpDifference)
	if err != nil {
		t.Fatal(err)
	}
	total := measure.Area(first)
	for _, p := range rest {
		total += measure.Area(p)
	}
	// subj area = 100 - 16 = 84
	// clip ∩ subj = 27 (computed in TestSubjectWithHoleIntersection)
	// subj \ clip = subj - (subj ∩ clip) = 84 - 27 = 57
	want := 57.0
	if math.Abs(total-want) > 0.5 {
		t.Errorf("difference area = %v, want ≈ %v", total, want)
	}
}

// TestBothInputsHaveHoles: subj is a 10x10 with hole at (3,3..7,7),
// clip is a 10x10 (offset by (5,0)) with a hole at (8,3..12,7). Their
// intersection should reflect both holes correctly.
func TestBothInputsHaveHoles(t *testing.T) {
	subjOuter := []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}, {X: 0, Y: 10}, {X: 0, Y: 0}}
	subjHole := []geom.XY{{X: 3, Y: 3}, {X: 3, Y: 7}, {X: 7, Y: 7}, {X: 7, Y: 3}, {X: 3, Y: 3}}
	subj := geom.NewPolygon(nil, subjOuter, subjHole)

	clipOuter := []geom.XY{{X: 5, Y: 0}, {X: 15, Y: 0}, {X: 15, Y: 10}, {X: 5, Y: 10}, {X: 5, Y: 0}}
	clipHole := []geom.XY{{X: 8, Y: 3}, {X: 8, Y: 7}, {X: 12, Y: 7}, {X: 12, Y: 3}, {X: 8, Y: 3}}
	clip := geom.NewPolygon(nil, clipOuter, clipHole)

	first, rest, err := Overlay(subj, clip, OpIntersection)
	if err != nil {
		t.Fatal(err)
	}
	total := measure.Area(first)
	for _, p := range rest {
		total += measure.Area(p)
	}
	// Intersection of outers: 5×10 = 50.
	// Subj hole inside that: (5,3..7,7) → 8.
	// Clip hole inside that: (8,3..10,7) → wait, clip hole is x ∈ [8,12] but
	// outer intersection is x ∈ [5,10], so hole-in-intersection is x ∈ [8,10] → 8.
	// Total intersection area: 50 - 8 - 8 = 34.
	want := 34.0
	if math.Abs(total-want) > 0.5 {
		t.Errorf("intersection area = %v, want ≈ %v", total, want)
	}
}
