package noding

import (
	"errors"
	"math"

	"github.com/exergy-dev/go-topology-suite/geom"
)

// IteratedNoder repeatedly applies an inner noder until the output set
// is stable (no new intersections are produced) or a maximum iteration
// count is reached. It is a port of
// org.locationtech.jts.noding.IteratedNoder.
//
// IteratedNoder is the right choice when the underlying noder uses a
// finite-tolerance intersection primitive (e.g. snap-rounding) that
// can introduce *new* near-coincident edges on each pass. A single
// noding pass leaves these unresolved; iteration converges them.
//
// The default MaxIter is 5; ErrNotConverged is returned (alongside the
// last partial result) if convergence is not achieved before then.
type IteratedNoder struct {
	inner   Noder
	MaxIter int
}

// ErrNotConverged is returned by NodeIter when the iterated noder
// reaches MaxIter without producing two consecutive identical
// output sets.
var ErrNotConverged = errors.New("noding: iterated noder did not converge")

// NewIteratedNoder wraps inner with iterative re-noding. maxIter ≤ 0
// is replaced by the JTS default of 5.
func NewIteratedNoder(inner Noder, maxIter int) *IteratedNoder {
	if maxIter <= 0 {
		maxIter = 5
	}
	return &IteratedNoder{inner: inner, MaxIter: maxIter}
}

// Node satisfies the Noder interface. Convergence failure is silent —
// the last produced output is returned. Use NodeIter to detect it.
func (n *IteratedNoder) Node(input []*SegmentString) []*SegmentString {
	out, _ := n.NodeIter(input)
	return out
}

// NodeIter iterates the inner noder until the output set is stable.
// Returns the last output produced and ErrNotConverged if MaxIter
// was reached without a stable output.
func (n *IteratedNoder) NodeIter(input []*SegmentString) ([]*SegmentString, error) {
	cur := input
	prevSig := signature(cur)
	for i := 0; i < n.MaxIter; i++ {
		next := n.inner.Node(cur)
		nextSig := signature(next)
		if nextSig == prevSig {
			return next, nil
		}
		cur = next
		prevSig = nextSig
	}
	return cur, ErrNotConverged
}

// signature reduces a slice of segment strings to a deterministic
// fingerprint suitable for "is this the same noding as last time?".
// Order-sensitive: the inner noder is expected to be deterministic
// across iterations.
func signature(strings []*SegmentString) string {
	// Use a length-prefixed encoding of every coord to avoid ambiguity
	// between [(0,0),(1,1)] and [(0,0,1),(1)].
	buf := make([]byte, 0, 32*len(strings))
	for _, ss := range strings {
		buf = appendInt(buf, len(ss.Coords))
		for _, p := range ss.Coords {
			buf = appendXY(buf, p)
		}
		buf = append(buf, ';')
	}
	return string(buf)
}

func appendInt(b []byte, n int) []byte {
	var tmp [20]byte
	pos := len(tmp)
	if n == 0 {
		return append(b, '0', ',')
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	for n > 0 {
		pos--
		tmp[pos] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		pos--
		tmp[pos] = '-'
	}
	b = append(b, tmp[pos:]...)
	return append(b, ',')
}

func appendXY(b []byte, p geom.XY) []byte {
	// Use raw float64 bit pattern for exact equality.
	bx := floatBitsHex(p.X)
	by := floatBitsHex(p.Y)
	b = append(b, bx[:]...)
	b = append(b, ':')
	b = append(b, by[:]...)
	return append(b, '|')
}

func floatBitsHex(f float64) [16]byte {
	const hex = "0123456789abcdef"
	bits := math.Float64bits(f)
	var out [16]byte
	for i := 15; i >= 0; i-- {
		out[i] = hex[bits&0xF]
		bits >>= 4
	}
	return out
}
