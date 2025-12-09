# Go Topology Suite (GTS)

[![Go Tests](https://github.com/go-topology-suite/gts/actions/workflows/test.yml/badge.svg)](https://github.com/go-topology-suite/gts/actions/workflows/test.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/go-topology-suite/gts.svg)](https://pkg.go.dev/github.com/go-topology-suite/gts)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-topology-suite/gts)](https://goreportcard.com/report/github.com/go-topology-suite/gts)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A computational geometry library for Go, providing a native implementation of the functionality found in the Java Topology Suite (JTS). GTS enables creation, manipulation, and analysis of 2D vector geometries according to the OGC Simple Features Specification.

## Features

- **Pure Go** - No CGO dependencies, easy cross-compilation
- **OGC Compliant** - Follows the Simple Features Specification
- **Full Geometry Support** - Point, LineString, Polygon, MultiPoint, MultiLineString, MultiPolygon, GeometryCollection
- **Spatial Predicates** - Intersects, Contains, Within, Overlaps, Touches, Crosses, Covers, CoveredBy, Equals, Disjoint
- **Spatial Operations** - Buffer, Union, Intersection, Difference, Symmetric Difference
- **Spherical Geometry** - Full spherical predicate support for WGS84 coordinates using Google's S2 library
- **CRS Support** - Coordinate Reference System support with EPSG registry and transformations
- **I/O Formats** - WKT, WKB, and GeoJSON
- **Spatial Indexes** - STR-tree, Quadtree, and KD-tree for fast spatial queries
- **Algorithms** - Convex hull, simplification (Douglas-Peucker), distance calculations, and more

## Installation

```bash
go get github.com/go-topology-suite/gts
```

Requires Go 1.21 or later.

## Quick Start

### Creating Geometries

```go
package main

import (
    "fmt"
    "github.com/go-topology-suite/gts/geom"
    "github.com/go-topology-suite/gts/io/wkt"
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
    "github.com/go-topology-suite/gts/geom"
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
    "github.com/go-topology-suite/gts/geom"
    "github.com/go-topology-suite/gts/spherical"
    "github.com/go-topology-suite/gts/geodetic"
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
    dist := geodetic.VincentyDistance(
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
    "github.com/go-topology-suite/gts/io/geojson"
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

### Spatial Indexing

```go
package main

import (
    "fmt"
    "github.com/go-topology-suite/gts/geom"
    "github.com/go-topology-suite/gts/index/strtree"
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
        tree.Insert(point.Envelope(), point)
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
    "github.com/go-topology-suite/gts/geom"
    "github.com/go-topology-suite/gts/transform/projection"
)

func main() {
    // Project WGS84 coordinates to Web Mercator
    webMercator := projection.WebMercator{}

    lon, lat := -122.4194, 37.7749
    x, y := webMercator.Forward(lon, lat)
    fmt.Printf("WGS84 (%.4f, %.4f) -> Web Mercator (%.2f, %.2f)\n", lon, lat, x, y)

    // Inverse projection
    lon2, lat2 := webMercator.Inverse(x, y)
    fmt.Printf("Web Mercator (%.2f, %.2f) -> WGS84 (%.4f, %.4f)\n", x, y, lon2, lat2)
}
```

## Package Structure

```
gts/
├── geom/              Core geometry types (Point, LineString, Polygon, etc.)
├── algorithm/         Geometric algorithms (distance, orientation, convex hull)
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
│   └── geojson/       GeoJSON
├── crs/               Coordinate Reference Systems
│   └── epsg/          EPSG registry
├── geodetic/          Geodetic calculations (Vincenty, Haversine)
├── spherical/         Spherical geometry operations (S2-based)
├── transform/         Coordinate transformations
│   └── projection/    Map projections (Mercator, UTM)
├── precision/         Precision models
├── noding/            Line segment noding
└── planar/            Planar graph structures
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
