#!/usr/bin/env bash
# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0
#
# Usage:
#   ./scripts/check-docs-current.sh
#
# Verifies that website/docs/ is up to date with what scripts/generate-docs.sh
# would produce. Files listed under "no_generate" in
# examples/document_generation_exceptions.json are excluded from the check.
#
# Exit codes:
#   0 - Docs are current
#   1 - Docs are stale

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROVIDER_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
EXCEPTIONS_FILE="${PROVIDER_DIR}/examples/document_generation_exceptions.json"

# Build pathspec exclusions for hand-authored files that are never generated.
exclusions=()
if [[ -f "$EXCEPTIONS_FILE" ]]; then
  while IFS= read -r entry; do
    exclusions+=(":!website/docs/${entry}")
  done < <(jq -r '.no_generate[]?' "$EXCEPTIONS_FILE")
fi

stale_files=()
while IFS= read -r f; do
  stale_files+=("$f")
done < <(
  git -C "$PROVIDER_DIR" diff --name-only -- \
    website/docs/r \
    website/docs/d \
    website/docs/ephemeral-resources \
    website/docs/index.html.markdown \
    "${exclusions[@]}"
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
  echo "To regenerate a single resource: make generate RESOURCE=\"resources/tfe_workspace.md\""
  echo "Files excluded from generation: examples/document_generation_exceptions.json"
  exit 1
fi

echo ""
echo "Result: PASSED"
