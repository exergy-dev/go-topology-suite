package geojson

import (
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// checkSame asserts the scanner-based Unmarshal agrees with the legacy
// encoding/json based decoder (decode_legacy_test.go) on in: same
// success/failure, identical geometry, identical re-encoding.
func checkSame(t *testing.T, in string) {
	t.Helper()
	want, werr := legacyUnmarshal([]byte(in))
	got, gerr := Unmarshal([]byte(in))
	if (werr != nil) != (gerr != nil) {
		t.Fatalf("error mismatch for %q:\n legacy: %v\n new:    %v", in, werr, gerr)
	}
	if werr != nil {
		return
	}
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("geometry mismatch for %q:\n legacy: %#v\n new:    %#v", in, want, got)
	}
	// DeepEqual treats -0 == 0; the textual re-encoding distinguishes them.
	wb, err1 := Marshal(want)
	gb, err2 := Marshal(got)
	if err1 != nil || err2 != nil || string(wb) != string(gb) {
		t.Fatalf("marshal mismatch for %q: %s vs %s (%v %v)", in, wb, gb, err1, err2)
	}
}

func TestDecodeDifferentialCorpus(t *testing.T) {
	corpus := []string{
		`{"type":"Point","coordinates":[1,2]}`,
		`{"type":"Point","coordinates":[1,2,3]}`,
		`{"type":"Point","coordinates":[1,2,3,4]}`,
		`{"type":"Point","coordinates":[]}`,
		`{"type":"Point","coordinates":[ ]}`,
		`{"type":"Point","coordinates":[1]}`,
		`{"type":"Point","coordinates":null}`,
		`{"type":"Point","coordinates":[1,null]}`,
		`{"type":"Point","coordinates":["a",2]}`,
		`{"type":"Point","coordinates":[true,2]}`,
		`{"type":"Point","coordinates":1}`,
		`{"type":"Point","coordinates":[[1,2]]}`,
		`{"type":"Point","coordinates":[1e999,2]}`,
		`{"type":"Point","coordinates":[-0,-0.0]}`,
		`{"type":"Point","coordinates":[1E2,2e-2,3.5E+1]}`,
		"{\"type\":\"Point\",\"coordinates\":\n\t[ 1 ,\r\n 2 ]}",
		`{"type":"LineString","coordinates":[]}`,
		`{"type":"LineString","coordinates":null}`,
		`{"type":"LineString","coordinates":[[0,0],[1,1]]}`,
		`{"type":"LineString","coordinates":[[0,0,5],[1,1],[2,2,7,9]]}`,
		`{"type":"LineString","coordinates":[[0,0],[1,1,1],[2,2,7,9]]}`,
		`{"type":"LineString","coordinates":[[0,0],[1]]}`,
		`{"type":"LineString","coordinates":[[0,0],[]]}`,
		`{"type":"LineString","coordinates":[[0,0],null]}`,
		`{"type":"LineString","coordinates":[[0,0],1]}`,
		`{"type":"LineString","coordinates":[ [ 0 , 0 ] , [ 1 , 1 ] ]}`,
		`{"type":"LineString","coordinates":[[-0,0],[1e0,1E1]]}`,
		`{"type":"LineString","coordinates":[[0,0],[1,1]],"extra":{"a":[1,2]}}`,
		`{"coordinates":[[0,0],[1,1]],"type":"LineString"}`,
		`{"type":"Polygon","coordinates":[]}`,
		`{"type":"Polygon","coordinates":null}`,
		`{"type":"Polygon","coordinates":[[]]}`,
		`{"type":"Polygon","coordinates":[[],[[0,0,1],[1,0,1],[1,1,1],[0,0,1]]]}`,
		`{"type":"Polygon","coordinates":[[[0,0],[1,0],[1,1],[0,0]]]}`,
		`{"type":"Polygon","coordinates":[[[0,0,1],[1,0,1],[1,1,1],[0,0,1]]]}`,
		`{"type":"Polygon","coordinates":[[[0,0],[4,0],[4,4],[0,0]],[[1,1,9],[2,1,9],[2,2,9],[1,1,9]]]}`,
		`{"type":"Polygon","coordinates":[[[0,0,1],[4,0,1],[4,4,1],[0,0,1]],[[1,1],[2,1],[2,2],[1,1]]]}`,
		`{"type":"Polygon","coordinates":[[[0,0],[4,0],[4,4],[0,0]],[[1,1,9,8],[2,1],[2,2],[1,1]]]}`,
		`{"type":"Polygon","coordinates":[[[0,0,1],[4,0],[4,4,2,3],[0,0]],[[1,1],[2,1],[2,2],[1,1]]]}`,
		`{"type":"Polygon","coordinates":[[[0,0,1],[4,0,1],[4,4,1],[0,0,1]],[[1,1,5,6],[2,1,2],[2,2],[1,1]]]}`,
		`{"type":"Polygon","coordinates":[[[0,0],[1,0],[1,1],[0,0]],null,[]]}`,
		`{"type":"Polygon","coordinates":[[[0,0],[1,0],[1,1],[0,0]]` + "\n" + `]}`,
		`{"type":"Polygon","coordinates":[[[0,0],[1,0],[1,1],[0,0]],5]}`,
		`{"type":"MultiPoint","coordinates":[]}`,
		`{"type":"MultiPoint","coordinates":[[0,0],[1,1,1]]}`,
		`{"type":"MultiPoint","coordinates":[[0,0,3],[1,1]]}`,
		`{"type":"MultiPoint","coordinates":[[0]]}`,
		`{"type":"MultiLineString","coordinates":[]}`,
		`{"type":"MultiLineString","coordinates":null}`,
		`{"type":"MultiLineString","coordinates":[[],[[0,0],[1,1]],[[0,0,1],[1,1,1]],null]}`,
		`{"type":"MultiLineString","coordinates":[[[0,0],[1]]]}`,
		`{"type":"MultiPolygon","coordinates":[]}`,
		`{"type":"MultiPolygon","coordinates":null}`,
		`{"type":"MultiPolygon","coordinates":[[],null,[[]],[[[0,0],[1,0],[1,1],[0,0]]],[[[0,0,1],[1,0,1],[1,1,1],[0,0,1]],[[0,0],[1,0],[1,1],[0,0]]]]}`,
		`{"type":"MultiPolygon","coordinates":[[[[0,0],[1,0],[1,1],[0,0]]],[[[0,0],[1,x]]]]}`,
		`{"type":"GeometryCollection","geometries":[{"type":"Point","coordinates":[1,2]},{"type":"LineString","coordinates":[[0,0],[1,1,1]]}]}`,
		`{"type":"GeometryCollection","geometries":[]}`,
		`{"type":"Nope","coordinates":[]}`,
		`{"type":"Point"}`,
		`{"coordinates":[1,2]}`,
		`[1,2]`,
		``,
	}
	for _, in := range corpus {
		checkSame(t, in)
	}
}

