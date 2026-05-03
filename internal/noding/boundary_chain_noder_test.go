package noding

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/terra-geo/terra/geom"
)

func TestBoundaryChainNoder_TwoAdjacentSquares(t *testing.T) {
	// Two adjacent squares share the edge x=1.
	// Left:  (0,0)-(1,0)-(1,1)-(0,1)-(0,0)
	// Right: (1,0)-(2,0)-(2,1)-(1,1)-(1,0)
	// Shared edge (1,0)-(1,1) appears twice → dropped.
	left := &SegmentString{Coords: []geom.XY{
		{0, 0}, {1, 0}, {1, 1}, {0, 1}, {0, 0},
	}, Tag: 1}
	right := &SegmentString{Coords: []geom.XY{
		{1, 0}, {2, 0}, {2, 1}, {1, 1}, {1, 0},
	}, Tag: 2}
	n := NewBoundaryChainNoder([]*SegmentString{left, right})
	out := n.NodedSubstrings()
	require.NotEmpty(t, out)

	// Total boundary length should be 6 segments (two squares minus
	// shared edge counted once on each side).
	totalSegs := 0
	for _, ss := range out {
		totalSegs += ss.NumSegments()
	}
	assert.Equal(t, 6, totalSegs, "boundary segment count")

	// Shared edge (1,0)-(1,1) should not appear in any output.
	for _, ss := range out {
		for j := 0; j < ss.NumSegments(); j++ {
			a, b := ss.Segment(j)
			key := canonicalSegKey(a, b)
			if key == canonicalSegKey(geom.XY{1, 0}, geom.XY{1, 1}) {
				t.Fatalf("shared interior edge present in output")
			}
		}
	}
}

func TestBoundaryChainNoder_SinglePolygonAllBoundary(t *testing.T) {
	square := &SegmentString{Coords: []geom.XY{
		{0, 0}, {1, 0}, {1, 1}, {0, 1}, {0, 0},
	}}
	n := NewBoundaryChainNoder([]*SegmentString{square})
	out := n.NodedSubstrings()
	totalSegs := 0
	for _, ss := range out {
		totalSegs += ss.NumSegments()
	}
	assert.Equal(t, 4, totalSegs)
}

func TestBoundaryChainNoder_EmptyInput(t *testing.T) {
	n := NewBoundaryChainNoder(nil)
	assert.Empty(t, n.NodedSubstrings())
}

func TestBoundaryChainNoder_NodeMethodMatchesSubstrings(t *testing.T) {
	square := &SegmentString{Coords: []geom.XY{
		{0, 0}, {1, 0}, {1, 1}, {0, 1}, {0, 0},
	}}
	n := NewBoundaryChainNoder([]*SegmentString{square})
	a := n.NodedSubstrings()
	b := n.Node(nil)
	assert.Equal(t, len(a), len(b))
}

