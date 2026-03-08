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
- **Crosses/Covers/Equals**: Additional topological relationship checks

### S2 Cell Indexing
- **Cell IDs**: Generate S2 cell identifiers at any level (0-30)
- **Cell Tokens**: String-based cell identifiers for indexing
- **Covering**: Cover geometries with S2 cells for spatial indexing
- **Interior Covering**: Get cells completely contained within geometries

## Installation

The package is part of GTS and requires the S2 geometry library:

```bash
go get github.com/robert-malhotra/go-topology-suite/spherical
```

## Usage Examples

### Distance Between Cities

```go
import (
    "github.com/robert-malhotra/go-topology-suite/geom"
    "github.com/robert-malhotra/go-topology-suite/spherical"
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

All predicate functions support the following geometry types:
- Point
- LineString
- LinearRing
- Polygon
- MultiPoint
- MultiLineString
- MultiPolygon
- GeometryCollection

#### Basic Predicates

- `Contains(g1, g2 geom.Geometry) bool` - Containment test
- `Intersects(g1, g2 geom.Geometry) bool` - Intersection test

#### Relationship Predicates

- `Within(g1, g2 geom.Geometry) bool` - Returns true if g1 is completely within g2
  - Inverse of Contains: `Within(a, b) == Contains(b, a)`
  - All points of g1 must be in the interior or on the boundary of g2

- `Disjoint(g1, g2 geom.Geometry) bool` - Returns true if geometries have no points in common
  - Inverse of Intersects: `Disjoint(a, b) == !Intersects(a, b)`

- `Overlaps(g1, g2 geom.Geometry) bool` - Returns true if geometries partially overlap
  - Both geometries must have the same dimension
  - Neither geometry can completely contain the other

- `Touches(g1, g2 geom.Geometry) bool` - Returns true if geometries touch at boundaries only
  - Geometries must share boundary points but NOT interior points
  - Common for adjacent polygons (like countries or states)

- `Crosses(g1, g2 geom.Geometry) bool` - Returns true if geometries cross each other
  - Typically used for line-line or line-polygon relationships

- `Covers(g1, g2 geom.Geometry) bool` - Returns true if g1 covers g2 (interior or boundary)
- `CoveredBy(g1, g2 geom.Geometry) bool` - Inverse of Covers
- `Equals(g1, g2 geom.Geometry) bool` - Topological equality test

### S2 Cell Functions

- `CellID(p *geom.Point) s2.CellID` - Cell ID at max level
- `CellIDAtLevel(p *geom.Point, level int) s2.CellID` - Cell ID at specific level
- `CellToken(p *geom.Point, level int) string` - Cell token string
- `Covering(g geom.Geometry, minLevel, maxLevel, maxCells int) []s2.CellID` - Cell covering
- `CoveringTokens(g geom.Geometry, minLevel, maxLevel, maxCells int) []string` - Token covering
- `InteriorCovering(g geom.Geometry, minLevel, maxLevel, maxCells int) []s2.CellID` - Interior cell covering (area geometries only)
- `InteriorCoveringTokens(g geom.Geometry, minLevel, maxLevel, maxCells int) []string` - Interior token covering (area geometries only)
- `CellUnion(g geom.Geometry, minLevel, maxLevel, maxCells int) s2.CellUnion` - Normalized cell union
- `GeometryFromCellID(cellID s2.CellID) *geom.Polygon` - Convert cell to polygon

### Conversion Functions

- `ToS2Point(p *geom.Point) s2.Point` - Convert GTS point to S2 point
- `FromS2Point(p s2.Point) *geom.Point` - Convert S2 point to GTS point
- `ToS2Polygon(poly *geom.Polygon) *s2.Polygon` - Convert GTS polygon to S2
- `FromS2Polygon(poly *s2.Polygon) *geom.Polygon` - Convert S2 polygon to GTS

## Spatial Predicates: Detailed Documentation

### Implementation Details

#### Spherical Geometry

All predicates use the S2 library for spherical geometry calculations, which means:
- Calculations are accurate on the Earth's surface
- Correctly handles the antimeridian (180°/-180° longitude line)
- Works properly near the poles
- Distances and areas are calculated on the sphere, not a flat plane

#### Point Location

The implementation includes a comprehensive `locatePointSpherical` function that determines whether a point is:
- **Interior**: Inside the geometry
- **Boundary**: On the edge/boundary of the geometry
- **Exterior**: Outside the geometry

This is crucial for implementing the `Touches` predicate correctly.

#### Tolerance

The implementation uses a 0.1-meter (10 cm) tolerance for coincidence tests. Two points are considered the same if they are within 0.1 meters of each other on the Earth's surface.

#### Interior Intersection Detection

For polygon-polygon relationships, the implementation includes `hasSphericalPolygonInteriorIntersection` which detects when two polygons have overlapping interior areas (not just touching boundaries). This is essential for correctly implementing the `Touches` predicate.

### Comparison with Planar Predicates

The planar predicates in `geom/predicates.go` perform similar operations but use planar (Euclidean) geometry. Key differences:

| Aspect | Planar | Spherical |
|--------|--------|-----------|
| Distance calculation | Euclidean distance | Great circle distance |
| Area calculation | Flat plane | Spherical surface |
| Accuracy | Good for small areas | Accurate anywhere on Earth |
| Antimeridian | Can cause issues | Handled correctly |
| Poles | Can cause issues | Handled correctly |
| Performance | Slightly faster | Slightly slower |

### When to Use Spherical vs Planar Predicates

**Use spherical predicates when:**
- Working with real-world geographic coordinates (latitude/longitude)
- Dealing with large areas where Earth's curvature matters
- Data crosses the antimeridian or includes polar regions
- Accuracy on the Earth's surface is important

**Use planar predicates when:**
- Working with projected coordinates (e.g., UTM, State Plane)
- Performance is critical and areas are small
- Data is already in a planar coordinate system
- Earth's curvature can be safely ignored

### Predicate Examples

#### Point in Polygon

```go
// Check if a point is within a polygon
point := geom.NewPoint(-73.985, 40.748) // Empire State Building
manhattan := geom.NewPolygon(
    geom.NewLinearRing(geom.CoordinateSequence{
        {X: -74.02, Y: 40.70},
        {X: -73.97, Y: 40.70},
        {X: -73.97, Y: 40.80},
        {X: -74.02, Y: 40.80},
        {X: -74.02, Y: 40.70},
    }),
    nil,
)

