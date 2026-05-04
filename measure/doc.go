// Package measure computes scalar measurements (distance, length, area,
// centroid) on go-topology-suite geometries.
//
// Every measurement is kernel-routed. When WithKernel is not provided
// the kernel is chosen from the geometry's CRS:
//
//   - Geographic CRS → geodesic kernel (metres on WGS84).
//   - Projected CRS or no CRS → planar kernel (units of the projection).
//
// Pass WithKernel to override.
//
// Caveat: Centroid on Polygon and MultiPolygon uses the planar
// Bashein–Detmer shoelace formula in input coordinates; the kernel
// influences only the relative weights between sub-geometries (and the
// length-weighting of line centroids). A fully geodesic polygon
// centroid is future work.
package measure
