package predicate

import (
	"testing"

	"github.com/terra-geo/terra/wkt"
)

// TestRelateNG_EdgePipeline_Agrees verifies that for inputs whose answer
// depends on proper interior segment crossings, RelateNG (UseRelateNG=true)
// produces the same DE-9IM matrix as the legacy path.
func TestRelateNG_EdgePipeline_Agrees(t *testing.T) {
	cases := []struct {
		name string
		a    string
		b    string
	}{
		{
			name: "two crossing lines (X)",
			a:    "LINESTRING (0 0, 10 10)",
			b:    "LINESTRING (0 10, 10 0)",
		},
		{
			name: "polygon edge crosses linestring",
			a:    "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))",
			b:    "LINESTRING (-1 5, 11 5)",
		},
		{
			name: "two polygons sharing a boundary segment",
			a:    "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))",
			b:    "POLYGON ((10 0, 20 0, 20 10, 10 10, 10 0))",
		},
		{
			name: "two overlapping polygons (interior cross)",
			a:    "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))",
			b:    "POLYGON ((5 5, 15 5, 15 15, 5 15, 5 5))",
		},
		{
			name: "T-junction line touching polygon edge",
			a:    "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))",
			b:    "LINESTRING (5 5, 5 0)",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			a, err := wkt.Unmarshal(c.a)
			if err != nil {
				t.Fatalf("decode A: %v", err)
			}
			b, err := wkt.Unmarshal(c.b)
			if err != nil {
				t.Fatalf("decode B: %v", err)
			}
			legacy, err := Relate(a, b)
			if err != nil {
				t.Fatalf("legacy Relate: %v", err)
			}
			ng, err := Relate(a, b, UseRelateNG(true))
			if err != nil {
				t.Fatalf("relateng Relate: %v", err)
			}
			if !matricesEquivalent(legacy, ng) {
				t.Errorf("legacy=%s relateng=%s", legacy, ng)
			}
			if string(legacy) != string(ng) {
				t.Logf("matrices differ but topologically equivalent: legacy=%s relateng=%s", legacy, ng)
			}
		})
	}
}

// matricesEquivalent compares two DE-9IM strings for equivalence,
// treating any cell where one side is '*' (or 'F') as a match if the
// other side is the same dimension or 'F'. RelateNG and the legacy
// path may differ in how they distinguish dimension precisely in
// degenerate cells, but the topological answer must agree.
func matricesEquivalent(a, b DE9IM) bool {
	if len(a) != 9 || len(b) != 9 {
		return string(a) == string(b)
	}
	for i := 0; i < 9; i++ {
		if a[i] == b[i] {
			continue
		}
		// '*' matches anything.
		if a[i] == '*' || b[i] == '*' {
			continue
		}
		return false
	}
	return true
}
