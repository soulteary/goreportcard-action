#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
profile="${COVERAGE_PROFILE:-coverage.out}"
go test -race -covermode=atomic -coverpkg=./... -coverprofile="$profile" ./...
go tool cover -func="$profile"
awk 'NR > 1 { key=$1; statements[key]=$2; if ($3 > 0) covered[key]=1 }
END {
  for (key in statements) { total+=statements[key]; if (covered[key]) hit+=statements[key] }
  if (total == 0) { print "No coverage statements found"; exit 1 }
  coverage=100*hit/total
  printf "Total statement coverage: %.2f%% (minimum 95%%)\n", coverage
  if (coverage < 95) exit 1
}' "$profile"
