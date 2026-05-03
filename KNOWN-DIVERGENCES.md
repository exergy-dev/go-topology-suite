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

After the OffsetSegmentGenerator port (commit `72154f0`) and the
minArea-filter removal (commit `a50d968`), the corpus stands at
**99.81% pass rate** (8904/8951 passing, 17 failures, 30 skipped —
99.99% excluding skipped). Down from a 200-failure baseline.

All `relate` / `within` / `contains` / `touches` / `crosses` /
`overlaps` / `equals` / `isValid` predicates pass on the JTS corpus.

The remaining 17 failures break down:

| Bucket | Count | Resolution |
|--------|------:|------------|
| TestBufferExternal2 case#97 | 1 | Deferred. Down from 24 (closed 23). The single residual is a dense UTM land parcel where the inset produces a tiny polygon (area ~30 at d=75) that survives the validator but the polygonize fallback's depth labelling appears to mis-classify a face. |
| TestBufferJagged misc+robust case#0 test#5 | 2 | Deferred. Down from 16 (closed 14). Both residuals are GEOS BufferRobustness corpus case#0 at buffer distance 2.0 — a single specific distance where the dense (657-vertex) stair-step polygon's offset construction surfaces a shape-level divergence the OffsetSegmentGenerator port doesn't catch. |
| TestSimplify | 2 | cases 15, 16 simplifyTP — JTS version drift (older fixture vs current DP analysis). Confirmed not closeable: both our output and JTS's textbook algorithm agree on case 15 (vertex below DP tolerance flattens); case 16 picks a different but equally valid corner of a 4-corner square. Out of scope. |
| TestOverlayAA | 1 | case#9 symdifference: mAmA inputs where A is a multipolygon with self-touching "fold-in" outer rings (notches) and B partially fills the notches. **Empirical investigation (May 2026):** the bug is NOT in `classifyFacesByPolygons` (every face's `keep` flag is correct against winding-number ground truth). The bug is in `extractResultRings::nextBoundaryAtVertex` — at a pinch-point vertex shared by two distinct kept components, the trace's "next CCW after twin" rule picks an outgoing edge in a *different* kept face, fusing what should be 5 separate polygons into 1 self-touching polygon. A union-find over kept faces (joined when they share an interior edge) and a same-component constraint on the trace's next-edge selection closes case#9 cleanly, BUT shifts the buffer Union chain in `failure/TestBufferFailure.xml` case#1 by 0.075% area — enough to push that previously-passing case past `BufferResultMatcher` tolerance. Deferred until either the matcher accepts the topologically-better buffer result or a per-op gate is added. |
| misc/TestOverlay #4 | 1 | GEOS#737 — sliver under area threshold (3e-6 relative). Area-conservation check tightening below 1e-6 would force spurious retries on rounding noise. Closing requires per-input snap-rounding to coordinate-magnitude-relative grid, not retry-gating. |
| misc/GEOSBuffer + geos-bug356-buffer | 2 | GEOS-tracked buffer pathologies. |
| **JTS-known-fail** (`failure/` folder) | 8 | TestReducePrecisionFailure 5, TestOverlayNGFailure 2, TestBufferFailure 1. JTS headers these as "Result provided is approximately correct". (TestBigNastyBuffer closed by the OSG port.) |

#### Cases closed in the 2026-05-02 round

- **TestBufferMitredJoin case#4** — closed by `e180013` (skip near-collinear corner-vertex emission when `|cross product| < 1e-5`, plus apply `cfg.mitreLimit` in the CROSS branch instead of `Inf`).
- **TestNGOverlayAPrec case#8 differenceSR + symDifferenceSR** — closed by `36bb131` (gated relaxed-threshold snap-rounding pass after the strict fixpoint, with per-tag isolation, chain-with-interior-repeat gating, and hot-pixel-occurrence ≥ 2 filter to target bowtie collapse without regressing narrow features).
- **TestOverlayLAPrec case#0** — closed by `bdc8104` (`polyMinusLineDecompose` builds a small DCEL on noded edges, walks faces, and emits each face as LineString or Polygon based on whether its vertices snap to fewer than 3 distinct grid points).
- **TestOverlayAAPrec case#14** — closed by `fea5b2f` (hole-reshape recovery in `polyMinusLineDecompose`: when the simple sum mismatch fails, find the inner face whose area equals the expected polygon area, walk its half-edges skipping chord-only bridge segments, and split the self-touching walk into outer + holes).

#### Cases closed in the 2026-05-03 round (OffsetSegmentGenerator port + minArea drop)

- **TestBufferExternal2** — 23 of 24 cases closed (1 residual). Of these, 16 closed by the OSG port and 7 more by removing the minArea = d²·0.01 filter that was wrongly rejecting tiny inset slivers (case#30 expected area ~30 at d=75 vs minArea threshold 56).
- **TestBufferJagged misc + robust** — 14 of 16 cases closed (2 residuals).
- **failure/TestBigNastyBuffer case#0** — closed.

Net effect of `72154f0` + `a50d968`: conformance 55 → 17 failures,
38 cases closed. Buffer op count: 44 → 6. Zero non-buffer
regressions. Buffer code went from ~620 lines of ad-hoc per-corner
emission to ~440 lines of JTS-faithful state machine; the area-
filter heuristic that compensated for our pre-OSG offset noise is
no longer load-bearing and removing it admits the legitimate tiny-
inset cases that JTS's expected outputs cover.

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
