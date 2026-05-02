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

### JTS testxml conformance residuals (2026-05-01)

After parallel-pillar work (DCEL pinch-point + multi-component face
classification; planar.SegmentIntersect endpoint snapping; negative-
buffer overshoot guards; multi-polygon overlap-via-centroid; precision-
result snapping; spur-edge lineal extraction + figure-8 ring split;
buffer hole-erosion via recursion + multi-line union; LinearRing type
with isValid self-intersection check; GC line-endpoint Mod2 boundary
rule) the corpus stands at **98.4% pass rate** (8806/8951 passing, 115
failures, 30 skipped). The remaining buckets are tracked here as known
divergences pending deeper engine work.

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
