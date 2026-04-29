package predicate

import (
	"testing"

	"github.com/terra-geo/terra/wkt"
)

func TestCoversBoundaryPoint(t *testing.T) {
	poly, _ := wkt.Unmarshal("POLYGON ((0 0, 0 10, 10 10, 10 0, 0 0))")
	boundary, _ := wkt.Unmarshal("POINT (0 5)")
	interior, _ := wkt.Unmarshal("POINT (5 5)")
	exterior, _ := wkt.Unmarshal("POINT (15 5)")

	if got, _ := Covers(poly, boundary); !got {
		t.Errorf("Covers(boundary point) should be true (Contains is false)")
	}
	if got, _ := Covers(poly, interior); !got {
		t.Errorf("Covers(interior point) should be true")
	}
	if got, _ := Covers(poly, exterior); got {
		t.Errorf("Covers(exterior point) should be false")
	}
}

func TestCoversIsContainsAtBoundary(t *testing.T) {
	// Contains is strict (interior only), Covers is inclusive (interior or boundary).
	poly, _ := wkt.Unmarshal("POLYGON ((0 0, 0 10, 10 10, 10 0, 0 0))")
	boundary, _ := wkt.Unmarshal("POINT (0 5)")

	cont, _ := Contains(poly, boundary)
	cov, _ := Covers(poly, boundary)
	if cont {
		t.Errorf("Contains(boundary) = true, want false")
	}
	if !cov {
		t.Errorf("Covers(boundary) = false, want true")
	}
}

func TestCoveredByIsCoversReversed(t *testing.T) {
	poly, _ := wkt.Unmarshal("POLYGON ((0 0, 0 10, 10 10, 10 0, 0 0))")
	pt, _ := wkt.Unmarshal("POINT (5 5)")

	c1, _ := CoveredBy(pt, poly)
	c2, _ := Covers(poly, pt)
	if c1 != c2 {
		t.Errorf("CoveredBy(a,b) != Covers(b,a)")
	}
}

func TestPointCoversPoint(t *testing.T) {
	a, _ := wkt.Unmarshal("POINT (1 2)")
	b, _ := wkt.Unmarshal("POINT (1 2)")
	c, _ := wkt.Unmarshal("POINT (1 3)")
	if got, _ := Covers(a, b); !got {
		t.Errorf("identical points should Cover")
	}
	if got, _ := Covers(a, c); got {
		t.Errorf("distinct points should not Cover")
	}
}
