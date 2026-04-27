# Production Readiness

This project has a production-oriented v2 API, but production use still depends on choosing the correct coordinate model, validation policy, and release gates for the workload. This document is the checklist for what GTS does and does not guarantee today.

## Recommended API Surface

Use `github.com/robert-malhotra/go-topology-suite/v2` for new production code. It provides:

- Constructors that return errors for invalid or nil inputs.
- Overlay and buffer functions that return `(geom.Geometry, error)`.
- Strict WKT/WKB readers, including EWKB SRID/Z/M handling for nested collections, and v2 wrappers for GeoJSON and KML.
- Optional normalization and precision-model application at operation boundaries.

The root module remains supported for compatibility. It should be treated as the legacy API surface when fail-fast behavior is required.

## Current Implemented State

Recent hardening that production users can rely on when covered by their workload fixtures:

- Relate golden fixtures are active and are part of the regression suite, including graph-backed line/polygon, polygonal-set, and mixed-collection cases.
- Polygon overlay uses selected labeled faces for polygonal output.
- WKB/EWKB nested SRID and Z/M dimensional metadata coverage is in place.
- Shared topology primitives exist behind overlay/relate work, including labeled polygon boundary edges and an initial labeled-face tracing bridge.
- Buffer negative-distance collapse behavior is fixed for cases that should return empty geometry instead of invalid remnants.
- v2 parser and operation entry points consistently prefer explicit errors over silent best-effort behavior.

## Correctness Boundaries

GTS is a pure-Go topology library with ongoing hardening. Do not treat v2 as a blanket guarantee of full JTS/GEOS parity for every pathological geometry.

Known boundaries to account for in production:

- The remaining core topology gap is completing full mixed-dimension collection DE-9IM coverage and broadening external JTS/GEOS parity fixture coverage.
- Highly degenerate inputs, near-coincident linework, extreme coordinate magnitudes, tiny slivers, and invalid rings can still expose numeric or topology edge cases.
- Validation catches many invalid geometries, but it is not a replacement for dataset-specific QA, snapping, simplification, deduplication, or repair workflows.
- Precision models can make coordinates deterministic before an operation, but inappropriate snapping can also change topology.
- Buffer and overlay should be tested with representative fixtures from the target domain before relying on exact output shape, component ordering, or boundary details.
- Output normalization is useful for reproducible tests and serialization, not for proving semantic correctness.

## CRS and Geographic Boundaries

Planar operations use the numeric coordinate plane supplied by the caller. They do not know whether values are meters, feet, degrees, pixels, or any other unit.

Production rules:

- Use planar overlay, buffer, area, and distance only with an appropriate projected CRS or another intentional planar coordinate system.
- Do not pass longitude/latitude directly to planar buffer or distance and interpret the result as meters.
- Do not mix CRSs in a single operation. GTS does not reproject, infer axis order, convert units, or verify SRID compatibility.
- Treat SRID as metadata carried by factories and geometries where supported.
- Use `geodetic` for WGS84 distance, bearing, destination, and geodesic area.
- Use `spherical` for WGS84-style predicates backed by S2.
- Project geographic data before planar overlay or buffering when the analysis requires planar topology.

Web Mercator is useful for visualization and some coarse workflows, but it is not a universal analysis projection. Choose a local or domain-appropriate CRS when measurements or buffers need defensible units.

## v2 Migration Gates

Before migrating a production workflow to v2:

- Replace root-module constructors with v2 constructors where invalid input should fail early.
- Handle returned errors from constructors, I/O, overlay, and buffer.
- Audit uses of `AllowInvalid`, `AllowInvalidInput`, and `AllowInvalidInputs`; keep them only for controlled compatibility, import quarantine, diagnostics, or repair.
- Decide whether `Normalize` or `NormalizeResult` is needed for deterministic tests or serialized outputs.
- Decide whether a `PrecisionModel` is appropriate for the dataset and document the tolerance.
- Add fixtures for representative valid, invalid, boundary-touching, near-coincident, empty, and large-coordinate cases.

## I/O Expectations

Strict v2 readers are intended to fail malformed input instead of accepting ambiguous payloads.

Production callers remain responsible for:

- Business-level schema validation.
- GeoJSON property validation.
- Shapefile sidecar completeness and attribute handling.
- CRS sidecar parsing and CRS compatibility checks.
- Enforcing application-specific dimensionality, bounds, and coordinate ranges.

## Example Compile Expectations

Compile-checked examples live in `example_test.go` files. These examples must pass as part of the normal test suites:

```bash
go test ./...
(cd v2 && go test ./...)
```

README examples are illustrative and should be kept aligned with compile-checked examples. New documented workflows should include an `Example...` test when practical.

## Release Gates

A production-ready release should pass the gates in [release-checklist.md](release-checklist.md), including:

- Root and v2 tests.
- Root and v2 race tests.
- Root and v2 vet.
- Root and v2 builds.
- Module tidiness checks.
- Coverage gates for high-risk packages.
- Bounded fuzz smoke for selected parser and topology targets.
- Benchmark smoke for representative geometry operations.
- Lint through CI when `.golangci.yml` is present; run `golangci-lint` locally when the binary is installed.

Passing these gates means the release met the current project quality bar. It does not eliminate the need for workload-specific fixtures, CRS review, and operational monitoring in downstream applications.
