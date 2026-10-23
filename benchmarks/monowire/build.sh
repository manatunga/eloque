#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"

for workload in call_loop arithmetic branch recursion; do
  echo "Building $workload..."
  (cd "$workload/go" && go build -o monowire-go main.go)
  (cd "$workload/rust" && rustc main.rs -C opt-level=3 -o monowire-rust)
  (cd "$workload/zig" && zig build-exe main_switch.zig -O ReleaseFast -lc -femit-bin=monowire-switch)
  (cd "$workload/zig" && zig build-exe main_labeled.zig -O ReleaseFast -lc -femit-bin=monowire-labeled)
done
echo "All 16 benchmark binaries built."
