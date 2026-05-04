package geom

import (
	"sync"
	"testing"

	"github.com/terra-geo/terra/crs"
)

// TestWithCRS_PointDoesNotShareEnvelopeCache regression-tests the fix
// for the *Point arm of WithCRS. The arm previously did `out := *v;
// return &out`, which copied the embedded sync/atomic.Pointer envelope
// cache by value (and tripped go vet's "copies lock value" warning).
//
// The semantic risk is that mutating one Point's cache (e.g. by
// computing its envelope) was visible on the other, and worse, copying
// an atomic value violates Go's noCopy contract.
func TestWithCRS_PointDoesNotShareEnvelopeCache(t *testing.T) {
	src := NewPoint(crs.WGS84, XY{X: 1, Y: 2})

	// Force the source's envelope cache to populate.
	_ = src.Envelope()

	// Re-brand under a different CRS.
	dst := WithCRS(src, crs.WebMercator).(*Point)

	if dst.CRS() != crs.WebMercator {
		t.Fatalf("WithCRS: dst.CRS = %v, want WebMercator", dst.CRS())
	}
	if src.CRS() != crs.WGS84 {
		t.Fatalf("WithCRS: src.CRS mutated to %v, want WGS84", src.CRS())
	}

	// The envelope cache MUST be independent: dst's atomic.Pointer
	// must start unloaded, even though src's is populated.
	if dst.env.Load() != nil && src.env.Load() != nil &&
		dst.env.Load() == src.env.Load() {
		t.Fatalf("WithCRS: src and dst share the same envelope-cache pointer; cache must be independent")
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
