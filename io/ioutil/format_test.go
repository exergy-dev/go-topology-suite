package ioutil_test

import (
	"math"
	"strings"
	"testing"

	"github.com/robert-malhotra/go-topology-suite/io/ioutil"
	"github.com/stretchr/testify/assert"
)

func TestWriteNumber_DefaultPrecision(t *testing.T) {
	var sb strings.Builder
	ioutil.WriteNumber(&sb, 3.14159, -1)
	assert.Equal(t, "3.14159", sb.String())
}

func TestWriteNumber_FixedPrecision(t *testing.T) {
	var sb strings.Builder
	ioutil.WriteNumber(&sb, 3.14159, 2)
	assert.Equal(t, "3.14", sb.String())
}

func TestWriteNumber_ZeroPrecision(t *testing.T) {
	var sb strings.Builder
	ioutil.WriteNumber(&sb, 3.7, 0)
	assert.Equal(t, "4", sb.String())
}

func TestWriteNumber_NaN(t *testing.T) {
	var sb strings.Builder
	ioutil.WriteNumber(&sb, math.NaN(), -1)
	assert.Equal(t, "NaN", sb.String())
}

func TestWriteNumber_LargePrecision(t *testing.T) {
	var sb strings.Builder
	ioutil.WriteNumber(&sb, 1.0, 10)
	assert.Equal(t, "1.0000000000", sb.String())
}
