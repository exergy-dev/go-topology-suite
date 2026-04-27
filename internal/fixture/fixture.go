package fixture

import (
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/robert-malhotra/go-topology-suite/io/wkt"
	"github.com/stretchr/testify/require"
)

// WKTCase describes a source-backed geometry operation fixture.
type WKTCase struct {
	Name           string
	Operation      string
	Predicate      string
	A              string
	B              string
	PrecisionModel geom.PrecisionModel
	ExpectedWKT    string
	ExpectedDE9IM  string
	Source         string
	DimA           int
	DimB           int
	Intersects     bool
	Disjoint       bool
	Touches        bool
	Contains       bool
	Within         bool
}

// MustGeometry parses WKT text and fails the test on invalid input.
func MustGeometry(t *testing.T, text string) geom.Geometry {
	t.Helper()

	g, err := wkt.UnmarshalString(text)
	require.NoError(t, err)
	return g
}
