package prepare_test

import (
	"testing"

	"github.com/exergy-dev/go-topology-suite/geom"
	"github.com/exergy-dev/go-topology-suite/kernel"
	"github.com/exergy-dev/go-topology-suite/kernel/planar"
	"github.com/exergy-dev/go-topology-suite/prepare"
	"pgregory.net/rapid"
)

func rectRing(x0, y0, x1, y1 float64) []geom.XY {
	return []geom.XY{{X: x0, Y: y0}, {X: x1, Y: y0}, {X: x1, Y: y1}, {X: x0, Y: y1}, {X: x0, Y: y0}}
}

// Differential: prepared ContainsPoint vs the per-ring kernel on polygons
// with holes, using integer grids so boundary hits are exercised.
func TestPreparedPolygon_ContainsPoint_HolesDifferential(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		shell := rectRing(0, 0, 40, 40)
		nh := rapid.IntRange(0, 12).Draw(t, "holes")
		var holes [][]geom.XY
		for i := 0; i < nh; i++ {
			cx, cy := float64(i%4)*10, float64(i/4)*10
			w := rapid.IntRange(1, 3).Draw(t, "w")
			h := rapid.IntRange(1, 3).Draw(t, "h")
			holes = append(holes, rectRing(cx+3, cy+3, cx+3+float64(w), cy+3+float64(h)))
		}
		pp := prepare.Polygon(geom.NewPolygon(nil, append([][]geom.XY{shell}, holes...)...))
		k := planar.Default()
		for q := 0; q < 50; q++ {
			p := geom.XY{
				X: float64(rapid.IntRange(-2, 42).Draw(t, "x")),
				Y: float64(rapid.IntRange(-2, 42).Draw(t, "y")),
			}
			want := k.PointInRing(p, shell)
			if want == kernel.Inside {
				for _, h := range holes {
					switch k.PointInRing(p, h) {
					case kernel.Inside:
						want = kernel.Outside
					case kernel.OnBoundary:
						want = kernel.OnBoundary
					}
					if want != kernel.Inside {
						break
					}
				}
			}
			if got := pp.ContainsPoint(p); got != want {
				t.Fatalf("p=%v prepared=%v kernel=%v", p, got, want)
			}
		}
	})
}
