#!/usr/bin/env bash
# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0
#
# Usage:
#   ./scripts/schema-diff.sh [--base REF] [--allow 'PATH=FROM->TO']... NAME...
#
# Compares the protocol-level schema of resources/data sources between a base
# commit (default: merge-base of HEAD and origin/main) and the working tree.
# Used during SDKv2 -> plugin framework migrations to prove the schema users
# and Terraform see did not change. See docs/framework-migration.md.
#
# NAME is a resource type (tfe_team) or "data." + data source type
# (data.tfe_team).
#
# --allow accepts one known, reviewed difference, written exactly as reported:
#   --allow 'resource.tfe_team.block.organization_access.max_items=1->0'
#
# The schema is read in-process by TestSchemaDump
# (internal/provider/schema_dump_test.go); the working tree's copy of that
# file is used on both sides so the output format is identical.
#
# Exit codes:
#   0 - No differences (other than allowed ones)
#   1 - Differences found
#   2 - Usage or execution error

set -euo pipefail

usage() {
  sed -n '5,14p' "$0" | sed 's/^# \{0,1\}//'
  exit 2
}

base=""
allows=()
names=()
while [ $# -gt 0 ]; do
  case "$1" in
    --base) base="${2:?--base needs a ref}"; shift 2 ;;
    --allow) allows+=("${2:?--allow needs a value}"); shift 2 ;;
    -h|--help) usage ;;
    -*) echo "unknown flag: $1" >&2; usage ;;
    *) names+=("$1"); shift ;;
  esac
done
[ ${#names[@]} -gt 0 ] || usage

repo="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo"

if [ -z "$base" ]; then
  upstream="origin/main"
  git rev-parse --verify --quiet "$upstream" >/dev/null || upstream="main"
  base="$(git merge-base HEAD "$upstream")"
fi
base_sha="$(git rev-parse --verify "$base^{commit}")"

work="$(mktemp -d "${TMPDIR:-/tmp}/schema-diff.XXXXXX")"
worktree="$work/base"
cleanup() {
  git worktree remove --force "$worktree" >/dev/null 2>&1 || true
  rm -rf "$work"
}
trap cleanup EXIT

dump_list="$(IFS=,; echo "${names[*]}")"

dump() { # dir out
  (cd "$1" && SCHEMA_DUMP="$dump_list" SCHEMA_DUMP_OUT="$2" \
    go test ./internal/provider -run '^TestSchemaDump$' -count=1 >"$2.log" 2>&1) || {
    echo "schema dump failed in $1:" >&2
    cat "$2.log" >&2
    exit 2
  }
}

echo "base:  ${base_sha:0:12} ($base)"
echo "after: working tree"
echo

git worktree add --detach --quiet "$worktree" "$base_sha"
cp internal/provider/schema_dump_test.go "$worktree/internal/provider/schema_dump_test.go"

dump "$worktree" "$work/before.txt"
dump "$repo" "$work/after.txt"

allow_file="$work/allow.txt"
: >"$allow_file"
for a in ${allows[@]+"${allows[@]}"}; do
  printf '%s\n' "$a" >>"$allow_file"
done

awk -v allowfile="$allow_file" '
  function split_line(line, kv,   i) {
    i = index(line, " = ")
    kv["path"] = substr(line, 1, i - 1)
    kv["value"] = substr(line, i + 3)
  }
  BEGIN {
    while ((getline a < allowfile) > 0) {
      eq = index(a, "="); arrow = 0
      for (i = length(a) - 1; i > eq; i--) if (substr(a, i, 2) == "->") { arrow = i; break }
      if (eq == 0 || arrow == 0) { print "invalid --allow: " a > "/dev/stderr"; exit 2 }
      allow[substr(a, 1, eq - 1) SUBSEP substr(a, eq + 1, arrow - eq - 1) SUBSEP substr(a, arrow + 2)] = 1
    }
  }
  FNR == NR { split_line($0, kv); before[kv["path"]] = kv["value"]; paths[kv["path"]] = 1; next }
  { split_line($0, kv); after[kv["path"]] = kv["value"]; paths[kv["path"]] = 1 }
  END {
    n = asorti_compat(paths, sorted)
    for (i = 1; i <= n; i++) {
      p = sorted[i]
      b = (p in before) ? before[p] : "<absent>"
      a = (p in after) ? after[p] : "<absent>"
      if (b == a) continue
      key = p SUBSEP b SUBSEP a
      if (key in allow) { printf "allowed  %s: %s -> %s\n", p, b, a; used[key] = 1; continue }
      printf "CHANGED  %s: %s -> %s\n", p, b, a
      diffs++
    }
    for (k in allow) if (!(k in used)) {
      split(k, parts, SUBSEP)
      printf "unused --allow %s=%s->%s (no such difference)\n", parts[1], parts[2], parts[3]
      diffs++
    }
    if (diffs > 0) { printf "\n%d unexpected difference(s)\n", diffs; exit 1 }
    print "no unexpected differences"
  }
  # portable sort of array keys (BSD awk has no asorti)
  function asorti_compat(src, dst,   k, n, i, j, t) {
    n = 0
    for (k in src) dst[++n] = k
    for (i = 2; i <= n; i++) { t = dst[i]; for (j = i - 1; j > 0 && dst[j] > t; j--) dst[j + 1] = dst[j]; dst[j + 1] = t }
    return n
  }
' "$work/before.txt" "$work/after.txt"
