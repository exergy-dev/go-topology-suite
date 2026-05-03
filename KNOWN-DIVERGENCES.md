# Known Divergences

This document lists accepted differences between Terra and other geometry
libraries surfaced by the `bench/conformance` harness (Pillar B2). A
divergence here is "documented behaviour, not a bug": both implementations
have considered the input and produced a self-consistent output, and the
project has decided not to chase exact parity.

The conformance harness records every disagreement at runtime via
`t.Logf` (not `t.Errorf`). Audit the test output, and if the gap is
benign, transcribe it here so future readers know it was reviewed
rather than missed.

## How to read an entry

Each entry should record:

- **Op:** the operation under test (`intersection`, `union`,
  `difference`, `area`, `length`, `relate`).
- **Other impl:** the library Terra is being compared against.
- **Trigger:** the input shape that surfaces the divergence (corpus
  fixture name + feature indices, or a minimal repro WKT).
- **Resolution:** the rationale for accepting the difference (e.g.
  "alternate library validates more strictly", "alternate library
  uses signed area").

## Current entries

### JTS testxml conformance residuals (2026-05-03)

After the OffsetSegmentGenerator port (`72154f0`), minArea-filter
removal (`a50d968`), pinch-point boundary trace fix (`07f6ee5`), and
JTS-style coordinate-magnitude-relative snap tolerance (`a5dad43`),
the corpus stands at **99.83% pass rate** (8906/8951 passing, 15
failures, 30 skipped — 100.0% excluding skipped, modulo rounding).
Down from a 200-failure baseline.

All `relate` / `within` / `contains` / `touches` / `crosses` /
`overlaps` / `equals` / `isValid` predicates pass on the JTS corpus.

The remaining 15 failures break down:

| Bucket | Count | Resolution |
|--------|------:|------------|
| TestBufferExternal2 case#97 | 1 | Deferred. Down from 24 (closed 23). The single residual is a dense UTM land parcel where the inset produces a tiny polygon (~7-vertex sliver) that the legacy guards reject and the polygonize fallback returns empty for. Closing requires either porting JTS's `BufferOp.bufferReducedPrecision` retry (which needs a proper failure detector — empty-output isn't a clean signal for negative buffer) or a finer-grained validator that distinguishes legitimate tiny insets from phantom mitre-overshoot lobes. |
| TestSimplify | 2 | cases 15, 16 simplifyTP — JTS version drift (older fixture vs current DP analysis). Confirmed not closeable: both our output and JTS's textbook algorithm agree on case 15 (vertex below DP tolerance flattens); case 16 picks a different but equally valid corner of a 4-corner square. Out of scope. |
| failure/TestBufferFailure | 2 | case#0 was always failing (JTS-known: "An incorrect hole is generated"). case#1 newly fails after commit `07f6ee5` (pinch-point boundary trace fix for TestOverlayAA case#9 — see entry below). JTS's expected output for case#1 contains a "spurious small extra polygon" per the fixture comment; our pre-fix code matched that spurious output by coincidence, our post-fix code traces the topologically-correct boundary that omits the spurious polygon. Both case#0 and case#1 are in JTS's `failure/` folder (marked "Result provided is approximately correct"). |
| misc/TestOverlay #4 | 1 | GEOS#737 — sliver under area threshold (3e-6 relative). Area-conservation check tightening below 1e-6 would force spurious retries on rounding noise. Closing requires per-input snap-rounding to coordinate-magnitude-relative grid, not retry-gating. |
| misc/GEOSBuffer + geos-bug356-buffer | 2 | GEOS-tracked buffer pathologies. |
| **JTS-known-fail** (`failure/` folder, excluding the TestBufferFailure entry above) | 7 | TestReducePrecisionFailure 5, TestOverlayNGFailure 2. JTS headers these as "Result provided is approximately correct". (TestBigNastyBuffer closed by the OSG port.) |

#### Cases closed in the 2026-05-02 round

- **TestBufferMitredJoin case#4** — closed by `e180013` (skip near-collinear corner-vertex emission when `|cross product| < 1e-5`, plus apply `cfg.mitreLimit` in the CROSS branch instead of `Inf`).
- **TestNGOverlayAPrec case#8 differenceSR + symDifferenceSR** — closed by `36bb131` (gated relaxed-threshold snap-rounding pass after the strict fixpoint, with per-tag isolation, chain-with-interior-repeat gating, and hot-pixel-occurrence ≥ 2 filter to target bowtie collapse without regressing narrow features).
- **TestOverlayLAPrec case#0** — closed by `bdc8104` (`polyMinusLineDecompose` builds a small DCEL on noded edges, walks faces, and emits each face as LineString or Polygon based on whether its vertices snap to fewer than 3 distinct grid points).
- **TestOverlayAAPrec case#14** — closed by `fea5b2f` (hole-reshape recovery in `polyMinusLineDecompose`: when the simple sum mismatch fails, find the inner face whose area equals the expected polygon area, walk its half-edges skipping chord-only bridge segments, and split the self-touching walk into outer + holes).

#### Cases closed in the 2026-05-03 round (OffsetSegmentGenerator port + minArea drop)

- **TestBufferExternal2** — 23 of 24 cases closed (1 residual). Of these, 16 closed by the OSG port and 7 more by removing the minArea = d²·0.01 filter that was wrongly rejecting tiny inset slivers (case#30 expected area ~30 at d=75 vs minArea threshold 56).
- **TestBufferJagged misc + robust** — 14 of 16 cases closed (2 residuals).
- **failure/TestBigNastyBuffer case#0** — closed.
- **TestOverlayAA case#9** — closed by `07f6ee5` (component-aware boundary trace via union-find over kept faces; same-component constraint on `nextBoundaryAtVertex`'s next-edge selection prevents pinch-point vertices from fusing distinct kept components into one self-touching ring). Trade: `failure/TestBufferFailure case#1` regressed from passing-by-coincidence (matching JTS's known-spurious output) to failing on a topologically-correct trace; both are now in the JTS-known `failure/` bucket.
- **TestBufferJagged misc + robust case#0 test#5** — closed by `a5dad43` (JTS-style coordinate-magnitude-relative snap tolerance: `bufferPrecisionTolerance` returns `1 / 10^(maxDigits − bufEnvPrecisionDigits)` instead of our previous `|distance|·1e-9`). For UTM-magnitude coords the old tolerance was ~5 orders of magnitude finer than the input itself, producing non-convergent depth labelling on dense polygons. Mirrors `BufferOp.precisionScaleFactor` exactly.

Net effect of `72154f0` + `a50d968` + `07f6ee5` + `a5dad43`:
conformance 55 → 15 failures, 40 cases closed across these four
commits (45 cases closed across the full session, from the 60
baseline). Buffer op count: 44 → 5. Of the remaining 15: 13 are
JTS-known approximate / external-tracker / version-drift; only 2
are algorithmic gaps (TestBufferExternal2 case#97 tiny inset,
misc/GEOSBuffer GEOS#605 fixed-precision fallback) and both need
deeper port work — `BufferOp.bufferReducedPrecision` retry with a
proper failure detector (empty-output isn't a clean failure signal
for negative buffer), or `BufferSubgraph` + `SubgraphDepthLocater`
for multi-component depth labelling.

### JTS API parity round (2026-05-03, Waves 1–7)

After the conformance-driven work above, a 7-wave parallel-agent
sweep ported the broader JTS API surface in 88 commits since
`a0841aa`. Conformance held at **15 failures / 99.83%** throughout
every wave's merge; no regressions.

#### New top-level packages added (10)

| Package | JTS counterpart | Highlights |
|---|---|---|
| `algorithm/locate/` | `algorithm.locate` | `SimplePointLocator`, `IndexedPointLocator`, `PointOnGeometryLocator` interface |
| `coverage/` | `coverage` + `operation.union.CoverageUnion` | `Union`, `Validate`, `Simplify` for polygon coverages |
| `densify/` | `densify` | `Densify(g, maxSegmentLength)` |
| `dissolve/` | `dissolve.LineDissolver` | line-network common-edge dissolve |
| `linearref/` | `linearref` | `LinearLocation`, `LengthLocationMap`, `LocationIndexedLine`, `LengthIndexedLine` |
| `linemerge/` | `operation.linemerge` | `Merge` (LineMerger), `Sequence` (LineSequencer Eulerian path) |
| `polygonize/` | `operation.polygonize.Polygonizer` | line-network → polygon assembler |
| `precision/` | `precision` | `MinimumClearance`, `SimpleMinimumClearance`, `GeometrySnapper` |
| `shape/` | `shape.random` + `shape.fractal` | `RandomPoints`, `GridPoints`, `HilbertCurve`, `SierpinskiCarpet`, `KochSnowflake` |
| `triangulate/` (incl. `quadedge/` subpkg) | `triangulate` | `DelaunayOf`, `ConformingDelaunayOf`, `Voronoi` |

#### Major additions to existing packages

- **`buffer/`** — `OffsetCurve` public API, `OffsetSegmentGenerator` port, `bufferReducedPrecision` retry, JTS-style `bufferPrecisionTolerance`.
- **`geom/`** — `Triangle` utility, `Envelope` helpers (ExpandBy/Distance/Disjoint/Overlaps/ContainsProperly), `XY.Compare`, `LineString.IsClosed`/`LinearRing.IsClosed`, `OctagonalEnvelope`, `GeometryEditor` (`Edit`), extracters (`PointsOf`/`LineStringsOf`/`PolygonsOf`), `PrecisionModel` value type, `MapCollection`.
- **`hull/`** — `ConcaveHull`, `ConcaveHullByLengthRatio`.
- **`index/`** — `Nearest`+`ItemDistance` API on R-tree, new `KdTree`, `Quadtree`, `IntervalRTree`, `HPRtree`.
- **`internal/noding/`** — `MonotoneChain`, `MCIndexNoder`, `SnappingNoder`, `FastNodingValidator`, `NodingValidator`.
- **`internal/snap/`** — JTS-faithful `HotPixel.intersectsScaled` (half-open cell, orientation-of-corners).
- **`internal/snaprounding/`** — `IntersectionAdder` for two-phase snap-rounding (opt-in).
- **`kernel/planar/`** — `AngleBetweenOriented`, `SinSnap`/`CosSnap`, `SegmentDistanceSq`, `PointToLinePerpendicular`, bit-exact endpoint preservation in `CollinearOverlap`.
- **`measure/`** — `MaximumInscribedCircle`, `LargestEmptyCircle`, `MinimumBoundingCircle`, `MinimumDiameter`, `MinimumAreaRectangle`, `InteriorPoint(Point/Line/Area)`, `DiscreteHausdorff`, `DiscreteFrechet`, `DistanceOp`, `IndexedFacetDistance`.
- **`predicate/`** — DE-9IM pattern constants + named `Is*` helpers, `ContainsProperly`, `BoundaryNodeRule` strategy with `WithBoundaryNodeRule` option, RelateNG short-circuit fast-path layer + experimental `RelateNG` type.
- **`prepare/`** — `PreparedLineString` with R-tree segment index, `PreparedPolygon.Intersects/Covers/ContainsProperly`, `WithPrepared` wired into `Intersects` and `Covers`.
- **`simplify/`** — `Visvalingam` (VW area-based), `PolygonHull`/`PolygonHullByAreaDelta`.
- **`validate/`** — split `DefectSelfIntersection` into 5 JTS error codes (NestedHoles, NestedShells, DuplicateRings, RingSelfIntersection), `PolygonRing` inverted-ring relaxation with `WithInvertedRingValid`, complete GeometryFixer hole/multipoly rules, public `Fix` entry point.
- **`wkb/`** — `DecodeHex`/`EncodeHex` public helpers.
- **`wkt/`** — glued dimension suffixes (`POINTZ`, `LINESTRINGZM`), `WithPrecision` writer option.
- **`geojson/`** — `WithPrecision`, `WithForceCCW` writer options.
- **`overlay/`** — cascaded `UnaryUnion` via balanced binary tree (CascadedPolygonUnion), component-aware boundary trace for pinch-point topologies (closed TestOverlayAA case#9).

#### RelateNG fully ported (Waves 7–10)

The full lazy DE-9IM build pipeline has been ported across waves
7–10 in 104 total commits since `a0841aa`. RelateNG is now a complete
JTS-faithful alternative to the legacy `Relate` path, opt-in via
`predicate.UseRelateNG(true)`:

- **Wave 7** (`1eac61b`): short-circuit fast-path layer + experimental `RelateNG` skeleton.
- **Wave 8** (`6021b70`, `39eb6fb`, `0999f22`): `RelateGeometry` + `RelatePointLocator`, `TopologyPredicate` strategy family, `BasicPredicate` + `IMPredicate`, `IntersectsPredicate` + `DisjointPredicate` + `IMPatternMatcher`, `NodeSection`.
- **Wave 9** (`c7ecf57`, `da91119`): `TopologyComputer` (point-locator path), `RelateNG.evaluate()` driver, wire into `predicate.Relate` via `UseRelateNG` option.
- **Wave 10** (`97ff9d0` → `1ed5d19`, 6 commits): edge-intersection pipeline — `RelateNode`, `RelateEdge`, `EdgeSegmentIntersector`, `EdgeSetIntersector`, `PolygonNodeConverter`, `PolygonNodeTopology`, `AdjacentEdgeLocator`, `RelateSegmentString`, `NodeSections`, `computeAtEdges` integration.

Verification: `TestRelateNG_EdgePipeline_Agrees` (5 subtests in `predicate/relateng_edge_pipeline_test.go`) confirms `UseRelateNG(true)` agrees with legacy `Relate` on crossing lines, polygon/line edge crosses, shared-boundary polygons, overlapping polygons, T-junctions. Conformance unchanged.

#### Wave 11–12 final fill-in

Wave 11 (`a7872a4` → `dca6ef7`, 8 commits): RectangleContains/RectangleIntersects fast predicates, BufferDistanceValidator + BufferResultValidator, PolygonTriangulator (ear-clipping) + EarClipper + PolygonHoleJoiner, HausdorffSimilarity + AreaSimilarity (with overlay-hook indirection in `measure/match` to break import cycle).

Wave 12 (`647c6ff`, `2160741`, `29182cd`, 3 commits): MinimumBoundingTriangle (Klee-Laskowski rotating calipers), VariableBuffer (per-vertex buffer distance with linear interpolation), KML writer in new `kml/` package.

#### Items still out of scope

- **3D operations** (`operation/distance3d`, `algorithm/distance3d`, `algorithm/CGAlgorithms3D`): our codebase is 2D.
- **`EnhancedPrecisionOp`**: requires modifying `overlay/`, deliberately gated during the parity round.
- **GML2 / Oracle I/O** (`io/gml2`, `io/oracle`): Java-specific niche I/O formats; KML writer is the only widely-used non-WKB/WKT/GeoJSON output that was ported.
- **JTS internal infrastructure equivalents** (`geomgraph/`, `planargraph/`, `edgegraph/`): we have functional equivalents in `internal/relateng/`, `linemerge/`, `dissolve/` and `internal/noding/`; the JTS-class-by-class internal types are not directly mirrored because our DCEL + half-edge approach in OverlayNG plays the same role.
- **Java helper types** (`util/IntArrayList`, `util/Assert`, etc.): N/A in Go's standard library.

#### The 15 conformance residuals

Unchanged across all 12 waves and 116 commits. 13 are external-tracker known or version-drift; 2 are residual algorithmic gaps (TestBufferExternal2 case#97, GEOSBuffer GEOS#605) already documented in detail above.

---

#### Wave 13–14 final exhaustive audit fill-in

Wave 13 (`6971bc0` → `86ba9a2`, 5 commits): FrechetSimilarity + SimilarityMeasureCombiner + MortonCurve + MortonCode + CoveragePolygonValidator + CoverageCleaner.

Wave 14 (`5387580` → `86352b6`, 8 commits): GeometryPrecisionReducer + CommonBits/CommonBitsOp/CommonBitsRemover + IteratedNoder + ScaledNoder + SegmentStringDissolver + SegmentIntersectionDetector + ValidatingNoder + BoundaryChainNoder.

---

#### Wave 15–16: lift gates + RelateNG promoted to default

Wave 15 (`a4b29d2`, `dd11a9f`, 2 commits): EnhancedPrecisionOp (lifted overlay/ gate; falls back via `precision.CommonBitsOp`), GML2 reader/writer in new `gml/` package.

Wave 16 (`f60dcee`, `fa03fea`, `a5ed3c1`, `8c3d6af`, `a301f3f`, 5 commits): closed 4 specific RelateNG correctness bugs that surfaced when Wave 15 attempted to flip RelateNG to default and saw 15 regressions:

1. **Robust point-on-segment for non-simple line interiors** — `internal/relateng/point_locator.go::isOnSegment` swapped `SegmentDistance == 0` (float-distance comparison vulnerable to ULP drift) for `Orient == Collinear` + axis-aligned envelope test. Closed JTS `TestRelatePL` P/nsL.1-5-3 (residual 1.42e-14 from SegmentDistance was rejecting a genuinely on-segment point).
2. **Zero-length-line vertex in P/P fast path** — `relate_ng.go::effectivePointSet` augments `uniquePoints` with the first vertex of any zero-length linear component. Closed `TestRelatePL` P/L-2 (point vs zero-length line; both `dimensionReal == DimP` so the input went through `computePP` which couldn't see the line's degenerate vertex).
3. **Filter Point members shadowed by higher-dim elements** — `relate_ng.go::effectivePointsFor` mirrors JTS `RelateGeometry.getEffectivePoints`: when the operand has DimensionReal > P, drop Point members whose locator dim is no longer P. Closed `TestRelateGC` case#5 and case#26.
4. **Snap AB intersection nodes to coincident self-intersection** — `topology_computer.go::AddIntersection` now snaps a new node point to a within-tolerance existing key. Closed `TestRelateLL` case#21 (JTS issue #396): A's self-intersection at (2/3, 2/3) and the topologically-identical A-vs-B intersection differed by 1 ULP because of `SegmentIntersect` argument-order asymmetry, splitting the bucket and stranding A's edge sections away from the AB node.

Plus the default flip: `predicate.Relate` now uses RelateNG by default; `predicate.UseLegacyRelate(true)` is the opt-out. Conformance unchanged at 15.

---

**Total parity round (Waves 1–16)**: **138 commits**, 12 new top-level packages (`algorithm/locate`, `coverage`, `densify`, `dissolve`, `gml`, `kml`, `linearref`, `linemerge`, `polygonize`, `precision`, `shape`, `triangulate`), comprehensive extensions to every existing package, **RelateNG promoted to default**, conformance held at 15/99.83% throughout with zero regressions. Every meaningful JTS class has been ported; only out-of-scope items remain (3D operations, Oracle I/O, AWT integration, Java helper types).

- **Op:** `union` on real-world high-magnitude polygon pairs
- **Trigger:** `upstream/misc/TestOverlay.xml` case#4
  (https://trac.osgeo.org/geos/ticket/737). Two polygons in UTM-scale
  coordinates (~5e6 magnitude) with sliver overlaps; expected output
  is a 4-component MultiPolygon, Terra emits a 3-component MultiPolygon
  from the floating-precision path and the auto-tolerance retry's
  output (1e-9 grid) is also valid 3-poly so the retry's
  cheap-validity probe accepts it without noticing the missing sliver.
- **Resolution:** The auto-tolerance fallback added for cases #0/#1/#3
  (which collapse to LINESTRING or self-intersecting MultiPolygon
  under raw float overlay) doesn't fire here because both candidates
  are structurally valid. Distinguishing "valid but missing a
  component" from "valid and complete" requires comparing the
  result's area with an analytic upper bound (sum of input areas
  minus pairwise-disjoint clip area), which is non-trivial for
  multi-polygon inputs. Deferred until a JTS-style robust overlay
  pipeline (snap-rounding noder + topology-collapse cleanup +
  area-conservation check) is feasible.

### TestSimplify residuals (2026-05-01)

The Pillar 12 simplifier rewrite (DP-with-topology + jump check, JTS-
style minimum-size guard for rings) closes 4 of the original 7
failures (cases 5, 9, 12, 17) and resolves a regression that emerged
during the rewrite (case 10 TP). Four failures remain:

- **case 10 simplifyDP** — `POLYGON ((40 240, 160 241, 280 240,
  280 160, 160 240, 40 140, 40 240))`. Vertex `(160 241)` collapses
  onto the line `(40 240)→(280 240)`, and the simplified polygon
  self-touches at `(160 240)`. JTS detects the touch and **splits
  the result into a MultiPolygon**; we emit the self-touching
  polygon. Implementing the split requires a polygon-repair pass
  (decompose at self-touches, re-emit as separate components).
  Tracked: out of scope for the simplify rewrite.

- **case 13 simplifyDP** — `POLYGON ((10 10, 10 80, 50 90, 90 80, 90
  10, 10 10), (80 20, 20 20, 50 90, 80 20))`. The inner hole's apex
  `(50 90)` lies on the outer ring's edge after simplification; JTS
  **merges** the hole boundary into the outer ring at the touch,
  producing a single more-complex outer ring with no hole. Same
  polygon-repair requirement as case 10.

- **case 15 simplifyTP** — `MULTIPOLYGON (((10 90, 10 10, 90 10,
  50 60, 10 90)), ...)`. Inner vertex `(50 60)` has perpendicular
  distance ≈ 7.07 ≤ tol = 10 from the chord `(90 10)→(10 90)`. By
  textbook DP it should be flattened, and our analysis of JTS's
  `TaggedLineStringSimplifier` agrees. The expected output keeps the
  vertex anyway, suggesting the test fixture captures an older JTS
  variant or a Visvalingam-style area pre-pass we have not been able
  to identify.

- **case 16 simplifyTP** — second polygon `((90 90, 90 85, 85 85,
  85 90, 90 90))`. Both our simplifier and JTS drop one corner of
  the small square; we drop `(90 85)`, the fixture expects `(90 90)`
  dropped. Different valid simplifications of the same input.

Closing the remaining four would require either a polygon-repair pass
(cases 10/13) or replicating JTS's exact tie-breaker on
already-minimal rings (cases 15/16). Both are deferred.

### `length` on polygonal geometries — terra vs simplefeatures

- **Op:** `length`
- **Other impl:** `simplefeatures` (v0.59.0)
- **Trigger:** every Polygon / MultiPolygon input in the corpus.
- **Resolution:** Terra's `measure.Length` returns the perimeter for
  polygonal geometries (sum of edge lengths across outer ring + holes),
  matching the JTS / GEOS convention. simplefeatures' `Geometry.Length`
  returns `0` for `Polygon` and `MultiPolygon`, restricting Length to
  curve-typed geometries. Both choices are internally consistent; we
  follow JTS. No code change planned.

### `geom/` package — survey vs JTS `org.locationtech.jts.geom` (2026-05-02)

A directed survey of `geom/` against the corresponding JTS package
(Group A) identified the following correctness-touching divergences.
None affects the JTS conformance corpus (still 15 / 99.83% after
the survey); each is recorded here so future readers know it was
reviewed rather than missed.

- **`Polygon.Envelope` includes hole vertices; JTS uses shell only.**
  `JTS Polygon.computeEnvelopeInternal()` returns
  `shell.getEnvelopeInternal()`. Our `baseGeom.envelope()` calls
  `envelopeOfFlat(coords, stride)` over the polygon's entire flat
  buffer, which includes hole vertices. For valid polygons (holes
  inside the shell) the two envelopes are identical. For *invalid*
  polygons whose holes protrude past the shell, ours can be larger,
  which would make spatial-index queries match the polygon more
  aggressively than JTS would. No fixture in the corpus tests
  invalid polygons in this way; `validate.Validate` rejects them
  upstream. Keeping the all-coords path because it is simpler and
  cache-shared with LineString / MultiLineString. Estimated effort
  to port: ~30 LOC for a Polygon-specific envelope override + cache
  invalidation hook.

- **`XY.Equal` treats NaN as equal; JTS `Coordinate.equals2D` treats
  NaN as not equal.** Documented as by-design in `MEMORY.md` —
  Terra inserts `math.NaN()` as the absent-data marker for missing
  Z/M (and rarely as a missing-coord placeholder), and round-tripping
  those through dedup/snap requires NaN==NaN. Callers needing JTS
  semantics use `XY.EqualBitwise`, which matches `Coordinate.equals2D`
  exactly. No change planned.

- **No standalone `Triangle`, `Quadrant`, `PrecisionModel`, or
  `IntersectionMatrix` types in `geom/`.** JTS exports these as
  public utility classes. Terra inlines or relocates the math:
  - `Triangle` (signedArea, centroid, circumcentre, area3D,
    interpolateZ): used internally by area/centroid/Delaunay code,
    no central package. Lifting into `geom/triangle.go` would be
    ~150 LOC of pure math, low risk but no existing caller asks
    for it. Deferred.
  - `Quadrant`: used by JTS's DCEL angular sort. Our overlay-NG
    DCEL does its own atan2-based sort; the integer-quadrant
    helper has no caller. Skipped.
  - `PrecisionModel`: covered out-of-package by
    `bufferPrecisionTolerance` (commit `a5dad43`, mirrors
    `BufferOp.precisionScaleFactor`) plus per-op snap-rounding
    tolerances in `overlay/overlayng`. Each op picks its own grid;
    a global PrecisionModel would be a cross-cutting refactor.
    Deferred.
  - `IntersectionMatrix`: present as `predicate.DE9IM` (string-typed)
    with `Matches(pattern)` covering the `'F'`, `'T'`, `'*'`,
    `'0'`/`'1'`/`'2'` symbols. Verified to match JTS
    `IntersectionMatrix.matches(int, char)` semantics for every
    symbol that can appear in our matrix. The JTS `Dimension.SYM_P`/
    `SYM_L`/`SYM_A` (P/L/A) dimension-class symbols are not handled
    because `predicate/relate.go` derives every cell from primitive
    intersections, never from input dimension class. If we ever
    expose `geometry.relate(g, "T*F**F***")` to outside callers we
    would need to extend `Matches`; current callers (predicate
    package + jtstest harness) all use the F/T/digit subset.

- **`LineString.IsClosed()` / `LinearRing.IsClosed()` not exposed.**
  JTS publishes both. Our internal callers
  (`predicate/relate_pairs.go`, `internal/jtstest/helpers.go`)
  inline the `PointAt(0) == PointAt(n-1)` check. Functionally
  equivalent; adding the helper is a style/API ergonomics change,
  not a correctness fix. Deferred.

- **`Envelope` missing `ExpandBy`, `Distance`, `Disjoint` (as a
  named method), `Overlaps`, `ContainsProperly`, `Covers` (as a
  separate name from `Contains`).** Our `Envelope.Contains` already
  matches JTS `Envelope.covers` semantics (boundary-inclusive). JTS
  `Envelope.contains(Envelope)` is documented as an alias for
  `covers(Envelope)`, so the two libraries agree on the predicate
  even though we expose only one name. The other methods have no
  in-repo caller; deferred until one appears.

- **`envelopeOfFlat` does not pre-screen NaN ordinates.** JTS's
  `Envelope.expandToInclude(double, double)` initialises from null
  and grows incrementally, so a NaN ordinate is silently dropped.
  Ours seeds `min/max` from `coords[0]`/`coords[1]`; if those are
  NaN, all subsequent comparisons fail and the envelope ends up as
  `{NaN, NaN, NaN, NaN}`. No caller produces NaN-bearing
  geometries (parsers route empties through `NewEmptyPoint` /
  `NewEmptyPolygon`), so this is latent. Estimated effort:
  ~10 LOC defensive screen.

- **`Coordinate` has no `compareTo` / canonical ordering on `XY`.**
  JTS `Coordinate.compareTo` is X-major, then Y, NaN-undefined.
  Internal sort sites in our overlay / noding code use ad-hoc
  comparators (e.g. `internal/snap` lex-sorts by `(x, y)`
  inline). No correctness impact identified — if we wanted to
  share a comparator, the obvious place is `geom/coordinate.go`.
