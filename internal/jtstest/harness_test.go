//go:build jts

package jtstest

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJTSConformance walks the embedded testdata corpus and runs every
// op against terra. The harness logs aggregate counts and reports
// per-failure detail via assert.Failf.
func TestJTSConformance(t *testing.T) {
	files, err := findCorpus("testdata")
	require.NoError(t, err, "walk testdata")
	require.NotEmpty(t, files, "no XML test files found under testdata/")

	var (
		total   int
		passed  int
		failed  int
		skipped int
	)

	for _, path := range files {
		rn, err := loadFile(path)
		if !assert.NoErrorf(t, err, "%s: load", path) {
			continue
		}
		rel, _ := filepath.Rel("testdata", path)
		for ci, c := range rn.Cases {
			for ti, tc := range c.Tests {
				total++
				res := runOp(&c, tc.Op)
				label := makeLabel(rel, ci, ti, c.Desc, tc.Desc, tc.Op.Name)
				switch {
				case res.Skipped:
					skipped++
					t.Logf("SKIP %s: %s", label, res.Reason)
				case res.Pass:
					passed++
				default:
					failed++
					assert.Failf(t, "FAIL "+label, "%s", res.Detail)
				}
			}
		}
	}

	t.Logf("JTS conformance: total=%d passed=%d failed=%d skipped=%d",
		total, passed, failed, skipped)
	if total > 0 {
		t.Logf("pass rate: %.1f%% (excluding skipped: %.1f%%)",
			100*float64(passed)/float64(total),
			percentExclSkip(passed, total, skipped))
	}
}

func makeLabel(file string, caseIdx, testIdx int, caseDesc, testDesc, op string) string {
	desc := testDesc
	if desc == "" {
		desc = caseDesc
	}
	return file + " case#" + itoa(caseIdx) + " test#" + itoa(testIdx) +
		" op=" + op + " — " + desc
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}

func percentExclSkip(passed, total, skipped int) float64 {
	denom := total - skipped
	if denom <= 0 {
		return 0
	}
	return 100 * float64(passed) / float64(denom)
}
