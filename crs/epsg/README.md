# EPSG Registry Package

The `epsg` package provides EPSG (European Petroleum Survey Group) coordinate reference system definitions and a registry for looking up CRS by code.

## Overview

The EPSG dataset is the de facto standard registry of coordinate reference systems used in GIS and surveying. Each CRS is identified by a unique integer code (e.g., 4326 for WGS84).

This package includes the most commonly used CRS definitions:

**Geographic CRS:**
- WGS84 (EPSG:4326) - World Geodetic System 1984
- NAD83 (EPSG:4269) - North American Datum 1983
- NAD27 (EPSG:4267) - North American Datum 1927
- ETRS89 (EPSG:4258) - European Terrestrial Reference System 1989

**Projected CRS:**
- Web Mercator (EPSG:3857) - WGS 84 / Pseudo-Mercator
- UTM zones (EPSG:326xx for north, 327xx for south)

## Usage Examples

### Looking up a CRS by EPSG code

```go
package main

import (
    "fmt"
    "log"

    "github.com/robert-malhotra/go-topology-suite/crs/epsg"
)

func main() {
    // Look up WGS84 by code
    wgs84, err := epsg.Lookup(4326)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(wgs84.Name())         // "WGS 84"
    fmt.Println(wgs84.Code())         // "EPSG:4326"
    fmt.Println(wgs84.IsGeographic()) // true
}
```

### Using predefined CRS

```go
package main

import (
    "fmt"

    "github.com/robert-malhotra/go-topology-suite/crs/epsg"
)

func main() {
    // Use predefined constants
    fmt.Println(epsg.WGS84.Name())        // "WGS 84"
    fmt.Println(epsg.WebMercator.Name())  // "WGS 84 / Pseudo-Mercator"

    // Check CRS type
    if epsg.WGS84.IsGeographic() {
        fmt.Println("WGS84 uses lat/lon coordinates")
    }

    if !epsg.WebMercator.IsGeographic() {
        fmt.Println("Web Mercator uses projected coordinates")
    }
}
```

### Working with UTM zones

```go
package main

import (
    "fmt"

    "github.com/robert-malhotra/go-topology-suite/crs/epsg"
)

func main() {
    // Generate UTM zone CRS dynamically
    utm10n := epsg.UTMZone(10, true)   // Zone 10 North (US West Coast)
    fmt.Println(utm10n.Code())         // "EPSG:32610"
    fmt.Println(utm10n.Name())         // "WGS 84 / UTM zone 10N"

    utm50s := epsg.UTMZone(50, false)  // Zone 50 South (Australia)
    fmt.Println(utm50s.Code())         // "EPSG:32750"

    // Or use predefined common zones
    fmt.Println(epsg.UTM10N.Name())    // "WGS 84 / UTM zone 10N"
    fmt.Println(epsg.UTM17N.Name())    // "WGS 84 / UTM zone 17N"
    fmt.Println(epsg.UTM32N.Name())    // "WGS 84 / UTM zone 32N"
}
```

### Registering custom CRS

```go
package main

import (
    "fmt"
    "log"

    "github.com/robert-malhotra/go-topology-suite/crs"
    "github.com/robert-malhotra/go-topology-suite/crs/epsg"
)

func main() {
    // Create a custom CRS
    customCRS, err := crs.NewProjectedCRS(
        "EPSG:2154",
        "RGF93 / Lambert-93",
        crs.WGS84,
        crs.CartesianCS2D,
        "Lambert Conformal Conic",
        []float64{-9.86, 41.15, 10.38, 51.56}, // France bounds
    )
    if err != nil {
        log.Fatal(err)
    }

    // Register it
    if err := epsg.Register(customCRS); err != nil {
        log.Fatal(err)
    }

    // Now it can be looked up
    retrieved, err := epsg.Lookup(2154)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(retrieved.Name())  // "RGF93 / Lambert-93"
}
```

### Listing all registered CRS

```go
package main

import (
    "fmt"

    "github.com/robert-malhotra/go-topology-suite/crs/epsg"
)

func main() {
    // Get all registered EPSG codes
    codes := epsg.Codes()
    fmt.Printf("Registered CRS count: %d\n", epsg.Count())

    for _, code := range codes {
        crs, _ := epsg.Lookup(code)
        fmt.Printf("EPSG:%d - %s\n", code, crs.Name())
    }

    // Check if a specific code is registered
    if epsg.IsRegistered(4326) {
        fmt.Println("WGS84 is registered")
    }
}
```

### Working with CRS properties

```go
package main

import (
    "fmt"

    "github.com/robert-malhotra/go-topology-suite/crs/epsg"
)

func main() {
    wgs84 := epsg.WGS84

    // Get datum information
    datum := wgs84.Datum()
    fmt.Println(datum.Name())  // "WGS 84"

    // Get ellipsoid parameters
    ellipsoid := datum.Ellipsoid()
    fmt.Printf("Semi-major axis: %.1f m\n", ellipsoid.SemiMajorAxis())
    fmt.Printf("Inverse flattening: %.9f\n", ellipsoid.InverseFlattening())

    // Get coordinate system
    cs := wgs84.CoordinateSystem()
    fmt.Printf("Dimensions: %d\n", cs.Dimension())

    axis0 := cs.Axis(0)
    fmt.Printf("First axis: %s (%s)\n", axis0.Name, axis0.Unit.Name)

    // Get area of use
    minLon, minLat, maxLon, maxLat := wgs84.AreaOfUse()
    fmt.Printf("Area of use: [%.1f, %.1f] to [%.1f, %.1f]\n",
        minLon, minLat, maxLon, maxLat)

    // Get WKT representation
    fmt.Println("\nWKT:")
    fmt.Println(wgs84.WKT())
}
```

## API Reference

### Registry Functions

- `Lookup(code int) (crs.CRS, error)` - Find a CRS by EPSG code
- `MustLookup(code int) crs.CRS` - Find a CRS, panic if not found
- `Register(c crs.CRS) error` - Register a custom CRS
- `Unregister(code int) error` - Remove a CRS from the registry
- `Codes() []int` - Get sorted list of all registered EPSG codes
- `Count() int` - Get number of registered CRS
- `IsRegistered(code int) bool` - Check if a code is registered

### Geographic CRS Constants

- `WGS84` - EPSG:4326 - World Geodetic System 1984
- `NAD83` - EPSG:4269 - North American Datum 1983
- `NAD27` - EPSG:4267 - North American Datum 1927
- `ETRS89` - EPSG:4258 - European Terrestrial Reference System 1989

### Projected CRS Constants

- `WebMercator` - EPSG:3857 - WGS 84 / Pseudo-Mercator
- `UTM10N` - EPSG:32610 - WGS 84 / UTM zone 10N (US West Coast)
- `UTM17N` - EPSG:32617 - WGS 84 / UTM zone 17N (US East Coast)
- `UTM32N` - EPSG:32632 - WGS 84 / UTM zone 32N (Central Europe)

### Utility Functions

- `UTMZone(zone int, north bool) crs.CRS` - Generate UTM zone CRS dynamically

## Thread Safety

The EPSG registry is thread-safe. All registry operations (lookup, register, unregister) use appropriate locking to ensure safe concurrent access.

## Testing

Run the tests:

```bash
go test github.com/robert-malhotra/go-topology-suite/crs/epsg
```

Run with coverage:

```bash
go test -cover github.com/robert-malhotra/go-topology-suite/crs/epsg
```

## References

- [EPSG Registry](https://epsg.org/)
- [OGC Simple Features Specification](https://www.ogc.org/standards/sfa)
- [Proj4](https://proj.org/)
