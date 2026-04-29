package predicate

import (
	"testing"

	"github.com/terra-geo/terra/wkt"
)

func TestCrossesLineLine(t *testing.T) {
	a, _ := wkt.Unmarshal("LINESTRING (0 0, 10 10)")
	b, _ := wkt.Unmarshal("LINESTRING (0 10, 10 0)")
	got, _ := Crosses(a, b)
	if !got {
		t.Errorf("two crossing lines should Cross")
	}

	disjoint, _ := wkt.Unmarshal("LINESTRING (20 0, 30 0)")
	if got, _ := Crosses(a, disjoint); got {
		t.Errorf("disjoint lines should not Cross")
	}

	parallel, _ := wkt.Unmarshal("LINESTRING (0 1, 10 11)")
	if got, _ := Crosses(a, parallel); got {
		t.Errorf("parallel lines should not Cross")
	}
}

func TestCrossesLinePolygon(t *testing.T) {
	poly, _ := wkt.Unmarshal("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	through, _ := wkt.Unmarshal("LINESTRING (-5 5, 15 5)")
	inside, _ := wkt.Unmarshal("LINESTRING (3 3, 7 7)")
	outside, _ := wkt.Unmarshal("LINESTRING (-5 5, -1 5)")

	if got, _ := Crosses(through, poly); !got {
		t.Errorf("line through polygon should Cross")
	}
	if got, _ := Crosses(inside, poly); got {
		t.Errorf("line entirely inside polygon should not Cross")
	}
	if got, _ := Crosses(outside, poly); got {
		t.Errorf("line entirely outside polygon should not Cross")
	}
}

func TestOverlapsPolygons(t *testing.T) {
	a, _ := wkt.Unmarshal("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	b, _ := wkt.Unmarshal("POLYGON ((5 5, 15 5, 15 15, 5 15, 5 5))")
	if got, _ := Overlaps(a, b); !got {
		t.Errorf("partially-overlapping polygons should Overlap")
	}

	contained, _ := wkt.Unmarshal("POLYGON ((2 2, 4 2, 4 4, 2 4, 2 2))")
	if got, _ := Overlaps(a, contained); got {
		t.Errorf("contained polygon should not Overlap")
	}

	disjoint, _ := wkt.Unmarshal("POLYGON ((20 20, 30 20, 30 30, 20 30, 20 20))")
	if got, _ := Overlaps(a, disjoint); got {
		t.Errorf("disjoint polygons should not Overlap")
	}

	equal, _ := wkt.Unmarshal("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	if got, _ := Overlaps(a, equal); got {
		t.Errorf("equal polygons should not Overlap")
	}
}

func TestOverlapsDifferentDimensionsReturnsFalse(t *testing.T) {
	pt, _ := wkt.Unmarshal("POINT (5 5)")
	poly, _ := wkt.Unmarshal("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	if got, _ := Overlaps(pt, poly); got {
		t.Errorf("Point/Polygon Overlaps should be false (different dim)")
	}
}

func TestRelateBasic(t *testing.T) {
	a, _ := wkt.Unmarshal("POINT (1 2)")
	b, _ := wkt.Unmarshal("POINT (1 2)")
	d, err := Relate(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if len(d) != 9 {
		t.Errorf("DE-9IM matrix should be 9 chars, got %d", len(d))
	}
	// Two identical points must intersect (II != F).
	if d[0] == 'F' {
		t.Errorf("identical points should have non-empty II, got %s", d)
	}
}

func TestDE9IMMatches(t *testing.T) {
	d := DE9IM("212111212")
	// Pattern T********: II non-F. Should match.
	if !d.Matches("T********") {
		t.Errorf("expected match for T********")
	}
	// Pattern F********: II must be F. d[0]='2' so should NOT match.
	if d.Matches("F********") {
		t.Errorf("expected mismatch for F********")
	}
}
