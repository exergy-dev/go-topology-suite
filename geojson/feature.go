package geojson

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/terra-geo/terra/geom"
)

// Feature is the GeoJSON Feature type.
//
// Foreign top-level members are stored verbatim and round-tripped through
// MarshalJSON/UnmarshalJSON. ID may be string, number, or null per RFC 7946.
type Feature struct {
	Geometry   geom.Geometry
	Properties map[string]any
	ID         any
	BBox       *geom.Envelope
	// Foreign holds top-level members not part of the GeoJSON spec
	// (e.g. "title", "renderer"). Keys overlapping the spec are silently
	// dropped on marshal.
	Foreign map[string]json.RawMessage
}

// FeatureCollection is the GeoJSON FeatureCollection type.
type FeatureCollection struct {
	Features []*Feature
	BBox     *geom.Envelope
	Foreign  map[string]json.RawMessage
}

// MarshalJSON encodes a Feature.
func (f *Feature) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer
	b.WriteString(`{"type":"Feature"`)
	if f.ID != nil {
		idJSON, err := json.Marshal(f.ID)
		if err != nil {
			return nil, fmt.Errorf("geojson: bad id: %w", err)
		}
		b.WriteString(`,"id":`)
		b.Write(idJSON)
	}
	if f.BBox != nil {
		b.WriteString(`,"bbox":`)
		writeBBox(&b, *f.BBox)
	}
	b.WriteString(`,"geometry":`)
	if f.Geometry == nil {
		b.WriteString("null")
	} else {
		geomJSON, err := Marshal(f.Geometry)
		if err != nil {
			return nil, err
		}
		b.Write(geomJSON)
	}
	b.WriteString(`,"properties":`)
	if f.Properties == nil {
		b.WriteString("null")
	} else {
		propJSON, err := json.Marshal(f.Properties)
		if err != nil {
			return nil, fmt.Errorf("geojson: properties: %w", err)
		}
		b.Write(propJSON)
	}
	for k, v := range f.Foreign {
		if isReservedFeatureKey(k) {
			continue
		}
		b.WriteString(`,`)
		k2, _ := json.Marshal(k)
		b.Write(k2)
		b.WriteByte(':')
		b.WriteString(rawJSONOrNull(v))
	}
	b.WriteByte('}')
	return b.Bytes(), nil
}

func isReservedFeatureKey(k string) bool {
	switch k {
	case "type", "id", "bbox", "geometry", "properties":
		return true
	}
	return false
}

func writeBBox(b *bytes.Buffer, env geom.Envelope) {
	bb, _ := json.Marshal([]float64{env.MinX, env.MinY, env.MaxX, env.MaxY})
	b.Write(bb)
}

// UnmarshalJSON decodes a Feature.
func (f *Feature) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("geojson: %w", err)
	}
	if t, ok := raw["type"]; ok {
		var typ string
		_ = json.Unmarshal(t, &typ)
		if typ != "Feature" {
			return fmt.Errorf("geojson: expected type Feature, got %q", typ)
		}
	}
	if g, ok := raw["geometry"]; ok && string(g) != "null" {
		geo, err := Unmarshal(g)
		if err != nil {
			return err
		}
		f.Geometry = geo
	}
	if p, ok := raw["properties"]; ok && string(p) != "null" {
		var props map[string]any
		if err := json.Unmarshal(p, &props); err != nil {
			return err
		}
		f.Properties = props
	}
	if id, ok := raw["id"]; ok {
		var v any
		_ = json.Unmarshal(id, &v)
		f.ID = v
	}
	if bb, ok := raw["bbox"]; ok {
		env, err := decodeBBox(bb)
		if err != nil {
			return err
		}
		f.BBox = env
	}
	// Foreign: everything else.
	for k, v := range raw {
		if isReservedFeatureKey(k) {
			continue
		}
		if f.Foreign == nil {
			f.Foreign = make(map[string]json.RawMessage)
		}
		f.Foreign[k] = v
	}
	return nil
}

func decodeBBox(raw json.RawMessage) (*geom.Envelope, error) {
	var arr []float64
	if err := json.Unmarshal(raw, &arr); err != nil {
		return nil, err
	}
	switch len(arr) {
	case 4:
		return &geom.Envelope{MinX: arr[0], MinY: arr[1], MaxX: arr[2], MaxY: arr[3]}, nil
	case 6:
		// 3D bbox: drop Z bounds.
		return &geom.Envelope{MinX: arr[0], MinY: arr[1], MaxX: arr[3], MaxY: arr[4]}, nil
	}
	return nil, fmt.Errorf("geojson: bbox length %d unsupported", len(arr))
}

// MarshalJSON encodes a FeatureCollection.
func (fc *FeatureCollection) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer
	b.WriteString(`{"type":"FeatureCollection"`)
	if fc.BBox != nil {
		b.WriteString(`,"bbox":`)
		writeBBox(&b, *fc.BBox)
	}
	b.WriteString(`,"features":[`)
	for i, f := range fc.Features {
		if i > 0 {
			b.WriteByte(',')
		}
		fb, err := f.MarshalJSON()
		if err != nil {
			return nil, err
		}
		b.Write(fb)
	}
	b.WriteByte(']')
	for k, v := range fc.Foreign {
		if isReservedCollectionKey(k) {
			continue
		}
		b.WriteByte(',')
		k2, _ := json.Marshal(k)
		b.Write(k2)
		b.WriteByte(':')
		b.WriteString(rawJSONOrNull(v))
	}
	b.WriteByte('}')
	return b.Bytes(), nil
}

func isReservedCollectionKey(k string) bool {
	switch k {
	case "type", "bbox", "features":
		return true
	}
	return false
}

// UnmarshalJSON decodes a FeatureCollection.
func (fc *FeatureCollection) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if t, ok := raw["type"]; ok {
		var typ string
		_ = json.Unmarshal(t, &typ)
		if typ != "FeatureCollection" {
			return fmt.Errorf("geojson: expected type FeatureCollection, got %q", typ)
		}
	}
	if features, ok := raw["features"]; ok {
		var arr []json.RawMessage
		if err := json.Unmarshal(features, &arr); err != nil {
			return err
		}
		fc.Features = make([]*Feature, 0, len(arr))
		for _, fr := range arr {
			f := &Feature{}
			if err := f.UnmarshalJSON(fr); err != nil {
				return err
			}
			fc.Features = append(fc.Features, f)
		}
	}
	if bb, ok := raw["bbox"]; ok {
		env, err := decodeBBox(bb)
		if err != nil {
			return err
		}
		fc.BBox = env
	}
	for k, v := range raw {
		if isReservedCollectionKey(k) {
			continue
		}
		if fc.Foreign == nil {
			fc.Foreign = make(map[string]json.RawMessage)
		}
		fc.Foreign[k] = v
	}
	return nil
}
