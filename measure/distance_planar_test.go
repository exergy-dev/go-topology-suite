package measure

import (
	"math"
	"math/rand"
	"testing"

	"github.com/exergy-dev/go-topology-suite/geom"
	"github.com/exergy-dev/go-topology-suite/internal/benchfix"
	"github.com/exergy-dev/go-topology-suite/kernel/planar"
)

func randGeom(r *rand.Rand, ox, oy float64) geom.Geometry {
	pt := func() geom.XY {
		return geom.XY{X: ox + r.Float64()*20 - 10, Y: oy + r.Float64()*20 - 10}
	}
	line := func() *geom.LineString {
		n := 2 + r.Intn(40)
		pts := make([]geom.XY, n)
		for i := range pts {
			pts[i] = pt()
		}
		return geom.NewLineString(nil, pts)
	}
	poly := func() *geom.Polygon {
		p := benchfix.Star(4+r.Intn(60), 1+r.Float64()*9, 1+r.Float64()*4)
		rings := make([][]geom.XY, p.NumRings())
		for i := range rings {
			for _, q := range p.Ring(i) {
				rings[i] = append(rings[i], geom.XY{X: q.X + ox, Y: q.Y + oy})
			}
		}
		return geom.NewPolygon(nil, rings...)
	}
	switch r.Intn(7) {
	case 0:
		return geom.NewPoint(nil, pt())
	case 1:
		pts := make([]geom.XY, 1+r.Intn(30))
		for i := range pts {
			pts[i] = pt()
		}
		return geom.NewMultiPoint(nil, pts)
	case 2:
		return line()
	case 3:
		return poly()
	case 4:
		return geom.NewMultiLineString(nil, line(), line())
	case 5:
		return geom.NewMultiPolygon(nil, poly(), poly())
	default:
		return geom.NewGeometryCollection(nil, poly(), line(), geom.NewPoint(nil, pt()))
	}
}

// distanceReference is the original algorithm: point-point shortcut,
// containment, then the brute-force minimum.
func distanceReference(a, b geom.Geometry) float64 {
	k := planar.Default()
	if pa, ok := a.(*geom.Point); ok {
		if pb, ok := b.(*geom.Point); ok {
			return k.Distance(pa.XY(), pb.XY())
		}
	}
	if areaContainsAnyPoint(a, b, k) || areaContainsAnyPoint(b, a, k) {
		return 0
	}
	return distanceBruteForce(a, b, k)
}

func TestDistancePlanarMatchesBruteForce(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	for i := 0; i < 3000; i++ {
		off := []float64{0, 5, 15, 30, 200}[r.Intn(5)]
		a := randGeom(r, 0, 0)
		b := randGeom(r, off*r.Float64(), off*r.Float64())
		got, err := Distance(a, b)
		if err != nil {
			t.Fatal(err)
		}
		want := distanceReference(a, b)
		if math.Float64bits(got) != math.Float64bits(want) {
			t.Fatalf("case %d: got %v want %v\na=%v\nb=%v", i, got, want, a, b)
		}
	}
}

func TestDistancePlanarLargeStars(t *testing.T) {
	for _, shift := range []float64{0, 100, 200, 220, 1000} {
		a := benchfix.Star(512, 100, 60)
		ring := benchfix.Star(300, 100, 60).Ring(0)
		for j := range ring {
			ring[j].X += shift
		}
		b := geom.NewPolygon(nil, ring)
		got, _ := Distance(a, b)
		want := distanceReference(a, b)
		if math.Float64bits(got) != math.Float64bits(want) {
			t.Fatalf("shift %v: got %v want %v", shift, got, want)
		}
	}
}
