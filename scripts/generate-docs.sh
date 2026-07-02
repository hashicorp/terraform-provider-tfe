#!/usr/bin/env bash
# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0

set -euo pipefail

[ -n "${BASH_VERSION:-}" ] || { echo "Run with bash"; exit 1; }
[ "${BASH_VERSION%%.*}" -ge 4 ] || { echo "Bash 4+ required (found $BASH_VERSION)"; exit 1; }

TFPLUGINDOCS_CMD=(
  go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@v0.25.0
  generate
  --provider-name tfe
)

# Input RESOURCE is optional. If provided, it will be used to filter the generated docs.
# RESOURCE is expected to be a glob pattern, e.g. "docs/resources/aws_instance.md" or "docs/resources/*.md".
resource_input="${1:-}"

# No RESOURCE provided: generate full docs in-place.
if [[ -z "$resource_input" ]]; then
  "${TFPLUGINDOCS_CMD[@]}"
  exit 0
fi

# Create a temp folder to generate into
# Clean up on exit
tmp_dir=".tfplugindocs-tmp"
cleanup() {
  rm -rf "$tmp_dir"
}
trap cleanup EXIT

# Clean up and start fresh in case there is an existing temp folder
rm -rf "$tmp_dir"
mkdir -p "$tmp_dir"

# Run the targeted docs generation into the temp folder
"${TFPLUGINDOCS_CMD[@]}" --rendered-website-dir "$tmp_dir"

matches=()
while IFS= read -r match; do
  matches+=("$match")
done < <(
  (
    # Enable globstar and nullglob to allow recursive matching
    # and ignore non-matching patterns
    shopt -s globstar nullglob

    cd "$tmp_dir"
    for file in $resource_input; do
      [[ -f "$file" ]] && printf '%s\n' "$file"
    done
  ) | sort -u
)

if [[ "${#matches[@]}" == "0" ]]; then
  echo "No generated docs matched RESOURCE=$resource_input"
  exit 1
fi

# Copy the matched files from the temp folder to the docs folder
for resolved in "${matches[@]}"; do
  mkdir -p "docs/$(dirname "$resolved")"
  cp "$tmp_dir/$resolved" "docs/$resolved"
  echo "updated docs/$resolved"
done