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

After Pillars 1, 2/3 (partial), 4 P1+P2 (buffer polygonization +
tolerance-aware spike removal), 5, 6 P1 (line-on-polygon-boundary
collinear overlap), 7, simplify rewrite, overlay auto-tolerance
retry for FLOATING-precision real-world ticket cases, Stream G
(snap-rounding lineal overlay), centroid-based reclassification
for snap-rounded sliver faces (closes TestOverlayAAPrec#16
union/symdifference), and isolated-touch-point emission for line-
polygon intersection (closes TestOverlayLA cases#2,#3), the corpus
stands at **99.0% pass rate** (8860/8951 passing, 61 failures,
30 skipped — 99.3% excluding skipped).

All `relate` / `within` / `contains` / `touches` / `crosses` /
`overlaps` / `equals` / `isValid` predicates pass on the JTS corpus
(from a starting point of 200 failures).

The remaining 85 failures break down as:

| Bucket | Count | Resolution |
|--------|------:|------------|
| TestBufferExternal2 (negative buffer of land parcels) | 31 | Needs JTS-style subgraph-finder for depth determination in `buffer/polygonize.go::labelFaceDepths`. Tolerance-aware spike removal landed (Pillar 4 P2 partial); subgraph propagation deferred. |
| TestBufferJagged misc+robust | 16 | Same Pillar 4 P2 deferred work; sharp-corner offset overshoots produce spurious lobes that face-validity heuristics drop overaggressively. |
| ~~TestNGOverlayLPrec~~ | 0 | Closed by Stream G (snap-rounding lineal overlay entry point in `overlay/overlayng/overlay_lineal.go`). |
| TestSimplify | 4 | See "TestSimplify residuals" below. Needs polygon-repair pass for self-touches (cases 10/13), and corner tie-break to match an older JTS (cases 15/16). |
| TestBufferMitredJoin | 4 | Mitre-join with reflex corners; same Pillar 4 P2 root cause. |
| TestOverlayAAPrec | 1 | Polygon-difference-LineString case#14: B is a `LineString`, routed through float `overlay.Difference` (not snap-rounded NG). Hole reshaping under tolerance=1 unhandled by float path; deferred. |
| TestNGOverlayAPrec | 2 | case#8 differenceSR/symDifferenceSR: JTS inserts `(4,1)` as a vertex on the `(2,1)→(4,2)` segment via an extended-cell hot-pixel test (perpendicular distance ≈0.894 > tolerance/2=0.5). Connectivity-restricted `MergeNearCollinear` pass attempted but introduced sliver-collapse regressions on case#2 (narrow wedge), case#4 (close shells), case#13 (outward-sliver hole). Deferred — needs JTS-style hot-pixel adjacency that doesn't merge legitimate narrow features. |
| ~~TestNGOverlayPPrec~~ | 0 | Closed by Stream G (asymmetric Point/Line topology check against original geometry). |
| ~~TestOverlayLA~~ | 0 | Closed by isolated-touch-point emission for line-polygon intersection (`overlay/line_overlay.go::linePolygonOverlay`). |
| TestOverlayLAPrec | 1 | case#0 difference at scale=1: polygon `(95 9, 81 414, 87 414, 95 9)` is a sliver whose two long edges both round to `x=95` at `y≤13`, dimensionally collapsing the lower portion into a vertical line `(95 9, 95 13)`. JTS decomposes the snap-rounded polygon into Polygon+LineString; we currently return the polygon unchanged from the (poly\line) branch. Closing requires a polygon-snap-rounding decomposer that emits the collapsed sliver as a separate lineal component (hybrid-DCEL rewrite). Deferred. |
| TestOverlayAA | 1 | Complex AA touching+overlapping; one residual not closed by Pillar 1 work. |
| TestUnaryUnionFloating | 1 | Real-world MultiPoint union with closely-clustered coords. |
| misc/TestOverlay #4 | 1 | GEOS ticket #737 — UTM-scale polygon pair with missing sliver under floating-precision; auto-tolerance retry produces structurally-valid 3-poly output but missing one component. Detecting "valid but incomplete" needs analytic area-conservation check; deferred. |
| misc/GEOSBuffer + geos-bug356-buffer | 2 | GEOS-tracked buffer pathologies. |
| **JTS-known-fail** (`failure/` folder) | 11 | TestReducePrecisionFailure 5, TestOverlayNGFailure 2, TestBufferFailure 1, TestBigNastyBuffer 1, plus 2 distributed. JTS itself headers these as "Result provided is approximately correct". Opportunistic to close. |

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
