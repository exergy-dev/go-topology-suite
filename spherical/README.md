# Spherical Geometry Package

The `spherical` package provides spherical geometry operations for the Go Topology Suite (GTS), integrating with Google's S2 geometry library to deliver accurate calculations on the WGS84 ellipsoid.

## Features

### Distance Calculations
- **Geodesic Distance**: Calculate accurate distances between points on Earth's surface
- **Path Length**: Compute total length of linestrings following great circles
- **Perimeter**: Calculate polygon perimeters on the sphere

### Area Calculations
- **Spherical Area**: Compute polygon areas accounting for Earth's curvature
- **Signed Area**: Determine polygon orientation based on area sign
- **Holes Support**: Correctly handle polygons with interior rings

### Spatial Predicates
- **Contains**: Point-in-polygon tests on the sphere
- **Intersects**: Check if polygons intersect
- **Disjoint/Within/Overlaps**: Standard spatial relationship tests
- **Touches**: Boundary-only intersection detection

### S2 Cell Indexing
- **Cell IDs**: Generate S2 cell identifiers at any level (0-30)
- **Cell Tokens**: String-based cell identifiers for indexing
- **Covering**: Cover geometries with S2 cells for spatial indexing
- **Interior Covering**: Get cells completely contained within geometries

## Installation

The package is part of GTS and requires the S2 geometry library:

```bash
go get github.com/go-topology-suite/gts/spherical
```

## Usage Examples

### Distance Between Cities

```go
import (
    "github.com/go-topology-suite/gts/geom"
    "github.com/go-topology-suite/gts/spherical"
)

// New York City and London
nyc := geom.NewPoint(-74.0060, 40.7128)
london := geom.NewPoint(-0.1278, 51.5074)

// Calculate distance in meters
distance := spherical.Distance(nyc, london)
fmt.Printf("Distance: %.0f km\n", distance/1000) // ~5570 km
```

### Polygon Area

```go
// Create a polygon (coordinates in lon, lat)
ring := geom.NewLinearRingXY(
    -122.5, 37.7,
    -122.3, 37.7,
    -122.3, 37.8,
    -122.5, 37.8,
    -122.5, 37.7,
)
poly := geom.NewPolygon(ring, nil)

// Calculate area in square meters
area := spherical.Area(poly)
fmt.Printf("Area: %.2f km²\n", area/1000000)
```

### Point in Polygon

```go
// Create a region
ring := geom.NewLinearRingXY(
    -122.5, 37.7,
    -122.3, 37.7,
    -122.3, 37.8,
    -122.5, 37.8,
    -122.5, 37.7,
)
sanFranciscoArea := geom.NewPolygon(ring, nil)

// Test if a point is inside
point := geom.NewPoint(-122.4194, 37.7749) // San Francisco
isInside := spherical.Contains(sanFranciscoArea, point)
fmt.Printf("Point in area: %v\n", isInside) // true
```

### S2 Cell Indexing

```go
point := geom.NewPoint(-122.4194, 37.7749)

// Get cell token at different levels
cityLevel := spherical.CellToken(point, 10)       // ~1000 km²
neighborhoodLevel := spherical.CellToken(point, 15) // ~10 km²
buildingLevel := spherical.CellToken(point, 20)    // ~400 m²

fmt.Println("Tokens:", cityLevel, neighborhoodLevel, buildingLevel)

// Cover a polygon with cells for indexing
poly := geom.NewPolygon(ring, nil)
cells := spherical.Covering(poly, 10, 15, 8)
// Use these cells in a spatial index
```

## Coordinate System

All coordinates are expected in **WGS84** (EPSG:4326) format:
- **X coordinate** = Longitude (-180 to 180 degrees)
- **Y coordinate** = Latitude (-90 to 90 degrees)

## S2 Cell Levels

S2 cells are organized in a hierarchy from level 0 (coarsest) to level 30 (finest):

| Level | Cell Size | Use Case |
|-------|-----------|----------|
| 0-2 | Continental | Global analysis |
| 5-7 | Country | National boundaries |
| 10 | City (~1000 km²) | Urban areas |
| 15 | Neighborhood (~10 km²) | Districts |
| 20 | Building (~400 m²) | Addresses |
| 25 | Room (~1 m²) | Indoor positioning |
| 30 | Centimeter (~1 cm²) | Precise locations |

## API Reference

### Distance Functions

