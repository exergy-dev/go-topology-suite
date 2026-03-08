package geom_test

import (
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
)

func TestPointOnSegment_Epsilon(t *testing.T) {
	a := geom.NewCoordinate(0, 0)
	b := geom.NewCoordinate(10, 0)

	on := geom.NewCoordinate(5, geom.DefaultEpsilon/20)
	if !geom.PointOnSegment(on, a, b) {
		t.Error("Expected point within epsilon to be on segment")
	}

	off := geom.NewCoordinate(5, geom.DefaultEpsilon/2)
	if geom.PointOnSegment(off, a, b) {
		t.Error("Expected point beyond epsilon to be off segment")
	}
}

func TestPointInRing_BoundarySeparation(t *testing.T) {
	ring := mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)

	onEdge := geom.NewCoordinate(0, 5)
	if !geom.PointOnRing(onEdge, ring) {
		t.Error("Expected boundary point to be on ring")
	}
	if !geom.PointInRing(onEdge, ring) {
		t.Error("Expected boundary point to be inside ring per ray casting")
	}

	inside := geom.NewCoordinate(5, 5)
	if !geom.PointInRing(inside, ring) {
		t.Error("Expected interior point to be inside ring")
	}
	if geom.PointOnRing(inside, ring) {
		t.Error("Expected interior point to be off ring boundary")
	}
}
