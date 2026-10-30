# Benchmarks

Performance evidence for go-topology-suite: a micro-benchmark suite across
previously-unbenchmarked hot packages, seven macro workloads, a timed
cross-implementation comparison against simplefeatures and GEOS, and three
profile-driven optimizations landed in this pass. Every number below traces
to committed raw output under `bench/results/` or to the three optimization
commits' own benchstat evidence — see [Reproduction](#reproduction).

This is a snapshot, not a tracked regression suite — there is no CI
performance gate. Numbers are "true on this machine, this day, these
fixtures" (see [Limitations](#limitations)).

## Methodology

**Hardware/OS.** Intel Core i9-14900KF (hybrid P/E-core, 32 threads) under
WSL2 on Windows — no CPU pinning, and the Windows host scheduler is not under
our control. This is a real source of noise, worse for benchmarks in the
low-microsecond range where one scheduler preemption is a large fraction of
the sample. Every table here is `benchstat` output with CIs left in: read the
`±` and `p=` columns, and treat any comparison benchstat marks `~` as noise,
not a result.

**Toolchain.** Go 1.25.3 linux/amd64. `benchstat`
(`golang.org/x/perf/cmd/benchstat`). Every run is `-count=10`.

**Comparison implementations.**
- **gts** — this library, a pure-Go JTS port.
- **simplefeatures v0.59.0** — pure-Go OGC Simple Features library, also with
  a JTS-ported overlay engine.
- **GEOS 3.10.2** — system `libgeos-dev`, via cgo. This is a 2022-era release
  that **predates GEOS's own RelateNG and OverlayNG performance work landed
  in 3.12+**; treat the GEOS column as "old GEOS on this machine," not
  current GEOS. `twpayne/go-geos` v0.21 needs GEOS ≥3.11 symbols this system
  lacks, so the adapter (`bench/compare/geos_impl.go`, `geos` build tag) is a
  minimal direct cgo binding limited to ops present in 3.10.

**Snapshots** (`bench/results/`):

| Snapshot | Directory | Code state | Committed as |
|---|---|---|---|
| Baseline | `2026-07-10-baseline/` | `06e534b0` (bench suite, pre-fix) | `27081c9c` |
| Final | `2026-07-10-final/` | `11b2b019` (all 3 fixes) | `dbf73e22`; `compare.txt` recaptured as `86c2e95b` |
| 2026-10-30 lean wave | `2026-10-30-m3-lean-wins/` | main `6c6d023` vs `perf/lean-wins` (9 changes) | this commit; Apple M3, interleaved |

**GEOS crash caveat.** The first final `compare.txt` capture died mid-run
with a **SIGFPE inside native GEOS 3.10.2** during
`BenchmarkBuffer/impl=geos/n=256`, truncating output at 90/300 lines. The
rerun (`86c2e95b`) completed cleanly with no other anomaly across ~30 total
`-count=10` runs this pass — recorded as a known-flaky native crash in this
GEOS build, not a defect in this library or its adapter.

**JTS conformance held throughout:** 8,951 total / 8,942 passed / 9 known
divergences / 0 skipped (`internal/jtstest/harness_test.go`), unchanged by
every commit in this pass.

## Cross-implementation comparison