if spherical.Within(point, manhattan) {
    fmt.Println("Point is in Manhattan")
}
```

#### Adjacent Polygons

```go
// Check if two countries share a border
usa := geom.NewPolygon(...) // USA boundary
canada := geom.NewPolygon(...) // Canada boundary

if spherical.Touches(usa, canada) {
    fmt.Println("USA and Canada share a border")
}

if spherical.Disjoint(usa, canada) {
    fmt.Println("USA and Canada don't touch")
}
```

#### Overlapping Regions

```go
// Check if two delivery zones overlap
zone1 := geom.NewPolygon(...)
zone2 := geom.NewPolygon(...)

if spherical.Overlaps(zone1, zone2) {
    fmt.Println("Zones overlap - need to resolve territory")
}
```

#### Multi-Geometry Support

```go
// Check if multiple points are all within a region
points := geom.NewMultiPoint([]*geom.Point{
    geom.NewPoint(-73.99, 40.75),
    geom.NewPoint(-73.98, 40.76),
    geom.NewPoint(-73.97, 40.74),
})

region := geom.NewPolygon(...)

if spherical.Within(points, region) {
    fmt.Println("All points are within the region")
}
```

## Performance

Benchmark results on Intel Core i9-14900KF:

```
BenchmarkDistance-32    40,656,997 ops     29.74 ns/op
BenchmarkArea-32         1,000,000 ops   1,054 ns/op
BenchmarkContains-32     1,000,000 ops   1,042 ns/op
BenchmarkCovering-32        55,890 ops  21,444 ns/op
```

### Performance Considerations

1. **Envelope Checking**: While the predicate functions don't explicitly check envelopes first (this is done in `Intersects` and `Contains`), consider checking bounding boxes first for large datasets.

2. **Spatial Indexing**: For queries involving many geometries, use the spatial indexing functions in `spherical/index.go`.

3. **Caching**: The S2 conversion functions create new S2 objects each time. For repeated operations on the same geometries, consider caching the S2 representations.

4. **Tolerance**: The 0.1-meter tolerance is reasonable for most applications. To adjust it, modify the `defaultToleranceMeters` constant in the source.

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
- `github.com/robert-malhotra/go-topology-suite/geom` - GTS geometry types

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

## References

- [S2 Geometry Library](https://s2geometry.io/)
- [OGC Simple Features Specification](https://www.ogc.org/standards/sfa)
- [JTS Topology Suite](https://github.com/locationtech/jts)
- [DE-9IM Model](https://en.wikipedia.org/wiki/DE-9IM)

## Contributing

Contributions are welcome! Please ensure:
1. All tests pass
2. Code follows Go conventions
3. New features include tests and examples
4. Documentation is updated

## License

Part of the Go Topology Suite project.
