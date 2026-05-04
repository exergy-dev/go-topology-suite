package buffer

import (
	"math"
	"testing"

	"github.com/exergy-dev/go-topology-suite/geom"
)

func TestOffsetCurveStraightLine(t *testing.T) {
	ls := geom.NewLineString(nil, []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}})
	out := OffsetCurve(ls, 1.0).(*geom.LineString)
	// LEFT side of (0,0)→(10,0) is +Y. Both endpoints should be at y=1.
	if math.Abs(out.PointAt(0).Y-1.0) > 1e-9 || math.Abs(out.PointAt(out.NumPoints()-1).Y-1.0) > 1e-9 {
		t.Errorf("positive offset of horizontal line should sit at y=+distance; got %+v",
			[]geom.XY{out.PointAt(0), out.PointAt(out.NumPoints() - 1)})
	}
}

func TestOffsetCurveNegativeDistanceRightSide(t *testing.T) {
	ls := geom.NewLineString(nil, []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}})
	out := OffsetCurve(ls, -1.0).(*geom.LineString)
	if math.Abs(out.PointAt(0).Y-(-1.0)) > 1e-9 || math.Abs(out.PointAt(out.NumPoints()-1).Y-(-1.0)) > 1e-9 {
		t.Errorf("negative offset of horizontal line should sit at y=-|distance|; got %+v",
			[]geom.XY{out.PointAt(0), out.PointAt(out.NumPoints() - 1)})
	}
}

func TestOffsetCurveZero(t *testing.T) {
	ls := geom.NewLineString(nil, []geom.XY{{X: 0, Y: 0}, {X: 5, Y: 0}})
	out := OffsetCurve(ls, 0).(*geom.LineString)
	if out.NumPoints() != 2 || out.PointAt(0) != (geom.XY{X: 0, Y: 0}) || out.PointAt(1) != (geom.XY{X: 5, Y: 0}) {
		t.Errorf("zero distance should return linework verbatim; got %v", out)
	}
}

func TestOffsetCurveClosedRing(t *testing.T) {
	// CCW square. positive distance => offset on LEFT of forward direction
	// of each edge => INWARD for a CCW ring.
	ring := []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}, {X: 0, Y: 10}, {X: 0, Y: 0}}
	ls := geom.NewLineString(nil, ring)
	out := OffsetCurve(ls, 1.0).(*geom.LineString)
	if !out.IsClosed() {
		t.Fatalf("offset of closed ring should be closed")
	}
	// Inset square should be 8x8 centered at (5,5). Check envelope.
	env := out.Envelope()
	if math.Abs(env.MinX-1) > 1e-6 || math.Abs(env.MaxX-9) > 1e-6 ||
		math.Abs(env.MinY-1) > 1e-6 || math.Abs(env.MaxY-9) > 1e-6 {
		t.Errorf("inset envelope wrong: %+v", env)
	}
}

func TestOffsetCurvePoint(t *testing.T) {
	p := geom.NewPoint(nil, geom.XY{X: 0, Y: 0})
	out := OffsetCurve(p, 1.0).(*geom.LineString)
	if !out.IsEmpty() {
		t.Errorf("offset of point should be empty linestring")
	}
}

func TestOffsetCurvePolygon(t *testing.T) {
	ring := []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}, {X: 0, Y: 10}, {X: 0, Y: 0}}
	p := geom.NewPolygon(nil, ring)
	out := OffsetCurve(p, 1.0)
	// Polygon with one ring → returns the single offset LineString
	// directly (packOffsetResult collapses len(lines)==1).
	if _, ok := out.(*geom.LineString); !ok {
		t.Errorf("polygon with single ring should produce LineString offset; got %T", out)
	}
}
