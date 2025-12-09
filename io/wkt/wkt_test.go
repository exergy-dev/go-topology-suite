package wkt_test

import (
	"testing"

	"github.com/go-topology-suite/gts/geom"
	"github.com/go-topology-suite/gts/io/wkt"
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
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}
			p, ok := g.(*geom.Point)
			if !ok {
				t.Fatalf("Expected Point, got %T", g)
			}
			if !tt.expected(p) {
				t.Errorf("Point validation failed for input: %s", tt.input)
			}
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
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}
			ls, ok := g.(*geom.LineString)
			if !ok {
				t.Fatalf("Expected LineString, got %T", g)
			}
			if !tt.expected(ls) {
				t.Errorf("LineString validation failed for input: %s", tt.input)
			}
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
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}
			p, ok := g.(*geom.Polygon)
			if !ok {
				t.Fatalf("Expected Polygon, got %T", g)
			}
			if !tt.expected(p) {
				t.Errorf("Polygon validation failed for input: %s", tt.input)
			}
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
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}
			mp, ok := g.(*geom.MultiPoint)
			if !ok {
				t.Fatalf("Expected MultiPoint, got %T", g)
			}
			if !tt.expected(mp) {
				t.Errorf("MultiPoint validation failed for input: %s", tt.input)
			}
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
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}
			mls, ok := g.(*geom.MultiLineString)
			if !ok {
				t.Fatalf("Expected MultiLineString, got %T", g)
			}
			if !tt.expected(mls) {
				t.Errorf("MultiLineString validation failed for input: %s", tt.input)
			}
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
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}
			mp, ok := g.(*geom.MultiPolygon)
			if !ok {
				t.Fatalf("Expected MultiPolygon, got %T", g)
			}
			if !tt.expected(mp) {
				t.Errorf("MultiPolygon validation failed for input: %s", tt.input)
			}
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
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}
			gc, ok := g.(*geom.GeometryCollection)
			if !ok {
				t.Fatalf("Expected GeometryCollection, got %T", g)
			}
			if !tt.expected(gc) {
				t.Errorf("GeometryCollection validation failed for input: %s", tt.input)
			}
		})
	}
}

func TestMarshalPoint(t *testing.T) {
	t.Run("Simple point", func(t *testing.T) {
		p := geom.NewPoint(1, 2)
		result := wkt.MarshalString(p)
		if result != "POINT (1 2)" {
			t.Errorf("Expected 'POINT (1 2)', got '%s'", result)
		}
	})

	t.Run("Empty point", func(t *testing.T) {
		p := geom.NewPointEmpty()
		result := wkt.MarshalString(p)
		if result != "POINT EMPTY" {
			t.Errorf("Expected 'POINT EMPTY', got '%s'", result)
		}
	})
}

func TestMarshalLineString(t *testing.T) {
	t.Run("Simple linestring", func(t *testing.T) {
		ls := geom.NewLineStringXY(0, 0, 10, 10)
		result := wkt.MarshalString(ls)
		if result != "LINESTRING (0 0, 10 10)" {
			t.Errorf("Expected 'LINESTRING (0 0, 10 10)', got '%s'", result)
		}
	})
}

func TestMarshalPolygon(t *testing.T) {
	t.Run("Simple polygon", func(t *testing.T) {
		shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
		p := geom.NewPolygon(shell, nil)
		result := wkt.MarshalString(p)
		expected := "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))"
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})
}

func TestMarshalBytes(t *testing.T) {
	p := geom.NewPoint(1, 2)
	data, err := wkt.Marshal(p)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}
	if string(data) != "POINT (1 2)" {
		t.Errorf("Expected 'POINT (1 2)', got '%s'", string(data))
	}
}

func TestUnmarshalBytes(t *testing.T) {
	data := []byte("POINT (1 2)")
	g, err := wkt.Unmarshal(data)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	p, ok := g.(*geom.Point)
	if !ok {
		t.Fatalf("Expected Point, got %T", g)
	}
	if p.X() != 1 || p.Y() != 2 {
		t.Errorf("Expected (1, 2), got (%v, %v)", p.X(), p.Y())
	}
}

func TestMarshalIndent(t *testing.T) {
	shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	p := geom.NewPolygon(shell, nil)
	data, err := wkt.MarshalIndent(p)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}
	// Should contain newlines
	if len(data) <= len("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))") {
		t.Error("Expected indented output to be longer")
	}
}

func TestMarshalWithOptions(t *testing.T) {
	p := geom.NewPoint(1.123456789, 2.987654321)

	opts := wkt.Options{
		Precision:       2,
		OutputDimension: 2,
	}
	result := wkt.MarshalStringWithOptions(p, opts)
	if result != "POINT (1.12 2.99)" {
		t.Errorf("Expected 'POINT (1.12 2.99)', got '%s'", result)
	}
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
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}
			output := wkt.MarshalString(g)
			// Re-parse to verify
			_, err = wkt.UnmarshalString(output)
			if err != nil {
				t.Fatalf("Failed to re-parse: %v", err)
			}
		})
	}
}

func TestUnmarshalWithFactory(t *testing.T) {
	factory := geom.NewGeometryFactoryWithSRID(4326)
	g, err := wkt.UnmarshalStringWithFactory("POINT (1 2)", factory)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}
	// Note: SRID is not encoded in WKT, so factory SRID is not automatically applied
	// unless the factory is configured to do so
	if g == nil {
		t.Error("Expected non-nil geometry")
	}
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