func fnum(r *rand.Rand) string {
	switch r.Intn(8) {
	case 0:
		return "-0"
	case 1:
		return strconv.FormatFloat(r.NormFloat64()*1e6, 'e', -1, 64)
	case 2:
		return strconv.Itoa(r.Intn(2000) - 1000)
	case 3:
		return strconv.FormatFloat(r.Float64(), 'f', -1, 64)
	default:
		return strconv.FormatFloat(r.NormFloat64()*180, 'g', -1, 64)
	}
}

func fws(r *rand.Rand) string {
	if r.Intn(4) != 0 {
		return ""
	}
	return []string{" ", "\n", "\t ", "\r\n"}[r.Intn(4)]
}

func fvertex(r *rand.Rand, dims int) string {
	n := dims
	if r.Intn(10) == 0 {
		n = 2 + r.Intn(3)
	}
	parts := make([]string, n)
	for i := range parts {
		parts[i] = fws(r) + fnum(r) + fws(r)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func fline(r *rand.Rand, dims int) string {
	n := r.Intn(6)
	vs := make([]string, n)
	for i := range vs {
		vs[i] = fws(r) + fvertex(r, dims) + fws(r)
	}
	return "[" + strings.Join(vs, ",") + "]"
}

func fpoly(r *rand.Rand, dims int) string {
	n := r.Intn(4)
	rs := make([]string, n)
	for i := range rs {
		d := dims
		if r.Intn(4) == 0 {
			d = 2 + r.Intn(2)
		}
		rs[i] = fline(r, d)
	}
	return "[" + strings.Join(rs, ",") + "]"
}

func TestDecodeDifferentialRandom(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	for i := 0; i < 5000; i++ {
		dims := 2 + r.Intn(2)
		var in string
		switch r.Intn(6) {
		case 0:
			in = `{"type":"Point","coordinates":` + fvertex(r, dims) + `}`
		case 1:
			in = `{"type":"LineString","coordinates":` + fline(r, dims) + `}`
		case 2:
			in = `{"type":"Polygon","coordinates":` + fpoly(r, dims) + `}`
		case 3:
			in = `{"type":"MultiPoint","coordinates":` + fline(r, dims) + `}`
		case 4:
			ls := make([]string, r.Intn(4))
			for j := range ls {
				ls[j] = fline(r, dims)
			}
			in = `{"type":"MultiLineString","coordinates":[` + strings.Join(ls, ",") + `]}`
		default:
			ps := make([]string, r.Intn(4))
			for j := range ps {
				ps[j] = fpoly(r, dims)
			}
			in = `{"type":"MultiPolygon","coordinates":[` + strings.Join(ps, ",") + `]}`
		}
		checkSame(t, in)
	}
}
