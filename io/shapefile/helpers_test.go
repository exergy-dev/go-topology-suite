package shapefile

import "github.com/robert-malhotra/go-topology-suite/geom"

func mustCreateLineStringXY(gf *geom.GeometryFactory, values ...float64) *geom.LineString {
	seq, err := geom.NewCoordinateSequenceXY(values...)
	if err != nil {
		panic(err)
	}
	return gf.CreateLineString(seq)
}

func mustCreateLinearRingXY(gf *geom.GeometryFactory, values ...float64) *geom.LinearRing {
	seq, err := geom.NewCoordinateSequenceXY(values...)
	if err != nil {
		panic(err)
	}
	return gf.CreateLinearRing(seq)
}
