# Spherical Extended Predicates

This document describes the generic spatial predicates implemented for spherical geometry in the `spherical` package.

## Overview

The extended predicates provide full support for spatial relationship testing between any combination of geometry types using spherical geometry calculations (via the S2 library). These predicates work correctly across the antimeridian and near the poles.

## Functions

### GenericWithin

```go
func GenericWithin(g1, g2 geom.Geometry) bool
```

Returns `true` if geometry `g1` is completely within geometry `g2`.

**Properties:**
- This is the inverse of `Contains`: `GenericWithin(a, b) == Contains(b, a)`
- All points of g1 must be in the interior or on the boundary of g2
- At least one point of g1 must be in the interior of g2

**Example:**
```go
point := geom.NewPoint(-73.985, 40.748) // Empire State Building
manhattan := geom.NewPolygon(...) // Manhattan polygon
isWithin := spherical.GenericWithin(point, manhattan) // true
```

### GenericDisjoint

```go
func GenericDisjoint(g1, g2 geom.Geometry) bool
```

Returns `true` if geometries `g1` and `g2` have no points in common.

**Properties:**
- This is the inverse of `Intersects`: `GenericDisjoint(a, b) == !Intersects(a, b)`
- Two geometries are disjoint if they share no boundary or interior points

**Example:**
```go
nyc := geom.NewPolygon(...) // New York City
london := geom.NewPolygon(...) // London
areDisjoint := spherical.GenericDisjoint(nyc, london) // true
```

### GenericOverlaps

```go
func GenericOverlaps(g1, g2 geom.Geometry) bool
```

Returns `true` if geometries `g1` and `g2` overlap.

**Properties:**
- Both geometries must have the same dimension (both polygons, both lines, etc.)
- The geometries must intersect
- Neither geometry can completely contain the other
- Typically used for polygons that partially overlap

**Example:**
```go
zone1 := geom.NewPolygon(...) // Delivery zone 1
zone2 := geom.NewPolygon(...) // Delivery zone 2
doOverlap := spherical.GenericOverlaps(zone1, zone2) // true if zones overlap
```

### GenericTouches

```go
func GenericTouches(g1, g2 geom.Geometry) bool
```

Returns `true` if geometries `g1` and `g2` touch at their boundaries only.

**Properties:**
- Geometries must share boundary points
- Geometries must NOT share interior points
- Common for adjacent polygons (like countries or states)

**Example:**
```go
stateA := geom.NewPolygon(...) // New York
stateB := geom.NewPolygon(...) // New Jersey
doTouch := spherical.GenericTouches(stateA, stateB) // true (share border)
```

## Supported Geometry Types

All functions support the following geometry types:
- Point
- LineString
- LinearRing
- Polygon
- MultiPoint
- MultiLineString
- MultiPolygon
- GeometryCollection

## Implementation Details

### Spherical Geometry

All predicates use the S2 library for spherical geometry calculations, which means:
- Calculations are accurate on the Earth's surface
- Correctly handles the antimeridian (180°/-180° longitude line)
- Works properly near the poles
- Distances and areas are calculated on the sphere, not a flat plane

### Point Location

The implementation includes a comprehensive `locatePointSpherical` function that determines whether a point is:
- **Interior**: Inside the geometry
- **Boundary**: On the edge/boundary of the geometry
- **Exterior**: Outside the geometry

This is crucial for implementing the `Touches` predicate correctly.

### Tolerance

The implementation uses a 1-meter tolerance for coincidence tests. Two points are considered the same if they are within 1 meter of each other on the Earth's surface.

### Interior Intersection Detection

For polygon-polygon relationships, the implementation includes `hasSphericalPolygonInteriorIntersection` which detects when two polygons have overlapping interior areas (not just touching boundaries). This is essential for correctly implementing the `Touches` predicate.

## Comparison with Planar Predicates

The planar predicates in `geom/predicates.go` perform similar operations but use planar (Euclidean) geometry. Key differences:

| Aspect | Planar | Spherical |
|--------|--------|-----------|
| Distance calculation | Euclidean distance | Great circle distance |
| Area calculation | Flat plane | Spherical surface |
| Accuracy | Good for small areas | Accurate anywhere on Earth |
| Antimeridian | Can cause issues | Handled correctly |
| Poles | Can cause issues | Handled correctly |
| Performance | Slightly faster | Slightly slower |

## Usage Guidelines

### When to Use Spherical Predicates

Use spherical predicates when:
- Working with real-world geographic coordinates (latitude/longitude)
- Dealing with large areas where Earth's curvature matters
- Data crosses the antimeridian or includes polar regions
- Accuracy on the Earth's surface is important

### When to Use Planar Predicates

Use planar predicates when:
- Working with projected coordinates (e.g., UTM, State Plane)
- Performance is critical and areas are small
- Data is already in a planar coordinate system
- Earth's curvature can be safely ignored

## Examples

### Example 1: Point in Polygon

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

if spherical.GenericWithin(point, manhattan) {
    fmt.Println("Point is in Manhattan")
}
```

### Example 2: Adjacent Polygons

```go
// Check if two countries share a border
usa := geom.NewPolygon(...) // USA boundary
canada := geom.NewPolygon(...) // Canada boundary

if spherical.GenericTouches(usa, canada) {
    fmt.Println("USA and Canada share a border")
}

if spherical.GenericDisjoint(usa, canada) {
    fmt.Println("USA and Canada don't touch")
}
```

### Example 3: Overlapping Regions

```go
// Check if two delivery zones overlap
zone1 := geom.NewPolygon(...)
zone2 := geom.NewPolygon(...)

if spherical.GenericOverlaps(zone1, zone2) {
    fmt.Println("Zones overlap - need to resolve territory")
}
```

### Example 4: Multi-Geometry Support

```go
// Check if multiple points are all within a region
points := geom.NewMultiPoint([]*geom.Point{
    geom.NewPoint(-73.99, 40.75),
    geom.NewPoint(-73.98, 40.76),
    geom.NewPoint(-73.97, 40.74),
})

region := geom.NewPolygon(...)

if spherical.GenericWithin(points, region) {
    fmt.Println("All points are within the region")
}
```

## Performance Considerations

1. **Envelope Checking**: While the generic functions don't explicitly check envelopes first (this is done in `Intersects` and `Contains`), consider checking bounding boxes first for large datasets.

2. **Spatial Indexing**: For queries involving many geometries, use the spatial indexing functions in `spherical/index.go`.

3. **Caching**: The S2 conversion functions create new S2 objects each time. For repeated operations on the same geometries, consider caching the S2 representations.

4. **Tolerance**: The 1-meter tolerance is reasonable for most applications but can be adjusted if needed for higher precision requirements.

## Testing

Comprehensive tests are provided in:
- `predicates_extended_test.go`: Unit tests for all functions
- `predicates_extended_example_test.go`: Executable examples

Run tests with:
```bash
go test ./spherical -v -run "TestGeneric"
```

Run examples with:
```bash
go test ./spherical -v -run "Example"
```

## Future Enhancements

Potential improvements:
1. Add configurable tolerance parameter
2. Optimize performance by caching S2 conversions
3. Add more specialized predicates (e.g., `Crosses`, `Covers`)
4. Support for custom distance metrics
5. Parallel processing for MultiGeometry types

## References

- [S2 Geometry Library](https://s2geometry.io/)
- [OGC Simple Features Specification](https://www.ogc.org/standards/sfa)
- [JTS Topology Suite](https://github.com/locationtech/jts)
- [DE-9IM Model](https://en.wikipedia.org/wiki/DE-9IM)
