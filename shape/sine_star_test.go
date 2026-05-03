package shape

import (
	"math"
	"testing"

	"github.com/terra-geo/terra/geom"
)

func TestSineStarVertexCountDefault(t *testing.T) {
	poly := SineStar(geom.XY{X: 0, Y: 0}, 100, 8)
	if poly.IsEmpty() {
		t.Fatalf("SineStar returned empty polygon")
	}
	ring := poly.Ring(0)
	// Default nPts=100 plus closing vertex.
	if len(ring) != 101 {
		t.Errorf("default ring vertices: got %d, want 101", len(ring))
	}
	if ring[0] != ring[len(ring)-1] {
		t.Errorf("ring not closed: first=%v last=%v", ring[0], ring[len(ring)-1])
	}
}

func TestSineStarVertexCountCustom(t *testing.T) {
	poly := SineStarWithOptions(geom.XY{X: 0, Y: 0}, 50, 5, SineStarOptions{NumPoints: 60, ArmLengthRatio: 0.3})
	ring := poly.Ring(0)
	if len(ring) != 61 {
		t.Errorf("custom ring vertices: got %d, want 61", len(ring))
	}
}

func TestSineStarBoundingBoxAndShape(t *testing.T) {
	const size = 100.0
	poly := SineStar(geom.XY{X: 0, Y: 0}, size, 8)
	env := poly.Envelope()
	// Outer radius = size/2 = 50; full outer-radius peaks land on the
	// envelope when the sine-wave hits its maximum, so |env| ≈ size.
	maxR := math.Max(math.Max(math.Abs(env.MinX), math.Abs(env.MaxX)),
		math.Max(math.Abs(env.MinY), math.Abs(env.MaxY)))
	if maxR > size/2+1e-9 || maxR < size/2*0.99 {
		t.Errorf("rough envelope radius %.4f outside [%.4f, %.4f]", maxR, size/2*0.99, size/2)
	}
}

func TestSineStarRadialSinusoid(t *testing.T) {
	// Verify the radial profile actually oscillates: at least one
	// vertex sits near the inner radius and at least one near the
	// outer radius.
	const size = 100.0
	const nArms = 6
	poly := SineStarWithOptions(geom.XY{X: 0, Y: 0}, size, nArms, SineStarOptions{NumPoints: 200, ArmLengthRatio: 0.5})
	ring := poly.Ring(0)
	radius := size / 2
	innerR := (1 - 0.5) * radius
	outerR := radius
	var sawInner, sawOuter bool
	for _, p := range ring {
		r := math.Hypot(p.X, p.Y)
		if math.Abs(r-innerR) < 1 {
			sawInner = true
		}
		if math.Abs(r-outerR) < 1 {
			sawOuter = true
		}
	}
	if !sawInner {
		t.Errorf("no vertex near inner radius %.2f", innerR)
	}
	if !sawOuter {
		t.Errorf("no vertex near outer radius %.2f", outerR)
	}
}

func TestSineStarOffsetCentre(t *testing.T) {
	poly := SineStar(geom.XY{X: 100, Y: 200}, 50, 4)
	env := poly.Envelope()
	// Centre should be near (100, 200): envelope midpoint matches.
	cx := (env.MinX + env.MaxX) / 2
	cy := (env.MinY + env.MaxY) / 2
	if math.Abs(cx-100) > 1e-6 || math.Abs(cy-200) > 1e-6 {
		t.Errorf("centred at (%v,%v), expected near (100,200)", cx, cy)
	}
}

func TestSineStarDegenerate(t *testing.T) {
	if !SineStar(geom.XY{X: 0, Y: 0}, 0, 8).IsEmpty() {
		t.Errorf("size=0 must yield empty polygon")
	}
	if !SineStar(geom.XY{X: 0, Y: 0}, 10, 0).IsEmpty() {
		t.Errorf("nArms=0 must yield empty polygon")
	}
}
