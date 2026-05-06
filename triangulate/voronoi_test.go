package triangulate

import (
	"math"
	"testing"

	"github.com/exergy-dev/go-topology-suite/geom"
	"github.com/exergy-dev/go-topology-suite/kernel/planar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.Equal(t, 4, len(cells))
	// Each cell should have area 0.25.
	totalArea := 0.0
	for _, p := range cells {
		ring := p.Ring(0)
		a := math.Abs((planar.Kernel{}).RingArea(ring))
		assert.InDeltaf(t, 0.25, a, 1e-9, "cell area: want 0.25, got %v (%v)", a, ring)
		totalArea += a
	}
	assert.InDelta(t, 1.0, totalArea, 1e-9)
}

func TestVoronoi_SingleSite(t *testing.T) {
	// A single site degenerate case has no Delaunay edges to dual.
	cells := Voronoi([]geom.XY{{X: 0, Y: 0}}, nil)
	require.Nilf(t, cells, "expected nil for single site, got %d cells", len(cells))
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
	require.Equal(t, 3, len(cells))
}

func TestVoronoi_CellsCoverClipBox(t *testing.T) {
	// Random sites — total clipped cell area must equal the clip box
	// (cells partition the box up to robustness noise).
	pts := randomPoints(20, 7)
	clip := geom.Envelope{MinX: -5, MinY: -5, MaxX: 105, MaxY: 105}
	cells := Voronoi(pts, &clip)
	require.NotEmpty(t, cells, "expected some cells")
	total := 0.0
	for _, p := range cells {
		total += math.Abs((planar.Kernel{}).RingArea(p.Ring(0)))
	}
	want := clip.Width() * clip.Height()
	assert.InDeltaf(t, want, total, 1e-6, "total cell area %v, want %v (clip box area)", total, want)
}
