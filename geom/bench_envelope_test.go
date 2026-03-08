package geom_test

import (
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
)

func BenchmarkPointEnvelope(b *testing.B) {
	p := geom.NewPoint(5, 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.Envelope()
	}
}

func BenchmarkLineStringEnvelope(b *testing.B) {
	ls := geom.NewLineString(geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(10, 10),
		geom.NewCoordinate(20, 5),
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ls.Envelope()
	}
}

func BenchmarkPolygonEnvelope(b *testing.B) {
	poly := geom.NewPolygon(
		geom.NewLinearRing(geom.CoordinateSequence{
			geom.NewCoordinate(0, 0),
			geom.NewCoordinate(10, 0),
			geom.NewCoordinate(10, 10),
			geom.NewCoordinate(0, 10),
			geom.NewCoordinate(0, 0),
		}), nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = poly.Envelope()
	}
}

func BenchmarkMultiPointEnvelope(b *testing.B) {
	mp := geom.NewMultiPoint([]*geom.Point{
		geom.NewPoint(0, 0),
		geom.NewPoint(10, 10),
		geom.NewPoint(5, 5),
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mp.Envelope()
	}
}

func BenchmarkGeometryCollectionEnvelope(b *testing.B) {
	gc := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(0, 0),
		geom.NewPoint(100, 100),
		geom.NewLineString(geom.CoordinateSequence{
			geom.NewCoordinate(50, 50),
			geom.NewCoordinate(75, 75),
		}),
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gc.Envelope()
	}
}
