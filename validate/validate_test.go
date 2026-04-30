package validate

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/wkt"
)

func TestValidPolygon(t *testing.T) {
	g, _ := wkt.Unmarshal("POLYGON ((0 0, 0 10, 10 10, 10 0, 0 0))")
	assert.NoError(t, Validate(g), "expected valid")
}

func TestUnclosedRing(t *testing.T) {
	// Build an unclosed ring directly (WKT parser would also be unclosed).
	p := geom.NewPolygon(nil, []geom.XY{
		{X: 0, Y: 0}, {X: 0, Y: 10}, {X: 10, Y: 10}, {X: 10, Y: 0},
	})
	err := Validate(p)
	var ve *ValidationError
	require.True(t, errors.As(err, &ve), "expected ValidationError, got %v", err)
	found := false
	for _, d := range ve.Defects {
		if d.Kind == DefectRingNotClosed {
			found = true
		}
	}
	assert.True(t, found, "expected DefectRingNotClosed in %v", ve.Defects)
}

func TestRingTooFewPoints(t *testing.T) {
	p := geom.NewPolygon(nil, []geom.XY{{X: 0, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 0}})
	err := Validate(p)
	var ve *ValidationError
	require.True(t, errors.As(err, &ve), "expected ValidationError")
	assert.Equal(t, DefectRingTooFewPoints, ve.Defects[0].Kind, "expected too-few-points")
}

func TestSelfIntersectingRing(t *testing.T) {
	// Bowtie: edges cross.
	p := geom.NewPolygon(nil, []geom.XY{
		{X: 0, Y: 0}, {X: 10, Y: 10}, {X: 10, Y: 0}, {X: 0, Y: 10}, {X: 0, Y: 0},
	})
	err := Validate(p)
	var ve *ValidationError
	require.True(t, errors.As(err, &ve), "expected ValidationError")
	found := false
	for _, d := range ve.Defects {
		if d.Kind == DefectSelfIntersection {
			found = true
		}
	}
	assert.True(t, found, "expected self-intersection defect")
}

func TestHoleOutsideShell(t *testing.T) {
	outer := []geom.XY{{X: 0, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 1}, {X: 1, Y: 0}, {X: 0, Y: 0}}
	hole := []geom.XY{{X: 5, Y: 5}, {X: 5, Y: 6}, {X: 6, Y: 6}, {X: 6, Y: 5}, {X: 5, Y: 5}}
	p := geom.NewPolygon(nil, outer, hole)
	err := Validate(p)
	var ve *ValidationError
	require.True(t, errors.As(err, &ve), "expected ValidationError")
	assert.Equal(t, DefectHoleOutsideShell, ve.Defects[0].Kind, "expected hole-outside-shell, got %+v", ve.Defects)
}

func TestLineStringTooFew(t *testing.T) {
	ls := geom.NewLineString(nil, []geom.XY{{X: 1, Y: 2}})
	err := Validate(ls)
	var ve *ValidationError
	require.True(t, errors.As(err, &ve), "expected error")
	assert.Equal(t, DefectLineTooFewPoints, ve.Defects[0].Kind, "got %v", ve.Defects)
}

func TestEmptyValid(t *testing.T) {
	g, _ := wkt.Unmarshal("POLYGON EMPTY")
	assert.NoError(t, Validate(g), "empty polygon should validate")
}
