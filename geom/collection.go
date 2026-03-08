package geom

import (
	"sort"
	"strings"
)

// GeometryCollection is a collection of arbitrary geometry types.
type GeometryCollection struct {
	baseGeometry
	geometries []Geometry
}

// NewGeometryCollection creates a new GeometryCollection from geometries.
func NewGeometryCollection(geometries []Geometry) *GeometryCollection {
	gc := &GeometryCollection{
		geometries: make([]Geometry, len(geometries)),
	}
	for i, g := range geometries {
		gc.geometries[i] = g.Clone()
	}
	return gc
}

// NewGeometryCollectionEmpty creates an empty GeometryCollection.
func NewGeometryCollectionEmpty() *GeometryCollection {
	return &GeometryCollection{geometries: []Geometry{}}
}

// GeometryType returns "GeometryCollection".
func (gc *GeometryCollection) GeometryType() string {
	return "GeometryCollection"
}

// Envelope returns the bounding box.
func (gc *GeometryCollection) Envelope() *Envelope {
	if env := gc.cachedEnvelope(); env != nil {
		return env.Clone()
	}
	env := NewEnvelopeEmpty()
	for _, g := range gc.geometries {
		env.ExpandToInclude(g.Envelope())
	}
	gc.setCachedEnvelope(env)
	return env.Clone()
}

// IsEmpty returns true if there are no geometries.
func (gc *GeometryCollection) IsEmpty() bool {
	return len(gc.geometries) == 0
}

// IsSimple returns true if all component geometries are simple.
func (gc *GeometryCollection) IsSimple() bool {
	for _, g := range gc.geometries {
		if !g.IsSimple() {
			return false
		}
	}
	return true
}

// IsValid returns true if all component geometries are valid.
func (gc *GeometryCollection) IsValid() bool {
	for _, g := range gc.geometries {
		if !g.IsValid() {
			return false
		}
	}
	return true
}

// Dimension returns the maximum dimension of all components.
func (gc *GeometryCollection) Dimension() Dimension {
	maxDim := DimensionEmpty
	for _, g := range gc.geometries {
		dim := g.Dimension()
		if dim > maxDim {
			maxDim = dim
		}
	}
	return maxDim
}

// Boundary returns the union of boundaries of all components.
func (gc *GeometryCollection) Boundary() Geometry {
	var boundaries []Geometry
	for _, g := range gc.geometries {
		b := g.Boundary()
		if !b.IsEmpty() {
			boundaries = append(boundaries, b)
		}
	}
	return NewGeometryCollection(boundaries)
}

// Coordinates returns all coordinates from all geometries.
func (gc *GeometryCollection) Coordinates() CoordinateSequence {
	var coords CoordinateSequence
	for _, g := range gc.geometries {
		coords = append(coords, g.Coordinates()...)
	}
	return coords
}

// ApplyCoordinateFilter applies a coordinate filter to the collection.
func (gc *GeometryCollection) ApplyCoordinateFilter(filter CoordinateFilter) {
	if filter == nil {
		return
	}
	for _, g := range gc.geometries {
		if cf, ok := g.(CoordinateFilterer); ok {
			cf.ApplyCoordinateFilter(filter)
		}
	}
	gc.invalidateEnvelope()
}

// NumGeometries returns the number of geometries.
func (gc *GeometryCollection) NumGeometries() int {
	return len(gc.geometries)
}

// GeometryN returns the nth geometry (0-indexed).
func (gc *GeometryCollection) GeometryN(n int) Geometry {
	if n < 0 || n >= len(gc.geometries) {
		return nil
	}
	return gc.geometries[n]
}

// Clone returns a deep copy.
func (gc *GeometryCollection) Clone() Geometry {
	clone := NewGeometryCollection(gc.geometries)
	clone.srid = gc.srid
	return clone
}

// Normalized returns a new GeometryCollection with all components normalized.
func (gc *GeometryCollection) Normalized() Geometry {
	clone := gc.Clone().(*GeometryCollection)
	for i, g := range clone.geometries {
		clone.geometries[i] = g.Normalized()
	}
	sort.Slice(clone.geometries, func(i, j int) bool {
		return Compare(clone.geometries[i], clone.geometries[j]) < 0
	})
	return clone
}

// EqualsExact returns true if the GeometryCollections are exactly equal.
func (gc *GeometryCollection) EqualsExact(other Geometry, tolerance float64) bool {
	if other == nil {
		return false
	}
	otherGC, ok := other.(*GeometryCollection)
	if !ok {
		return false
	}
	if len(gc.geometries) != len(otherGC.geometries) {
		return false
	}
	for i, g := range gc.geometries {
		if !g.EqualsExact(otherGC.geometries[i], tolerance) {
			return false
		}
	}
	return true
}

// String returns the WKT representation.
func (gc *GeometryCollection) String() string {
	if gc.IsEmpty() {
		return "GEOMETRYCOLLECTION EMPTY"
	}

	var sb strings.Builder
	sb.WriteString("GEOMETRYCOLLECTION (")
	for i, g := range gc.geometries {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(g.String())
	}
	sb.WriteString(")")
	return sb.String()
}

