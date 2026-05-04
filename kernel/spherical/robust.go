package spherical

import (
	"math"
	"math/big"
	"sync"

	"github.com/exergy-dev/go-topology-suite/kernel"
)

// adaptiveOrient3D returns the sign of det(a, b, c) where a, b, c are
// unit 3-vectors on the sphere. It is a Shewchuk-style adaptive
// predicate: a float64 fast path with a running absolute-error bound,
// falling back to math/big.Rat exact arithmetic when the float sign
// cannot be trusted.
//
// The result is exact for any triple of unit vectors derived from
// finite lon/lat inputs: input scale is bounded and the rational fallback
// preserves every bit of every float64 operand, so the sign of the
// returned value is the sign of the true determinant.
//
// The bound below is Shewchuk's o3derrboundA-style coefficient: a
// conservative scalar that guarantees |float det - true det| < bound.
const orient3DerrboundA = (7.0 + 56.0*1.1102230246251565e-16) * 1.1102230246251565e-16

func adaptiveOrient3D(a, b, c vec3) kernel.Orientation {
	bycz := b.Y * c.Z
	bzcy := b.Z * c.Y
	bxcz := b.X * c.Z
	bzcx := b.Z * c.X
	bxcy := b.X * c.Y
	bycx := b.Y * c.X

	minor1 := bycz - bzcy
	minor2 := bxcz - bzcx
	minor3 := bxcy - bycx

	det := a.X*minor1 - a.Y*minor2 + a.Z*minor3

	// Conservative error bound. Each minor's bound is the sum of
	// absolute values of its two product terms; the full bound is the
	// |a.*|-weighted sum of those, scaled by Shewchuk's coefficient.
	bound := orient3DerrboundA * (math.Abs(a.X)*(math.Abs(bycz)+math.Abs(bzcy)) +
		math.Abs(a.Y)*(math.Abs(bxcz)+math.Abs(bzcx)) +
		math.Abs(a.Z)*(math.Abs(bxcy)+math.Abs(bycx)))

	if math.Abs(det) > bound {
		return signToOrientation3D(det)
	}

	// Filter fail: cache, then exact.
	key := orient3DKey{
		ax: math.Float64bits(a.X), ay: math.Float64bits(a.Y), az: math.Float64bits(a.Z),
		bx: math.Float64bits(b.X), by: math.Float64bits(b.Y), bz: math.Float64bits(b.Z),
		cx: math.Float64bits(c.X), cy: math.Float64bits(c.Y), cz: math.Float64bits(c.Z),
	}
	if o, ok := orient3DCache.lookup(key); ok {
		return o
	}
	o := exactOrient3D(a, b, c)
	orient3DCache.store(key, o)
	return o
}

func signToOrientation3D(v float64) kernel.Orientation {
	switch {
	case v > 0:
		return kernel.CounterClockwise
	case v < 0:
		return kernel.Clockwise
	default:
		return kernel.Collinear
	}
}

// exactOrient3D computes det(a, b, c) using math/big.Rat. Like the
// planar exactOrient, this is genuinely exact for all float64 inputs;
// no precision parameter to tune.
func exactOrient3D(a, b, c vec3) kernel.Orientation {
	ax := new(big.Rat).SetFloat64(a.X)
	ay := new(big.Rat).SetFloat64(a.Y)
	az := new(big.Rat).SetFloat64(a.Z)
	bx := new(big.Rat).SetFloat64(b.X)
	by := new(big.Rat).SetFloat64(b.Y)
	bz := new(big.Rat).SetFloat64(b.Z)
	cx := new(big.Rat).SetFloat64(c.X)
	cy := new(big.Rat).SetFloat64(c.Y)
	cz := new(big.Rat).SetFloat64(c.Z)

	if ax == nil || ay == nil || az == nil ||
		bx == nil || by == nil || bz == nil ||
		cx == nil || cy == nil || cz == nil {
		return kernel.Collinear
	}

	// minor1 = b.Y * c.Z - b.Z * c.Y
	m1 := new(big.Rat).Sub(new(big.Rat).Mul(by, cz), new(big.Rat).Mul(bz, cy))
	// minor2 = b.X * c.Z - b.Z * c.X
	m2 := new(big.Rat).Sub(new(big.Rat).Mul(bx, cz), new(big.Rat).Mul(bz, cx))
	// minor3 = b.X * c.Y - b.Y * c.X
	m3 := new(big.Rat).Sub(new(big.Rat).Mul(bx, cy), new(big.Rat).Mul(by, cx))

	// det = a.X*m1 - a.Y*m2 + a.Z*m3
	det := new(big.Rat).Mul(ax, m1)
	det.Sub(det, new(big.Rat).Mul(ay, m2))
	det.Add(det, new(big.Rat).Mul(az, m3))

	switch det.Sign() {
	case 1:
		return kernel.CounterClockwise
	case -1:
		return kernel.Clockwise
	default:
		return kernel.Collinear
	}
}

// orient3DKey is the bit-pattern key for the spherical-orient
// memoization cache, mirroring the planar predicate's design.
type orient3DKey struct {
	ax, ay, az, bx, by, bz, cx, cy, cz uint64
}

const orient3DCacheCap = 1024

type orient3DCacheT struct {
	mu      sync.RWMutex
	entries map[orient3DKey]kernel.Orientation
	order   [orient3DCacheCap]orient3DKey
	next    int
}

var orient3DCache = &orient3DCacheT{
	entries: make(map[orient3DKey]kernel.Orientation, orient3DCacheCap),
}

func (c *orient3DCacheT) lookup(k orient3DKey) (kernel.Orientation, bool) {
	c.mu.RLock()
	o, ok := c.entries[k]
	c.mu.RUnlock()
	return o, ok
}

func (c *orient3DCacheT) store(k orient3DKey, o kernel.Orientation) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.entries[k]; ok {
		return
	}
	if len(c.entries) == orient3DCacheCap {
		delete(c.entries, c.order[c.next])
	}
	c.order[c.next] = k
	c.entries[k] = o
	c.next++
	if c.next >= orient3DCacheCap {
		c.next = 0
	}
}
