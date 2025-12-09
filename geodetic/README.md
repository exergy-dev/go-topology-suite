# Geodetic Calculations Package

The `geodetic` package provides accurate geodetic calculations on ellipsoidal Earth models. It implements geodesic distance, azimuth, and area calculations using various reference ellipsoids including WGS84, GRS80, and Clarke1866.

## Features

- **Geodesic Distance**: Vincenty's formula for sub-millimeter accuracy
- **Fast Approximations**: Haversine formula for spherical calculations
- **Azimuth/Bearing**: Initial and final bearings along geodesics
- **Direct Problem**: Find destination given start point, bearing, and distance
- **Inverse Problem**: Find distance and bearings given two points
- **Polygon Area**: Geodetic area calculation on ellipsoids
- **Multiple Ellipsoids**: Support for WGS84, GRS80, Clarke1866, and custom ellipsoids

## Installation

```bash
go get github.com/go-topology-suite/gts/geodetic
```

## Quick Start

```go
import "github.com/go-topology-suite/gts/geodetic"

// Calculate distance between New York and London
lat1, lon1 := 40.7128, -74.0060  // NYC
lat2, lon2 := 51.5074, -0.1278   // London

distance := geodetic.DistanceWGS84(lat1, lon1, lat2, lon2)
fmt.Printf("Distance: %.0f km\n", distance/1000)
// Output: Distance: 5585 km

// Calculate initial bearing
bearing := geodetic.InitialBearing(lat1, lon1, lat2, lon2)
fmt.Printf("Bearing: %.1f°\n", bearing)
// Output: Bearing: 51.2°
```

## Coordinate Convention

All latitude and longitude values are in **degrees** (not radians):
- Latitude: -90 (South Pole) to +90 (North Pole)
- Longitude: -180 to +180 (or 0 to 360)
- Azimuth/Bearing: 0 to 360 (0 = North, 90 = East, 180 = South, 270 = West)

## Functions

### Distance Calculations

#### `Distance(lat1, lon1, lat2, lon2 float64, e *Ellipsoid) float64`
Calculates geodesic distance using Vincenty's inverse formula. Accurate to ~0.5mm.

```go
distance := geodetic.Distance(lat1, lon1, lat2, lon2, geodetic.WGS84)
```

#### `DistanceWGS84(lat1, lon1, lat2, lon2 float64) float64`
Convenience function using WGS84 ellipsoid.

#### `Vincenty(lat1, lon1, lat2, lon2 float64, e *Ellipsoid) (float64, error)`
Vincenty's formula with error handling for non-convergent cases (rare).

#### `Haversine(lat1, lon1, lat2, lon2, radius float64) float64`
Spherical approximation. Faster but less accurate (~0.5% error).

```go
// Use mean Earth radius for spherical calculation
distance := geodetic.Haversine(lat1, lon1, lat2, lon2, geodetic.EarthMeanRadius)
```

### Azimuth/Bearing

#### `InitialBearing(lat1, lon1, lat2, lon2 float64) float64`
Returns initial bearing (0-360°) from point 1 to point 2.

#### `FinalBearing(lat1, lon1, lat2, lon2 float64) float64`
Returns final bearing when arriving at point 2 from point 1.

#### `Inverse(lat1, lon1, lat2, lon2 float64, e *Ellipsoid) (distance, azimuth1, azimuth2 float64, err error)`
Solves inverse problem: returns distance and both azimuths.

```go
distance, az1, az2, err := geodetic.Inverse(lat1, lon1, lat2, lon2, geodetic.WGS84)
```

### Destination/Direct Problem

#### `DestinationPoint(lat, lon, bearing, distance float64, e *Ellipsoid) (lat2, lon2 float64)`
Calculates destination point given start point, bearing, and distance.

```go
lat2, lon2 := geodetic.DestinationPoint(lat, lon, bearing, distance, geodetic.WGS84)
```

#### `Direct(lat1, lon1, azimuth1, distance float64, e *Ellipsoid) (lat2, lon2, azimuth2 float64, err error)`
Solves direct problem: returns destination and final azimuth.

### Area Calculations

#### `PolygonArea(lats, lons []float64, e *Ellipsoid) float64`
Calculates geodetic area of a polygon in square meters.

```go
lats := []float64{0, 0, 1, 1, 0}
lons := []float64{0, 1, 1, 0, 0}
area := geodetic.PolygonArea(lats, lons, geodetic.WGS84)
```

#### `SphericalPolygonArea(lats, lons []float64, radius float64) float64`
Faster spherical approximation for area calculation.

#### `SignedPolygonArea(lats, lons []float64, e *Ellipsoid) float64`
Returns signed area (positive for CCW, negative for CW).

## Ellipsoids

### Pre-defined Ellipsoids

- **WGS84**: World Geodetic System 1984 (most common, used by GPS)
- **GRS80**: Geodetic Reference System 1980 (used by NAD83)
- **Clarke1866**: Historical US ellipsoid (used by NAD27)
- **Sphere**: Spherical Earth model (radius = 6,371,000 m)

### Constants

