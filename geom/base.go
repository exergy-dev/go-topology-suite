package geom

import (
	"sync/atomic"

	"github.com/terra-geo/terra/crs"
)

// baseGeom is embedded by every concrete geometry type. It owns the flat
// coordinate buffer, the CRS pointer, and the envelope cache.
//
// The envelope cache is an atomic.Pointer for lock-free lazy initialisation
// — read-only operations on a constructed geometry are safe for concurrent
// use. ApplyCoordinateFilter and similar mutators are NOT safe for
// concurrent use; they invalidate the cache.
type baseGeom struct {
	layout Layout
	coords []float64
	crs    *crs.CRS
	env    atomic.Pointer[Envelope]
}

func (b *baseGeom) Layout() Layout { return b.layout }
func (b *baseGeom) CRS() *crs.CRS  { return b.crs }
func (b *baseGeom) FlatCoords() []float64 {
	// Returns the underlying buffer. Callers MUST treat as read-only;
	// mutating it bypasses the envelope cache invariant.
	return b.coords
}

// stride returns the number of float64 values per coordinate.
func (b *baseGeom) stride() int { return b.layout.Stride() }

// numCoords returns the number of vertices stored.
func (b *baseGeom) numCoords() int {
	s := b.stride()
	if s == 0 {
		return 0
	}
	return len(b.coords) / s
}

// envelope returns the cached envelope, computing it on first call.
// Multiple concurrent callers may compute it; one wins the CAS, the rest
// discard their result. This is the same pattern used in the v2 codebase.
func (b *baseGeom) envelope() Envelope {
	if e := b.env.Load(); e != nil {
		return *e
	}
	computed := envelopeOfFlat(b.coords, b.stride())
	b.env.CompareAndSwap(nil, &computed)
	if e := b.env.Load(); e != nil {
		return *e
	}
	return computed
}

// cloneFloats returns a defensive copy of in. Constructors clone inputs by
// default so callers retain ownership of their own slices — the same rule
// the v2 codebase used.
func cloneFloats(in []float64) []float64 {
	if len(in) == 0 {
		return nil
	}
	out := make([]float64, len(in))
	copy(out, in)
	return out
}
