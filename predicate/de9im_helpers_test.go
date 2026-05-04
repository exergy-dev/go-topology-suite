package predicate

import (
	"testing"

	"github.com/exergy-dev/go-topology-suite/wkt"
)

func TestDE9IMNamedHelpers(t *testing.T) {
	a, _ := wkt.Unmarshal("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	b, _ := wkt.Unmarshal("POLYGON ((2 2, 4 2, 4 4, 2 4, 2 2))") // strictly inside
	d, err := Relate(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if !d.IsContains() {
		t.Errorf("IsContains should be true for strictly contained polygon, got %s", d)
	}
	if !d.IsCovers() {
		t.Errorf("IsCovers should be true: %s", d)
	}
	if !d.IsContainsProperly() {
		t.Errorf("IsContainsProperly should be true for strict interior containment: %s", d)
	}
	if !d.IsIntersects() {
		t.Errorf("IsIntersects should be true: %s", d)
	}
	if d.IsDisjoint() {
		t.Errorf("IsDisjoint should be false: %s", d)
	}
	if d.IsEquals() {
		t.Errorf("IsEquals should be false for proper subset: %s", d)
	}

	// Boundary touch — contains true, contains-properly false.
	bb, _ := wkt.Unmarshal("POLYGON ((0 2, 4 2, 4 4, 0 4, 0 2))") // touches a's boundary at x=0
	d2, _ := Relate(a, bb)
	if d2.IsContainsProperly() {
		t.Errorf("IsContainsProperly should be false when boundary contact present: %s", d2)
	}
}

func TestContainsProperly(t *testing.T) {
	a, _ := wkt.Unmarshal("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	bInside, _ := wkt.Unmarshal("POLYGON ((2 2, 4 2, 4 4, 2 4, 2 2))")
	bTouch, _ := wkt.Unmarshal("POLYGON ((0 2, 4 2, 4 4, 0 4, 0 2))")
	bOut, _ := wkt.Unmarshal("POLYGON ((20 20, 25 20, 25 25, 20 25, 20 20))")

	got, _ := ContainsProperly(a, bInside)
	if !got {
		t.Errorf("ContainsProperly(inside) want true")
	}
	got, _ = ContainsProperly(a, bTouch)
	if got {
		t.Errorf("ContainsProperly(boundary touch) want false")
	}
	got, _ = ContainsProperly(a, bOut)
	if got {
		t.Errorf("ContainsProperly(disjoint) want false")
	}
}

func TestPatternConstants(t *testing.T) {
	if PatternAdjacent != "F***1****" {
		t.Errorf("PatternAdjacent constant drift: %q", PatternAdjacent)
	}
	if PatternContainsProperly != "T**FF*FF*" {
		t.Errorf("PatternContainsProperly constant drift: %q", PatternContainsProperly)
	}
	if PatternInteriorIntersects != "T********" {
		t.Errorf("PatternInteriorIntersects constant drift: %q", PatternInteriorIntersects)
	}
}
