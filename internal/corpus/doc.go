// Package corpus exposes a small bundle of real-world-style GeoJSON
// FeatureCollections that ship with the Terra source tree, plus a smoke
// harness that runs the core pipeline (validate, measure, buffer, overlay)
// across every feature.
//
// # Role
//
// The corpus exists to catch regressions that unit tests on hand-crafted
// shapes miss: ring-orientation quirks, coordinate-density extremes,
// holes, multi-parts, and the kinds of awkward overlap geometries
// produced by real cartographic data. It is intentionally small - the
// goal is breadth of shape topology, not benchmark scale.
//
// # Embedding strategy
//
// All fixtures live under testdata/ and are pulled in via go:embed at
// build time. No filesystem or network access is required at test time,
// so the harness runs identically in CI, in offline environments, and
// inside go test -count=N loops. The total embedded payload is kept
// well under 200 KB.
//
// Three fixtures are provided:
//
//   - "ne"    : Natural-Earth-style country boundaries (~10 polygons,
//     including a polygon with a hole and a multi-polygon).
//   - "tiger" : TIGER-style county polygons (~10 rectangles, one with a
//     courtyard hole).
//   - "osm"   : OSM-style building footprints (~20 small polygons,
//     including one footprint with a courtyard hole).
//
// Shapes are synthesised - they are plausible in topology and density
// but do not correspond to real-world places. This keeps the embedded
// payload small and avoids licensing concerns.
//
// # Smoke harness
//
// TestCorpusSmoke (in smoke_test.go) iterates every fixture and every
// feature, exercising:
//
//   - validate.Validate         - must not error on any fixture geometry.
//   - measure.Area / Length     - results must be finite for every feature
//     of dimension >= 2 (Area) and dimension >= 1 (Length).
//   - buffer.Buffer with a small positive distance - must succeed and
//     produce a non-empty result for non-empty inputs.
//   - overlay.Union              - run on the first three feature pairs of
//     each fixture; errors are tolerated (logged, not failed) because
//     overlay is not yet expected to be robust against every real-world
//     polygon configuration.
//
// The harness is intended to complete in well under two seconds.
package corpus
