#!/usr/bin/env bash
# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0
#
# Usage:
#   ./scripts/check-docs-current.sh
#
# This script is primarily made for CI and assumes that all changes have already been committed.
# Runs generate-docs.sh, then verifies that website/docs/ is up to date with the git diff.
# Beware that doing so may destroy manually written documentation which has not been committed nor excluded. This is intended behaviour.
# Files under website/docs/cdktf/ and files listed under "no_generate" in examples/document_generation_exceptions.json are excluded from the check.
#
# Exit codes:
#   0 - Docs are current
#   1 - Docs are stale

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROVIDER_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
EXCEPTIONS_FILE="${PROVIDER_DIR}/examples/document_generation_exceptions.json"

# Regenerate docs before diffing.
"${SCRIPT_DIR}/generate-docs.sh"

# Build pathspec exclusions: cdktf (deprecated) and files listed in exceptions.
exclusions=(":!website/docs/cdktf/")
if [[ -f "$EXCEPTIONS_FILE" ]]; then
  while IFS= read -r entry; do
    exclusions+=(":!website/docs/${entry}")
  done < <(jq -r '.no_generate[]?' "$EXCEPTIONS_FILE")
fi

# Build the ls-files exclude args (same set, different flag syntax).
ls_excludes=("website/docs/cdktf/")
if [[ -f "$EXCEPTIONS_FILE" ]]; then
  while IFS= read -r entry; do
    ls_excludes+=("website/docs/${entry}")
  done < <(jq -r '.no_generate[]?' "$EXCEPTIONS_FILE")
fi

stale_files=()
while IFS= read -r f; do
  stale_files+=("$f")
done < <(
  {
    # Modified/deleted tracked files.
    git -C "$PROVIDER_DIR" diff --name-only -- \
      website/docs \
      "${exclusions[@]}"
    # Untracked new files (e.g. a generated file never previously committed).
    git -C "$PROVIDER_DIR" ls-files --others --exclude-standard -- \
      website/docs \
      "${ls_excludes[@]}"
  } | sort -u
)

echo ""
echo "========================================"

if [[ "${#stale_files[@]}" -gt 0 ]]; then
  echo ""
  echo "Stale docs (${#stale_files[@]}):"
  for f in "${stale_files[@]}"; do
    echo "  - $f"
  done
  echo ""
  echo "Result: FAILED"
  echo ""
  echo "Run 'make generate' locally and commit the result."
  echo "Files excluded from generation: examples/document_generation_exceptions.json"
  exit 1
fi

echo ""
echo "Result: PASSED"
