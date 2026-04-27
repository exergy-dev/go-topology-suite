# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [0.1.0] - 2026-03-08

### Added
- Core geometry types: Point, LineString, LinearRing, Polygon, MultiPoint, MultiLineString, MultiPolygon, GeometryCollection
- Spatial predicates: Intersects, Contains, Within, Overlaps, Touches, Crosses, Covers, CoveredBy, Equals, Disjoint
- DE-9IM relate operation with intersection matrix support
- Spatial operations: Buffer, Union, Intersection, Difference, Symmetric Difference
- Spherical geometry support for WGS84 coordinates (via Google S2)
- Geodetic calculations: Vincenty/Haversine distance, geodesic area, bearing, destination point
- I/O formats: WKT, WKB (with EWKB/SRID support), GeoJSON, KML, Shapefile
- Spatial indexes: STR-tree, Quadtree, KD-tree
- Algorithms: Convex hull, Douglas-Peucker/Visvalingam-Whyatt/Radial simplification, distance calculations
- Coordinate transformations: Mercator, Transverse Mercator projections
- CRS support with EPSG registry
- Geometry validation
- Line merging and polygonization operations
- Precision model support

### Known Limitations
- Geometry objects are not safe for concurrent modification
