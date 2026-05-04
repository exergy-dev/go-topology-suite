# Changelog

All notable changes to Terra will be documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **CRS subsystem.** `crs.Definition`, `crs.Datum`, `crs.Ellipsoid`, `crs.Projection`, `crs.OperationFor`, plus the projection set under `crs/proj/` (Web Mercator, Transverse Mercator, Lambert Conformal Conic 2SP, Albers Equal-Area Conic, Lambert Azimuthal Equal-Area) and a 7-parameter Helmert datum shift. Validated against PROJ's GIE regression corpus.
- **Root `terra.Transform` facade.** Reprojects a geometry through `crs.OperationFor` and rebrands via `geom.WithCRS`.
- **`geom.WithCRS`.** Shallow-copy rebrand for swapping the CRS pointer on a geometry without reallocating coordinates.
- **`geom.NewLineStringOwned`, `geom.NewPolygonOwned`, `geom.NewEmptyLineString`.** Donate-ownership constructors for format decoders that have just allocated their own buffers.

### Changed

- **`geom.NewLineStringFlat` / `geom.NewLineStringFlatNoClone` collapsed into `geom.NewLineStringOwned`.** Single ownership-donation constructor; the cloning variant was removed because no caller needed it.
- **`predicate.Intersects` Polygon-vs-Point fast path.** Reordered to check `ContainsPoint` (R-tree direct) before `preparedIntersector.Intersects` (closure dispatch) for the common single-point query.
- **`go vet ./...` is now clean.** All `geom.XY` composite literals are keyed; gofmt drift swept across the tree.

### Fixed

- **`geom/withcrs.go` Point arm.** No longer copies the embedded `sync/atomic.Pointer` envelope cache by value (`out := *v`). Reconstructed field-by-field, matching the LineString/LinearRing/MultiPoint arms. Resolved a `go vet` "copies lock value" warning that masked a real concurrency hazard for cached envelopes.

### Removed

- Unreachable helpers in `buffer/` (`reverseRing`, `cleanRingPolygon`), `overlay/overlayng/` (`overlayCore`, `overlayDisjoint`, plus several smaller shims in `classify`, `spatial_index`, `overlay_lineal`), `predicate/` (`relate_short_circuit` unused branches), `geom/base.go` (`invalidateEnvelope`, no callers), and a handful of internal/relateng + relate-test paths superseded by the Wave 16 RelateNG promotion and the Wave 19/20 buffer rewrite.

## Project history (pre-v1)

Terra was developed over a series of waves driven by the JTS conformance harness and a parallel JTS-API parity sweep. Each wave landed against the JTS `testxml` corpus baseline; the running tally is preserved in [`KNOWN-DIVERGENCES.md`](./KNOWN-DIVERGENCES.md). Highlights:

- **Waves 1–6 (port build-out).** Geometry types, kernels, indexing, R-tree, snap-rounding, the OffsetSegmentGenerator buffer port, and the overlay-NG DCEL pipeline.
- **Waves 7–10 (RelateNG).** Lazy DE-9IM build pipeline ported in 104 commits; `RelateGeometry`, `TopologyComputer`, `RelateNode`, `EdgeSegmentIntersector`, and the surrounding edge/node infrastructure.
- **Wave 11–14 (API parity).** RectangleContains/Intersects, BufferDistance/ResultValidator, PolygonTriangulator, HausdorffSimilarity, AreaSimilarity, MinimumBoundingTriangle, VariableBuffer, KML writer, FrechetSimilarity, MortonCurve, CoveragePolygonValidator, CoverageCleaner, GeometryPrecisionReducer, CommonBits family, IteratedNoder, ScaledNoder, SegmentStringDissolver, SegmentIntersectionDetector, ValidatingNoder, BoundaryChainNoder.
- **Wave 15 (lift gates).** EnhancedPrecisionOp; GML2 reader/writer.
- **Wave 16 (RelateNG promotion).** Closed four RelateNG correctness bugs (point-on-segment robustness, zero-length-line vertex, Point shadowing, AB intersection node snapping); flipped `predicate.Relate` to RelateNG by default and removed the `UseLegacyRelate` opt-out.
- **Waves 17–19 (residual algorithmic gaps).** Closed `TestBufferExternal2 case#97` by reverting `OFFSET_SEGMENT_SEPARATION_FACTOR` to JTS pre-2023 (1e-3); diagnosed `GEOSBuffer#2`.
- **Wave 20 (LineString buffer rewrite).** Routed `bufferLineString` through the polygonizer pipeline (offset segments → snap-round → DCEL → per-subgraph depth labelling → kept-ring extraction → reduced-precision retry); zero algorithmic gaps remain.

The conformance baseline at HEAD is **8 940 / 8 951 (99.88 %)**; the 11 residual failures are all external-tracker known or fixture version drift, documented case-by-case in `KNOWN-DIVERGENCES.md`.

[Unreleased]: https://github.com/terra-geo/terra/compare/HEAD
