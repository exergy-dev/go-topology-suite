package kml

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-topology-suite/gts/geom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_RealKMLFile(t *testing.T) {
	// Read real KML file from Google's samples
	data, err := os.ReadFile(filepath.Join("..", "testdata", "placemark.kml"))
	require.NoError(t, err, "Failed to read test KML file")

	// Test Feature iteration
	count := 0
	for f, err := range UnmarshalFeatures(data) {
		require.NoError(t, err, "Error during iteration")
		
		assert.Equal(t, "My office", f.Name)
		assert.Equal(t, "This is the location of my office.", f.Description)
		assert.NotNil(t, f.Geometry, "Geometry should not be nil")
		
		// Verify coordinates (Mountain View, CA)
		point, ok := f.Geometry.(*geom.Point)
		require.True(t, ok, "Expected Point geometry")
		coord := point.Coordinate()
		assert.InDelta(t, -122.087461, coord.X, 0.0001, "Longitude mismatch")
		assert.InDelta(t, 37.422069, coord.Y, 0.0001, "Latitude mismatch")
		
		count++
	}
	assert.Equal(t, 1, count, "Expected 1 feature")
}

func TestIntegration_MultiplePlacemarks(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "testdata", "multi_placemark.kml"))
	require.NoError(t, err, "Failed to read test KML file")

	// Collect all features
	var features []*Feature
	for f, err := range UnmarshalFeatures(data) {
		require.NoError(t, err, "Error during iteration")
		features = append(features, f)
	}

	// Should have 4 placemarks (3 in Document, 1 in Folder)
	require.Len(t, features, 4, "Expected 4 features")

	// Verify feature properties
	expectedNames := []string{"Google HQ", "Apple Park", "Stanford University", "Golden Gate Bridge"}
	expectedIDs := []string{"p1", "p2", "p3", "p4"}
	
	for i, f := range features {
		assert.Equal(t, expectedNames[i], f.Name, "Name mismatch at index %d", i)
		assert.Equal(t, expectedIDs[i], f.ID, "ID mismatch at index %d", i)
		assert.NotEmpty(t, f.Description, "Description should not be empty at index %d", i)
		assert.NotNil(t, f.Geometry, "Geometry should not be nil at index %d", i)
	}

	// Verify Golden Gate Bridge coordinates
	ggb := features[3]
	ggbPoint, ok := ggb.Geometry.(*geom.Point)
	require.True(t, ok, "Expected Point geometry for Golden Gate Bridge")
	coord := ggbPoint.Coordinate()
	assert.InDelta(t, -122.478255, coord.X, 0.0001, "Golden Gate Bridge longitude")
	assert.InDelta(t, 37.819929, coord.Y, 0.0001, "Golden Gate Bridge latitude")
}

func TestIntegration_RoundTrip(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "testdata", "multi_placemark.kml"))
	require.NoError(t, err)

	// Collect features
	var features []*Feature
	for f, err := range UnmarshalFeatures(data) {
		require.NoError(t, err)
		features = append(features, f)
	}

	// Marshal back to KML
	output, err := MarshalFeatures(features)
	require.NoError(t, err, "Failed to marshal features")
	
	// Parse again
	var roundtripped []*Feature
	for f, err := range UnmarshalFeatures(output) {
		require.NoError(t, err)
		roundtripped = append(roundtripped, f)
	}

	// Verify same count and names
	require.Len(t, roundtripped, len(features), "Round-trip feature count mismatch")
	for i, f := range roundtripped {
		assert.Equal(t, features[i].Name, f.Name, "Name mismatch at index %d", i)
		assert.Equal(t, features[i].Description, f.Description, "Description mismatch at index %d", i)
	}
}
