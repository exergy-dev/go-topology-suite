package proj

import (
	"math"
	"testing"

	"github.com/exergy-dev/go-topology-suite/crs"
)

func BenchmarkUTMForward(b *testing.B) {
	p := UTM(33, false, crs.WGS84Ellipsoid)
	d2r := math.Pi / 180.0
	lon, lat := 12*d2r, 50*d2r
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Forward(lon, lat)
	}
}

func BenchmarkUTMInverse(b *testing.B) {
	p := UTM(33, false, crs.WGS84Ellipsoid)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Inverse(285015.5, 5542944.0)
	}
}
