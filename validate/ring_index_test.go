package validate

import (
	"testing"

	"github.com/exergy-dev/go-topology-suite/geom"
	"pgregory.net/rapid"
)

// TestRingSelfIntersectionIndexedMatchesBrute checks the monotone-chain
// indexed search against the all-pairs scan, including the reported
// location, on random closed rings (small integer grid to provoke touches
// and collinear overlaps, plus a float variant).
func TestRingSelfIntersectionIndexedMatchesBrute(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(4, 150).Draw(t, "n")
		grid := rapid.Bool().Draw(t, "grid")
		pts := make([]geom.XY, 0, n+1)
		for i := 0; i < n; i++ {
			var p geom.XY
			if grid {
				p = geom.XY{X: float64(rapid.IntRange(0, 40).Draw(t, "x")), Y: float64(rapid.IntRange(0, 40).Draw(t, "y"))}
			} else {
				p = geom.XY{X: rapid.Float64Range(-100, 100).Draw(t, "x"), Y: rapid.Float64Range(-100, 100).Draw(t, "y")}
			}
			pts = append(pts, p)
		}
		pts = append(pts, pts[0])
		c := collapseConsecutiveDuplicates(pts)
		if len(c) < 4 {
			return
		}
		wp, wok := ringSelfIntersectionBrute(c)
		gp, gok := ringSelfIntersectionIndexed(c)
		if wok != gok || (wok && wp != gp) {
			t.Fatalf("brute=(%v,%v) indexed=(%v,%v) ring=%v", wp, wok, gp, gok, c)
		}
	})
}
