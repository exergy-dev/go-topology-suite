// Package kernel defines the strategy interface that supplies geometric
// primitives — orientation, segment intersection, distance, point-in-ring,
// and so on — to every higher-level operation.
//
// There are exactly three implementations, each in its own subpackage:
//
//   - kernel/planar    Cartesian; the textbook 2D plane.
//   - kernel/spherical Treats Earth as a sphere; lon/lat input.
//   - kernel/geodesic  WGS84 ellipsoid; Karney's algorithm.
//
// Operations select the kernel via a functional Option (predicate.WithKernel
// etc.); when omitted the default is inferred from the geometry's CRS.
//
// Phase 0 of the implementation plan ships only this interface and the
// Orientation/Containment enums. Concrete kernels land in Phase 1.
package kernel
