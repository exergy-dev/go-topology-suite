// Port of org.locationtech.jts.geom.util.SineStarFactory.
//
// Creates polygons shaped like multi-armed stars where each arm is a
// complete sine-wave cycle. Useful as a non-trivial test geometry for
// algorithms (overlay, simplification, buffering, ...).

package shape

import (
	"math"

	"github.com/exergy-dev/go-topology-suite/geom"
)

// SineStarOptions captures the JTS SineStarFactory parameters that
// don't have natural defaults at the call site. Use SineStar for the
// common case (default nPts and armLengthRatio).
type SineStarOptions struct {
	// NumPoints is the total number of vertices around the star.
	// Defaults to 100 if zero or negative.
	NumPoints int
	// ArmLengthRatio is the ratio of arm length to total radius, in
	// [0, 1]. Defaults to 0.5 if zero (clamped silently if outside).
	ArmLengthRatio float64
}

// SineStar returns a sinusoidal star polygon centred at centre, with
// bounding-box width = size, with nArms peaks. Uses JTS defaults
// (100 points, arm length ratio 0.5).
//
// JTS: SineStarFactory.create(origin, size, 100, nArms, 0.5).
func SineStar(centre geom.XY, size float64, nArms int) *geom.Polygon {
	return SineStarWithOptions(centre, size, nArms, SineStarOptions{})
}

// SineStarWithOptions is SineStar with caller-controlled point count
// and arm-length ratio.
func SineStarWithOptions(centre geom.XY, size float64, nArms int, opt SineStarOptions) *geom.Polygon {
	nPts := opt.NumPoints
	if nPts <= 0 {
		nPts = 100
	}
	armRatio := opt.ArmLengthRatio
	if armRatio == 0 {
		armRatio = 0.5
	}
	if armRatio < 0 {
		armRatio = 0
	}
	if armRatio > 1 {
		armRatio = 1
	}
	if size <= 0 || nArms <= 0 {
		return geom.NewEmptyPolygon(nil, geom.LayoutXY)
	}

	radius := size / 2
	armMaxLen := armRatio * radius
	insideRadius := (1 - armRatio) * radius

	pts := make([]geom.XY, nPts+1)
	for i := 0; i < nPts; i++ {
		// Fraction of the way through the current arm in [0,1].
		ptArcFrac := (float64(i) / float64(nPts)) * float64(nArms)
		armAngFrac := ptArcFrac - math.Floor(ptArcFrac)

		// Each arm is a complete sine-wave cycle.
		armAng := 2 * math.Pi * armAngFrac
		armLenFrac := (math.Cos(armAng) + 1.0) / 2.0
		curveRadius := insideRadius + armMaxLen*armLenFrac

		ang := float64(i) * (2 * math.Pi / float64(nPts))
		pts[i] = geom.XY{
			X: curveRadius*math.Cos(ang) + centre.X,
			Y: curveRadius*math.Sin(ang) + centre.Y,
		}
	}
	// Close the ring.
	pts[nPts] = pts[0]
	return geom.NewPolygon(nil, pts)
}
