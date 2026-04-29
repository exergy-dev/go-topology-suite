// Package hull computes convex (and eventually concave) hulls of Terra
// geometries. The convex-hull implementation is Andrew's monotone-chain
// algorithm — O(n log n) and stable on collinear inputs.
package hull
