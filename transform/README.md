# Transform Package

The `transform` package provides coordinate transformation functionality for the Go Topology Suite. It includes support for affine transformations, map projections, and utilities for transforming entire geometries.

## Features

### Affine Transformations

Affine transformations support common 2D geometric operations:

- **Translation**: Shift coordinates by a fixed offset
- **Scaling**: Scale coordinates by factors in x and y directions
- **Rotation**: Rotate coordinates around a point
- **Shearing**: Apply shear transformations
- **Composition**: Combine multiple transformations

```go
// Create a composite transformation
translate := transform.NewAffineTranslation(10, 20)
scale := transform.NewAffineScale(2, 2)
rotate := transform.NewAffineRotation(math.Pi / 4)
composite := transform.NewComposite(translate, scale, rotate)

// Transform a point
x, y, err := composite.Forward(5, 5)
```

### Map Projections

The `projection` subpackage provides map projection transformations:

#### Web Mercator (EPSG:3857)
The standard projection used by web mapping applications:

```go
wm := projection.WebMercator()
x, y, err := wm.Forward(lon, lat) // degrees -> meters
lon, lat, err := wm.Inverse(x, y) // meters -> degrees
```

#### UTM (Universal Transverse Mercator)
Accurate projection for specific geographic zones:

```go
// UTM Zone 10N (covers San Francisco area)
utm := projection.UTM(10, true, nil)
easting, northing, err := utm.Forward(lon, lat)
lon, lat, err := utm.Inverse(easting, northing)
```

#### Custom Projections
Create custom Mercator or Transverse Mercator projections:

```go
// Custom Mercator with specific ellipsoid and parameters
merc := projection.NewMercator(
    projection.WGS84(),
    centralMeridian,
    falseEasting,
    falseNorthing,
)

// Custom Transverse Mercator
tm := projection.NewTransverseMercator(
    projection.GRS80(),
    centralMeridian,
    latitudeOfOrigin,
    scaleFactor,
    falseEasting,
    falseNorthing,
)
```

### Geometry Transformations

Transform entire geometries while preserving their structure:

```go
// Transform any geometry type
polygon := geom.NewPolygon(shell, holes)
transformed, err := transform.TransformGeometry(affineTransform, polygon)

// Works with all geometry types:
// - Point, MultiPoint
// - LineString, MultiLineString
// - Polygon, MultiPolygon
// - GeometryCollection (recursive)
```

## Transform Interface

All transformations implement the `Transform` interface:

```go
type Transform interface {
    Forward(x, y float64) (float64, float64, error)
    Inverse(x, y float64) (float64, float64, error)
}
```

This allows custom transformations to be used with the geometry transformation utilities.

## Supported Ellipsoids

- **WGS84**: World Geodetic System 1984 (GPS standard)
- **GRS80**: Geodetic Reference System 1980
- **Clarke1866**: Used in NAD27
- **Sphere**: Spherical model with custom radius

```go
ellipsoid := projection.WGS84()
ellipsoid := projection.GRS80()
ellipsoid := projection.Clarke1866()
ellipsoid := projection.Sphere(6371000) // Custom radius in meters
```

## Utilities

### Coordinate Transformation
```go
// Transform single coordinate
coord := geom.NewCoordinate(x, y)
transformed, err := transform.TransformCoordinate(t, coord)

// Transform coordinate sequence
coords := geom.CoordinateSequence{...}
transformed, err := transform.TransformCoordinates(t, coords)
```

### Composite Transformations
```go
// Chain multiple transformations
composite := transform.NewComposite(
    transform1,
    transform2,
    transform3,
)
```

### Inverse Transformations
```go
// Wrap a transform to swap forward/inverse
inverse := transform.NewInverse(originalTransform)
```

## Implementation Notes

### Precision
- Affine transformations maintain high precision (double-precision floating point)
- Map projections may introduce small errors due to iterative algorithms
- Round-trip transformations (forward then inverse) typically accurate to ~1e-9 degrees

### Performance
- Affine transformations: ~0.1 ns/op
- Web Mercator: ~15 ns/op (forward), ~13 ns/op (inverse)
- UTM: ~47 ns/op (forward), ~99 ns/op (inverse)
- Geometry transformations scale with coordinate count

### Thread Safety
All transformation objects are immutable and safe for concurrent use.

## Testing

The package includes comprehensive tests:

```bash
# Run all tests
go test ./transform/...

# Run with coverage
go test ./transform/... -cover

# Run benchmarks
go test ./transform/... -bench=.
```

Coverage: ~80% of statements

## Future Enhancements

Potential additions:
- Additional projection types (Lambert Conformal Conic, Albers Equal Area)
- Datum transformations (coordinate system conversions)
- 3D transformations
- Improved ellipsoidal Mercator inverse formula
- PROJ string parsing for projection definitions

## References

- [OGC Simple Features Specification](https://www.ogc.org/standards/sfa)
- [EPSG Geodetic Parameter Dataset](https://epsg.org/)
- [Map Projections - A Working Manual (Snyder, 1987)](https://pubs.usgs.gov/pp/1395/report.pdf)
- [Coordinate Conversions and Transformations (EPSG Guidance Note 7)](https://www.iogp.org/bookstore/)
