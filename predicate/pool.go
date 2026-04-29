package predicate

import (
	"sync"

	"github.com/terra-geo/terra/geom"
)

// ringBufPool reuses []geom.XY scratch buffers for ring snapshots taken
// during predicate evaluation (point-in-polygon, polygon-polygon edge
// scans, line-ring crossings). Each call site borrows a buffer, fills it
// via Polygon.RingInto, then returns it. Buffers grow as needed and are
// released back to the pool capped at maxRingBufCap to avoid unbounded
// retention by GC.
//
// The borrowed buffer must NOT escape the calling function: the contents
// are reused on the next borrow and are not safe to retain.
var ringBufPool = sync.Pool{
	New: func() any {
		buf := make([]geom.XY, 0, 64)
		return &buf
	},
}

// maxRingBufCap caps pooled buffers. A run that processes a single huge
// polygon shouldn't pin a multi-MB scratch buffer in the pool forever.
const maxRingBufCap = 8192

func borrowRingBuf() *[]geom.XY {
	return ringBufPool.Get().(*[]geom.XY)
}

func releaseRingBuf(buf *[]geom.XY) {
	if buf == nil {
		return
	}
	if cap(*buf) > maxRingBufCap {
		// Drop oversized buffers to bound steady-state pool memory.
		return
	}
	*buf = (*buf)[:0]
	ringBufPool.Put(buf)
}
