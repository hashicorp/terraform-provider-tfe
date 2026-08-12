#!/usr/bin/env bash
# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0
#
# Usage:
#   ./scripts/generate-docs.sh
#   RESOURCE="resources/tfe_workspace.md" ./scripts/generate-docs.sh
#
# Environment variables:
#   RESOURCE          Optional glob pattern (relative to tfplugindocs output) to
#                     limit which generated files are written. Defaults to "**"
#                     (all files). Always pass as a quoted env var — never as a
#                     shell argument — to prevent the calling shell from expanding
#                     the glob before the script sees it.
#                     Example: RESOURCE="resources/tfe_workspace.md"
#                              RESOURCE="resources/*.md"
#   EXCEPTIONS_FILE   Path to JSON exceptions file.
#                     Default: examples/document_generation_exceptions.json
#
# The script generates docs via tfplugindocs, maps the new-layout output paths
# to the legacy website/docs layout (resources/ → r/, data-sources/ → d/,
# tfe_ prefix stripped, .md → .html.markdown), then writes the result to
# website/docs/. Files listed under "no_generate" in the exceptions file are
# never overwritten. Any file under website/docs/ (except cdktf/) that is no
# longer generated and not excepted is removed.

set -euo pipefail

[ -n "${BASH_VERSION:-}" ] || { echo "Run with bash"; exit 1; }
[ "${BASH_VERSION%%.*}" -ge 4 ] || { echo "Bash 4+ required (found $BASH_VERSION)"; exit 1; }

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROVIDER_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
WEBSITE_DOCS_DIR="${PROVIDER_DIR}/website/docs"

EXCEPTIONS_FILE="${EXCEPTIONS_FILE:-${PROVIDER_DIR}/examples/document_generation_exceptions.json}"

# pins to a version which has actions
TFPLUGINDOCS_CMD=(
  go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@v0.25.1-0.20260727142635-b766be144d67
  generate
  --provider-name tfe
)

# ---------------------------------------------------------------------------
# Temp dirs — all inside $tmp_root so cleanup is a single rm -rf
# ---------------------------------------------------------------------------
tmp_root="$(mktemp -d "${TMPDIR:-/tmp}/tfplugindocs.XXXXXX")"
schema_work_dir="$(mktemp -d "$tmp_root/schema-build.XXXXXX")"
augmented_schema_file="$(mktemp "$tmp_root/provider-schema.augmented.XXXXXX.json")"
# rendered_tmp_dir must be relative to PROVIDER_DIR since --rendered-website-dir
# is always resolved relative to --provider-dir by tfplugindocs.
rendered_tmp_dir="$(mktemp -d "${PROVIDER_DIR}/.tfplugindocs-rendered.XXXXXX")"

cleanup() {
  rm -rf "$tmp_root"
  # rendered_tmp_dir lives inside PROVIDER_DIR (not tmp_root) so clean separately
  [[ -n "${rendered_tmp_dir:-}" ]] && rm -rf "$rendered_tmp_dir"
}
trap cleanup EXIT INT TERM

# ---------------------------------------------------------------------------
# Build augmented provider schema
# ---------------------------------------------------------------------------
generate_augmented_schema() {
  local provider_dir provider_name provider_short_name os_arch plugin_dir provider_binary jq_tmp_file

  provider_dir="$(pwd)"
  provider_name="$(basename "$provider_dir")"
  provider_short_name="${provider_name#terraform-provider-}"

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
  # deprecation_message text. For top-level components (identified by having an
  # "attributes" key — i.e. the .block node), prepend a warning admonition to
  # the description. For nested attributes/blocks, append a plain inline notice.
  jq_tmp_file="$(mktemp "$tmp_root/provider-schema.filtered.XXXXXX.json")"
  jq '
    def augment:
      if type == "object" then
        (
          if (.deprecated == true and (.deprecation_message | type) == "string") then
            (.deprecation_message | gsub("^\\s+|\\s+$"; "")) as $msg
            | if ($msg | length) > 0 then
                if (has("attributes") or has("block_types")) then
                  ("~> **Deprecated:** " + $msg) as $notice
                  | .description = (
                      if ((.description | type) == "string" and (.description | length) > 0) then
                        $notice + "\n\n" + .description
                      else
                        $notice
                      end
                    )
                  | .description_kind = "markdown"
                else
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
                end
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

# ---------------------------------------------------------------------------
# Path mapping: tfplugindocs new layout → website/docs legacy layout
#
#   resources/tfe_X.md           → r/X.html.markdown
#   data-sources/tfe_X.md        → d/X.html.markdown
#   ephemeral-resources/tfe_X.md → ephemeral-resources/X.html.markdown
#   actions/tfe_X.md             → actions/X.html.markdown
#   index.md                     → index.html.markdown
#
# Returns the mapped path in $MAPPED, or an empty string if no mapping applies.
# ---------------------------------------------------------------------------
map_generated_path() {
  local gen_path="$1"
  local dir base stem

  dir="$(dirname "$gen_path")"
  base="$(basename "$gen_path" .md)"
  # Strip the provider prefix (tfe_) from the filename
  stem="${base#tfe_}"

  # Any new directories will need to be added here
  case "$dir" in
    resources)           MAPPED="r/${stem}.html.markdown" ;;
    data-sources)        MAPPED="d/${stem}.html.markdown" ;;
    ephemeral-resources) MAPPED="ephemeral-resources/${stem}.html.markdown" ;;
    actions)             MAPPED="actions/${stem}.html.markdown" ;;
    .)
      if [[ "$gen_path" == "index.md" ]]; then
        MAPPED="index.html.markdown"
      else
        MAPPED=""
      fi
      ;;
    *) MAPPED="" ;;
  esac
}

