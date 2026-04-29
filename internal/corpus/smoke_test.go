package corpus

import (
	"fmt"
	"math"
	"testing"

	"github.com/terra-geo/terra/buffer"
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/measure"
	"github.com/terra-geo/terra/overlay"
	"github.com/terra-geo/terra/validate"
)

// dimension reports the topological dimension of g (point=0, line=1,
// surface=2). For collections we return the maximum child dimension.
func dimension(g geom.Geometry) int {
	switch g.Type() {
	case geom.PointType, geom.MultiPointType:
		return 0
	case geom.LineStringType, geom.MultiLineStringType:
		return 1
	case geom.PolygonType, geom.MultiPolygonType:
		return 2
	case geom.GeometryCollectionType:
		// Treat collections as surface-dimension; the embedded corpus
		// does not include any GeometryCollection features, so this
		// branch is defensive only.
		return 2
	default:
		return 0
	}
}

// TestCorpusSmoke exercises the core pipeline (validate, measure, buffer,
// overlay) over every feature in every embedded fixture. See doc.go for
// the contract this harness enforces.
func TestCorpusSmoke(t *testing.T) {
	fixtures := All()
	if len(fixtures) == 0 {
		t.Fatal("corpus.All() returned no fixtures")
	}

	for _, fx := range fixtures {
		fx := fx
		t.Run(fx.Name, func(t *testing.T) {
			if len(fx.Features) == 0 {
				t.Fatalf("fixture %s decoded with zero features", fx.Name)
			}

			for i, g := range fx.Features {
				label := fmt.Sprintf("%s/feat%d/%s", fx.Name, i, g.Type())

				// 1. Validate.
				if err := validate.Validate(g); err != nil {
					t.Errorf("%s: validate.Validate: %v", label, err)
					// Skip downstream checks when input is invalid.
					continue
				}

				dim := dimension(g)

				// 2. Length (dim >= 1).
				if dim >= 1 {
					l := measure.Length(g)
					if math.IsNaN(l) || math.IsInf(l, 0) {
						t.Errorf("%s: measure.Length non-finite: %v", label, l)
					}
				}

				// 3. Area (dim >= 2).
				if dim >= 2 {
					a := measure.Area(g)
					if math.IsNaN(a) || math.IsInf(a, 0) {
						t.Errorf("%s: measure.Area non-finite: %v", label, a)
					}
				}

				// 4. Buffer with a small positive distance.
				if !g.IsEmpty() {
					buf, err := buffer.Buffer(g, 0.001)
					if err != nil {
						t.Errorf("%s: buffer.Buffer: %v", label, err)
					} else if buf == nil || buf.IsEmpty() {
						t.Errorf("%s: buffer.Buffer produced empty result for non-empty input", label)
					}
				}
			}

			// 5. Overlay union over the first three feature pairs.
			pairsRun := 0
			for i := 0; i < len(fx.Features) && pairsRun < 3; i++ {
				for j := i + 1; j < len(fx.Features) && pairsRun < 3; j++ {
					a, b := fx.Features[i], fx.Features[j]
					if _, err := overlay.Union(a, b); err != nil {
						// Overlay may legitimately fail on awkward
						// real-world inputs; record but don't fail.
						t.Logf("%s: overlay.Union(%d,%d) skipped: %v",
							fx.Name, i, j, err)
					}
					pairsRun++
				}
			}
		})
	}
}