- **EarthMeanRadius**: 6,371,008.8 meters
- **EarthAuthalicRadius**: 6,371,007.2 meters (equal-area sphere)

### Creating Custom Ellipsoids

```go
// From semi-major and semi-minor axes
custom := geodetic.NewEllipsoid("Custom", 6378137.0, 6356752.0)

// From semi-major axis and flattening
custom := geodetic.NewEllipsoidFromAF("Custom", 6378137.0, 1.0/298.257223563)

// From semi-major axis and inverse flattening
custom := geodetic.NewEllipsoidFromAInvF("Custom", 6378137.0, 298.257223563)
```

### Ellipsoid Properties

```go
e := geodetic.WGS84

e.SemiMajorAxis()           // 6378137.0 m
e.SemiMinorAxis()           // 6356752.314... m
e.Flattening()              // ~0.00335...
e.InverseFlattening()       // 298.257223563
e.EccentricitySquared()     // ~0.00669...
e.Eccentricity()            // ~0.0818...
e.SecondEccentricitySquared() // ~0.00673...
```

## Accuracy and Performance

### Vincenty's Formula
- **Accuracy**: ~0.5mm on Earth ellipsoid
- **Performance**: ~215 ns/op (very fast)
- **Convergence**: May fail for nearly antipodal points (extremely rare)

### Haversine Formula
- **Accuracy**: ~0.5% error due to Earth's ellipsoidal shape
- **Performance**: ~37 ns/op (6x faster than Vincenty)
- **Use case**: When speed is critical and sub-kilometer accuracy is acceptable

### Benchmarks (on Intel i9-14900KF)

```
BenchmarkVincenty-32         5,544,523 ops    214.9 ns/op    0 allocs
BenchmarkHaversine-32       32,026,393 ops     37.2 ns/op    0 allocs
BenchmarkDirect-32           8,615,131 ops    140.1 ns/op    0 allocs
BenchmarkPolygonArea-32      4,619,252 ops    259.1 ns/op    0 allocs
```

All functions are allocation-free for optimal performance.

## Examples

### Calculate Distance

```go
// New York to London
distance := geodetic.DistanceWGS84(40.7128, -74.0060, 51.5074, -0.1278)
fmt.Printf("Distance: %.0f km\n", distance/1000)
// Output: Distance: 5585 km
```

### Find Destination Point

```go
// Start at Sydney, travel 1000km northeast
lat, lon := -33.8688, 151.2093
lat2, lon2 := geodetic.DestinationPoint(lat, lon, 45.0, 1000000.0, geodetic.WGS84)
fmt.Printf("Destination: %.4f°, %.4f°\n", lat2, lon2)
```

### Calculate Polygon Area

```go
// Small square near equator
lats := []float64{0, 0, 0.1, 0.1, 0}
lons := []float64{0, 0.1, 0.1, 0, 0}
area := geodetic.PolygonAreaWGS84(lats, lons)
fmt.Printf("Area: %.0f km²\n", area/1e6)
// Output: Area: 123 km²
```

### Round Trip (Direct + Inverse)

```go
// Forward: calculate destination
lat1, lon1 := 35.0, 45.0
lat2, lon2, _, _ := geodetic.Direct(lat1, lon1, 60.0, 500000.0, geodetic.WGS84)

// Reverse: calculate back
dist, az, _, _ := geodetic.Inverse(lat1, lon1, lat2, lon2, geodetic.WGS84)

fmt.Printf("Distance: %.0f m\n", dist)      // 500000 m
fmt.Printf("Azimuth: %.1f°\n", az)          // 60.0°
```

## Algorithm References

### Vincenty's Formula
Vincenty, T. (1975). "Direct and Inverse Solutions of Geodesics on the Ellipsoid with application of nested equations". Survey Review, Vol. 23, No. 176, pp. 88-93.

### Haversine Formula
Spherical law of haversines for great circle distance calculation.

### Polygon Area
Authalic latitude correction with spherical excess formula for accurate area calculation on ellipsoids.

## Edge Cases

### Antipodal Points
Points on opposite sides of the Earth may cause Vincenty's formula to fail to converge. The package automatically falls back to spherical approximation in these rare cases.

### International Date Line
The package correctly handles longitude discontinuities at ±180°.

### Poles
Calculations near the poles are handled correctly, though bearing becomes undefined exactly at the poles.

## Thread Safety

All functions are pure (no shared state) and safe for concurrent use.

## Testing

The package includes comprehensive tests:
- Known distance test cases (e.g., Flinders Peak to Buninyong)
- Symmetry tests (Distance(A,B) == Distance(B,A))
- Round-trip tests (Direct then Inverse)
- Edge case tests (same point, poles, date line)

Run tests:
```bash
go test ./geodetic
go test -bench=. ./geodetic
```

## See Also

- [JTS Topology Suite](https://github.com/locationtech/jts) - Java geometric library
- [PROJ](https://proj.org/) - Cartographic projections library
- [GeographicLib](https://geographiclib.sourceforge.io/) - C++ geodesic library

## License

Part of the Go Topology Suite project.
