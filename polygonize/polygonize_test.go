package polygonize

import (
	"testing"

	"github.com/exergy-dev/go-topology-suite/geom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func line(pts ...geom.XY) *geom.LineString {
	return geom.NewLineString(nil, pts)
}

func TestPolygonizeClosedRing(t *testing.T) {
	// A single closed quad as one LineString.
	ls := line(
		geom.XY{X: 0, Y: 0}, geom.XY{X: 10, Y: 0},
		geom.XY{X: 10, Y: 10}, geom.XY{X: 0, Y: 10},
		geom.XY{X: 0, Y: 0},
	)
	polys, dangles, cuts, invalid := Polygonize([]geom.Geometry{ls})
	require.Equalf(t, 1, len(polys), "expected 1 polygon, got %d", len(polys))
	assert.Equalf(t, 0, len(dangles), "expected no dangles; got %d", len(dangles))
	assert.Equalf(t, 0, len(cuts), "expected no cuts; got %d", len(cuts))
	assert.Equalf(t, 0, len(invalid), "expected no invalid; got %d", len(invalid))
	p := polys[0].(*geom.Polygon)
	assert.Equalf(t, 1, p.NumRings(), "expected 1 ring, got %d", p.NumRings())
}

func TestPolygonizeTwoAdjacentBoxes(t *testing.T) {
	// Two unit squares sharing the vertical line x=10. Six edges total
	// (the shared edge counts once as a cut between the two polygons).
	//
	//   (0,0)──(10,0)──(20,0)
	//     │       │       │
	//   (0,10)─(10,10)─(20,10)
	a := line(geom.XY{X: 0, Y: 0}, geom.XY{X: 10, Y: 0})    // bottom-left edge
	b := line(geom.XY{X: 10, Y: 0}, geom.XY{X: 20, Y: 0})   // bottom-right edge
	c := line(geom.XY{X: 0, Y: 0}, geom.XY{X: 0, Y: 10})    // left edge
	d := line(geom.XY{X: 10, Y: 0}, geom.XY{X: 10, Y: 10})  // shared vertical
	e := line(geom.XY{X: 20, Y: 0}, geom.XY{X: 20, Y: 10})  // right edge
	f := line(geom.XY{X: 0, Y: 10}, geom.XY{X: 10, Y: 10})  // top-left edge
	g := line(geom.XY{X: 10, Y: 10}, geom.XY{X: 20, Y: 10}) // top-right edge

	polys, dangles, cuts, invalid := Polygonize([]geom.Geometry{a, b, c, d, e, f, g})
	require.Equalf(t, 2, len(polys), "expected 2 polygons, got %d", len(polys))
	assert.Equalf(t, 0, len(dangles), "expected no dangles; got %d", len(dangles))
	assert.Equalf(t, 0, len(invalid), "expected no invalid; got %d", len(invalid))
	_ = cuts // shared edge is a ring boundary for both polygons → not a cut
}

func TestPolygonizeDangle(t *testing.T) {
	// A square plus a dangling stub off one corner.
	ring := line(
		geom.XY{X: 0, Y: 0}, geom.XY{X: 10, Y: 0},
		geom.XY{X: 10, Y: 10}, geom.XY{X: 0, Y: 10},
		geom.XY{X: 0, Y: 0},
	)
	stub := line(geom.XY{X: 10, Y: 10}, geom.XY{X: 15, Y: 15})
	polys, dangles, _, _ := Polygonize([]geom.Geometry{ring, stub})
	assert.Equalf(t, 1, len(polys), "expected 1 polygon, got %d", len(polys))
	assert.Equalf(t, 1, len(dangles), "expected 1 dangle, got %d", len(dangles))
}

func TestPolygonizeInvalidRing(t *testing.T) {
	// A self-intersecting bowtie loop (figure-eight) as a single
	// closed LineString. Should be reported as invalid, not as a
	// polygon.
	bowtie := line(
		geom.XY{X: 0, Y: 0},
		geom.XY{X: 10, Y: 10},
		geom.XY{X: 10, Y: 0},
		geom.XY{X: 0, Y: 10},
		geom.XY{X: 0, Y: 0},
	)
	polys, _, _, invalid := Polygonize([]geom.Geometry{bowtie})
	assert.Equalf(t, 1, len(invalid), "expected 1 invalid ring, got %d (polys=%d)", len(invalid), len(polys))
}

func TestPolygonizeEmpty(t *testing.T) {
	polys, dangles, cuts, invalid := Polygonize(nil)
	assert.Equalf(t, 0, len(polys)+len(dangles)+len(cuts)+len(invalid), "empty input should yield empty output")
}
