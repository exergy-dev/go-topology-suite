package densify

import (
	"math"
	"testing"

	"github.com/exergy-dev/go-topology-suite/geom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func segLen(a, b geom.XY) float64 {
	return math.Hypot(b.X-a.X, b.Y-a.Y)
}

func TestDensifyLineString(t *testing.T) {
	ls := geom.NewLineString(nil, []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}})
	out := Densify(ls, 3.0).(*geom.LineString)
	require.GreaterOrEqualf(t, out.NumPoints(), 4, "expected at least 4 points after densifying, got %d", out.NumPoints())
	for i := 0; i < out.NumPoints()-1; i++ {
		d := segLen(out.PointAt(i), out.PointAt(i+1))
		assert.LessOrEqualf(t, d, 3.0+1e-9, "segment %d has length %g > tol", i, d)
	}
	// endpoints preserved
	assert.Equal(t, geom.XY{X: 0, Y: 0}, out.PointAt(0), "endpoints not preserved")
	assert.Equal(t, geom.XY{X: 10, Y: 0}, out.PointAt(out.NumPoints()-1), "endpoints not preserved")
}

func TestDensifyShortLineUntouched(t *testing.T) {
	ls := geom.NewLineString(nil, []geom.XY{{X: 0, Y: 0}, {X: 1, Y: 0}})
	out := Densify(ls, 5.0).(*geom.LineString)
	assert.Equalf(t, 2, out.NumPoints(), "expected 2 points (no densification), got %d", out.NumPoints())
}

func TestDensifyPoint(t *testing.T) {
	p := geom.NewPoint(nil, geom.XY{X: 5, Y: 5})
	out := Densify(p, 1.0)
	assert.Equal(t, p, out, "Point should be returned as-is")
}

func TestDensifyPolygon(t *testing.T) {
	ring := []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}, {X: 0, Y: 10}, {X: 0, Y: 0}}
	p := geom.NewPolygon(nil, ring)
	out := Densify(p, 3.0).(*geom.Polygon)
	r := out.Ring(0)
	require.Greaterf(t, len(r), len(ring), "expected polygon ring densified, got %d vertices", len(r))
	for i := 0; i < len(r)-1; i++ {
		d := segLen(r[i], r[i+1])
		assert.LessOrEqualf(t, d, 3.0+1e-9, "ring segment %d length %g > tol", i, d)
	}
	assert.Equal(t, ring[0], r[0], "ring endpoints not preserved")
	assert.Equal(t, ring[len(ring)-1], r[len(r)-1], "ring endpoints not preserved")
}

func TestDensifyNonPositiveTol(t *testing.T) {
	ls := geom.NewLineString(nil, []geom.XY{{X: 0, Y: 0}, {X: 100, Y: 0}})
	assert.Equal(t, 2, Densify(ls, 0).(*geom.LineString).NumPoints(), "zero tol should be no-op")
	assert.Equal(t, 2, Densify(ls, -1).(*geom.LineString).NumPoints(), "negative tol should be no-op")
}

func TestDensifyMultiLineString(t *testing.T) {
	ls1 := geom.NewLineString(nil, []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}})
	ls2 := geom.NewLineString(nil, []geom.XY{{X: 0, Y: 5}, {X: 8, Y: 5}})
	mls := geom.NewMultiLineString(nil, ls1, ls2)
	out := Densify(mls, 3.0).(*geom.MultiLineString)
	require.Equalf(t, 2, out.NumGeometries(), "expected 2 lines, got %d", out.NumGeometries())
}