`benchstat -col /impl bench/results/2026-07-11-wave3/compare.txt`, verbatim.
Fixtures are identical gts-origin geometries converted via WKB
(`wkb.Marshal` → each impl's native unmarshal), untimed. `Length` runs on
an open 1,024-vertex LineString (coastlineA's ring, unclosed): polygon
Length is not comparable across implementations — simplefeatures defines
it as 0 and answers from a bare type check, while gts and GEOS return the
perimeter. Earlier snapshots (2026-07-10-final) used the polygon fixture,
so their Length rows measure different operations.

```
                             │      gts       │             simplefeatures             │                  geos                  │
                             │     sec/op     │    sec/op      vs base                 │    sec/op      vs base                 │
Intersection/n=1024-32            92.58m ± 1%    219.20m ± 5%  +136.77% (p=0.000 n=10)   124.28m ±  7%   +34.24% (p=0.000 n=10)
Union/n=1024-32                   123.1m ± 1%     233.0m ± 2%   +89.29% (p=0.000 n=10)    142.1m ± 10%   +15.41% (p=0.000 n=10)
Difference/n=1024-32              83.03m ± 3%    210.38m ± 1%  +153.38% (p=0.000 n=10)   115.31m ±  9%   +38.87% (p=0.000 n=10)
Relate/n=1024-32                 34.878m ± 2%    42.058m ± 1%   +20.58% (p=0.000 n=10)    5.821m ±  3%   -83.31% (p=0.000 n=10)
Area/n=1024-32                    728.7n ± 1%    1049.5n ± 1%   +44.02% (p=0.000 n=10)   1922.5n ±  1%  +163.83% (p=0.000 n=10)
Length/n=1024-32                  1.287µ ± 1%     1.612µ ± 1%   +25.21% (p=0.000 n=10)    2.026µ ±  1%   +57.38% (p=0.000 n=10)
Buffer/n=256-32                  2615.5µ ± 3%    2704.0µ ± 1%    +3.38% (p=0.002 n=10)    371.7µ ±  3%   -85.79% (p=0.000 n=10)
Buffer/n=1024-32                  36.96m ± 3%     84.04m ± 3%  +127.38% (p=0.000 n=10)    31.41m ±  5%   -15.02% (p=0.000 n=10)
PrepareBuild/n=1024-32         215266.5n ± 3%   82066.0n ± 3%   -61.88% (p=0.000 n=10)    560.2n ± 15%   -99.74% (p=0.000 n=10)
PreparedIntersects/n=1024-32      815.5µ ± 1%    2516.3µ ± 1%  +208.57% (p=0.000 n=10)    685.7µ ±  1%   -15.92% (p=0.000 n=10)
geomean                           1.871m          2.854m        +52.56%                   854.3µ         -54.34%

                             │        gts         │                simplefeatures                │                  geos                  │
                             │        B/op        │       B/op         vs base                   │    B/op      vs base                   │
Intersection/n=1024-32         43045908.50 ± 0%     121449654.00 ± 0%  +182.14% (p=0.000 n=10)     16.00 ± 62%  -100.00% (p=0.000 n=10)
Union/n=1024-32                40125652.00 ± 0%     121126527.50 ± 0%  +201.87% (p=0.000 n=10)     16.00 ±  0%  -100.00% (p=0.000 n=10)
Difference/n=1024-32           41555557.00 ± 0%     121541806.00 ± 0%  +192.48% (p=0.000 n=10)     16.00 ±  0%  -100.00% (p=0.000 n=10)
Relate/n=1024-32               23175655.50 ± 0%      22237507.00 ± 0%    -4.05% (p=0.000 n=10)     16.00 ±  0%  -100.00% (p=0.000 n=10)
Area/n=1024-32                       0.000 ± 0%           16.000 ± 0%         ? (p=0.000 n=10)     8.000 ±  0%         ? (p=0.000 n=10)
Length/n=1024-32                     0.000 ± 0%            0.000 ± 0%         ~ (p=1.000 n=10) ¹   8.000 ±  0%         ? (p=0.000 n=10)
Buffer/n=256-32                 2497544.00 ± 0%       1872684.50 ± 0%   -25.02% (p=0.000 n=10)     16.00 ±  0%  -100.00% (p=0.000 n=10)
Buffer/n=1024-32               20356489.00 ± 0%      45918856.50 ± 0%  +125.57% (p=0.000 n=10)     16.00 ±  0%  -100.00% (p=0.000 n=10)
PrepareBuild/n=1024-32           149329.00 ± 0%        102449.00 ± 0%   -31.39% (p=0.000 n=10)     24.00 ±  0%   -99.98% (p=0.000 n=10)
PreparedIntersects/n=1024-32         0.0Ki ± 0%          383.3Ki ± 0%         ? (p=0.000 n=10)     0.0Ki ±  0%         ~ (p=1.000 n=10) ¹
geomean                                         ²                      ?                       ²                ?                       ²
¹ all samples are equal
² summaries must be >0 to compute geomean

                             │        gts        │                simplefeatures                │                 geos                  │
                             │     allocs/op     │    allocs/op      vs base                    │ allocs/op   vs base                   │
Intersection/n=1024-32         156830.000 ± 0%     1391609.500 ± 0%   +787.34% (p=0.000 n=10)     1.000 ± 0%  -100.00% (p=0.000 n=10)
Union/n=1024-32                150513.500 ± 0%     1397286.000 ± 0%   +828.35% (p=0.000 n=10)     1.000 ± 0%  -100.00% (p=0.000 n=10)
Difference/n=1024-32           159607.000 ± 0%     1398098.000 ± 0%   +775.96% (p=0.000 n=10)     1.000 ± 0%  -100.00% (p=0.000 n=10)
Relate/n=1024-32               390920.000 ± 0%      493385.000 ± 0%    +26.21% (p=0.000 n=10)     1.000 ± 0%  -100.00% (p=0.000 n=10)
Area/n=1024-32                      0.000 ± 0%           1.000 ± 0%          ? (p=0.000 n=10)     1.000 ± 0%         ? (p=0.000 n=10)
Length/n=1024-32                    0.000 ± 0%           0.000 ± 0%          ~ (p=1.000 n=10) ¹   1.000 ± 0%         ? (p=0.000 n=10)
Buffer/n=256-32                   998.000 ± 0%       21690.000 ± 0%  +2073.35% (p=0.000 n=10)     1.000 ± 0%   -99.90% (p=0.000 n=10)
Buffer/n=1024-32                46723.000 ± 0%      813942.500 ± 0%  +1642.06% (p=0.000 n=10)     1.000 ± 0%  -100.00% (p=0.000 n=10)
PrepareBuild/n=1024-32            149.000 ± 0%        2076.000 ± 0%  +1293.29% (p=0.000 n=10)     1.000 ± 0%   -99.33% (p=0.000 n=10)
PreparedIntersects/n=1024-32        0.00k ± 0%          14.51k ± 0%          ? (p=0.000 n=10)     0.00k ± 0%         ~ (p=1.000 n=10) ¹
geomean                                        ²                     ?                        ²               ?                       ²
¹ all samples are equal
² summaries must be >0 to compute geomean
```

**Notes.** gts leads simplefeatures on every timed operation in the table
except `PrepareBuild` (see Deferred backlog), and leads GEOS 3.10.2 on the
binary overlays, `Area`, and `Length` (the cgo boundary dominates the tiny
ops — read those GEOS cells as cgo-vs-Go-loop artifacts, not algorithmic
claims). Conversion (WKB round trip into each impl's native type) is
untimed. Prepared-build cost is reported separately (`PrepareBuild`), not
folded into `PreparedIntersects`. GEOS 3.10.2 predates its own recent perf
work (see Methodology). History of the headline `Intersection/n=1024` gts
column: 4446.2ms at the 2026-07-10 baseline, 547.2ms after that pass
(2026-07-10-final), 92.6ms after the 2026-07-11 wave below — from ~20.9x
slower than simplefeatures to ~2.4x faster.
## Optimization log

### 2026-10-30 wave — "lean wins": scaling fixes under a memory ceiling

**Machine and method differ from the rest of this document.** Apple M3
(8 cores, macOS), Go 1.23.1 for the main module and the go1.25 toolchain for
the `bench` module via GOTOOLCHAIN=auto. Every table here is `benchstat` over
count=6 of `main` (`6c6d023`) against branch `perf/lean-wins`. The macro and
cross-implementation tables used *interleaved* runs of two prebuilt test
binaries (main, branch, main, branch, ...) to cancel machine drift after a
first pass was contaminated by an unrelated process; the micro tables were
run back-to-back on an otherwise idle machine. Raw files:
`bench/results/2026-10-30-m3-lean-wins/` (`micro-*.txt`, `macro-*.txt`,
`compare-*.txt`, plus the `*-benchstat.txt` outputs quoted below).

**Constraint.** The user's rule for this wave was "no improvement may balloon
memory": bytes/op had to stay within 1.3x of main. Three changes that
initially traded memory for time (indexed Distance at 20x bytes, indexed
TopologyPreserving at 7x, indexed simplify repair at +62%) were reworked onto
flat, pooled, sort-free structures until they met it; all three now *lower*
bytes/op at the sizes benchmarked except TopologyPreserving n=8192 (+5%) and
Validate on jagged stars (+15%). One candidate was **held out** for the same
reason: capping the convex Sutherland–Hodgman fast path in
`overlay/overlay.go` by segment-pair count (-48% at n=1024, -94% at n=8192 on
Intersection/ngon) routes large inputs onto the general overlay path, whose
allocation footprint is 18–37x the alloc-free fast path's. It is correct and
measured (worktree branch `worktree-agent-af2bdae9ef6264284`, commit
`21cedc4`) but not merged.

**Process.** Profiling of the full suite (120 micro + 7 macro + compare)
found six paths with quadratic or worse scaling, each pinned by a CPU profile
to one brute-force loop. Each fix was implemented by a separate agent in its
own worktree, gated on byte-identical output (a differential test against the
retained old implementation, usually with `pgregory.net/rapid`), the full
test suite, and the JTS corpus (8951/8942/9 throughout), then A/B-benchmarked
by the orchestrator. Agent-reported numbers were treated as direction only;
every figure below is from the orchestrator's runs.

**What landed** (9 changes, 15 commits, in branch order):

| Area | Mechanism | Headline (main → branch) | Bytes/op |
|---|---|---|---|
| `predicate` early exit | `Intersects`/`Disjoint`/`Contains`/`Within`/`Equals` evaluate their RelateNG predicate via `Evaluate`; scan stops when the result is known | IntersectsCold poly-poly n=1024 2.79 ms → 0.59 ms (-79%); RelateNGReuse -51% | -22% to -52% |
| `wkt` codec | ordinates appended into one flat buffer; `AppendFloat` on encode | Unmarshal n=1024 1,043 → 18 allocs (-6%); Marshal 2,071 → 21 (-21%) | -12% to -57% |
| `geojson` codec | streaming `coordinates` scanner; `AppendFloat` on encode | Unmarshal n=1024 988 µs → 380 µs (-62%), 6,198 → 20 allocs | -20% to -90% |
| `overlayng`/`noding` allocs | face-edge arena, flat split records, chain slab, dense union-find, `NewPolygon` `[][]XY` fast path | star overlays -8% to -15% time, -85% to -90% allocs; Union/ngon -40% | -1% to -30% |
| `prepare.Polygon` | sort-free packed segment tree, 8-edge leaves, precomputed envelopes | build n=1024 123 µs → 12 µs (-91%); ContainsPoint -38% | -40% to -70% |
| `overlayng` simplify repair | flat packed segment tree for `injectRingSelfContacts`, reused across passes | Simplify n=1024 2.66 ms → 0.32 ms (-88%) | -10% to -19% |
| `simplify.TopologyPreserving` | flat block trees for jump/crossing checks above 1,536 vertices; lazy build | n=8192 295 ms → 14 ms (-95%); n=1024 -87% (shared repair fix) | -12% (n=1024), +5% (n=8192) |
| `measure.Distance` | pooled x-sorted facet sweep with envelope-gap pruning; bit-identical | n=1024 38 ms → 0.16 ms (-99.6%); 2 → 0 allocs | -100% |
| `validate` self-intersection | slab-allocated monotone chains, closure-free sweep, brute force under 32 segments | ngon n=8192 108 ms → 2.3 ms (-98%); star -53% to -64% | 0% (ngon), +15% (star) |

**Micro benchmarks** (`benchstat micro-main.txt micro-lean.txt`, condensed to
the rows the wave touches; the full output is `micro-benchstat.txt`):

```
                                  │    main     │           perf/lean-wins            │
                                  │   sec/op    │   sec/op     vs base                │
Distance/near/n=1024-8             38382.4µ ± 1%   164.5µ ± 1%  -99.57% (p=0.002 n=6)
Distance/disjoint/n=1024-8         38003.2µ ± 1%   162.1µ ± 1%  -99.57% (p=0.002 n=6)
Distance/near/n=64-8               180.222µ ± 0%   8.314µ ± 0%  -95.39% (p=0.002 n=6)
Validate/ngon/n=8192-8             111.792m ±11%   2.304m ±11%  -97.94% (p=0.002 n=6)
Validate/ngon/n=1024-8             1671.6µ ± 0%    286.0µ ± 3%  -82.89% (p=0.002 n=6)
Validate/star/n=8192-8             119.10m ± 4%    43.19m ±15%  -63.73% (p=0.002 n=6)
Validate/star/n=1024-8             1664.5µ ± 0%    777.4µ ± 2%  -53.30% (p=0.002 n=6)
Validate/multipoly=100-8           3.521m ± 3%     3.555m ±26%        ~ (p=0.937 n=6)
TopologyPreserving/n=8192-8        294.99m ± 1%    13.98m ±30%  -95.26% (p=0.002 n=6)
TopologyPreserving/n=1024-8        2695.3µ ± 0%    347.8µ ± 6%  -87.09% (p=0.002 n=6)
Simplify/n=1024-8                  2664.5µ ± 0%    318.3µ ± 1%  -88.05% (p=0.002 n=6)
Simplify/n=8192-8                  8.857m ± 0%     2.107m ± 6%  -76.21% (p=0.002 n=6)
PolygonBuild/n=1024-8              140.29µ ±32%    11.69µ ±51%  -91.67% (p=0.002 n=6)
PolygonBuild/n=8192-8              1694.00µ ±59%   76.51µ ±76%  -95.48% (p=0.002 n=6)
PreparedContainsPoint/prepared-8   232.2n ± 0%     144.4n ± 1%  -37.79% (p=0.002 n=6)
IntersectsCold/poly-poly/n=1024-8  2793.6µ ± 1%    592.4µ ± 0%  -78.80% (p=0.002 n=6)
IntersectsCold/poly-poly/n=64-8    151.46µ ± 1%    79.11µ ± 0%  -47.77% (p=0.002 n=6)
IntersectsCold/poly-point/n=1024-8 2.658µ ± 1%     2.560µ ± 1%   -3.69% (p=0.002 n=6)
RelateNGReuse/n=1024-8             77.95µ ± 1%     38.07µ ± 1%  -51.17% (p=0.002 n=6)
RelateNGOneShot/n=1024-8           165.5µ ± 0%     127.2µ ± 1%  -23.14% (p=0.002 n=6)
Union/ngon/n=1024-8                1172.8µ ±15%    699.2µ ± 1%  -40.38% (p=0.002 n=6)
Union/star/n=1024-8                53.64m ± 1%     45.77m ± 1%  -14.67% (p=0.002 n=6)
Difference/ngon/n=1024-8           745.7µ ± 1%     657.8µ ± 1%  -11.79% (p=0.002 n=6)
Difference/star/n=1024-8           94.19m ± 1%     86.79m ± 1%   -7.85% (p=0.002 n=6)
Intersection/star/n=1024-8         41.32m ±11%     37.69m ± 1%   -8.78% (p=0.002 n=6)
Intersection/ngon/n=1024-8         1.623m ± 0%     1.620m ± 0%        ~ (p=0.180 n=6)
UnaryUnion/grid/n=256-8            2.483m ± 2%     2.209m ± 2%  -11.05% (p=0.002 n=6)
geojson Unmarshal/n=1024-8         1000.4µ ± 8%    380.2µ ± 0%  -61.99% (p=0.002 n=6)
geojson Unmarshal/multipoly=100-8  420.59µ ± 0%    77.95µ ± 4%  -81.47% (p=0.002 n=6)
geojson Marshal/n=1024-8           176.4µ ±10%     135.9µ ± 0%  -22.98% (p=0.002 n=6)
wkt Marshal/n=1024-8               173.9µ ±15%     136.7µ ± 2%  -21.41% (p=0.002 n=6)
wkt Unmarshal/n=1024-8             238.2µ ± 9%     222.9µ ± 3%   -6.42% (p=0.002 n=6)

                                  │     main     │          perf/lean-wins           │
                                  │     B/op     │     B/op      vs base             │
Distance/near/n=1024-8             36864.00 ± 0%     28.00 ± ?   -99.92% (p=0.002 n=6)
Validate/star/n=8192-8             1.991Mi ± 0%    2.304Mi ± 0%  +15.70% (p=0.002 n=6)
Validate/ngon/n=8192-8             1.991Mi ± 0%    1.991Mi ± 0%        ~ (p=0.394 n=6)
TopologyPreserving/n=8192-8        973.5Ki ± 0%   1020.2Ki ± 0%   +4.80% (p=0.002 n=6)
TopologyPreserving/n=1024-8        293.2Ki ± 0%    259.4Ki ± 0%  -11.51% (p=0.002 n=6)
Simplify/n=1024-8                  176.5Ki ± 0%    142.8Ki ± 0%  -19.12% (p=0.002 n=6)
PolygonBuild/n=1024-8              145.81Ki ± 0%   87.05Ki ± 0%  -40.30% (p=0.002 n=6)
IntersectsCold/poly-poly/n=1024-8  2.300Mi ± 0%    1.114Mi ± 0%  -51.55% (p=0.002 n=6)
Union/ngon/n=1024-8                1.641Mi ± 0%    1.442Mi ± 0%  -12.13% (p=0.002 n=6)
Union/star/n=1024-8                28.44Mi ± 0%    27.45Mi ± 0%   -3.47% (p=0.002 n=6)
Difference/star/n=1024-8           65.70Mi ± 0%    64.89Mi ± 0%   -1.23% (p=0.002 n=6)
geojson Unmarshal/n=1024-8         423.74Ki ± 0%   58.94Ki ± 0%  -86.09% (p=0.002 n=6)
wkt Unmarshal/n=1024-8             132.7Ki ± 0%    116.7Ki ± 0%  -12.07% (p=0.002 n=6)

                                  │     main     │          perf/lean-wins           │
                                  │  allocs/op   │  allocs/op    vs base             │
Distance/near/n=1024-8             2.000 ± 0%       0.000 ± 0%  -100.00% (p=0.002 n=6)
Validate/star/n=8192-8             621.5 ± 2%       624.0 ± 1%         ~ (p=0.682 n=6)
TopologyPreserving/n=8192-8        1.146k ± 0%      1.211k ± 0%   +5.67% (p=0.002 n=6)
Simplify/n=1024-8                  26.00 ± 0%       25.00 ± 0%   -3.85% (p=0.002 n=6)
PolygonBuild/n=1024-8              148.00 ± 0%      10.00 ± 0%  -93.24% (p=0.002 n=6)
IntersectsCold/poly-poly/n=1024-8  29.140k ± 0%     8.597k ± 0%  -70.50% (p=0.002 n=6)
Union/star/n=1024-8                113.94k ± 0%     11.20k ± 0%  -90.17% (p=0.002 n=6)
Difference/star/n=1024-8           244.67k ± 0%     36.19k ± 0%  -85.21% (p=0.002 n=6)
Intersection/star/n=1024-8         120.68k ± 0%     17.93k ± 0%  -85.14% (p=0.002 n=6)
geojson Unmarshal/n=1024-8         6198.00 ± 0%     20.00 ± 0%  -99.68% (p=0.002 n=6)
wkt Unmarshal/n=1024-8             1043.00 ± 0%     18.00 ± 0%  -98.27% (p=0.002 n=6)
wkt Marshal/n=1024-8               2071.00 ± 0%     21.00 ± 0%  -98.99% (p=0.002 n=6)
```

The wide `±` on `PolygonBuild` and `TopologyPreserving/n=8192` is spread in
the *branch* column of sub-100µs or GC-heavy ops on a desktop machine, not
disagreement about direction (both p=0.002). `Validate/multipoly=100` and
`Intersection/ngon/n=1024` are the two touched-adjacent rows that did not
move, as expected (small rings stay on brute force; the convex fast path is
unchanged).

**Macro workloads** (`bench/workloads.go`, interleaved prebuilt binaries,
`benchstat macro-main.txt macro-lean.txt`):

```
                         │ macro-main.txt │ macro-lean.txt │
                         │                                                              sec/op                                                               │                                                  sec/op                                                    vs base               │
Buffer-8                                                                                                                                         28.08m ± 1%                                                                                                 27.13m ± 2%   -3.40% (p=0.002 n=6)
GeographicIntersection-8                                                                                                                         1.467m ± 1%                                                                                                 1.467m ± 1%        ~ (p=0.937 n=6)
IngestWKB-8                                                                                                                                      888.1µ ± 2%                                                                                                 877.0µ ± 1%        ~ (p=0.093 n=6)
PairwiseIntersection-8                                                                                                                           63.54µ ± 1%                                                                                                 62.62µ ± 1%   -1.44% (p=0.002 n=6)
PointInPolygon-8                                                                                                                                 20.54m ± 1%                                                                                                 19.83m ± 1%   -3.48% (p=0.002 n=6)
PointInPolygonPrepared-8                                                                                                                         11.13m ± 1%                                                                                                 10.52m ± 1%   -5.50% (p=0.002 n=6)
UnaryUnion-8                                                                                                                                     8.553m ± 2%                                                                                                 7.648m ± 1%  -10.58% (p=0.002 n=6)
geomean                                                                                                                                          3.331m                                                                                                      3.207m        -3.72%

                         │ macro-main.txt │ macro-lean.txt │
                         │                                                               B/op                                                                │                                                   B/op                                                     vs base               │
Buffer-8                                                                                                                                        22.90Mi ± 0%                                                                                                23.18Mi ± 0%   +1.22% (p=0.002 n=6)
GeographicIntersection-8                                                                                                                        243.9Ki ± 0%                                                                                                241.0Ki ± 0%   -1.22% (p=0.002 n=6)
IngestWKB-8                                                                                                                                     1.602Mi ± 0%                                                                                                1.602Mi ± 0%        ~ (p=0.675 n=6)
PairwiseIntersection-8                                                                                                                          22.77Ki ± 0%                                                                                                19.82Ki ± 0%  -12.98% (p=0.002 n=6)
PointInPolygon-8                                                                                                                                6.106Mi ± 0%                                                                                                6.106Mi ± 0%        ~ (p=0.818 n=6)
PointInPolygonPrepared-8                                                                                                                        6.104Mi ± 0%                                                                                                6.104Mi ± 0%        ~ (p=0.563 n=6)
UnaryUnion-8                                                                                                                                    26.57Mi ± 0%                                                                                                23.96Mi ± 0%   -9.80% (p=0.002 n=6)
geomean                                                                                                                                         2.120Mi                                                                                                     2.048Mi        -3.40%

                         │ macro-main.txt │ macro-lean.txt │
                         │                                                             allocs/op                                                             │                                                allocs/op                                                 vs base                 │
Buffer-8                                                                                                                                        56.924k ± 0%                                                                                               9.331k ± 0%  -83.61% (p=0.002 n=6)
GeographicIntersection-8                                                                                                                         2.865k ± 0%                                                                                               2.828k ± 0%   -1.29% (p=0.002 n=6)
IngestWKB-8                                                                                                                                      30.00k ± 0%                                                                                               30.00k ± 0%        ~ (p=1.000 n=6) ¹
PairwiseIntersection-8                                                                                                                            317.0 ± 0%                                                                                                280.0 ± 0%  -11.67% (p=0.002 n=6)
PointInPolygon-8                                                                                                                                 200.0k ± 0%                                                                                               200.0k ± 0%        ~ (p=1.000 n=6)
PointInPolygonPrepared-8                                                                                                                         200.0k ± 0%                                                                                               200.0k ± 0%        ~ (p=1.000 n=6) ¹
UnaryUnion-8                                                                                                                                    122.29k ± 0%                                                                                               94.95k ± 0%  -22.36% (p=0.002 n=6)
geomean                                                                                                                                          25.79k                                                                                                    18.84k       -26.95%
¹ all samples are equal
```

Only `UnaryUnion` moves by double digits in time; the other six are within a
few percent, and none regress. The wave's macro effect is mostly memory:
`Buffer` allocations -84% and `UnaryUnion` -22% at -10% bytes. `Buffer`
bytes/op is +1.2% (the `NewPolygon` fast path writes directly into the flat
buffer, which reserves slightly more up front) — the only macro memory
increase, well inside the 1.3x ceiling.

**Cross-implementation suite, gts column only** (`bench/compare`, same
fixtures as the July tables, interleaved; simplefeatures and GEOS were not
re-run — this machine has no libgeos — so compare against the gts column of
the 2026-07-11 table only by ratio, not absolute):

```
                                     │ compare-main.txt │ compare-lean.txt │
                                     │                                                               sec/op                                                                │                                                   sec/op                                                     vs base               │
Intersection/impl=gts/n=1024-8                                                                                                                                 53.60m ± 1%                                                                                                   49.66m ± 1%   -7.35% (p=0.002 n=6)
Union/impl=gts/n=1024-8                                                                                                                                        68.46m ± 1%                                                                                                   59.44m ± 0%  -13.17% (p=0.002 n=6)
Difference/impl=gts/n=1024-8                                                                                                                                   120.9m ± 1%                                                                                                   113.7m ± 1%   -5.94% (p=0.002 n=6)
Relate/impl=gts/n=1024-8                                                                                                                                       19.33m ± 1%                                                                                                   19.21m ± 2%        ~ (p=0.394 n=6)
Area/impl=gts/n=1024-8                                                                                                                                         714.5n ± 1%                                                                                                   713.2n ± 0%        ~ (p=0.610 n=6)
Length/impl=gts/n=1024-8                                                                                                                                       678.9n ± 0%                                                                                                   679.0n ± 1%        ~ (p=0.937 n=6)
Buffer/impl=gts/n=256-8                                                                                                                                        1.377m ± 1%                                                                                                   1.336m ± 1%   -2.97% (p=0.002 n=6)
Buffer/impl=gts/n=1024-8                                                                                                                                       26.31m ± 1%                                                                                                   25.57m ± 0%   -2.79% (p=0.002 n=6)
PrepareBuild/impl=gts/n=1024-8                                                                                                                               104.902µ ± 1%                                                                                                   9.242µ ± 2%  -91.19% (p=0.002 n=6)
PreparedIntersects/impl=gts/n=1024-8                                                                                                                           485.5µ ± 1%                                                                                                   372.6µ ± 1%  -23.25% (p=0.002 n=6)
geomean                                                                                                                                                        1.226m                                                                                                        904.7µ       -26.21%

                                     │ compare-main.txt │ compare-lean.txt │
                                     │                                                                B/op                                                                 │                                                   B/op                                                     vs base                 │
Intersection/impl=gts/n=1024-8                                                                                                                              40.52Mi ± 0%                                                                                                  39.42Mi ± 0%   -2.72% (p=0.002 n=6)
Union/impl=gts/n=1024-8                                                                                                                                     37.75Mi ± 0%                                                                                                  35.49Mi ± 0%   -5.99% (p=0.002 n=6)
Difference/impl=gts/n=1024-8                                                                                                                                86.85Mi ± 0%                                                                                                  84.54Mi ± 0%   -2.66% (p=0.002 n=6)
Relate/impl=gts/n=1024-8                                                                                                                                    22.10Mi ± 0%                                                                                                  22.10Mi ± 0%        ~ (p=0.619 n=6)
Area/impl=gts/n=1024-8                                                                                                                                        0.000 ± 0%                                                                                                    0.000 ± 0%        ~ (p=1.000 n=6) ¹
Length/impl=gts/n=1024-8                                                                                                                                      0.000 ± 0%                                                                                                    0.000 ± 0%        ~ (p=1.000 n=6) ¹
Buffer/impl=gts/n=256-8                                                                                                                                     2.382Mi ± 0%                                                                                                  2.204Mi ± 0%   -7.48% (p=0.002 n=6)
Buffer/impl=gts/n=1024-8                                                                                                                                    19.41Mi ± 0%                                                                                                  19.89Mi ± 0%   +2.44% (p=0.002 n=6)
PrepareBuild/impl=gts/n=1024-8                                                                                                                             145.83Ki ± 0%                                                                                                  87.06Ki ± 0%  -40.30% (p=0.002 n=6)
PreparedIntersects/impl=gts/n=1024-8                                                                                                                          0.000 ± 0%                                                                                                    0.000 ± 0%        ~ (p=1.000 n=6) ¹
geomean                                                                                                                                                                  ²                                                                                                               -6.63%               ²
¹ all samples are equal
² summaries must be >0 to compute geomean

                                     │ compare-main.txt │ compare-lean.txt │
                                     │                                                              allocs/op                                                              │                                                 allocs/op                                                  vs base                 │
Intersection/impl=gts/n=1024-8                                                                                                                              160.37k ± 0%                                                                                                   24.80k ± 0%  -84.54% (p=0.002 n=6)
Union/impl=gts/n=1024-8                                                                                                                                     154.05k ± 0%                                                                                                   13.91k ± 0%  -90.97% (p=0.002 n=6)
Difference/impl=gts/n=1024-8                                                                                                                                324.08k ± 0%                                                                                                   50.14k ± 0%  -84.53% (p=0.002 n=6)
Relate/impl=gts/n=1024-8                                                                                                                                     390.9k ± 0%                                                                                                   390.9k ± 0%        ~ (p=1.000 n=6)
Area/impl=gts/n=1024-8                                                                                                                                        0.000 ± 0%                                                                                                    0.000 ± 0%        ~ (p=1.000 n=6) ¹
Length/impl=gts/n=1024-8                                                                                                                                      0.000 ± 0%                                                                                                    0.000 ± 0%        ~ (p=1.000 n=6) ¹
Buffer/impl=gts/n=256-8                                                                                                                                       998.0 ± 0%                                                                                                    212.0 ± 0%  -78.76% (p=0.002 n=6)
Buffer/impl=gts/n=1024-8                                                                                                                                    46723.0 ± 0%                                                                                                    530.0 ± 0%  -98.87% (p=0.002 n=6)
PrepareBuild/impl=gts/n=1024-8                                                                                                                               149.00 ± 0%                                                                                                    11.00 ± 0%  -92.62% (p=0.002 n=6)
PreparedIntersects/impl=gts/n=1024-8                                                                                                                          0.000 ± 0%                                                                                                    0.000 ± 0%        ~ (p=1.000 n=6) ¹
geomean                                                                                                                                                                  ²                                                                                                              -77.17%               ²
¹ all samples are equal
² summaries must be >0 to compute geomean
```

`PrepareBuild` is the headline: 105µs → 9.2µs, which turns the one
comparison simplefeatures led (52µs on this machine) into a ~5.6x gts lead.
`Buffer/n=1024` bytes/op is +2.4%, the same `NewPolygon` reservation as the
macro `Buffer` row.


**Caveats specific to this wave.** The validator's indexed path pads chain
envelopes by 1e-9 relative and additionally checks each chain against itself
because the float predicate can report contact when envelopes miss by a few
ULPs; a 50,000-case differential test found no mismatch, but this is a
guard, not a proof. GeoJSON decode-error *text* for malformed coordinate
arrays changed wording (accept/reject behaviour did not). The predicates
change adds ~3% to the already sub-3µs poly-point path. Memory numbers for
`Distance` show a few bytes/op from benchstat averaging cold `sync.Pool`
refills; allocs/op is 0.


### 2026-07-11 wave — "beat simplefeatures everywhere it led"

Profile-driven, conformance-gated (JTS baseline 8951/8942/9 held
throughout; `bench/conformance` and `TestCompareAgreement` green). The
headline levers, in decreasing measured impact:

- **`needsCanonicalize` touch detection indexed** (`internal/overlayng/overlay.go`) —
  the brute-force `polygonHasTouchingHole` / `multiPolygonsTouch` per-pair
  O(V×L) vertex-on-segment sweeps were 59% of Union and 49% of Intersection
  CPU on the jagged-star fixture. Replaced with one STR R-tree over all
  result-ring segment envelopes probed once per vertex, preserving the
  exact predicates (same pair eligibility, vertex-set skip,
  `pointOnSegmentInterior`). The repeated-interior-vertex probe also lost
  its per-ring map (sort + adjacent-dup scan on a shared scratch).
- **Prepared point-in-polygon for face classification**
  (`internal/overlayng/prepared_pip.go`) — classification probed each DCEL
  face against the original input rings via full `geomath.PointInRing`
  scans (19–53% of overlay CPU depending on op). The prepared form buckets
  segments by y-extent and precomputes per-segment x-ranges so the
  crossing division only runs for segments genuinely bracketing the probe;
  a guard band (1e-12 relative) keeps the parity identical to the full
  scan (differentially tested in `prepared_pip_test.go`).
- **Buffer noding switched to the monotone-chain noder**
  (`buffer/polygonize.go`, `internal/snaprounding/noder.go`) — offset
  curves chain into few monotone pieces, so `MCIndexNoder` replaces the
  per-segment R-tree of `IndexedNoder` on both the tolerance-0 branch and
  inside the snap-rounding fixpoint's `adaptiveNode`.
- **Hot-pixel set re-indexed as an x-sorted array** (`internal/snap/hotpixel.go`) —
  pixels are points, so the R-tree (build + per-segment rectangle search,
  ~32% of Buffer) became one sort-and-dedupe plus binary search + short
  scan per query, bit-identical candidate sets, zero per-query allocation.
- **DCEL build dedup without hashing** (`internal/overlayng/dcel.go`) —
  vertex dedup by sort + binary search on packed coordinate keys (the key
  IS the coordinate, bit-cast); edge dedup keyed by packed vertex-id
  uint64 (and skipped entirely for depth-mode builds, whose single caller
  pre-merges coincident segments in `flattenChains`); one shared arena for
  all `Vertex.Out` slices sized from key-run lengths; manual compare-swap
  for the degree-≤2 angular sort that dominates noded arrangements.
- **Dense-index consumers replace pointer-keyed maps** — buffer subgraph
  union-find/labelling and ring extraction now key off new
  `Vertex/Face/HalfEdge.Index()` ids (epoch-stamped scratch) instead of
  `map[*T]bool`; `flattenChains` accumulates via sort-and-merge on the
  canonical key instead of `map[canonKey]*accum`.
- **Buffer validator rep points** (`buffer/polygonize.go`) — the
  positive-buffer winding validator now derives a cheap O(ring) interior
  point (longest-edge midpoint nudge, parity-verified) instead of the
  ~97-probe inscribed-circle search, which only the negative-buffer
  clearance check actually needs; validators receive the ring and choose.
- **Measure fast paths** (`measure/measure.go`) — planar `Area` shoelaces
  the flat coordinate buffer via `internal/flatref` (no 18KB `Ring(0)`
  copy); planar `Length` sums sqrt(dx²+dy²) off the flat buffer with a
  2x-unrolled dual accumulator (one accumulator pins the loop to SQRTSD
  latency); `resolve` no longer heap-allocates its config on the
  zero-option path. Length's sqrt matches JTS `Length.ofLine`/GEOS;
  kernel `Distance` (math.Hypot) is unchanged.
- **Misc**: `dropPhantomSliverHoles` early-outs before building its input
  grid when the Union result has no holes; STR bulk-load uses unstable
  sorts (query results are sets; tie order was never meaningful);
  `insertSplitsInto` copies chains lazily (zero-split chains, the common
  case, are returned as-is).

Cost profile after the wave (same fixtures): binary overlays carry ~40%
fewer bytes and ~60% fewer allocations than before it; `Buffer/n=256` went
from 12,540 allocs / 4.7MB to 998 allocs / 2.5MB.

### 2026-07-10 pass

Three profile-driven fixes, each gated on a profile confirming the
hypothesis first, byte-identical output, and the JTS baseline (8951/8942/0)
holding after. Each commit quotes its own before/after in the message; the
tables below independently reproduce the same comparisons from the committed
baseline/final `micro.txt`. Where the two disagree slightly, that's WSL2
noise between runs taken hours apart, not a different result (noted per fix).

### 1. overlayng: hoist ring materialization out of the O(P²) touch check

**Commit `c8cd92b2`.** `needsCanonicalize` (`internal/overlayng/overlay.go`)
ran `multiPolygonsTouch` after every overlay producing ≥2 output polygons.
Its all-pairs loop re-fetched `polys[j].Ring(0)` — an allocating copy — and
rebuilt polygon `j`'s vertex-set map on *every* `(i,j)` iteration, before the
O(ring) segment scan it actually needed. A jagged-star intersection emitting
3,306 fragments spent **79% cumulative CPU and 93% of allocations** there,
scaling worse than cubically. Fix: materialize rings/vertex-sets/envelopes
once per polygon; skip pairs whose cached envelopes don't intersect (touching
rings must have intersecting envelopes). Byte-identical by construction.

```
$ benchstat overlayng_old.txt overlayng_new.txt   # Intersection|Difference/star/n=1024
                              │  baseline  │        final                         │
                              │   sec/op   │   sec/op     vs base                 │
Intersection/star/n=1024-32    2497.0m ± 2%   255.9m ± 2%  -89.75% (p=0.000 n=10)
Difference/star/n=1024-32      2520.7m ±14%   251.5m ± 1%  -90.02% (p=0.000 n=10)
                              │  baseline  │        final                         │
                              │    B/op    │    B/op      vs base                 │
Intersection/star/n=1024-32    450.69Mi ± 0%  34.60Mi ± 0%  -92.32% (p=0.000 n=10)
Difference/star/n=1024-32      577.47Mi ± 0%  33.74Mi ± 0%  -94.16% (p=0.000 n=10)
                              │  baseline  │        final                         │
                              │ allocs/op  │ allocs/op    vs base                 │
Intersection/star/n=1024-32    5750.7k ± 0%   287.5k ± 0%  -95.00% (p=0.000 n=10)
Difference/star/n=1024-32      6729.6k ± 0%   291.0k ± 0%  -95.68% (p=0.000 n=10)
```

**Headline:** `Intersection/star/n=1024` **-89.75%** time, allocs -95.00%.
Cross-impl gap narrowed ~20.9x → ~2.5x (above). *Discrepancy:* the commit
message's own before/after quotes -89.6% and 4.46s→528ms — its point-in-time
run at commit time, vs -89.75%/547.2ms from the hours-later committed
snapshot. Same effect, same order of magnitude.

### 2. index: expand ancestor envelopes along the insert path, not the whole tree

**Commit `099916df`.** `RTree.Insert` (`index/rtree.go`) unconditionally ran
`recomputeEnvelopeRecursive` — a full-tree walk whose own doc comment scoped
it to bulk rebuilds — after every single insert, making sequential builds
O(n²). `buffer`'s snap-rounding noder pays this directly, inserting one hot
pixel per vertex. Profiles put **91% cumulative CPU** of
`BenchmarkBuffer/star/n=1024` inside the recursive recompute. Fix:
`chooseLeaf` records the root-to-leaf descent path; plain inserts expand
ancestor envelopes bottom-up along it (they can only grow); splits recompute
shallowly along the same path instead of a second full-tree search
(`findParent` retired). A new differential test checks bit-exact envelopes
and `Search` parity against brute force.

```
$ benchstat rtree_old.txt rtree_new.txt        # RTreeInsertSequential/n=10000
RTreeInsertSequential/n=10000-32   272.96m ± 2%  16.38m ± 1%  -94.00% (p=0.000 n=10)

$ benchstat bufstar_old.txt bufstar_new.txt    # Buffer/star/n=1024
Buffer/star/n=1024-32              2013.9m ± 5%  159.1m ± 2%  -92.10% (p=0.000 n=10)
```

**Headline:** `RTreeInsertSequential/n=10000` **-94.00%** time, allocs
unchanged (a reused scratch buffer, not fewer allocations, is the fix).
`Buffer/star/n=1024` **-92.10%** time. *Discrepancy:* commit message quotes
149ms (-92.6%) and "macro Buffer ~8-9x faster"; the committed macro snapshot
(below) shows Buffer -86.58%, ~7.45x, not 8-9x — the commit-time quote used a
different, earlier ad-hoc baseline than the one committed under
`bench/results/2026-07-10-baseline/macro.txt`. Trust the committed-snapshot
number; direction and order of magnitude agree.

### 3. measure: traverse rings via RingLen/RingVertex, not per-iteration copies

**Commit `11b2b019`.** `geometryDistance` (`measure/measure.go`) nests
`visitSegments(b,...)` inside `visitVertices(a,...)`; both materialized a
polygon ring with an allocating `Ring()` call **inside the loop body**, once
per outer vertex — O(vertices × ring length) allocation per `Distance` call.
`polygonContains` (`measure/distance_op.go`) had the identical pattern.
Profiles put **99% of `alloc_space`** for `Distance/near/n=1024` in
`Ring`/`RingInto`. Fix: walk rings via the existing non-allocating
`RingLen`/`RingVertex` accessors; hoist `polygonContains`'s ring fetch to
once per call. `Length`, Hausdorff variants, and `IndexedFacetDistance` share
the helpers and inherit the fix. Bit-identical output.

```
$ benchstat distance_old.txt distance_new.txt   # Distance/{disjoint,near}/n={64,1024}
                              │   baseline    │        final                        │
                              │    sec/op      │   sec/op     vs base                │
Distance/disjoint/n=64-32       257.2µ ± 2%      192.6µ ± 0%  -25.11% (p=0.000 n=10)
Distance/disjoint/n=1024-32     55.35m ± 3%      43.20m ± 1%  -21.96% (p=0.000 n=10)
Distance/near/n=64-32           259.5µ ± 1%      193.5µ ± 7%  -25.45% (p=0.000 n=10)
Distance/near/n=1024-32         58.03m ± 1%      45.88m ± 3%  -20.94% (p=0.000 n=10)
                              │    baseline     │        final                       │
                              │      B/op        │     B/op      vs base              │
Distance/disjoint/n=1024-32   110791.43Ki ± 0%     36.02Ki ± 0%  -99.97% (p=0.000 n=10)
Distance/near/n=1024-32       110791.50Ki ± 0%     36.02Ki ± 0%  -99.97% (p=0.000 n=10)
                              │   baseline    │        final                        │
                              │  allocs/op     │  allocs/op    vs base               │
Distance/disjoint/n=1024-32   6170.500 ± 0%      3.000 ± 0%   -99.95% (p=0.000 n=10)
Distance/near/n=1024-32       6171.000 ± 0%      3.000 ± 0%   -99.95% (p=0.000 n=10)
```

**Headline:** `Distance/near/n=1024` allocs **-99.97%** (110.8 MB/op → 36.0
KB/op, 6171 → 3 allocs/op); time -21% to -25% across sizes. Envelope
short-circuit for distant pairs deliberately deferred (backlog below).

## Macro workload table

`benchstat` on the 7 workloads in `bench/workloads.go`, baseline vs final:

```
$ benchstat bench/results/2026-07-10-baseline/macro.txt bench/results/2026-07-10-final/macro.txt
goos: linux
goarch: amd64
pkg: github.com/exergy-dev/go-topology-suite/bench
                          │       baseline        │                final                │
                          │         sec/op         │      sec/op       vs base            │
Buffer-32                                   816.4m ± 5%        109.6m ± 3%  -86.58% (p=0.000 n=10)
GeographicIntersection-32                   1.739m ± 2%        1.805m ± 2%   +3.80% (p=0.000 n=10)
IngestWKB-32                                1.358m ± 5%        1.403m ± 1%        ~ (p=0.105 n=10)
PairwiseIntersection-32                     92.10µ ± 1%        93.41µ ± 2%   +1.42% (p=0.007 n=10)
PointInPolygon-32                           29.60m ± 2%        29.97m ± 3%   +1.25% (p=0.043 n=10)
PointInPolygonPrepared-32                   15.27m ± 3%        15.56m ± 0%   +1.90% (p=0.000 n=10)
UnaryUnion-32                               42.11m ± 2%        42.31m ± 3%        ~ (p=0.912 n=10)
geomean                                     8.564m             6.540m       -23.64%
                          │       baseline        │                final                │
                          │       allocs/op        │     allocs/op      vs base           │
Buffer-32                                  309.4k ± 0%         309.5k ± 0%  +0.05% (p=0.000 n=10)
UnaryUnion-32                              332.6k ± 0%         332.7k ± 0%  +0.04% (p=0.000 n=10)
(remaining 5 workloads: allocs/op and B/op flat, ~0%, p ≥ 0.07 or exact ties)
```

Only `Buffer` moved outside noise: **-86.58%** time, from the R-tree
insert-path fix directly. The four workloads showing small (<4%) but
significant (p<0.05) *slower* moves don't touch any of the three optimized
paths at meaningful scale (no fragment-heavy overlay, no sequential R-tree
build, no nested-ring `Distance` visits) — read as WSL2 run-to-run noise
crossing significance at n=10, not a regression.

## Geographic overhead

From `overlay/geographic_bench_test.go`: `IntersectionGeographic` runs the
geoframe round trip (project → intersect → project back);
`IntersectionPlanarEquivalent` runs identical shapes with the CRS stripped,
isolating the projection cost.

```
$ benchstat geo_planar.txt geo_geographic.txt
            │ IntersectionPlanarEquivalent │        IntersectionGeographic        │
            │            sec/op            │   sec/op     vs base                 │
*/n=64-32                      10.15µ ± 2%   71.12µ ± 1%  +600.91% (p=0.000 n=10)
*/n=1024-32                    1.899m ± 1%   2.802m ± 1%   +47.54% (p=0.000 n=10)
```

Ratio: **~7.0x at n=64, ~1.48x at n=1024** — the geoframe round trip (frame
selection + two coordinate-projection passes) is close to a fixed per-call
cost, so it dominates at small n and amortizes as the planar op's own cost
grows with vertex count. Profiling attributes **87% of the geographic path's
cost to inherent projection trigonometry** (Transverse Mercator / polar LAEA
forward+inverse, `internal/geoframe`) and **~3.5% to the planar op itself**
at the small-n end. This is the disclosed cost of the correctness fix in
`CHANGELOG.md` ("overlay on geographic input now computes in a local
projected frame"), not a regression to chase — the alternative was degree-
space math that fix replaced for correctness.

## Refuted candidates

- **`planar.Orient` interface devirtualization** — zero `assertI2I`/indirect-call frames attributable to `Orient` across 8 profiled workloads; the compiler already devirtualizes the ~19 monomorphic call sites. No commit.
- **`orientCache` mutex removal** — `orientCache` (`kernel/planar/robust.go`) is absent from every profile taken this pass (off-hot-path: consulted only after a rare Shewchuk filter failure); nothing to gain by removing the guarding mutex. Left in place.
- **HPRtree `Query`→`QueryVisit` steering for `PointInPolygon` allocs** — the macro workload's unprepared path doesn't touch the HPRtree index at all; nothing to steer.
- **`mcindex_noder` scratch reuse** — measured at 0.2%–11% of total time depending on workload, real but small and workload-dependent, not an outsized single-site win; moved to backlog rather than implemented.
- **Optimistic tolerance-0 buffer first attempt** (2026-07-11) — skipping the snap-rounding fixpoint when the full-precision polygonize succeeds regressed `TestBufferJagged` buffer-5/buffer-10: jagged inputs are exactly where the digits=12 grid's stabilisation is load-bearing. Reverted; the snap path's noding cost was addressed inside the noder (monotone chains) instead.
- **Union double-overlay hypothesis** (2026-07-11) — `canonicalizeTouchingRings`' self-union re-run never fires on the star fixtures; the cost was the `needsCanonicalize` *check* itself (fixed above). Don't chase the re-run without a profile showing `OverlayPolygonalMixedDim` twice-deep.
- **Dropping the legacy negative-buffer pipeline** (2026-07-11 cleanup pass) — routing all insets through the polygonizer + hybrid validator fails immediately on `TestBufferNegativeGrowsHole` (10×10 outer, 4×4 hole, d=-1 → empty instead of area 28). The legacy offset+Difference path is not a quality guard but the only path handling typical hole-carrying insets; the polygonizer remains a fallback for the thin-parcel cases legacy collapses to empty (JTS TestBufferExternal2/Jagged/MitredJoin). Two-phase structure stays.

### 2026-07-11 cleanup pass

After the wave landed, a dedicated pass deleted everything it had
superseded, with the full gate suite (root tests, JTS corpus
8951/8942/9, `bench/conformance`, `TestCompareAgreement`) green after
each commit and benchstat pre/post confirming no timed compare
benchmark moved outside noise. Removed: the brute-force overlayng touch
checks and the unreachable disjoint-overlay subsystem (~450 lines,
`disjoint.go` gone entirely); `IndexedNoder` and `IntersectionAdder`
(overlayng now shares `noding.NodeAdaptive`'s MCIndexNoder selection —
measured -1.3% Intersection time, within noise elsewhere); the
single-value snap-rounding knobs (`SeedIntersections`, `MaxIter`);
test-only entry point `SnapRoundRings`.
Consolidated: the seven duplicated predicate method/free-function
orchestrations (generic `xxxWith` helpers, allocation-free), the two
figure-8 ring predicates, the two pairwise-fuse union loops, and
`dropPhantomSliverHoles`' nested closures. Net ≈ -1,350 lines.

One cleanup was itself refuted: deleting the production-dead `minArea`
branch from `polygonizeBufferWithFilter` reproducibly cost ~6% on
`Buffer/n=1024` — interleaved A/B bisection isolated it to that exact
deletion, comment-only edits moved nothing, and inlining decisions were
unchanged, so the mechanism is instruction-layout displacement of the
hot loops later in `buffer/polygonize.go`. The branch was reinstated
byte-for-byte with a warning comment. Treat single-digit `Buffer`
swings on this machine as layout-sensitive until proven algorithmic.

## Deferred backlog

- ~~**`PrepareBuild` vs simplefeatures**~~ — landed in the 2026-10-30 wave (sort-free packed segment tree; build 123µs → 12µs on M3, now ~4x faster than simplefeatures).
- ~~**Parallel cascaded `UnaryUnion`**~~ — landed earlier (`overlay/unary_union.go`, bounded fork-join, commit 56084e3); this entry was stale.
- **Envelope short-circuit for `measure.Distance` disjoint pairs** — explicitly left for later in the ring-traversal fix's commit message.
- **`relateng` geomB caching** — not profiled in depth this pass.
- **PGO** — needs a representative production profile corpus, not these synthetic fixtures, to be meaningful.
- **JTS-corpus-derived realistic fixtures** — corpus extraction was out of scope (tag-gated, ~150KB+ blobs with EPL/EDL provenance bookkeeping, not size-parameterizable); stars/grids stand in for now.
- ~~**DCEL arena allocation + twin-scan fix**~~ — landed in the 2026-07-11 wave (sort-based dedup, Out arena, dense ids, O(1) twin lookup was already in).
- ~~**`mcindex_noder` scratch reuse**~~ — superseded: the 2026-07-11 wave routed buffer noding through `MCIndexNoder` wholesale.

## Limitations

- **Synthetic fixtures only** — generated n-gons, jagged stars, overlapping grids (`internal/benchfix`), no real-world coastlines, cadastral slivers, or precision pathologies. Relative wins (e.g. -95% allocs on the star-intersection touch check) are believable because stars are specifically the adversarial case for fragment-count-driven cost; absolute numbers are not production throughput estimates.
- **Single machine, WSL2** — no cross-machine, cross-OS, or bare-metal-Linux confirmation exists here. Hybrid P/E-core scheduling and the WSL2 virtualization layer are both plausible sources of numbers that wouldn't reproduce identically elsewhere.
- **GEOS numbers include the cgo boundary and an old GEOS** — every GEOS column includes cgo call overhead and a context mutex, and reflects GEOS 3.10.2 (pre-3.12 RelateNG/OverlayNG work). Not disclosed per-cell in the tables; applies throughout.
- **The GEOS 3.10.2 SIGFPE** during the first final `compare.txt` capture didn't recur on rerun and affected nothing else across ~30 total `-count=10` runs — but it's a reminder the GEOS numbers depend on a native library whose stability this project doesn't control.
- **No CI performance gate** — this document is a point-in-time snapshot from one optimization pass; a future regression would not be caught automatically.

## Reproduction

Micro-benchmarks (main module):

```sh
go test -run '^$' -bench . -benchmem -count=10 ./buffer/... ./overlay/... \
  ./predicate/... ./prepare/... ./wkb/... ./wkt/... ./geojson/... \
  ./index/... ./measure/... ./validate/... ./simplify/... > micro.txt
```

Macro workloads (`bench` module):

```sh
go test -C bench -run '^$' -bench . -benchmem -count=10 . > macro.txt

# Or via the CLI (tabwriter table by default; standard bench lines with -benchfmt):
go run ./bench/cmd/gts-bench
go run ./bench/cmd/gts-bench -benchfmt
```

Cross-implementation comparison, without and with the GEOS adapter:

```sh
# gts + simplefeatures only (geos adapter compiles to an Available()=false stub)
go test -C bench -run '^$' -bench . -benchmem -count=10 ./compare/... > compare.txt

# gts + simplefeatures + GEOS (requires system libgeos-dev; cgo)
go test -C bench -tags geos -run '^$' -bench . -benchmem -count=10 ./compare/... > compare.txt
```

benchstat invocations used for every table above:

```sh
# cross-implementation comparison (pivots impl=gts/simplefeatures/geos into columns)
benchstat -col /impl bench/results/2026-07-10-final/compare.txt

# macro workloads, baseline vs final
benchstat bench/results/2026-07-10-baseline/macro.txt bench/results/2026-07-10-final/macro.txt

# a single optimization's before/after (example: the overlayng fix)
grep -E '^Benchmark(Intersection|Difference)/star/n=1024-' bench/results/2026-07-10-baseline/micro.txt > old.txt
grep -E '^Benchmark(Intersection|Difference)/star/n=1024-' bench/results/2026-07-10-final/micro.txt > new.txt
benchstat old.txt new.txt
```

JTS conformance baseline (must hold at 8951/8942/0 skipped; run alongside any optimization):

```sh
go test -tags=jts ./internal/jtstest/...
```
