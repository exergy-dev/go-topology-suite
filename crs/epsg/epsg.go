package epsg

import (
	"sync"

	"github.com/exergy-dev/go-topology-suite/crs"
)

// registry maps EPSG codes to their *crs.CRS entry. It is populated by
// init() across the files in this package and is read-only after init.
var (
	registryMu sync.RWMutex
	registry   = map[int]*crs.CRS{}
)

// register inserts c into the registry, keyed by c.Code. It panics on a
// duplicate registration; this can only fire at init time and indicates a
// programming error in this package.
func register(c *crs.CRS) *crs.CRS {
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, dup := registry[c.Code]; dup {
		panic("epsg: duplicate registration for code")
	}
	registry[c.Code] = c
	return c
}

// Lookup returns the registered CRS for the given EPSG code, or nil if the
// code is not known to this package. The returned pointer is the same
// instance shared by the corresponding exported variable (when one exists)
// and by any other Lookup call for the same code.
func Lookup(code int) *crs.CRS {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return registry[code]
}

// Codes returns the sorted list of EPSG codes registered in this package.
// It is intended for diagnostics and tests.
func Codes() []int {
	registryMu.RLock()
	defer registryMu.RUnlock()
	out := make([]int, 0, len(registry))
	for code := range registry {
		out = append(out, code)
	}
	// Insertion-sort: registry is < 200 entries, allocations dominate.
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j-1] > out[j]; j-- {
			out[j-1], out[j] = out[j], out[j-1]
		}
	}
	return out
}

// Geographic CRSes (2D unless noted).
//
// The 4326/3857/4269 instances are re-exported here for convenience. They
// are *separate* pointers from crs.WGS84/WebMercator/NAD83 (since this
// package must not mutate the parent crs package), but compare equal under
// crs.Equal because the (Authority, Code) pair matches.
var (
	WGS84       = register(&crs.CRS{Authority: "EPSG", Code: 4326, Kind: crs.Geographic})
	NAD83       = register(&crs.CRS{Authority: "EPSG", Code: 4269, Kind: crs.Geographic})
	NAD27       = register(&crs.CRS{Authority: "EPSG", Code: 4267, Kind: crs.Geographic})
	WGS72       = register(&crs.CRS{Authority: "EPSG", Code: 4322, Kind: crs.Geographic})
	ETRS89      = register(&crs.CRS{Authority: "EPSG", Code: 4258, Kind: crs.Geographic})
	WGS84_3D    = register(&crs.CRS{Authority: "EPSG", Code: 4979, Kind: crs.Geographic})
	CGCS2000    = register(&crs.CRS{Authority: "EPSG", Code: 4490, Kind: crs.Geographic})
	Beijing1954 = register(&crs.CRS{Authority: "EPSG", Code: 4214, Kind: crs.Geographic})
)

// Named projected CRSes. UTM zones are registered programmatically in init()
// (see projected.go) and are reachable only via Lookup since there are 120
// of them.
var (
	WebMercator         = register(&crs.CRS{Authority: "EPSG", Code: 3857, Kind: crs.Projected})
	Lambert93           = register(&crs.CRS{Authority: "EPSG", Code: 2154, Kind: crs.Projected})
	BritishNationalGrid = register(&crs.CRS{Authority: "EPSG", Code: 27700, Kind: crs.Projected})
	ConusAlbers         = register(&crs.CRS{Authority: "EPSG", Code: 5070, Kind: crs.Projected})
	EuropeLAEA          = register(&crs.CRS{Authority: "EPSG", Code: 3035, Kind: crs.Projected})
)
