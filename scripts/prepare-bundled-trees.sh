#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
output="$repo_root/internal/ipm/bundled_trees_generated.go"
trees_dir="$repo_root/trees/main"

python3 - "$trees_dir" "$output" <<'PY'
from __future__ import annotations

import json
import pathlib
import sys

trees_dir = pathlib.Path(sys.argv[1])
output = pathlib.Path(sys.argv[2])
repo_root = output.parents[2]

entries: list[tuple[str, str]] = []
if trees_dir.is_dir():
    for path in sorted(p for p in trees_dir.rglob("*") if p.is_file()):
        rel = path.relative_to(repo_root).as_posix()
        entries.append((rel, path.read_text(encoding="utf-8")))

with output.open("w", encoding="utf-8") as fh:
    fh.write("package ipm\n\n")
    fh.write("var bundledTreeFiles = map[string]string{\n")
    for rel, content in entries:
        fh.write(f"\t{json.dumps(rel)}: {json.dumps(content)},\n")
    fh.write("}\n")
PY
