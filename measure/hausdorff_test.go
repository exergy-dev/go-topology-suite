package measure

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiscreteHausdorffIdentical(t *testing.T) {
	a := mustParse(t, "LINESTRING (0 0, 10 0, 10 10)")
	b := mustParse(t, "LINESTRING (0 0, 10 0, 10 10)")
	d := DiscreteHausdorff(a, b)
	assert.Equal(t, 0.0, d, "DHD of identical geometries must be 0")
}

func TestDiscreteHausdorffTranslated(t *testing.T) {
	a := mustParse(t, "LINESTRING (0 0, 100 0)")
	b := mustParse(t, "LINESTRING (0 10, 100 10)")
	d := DiscreteHausdorff(a, b)
	assert.InDelta(t, 10.0, d, 1e-9, "translated copy y+10 expects DHD=10, got %v", d)
}

func TestDiscreteHausdorffJTSExample(t *testing.T) {
	// JTS javadoc example:
	// A = LINESTRING (0 0, 100 0, 10 100, 10 100)
	// B = LINESTRING (0 100, 0 10, 80 10)
	// DHD(A, B) ~= 22.360679774997898
	a := mustParse(t, "LINESTRING (0 0, 100 0, 10 100, 10 100)")
	b := mustParse(t, "LINESTRING (0 100, 0 10, 80 10)")
	d := DiscreteHausdorff(a, b)
	assert.InDelta(t, 22.360679774997898, d, 1e-9, "JTS example DHD = %v", d)
}

func TestDiscreteHausdorffPointPoint(t *testing.T) {
	a := mustParse(t, "POINT (0 0)")
	b := mustParse(t, "POINT (3 4)")
	assert.Equal(t, 5.0, DiscreteHausdorff(a, b), "point-point DHD == 5")
}

func TestOrientedHausdorffAsymmetric(t *testing.T) {
	// A vertices are all near B; but B has a vertex far from A.
	a := mustParse(t, "LINESTRING (0 0, 1 0)")
	b := mustParse(t, "LINESTRING (0 0, 1 0, 50 0)")
	dAB := OrientedHausdorff(a, b)
	dBA := OrientedHausdorff(b, a)
	assert.Equal(t, 0.0, dAB, "A→B: every A vertex lies on B segment")
	assert.InDelta(t, 49.0, dBA, 1e-9, "B→A: vertex (50,0) is 49 from (1,0)")
}

func TestDiscreteHausdorffEmpty(t *testing.T) {
	a := mustParse(t, "LINESTRING EMPTY")
	b := mustParse(t, "POINT (0 0)")
	assert.Equal(t, 0.0, DiscreteHausdorff(mustParse(t, "POINT EMPTY"), mustParse(t, "POINT EMPTY")))
	assert.True(t, math.IsInf(DiscreteHausdorff(a, b), +1), "one empty → +Inf")
}
