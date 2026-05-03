// Package locate ports org.locationtech.jts.algorithm.locate.
//
// It provides point-in-area locators that classify a point's relation to
// an areal Geometry as INTERIOR, BOUNDARY, or EXTERIOR, mirroring the
// JTS Location enum.
package locate

import "github.com/terra-geo/terra/geom"

// Location classifies a point's relationship to a Geometry.
//
// Mirrors org.locationtech.jts.geom.Location (we deliberately don't add
// it to geom because the kernel package already exposes a Containment
// enum used by point-in-ring; Location here is the JTS-style trichotomy
// returned by point-on-geometry locators).
type Location int8

const (
	// Exterior — the point lies outside the geometry.
	Exterior Location = iota
	// Boundary — the point lies exactly on the geometry boundary.
	Boundary
	// Interior — the point lies in the geometry interior.
	Interior
)

// String returns "EXTERIOR", "BOUNDARY", or "INTERIOR".
func (l Location) String() string {
	switch l {
	case Interior:
		return "INTERIOR"
	case Boundary:
		return "BOUNDARY"
	default:
		return "EXTERIOR"
	}
}

// PointOnGeometryLocator is implemented by classes that determine the
// Location of points in a Geometry. Mirrors JTS
// org.locationtech.jts.algorithm.locate.PointOnGeometryLocator.
type PointOnGeometryLocator interface {
	Locate(p geom.XY) Location
}
