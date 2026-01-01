package geojson

import (
	"encoding/json"
	"fmt"

	"github.com/robert-malhotra/go-topology-suite/geom"
)

// FeatureID represents a GeoJSON feature ID which can be a string or number.
type FeatureID struct {
	String  string
	Number  float64
	IsNum   bool
	IsValid bool
}

// NewStringID creates a string feature ID.
func NewStringID(s string) FeatureID {
	return FeatureID{String: s, IsValid: true}
}

// NewNumberID creates a numeric feature ID.
func NewNumberID(n float64) FeatureID {
	return FeatureID{Number: n, IsNum: true, IsValid: true}
}

// MarshalJSON implements json.Marshaler.
func (id FeatureID) MarshalJSON() ([]byte, error) {
	if !id.IsValid {
		return []byte("null"), nil
	}
	if id.IsNum {
		return json.Marshal(id.Number)
	}
	return json.Marshal(id.String)
}

// UnmarshalJSON implements json.Unmarshaler.
func (id *FeatureID) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		id.IsValid = false
		return nil
	}

	// Try string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		id.String = s
		id.IsValid = true
		return nil
	}

	// Try number
	var n float64
	if err := json.Unmarshal(data, &n); err == nil {
		id.Number = n
		id.IsNum = true
		id.IsValid = true
		return nil
	}

	return fmt.Errorf("feature id must be string or number")
}

// BBox represents a GeoJSON bounding box.
// Can be 4 elements (2D) or 6 elements (3D).
type BBox []float64

// Is2D returns true if this is a 2D bounding box.
func (b BBox) Is2D() bool {
	return len(b) == 4
}

// Is3D returns true if this is a 3D bounding box.
func (b BBox) Is3D() bool {
	return len(b) == 6
}

// ToEnvelope converts the BBox to a geom.Envelope.
func (b BBox) ToEnvelope() *geom.Envelope {
	if len(b) < 4 {
		return geom.NewEnvelopeEmpty()
	}
	return geom.NewEnvelope(b[0], b[1], b[2], b[3])
}

// NewBBox2D creates a 2D bounding box.
func NewBBox2D(minX, minY, maxX, maxY float64) BBox {
	return BBox{minX, minY, maxX, maxY}
}

// NewBBox3D creates a 3D bounding box.
func NewBBox3D(minX, minY, minZ, maxX, maxY, maxZ float64) BBox {
	return BBox{minX, minY, minZ, maxX, maxY, maxZ}
}

// BBoxFromEnvelope creates a BBox from a geom.Envelope.
func BBoxFromEnvelope(env *geom.Envelope) BBox {
	if env == nil || env.IsNull() {
		return nil
	}
	return NewBBox2D(env.MinX, env.MinY, env.MaxX, env.MaxY)
}

// ForeignMembers holds additional top-level members not defined by the GeoJSON spec.
type ForeignMembers map[string]any

// Geometry wraps a geom.Geometry for GeoJSON serialization.
type Geometry struct {
	geom.Geometry
}

// Feature represents a GeoJSON Feature with generic properties.
type Feature[P any] struct {
	ID             FeatureID      `json:"id,omitempty"`
	Geometry       *Geometry      `json:"geometry"`
	Properties     P              `json:"properties"`
	BBox           BBox           `json:"bbox,omitempty"`
	ForeignMembers ForeignMembers `json:"-"`
}

// NewFeature creates a new Feature with the given geometry and properties.
func NewFeature[P any](g geom.Geometry, props P) *Feature[P] {
	var geometry *Geometry
	if g != nil {
		geometry = &Geometry{Geometry: g}
	}
	return &Feature[P]{
		Geometry:   geometry,
		Properties: props,
	}
}

// NewFeatureWithID creates a new Feature with an ID.
func NewFeatureWithID[P any](id FeatureID, g geom.Geometry, props P) *Feature[P] {
	f := NewFeature(g, props)
	f.ID = id
	return f
}

// SetBBox sets the bounding box from the geometry's envelope.
func (f *Feature[P]) SetBBox() {
	if f.Geometry != nil && f.Geometry.Geometry != nil {
		f.BBox = BBoxFromEnvelope(f.Geometry.Envelope())
	}
}

// FeatureCollection represents a GeoJSON FeatureCollection with generic properties.
type FeatureCollection[P any] struct {
	Features       []*Feature[P]  `json:"features"`
	BBox           BBox           `json:"bbox,omitempty"`
	ForeignMembers ForeignMembers `json:"-"`
}

// NewFeatureCollection creates an empty FeatureCollection.
func NewFeatureCollection[P any]() *FeatureCollection[P] {
	return &FeatureCollection[P]{
		Features: make([]*Feature[P], 0),
	}
}

// Add appends a feature to the collection.
func (fc *FeatureCollection[P]) Add(f *Feature[P]) {
	fc.Features = append(fc.Features, f)
}

// AddGeometry creates a feature from a geometry and adds it.
func (fc *FeatureCollection[P]) AddGeometry(g geom.Geometry, props P) {
	fc.Add(NewFeature(g, props))
}

// Len returns the number of features.
func (fc *FeatureCollection[P]) Len() int {
	return len(fc.Features)
}

// SetBBox computes and sets the bounding box from all features.
func (fc *FeatureCollection[P]) SetBBox() {
	if len(fc.Features) == 0 {
		return
	}

	env := geom.NewEnvelopeEmpty()
	for _, f := range fc.Features {
		if f.Geometry != nil && f.Geometry.Geometry != nil {
			env.ExpandToInclude(f.Geometry.Envelope())
		}
	}

	if !env.IsNull() {
		fc.BBox = BBoxFromEnvelope(env)
	}
}

// Geometries returns all geometries from the features.
func (fc *FeatureCollection[P]) Geometries() []geom.Geometry {
	geoms := make([]geom.Geometry, 0, len(fc.Features))
	for _, f := range fc.Features {
		if f.Geometry != nil && f.Geometry.Geometry != nil {
			geoms = append(geoms, f.Geometry.Geometry)
		}
	}
	return geoms
}

// UntypedFeature is a Feature with map[string]any properties for dynamic use.
type UntypedFeature = Feature[map[string]any]

// UntypedFeatureCollection is a FeatureCollection with map[string]any properties.
type UntypedFeatureCollection = FeatureCollection[map[string]any]

// NewUntypedFeature creates a feature with untyped properties.
func NewUntypedFeature(g geom.Geometry, props map[string]any) *UntypedFeature {
	if props == nil {
		props = make(map[string]any)
	}
	return NewFeature(g, props)
}

// NewUntypedFeatureCollection creates a collection with untyped properties.
func NewUntypedFeatureCollection() *UntypedFeatureCollection {
	return NewFeatureCollection[map[string]any]()
}