- `Distance(p1, p2 *geom.Point) float64` - Geodesic distance in meters
- `DistanceCoords(lon1, lat1, lon2, lat2 float64) float64` - Distance from coordinates
- `Length(ls *geom.LineString) float64` - LineString length in meters
- `Perimeter(poly *geom.Polygon) float64` - Polygon perimeter in meters

### Area Functions

- `Area(poly *geom.Polygon) float64` - Polygon area in square meters
- `SignedArea(poly *geom.Polygon) float64` - Signed area (orientation-dependent)
- `RingArea(ring *geom.LinearRing) float64` - Ring area in square meters
- `Centroid(poly *geom.Polygon) *geom.Point` - Spherical centroid

### Spatial Predicates

- `Contains(poly *geom.Polygon, p *geom.Point) bool` - Point-in-polygon test
- `Intersects(p1, p2 *geom.Polygon) bool` - Polygon intersection test
- `Disjoint(p1, p2 *geom.Polygon) bool` - Non-intersection test
- `Within(p1, p2 *geom.Polygon) bool` - Complete containment test
- `Overlaps(p1, p2 *geom.Polygon) bool` - Partial overlap test
- `Touches(p1, p2 *geom.Polygon) bool` - Boundary-only intersection

### S2 Cell Functions

- `CellID(p *geom.Point) s2.CellID` - Cell ID at max level
- `CellIDAtLevel(p *geom.Point, level int) s2.CellID` - Cell ID at specific level
- `CellToken(p *geom.Point, level int) string` - Cell token string
- `Covering(g geom.Geometry, minLevel, maxLevel, maxCells int) []s2.CellID` - Cell covering
- `CoveringTokens(g geom.Geometry, minLevel, maxLevel, maxCells int) []string` - Token covering
- `GeometryFromCellID(cellID s2.CellID) *geom.Polygon` - Convert cell to polygon

### Conversion Functions

- `ToS2Point(p *geom.Point) s2.Point` - Convert GTS point to S2 point
- `FromS2Point(p s2.Point) *geom.Point` - Convert S2 point to GTS point
- `ToS2Polygon(poly *geom.Polygon) *s2.Polygon` - Convert GTS polygon to S2
- `FromS2Polygon(poly *s2.Polygon) *geom.Polygon` - Convert S2 polygon to GTS

## Performance

Benchmark results on Intel Core i9-14900KF:

```
BenchmarkDistance-32    40,656,997 ops     29.74 ns/op
BenchmarkArea-32         1,000,000 ops   1,054 ns/op
BenchmarkContains-32     1,000,000 ops   1,042 ns/op
BenchmarkCovering-32        55,890 ops  21,444 ns/op
```

## Accuracy

The spherical package uses Google's S2 library which provides:
- **Distance accuracy**: Sub-meter precision for typical distances
- **Area accuracy**: High precision for polygons up to continental scale
- **Robustness**: Handles edge cases like antipodal points and poles

For extremely high-precision requirements (millimeter-level), consider using a proper ellipsoidal model instead of spherical approximations.

## Limitations

1. **Spherical Model**: Uses a spherical Earth approximation. For highest accuracy, use an ellipsoidal model like Vincenty's formulae.
2. **Polygon Validity**: Input polygons should follow GTS validity rules (CCW exterior, CW holes).
3. **S2 Specifics**: Some operations (like `Touches`) are approximations due to S2 API limitations.

## Dependencies

- `github.com/golang/geo/s2` - Google's S2 geometry library
- `github.com/go-topology-suite/gts/geom` - GTS geometry types

## Testing

Run tests:
```bash
go test ./spherical/...
```

Run benchmarks:
```bash
go test ./spherical/... -bench=.
```

View coverage:
```bash
go test ./spherical/... -cover
```

## Use Cases

### Spatial Indexing
Use S2 cell covering to create efficient spatial indexes:
```go
cells := spherical.Covering(polygon, 10, 15, 8)
// Store cells in database for fast spatial queries
```

### Geographic Queries
Calculate distances and areas for geographic data:
```go
distance := spherical.Distance(warehouse, customer)
deliveryTime := distance / averageSpeed
```

### Geofencing
Check if points are within regions:
```go
if spherical.Contains(serviceArea, userLocation) {
    // User is in service area
}
```

### Route Planning
Calculate path lengths:
```go
totalDistance := spherical.Length(route)
estimatedFuel := totalDistance * fuelConsumption
```

## Contributing

Contributions are welcome! Please ensure:
1. All tests pass
2. Code follows Go conventions
3. New features include tests and examples
4. Documentation is updated

## License

Part of the Go Topology Suite project.
