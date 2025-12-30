package wkt_test

import (
	"testing"

	"github.com/go-topology-suite/gts/geom"
	"github.com/go-topology-suite/gts/io/wkt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalPoint(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected func(*geom.Point) bool
	}{
		{
			name:  "Simple point",
			input: "POINT (1 2)",
			expected: func(p *geom.Point) bool {
				return p.X() == 1 && p.Y() == 2
			},
		},
		{
			name:  "Point with decimals",
			input: "POINT (1.5 2.5)",
			expected: func(p *geom.Point) bool {
				return p.X() == 1.5 && p.Y() == 2.5
			},
		},
		{
			name:  "Point with negative values",
			input: "POINT (-1 -2)",
			expected: func(p *geom.Point) bool {
				return p.X() == -1 && p.Y() == -2
			},
		},
		{
			name:  "Empty point",
			input: "POINT EMPTY",
			expected: func(p *geom.Point) bool {
				return p.IsEmpty()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := wkt.UnmarshalString(tt.input)
			require.NoError(t, err, "Failed to parse")
			p, ok := g.(*geom.Point)
			require.True(t, ok, "Expected Point, got %T", g)
			assert.True(t, tt.expected(p), "Point validation failed for input: %s", tt.input)
		})
	}
}

func TestUnmarshalLineString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected func(*geom.LineString) bool
	}{
		{
			name:  "Simple linestring",
			input: "LINESTRING (0 0, 10 10, 20 0)",
			expected: func(ls *geom.LineString) bool {
				return ls.NumPoints() == 3
			},
		},
		{
			name:  "Empty linestring",
			input: "LINESTRING EMPTY",
			expected: func(ls *geom.LineString) bool {
				return ls.IsEmpty()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := wkt.UnmarshalString(tt.input)
			require.NoError(t, err, "Failed to parse")
			ls, ok := g.(*geom.LineString)
			require.True(t, ok, "Expected LineString, got %T", g)
			assert.True(t, tt.expected(ls), "LineString validation failed for input: %s", tt.input)
		})
	}
}

func TestUnmarshalPolygon(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected func(*geom.Polygon) bool
	}{
		{
			name:  "Simple polygon",
			input: "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))",
			expected: func(p *geom.Polygon) bool {
				return !p.IsEmpty() && p.NumInteriorRings() == 0
			},
		},
		{
			name:  "Polygon with hole",
			input: "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0), (2 2, 8 2, 8 8, 2 8, 2 2))",
			expected: func(p *geom.Polygon) bool {
				return !p.IsEmpty() && p.NumInteriorRings() == 1
			},
		},
		{
			name:  "Empty polygon",
			input: "POLYGON EMPTY",
			expected: func(p *geom.Polygon) bool {
				return p.IsEmpty()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := wkt.UnmarshalString(tt.input)
			require.NoError(t, err, "Failed to parse")
			p, ok := g.(*geom.Polygon)
			require.True(t, ok, "Expected Polygon, got %T", g)
			assert.True(t, tt.expected(p), "Polygon validation failed for input: %s", tt.input)
		})
	}
}

func TestUnmarshalMultiPoint(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected func(*geom.MultiPoint) bool
	}{
		{
			name:  "MultiPoint with parens",
			input: "MULTIPOINT ((0 0), (1 1), (2 2))",
			expected: func(mp *geom.MultiPoint) bool {
				return mp.NumGeometries() == 3
			},
		},
		{
			name:  "Empty multipoint",
			input: "MULTIPOINT EMPTY",
			expected: func(mp *geom.MultiPoint) bool {
				return mp.IsEmpty()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := wkt.UnmarshalString(tt.input)
			require.NoError(t, err, "Failed to parse")
			mp, ok := g.(*geom.MultiPoint)
			require.True(t, ok, "Expected MultiPoint, got %T", g)
			assert.True(t, tt.expected(mp), "MultiPoint validation failed for input: %s", tt.input)
		})
	}
}

func TestUnmarshalMultiLineString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected func(*geom.MultiLineString) bool
	}{
		{
			name:  "MultiLineString",
			input: "MULTILINESTRING ((0 0, 10 10), (20 20, 30 30))",
			expected: func(mls *geom.MultiLineString) bool {
				return mls.NumGeometries() == 2
			},
		},
		{
			name:  "Empty multilinestring",
			input: "MULTILINESTRING EMPTY",
			expected: func(mls *geom.MultiLineString) bool {
				return mls.IsEmpty()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := wkt.UnmarshalString(tt.input)
			require.NoError(t, err, "Failed to parse")
			mls, ok := g.(*geom.MultiLineString)
			require.True(t, ok, "Expected MultiLineString, got %T", g)
			assert.True(t, tt.expected(mls), "MultiLineString validation failed for input: %s", tt.input)
		})
	}
}

func TestUnmarshalMultiPolygon(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected func(*geom.MultiPolygon) bool
	}{
		{
			name:  "MultiPolygon",
			input: "MULTIPOLYGON (((0 0, 10 0, 10 10, 0 10, 0 0)), ((20 20, 30 20, 30 30, 20 30, 20 20)))",
			expected: func(mp *geom.MultiPolygon) bool {
				return mp.NumGeometries() == 2
			},
		},
		{
			name:  "Empty multipolygon",
			input: "MULTIPOLYGON EMPTY",
			expected: func(mp *geom.MultiPolygon) bool {
				return mp.IsEmpty()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := wkt.UnmarshalString(tt.input)
			require.NoError(t, err, "Failed to parse")
			mp, ok := g.(*geom.MultiPolygon)
			require.True(t, ok, "Expected MultiPolygon, got %T", g)
			assert.True(t, tt.expected(mp), "MultiPolygon validation failed for input: %s", tt.input)
		})
	}
}

