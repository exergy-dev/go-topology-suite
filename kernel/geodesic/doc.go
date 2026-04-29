// Package geodesic implements the kernel.Kernel interface on the WGS84
// ellipsoid.
//
// Convention matches kernel/spherical: geom.XY.X is longitude in degrees,
// geom.XY.Y is latitude in degrees. All distances are returned in metres.
//
// Distance, initial bearing, and the midpoint use Vincenty's inverse
// formula as the fast path. For pairs where Vincenty's iteration fails
// to converge (typically within ~0.5° of antipodal), the implementation
// falls through to a port of Karney's exact inverse (Karney 2013),
// which converges everywhere on the ellipsoid. The destination point
// uses Vincenty's direct formula.
//
// Ring area uses Karney's exact ellipsoidal-polygon algorithm
// (Karney 2013, Section 6) with the series expansion truncated at
// eccentricity^6 — sub-millimetre accuracy on continent-scale rings.
//
// Topology primitives (Orient, PointInRing, AngleBetween) are independent
// of the surface model — they depend only on the chirality of points on
// a smooth surface — so this kernel delegates them to kernel/spherical.
// SegmentIntersection and SegmentDistance also delegate to the spherical
// kernel; geodesic-aware overlay is a future-phase deliverable.
package geodesic
