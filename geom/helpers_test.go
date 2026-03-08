package geom_test

import "github.com/robert-malhotra/go-topology-suite/geom"

func mustCoordsXY(values ...float64) geom.CoordinateSequence {
	seq, err := geom.NewCoordinateSequenceXY(values...)
	if err != nil {
		panic(err)
	}
	return seq
}

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