func TestUnmarshalGeometryCollection(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected func(*geom.GeometryCollection) bool
	}{
		{
			name:  "GeometryCollection",
			input: "GEOMETRYCOLLECTION (POINT (0 0), LINESTRING (0 0, 10 10))",
			expected: func(gc *geom.GeometryCollection) bool {
				return gc.NumGeometries() == 2
			},
		},
		{
			name:  "Empty geometrycollection",
			input: "GEOMETRYCOLLECTION EMPTY",
			expected: func(gc *geom.GeometryCollection) bool {
				return gc.IsEmpty()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := wkt.UnmarshalString(tt.input)
			require.NoError(t, err, "Failed to parse")
			gc, ok := g.(*geom.GeometryCollection)
			require.True(t, ok, "Expected GeometryCollection, got %T", g)
			assert.True(t, tt.expected(gc), "GeometryCollection validation failed for input: %s", tt.input)
		})
	}
}

func TestMarshalPoint(t *testing.T) {
	t.Run("Simple point", func(t *testing.T) {
		p := geom.NewPoint(1, 2)
		result := wkt.MarshalString(p)
		assert.Equal(t, "POINT (1 2)", result)
	})

	t.Run("Empty point", func(t *testing.T) {
		p := geom.NewPointEmpty()
		result := wkt.MarshalString(p)
		assert.Equal(t, "POINT EMPTY", result)
	})
}

func TestMarshalLineString(t *testing.T) {
	t.Run("Simple linestring", func(t *testing.T) {
		ls := geom.NewLineStringXY(0, 0, 10, 10)
		result := wkt.MarshalString(ls)
		assert.Equal(t, "LINESTRING (0 0, 10 10)", result)
	})
}

func TestMarshalPolygon(t *testing.T) {
	t.Run("Simple polygon", func(t *testing.T) {
		shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
		p := geom.NewPolygon(shell, nil)
		result := wkt.MarshalString(p)
		expected := "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))"
		assert.Equal(t, expected, result)
	})
}

func TestMarshalBytes(t *testing.T) {
	p := geom.NewPoint(1, 2)
	data, err := wkt.Marshal(p)
	require.NoError(t, err, "Failed to marshal")
	assert.Equal(t, "POINT (1 2)", string(data))
}

func TestUnmarshalBytes(t *testing.T) {
	data := []byte("POINT (1 2)")
	g, err := wkt.Unmarshal(data)
	require.NoError(t, err, "Failed to unmarshal")
	p, ok := g.(*geom.Point)
	require.True(t, ok, "Expected Point, got %T", g)
	assert.Equal(t, float64(1), p.X())
	assert.Equal(t, float64(2), p.Y())
}

func TestMarshalIndent(t *testing.T) {
	shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	p := geom.NewPolygon(shell, nil)
	data, err := wkt.MarshalIndent(p)
	require.NoError(t, err, "Failed to marshal")
	// Should contain newlines
	assert.Greater(t, len(data), len("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))"), "Expected indented output to be longer")
}

func TestMarshalWithOptions(t *testing.T) {
	p := geom.NewPoint(1.123456789, 2.987654321)

	opts := wkt.Options{
		Precision:       2,
		OutputDimension: 2,
	}
	result := wkt.MarshalStringWithOptions(p, opts)
	assert.Equal(t, "POINT (1.12 2.99)", result)
}

func TestRoundTrip(t *testing.T) {
	tests := []string{
		"POINT (1 2)",
		"LINESTRING (0 0, 10 10, 20 0)",
		"POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))",
		"MULTIPOINT ((0 0), (1 1))",
		"MULTILINESTRING ((0 0, 10 10), (20 20, 30 30))",
		"MULTIPOLYGON (((0 0, 10 0, 10 10, 0 10, 0 0)))",
		"GEOMETRYCOLLECTION (POINT (0 0), LINESTRING (0 0, 10 10))",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			g, err := wkt.UnmarshalString(input)
			require.NoError(t, err, "Failed to parse")
			output := wkt.MarshalString(g)
			// Re-parse to verify
			_, err = wkt.UnmarshalString(output)
			require.NoError(t, err, "Failed to re-parse")
		})
	}
}

func TestUnmarshalWithFactory(t *testing.T) {
	factory := geom.NewGeometryFactoryWithSRID(4326)
	g, err := wkt.UnmarshalStringWithFactory("POINT (1 2)", factory)
	require.NoError(t, err, "Failed to parse")
	// Note: SRID is not encoded in WKT, so factory SRID is not automatically applied
	// unless the factory is configured to do so
	assert.NotNil(t, g, "Expected non-nil geometry")
}

func BenchmarkMarshalPoint(b *testing.B) {
	p := geom.NewPoint(1.5, 2.5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wkt.MarshalString(p)
	}
}

func BenchmarkUnmarshalPoint(b *testing.B) {
	input := "POINT (1.5 2.5)"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wkt.UnmarshalString(input)
	}
}

func BenchmarkMarshalPolygon(b *testing.B) {
	shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	p := geom.NewPolygon(shell, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wkt.MarshalString(p)
	}
}

func BenchmarkUnmarshalPolygon(b *testing.B) {
	input := "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wkt.UnmarshalString(input)
	}
}
