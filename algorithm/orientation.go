// Package algorithm provides geometric algorithms for use with geometry types.
package algorithm

import (
	"math"

	"github.com/robert-malhotra/go-topology-suite/geom"
)

// Orientation constants.
const (
	// Clockwise orientation (negative signed area).
	Clockwise = -1
	// Collinear points (zero signed area).
	Collinear = 0
	// CounterClockwise orientation (positive signed area).
	CounterClockwise = 1
)

// OrientationIndex computes the orientation of three points.
// Returns:
//
//	Clockwise (-1) if p1-p2-p3 makes a clockwise turn
//	Collinear (0) if p1-p2-p3 are collinear
//	CounterClockwise (1) if p1-p2-p3 makes a counter-clockwise turn
func OrientationIndex(p1, p2, p3 geom.Coordinate) int {
	return geom.OrientationIndex(p1, p2, p3)
}

// IsCCW returns true if the ring has counter-clockwise orientation.
// Uses the signed area test.
func IsCCW(ring geom.CoordinateSequence) bool {
	return SignedArea(ring) > 0
}

// IsCW returns true if the ring has clockwise orientation.
func IsCW(ring geom.CoordinateSequence) bool {
	return SignedArea(ring) < 0
}

// SignedArea computes the signed area of a ring.
// Returns positive for counter-clockwise rings, negative for clockwise.
func SignedArea(ring geom.CoordinateSequence) float64 {
	return geom.SignedArea(ring)
}

// Angle computes the angle of the vector from p1 to p2.
// Returns the angle in radians, in range (-Pi, Pi].
func Angle(p1, p2 geom.Coordinate) float64 {
	dx := p2.X - p1.X
	dy := p2.Y - p1.Y
	return math.Atan2(dy, dx)
}

// AngleBetween computes the angle formed by the vectors p0-p1 and p1-p2.
// Returns the angle in radians, in range [0, Pi].
func AngleBetween(p0, p1, p2 geom.Coordinate) float64 {
	a1 := Angle(p1, p0)
	a2 := Angle(p1, p2)
	diff := math.Abs(a1 - a2)
	if diff > math.Pi {
		diff = 2*math.Pi - diff
	}
	return diff
}

// AngleBetweenOriented computes the oriented angle between two vectors.
// Returns the angle in radians, in range (-Pi, Pi].
func AngleBetweenOriented(p0, p1, p2 geom.Coordinate) float64 {
	a1 := Angle(p1, p0)
	a2 := Angle(p1, p2)
	return NormalizeAngle(a2 - a1)
}

// NormalizeAngle normalizes an angle to the range (-Pi, Pi].
func NormalizeAngle(angle float64) float64 {
	for angle > math.Pi {
		angle -= 2 * math.Pi
	}
	for angle <= -math.Pi {
		angle += 2 * math.Pi
	}
	return angle
}

// NormalizePositiveAngle normalizes an angle to the range [0, 2*Pi).
func NormalizePositiveAngle(angle float64) float64 {
	for angle >= 2*math.Pi {
		angle -= 2 * math.Pi
	}
	for angle < 0 {
		angle += 2 * math.Pi
	}
	return angle
}

// ToDegrees converts radians to degrees.
func ToDegrees(radians float64) float64 {
	return radians * 180 / math.Pi
}

// ToRadians converts degrees to radians.
func ToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

// AngleOfLine computes the angle of the line from start to end.
func AngleOfLine(start, end geom.Coordinate) float64 {
	return Angle(start, end)
}

// InteriorAngle computes the interior angle between two line segments.
func InteriorAngle(p0, p1, p2 geom.Coordinate) float64 {
	angle := AngleBetween(p0, p1, p2)
	return math.Pi - angle
}

// IsAcute returns true if the angle at p1 is acute (< 90 degrees).
func IsAcute(p0, p1, p2 geom.Coordinate) bool {
	// Vector from p1 to p0
	dx0 := p0.X - p1.X
	dy0 := p0.Y - p1.Y
	// Vector from p1 to p2
	dx2 := p2.X - p1.X
	dy2 := p2.Y - p1.Y
	// Dot product
	dot := dx0*dx2 + dy0*dy2
	return dot > 0
}

// IsObtuse returns true if the angle at p1 is obtuse (> 90 degrees).
func IsObtuse(p0, p1, p2 geom.Coordinate) bool {
	dx0 := p0.X - p1.X
	dy0 := p0.Y - p1.Y
	dx2 := p2.X - p1.X
	dy2 := p2.Y - p1.Y
	dot := dx0*dx2 + dy0*dy2
	return dot < 0
}

// IsRight returns true if the angle at p1 is a right angle.
func IsRight(p0, p1, p2 geom.Coordinate) bool {
	dx0 := p0.X - p1.X
	dy0 := p0.Y - p1.Y
	dx2 := p2.X - p1.X
	dy2 := p2.Y - p1.Y
	dot := dx0*dx2 + dy0*dy2
	return math.Abs(dot) < geom.DefaultEpsilon
}

// TurnDirection returns the turn direction going from p0 through p1 to p2.
func TurnDirection(p0, p1, p2 geom.Coordinate) int {
	return OrientationIndex(p0, p1, p2)
}

// LeftTurn returns true if going from p0 through p1 to p2 is a left turn.
func LeftTurn(p0, p1, p2 geom.Coordinate) bool {
	return OrientationIndex(p0, p1, p2) == CounterClockwise
}

// RightTurn returns true if going from p0 through p1 to p2 is a right turn.
func RightTurn(p0, p1, p2 geom.Coordinate) bool {
	return OrientationIndex(p0, p1, p2) == Clockwise
}

// StraightTurn returns true if p0, p1, p2 are collinear.
func StraightTurn(p0, p1, p2 geom.Coordinate) bool {
	return OrientationIndex(p0, p1, p2) == Collinear
}

// Bisector computes the angle bisector of the angle formed at p1.
// Returns the direction of the bisector as an angle.
func Bisector(p0, p1, p2 geom.Coordinate) float64 {
	a0 := Angle(p1, p0)
	a2 := Angle(p1, p2)
	return NormalizeAngle((a0 + a2) / 2)
}
