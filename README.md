# Go Topology Suite (GTS)

[![Go Tests](https://github.com/robert-malhotra/go-topology-suite/actions/workflows/test.yml/badge.svg)](https://github.com/robert-malhotra/go-topology-suite/actions/workflows/test.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/robert-malhotra/go-topology-suite.svg)](https://pkg.go.dev/github.com/robert-malhotra/go-topology-suite)
[![Go Report Card](https://goreportcard.com/badge/github.com/robert-malhotra/go-topology-suite)](https://goreportcard.com/report/github.com/robert-malhotra/go-topology-suite)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A computational geometry library for Go, providing a native implementation of the functionality found in the Java Topology Suite (JTS). GTS enables creation, manipulation, and analysis of 2D vector geometries according to the OGC Simple Features Specification.

## Features

### Core Capabilities
- **Pure Go** - No CGO dependencies, easy cross-compilation
- **OGC Compliant** - Follows the Simple Features Specification
- **Full Geometry Support** - Point, LineString, Polygon, MultiPoint, MultiLineString, MultiPolygon, GeometryCollection

### Spatial Analysis
- **Spatial Predicates** - Intersects, Contains, Within, Overlaps, Touches, Crosses, Covers, CoveredBy, Equals, Disjoint
- **DE-9IM Relations** - Full relate operation with intersection matrix support
- **Spatial Operations** - Buffer, Union, Intersection, Difference, Symmetric Difference
- **Geometry Validation** - OGC-compliant validity checking

### Geographic Support
- **Spherical Geometry** - Full spherical predicate support for WGS84 coordinates using Google's S2 library
- **Geodetic Calculations** - Vincenty and Haversine distance, geodesic area, bearing, and destination point
- **CRS Support** - Coordinate Reference System support with EPSG registry

### I/O Formats
- **WKT** - Well-Known Text (read/write)
- **WKB** - Well-Known Binary (read/write)
- **GeoJSON** - Feature and FeatureCollection support
- **KML** - Keyhole Markup Language (Google Earth format)
- **Shapefile** - ESRI Shapefile format (read/write)

### Performance
- **Spatial Indexes** - STR-tree, Quadtree, and KD-tree for fast spatial queries
- **Algorithms** - Convex hull, simplification (Douglas-Peucker, Visvalingam-Whyatt), distance calculations
- **Projections** - Mercator, Transverse Mercator, and coordinate transformations

## Installation

```bash
go get github.com/robert-malhotra/go-topology-suite
```

Requires Go 1.21 or later.

## Quick Start

### Creating Geometries

```go
package main

import (
    "fmt"
    "github.com/robert-malhotra/go-topology-suite/geom"
    "github.com/robert-malhotra/go-topology-suite/io/wkt"
)

func main() {
    // Create a geometry factory
    factory := geom.NewGeometryFactory()

    // Create a point
    point := factory.CreatePoint(geom.Coordinate{X: -122.4194, Y: 37.7749})

    // Create a polygon from coordinates
    coords := geom.CoordinateSequence{
        {X: 0, Y: 0},
        {X: 10, Y: 0},
        {X: 10, Y: 10},
        {X: 0, Y: 10},
        {X: 0, Y: 0},
    }
    polygon := factory.CreatePolygon(factory.CreateLinearRing(coords), nil)

    // Parse WKT
    reader := wkt.NewReader()
    geom, _ := reader.Read("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")

    fmt.Println(polygon.String()) // WKT output
}
```

### Spatial Predicates

```go
package main

import (
    "fmt"
    "github.com/robert-malhotra/go-topology-suite/geom"
)

func main() {
    factory := geom.NewGeometryFactory()

    // Create two overlapping polygons
    poly1 := factory.CreatePolygon(
        factory.CreateLinearRing(geom.CoordinateSequence{
            {X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}, {X: 0, Y: 10}, {X: 0, Y: 0},
        }), nil)

    poly2 := factory.CreatePolygon(
        factory.CreateLinearRing(geom.CoordinateSequence{
            {X: 5, Y: 5}, {X: 15, Y: 5}, {X: 15, Y: 15}, {X: 5, Y: 15}, {X: 5, Y: 5},
        }), nil)

    point := factory.CreatePoint(geom.Coordinate{X: 5, Y: 5})

    // Check spatial relationships
    fmt.Println("Intersects:", poly1.Intersects(poly2))   // true
    fmt.Println("Contains:", poly1.Contains(point))       // true
    fmt.Println("Overlaps:", poly1.Overlaps(poly2))       // true
}
```

### Spherical Operations (WGS84)

For geographic coordinates, use the `spherical` package for accurate results on a spherical Earth:

```go
package main

import (
    "fmt"
    "github.com/robert-malhotra/go-topology-suite/geom"
    "github.com/robert-malhotra/go-topology-suite/spherical"
    "github.com/robert-malhotra/go-topology-suite/geodetic"
)

func main() {
    factory := geom.NewGeometryFactoryWithSRID(4326) // WGS84

    // San Francisco
    sf := factory.CreatePoint(geom.Coordinate{X: -122.4194, Y: 37.7749})

    // A polygon around the Bay Area
    bayArea := factory.CreatePolygon(
        factory.CreateLinearRing(geom.CoordinateSequence{
            {X: -122.6, Y: 37.4},
            {X: -121.8, Y: 37.4},
            {X: -121.8, Y: 37.9},
            {X: -122.6, Y: 37.9},
            {X: -122.6, Y: 37.4},
        }), nil)

    // Spherical predicates
    fmt.Println("SF in Bay Area:", spherical.Contains(bayArea, sf)) // true

    // Geodesic distance (meters)
    oakland := factory.CreatePoint(geom.Coordinate{X: -122.2711, Y: 37.8044})
    dist := geodetic.DistanceWGS84(
        sf.Coordinate().Y, sf.Coordinate().X,
        oakland.Coordinate().Y, oakland.Coordinate().X,
    )
    fmt.Printf("SF to Oakland: %.2f km\n", dist/1000)
}
```

### GeoJSON Support

```go
package main

import (
    "fmt"
    "github.com/robert-malhotra/go-topology-suite/io/geojson"
)

func main() {
    // Parse GeoJSON (automatically uses WGS84/SRID 4326)
    data := []byte(`{
        "type": "Feature",
        "geometry": {
            "type": "Point",
            "coordinates": [-122.4194, 37.7749]
        },
        "properties": {"name": "San Francisco"}
    }`)

    feature, _ := geojson.UnmarshalFeature(data)
    fmt.Println("Geometry:", feature.Geometry.String())

    // Marshal back to GeoJSON
    output, _ := geojson.MarshalFeature(feature)
    fmt.Println(string(output))
}
```

### KML Support

```go
package main

import (
    "fmt"
    "github.com/robert-malhotra/go-topology-suite/geom"
    "github.com/robert-malhotra/go-topology-suite/io/kml"
)

func main() {
    factory := geom.NewGeometryFactoryWithSRID(4326)
    point := factory.CreatePoint(geom.Coordinate{X: -122.4194, Y: 37.7749})

    // Marshal to KML
    data, _ := kml.Marshal(point)
    fmt.Println(string(data))

    // Unmarshal KML
    geom, _ := kml.Unmarshal(data)
    fmt.Println(geom.String())
}
```

### Shapefile Support

```go
package main

import (
    "fmt"
    "github.com/robert-malhotra/go-topology-suite/io/shapefile"
)

func main() {
    // Read all geometries from a shapefile
    geometries, err := shapefile.ReadAll("input.shp")
    if err != nil {
        panic(err)
    }
    fmt.Printf("Read %d geometries\n", len(geometries))

    // Write geometries to a shapefile
    err = shapefile.WriteAll("output.shp", geometries)
    if err != nil {
        panic(err)
    }
}
```

### Spatial Indexing

```go
package main

import (
    "fmt"
    "github.com/robert-malhotra/go-topology-suite/geom"
    "github.com/robert-malhotra/go-topology-suite/index/strtree"
)

func main() {
    factory := geom.NewGeometryFactory()

    // Create an STR-tree index
    tree := strtree.New(10) // node capacity

    // Insert geometries
    for i := 0; i < 1000; i++ {
        point := factory.CreatePoint(geom.Coordinate{
            X: float64(i % 100),
            Y: float64(i / 100),
        })
        if err := tree.Insert(point.Envelope(), point); err != nil {
            panic(err)
        }
    }
    tree.Build()

    // Query by bounding box
    queryEnv := geom.NewEnvelope(10, 20, 10, 20)
    results := tree.Query(queryEnv)
    fmt.Printf("Found %d geometries in query region\n", len(results))
}
```

### Coordinate Transformations

```go
package main

import (
    "fmt"
    "github.com/robert-malhotra/go-topology-suite/transform/projection"
)

func main() {
    // Project WGS84 coordinates to Web Mercator
    mercator := &projection.Mercator{}

    lon, lat := -122.4194, 37.7749
    x, y, _ := mercator.Forward(lon, lat)
    fmt.Printf("WGS84 (%.4f, %.4f) -> Mercator (%.2f, %.2f)\n", lon, lat, x, y)

    // Inverse projection
    lon2, lat2, _ := mercator.Inverse(x, y)
    fmt.Printf("Mercator (%.2f, %.2f) -> WGS84 (%.4f, %.4f)\n", x, y, lon2, lat2)
}
```

### Geometry Simplification

```go
package main

import (
    "fmt"
    "github.com/robert-malhotra/go-topology-suite/algorithm"
    "github.com/robert-malhotra/go-topology-suite/geom"
)

func main() {
    factory := geom.NewGeometryFactory()

    // Create a line with many points
    coords := geom.CoordinateSequence{
        {X: 0, Y: 0}, {X: 1, Y: 0.1}, {X: 2, Y: -0.1},
        {X: 3, Y: 5}, {X: 4, Y: 6}, {X: 5, Y: 7},
    }
    line := factory.CreateLineString(coords)

    // Simplify using Douglas-Peucker
    simplified := algorithm.DouglasPeuckerLineString(line, 1.0)
    fmt.Printf("Original: %d points, Simplified: %d points\n",
        line.NumPoints(), simplified.NumPoints())
}
```

## Package Structure

```
gts/
├── geom/              Core geometry types (Point, LineString, Polygon, etc.)
├── algorithm/         Geometric algorithms
│   ├── area.go        Area and centroid calculations
│   ├── convexhull.go  Convex hull computation
│   ├── distance.go    Distance calculations
│   ├── intersection.go Line intersection algorithms
│   ├── locate.go      Point location algorithms
│   ├── orientation.go Orientation and angle calculations
│   └── simplify.go    Douglas-Peucker, Visvalingam-Whyatt, Radial Distance
├── operation/         High-level operations
│   ├── buffer/        Buffer operations
│   ├── overlay/       Boolean operations (union, intersection, difference)
│   ├── relate/        DE-9IM spatial relationships
│   ├── valid/         Geometry validation
│   ├── polygonize/    Build polygons from lines
│   └── linemerge/     Merge line segments
├── index/             Spatial indexes
│   ├── strtree/       STR-tree (Sort-Tile-Recursive)
│   ├── quadtree/      Quadtree
│   └── kdtree/        KD-tree for point data
├── io/                Input/output formats
│   ├── wkt/           Well-Known Text
│   ├── wkb/           Well-Known Binary
│   ├── geojson/       GeoJSON
│   ├── kml/           KML (Keyhole Markup Language)
│   └── shapefile/     ESRI Shapefile
├── crs/               Coordinate Reference Systems
│   └── epsg/          EPSG registry
├── geodetic/          Geodetic calculations
│   ├── distance.go    Vincenty and Haversine distance
│   ├── area.go        Geodesic polygon area
│   ├── azimuth.go     Bearing calculations
│   ├── destination.go Direct geodesic problem
│   └── ellipsoid.go   Reference ellipsoids (WGS84, etc.)
├── spherical/         Spherical geometry operations (S2-based)
├── transform/         Coordinate transformations
│   └── projection/    Map projections (Mercator, Transverse Mercator)
├── precision/         Precision models
├── noding/            Line segment noding
└── testing/           Testing utilities
```

## Supported Geometry Types

| Type | Description |
|------|-------------|
| Point | Single coordinate |
| LineString | Connected sequence of points |
| LinearRing | Closed LineString (forms polygon boundary) |
| Polygon | Area with optional holes |
| MultiPoint | Collection of Points |
| MultiLineString | Collection of LineStrings |
| MultiPolygon | Collection of Polygons |
| GeometryCollection | Heterogeneous collection |

## Spatial Predicates

All predicates are available in both planar (`geom` package) and spherical (`spherical` package) variants:

| Predicate | Description |
|-----------|-------------|
| Equals | Geometries are topologically equal |
| Disjoint | No points in common |
| Intersects | Share at least one point |
| Touches | Share boundary but not interior |
| Crosses | Intersect with different dimensions |
| Within | All points of A in B |
| Contains | All points of B in A |
| Overlaps | Share some but not all points |
| Covers | B is within A including boundary |
| CoveredBy | A is within B including boundary |

## I/O Format Support

| Format | Read | Write | Notes |
|--------|------|-------|-------|
| WKT | Yes | Yes | Well-Known Text |
| WKB | Yes | Yes | Well-Known Binary (little/big endian) |
| GeoJSON | Yes | Yes | Feature and FeatureCollection support |
| KML | Yes | Yes | Google Earth format, WGS84 coordinates |
| Shapefile | Yes | Yes | ESRI format, geometry only (no DBF attributes) |

## Geodetic Functions

The `geodetic` package provides accurate calculations on the Earth's surface:

| Function | Description |
|----------|-------------|
| `DistanceWGS84` | Vincenty distance between two points (meters) |
| `Haversine` | Spherical distance (faster, less accurate) |
| `InitialBearing` | Bearing from point A to B |
| `FinalBearing` | Bearing arriving at point B |
| `DestinationPointWGS84` | Point at distance/bearing from origin |
| `PolygonAreaWGS84` | Geodesic area of polygon (sq meters) |

## Performance

GTS is designed for high performance:

- **Zero-allocation** geodetic calculations
- **Spatial indexing** for large datasets (STR-tree recommended)
- **Lazy envelope computation** with caching
- **Efficient coordinate sequences**

Benchmark results (Apple M1):

| Operation | Time | Allocations |
|-----------|------|-------------|
| Vincenty Distance | ~260 ns | 0 |
| Haversine Distance | ~30 ns | 0 |
| Point-in-Polygon (planar) | ~50 ns | 0 |
| Point-in-Polygon (spherical) | ~1.5 μs | 2-3 |
| Polygon Intersection | ~5-50 μs | varies |
| STR-tree Query (10k items) | ~1 μs | 1 |

## Dependencies

- [github.com/golang/geo](https://github.com/golang/geo) - S2 geometry library for spherical operations
- [github.com/jonas-p/go-shp](https://github.com/jonas-p/go-shp) - Shapefile format support
- [github.com/stretchr/testify](https://github.com/stretchr/testify) - Testing utilities (test only)

## Comparison with JTS

GTS follows the JTS architecture and algorithms closely, with Go-specific adaptations:

| Aspect | JTS | GTS |
|--------|-----|-----|
| Language | Java | Go |
| Inheritance | Class hierarchy | Interface composition |
| Errors | Exceptions | Return values |
| Nulls | Null pointers | Nil + ok patterns |
| CRS | Via GeoTools | Built-in (crs/, spherical/) |

## Contributing

Contributions are welcome. Please:

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass: `go test ./...`
5. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

- [JTS Topology Suite](https://github.com/locationtech/jts) - The original Java implementation
- [GEOS](https://libgeos.org/) - C++ port of JTS
- [Google S2 Geometry](https://s2geometry.io/) - Spherical geometry library
- [OGC Simple Features](https://www.ogc.org/standards/sfa) - Standards specification

## Related Projects

- [orb](https://github.com/paulmach/orb) - Alternative Go geometry library
- [go.geojson](https://github.com/paulmach/go.geojson) - GeoJSON library
- [golang/geo](https://github.com/golang/geo) - S2 geometry library
