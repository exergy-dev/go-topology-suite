// Package kml provides KML (Keyhole Markup Language) encoding/decoding for geometries.
//
// KML is an XML-based format used by Google Earth and other geospatial applications.
// Coordinates are always in WGS84 (EPSG:4326), with longitude first (X), then latitude (Y),
// and optional altitude (Z).
//
// Basic usage:
//
//	// Marshal a geometry to KML
//	data, err := kml.Marshal(point)
//
//	// Unmarshal KML to a geometry
//	geom, err := kml.Unmarshal(data)
//
// Coordinate format in KML is "lon,lat[,alt]" with tuples separated by whitespace:
//
//	<coordinates>-122.084,37.422,0 -122.085,37.423,0</coordinates>
//
// All parsed geometries will have SRID set to 4326 (WGS84).
package kml
