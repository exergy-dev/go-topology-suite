package kml

import (
	"encoding/xml"

	"github.com/robert-malhotra/go-topology-suite/geom"
)

// AltitudeMode specifies how altitude values are interpreted.
type AltitudeMode string

const (
	// AltitudeModeClampToGround indicates altitude values are ignored and the geometry
	// is clamped to the terrain surface.
	AltitudeModeClampToGround AltitudeMode = "clampToGround"

	// AltitudeModeRelativeToGround indicates altitude values are relative to the
	// ground elevation at each point.
	AltitudeModeRelativeToGround AltitudeMode = "relativeToGround"

	// AltitudeModeAbsolute indicates altitude values are absolute (relative to sea level).
	AltitudeModeAbsolute AltitudeMode = "absolute"
)

// KML is the root element of a KML document.
type KML struct {
	XMLName   xml.Name   `xml:"kml"`
	Namespace string     `xml:"xmlns,attr,omitempty"`
	Document  *Document  `xml:"Document,omitempty"`
	Folder    *Folder    `xml:"Folder,omitempty"`
	Placemark *Placemark `xml:"Placemark,omitempty"`
}

// Document is a container for features and styles.
type Document struct {
	XMLName    xml.Name     `xml:"Document"`
	Name       string       `xml:"name,omitempty"`
	Placemarks []*Placemark `xml:"Placemark,omitempty"`
	Folders    []*Folder    `xml:"Folder,omitempty"`
}

// Folder is a container for organizing features.
type Folder struct {
	XMLName    xml.Name     `xml:"Folder"`
	Name       string       `xml:"name,omitempty"`
	Placemarks []*Placemark `xml:"Placemark,omitempty"`
	Folders    []*Folder    `xml:"Folder,omitempty"`
}

// Placemark is a feature with geometry.
type Placemark struct {
	XMLName       xml.Name       `xml:"Placemark"`
	ID            string         `xml:"id,attr,omitempty"`
	Name          string         `xml:"name,omitempty"`
	Description   string         `xml:"description,omitempty"`
	Point         *Point         `xml:"Point,omitempty"`
	LineString    *LineString    `xml:"LineString,omitempty"`
	LinearRing    *LinearRing    `xml:"LinearRing,omitempty"`
	Polygon       *Polygon       `xml:"Polygon,omitempty"`
	MultiGeometry *MultiGeometry `xml:"MultiGeometry,omitempty"`
}

// Point represents a KML Point geometry.
type Point struct {
	XMLName      xml.Name     `xml:"Point"`
	Extrude      bool         `xml:"extrude,omitempty"`
	AltitudeMode AltitudeMode `xml:"altitudeMode,omitempty"`
	Coordinates  string       `xml:"coordinates"`
}

// LineString represents a KML LineString geometry.
type LineString struct {
	XMLName      xml.Name     `xml:"LineString"`
	Extrude      bool         `xml:"extrude,omitempty"`
	Tessellate   bool         `xml:"tessellate,omitempty"`
	AltitudeMode AltitudeMode `xml:"altitudeMode,omitempty"`
	Coordinates  string       `xml:"coordinates"`
}

// LinearRing represents a KML LinearRing geometry (closed ring for polygon boundaries).
type LinearRing struct {
	XMLName      xml.Name     `xml:"LinearRing"`
	Extrude      bool         `xml:"extrude,omitempty"`
	Tessellate   bool         `xml:"tessellate,omitempty"`
	AltitudeMode AltitudeMode `xml:"altitudeMode,omitempty"`
	Coordinates  string       `xml:"coordinates"`
}

// Polygon represents a KML Polygon geometry.
type Polygon struct {
	XMLName        xml.Name         `xml:"Polygon"`
	Extrude        bool             `xml:"extrude,omitempty"`
	Tessellate     bool             `xml:"tessellate,omitempty"`
	AltitudeMode   AltitudeMode     `xml:"altitudeMode,omitempty"`
	OuterBoundary  *OuterBoundaryIs `xml:"outerBoundaryIs"`
	InnerBoundries []*InnerBoundaryIs `xml:"innerBoundaryIs,omitempty"`
}

// OuterBoundaryIs contains the outer boundary of a polygon.
type OuterBoundaryIs struct {
	XMLName    xml.Name    `xml:"outerBoundaryIs"`
	LinearRing *LinearRing `xml:"LinearRing"`
}

// InnerBoundaryIs contains an inner boundary (hole) of a polygon.
type InnerBoundaryIs struct {
	XMLName    xml.Name    `xml:"innerBoundaryIs"`
	LinearRing *LinearRing `xml:"LinearRing"`
}

// MultiGeometry represents a collection of geometries.
type MultiGeometry struct {
	XMLName        xml.Name         `xml:"MultiGeometry"`
	Points         []*Point         `xml:"Point,omitempty"`
	LineStrings    []*LineString    `xml:"LineString,omitempty"`
	LinearRings    []*LinearRing    `xml:"LinearRing,omitempty"`
	Polygons       []*Polygon       `xml:"Polygon,omitempty"`
	MultiGeometries []*MultiGeometry `xml:"MultiGeometry,omitempty"`
}

// hasGeometry returns true if the Placemark contains any geometry.
func (p *Placemark) hasGeometry() bool {
	return p.Point != nil || p.LineString != nil || p.LinearRing != nil ||
		p.Polygon != nil || p.MultiGeometry != nil
}

// getGeometry returns the first non-nil geometry element.
func (p *Placemark) getGeometry() interface{} {
	if p.Point != nil {
		return p.Point
	}
	if p.LineString != nil {
		return p.LineString
	}
	if p.LinearRing != nil {
		return p.LinearRing
	}
	if p.Polygon != nil {
		return p.Polygon
	}
	if p.MultiGeometry != nil {
		return p.MultiGeometry
	}
	return nil
}

// Feature represents a KML feature with geometry and properties.
type Feature struct {
	ID          string
	Name        string
	Description string
	Geometry    geom.Geometry
}
