// Package geojson provides GeoJSON encoding/decoding for geometries.
//
// GeoJSON coordinates are always in WGS84 (EPSG:4326) per RFC 7946.
// All geometries parsed by this package will have SRID set to 4326.
//
// This package follows standard Go conventions - use json.Marshal/json.Unmarshal
// with the provided types (Feature, FeatureCollection, Geometry).
//
// Example usage:
//
//	// Marshal a feature
//	f := geojson.NewFeature(point, MyProps{Name: "test"})
//	data, _ := json.Marshal(f)
//
//	// Unmarshal a feature
//	var f geojson.Feature[MyProps]
//	json.Unmarshal(data, &f)
//
//	// Marshal raw geometry
//	data, _ := geojson.MarshalGeometry(polygon)
//
//	// Unmarshal raw geometry
//	g, _ := geojson.UnmarshalGeometry(data)
//	fmt.Println(g.SRID()) // Output: 4326
package geojson

import (
	"encoding/json"

	"github.com/robert-malhotra/go-topology-suite/geom"
)

// SRID4326 is the SRID for WGS84, which is the coordinate reference system
// used by GeoJSON as mandated by RFC 7946.
const SRID4326 = 4326

// DefaultFactory is a geometry factory configured for GeoJSON (WGS84/EPSG:4326).
var DefaultFactory = geom.NewGeometryFactoryWithSRID(SRID4326)

// MarshalGeometry marshals a geometry to GeoJSON bytes.
// For Feature or FeatureCollection, use json.Marshal directly.
func MarshalGeometry(g geom.Geometry) ([]byte, error) {
	return json.Marshal(geometryToMap(g))
}

// MarshalGeometryIndent marshals a geometry with indentation.
func MarshalGeometryIndent(g geom.Geometry, indent string) ([]byte, error) {
	return json.MarshalIndent(geometryToMap(g), "", indent)
}

// UnmarshalGeometry unmarshals GeoJSON bytes to a geometry.
// Handles raw geometry, Feature (extracts geometry), and FeatureCollection (returns GeometryCollection).
// The returned geometry will have SRID set to 4326 (WGS84) per RFC 7946.
func UnmarshalGeometry(data []byte) (geom.Geometry, error) {
	return UnmarshalGeometryWithFactory(data, DefaultFactory)
}

// UnmarshalGeometryWithFactory unmarshals using a custom geometry factory.
func UnmarshalGeometryWithFactory(data []byte, factory *geom.GeometryFactory) (geom.Geometry, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	var typeName string
	if err := json.Unmarshal(raw["type"], &typeName); err != nil {
		return nil, err
	}

	switch typeName {
	case "Feature":
		geomData := raw["geometry"]
		if string(geomData) == "null" {
			return factory.CreateGeometryCollection(nil), nil
		}
		return UnmarshalGeometryWithFactory(geomData, factory)

	case "FeatureCollection":
		var features []json.RawMessage
		if err := json.Unmarshal(raw["features"], &features); err != nil {
			return nil, err
		}
		geoms := make([]geom.Geometry, 0, len(features))
		for _, f := range features {
			g, err := UnmarshalGeometryWithFactory(f, factory)
			if err != nil {
				return nil, err
			}
			if g != nil && !g.IsEmpty() {
				geoms = append(geoms, g)
			}
		}
		return factory.CreateGeometryCollection(geoms), nil

	default:
		return parseGeometry(raw, factory)
	}
}
