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

After Pillar 1, 2/3 (partial), 4 P1 (positive buffer polygonization),
5, 6 P1 (line-on-polygon-boundary collinear overlap), and 7 work, the
corpus stands at **98.5% pass rate** (8817/8951 passing, 104 failures,
30 skipped). All `relate` / `within` / `contains` / `touches` /
`crosses` / `overlaps` / `equals` predicates pass on the JTS corpus.

The remaining 104 failures concentrate in:

- **55 buffer** (TestBufferExternal2: 31, TestBufferJagged misc+robust: 16,
  TestBufferMitredJoin: 4, plus a handful in failure/ folder).
  Negative-buffer cases need full Pillar 4 P2: tolerance-based spike
  removal in offset emission + JTS-style depth-determination via
  subgraph-finder. Half-measures (overshoot guards) catch only the
  trivially-thin polygons.
- **30 overlay-precision** (TestOverlayAAPrec, TestNGOverlayAPrec,
  TestNGOverlayLPrec, etc.). Sliver dimensional collapse cases that
  Pillar 3's spur-edge / figure-8 work caught the easy ones; tight
  near-collinear segments still fail.
- **4 TestSimplify**: Douglas-Peucker / topology-preserving edge cases.
  See "TestSimplify residuals (2026-05-01)" below for case-by-case
  detail. The remaining failures all need either polygon-repair after
  edge collapse (DP cases 10, 13) or matching JTS's exact corner-
  selection heuristic on tiny rings, which the test expectations seem
  to capture from a non-current JTS version.
- **5 TestReducePrecisionFailure** + 3 TestOverlayNGFailure + 2
  TestBufferFailure + 1 TestBigNastyBuffer: documented JTS-known-fail
  cases ("Result provided is approximately correct").

- **Op:** `buffer` (~87 failures)
- **Other impl:** JTS / GEOS overlay-NG buffer with snap-rounding
- **Trigger:** TestBuffer, TestBufferExternal2, TestBigNastyBuffer,
  TestBufferFailure, TestBufferInsideNonEmpty.
- **Resolution:** Terra's `buffer.Buffer` uses an offset-curve
  generator without robust snap-rounding cleanup. Buffer correctness
  for complex inputs (long zig-zag rings, sharp concavities, nearly-
  parallel edges, exact JTS round-cap subdivision) requires a
  snap-rounding noder + boundary-merge pass. Tracked.

- **Op:** `relate` / cascading predicates on GeometryCollection /
  overlapping multi cases (~150 across `relate`/`within`/`contains`/
  `touches`)
- **Trigger:** TestRelateGC, TestRelateLL (validate suite),
  TestRelateLA, TestRelateAA-big (skinny-polygon precision).
- **Resolution:** Terra's `predicate.Relate` flattens multi-geometries
  and combines per-member matrices. Aggregate-boundary post-processing
  (mod-2 for MLS, closed-line detection) is implemented in Phase 11,
  but exact results for overlapping GC members + segment-level GC
  boundary semantics need a global noded relate engine. Tracked.

- **Op:** mixed-dimension overlay results
- **Trigger:** TestOverlayAA case#3 (intersection that produces
  P+L+A in a GeometryCollection), TestNGOverlayA precision sliver
  cases.
- **Resolution:** Terra's overlay-NG result extractor only emits
  result polygons; result lines (overlay-shared edges that don't
  bound result polygons) and result points (vertex-only intersections)
  need additional DCEL traversal. Phase 12 in the followup plan
  proposes this work but it has not yet landed.

- **Op:** snap-rounding precision overlays (`*SR` arg3 ≠ 1, `*Prec`
  test files ~120)
- **Trigger:** TestOverlayAAPrec, TestNGOverlayAPrec, TestNGOverlayLPrec.
- **Resolution:** Terra has no true snap-rounding noder. The harness
  pre-snaps inputs to the precision scale before running overlay
  (Phase 14), which handles the simpler cases; sliver-precision
  overlays where snap-rounding is required during the noding step
  itself still fail. A real `internal/snaprounding/noder.go` is the
  proper fix.

- **Op:** `isValid` on complex polygon validity
- **Trigger:** TestValid case#74–86 (interior-disconnected via hole
  chains), TestInvalidA (adjacent-hole chains, zero-width spikes
  along boundary).
- **Resolution:** These need a noded validation analysis (build
  edge graph of all rings, check connectedness of polygon interior).
  Single-ring + simple hole-pair checks are implemented in Phase 10.

- **Op:** various AA real overlay correctness (~30)
- **Trigger:** TestOverlayAA case#1 (polygon with hole intersecting
  another polygon) and similar.
- **Resolution:** Terra's overlay-NG doesn't always correctly
  subtract holes when computing intersection of polygon-with-hole
  vs polygon. Needs targeted fixes in
  `overlay/overlayng/result.go`.

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
