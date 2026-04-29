package predicate

import (
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/internal/xybuf"
)

// borrowRingBuf / releaseRingBuf are package-local conveniences over
// internal/xybuf — kept short because the borrow/release pattern shows
// up at every PIP and segment-scan call site.
//
// The borrowed buffer must NOT escape the calling function: contents
// are reused on the next borrow.

func borrowRingBuf() *[]geom.XY {
	return xybuf.Borrow()
}

func releaseRingBuf(buf *[]geom.XY) {
	xybuf.Release(buf)
}
