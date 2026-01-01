package shapefile

import (
	"path/filepath"
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_RealShapefile(t *testing.T) {
	// Read real shapefile - US Census blockgroups near San Francisco
	// Source: https://github.com/GeospatialPython/pyshp
	filename := filepath.Join("..", "testdata", "blockgroups.shp")

	reader, err := NewReader(filename)
	require.NoError(t, err, "Failed to open shapefile")
	defer reader.Close()

	// Verify shapefile properties
	assert.Equal(t, ShapeTypePolygon, reader.ShapeType(), "Expected polygon shapefile")

	// Get field names
	fields := reader.Fields()
	require.NotEmpty(t, fields, "Expected DBF fields")
	t.Logf("Found %d fields: %v", len(fields), fields[:min(5, len(fields))])

	// Count features and verify we can read geometry + attributes
	count := 0
	for reader.Next() {
		f, err := reader.Feature()
		require.NoError(t, err, "Failed to read feature at index %d", count)

		assert.NotNil(t, f.Geometry, "Geometry should not be nil at index %d", count)
		assert.NotEmpty(t, f.Properties, "Properties should not be empty at index %d", count)
		assert.Equal(t, count, f.Index, "Feature index mismatch")

		// Verify it's a polygon
		_, isPoly := f.Geometry.(*geom.Polygon)
		_, isMultiPoly := f.Geometry.(*geom.MultiPolygon)
		assert.True(t, isPoly || isMultiPoly, "Expected Polygon or MultiPolygon at index %d, got %T", count, f.Geometry)

		count++
		if count >= 10 {
			break // Just check first 10 for speed
		}
	}

	assert.GreaterOrEqual(t, count, 10, "Expected at least 10 features")
	t.Logf("Read %d features successfully", count)
}

func TestIntegration_FeatureIterator(t *testing.T) {
	filename := filepath.Join("..", "testdata", "blockgroups.shp")

	// Use iterator to read features
	count := 0
	var firstFeature *Feature

	for f, err := range Features(filename) {
		require.NoError(t, err, "Error during iteration at feature %d", count)
		
		if count == 0 {
			firstFeature = f
		}

		assert.NotNil(t, f.Geometry, "Geometry nil at index %d", count)
		assert.NotEmpty(t, f.Properties, "Properties empty at index %d", count)

		count++
		if count >= 50 {
			break
		}
	}

	require.NotNil(t, firstFeature, "Should have read at least one feature")
	assert.GreaterOrEqual(t, count, 50, "Expected at least 50 features")
	
	// Log sample attribute values from first feature
	t.Logf("First feature has %d properties", len(firstFeature.Properties))
	for k, v := range firstFeature.Properties {
		t.Logf("  %s: %v", k, v)
		break // Just log one
	}
}

func TestIntegration_ReadAllFields(t *testing.T) {
	filename := filepath.Join("..", "testdata", "blockgroups.shp")

	reader, err := NewReader(filename)
	require.NoError(t, err)
	defer reader.Close()

	// Get and verify fields
	fields := reader.Fields()
	
	// The blockgroups shapefile should have multiple fields
	assert.GreaterOrEqual(t, len(fields), 5, "Expected at least 5 fields")
	
	// Log all field names
	t.Logf("Field names: %v", fields)

	// Read one feature and verify all properties are populated
	require.True(t, reader.Next(), "Expected at least one record")
	
	f, err := reader.Feature()
	require.NoError(t, err)
	
	// Each field should have a corresponding property
	for _, fieldName := range fields {
		_, exists := f.Properties[fieldName]
		assert.True(t, exists, "Property %q should exist", fieldName)
	}
}

func TestIntegration_BoundingBox(t *testing.T) {
	filename := filepath.Join("..", "testdata", "blockgroups.shp")

	reader, err := NewReader(filename)
	require.NoError(t, err)
	defer reader.Close()

	bbox := reader.BoundingBox()
	require.NotNil(t, bbox, "BoundingBox should not be nil")

	// San Francisco area coordinates (approximate)
	// The blockgroups shapefile covers SF Bay Area
	assert.Greater(t, bbox.MaxX, bbox.MinX, "MaxX should be > MinX")
	assert.Greater(t, bbox.MaxY, bbox.MinY, "MaxY should be > MinY")
	
	// Should be somewhere in California (rough check)
	assert.Less(t, bbox.MinX, -100.0, "MinX should be negative (western hemisphere)")
	assert.Greater(t, bbox.MinY, 30.0, "MinY should be > 30 (California latitude)")
	
	t.Logf("Bounding box: (%f, %f) to (%f, %f)", bbox.MinX, bbox.MinY, bbox.MaxX, bbox.MaxY)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
