package topology

import (
	"fmt"

	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/robert-malhotra/go-topology-suite/io/geojson"
	"github.com/robert-malhotra/go-topology-suite/io/kml"
	"github.com/robert-malhotra/go-topology-suite/io/wkb"
	"github.com/robert-malhotra/go-topology-suite/io/wkt"
)

type WKTOptions = wkt.Options
type WKBOptions = wkb.Options
type KMLOptions = kml.Options

type GeoJSONOptions struct {
	Indent string
}

type ReadOptions struct {
	Factory *geom.GeometryFactory
}

func ReadWKT(input string, opts ...ReadOptions) (geom.Geometry, error) {
	cfg := readOptions(opts...)
	return wkt.UnmarshalStringWithFactory(input, cfg.Factory)
}

func WriteWKT(g geom.Geometry, opts ...wkt.Options) (string, error) {
	if g == nil {
		return "", fmt.Errorf("v2 wkt: geometry is nil")
	}
	if len(opts) == 0 {
		return wkt.MarshalString(g), nil
	}
	return wkt.MarshalStringWithOptions(g, opts[0]), nil
}

func ReadWKB(data []byte, opts ...ReadOptions) (geom.Geometry, error) {
	cfg := readOptions(opts...)
	return wkb.UnmarshalWithFactory(data, cfg.Factory)
}

func WriteWKB(g geom.Geometry, opts ...wkb.Options) ([]byte, error) {
	if len(opts) == 0 {
		return wkb.Marshal(g)
	}
	return wkb.MarshalWithOptions(g, opts[0])
}

func ReadGeoJSON(data []byte, opts ...ReadOptions) (geom.Geometry, error) {
	cfg := readOptionsWithDefaultFactory(geojson.DefaultFactory, opts...)
	return geojson.UnmarshalGeometryWithFactory(data, cfg.Factory)
}

func WriteGeoJSON(g geom.Geometry, opts ...GeoJSONOptions) ([]byte, error) {
	if g == nil {
		return nil, fmt.Errorf("v2 geojson: geometry is nil")
	}
	if len(opts) > 0 && opts[0].Indent != "" {
		return geojson.MarshalGeometryIndent(g, opts[0].Indent)
	}
	return geojson.MarshalGeometry(g)
}

func ReadKML(data []byte, opts ...ReadOptions) (geom.Geometry, error) {
	cfg := readOptionsWithDefaultFactory(kml.DefaultFactory, opts...)
	return kml.UnmarshalWithFactory(data, cfg.Factory)
}

func WriteKML(g geom.Geometry, opts ...kml.Options) ([]byte, error) {
	if g == nil {
		return nil, fmt.Errorf("v2 kml: geometry is nil")
	}
	if len(opts) == 0 {
		return kml.Marshal(g)
	}
	return kml.MarshalWithOptions(g, opts[0])
}

func readOptions(opts ...ReadOptions) ReadOptions {
	return readOptionsWithDefaultFactory(geom.DefaultFactory, opts...)
}

func readOptionsWithDefaultFactory(defaultFactory *geom.GeometryFactory, opts ...ReadOptions) ReadOptions {
	cfg := ReadOptions{Factory: defaultFactory}
	if len(opts) > 0 {
		cfg = opts[0]
	}
	if cfg.Factory == nil {
		cfg.Factory = defaultFactory
	}
	return cfg
}
