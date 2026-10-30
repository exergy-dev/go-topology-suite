package prepare

import (
	"testing"

	"github.com/exergy-dev/go-topology-suite/geom"
	"pgregory.net/rapid"
)

func TestSegTreeMatchesBruteForce(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(0, 400).Draw(t, "n")
		coord := rapid.Float64Range(-50, 50)
		segs := make([]seg, n)
		for i := range segs {
			segs[i] = seg{
				geom.XY{X: coord.Draw(t, "ax"), Y: coord.Draw(t, "ay")},
				geom.XY{X: coord.Draw(t, "bx"), Y: coord.Draw(t, "by")},
			}
		}
		tr := newSegTree(segs)
		for k := 0; k < 10; k++ {
			x0, x1 := coord.Draw(t, "x0"), coord.Draw(t, "x1")
			y0, y1 := coord.Draw(t, "y0"), coord.Draw(t, "y1")
			q := geom.Envelope{MinX: min(x0, x1), MaxX: max(x0, x1), MinY: min(y0, y1), MaxY: max(y0, y1)}
			want := map[*seg]bool{}
			for i := range segs {
				if q.Intersects(geom.SegmentEnvelope(segs[i].a, segs[i].b)) {
					want[&segs[i]] = true
				}
			}
			got := map[*seg]bool{}
			tr.query(q, func(s *seg) bool { got[s] = true; return true })
			if len(got) != len(want) {
				t.Fatalf("got %d want %d", len(got), len(want))
			}
			for s := range want {
				if !got[s] {
					t.Fatalf("missing segment")
				}
			}
		}
	})
}
