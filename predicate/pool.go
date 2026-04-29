// Pool helpers thinly wrap internal/xybuf for the ring-snapshot scratch
// used by point-in-polygon, segment-pair scanning, and the relate
// matrix builders. The aliases keep call sites terse; xybuf owns the
// buffer cap and oversize-drop policy.
//
// Borrowed buffers must NOT escape the calling function — contents are
// reused on the next borrow.

package predicate

import (
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/internal/xybuf"
)

func borrowRingBuf() *[]geom.XY {
	return xybuf.Borrow()
}

func releaseRingBuf(buf *[]geom.XY) {
	xybuf.Release(buf)
}
