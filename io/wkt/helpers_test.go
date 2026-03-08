package wkt_test

import "github.com/robert-malhotra/go-topology-suite/geom"

func mustLineStringXY(values ...float64) *geom.LineString {
	seq, err := geom.NewCoordinateSequenceXY(values...)
	if err != nil {
		panic(err)
	}
	return geom.NewLineString(seq)
}

func mustLinearRingXY(values ...float64) *geom.LinearRing {
	seq, err := geom.NewCoordinateSequenceXY(values...)
	if err != nil {
		panic(err)
	}
	return geom.NewLinearRing(seq)
}
