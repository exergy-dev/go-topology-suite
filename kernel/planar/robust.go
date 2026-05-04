package planar

import (
	"math"
	"math/big"
	"sync"

	"github.com/exergy-dev/go-topology-suite/geom"
	"github.com/exergy-dev/go-topology-suite/kernel"
)

// adaptiveOrient is the Shewchuk-style adaptive 2D orientation predicate.
// Common case (well-conditioned inputs) runs at native float64 speed; on
// near-collinear inputs the result is verified using big.Float at 113-bit
// (quadruple) precision and only the sign of the exact result is returned.
//
// Reference: Jonathan R. Shewchuk, "Adaptive Precision Floating-Point
// Arithmetic and Fast Robust Geometric Predicates," Discrete & Computational
// Geometry 18(3):305-363, 1997.
//
// The error bound below (ccwerrboundA) is the first-level filter from
// Shewchuk §4.2, derived from the worst-case rounding of the four
// multiplications and one subtraction in the naive computation.
const ccwerrboundA = (3.0 + 16.0*1.1102230246251565e-16) * 1.1102230246251565e-16

// adaptiveOrient returns +1 (CCW), -1 (CW), or 0 (collinear) for the
// triangle (a, b, c). The algorithm computes (b-a) × (c-a) using the same
// formula as the naive path, then either trusts the result (filter pass)
// or verifies via exact arithmetic (filter fail).
func adaptiveOrient(a, b, c geom.XY) kernel.Orientation {
	detLeft := (b.X - a.X) * (c.Y - a.Y)
	detRight := (b.Y - a.Y) * (c.X - a.X)
	det := detLeft - detRight

	// Sign-only fast paths: when the two product terms have strictly
	// different signs, the subtraction is exact and `det` is reliable.
	var detSum float64
	switch {
	case detLeft > 0:
		if detRight <= 0 {
			return signToOrientation(det)
		}
		detSum = detLeft + detRight
	case detLeft < 0:
		if detRight >= 0 {
			return signToOrientation(det)
		}
		detSum = -detLeft - detRight
	default:
		return signToOrientation(det)
	}

	errBound := ccwerrboundA * detSum
	if det >= errBound || -det >= errBound {
		return signToOrientation(det)
	}

	// Filter fail: inputs are too near collinear for float64 to decide
	// the sign safely. The math/big recomputation is expensive — try the
	// memoization cache first. The cache is consulted ONLY here so that
	// well-conditioned inputs (the hot path) pay zero overhead.
	key := orientKey{
		ax: math.Float64bits(a.X), ay: math.Float64bits(a.Y),
		bx: math.Float64bits(b.X), by: math.Float64bits(b.Y),
		cx: math.Float64bits(c.X), cy: math.Float64bits(c.Y),
	}
	if o, ok := orientCache.lookup(key); ok {
		return o
	}
	o := exactOrient(a, b, c)
	orientCache.store(key, o)
	return o
}

// orientKey is the bit-pattern key for the adaptive-orient memoization
// cache. Using math.Float64bits instead of the raw float values gives a
// strict-equality comparison (NaN==NaN, +0==-0 distinguished) which is
// what we want: the cached result is the exact predicate output for a
// specific bit pattern, nothing more.
type orientKey struct {
	ax, ay, bx, by, cx, cy uint64
}

// orientCacheCap bounds the cache at 1024 entries. The eviction strategy
// is round-robin: when full, the next insert overwrites the slot pointed
// to by `next`, which advances modulo the capacity. This avoids the bookkeeping
// of a true LRU while still preventing unbounded growth on adversarial inputs.
const orientCacheCap = 1024

type orientCacheT struct {
	mu sync.RWMutex
	// entries maps key → orientation. Capped at orientCacheCap.
	entries map[orientKey]kernel.Orientation
	// order records insertion order so eviction can find the oldest slot.
	// Round-robin: advance `next` and drop entries[order[next]].
	order [orientCacheCap]orientKey
	next  int
}

var orientCache = &orientCacheT{
	entries: make(map[orientKey]kernel.Orientation, orientCacheCap),
}

func (c *orientCacheT) lookup(k orientKey) (kernel.Orientation, bool) {
	c.mu.RLock()
	o, ok := c.entries[k]
	c.mu.RUnlock()
	return o, ok
}

func (c *orientCacheT) store(k orientKey, o kernel.Orientation) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// Re-check inside the write lock; another goroutine may have raced us.
	if _, ok := c.entries[k]; ok {
		return
	}
	// Map size == capacity ⇒ we've wrapped at least once and the slot
	// at order[next] is occupied; evict it before we overwrite.
	if len(c.entries) == orientCacheCap {
		delete(c.entries, c.order[c.next])
	}
	c.order[c.next] = k
	c.entries[k] = o
	c.next++
	if c.next >= orientCacheCap {
		c.next = 0
	}
}

func signToOrientation(v float64) kernel.Orientation {
	switch {
	case v > 0:
		return kernel.CounterClockwise
	case v < 0:
		return kernel.Clockwise
	default:
		return kernel.Collinear
	}
}

// exactOrient computes (b-a) × (c-a) using exact rational arithmetic and
// returns only the sign. Slower than full Shewchuk error-free
// transformations but obviously correct for ALL pairs of float64 inputs
// that don't overflow.
//
// Why rationals (math/big.Rat) instead of fixed-precision big.Float:
// the input magnitudes can span the full float64 dynamic range (~617
// decimal orders of magnitude). A 256-bit big.Float (~77 decimal digits)
// silently truncates subnormal contributions when subtracted from
// large-magnitude values, producing wrong-sign zeros for adversarial
// triples. big.Rat preserves every bit of every float64 input and
// performs the subtraction/multiplication symbolically — no precision
// parameter to tune.
func exactOrient(a, b, c geom.XY) kernel.Orientation {
	ax := new(big.Rat).SetFloat64(a.X)
	ay := new(big.Rat).SetFloat64(a.Y)
	bx := new(big.Rat).SetFloat64(b.X)
	by := new(big.Rat).SetFloat64(b.Y)
	cx := new(big.Rat).SetFloat64(c.X)
	cy := new(big.Rat).SetFloat64(c.Y)

	// SetFloat64 returns nil for NaN/±Inf inputs. The fast path won't
	// route NaN here (all comparisons against NaN are false, so the
	// default branch returns Collinear). Inf-containing inputs likewise
	// short-circuit at the fast path. Treat any nil here as Collinear.
	if ax == nil || ay == nil || bx == nil || by == nil || cx == nil || cy == nil {
		return kernel.Collinear
	}

	// (b.X - a.X) * (c.Y - a.Y)
	tmp1 := new(big.Rat).Sub(bx, ax)
	tmp2 := new(big.Rat).Sub(cy, ay)
	left := new(big.Rat).Mul(tmp1, tmp2)

	// (b.Y - a.Y) * (c.X - a.X)
	tmp3 := new(big.Rat).Sub(by, ay)
	tmp4 := new(big.Rat).Sub(cx, ax)
	right := new(big.Rat).Mul(tmp3, tmp4)

	diff := new(big.Rat).Sub(left, right)
	switch diff.Sign() {
	case 1:
		return kernel.CounterClockwise
	case -1:
		return kernel.Clockwise
	default:
		return kernel.Collinear
	}
}
