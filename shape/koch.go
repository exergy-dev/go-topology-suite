package shape

import (
	"math"

	"github.com/exergy-dev/go-topology-suite/geom"
)

// kochHeightFactor is the height of an equilateral triangle of side 1.
var kochHeightFactor = math.Sin(math.Pi / 3.0)

const (
	kochOneThird  = 1.0 / 3.0
	kochTwoThirds = 2.0 / 3.0
)

// kochThirdHeight is one third of an equilateral triangle's height.
var kochThirdHeight = kochHeightFactor / 3.0

// KochSnowflake returns a closed Polygon tracing a Koch snowflake at
// the given recursion level (0 = plain equilateral triangle), centred
// roughly at centre with the requested base width "size".
//
// The triangle is anchored at (centre.X - size/2, centre.Y); for level >
// 0 the shape is shifted vertically by THIRD_HEIGHT*size so the figure
// is visually centred on centre, matching JTS KochSnowflakeBuilder.
//
// JTS: org.locationtech.jts.shape.fractal.KochSnowflakeBuilder
func KochSnowflake(level int, centre geom.XY, size float64) *geom.Polygon {
	if level < 0 {
		level = 0
	}
	if size <= 0 {
		return geom.NewEmptyPolygon(nil, geom.LayoutXY)
	}

	originX := centre.X - size/2
	originY := centre.Y

	y := originY
	if level > 0 {
		y += kochThirdHeight * size
	}

	p0 := geom.XY{X: originX, Y: y}
	p1 := geom.XY{X: originX + size/2, Y: y + size*kochHeightFactor}
	p2 := geom.XY{X: originX + size, Y: y}

	pts := make([]geom.XY, 0, 3*pow4(level)+1)
	pts = append(pts, p0)
	pts = kochAddSide(level, p0, p1, pts)
	pts = kochAddSide(level, p1, p2, pts)
	pts = kochAddSide(level, p2, p0, pts)
	// Close the ring.
	if pts[0] != pts[len(pts)-1] {
		pts = append(pts, pts[0])
	}
	return geom.NewPolygon(nil, pts)
}

// pow4 returns 4^n.
func pow4(n int) int {
	if n <= 0 {
		return 1
	}
	return 1 << (2 * n)
}

// kochAddSide recursively subdivides segment [p0,p1] and appends the
// generated vertices (excluding p0, including endpoint) to pts.
// Mirrors JTS KochSnowflakeBuilder.addSide.
func kochAddSide(level int, p0, p1 geom.XY, pts []geom.XY) []geom.XY {
	if level == 0 {
		// JTS addSegment only writes p1 (p0 was emitted by the previous
		// side, or by the initial vertex).
		return append(pts, p1)
	}
	bx := p1.X - p0.X
	by := p1.Y - p0.Y

	// midPt = p0 + 0.5*base
	midPt := geom.XY{X: p0.X + 0.5*bx, Y: p0.Y + 0.5*by}

	// heightVec = base * THIRD_HEIGHT, then rotate by +90° (quarter circle).
	hx := bx * kochThirdHeight
	hy := by * kochThirdHeight
	// rotateByQuarterCircle(1) in JTS Vector2D: (x,y) -> (-y, x).
	ox := -hy
	oy := hx
	offsetPt := geom.XY{X: midPt.X + ox, Y: midPt.Y + oy}

	thirdPt := geom.XY{X: p0.X + bx*kochOneThird, Y: p0.Y + by*kochOneThird}
	twoThirdPt := geom.XY{X: p0.X + bx*kochTwoThirds, Y: p0.Y + by*kochTwoThirds}

	n2 := level - 1
	pts = kochAddSide(n2, p0, thirdPt, pts)
	pts = kochAddSide(n2, thirdPt, offsetPt, pts)
	pts = kochAddSide(n2, offsetPt, twoThirdPt, pts)
	pts = kochAddSide(n2, twoThirdPt, p1, pts)
	return pts
}
