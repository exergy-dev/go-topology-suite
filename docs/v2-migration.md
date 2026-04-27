# v2 Migration Notes

The `github.com/robert-malhotra/go-topology-suite/v2` module is the production-oriented API surface for workflows that need explicit errors and stricter input handling. It remains a compatibility facade over the pure-Go implementation in places, but shared topology primitives now exist behind the overlay/relate hardening work and are being expanded where they reduce duplicate graph and labeling behavior.

Use v2 when new code needs validation at construction boundaries, strict parser failures instead of best-effort reads, and operation calls that report invalid input rather than silently returning an unusable result. Keep the root module when preserving existing behavior is more important than fail-fast semantics.

## Module Path

```bash
go get github.com/robert-malhotra/go-topology-suite/v2
```

```go
import topology "github.com/robert-malhotra/go-topology-suite/v2"
```

## Constructor Changes

The root module keeps simple constructors that mirror the original API. The v2 module adds validated constructors that return errors and accept `ConstructorOptions`:

```go
polygon, err := topology.NewPolygon(topology.CoordinateSequence{
	topology.NewCoordinate(0, 0),
	topology.NewCoordinate(10, 0),
	topology.NewCoordinate(10, 10),
	topology.NewCoordinate(0, 10),
	topology.NewCoordinate(0, 0),
}, nil)
```

By default, v2 constructors reject nil coordinate slices, nil child geometries, NaN coordinates where validation catches them, and geometries whose `IsValid` result is false. Use `ConstructorOptions{AllowInvalid: true}` only when intentionally carrying invalid geometries for diagnostics, import quarantine, or repair. `ConstructorOptions{Normalize: true}` normalizes successful outputs when deterministic ordering is useful for tests or serialization.

To attach an SRID, pass a factory:

```go
factory := topology.NewGeometryFactory(3857)
point, err := topology.NewPoint(0, 0, topology.ConstructorOptions{Factory: factory})
```

SRID is metadata only. It does not trigger reprojection or unit conversion.

## Operation Changes

Overlay and buffer operations return `(geom.Geometry, error)` in v2:

```go
result, err := topology.Intersection(a, b)
buffered, err := topology.Buffer(g, 10)
```

By default, v2 rejects nil and invalid inputs before running operations. Use the `AllowInvalidInputs` or `AllowInvalidInput` options only for compatibility with legacy behavior or for controlled repair pipelines. `NormalizeResult` can be used when stable output ordering matters. `PrecisionModel` can snap cloned inputs before the operation; it is not a replacement for choosing a suitable CRS or for data-specific topology QA.

Custom buffer parameters are validated before the operation runs. `QuadrantSegments` must be positive, cap and join styles must be supported values from `operation/buffer`, and mitre joins require a finite positive `MitreLimit`.

Relate pattern helpers are strict in v2. `RelatePattern` returns an error unless the DE-9IM pattern is exactly nine characters and each character is one of `T`, `F`, `*`, `0`, `1`, or `2`.

## I/O Changes

Use `ReadWKT`, `WriteWKT`, `ReadWKB`, and `WriteWKB` for v2 I/O. WKT and WKB parsing is strict: trailing payloads, malformed collection syntax, unsupported dimension flags, and inconsistent payloads are errors. WKB/EWKB coverage includes nested SRID and Z/M dimensional metadata cases. Use `ReadGeoJSON`, `WriteGeoJSON`, `ReadKML`, and `WriteKML` when callers need a single v2 entry point for those formats.

`ReadOptions{Factory: ...}` controls the output factory and SRID metadata where the underlying format reader supports it. GeoJSON and KML default to WGS84/SRID 4326 when no factory is supplied.

## CRS Boundary

Planar topology operations assume coordinates are already in an appropriate projected planar CRS. For longitude/latitude coordinates, use `spherical` or `geodetic` APIs for geographic predicates and measurements, or project coordinates before planar overlay/buffer work.

Do not mix coordinate systems in one operation. v2 does not compare SRIDs, reproject coordinates, infer axis order, or convert between degrees and meters. The caller is responsible for selecting a projection whose distortion is acceptable for the operation and geographic extent.

Practical guidance:

- Use `geodetic` for distances, bearings, destination points, and geodesic polygon area on WGS84.
- Use `spherical` for WGS84-style predicates backed by S2.
- Use planar overlay, buffer, area, and distance only after projection, or when the data is already in planar units.
- Treat Web Mercator as display-friendly, not a universal analysis CRS.

## Examples and Compile Expectations

The compile-checked examples live in `example_test.go` files. The v2 examples in `v2/example_test.go` demonstrate validated constructors, overlay, default and custom buffer options, relate pattern errors, WKT, GeoJSON, and KML. These examples must pass under:

```bash
go test ./...
(cd v2 && go test ./...)
```

README snippets should stay aligned with the current APIs, but they are not the primary compile gate. When adding a new documented workflow, prefer adding or updating an `Example...` test so the compiler and `go test` protect it.

## Implemented Hardening Since Early v2

The current tree includes several production-relevant hardening items that older migration notes may not have captured:

- Relate golden fixtures are active, including graph-backed line/polygon, polygonal-set, and mixed-collection cases.
- Polygon overlay now uses selected labeled faces for polygonal output.
- Root `OverlayWithPrecision` is additive. v2 precision behavior remains deterministic snapping before an operation; that improves reproducibility but can change topology.
- WKB/EWKB nested SRID and Z/M coverage is present.
- Shared topology primitives are available for ongoing overlay/relate convergence, including labeled polygon boundary edges and an initial labeled-face tracing bridge.
- Negative buffer collapse cases that should become empty geometry are fixed.
- Parser and operation wrappers continue to route malformed inputs to explicit errors.

## Remaining Correctness Limitations

v2 hardens input validation and error reporting, but it does not remove every algorithmic limitation inherited from the current pure-Go engine:

- The main remaining core gap is completing full mixed-dimension collection DE-9IM coverage and adding a larger external parity suite against JTS/GEOS fixtures.
- Robustness for highly degenerate overlay, near-coincident segments, extreme coordinate magnitudes, and precision-sensitive self-intersections still requires caller-side QA and representative fixtures.
- Buffer and overlay behavior should be validated against domain data before relying on exact equivalence with JTS or GEOS for difficult cases.
- CRS handling is metadata-oriented. No operation automatically reprojects or verifies that units and axis order are correct.
- Geographic overlay and geographic buffering are not provided by the v2 facade. Use spherical/geodetic APIs or an explicit projection step.
- I/O strictness catches malformed payloads, but schema-level business rules, attribute validation, shapefile sidecar completeness, and CRS sidecar parsing remain caller responsibilities.

## Compatibility Status

The v2 API is intended to preserve a migration path while breaking unsafe behavior intentionally. Current hardening includes validated constructors, strict WKT/WKB parsing, WKB/EWKB nested SRID/Z/M coverage, stricter GeoJSON/KML/Shapefile validation, relate golden fixtures, selected-face polygon overlay, mixed-collection relate corrections, bounded buffer dissolve, fixed negative buffer collapse, and release gates for tests, race, vet, fuzz smoke, lint, and benchmark smoke.

For production signoff criteria, see [production-readiness.md](production-readiness.md) and [release-checklist.md](release-checklist.md).
