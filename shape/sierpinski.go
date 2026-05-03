package shape

import (
	"github.com/terra-geo/terra/geom"
)

// SierpinskiCarpet generates a Sierpinski carpet at recursion depth
// level fitted to env. The result is returned as a MultiPolygon with a
// single member: an outer ring covering env, plus 8^level square holes
// representing the carpet's removed cells. Returning a MultiPolygon (per
// the task signature) keeps callers uniform with the other fractal
// builders that may decompose into multiple parts.
//
// JTS returns a Polygon; we wrap it in a single-element MultiPolygon.
//
// JTS: org.locationtech.jts.shape.fractal.SierpinskiCarpetBuilder
func SierpinskiCarpet(level int, env geom.Envelope) *geom.MultiPolygon {
	if env.IsEmpty() {
		return geom.NewMultiPolygon(nil)
	}
	if level < 0 {
		level = 0
	}

	// Match JTS getSquareBaseLine / getSquareExtent: anchor at (MinX,MinY)
	// and use the smaller side so the shape fits inside env.
	side := env.Width()
	if env.Height() < side {
		side = env.Height()
	}
	originX := env.MinX
	originY := env.MinY

	shell := []geom.XY{
		{X: originX, Y: originY},
		{X: originX + side, Y: originY},
		{X: originX + side, Y: originY + side},
		{X: originX, Y: originY + side},
		{X: originX, Y: originY},
	}

	holes := make([][]geom.XY, 0)
	holes = addCarpetHoles(level, originX, originY, side, holes)

	rings := make([][]geom.XY, 0, 1+len(holes))
	rings = append(rings, shell)
	rings = append(rings, holes...)
	poly := geom.NewPolygon(nil, rings...)
	return geom.NewMultiPolygon(nil, poly)
}

// addCarpetHoles recursively appends the centre-square hole at each
// depth, mirroring JTS SierpinskiCarpetBuilder.addHoles.
func addCarpetHoles(n int, originX, originY, width float64, holes [][]geom.XY) [][]geom.XY {
	if n < 0 {
		return holes
	}
	n2 := n - 1
	w3 := width / 3.0
	holes = addCarpetHoles(n2, originX, originY, w3, holes)
	holes = addCarpetHoles(n2, originX+w3, originY, w3, holes)
	holes = addCarpetHoles(n2, originX+2*w3, originY, w3, holes)

	holes = addCarpetHoles(n2, originX, originY+w3, w3, holes)
	holes = addCarpetHoles(n2, originX+2*w3, originY+w3, w3, holes)

	holes = addCarpetHoles(n2, originX, originY+2*w3, w3, holes)
	holes = addCarpetHoles(n2, originX+w3, originY+2*w3, w3, holes)
	holes = addCarpetHoles(n2, originX+2*w3, originY+2*w3, w3, holes)

	holes = append(holes, squareRing(originX+w3, originY+w3, w3))
	return holes
}

func squareRing(x, y, w float64) []geom.XY {
	return []geom.XY{
		{X: x, Y: y},
		{X: x + w, Y: y},
		{X: x + w, Y: y + w},
		{X: x, Y: y + w},
		{X: x, Y: y},
	}
}
