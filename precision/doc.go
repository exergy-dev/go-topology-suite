// Package precision ports org.locationtech.jts.precision.
//
// The package gathers the JTS utilities that operate at, or
// transform between, fixed-resolution precision models:
//
//   - MinimumClearance / SimpleMinimumClearance — how much vertex
//     perturbation a geometry can absorb before becoming topologically
//     invalid (port of MinimumClearance / SimpleMinimumClearance).
//   - PrecisionReducer / Reduce / ReducePointwise — snap coordinates
//     to a target PrecisionModel grid (GeometryPrecisionReducer).
//   - SnapTo / SnapToSelf / SnapBoth — vertex/segment snapping by a
//     tolerance (GeometrySnapper).
//   - CommonBits / CommonBitsRemover / CommonBitsOp — translate
//     near-collocated operands toward the origin to recover
//     floating-point precision before an overlay (CommonBitsOp).
//   - ComputeOverlaySnapTolerance — the heuristic JTS uses to pick a
//     snap tolerance for overlay inputs.
package precision
