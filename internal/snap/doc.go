// Package snap provides snap-rounding infrastructure used by the overlay-NG
// port (Pillars A1/A2 in the Terra parallel-implementation plan).
//
// # What snap rounding does
//
// Snap rounding is a numerical-robustness preprocessing pass. Given a set of
// input vertices, it rounds each vertex to the nearest point on a regular
// grid whose cell side equals the configured tolerance. Vertices that fall
// in the same grid cell collapse to a single point; consecutive duplicates
// in a ring are removed; and rings that collapse below four distinct
// vertices are reported as degenerate.
//
// The output is a set of segments where:
//   - Every endpoint sits exactly on the regular tolerance grid.
//   - No two distinct snapped vertices are closer than tolerance.
//   - Two near-coincident, near-parallel segments either become exactly
//     identical or remain exactly distinct after snapping — the
//     "near-equal" case (which kills topology graph algorithms with
//     floating-point comparisons) is removed.
//
// Snap rounding is the precondition that lets the JTS overlay-NG topology
// graph operate without ad-hoc tolerance hacks: every comparison after
// snap-round can be exact equality on the rounded grid.
//
// # When to use it
//
// Use snap rounding before feeding geometry to overlay, polygonization, or
// any other algorithm that constructs a topology graph from segment
// intersections. Do not use it when you need to preserve the exact input
// coordinates — snap-round is a lossy transform.
//
// A reasonable tolerance for unit-scale Cartesian inputs is 1e-9; for
// lon/lat inputs ~1e-7 (≈ 1 cm at the equator) is conservative.
//
// # v1.0 limitations
//
// This package implements grid-snap rounding, not the full Goodrich-Guibas-
// Hershberger-Tanenbaum (1997) "Snap Rounding Line Segments Efficiently in
// Two and Three Dimensions" algorithm.
//
// Specifically, the v1.0 implementation:
//   - Rounds every vertex to the nearest grid point and removes consecutive
//     duplicates.
//   - Does NOT detect or split a segment that passes through a hot pixel
//     belonging to some other input vertex without already terminating
//     there. A true Goodrich-style implementation would insert a vertex at
//     the hot-pixel center; we currently rely on the caller to have noded
//     intersections separately (see package noding, planned).
//
// This is correct for non-pathological inputs — real-world cadastral and
// hydrography data — but it can leave coincident-edge degeneracies if the
// input contains a segment that passes within tolerance of an unrelated
// vertex without crossing any segment incident to that vertex. The full
// algorithm will arrive in a later gate together with the noder port.
//
// The package is internal: only Terra packages can import it.
package snap
