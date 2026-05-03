package shape

import (
	"math"

	"github.com/terra-geo/terra/geom"
)

// GridPoints returns points placed by jittering cell centres of a square
// grid covering env. The number of returned points is at least n (it is
// rounded up to the next perfect square so the grid is square).
//
// jitterFraction in [0,1] controls how much each point is randomised
// inside its cell: 0 returns exact cell-aligned positions (cell origin)
// and 1 returns fully uniform random within each cell. The fraction is
// clamped to the JTS range [0,1].
//
// Geometrically jitterFraction maps to JTS's "1 - gutterFraction":
// JTS RandomPointsInGridBuilder uses gutterFraction to *shrink* the
// usable cell so points stay away from cell boundaries; we expose the
// inverse, which is friendlier when the caller wants "more random" with
// larger fraction.
//
// JTS: org.locationtech.jts.shape.random.RandomPointsInGridBuilder
func GridPoints(n int, env geom.Envelope, jitterFraction float64, opts ...Option) []geom.XY {
	if n <= 0 || env.IsEmpty() {
		return nil
	}
	cfg := newConfig(opts)

	// Pick smallest square grid that fits >= n cells.
	nCells := int(math.Sqrt(float64(n)))
	if nCells*nCells < n {
		nCells++
	}
	if nCells == 0 {
		return nil
	}

	gridDX := env.Width() / float64(nCells)
	gridDY := env.Height() / float64(nCells)

	// Clamp jitter to [0,1]; JTS clamps "gutterFraction" symmetrically.
	jf := jitterFraction
	if jf < 0 {
		jf = 0
	}
	if jf > 1 {
		jf = 1
	}
	// gutterFraction = 1 - jf. Half the gutter on each side.
	gutterFrac := 1.0 - jf
	gutterOffsetX := gridDX * gutterFrac / 2
	gutterOffsetY := gridDY * gutterFrac / 2
	cellDX := jf * gridDX
	cellDY := jf * gridDY

	out := make([]geom.XY, 0, nCells*nCells)
	for i := 0; i < nCells; i++ {
		for j := 0; j < nCells; j++ {
			orgX := env.MinX + float64(i)*gridDX + gutterOffsetX
			orgY := env.MinY + float64(j)*gridDY + gutterOffsetY
			out = append(out, geom.XY{
				X: orgX + cellDX*cfg.random64(),
				Y: orgY + cellDY*cfg.random64(),
			})
		}
	}
	return out
}
