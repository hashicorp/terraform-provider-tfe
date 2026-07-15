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

tmp_root="$(mktemp -d "${TMPDIR:-/tmp}/tfplugindocs.XXXXXX")"
schema_work_dir="$(mktemp -d "$tmp_root/schema-build.XXXXXX")"
augmented_schema_file="$(mktemp "$tmp_root/provider-schema.augmented.XXXXXX.json")"
rendered_tmp_dir=""

cleanup() {
  rm -rf "$tmp_root"
  if [[ -n "$rendered_tmp_dir" ]]; then
    rm -rf "$rendered_tmp_dir"
  fi
}
trap cleanup EXIT

generate_augmented_schema() {
  local provider_dir provider_name provider_short_name os_arch plugin_dir provider_binary jq_tmp_file

  provider_dir="$(pwd)"
  provider_name="$(basename "$provider_dir")"
  provider_short_name="${provider_name#terraform-provider-}"

  # Build the provider in a temporary plugin directory so Terraform can load it
  # without fetching from the network.
  os_arch="$(go env GOOS)_$(go env GOARCH)"
  plugin_dir="$schema_work_dir/plugins/registry.terraform.io/hashicorp/$provider_short_name/0.0.1/$os_arch"
  mkdir -p "$plugin_dir"
  provider_binary="$plugin_dir/terraform-provider-$provider_short_name"
  (cd "$provider_dir" && go build -o "$provider_binary")

  cat > "$schema_work_dir/provider.tf" <<EOF
provider "$provider_short_name" {
}
EOF

  (cd "$schema_work_dir" && terraform init -get=false -plugin-dir=./plugins > /dev/null)
  (cd "$schema_work_dir" && terraform providers schema -json > "$augmented_schema_file")

  # tfplugindocs currently marks deprecated fields but does not render
  # deprecation_message text. Append a short deprecation notice into each
  # deprecated node's description so generated docs include the guidance.
  jq_tmp_file="$(mktemp "$tmp_root/provider-schema.filtered.XXXXXX.json")"
  jq '
    def augment:
      if type == "object" then
        (
          if (.deprecated == true and (.deprecation_message | type) == "string") then
            (.deprecation_message | gsub("^\\s+|\\s+$"; "")) as $msg
            | if ($msg | length) > 0 then
                ("**Deprecation notes**: " + $msg) as $notice
                | .description = (
                    if ((.description | type) == "string") then
                      if (.description | endswith($notice)) then
                        .description
                      elif (.description | length) > 0 then
                        .description + " " + $notice
                      else
                        $notice
                      end
                    else
                      $notice
                    end
                  )
                | .description_kind = (.description_kind // "plain")
              else
                .
              end
          else
            .
          end
        )
        | with_entries(.value |= augment)
      elif type == "array" then
        map(augment)
      else
        .
      end;
    augment
  ' "$augmented_schema_file" > "$jq_tmp_file"
  mv "$jq_tmp_file" "$augmented_schema_file"
}

# Input RESOURCE is an optional glob pattern to filter which generated docs are copied to docs/.
# Defaults to "**" (all files). Example: "resources/tfe_workspace.md" or "resources/*.md".
resource_input="${1:-**}"

generate_augmented_schema

# Run the targeted docs generation into the temp folder
rendered_tmp_dir="$(mktemp -d "./.tfplugindocs-rendered.XXXXXX")"
"${TFPLUGINDOCS_CMD[@]}" \
  --providers-schema "$augmented_schema_file" \
  --rendered-website-dir "$rendered_tmp_dir"

matches=()
while IFS= read -r match; do
  matches+=("$match")
done < <(
  (
    # Enable globstar and nullglob to allow recursive matching
    # and ignore non-matching patterns
    shopt -s globstar nullglob

    cd "$rendered_tmp_dir"
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
  cp "$rendered_tmp_dir/$resolved" "docs/$resolved"
  echo "updated docs/$resolved"
done
