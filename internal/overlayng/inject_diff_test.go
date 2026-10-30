package overlayng

import (
	"math/rand"
	"testing"

	"github.com/exergy-dev/go-topology-suite/geom"
)

// injectRingSelfContactsRef is the original brute-force implementation.
func injectRingSelfContactsRef(ring []geom.XY) ([]geom.XY, bool) {
	if len(ring) < 5 {
		return ring, false
	}
	out := append([]geom.XY(nil), ring...)
	changed := false
	for pass := 0; pass < 8; pass++ {
		passChanged := false
		for vi := 0; vi < len(out)-1; vi++ {
			v := out[vi]
			for si := 0; si+1 < len(out); si++ {
				if si == vi || si == vi-1 || (vi == 0 && si == len(out)-2) {
					continue
				}
				a, b := out[si], out[si+1]
				if a == v || b == v {
					continue
				}
				if !pointOnSegmentInterior(v, a, b) {
					continue
				}
				newRing := make([]geom.XY, 0, len(out)+1)
				newRing = append(newRing, out[:si+1]...)
				newRing = append(newRing, v)
				newRing = append(newRing, out[si+1:]...)
				out = newRing
				passChanged = true
				changed = true
				break
			}
			if passChanged {
				break
			}
		}
		if !passChanged {
			break
		}
	}
	return out, changed
}

func TestInjectRingSelfContactsDifferential(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	check := func(ring []geom.XY) {
		t.Helper()
		g, gc := injectRingSelfContacts(ring)
		w, wc := injectRingSelfContactsRef(ring)
		if gc != wc || len(g) != len(w) {
			t.Fatalf("mismatch on %v: got %v,%v want %v,%v", ring, g, gc, w, wc)
		}
		for i := range g {
			if g[i] != w[i] {
				t.Fatalf("mismatch on %v: got %v want %v", ring, g, w)
			}
		}
	}
	for it := 0; it < 20000; it++ {
		n := 4 + rng.Intn(14)
		grid := 2 + rng.Intn(6)
		scale := []float64{1, 0.1, 1e-120, 1e6}[rng.Intn(4)]
		ring := make([]geom.XY, 0, n+1)
		for i := 0; i < n; i++ {
			ring = append(ring, geom.XY{X: float64(rng.Intn(grid)) * scale, Y: float64(rng.Intn(grid)) * scale})
		}
		if rng.Intn(4) == 0 {
			ring[rng.Intn(n)] = ring[rng.Intn(n)]
		}
		ring = append(ring, ring[0])
		check(ring)
	}
	check([]geom.XY{{X: 0, Y: 0}, {X: 4, Y: 0}, {X: 2, Y: 0}, {X: 2, Y: 2}, {X: 0, Y: 2}, {X: 0, Y: 0}})
	check([]geom.XY{{X: 0, Y: 0}, {X: 4, Y: 0}, {X: 4, Y: 4}, {X: 2, Y: 0}, {X: 0, Y: 4}, {X: 0, Y: 0}})
	check([]geom.XY{{X: 0, Y: 0}, {X: 2, Y: 2}, {X: 4, Y: 0}, {X: 4, Y: 4}, {X: 2, Y: 2}, {X: 0, Y: 4}, {X: 0, Y: 0}})
	for it := 0; it < 5000; it++ {
		ring := make([]geom.XY, 0, 12)
		for i := 0; i < 11; i++ {
			ring = append(ring, geom.XY{X: rng.Float64(), Y: rng.Float64()})
		}
		a, b := ring[0], ring[1]
		ring[5] = geom.XY{X: a.X + (b.X-a.X)*0.5, Y: a.Y + (b.Y-a.Y)*0.5}
		ring = append(ring, ring[0])
		check(ring)
	}
}
