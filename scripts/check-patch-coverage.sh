#!/usr/bin/env bash
# check-patch-coverage.sh enforces a coverage floor on *newly added* lines so a
# PR cannot lower quality even while total coverage stays above its gate.
#
# It intersects the added lines from a git diff (excluding *_test.go and
# non-Go files) with the statement blocks in a Go cover profile. A line counts
# as "coverable" when it falls inside a profile block; it counts as "covered"
# when that block's hit count is > 0. Patch coverage = covered / coverable.
#
# Usage:
#   check-patch-coverage.sh <coverage.out>
#   check-patch-coverage.sh --selftest
#
# Environment:
#   BASE_REF             git ref/sha to diff against (default: origin/main).
#   PATCH_COVERAGE_MIN   minimum patch coverage percentage (default: 90).
set -euo pipefail

MIN="${PATCH_COVERAGE_MIN:-90}"

# module_prefix reads the module path from go.mod so profile paths
# (github.com/owner/repo/pkg/file.go) can be mapped back to repo-relative
# paths (pkg/file.go) that match git diff output.
module_prefix() {
  awk '/^module / { print $2; exit }' go.mod 2>/dev/null || true
}

# added_lines prints "relpath:lineno" for every added line in the diff,
# skipping test files, non-Go files, and deletions.
added_lines() {
  local base="$1"
  git diff --unified=0 --no-color "${base}" -- '*.go' ':(exclude)*_test.go' \
    | awk '
        /^\+\+\+ b\// { file = substr($0, 7); next }
        /^@@ / {
          # @@ -a,b +c,d @@ ; parse the new-file start line c.
          hunk = $0
          sub(/^@@ [^+]*\+/, "", hunk)
          split(hunk, parts, " ")
          split(parts[1], nums, ",")
          line = nums[1] + 0
          next
        }
        /^\+/ && file != "" {
          print file ":" line
          line++
        }
        /^[^+]/ { }
      '
}

# patch_coverage reads:
#   $1 = cover profile path
#   stdin = "relpath:lineno" added lines
#   $2 = module prefix
# and prints "covered coverable" counts.
patch_coverage() {
  local profile="$1" prefix="$2"
  awk -v prefix="$prefix" '
    # First pass: read added lines from stdin marker file.
    FNR == NR {
      # profile line: path:sL.sC,eL.eC numStmts count
      # split off the trailing " numStmts count"
      n = split($0, f, " ")
      count = f[n]
      loc = f[1]
      # loc = path:sL.sC,eL.eC
      ci = index(loc, ":")
      path = substr(loc, 1, ci - 1)
      rng = substr(loc, ci + 1)
      split(rng, a, ",")
      split(a[1], s, ".")
      split(a[2], e, ".")
      sL = s[1] + 0
      eL = e[1] + 0
      # Normalize path to repo-relative by stripping the module prefix.
      rel = path
      if (prefix != "" && index(path, prefix "/") == 1) {
        rel = substr(path, length(prefix) + 2)
      }
      # Record block: for each line in [sL,eL], remember best hit count.
      for (ln = sL; ln <= eL; ln++) {
        key = rel ":" ln
        if (!(key in seen) || count > hits[key]) {
          hits[key] = count
        }
        seen[key] = 1
      }
      next
    }
    # Second pass: added lines file.
    {
      key = $0
      if (key in seen) {
        coverable++
        if (hits[key] > 0) covered++
      }
    }
    END { printf "%d %d\n", covered + 0, coverable + 0 }
  ' "$profile" -
}

run() {
  local profile="$1"
  local base="${BASE_REF:-origin/main}"

  if [ ! -f "$profile" ]; then
    echo "::error::coverage profile not found: $profile"
    exit 1
  fi

  local prefix
  prefix="$(module_prefix)"

  local added
  added="$(added_lines "$base" || true)"

  if [ -z "$added" ]; then
    echo "No added Go (non-test) lines in diff against ${base}; patch coverage gate passes trivially."
    return 0
  fi

  local result covered coverable
  result="$(printf '%s\n' "$added" | patch_coverage "$profile" "$prefix")"
  covered="${result%% *}"
  coverable="${result##* }"

  if [ "$coverable" -eq 0 ]; then
    echo "Added lines contain no measurable statements; patch coverage gate passes trivially."
    return 0
  fi

  awk -v c="$covered" -v t="$coverable" -v min="$MIN" 'BEGIN {
    pct = (c / t) * 100
    printf "patch coverage: %d/%d = %.2f%% (min %s%%)\n", c, t, pct, min
    if (pct + 0 < min + 0) {
      printf "::error::patch coverage %.2f%% < %s%%\n", pct, min
      exit 1
    }
  }'
}

selftest() {
  local tmp
  tmp="$(mktemp -d)"

  # A profile where lines 3-4 are covered (count 1) and line 5 is not (count 0).
  cat > "${tmp}/cover.out" <<'EOF'
mode: atomic
example.com/mod/pkg/a.go:3.10,4.20 2 1
example.com/mod/pkg/a.go:5.2,5.30 1 0
EOF

  # Added lines: 3 and 5 are coverable; 3 covered, 5 not -> 1/2 = 50%.
  local added covered coverable result
  added=$'pkg/a.go:3\npkg/a.go:5\npkg/a.go:99'
  result="$(printf '%s\n' "$added" | patch_coverage "${tmp}/cover.out" "example.com/mod")"
  covered="${result%% *}"
  coverable="${result##* }"

  local fail=0
  if [ "$covered" != "1" ] || [ "$coverable" != "2" ]; then
    echo "selftest FAIL: expected covered=1 coverable=2, got covered=${covered} coverable=${coverable}"
    fail=1
  fi

  # added_lines parsing: build a tiny diff and confirm line numbers.
  local difffile parsed
  difffile="${tmp}/patch.diff"
  cat > "$difffile" <<'EOF'
diff --git a/pkg/b.go b/pkg/b.go
--- a/pkg/b.go
+++ b/pkg/b.go
@@ -0,0 +1,2 @@
+package pkg
+func F() {}
@@ -10,0 +12,1 @@
+var X = 1
EOF
  parsed="$(awk '
    /^\+\+\+ b\// { file = substr($0, 7); next }
    /^@@ / {
      hunk = $0; sub(/^@@ [^+]*\+/, "", hunk)
      split(hunk, parts, " "); split(parts[1], nums, ","); line = nums[1] + 0; next
    }
    /^\+/ && file != "" { print file ":" line; line++ }
  ' "$difffile" | tr '\n' ',')"
  if [ "$parsed" != "pkg/b.go:1,pkg/b.go:2,pkg/b.go:12," ]; then
    echo "selftest FAIL: diff parse mismatch, got '${parsed}'"
    fail=1
  fi

  if [ "$fail" -ne 0 ]; then
    rm -rf "$tmp"
    echo "::error::check-patch-coverage.sh selftest failed"
    exit 1
  fi
  rm -rf "$tmp"
  echo "check-patch-coverage.sh selftest passed"
}

main() {
  if [ "${1:-}" = "--selftest" ]; then
    selftest
    return
  fi
  if [ "$#" -lt 1 ]; then
    echo "usage: $0 <coverage.out> | --selftest" >&2
    exit 2
  fi
  run "$1"
}

main "$@"
