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

### JTS testxml conformance residuals (2026-05-02)

After Pillars 1–7 + Streams A–G + G1–G4 + post-G4 round (buffer-polygonize
upper bound, polygon-vs-line touch-point emission, orientation-tolerant
Polygon Equals, area-conservation upper-bound check, phantom-sliver-hole
filter for floating-precision Union), the corpus stands at **99.0% pass
rate** (8861/8951 passing, 60 failures, 30 skipped — 99.3% excluding
skipped). Down from a 200-failure baseline.

All `relate` / `within` / `contains` / `touches` / `crosses` /
`overlaps` / `equals` / `isValid` predicates pass on the JTS corpus.

The remaining 60 failures break down:

| Bucket | Count | Resolution |
|--------|------:|------------|
| TestBufferExternal2 (negative buffer of land parcels) | 24 | Needs deeper buffer rework: better representative-point selection (inscribed-circle vs midpoint-of-longest-segment) + JTS-style winding-number depth-against-original (continuous quantity, not binary point-in-polygon). Phantom subgraph rep-points sit at distances 37–46 from the original boundary while legitimate inset rep-points sit at 1–15 due to ULP-noise on dense polygons; relaxing the validator threshold admits MORE phantoms (per agent A's empirical investigation). Deferred — requires medial-axis-style analysis to distinguish phantom from legitimate at this geometric resolution. |
| TestBufferJagged misc+robust | 16 | Positive buffer on jagged polygons; routes through `polygonizeBuffer` without filter. Failures are vertex-sequence Hausdorff mismatches (shape mismatch), not face-count issues. Adding minArea filter is a no-op. Closing needs robust noding + JTS-style depth labelling that handles dense near-collinear vertices without snap-rounding artefacts (1–2 weeks of focused work). |
| TestSimplify | 2 | cases 15, 16 simplifyTP — JTS version drift (older fixture vs current DP analysis). Out of scope. |
| TestNGOverlayAPrec | 2 | case#8 differenceSR/symDifferenceSR: JTS inserts `(4,1)` as a vertex on the `(2,1)→(4,2)` segment via an extended-cell hot-pixel test (perpendicular distance ≈0.894 > tolerance/2=0.5). Connectivity-restricted `MergeNearCollinear` pass attempted but introduced sliver-collapse regressions; needs JTS-style hot-pixel adjacency rule that doesn't merge legitimate narrow features. |
| TestBufferMitredJoin | 1 | Mitre-join with reflex corner; same root cause as TestBufferExternal2. |
| TestOverlayAAPrec | 1 | Polygon-difference-LineString case#14: B is a `LineString`, routed through float `overlay.Difference` (not snap-rounded NG). Hole reshaping under tolerance=1 unhandled by float path. |
| TestOverlayLAPrec | 1 | case#0 difference at scale=1: sliver polygon dimensionally collapses to a vertical line; JTS decomposes into Polygon+LineString, we return the polygon unchanged. Hybrid-DCEL polygon-snap-rounding decomposer needed. |
| TestOverlayAA | 1 | case#9 symdifference: mAmA inputs where A is a multipolygon with self-touching "fold-in" outer rings (notches) and B partially fills the notches. The DCEL classifier mis-keeps the unfilled notch face when both inputs are multi-polygon; root cause sits in `classifyFacesByPolygons` boundary-tag-aware nudging interaction with multi-polygon `inSubj/inClip` membership. SymmetricDifference's Union-of-(A\B,B\A) path inherits the bug. Deferred — needs face-classification rework for multi-polygon inputs with shared seams. |
| misc/TestOverlay #4 | 1 | GEOS#737 — sliver under area threshold (3e-6 relative). Area-conservation check tightening below 1e-6 would force spurious retries on rounding noise. Closing requires per-input snap-rounding to coordinate-magnitude-relative grid, not retry-gating. |
| misc/GEOSBuffer + geos-bug356-buffer | 2 | GEOS-tracked buffer pathologies. |
| **JTS-known-fail** (`failure/` folder) | 9 | TestReducePrecisionFailure 5, TestOverlayNGFailure 2, TestBufferFailure 1, TestBigNastyBuffer 1. JTS headers these as "Result provided is approximately correct". |

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
