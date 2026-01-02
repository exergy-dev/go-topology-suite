package geom

import "math"

// OrientationIndex computes the orientation of three points.
// Returns -1 for clockwise, 0 for collinear, 1 for counter-clockwise.
func OrientationIndex(p1, p2, p3 Coordinate) int {
	dx1 := p2.X - p1.X
	dy1 := p2.Y - p1.Y
	dx2 := p3.X - p2.X
	dy2 := p3.Y - p2.Y

	cross := dx1*dy2 - dy1*dx2
	if math.Abs(cross) < DefaultEpsilon {
		return 0
	}
	if cross > 0 {
		return 1
	}
	return -1
}
