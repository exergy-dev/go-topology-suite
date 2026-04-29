// Package planar implements the kernel.Kernel interface in Cartesian 2D
// space. All distances, areas, and bearings are computed as if the
// coordinates lie on a flat plane.
//
// This is the right kernel for projected CRSes (UTM, Web Mercator, state
// plane, etc.). For lon/lat coordinates, prefer kernel/spherical or
// kernel/geodesic — using the planar kernel on geographic CRSes silently
// produces wrong answers near the poles and across the antimeridian.
package planar