# ---------------------------------------------------------------------------
# Load exceptions
# ---------------------------------------------------------------------------
NO_GENERATE=()
if [[ -f "$EXCEPTIONS_FILE" ]]; then
  while IFS= read -r entry; do
    NO_GENERATE+=("$entry")
  done < <(jq -r '.no_generate[]?' "$EXCEPTIONS_FILE")
else
  echo "Warning: exceptions file not found at $EXCEPTIONS_FILE — proceeding without exceptions" >&2
fi

is_excepted() {
  local path="$1"
  local entry
  for entry in "${NO_GENERATE[@]}"; do
    [[ "$entry" == "$path" ]] && return 0
  done
  return 1
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

# Optional glob filter on generated output paths (default: all files).
# Read from env var only — never a positional arg — so the calling shell cannot
# expand the glob pattern before this script sees it.
resource_input="${RESOURCE:-**}"

generate_augmented_schema

# Pass rendered-website-dir as relative to PROVIDER_DIR (tfplugindocs resolves it that way)
rendered_website_rel="${rendered_tmp_dir#${PROVIDER_DIR}/}"
"${TFPLUGINDOCS_CMD[@]}" \
  --provider-dir "$PROVIDER_DIR" \
  --providers-schema "$augmented_schema_file" \
  --rendered-website-dir "$rendered_website_rel"

# Collect all generated files matching the requested pattern.
# find -path handles ** as a path wildcard correctly; for the default "**" case
# we skip the filter entirely since all files are wanted.
generated_paths=()
while IFS= read -r match; do
  generated_paths+=("$match")
done < <(
  if [[ "$resource_input" == "**" ]]; then
    find "$rendered_tmp_dir" -type f \
      | sed "s|^${rendered_tmp_dir}/||" \
      | sort -u
  else
    find "$rendered_tmp_dir" -type f -path "${rendered_tmp_dir}/${resource_input}" \
      | sed "s|^${rendered_tmp_dir}/||" \
      | sort -u
  fi
)

if [[ "${#generated_paths[@]}" -eq 0 ]]; then
  echo "No generated docs matched pattern: $resource_input"
  echo ""
  echo "Run 'make generate' to regenerate all docs, or set RESOURCE to a valid"
  echo "path pattern relative to the tfplugindocs output (e.g. RESOURCE=\"resources/tfe_workspace.md\")."
  exit 1
fi

# Track which website/docs destination paths were written or excepted,
# for the stale-file removal pass at the end.
declare -A touched_dest_paths=()

for gen_path in "${generated_paths[@]}"; do
  MAPPED=""
  map_generated_path "$gen_path"

  if [[ -z "$MAPPED" ]]; then
    echo "skipped (no mapping): $gen_path"
    continue
  fi

  if is_excepted "$MAPPED"; then
    echo "skipped (excepted): $MAPPED"
    touched_dest_paths["$MAPPED"]=1
    continue
  fi

  dest="${WEBSITE_DOCS_DIR}/${MAPPED}"
  mkdir -p "$(dirname "$dest")"
  cp "$rendered_tmp_dir/$gen_path" "$dest"
  echo "updated website/docs/$MAPPED"
  touched_dest_paths["$MAPPED"]=1
done

# ---------------------------------------------------------------------------
# Stale file removal — delete any file in the managed directories that was
# neither produced by this run nor listed as an exception.
# Covers all of website/docs/ except cdktf/ (deprecated).
# Only runs when the full set was generated (no RESOURCE filter in effect).
# ---------------------------------------------------------------------------
if [[ "$resource_input" == "**" ]]; then
  while IFS= read -r existing; do
    rel="${existing#${WEBSITE_DOCS_DIR}/}"
    if [[ -z "${touched_dest_paths[$rel]+x}" ]]; then
      echo "removed stale: website/docs/$rel"
      rm "$existing"
    fi
  done < <(
    find "${WEBSITE_DOCS_DIR}" -type f \
      ! -path "${WEBSITE_DOCS_DIR}/cdktf/*" \
      2>/dev/null | sort
  )
fi
