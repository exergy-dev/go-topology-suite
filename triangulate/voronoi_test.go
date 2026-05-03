package triangulate

import (
	"math"
	"testing"

	"github.com/terra-geo/terra/geom"
)

func TestVoronoi_FourCorners(t *testing.T) {
	// Four corners of a unit square. Each site owns one quadrant. We clip
	// to the unit square so each cell becomes exactly a quarter of it.
	pts := []geom.XY{
		{X: 0, Y: 0},
		{X: 1, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 1},
	}
	clip := geom.Envelope{MinX: 0, MinY: 0, MaxX: 1, MaxY: 1}
	cells := Voronoi(pts, &clip)
	if len(cells) != 4 {
		t.Fatalf("want 4 Voronoi cells, got %d", len(cells))
	}
	// Each cell should have area 0.25.
	totalArea := 0.0
	for _, p := range cells {
		ring := p.Ring(0)
		a := math.Abs(signedRingArea(ring))
		if math.Abs(a-0.25) > 1e-9 {
			t.Fatalf("cell area: want 0.25, got %v (%v)", a, ring)
		}
		totalArea += a
	}
	if math.Abs(totalArea-1.0) > 1e-9 {
		t.Fatalf("total Voronoi area: want 1.0, got %v", totalArea)
	}
}

func TestVoronoi_SingleSite(t *testing.T) {
	// A single site degenerate case has no Delaunay edges to dual.
	cells := Voronoi([]geom.XY{{X: 0, Y: 0}}, nil)
	if cells != nil {
		t.Fatalf("expected nil for single site, got %d cells", len(cells))
	}
}

func TestVoronoi_NoClipBox(t *testing.T) {
	// Without an explicit clip box we still get one cell per site, bounded
	// by the auto-expanded frame.
	pts := []geom.XY{
		{X: 0, Y: 0},
		{X: 1, Y: 0},
		{X: 0, Y: 1},
	}
	cells := Voronoi(pts, nil)
	if len(cells) != 3 {
		t.Fatalf("want 3 cells, got %d", len(cells))
	}
}

func TestVoronoi_CellsCoverClipBox(t *testing.T) {
	// Random sites — total clipped cell area must equal the clip box
	// (cells partition the box up to robustness noise).
	pts := randomPoints(20, 7)
	clip := geom.Envelope{MinX: -5, MinY: -5, MaxX: 105, MaxY: 105}
	cells := Voronoi(pts, &clip)
	if len(cells) == 0 {
		t.Fatal("expected some cells")
	}
	total := 0.0
	for _, p := range cells {
		total += math.Abs(signedRingArea(p.Ring(0)))
	}
	want := clip.Width() * clip.Height()
	if math.Abs(total-want) > 1e-6 {
		t.Fatalf("total cell area %v, want %v (clip box area)", total, want)
	}
}

// signedRingArea returns the signed shoelace area of an XY ring (which
// may or may not be explicitly closed).
func signedRingArea(ring []geom.XY) float64 {
	n := len(ring)
	if n < 3 {
		return 0
	}
	if ring[0] == ring[n-1] {
		n--
	}
	var s float64
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		s += ring[i].X*ring[j].Y - ring[j].X*ring[i].Y
	}
	return s / 2
}
