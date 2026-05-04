package geom

import (
	"sync"
	"testing"

	"github.com/terra-geo/terra/crs"
)

// TestWithCRS_PointSwapsCRSWithoutMutatingSource regression-tests the
// fix for the *Point arm of WithCRS. The previous `out := *v; return
// &out` copied the embedded sync/atomic.Pointer envelope cache by
// value, tripping the noCopy contract; the concurrent-reads test
// below covers the race-detector side of the bug.
func TestWithCRS_PointSwapsCRSWithoutMutatingSource(t *testing.T) {
	src := NewPoint(crs.WGS84, XY{X: 1, Y: 2})
	dst := WithCRS(src, crs.WebMercator).(*Point)

	if dst.CRS() != crs.WebMercator {
		t.Fatalf("dst.CRS = %v, want WebMercator", dst.CRS())
	}
	if src.CRS() != crs.WGS84 {
		t.Fatalf("src.CRS mutated to %v, want WGS84", src.CRS())
	}
}

// TestWithCRS_PointConcurrentReadsAreSafe ensures that after WithCRS,
// concurrent Envelope() calls on src and dst race-free under -race.
func TestWithCRS_PointConcurrentReadsAreSafe(t *testing.T) {
	src := NewPoint(crs.WGS84, XY{X: 1, Y: 2})
	dst := WithCRS(src, crs.WebMercator).(*Point)

	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(2)
		go func() { defer wg.Done(); _ = src.Envelope() }()
		go func() { defer wg.Done(); _ = dst.Envelope() }()
	}
	wg.Wait()
}

// TestWithCRS_NilReturnsNil documents the nil-passthrough contract.
func TestWithCRS_NilReturnsNil(t *testing.T) {
	if got := WithCRS(nil, crs.WGS84); got != nil {
		t.Fatalf("WithCRS(nil) = %v, want nil", got)
	}
}
