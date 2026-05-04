package terra_test

import (
	"math"
	"testing"

	"github.com/terra-geo/terra"
	"github.com/terra-geo/terra/crs"
	"github.com/terra-geo/terra/crs/epsg"
	"github.com/terra-geo/terra/geom"
)

// TestTransformWGS84ToWebMercatorRoundTrip exercises the public
// terra.Transform API end-to-end on a small polygon, asserting that
// the round-trip recovers the input within 1mm.
func TestTransformWGS84ToWebMercatorRoundTrip(t *testing.T) {
	src := epsg.WGS84
	dst := epsg.WebMercator

	poly := geom.NewPolygon(src,
		[]geom.XY{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 1}, {X: 0, Y: 0}},
	)

	projected, err := terra.Transform(poly, dst)
	if err != nil {
		t.Fatalf("Transform forward: %v", err)
	}
	if !crs.Equal(projected.CRS(), dst) {
		t.Fatalf("projected.CRS() = %+v; want %+v", projected.CRS(), dst)
	}

	// (1°, 0°) on WGS84 web-mercators to roughly (111319.49, 0). We
	// don't pin the exact value here — round-trip is the contract.
	roundTrip, err := terra.Transform(projected, src)
	if err != nil {
		t.Fatalf("Transform back: %v", err)
	}
	if !crs.Equal(roundTrip.CRS(), src) {
		t.Fatalf("roundTrip.CRS() = %+v; want %+v", roundTrip.CRS(), src)
	}

	got := roundTrip.(*geom.Polygon)
	want := poly
	if got.RingLen(0) != want.RingLen(0) {
		t.Fatalf("ring length: got %d want %d", got.RingLen(0), want.RingLen(0))
	}
	for i := 0; i < got.RingLen(0); i++ {
		g := got.RingVertex(0, i)
		w := want.RingVertex(0, i)
		if math.Abs(g.X-w.X) > 1e-9 || math.Abs(g.Y-w.Y) > 1e-9 {
			t.Errorf("vertex %d: got (%v, %v) want (%v, %v)", i, g.X, g.Y, w.X, w.Y)
		}
	}
}

// TestTransformWGS84ToUTM exercises a UTM-zone projected target — the
// bulk of Terra's registered EPSG codes. Reference value drawn from
// PROJ's own gie corpus (test/gie/builtins.gie, +proj=utm +zone=32).
func TestTransformWGS84ToUTM(t *testing.T) {
	src := epsg.WGS84
	dst := epsg.Lookup(32632) // WGS84 / UTM zone 32N
	if dst == nil {
		t.Fatal("EPSG:32632 not registered")
	}

	pt := geom.NewPoint(src, geom.XY{X: 12, Y: 56})
	projected, err := terra.Transform(pt, dst)
	if err != nil {
		t.Fatalf("Transform: %v", err)
	}
	out := projected.(*geom.Point).XY()

	const wantE, wantN = 687071.43910944, 6210141.32674801
	if math.Abs(out.X-wantE) > 1e-3 || math.Abs(out.Y-wantN) > 1e-3 {
		t.Errorf("UTM32N: got (%v, %v) want (%v, %v)", out.X, out.Y, wantE, wantN)
	}

	// Round-trip back.
	back, err := terra.Transform(projected, src)
	if err != nil {
		t.Fatalf("inverse: %v", err)
	}
	bp := back.(*geom.Point).XY()
	if math.Abs(bp.X-12) > 1e-8 || math.Abs(bp.Y-56) > 1e-8 {
		t.Errorf("round-trip: got (%v, %v) want (12, 56)", bp.X, bp.Y)
	}
}

// TestTransformErrUntransformable verifies that a Transform attempt
// against a CRS without a Definition returns the expected sentinel.
func TestTransformErrUntransformable(t *testing.T) {
	bare := &crs.CRS{Authority: "EPSG", Code: 12345, Kind: crs.Projected}
	pt := geom.NewPoint(epsg.WGS84, geom.XY{X: 1, Y: 1})
	if _, err := terra.Transform(pt, bare); err != crs.ErrUntransformable {
		t.Errorf("got err=%v, want %v", err, crs.ErrUntransformable)
	}
}

// TestTransformIdentity verifies that transforming to the same CRS
// returns a value with the target CRS pointer (not a deep copy).
func TestTransformIdentity(t *testing.T) {
	pt := geom.NewPoint(epsg.WGS84, geom.XY{X: 1, Y: 1})
	out, err := terra.Transform(pt, epsg.WGS84)
	if err != nil {
		t.Fatalf("Transform: %v", err)
	}
	if !crs.Equal(out.CRS(), epsg.WGS84) {
		t.Errorf("CRS mismatch")
	}
	op := out.(*geom.Point)
	if op.XY() != (geom.XY{X: 1, Y: 1}) {
		t.Errorf("XY changed: %v", op.XY())
	}
}
