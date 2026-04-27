# Release Checklist

Use this checklist before tagging or publishing a release. The GitHub Actions workflow installs CI-only tools where needed, so local release preparation does not require every tool to be present on a developer machine.

## Required CI Gates

- Root module tests: `go test ./...`
- Root module race tests: `go test -race ./...`
- Root module vet: `go vet ./...`
- Root module build: `go build ./...`
- Root module tidiness check: `go mod tidy` followed by a clean `go.mod`/`go.sum` diff
- `v2` module tests: `cd v2 && go test ./...`
- `v2` module race tests: `cd v2 && go test -race ./...`
- `v2` module vet: `cd v2 && go vet ./...`
- `v2` module build: `cd v2 && go build ./...`
- `v2` module tidiness check: `cd v2 && go mod tidy` followed by a clean `go.mod`/`go.sum` diff
- High-risk package coverage gates using Go coverage profiles:
  - `geom` >= 60.0%
  - `operation/overlay` >= 70.0%
  - `operation/buffer` >= 80.0%
  - `operation/relate` >= 65.0%
  - `io/wkt` >= 70.0%
  - `io/wkb` >= 70.0%
  - `io/geojson` >= 75.0%
  - `io/kml` >= 75.0%
  - `io/shapefile` >= 75.0%
- Bounded fuzz smoke hooks for selected fuzz targets
- Benchmark smoke hooks for point-in-polygon, noding, overlay, buffer, WKT, WKB, and STRtree query baselines
- `golangci-lint` through `golangci/golangci-lint-action@v7` with the pinned version in CI when `.golangci.yml` is present

## Regression Focus

Before a production-oriented release, confirm the regression suite still exercises the hardening items that define the current implementation state:

- Active relate golden fixtures, including polygonal-set and mixed-collection cases.
- External parity fixture review for overlay, relate, parser, and buffer groups.
- WKB/EWKB nested SRID and Z/M dimensional metadata cases.
- Shared topology primitive behavior used by overlay/relate work.
- Negative buffer collapse cases that should return empty geometry.
- Parser and v2 operation errors for malformed or invalid inputs.

## Documentation and Example Gates

- Package examples compile through the normal test gates:
  - Root examples: `go test ./...`
  - v2 examples: `cd v2 && go test ./...`
- `example_test.go` files are the compile-checked documentation source of truth. README snippets should be reviewed against those examples when API names or signatures change.
- Documentation must state whether an operation is planar, spherical, or geodetic. Do not describe longitude/latitude planar overlay, buffer, area, or distance as production-safe without an explicit projection step or a documented distortion tolerance.
- v2 docs must call out any compatibility escape hatches, including `AllowInvalid`, `AllowInvalidInput`, and `AllowInvalidInputs`.
- Release notes must distinguish production hardening from full algorithmic parity claims. Avoid implying complete JTS/GEOS equivalence for degenerate or precision-sensitive cases unless backed by fixtures and tests.
- Release notes must keep the remaining core gap explicit: completing full mixed-dimension collection DE-9IM coverage and expanding external JTS/GEOS parity fixtures.

## Optional Local Preflight

Run the commands that are practical in your local environment:

```bash
go test ./...
go test -race ./...
go vet ./...
go test -cover ./geom ./operation/overlay ./operation/buffer ./operation/relate ./io/wkt ./io/wkb ./io/geojson ./io/kml ./io/shapefile
go test -run '^$' -fuzz '^FuzzBufferPoint$' -fuzztime=10s ./operation/buffer
go test -run '^$' -fuzz '^FuzzIntersectionPolygons$' -fuzztime=10s ./operation/overlay
go test -run '^$' -bench '^BenchmarkNodeLineSetsGrid$' -benchtime=100ms -count=1 ./internal/topology
(cd v2 && go test ./...)
(cd v2 && go test -race ./...)
(cd v2 && go vet ./...)
```

If `golangci-lint` is installed locally, run:

```bash
golangci-lint run --timeout=5m
```

If it is not installed, rely on the CI action rather than adding a local tooling requirement.

To install the same major lint tool locally, use the upstream installer or package manager for `golangci-lint` v2, then run the command above. CI remains authoritative for lint availability.

## Release Notes

- Confirm CI is green on the release branch.
- Confirm the release tag targets the intended commit.
- Confirm `CHANGELOG.md` has an entry for the release when user-facing changes are included.
- Confirm module paths and versions are correct for both root and `v2`.
- Confirm [production-readiness.md](production-readiness.md) still reflects known correctness limitations, CRS boundaries, and production gates.
