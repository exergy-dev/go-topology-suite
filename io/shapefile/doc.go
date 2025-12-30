// Package shapefile provides reading and writing of ESRI Shapefile format,
// a popular geospatial vector data format.
//
// This package focuses on geometry I/O only and does not handle DBF attribute
// data. For full shapefile support including attributes, use the underlying
// github.com/jonas-p/go-shp library directly.
//
// # Basic Usage
//
// Reading geometries from a shapefile:
//
//	reader, err := shapefile.NewReader("input.shp")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer reader.Close()
//
//	for reader.Next() {
//	    geom, err := reader.Geometry()
//	    if err != nil {
//	        log.Println("Error reading geometry:", err)
//	        continue
//	    }
//	    // Process geometry...
//	}
//
// Writing geometries to a shapefile:
//
//	writer, err := shapefile.NewWriter("output.shp", shapefile.ShapeTypePolygon)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer writer.Close()
//
//	for _, geom := range geometries {
//	    if err := writer.Write(geom); err != nil {
//	        log.Println("Error writing geometry:", err)
//	    }
//	}
//
// # Convenience Functions
//
// For simple use cases, ReadAll and WriteAll provide one-line operations:
//
//	// Read all geometries
//	geometries, err := shapefile.ReadAll("input.shp")
//
//	// Write all geometries
//	err = shapefile.WriteAll("output.shp", geometries)
//
// # Shapefile Format Notes
//
// Shapefiles store geometries in a binary format with the following characteristics:
//   - All geometries in a single shapefile must be the same type
//   - Multi-part polylines become MultiLineString geometries
//   - Multi-part polygons become MultiPolygon or Polygon with holes
//   - Ring orientation determines exterior vs hole: clockwise = exterior, counter-clockwise = hole
//
// # Supported Shape Types
//
// This package supports the following shapefile geometry types:
//   - Point (ShapeTypePoint)
//   - PolyLine (ShapeTypePolyLine) - mapped to LineString or MultiLineString
//   - Polygon (ShapeTypePolygon) - mapped to Polygon or MultiPolygon
//   - MultiPoint (ShapeTypeMultiPoint)
//   - PointZ, PolyLineZ, PolygonZ, MultiPointZ - 3D variants
package shapefile
