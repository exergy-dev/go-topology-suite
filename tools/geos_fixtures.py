#!/usr/bin/env python3
"""Developer-only GEOS fixture helper.

This script is intentionally not used by runtime code or Go tests. It prints
DE-9IM and operation output for ad hoc WKT pairs when Shapely/GEOS is present.
"""

import argparse
import json
import sys

from shapely import wkt


def operation_result(a, b, operation):
    if operation == "relate":
        return a.relate(b)
    if operation == "intersection":
        return a.intersection(b).wkt
    if operation == "union":
        return a.union(b).wkt
    if operation == "difference":
        return a.difference(b).wkt
    if operation == "symdifference":
        return a.symmetric_difference(b).wkt
    raise ValueError(f"unsupported operation: {operation}")


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("a", help="WKT for geometry A")
    parser.add_argument("b", help="WKT for geometry B")
    parser.add_argument(
        "--operation",
        choices=("intersection", "union", "difference", "symdifference", "relate"),
        default="relate",
    )
    parser.add_argument("--name", default="", help="optional fixture name")
    parser.add_argument("--source", default="GEOS", help="fixture source note")
    parser.add_argument(
        "--json",
        action="store_true",
        help="print a JSON fixture record instead of only the expected value",
    )
    args = parser.parse_args()

    a = wkt.loads(args.a)
    b = wkt.loads(args.b)
    expected = operation_result(a, b, args.operation)

    if args.json:
        key = "expected_de9im" if args.operation == "relate" else "expected_wkt"
        print(
            json.dumps(
                {
                    "name": args.name,
                    "operation": args.operation,
                    "a": args.a,
                    "b": args.b,
                    key: expected,
                    "source": args.source,
                },
                indent=2,
                sort_keys=True,
            )
        )
    else:
        print(expected)
    return 0


if __name__ == "__main__":
    sys.exit(main())
