package snap

import (
	"testing"

	"github.com/terra-geo/terra/geom"
)

func TestHotPixelIndex_AddRoundsAndDedupes(t *testing.T) {
	idx := NewHotPixelIndex(1.0)
	a := idx.Add(geom.XY{X: 1.4, Y: 0.7})
	b := idx.Add(geom.XY{X: 1.49, Y: 0.51})
	if a != b {
		t.Fatalf("near-duplicate not deduped: a=%p b=%p", a, b)
	}
	if !b.IsNode {
		t.Errorf("second add should mark pixel as node")
	}
	if got := idx.Len(); got != 1 {
		t.Errorf("Len = %d, want 1", got)
	}
}

func TestHotPixelIndex_QuerySegment(t *testing.T) {
	idx := NewHotPixelIndex(1.0)
	for _, p := range []geom.XY{
		{X: 0, Y: 0}, {X: 1, Y: 1}, {X: 2, Y: 2},
		{X: 5, Y: 5}, {X: -3, Y: 4},
	} {
		idx.Add(p)
	}
	var got []geom.XY
	idx.QuerySegment(geom.XY{X: 0, Y: 0}, geom.XY{X: 2, Y: 2}, func(e *HotPixelEntry) {
		got = append(got, e.Centre)
	})
	want := map[geom.XY]bool{
		{X: 0, Y: 0}: true, {X: 1, Y: 1}: true, {X: 2, Y: 2}: true,
	}
	if len(got) < len(want) {
		t.Errorf("QuerySegment returned %d entries, want >= %d (got=%v)", len(got), len(want), got)
	}
	for _, c := range got {
		delete(want, c)
	}
	if len(want) != 0 {
		t.Errorf("missing centres %v", want)
	}
}

func TestHotPixelIndex_AddNodes(t *testing.T) {
	idx := NewHotPixelIndex(1.0)
	idx.AddNodes([]geom.XY{{X: 0, Y: 0}, {X: 5, Y: 5}})
	pixels := idx.Pixels()
	if len(pixels) != 2 {
		t.Fatalf("got %d pixels, want 2", len(pixels))
	}
	for _, p := range pixels {
		if !p.IsNode {
			t.Errorf("pixel %v not marked as node", p.Centre)
		}
	}
}

func TestHotPixelIndex_AddShuffledDeterministic(t *testing.T) {
	pts := []geom.XY{}
	for i := 0; i < 20; i++ {
		pts = append(pts, geom.XY{X: float64(i), Y: float64(i)})
	}
	idx1 := NewHotPixelIndex(1.0)
	idx1.AddShuffled(pts)
	idx2 := NewHotPixelIndex(1.0)
	idx2.AddShuffled(pts)
	if idx1.Len() != idx2.Len() {
		t.Errorf("len mismatch %d vs %d", idx1.Len(), idx2.Len())
	}
	if idx1.Len() != len(pts) {
		t.Errorf("Len = %d, want %d", idx1.Len(), len(pts))
	}
}
