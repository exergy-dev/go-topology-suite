package relateng

import (
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/kernel"
	"github.com/terra-geo/terra/kernel/planar"
)

// compareAngle compares the angles of two vectors p and q anchored at
// origin, increasing CCW from the positive X-axis. Returns:
//
//	-1 if angle(P) < angle(Q)
//	 0 if collinear (same direction)
//	+1 if angle(P) > angle(Q)
//
// Port of org.locationtech.jts.algorithm.PolygonNodeTopology.compareAngle.
func compareAngle(origin, p, q geom.XY) int {
	quadP := quadrantOf(origin, p)
	quadQ := quadrantOf(origin, q)
	if quadP > quadQ {
		return 1
	}
	if quadP < quadQ {
		return -1
	}
	// same quadrant — relative orientation determines order
	o := planar.Default.Orient(origin, q, p)
	switch o {
	case kernel.CounterClockwise:
		return 1
	case kernel.Clockwise:
		return -1
	}
	return 0
}

// quadrantOf returns the JTS quadrant number for the directed vector
// origin→p. Numbering matches org.locationtech.jts.geom.Quadrant:
//
//	1 | 0
//	-----
//	2 | 3
func quadrantOf(origin, p geom.XY) int {
	dx := p.X - origin.X
	dy := p.Y - origin.Y
	switch {
	case dx >= 0 && dy >= 0:
		return 0
	case dx < 0 && dy >= 0:
		return 1
	case dx < 0 && dy < 0:
		return 2
	default:
		return 3
	}
}
