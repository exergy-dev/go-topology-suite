// Package overlayng implements a JTS overlay-NG-style polygon overlay
// for Terra. It replaces the v0.1 Greiner-Hormann implementation in
// cases where GH degenerates — coincident edges, shared boundary
// vertices, axis-aligned inputs.
//
// Algorithm sketch:
//
//  1. Compute noded edges from both rings — every edge crossing or
//     coincidence becomes a vertex shared by both polygons' edge sets.
//  2. Build a planar subdivision (half-edge DCEL): each undirected edge
//     produces two oppositely-directed half-edges; at each vertex the
//     outgoing half-edges are angularly sorted; "next" pointers traverse
//     each face counter-clockwise.
//  3. Trace all faces.
//  4. Classify each face by ray-casting an interior point against the
//     ORIGINAL subj and clip polygons. This sidesteps complex label
//     propagation and gives correct answers regardless of how edges are
//     shared between inputs.
//  5. For the chosen operation, mark each face keep/drop:
//       - Intersection: keep iff inSubj && inClip
//       - Union:        keep iff inSubj || inClip
//       - Difference:   keep iff inSubj && !inClip
//  6. Output the boundary: edges separating a kept face from a not-kept
//     face (or the outer face). Trace those edges into rings.
//
// v1.0 status: this is the foundation for the full JTS overlay-NG port.
// Currently it handles shells (no holes) and produces correct results
// for the cases v0.1 Greiner-Hormann fails on (axis-aligned coincident
// edges, shared boundaries). Hole support, full DE-9IM derivation, and
// MULTIPOLYGON↔MULTIPOLYGON via this path remain future work — the
// public Overlay function returns ErrUnsupportedKernel for those inputs
// and the caller (overlay package) falls back to GH where appropriate.
package overlayng
