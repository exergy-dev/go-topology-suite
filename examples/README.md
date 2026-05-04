# Examples

Self-contained programs that exercise the most common parts of the
go-topology-suite API. Each subdirectory is a runnable `package main`.

| Example | What it shows |
|---|---|
| [`predicates/`](./predicates) | Decode WKT, run spatial predicates (`Intersects`, `Contains`, `Touches`). |
| [`overlay/`](./overlay) | Boolean operations on polygons: `Union`, `Intersection`, `Difference`, `SymmetricDifference`. |
| [`buffer-geojson/`](./buffer-geojson) | Buffer a geometry by a distance and emit GeoJSON. |
| [`transform/`](./transform) | Reproject a geometry between coordinate reference systems. |
| [`spatial-index/`](./spatial-index) | Bulk-load an R-tree and query for envelope-overlapping items. |
| [`validate/`](./validate) | Validate a geometry, then `Fix` an invalid input into a valid one. |

Run any example with:

```sh
go run ./examples/<name>
```
