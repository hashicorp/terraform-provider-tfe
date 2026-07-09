#!/usr/bin/env bash
# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0
#
# Usage:
#   ./scripts/validate-schema-descriptions.sh [-o <json output file; ./missing_descriptions.json>] [-e <exceptions file; examples/error_exceptions.json>] [--help]
#
# Exit codes:
#  0 - Complete success
#  1 - Generic errors (including command line args)
#  3 - Warning: stale entries found in no_description_required
#  5 - Errors found: one or more components, attributes, or blocks are missing descriptions
#  6 - Required commands (terraform, jq, go) not found
#  7 - Provider directory not found or its schema could not be generated


# Crash on error
set -e

# Variables with defaults
SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]}")"
PROVIDER_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
EXCEPTIONS_FILE="${PROVIDER_DIR}/examples/error_exceptions.json"
JSON_OUTPUT_NAME="missing_descriptions.json"

# Arg parsing
while [[ $# -gt 0 ]]; do
    case $1 in
        -o|--output)
            JSON_OUTPUT_NAME="$2"
            shift 2
            ;;
        -e|--exceptions)
            EXCEPTIONS_FILE="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Validates that every provider schema attribute and block has a non-empty description."
            echo ""
            echo "Options:"
            echo "  -o, --output FILE        Output JSON file for failures (default: ./missing_descriptions.json)"
            echo "  -e, --exceptions FILE    JSON file with description exceptions (default: examples/error_exceptions.json)"
            echo "  -h, --help               Show this help message"
            echo ""
            echo "Exit Codes:"
            echo "  0 - Complete success"
            echo "  1 - Generic errors (including command line args)"
            echo "  3 - Warning: stale entries found in no_description_required"
            echo "  5 - Errors found: one or more components, attributes, or blocks are missing descriptions"
            echo "  6 - Required commands (terraform, jq, go) not found"
            echo "  7 - Provider directory not found or its schema could not be generated"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use -h or --help for usage information"
            exit 1
            ;;
    esac
done

# Check dependencies
# These can erroneously pass if the command name exists, but don't refer to the real tool
if ! command -v jq >/dev/null 2>&1; then
    echo "Error: jq command not found. Please install jq for JSON processing." >&2
    exit 6
fi

if ! command -v terraform >/dev/null 2>&1; then
    echo "Error: terraform command not found. Please install Terraform." >&2
    exit 6
fi

if ! command -v go >/dev/null 2>&1; then
    echo "Error: go command not found. Please install Go." >&2
    exit 6
fi

# Check if EXCEPTIONS_FILE exists and make it absolute, if provided via -e
if [ "${EXCEPTIONS_FILE}" != "${PROVIDER_DIR}/examples/error_exceptions.json" ]; then
    if [ ! -f "${EXCEPTIONS_FILE}" ]; then
        echo "Error: Exceptions file does not exist: ${EXCEPTIONS_FILE}" >&2
        exit 7
    fi
    EXCEPTIONS_FILE="$(cd "$(dirname "${EXCEPTIONS_FILE}")" && pwd)/$(basename "${EXCEPTIONS_FILE}")"
elif [ ! -f "${EXCEPTIONS_FILE}" ]; then
    echo "Warning: error_exceptions.json not found at ${EXCEPTIONS_FILE}" >&2
    echo "Proceeding without exceptions..." >&2
fi

# Generate provider schema to temporary file
echo "Generating provider schema..."
TEMP_DIR=$(mktemp -d)
PROVIDER_SCHEMA="${TEMP_DIR}/provider-schema.json"

# Exit cleanup trap
trap 'rm -rf "${TEMP_DIR}"' EXIT INT TERM

# Build provider binary
GOOS="${GOOS:-$(go env GOOS)}"
GOARCH="${GOARCH:-$(go env GOARCH)}"
if [ -z "${GOOS}" ] || [ -z "${GOARCH}" ]; then
    echo "Error: could not determine GOOS/GOARCH from go env." >&2
    exit 7
fi
OS_ARCH="${GOOS}_${GOARCH}"
PLUGIN_DIR="${TEMP_DIR}/plugins/registry.terraform.io/hashicorp/tfe/0.0.1/${OS_ARCH}" # tfe version is somewhat arbitrary for our particular usage of terraform init; this is the same as in tfplugindocs
mkdir -p "${PLUGIN_DIR}"
PROVIDER_BINARY="${PLUGIN_DIR}/terraform-provider-tfe"
if ! (cd "${PROVIDER_DIR}" && go build -o "${PROVIDER_BINARY}" 2>&1) >/dev/null; then
    echo "Error: failed to build provider binary." >&2
    exit 7
fi

# Create minimal provider configuration
cat > "${TEMP_DIR}/provider.tf" <<EOF
provider "tfe" {
}
EOF

# Initialize and extract schema
if ! (cd "${TEMP_DIR}" && terraform init -get=false -plugin-dir=./plugins >/dev/null 2>&1); then
    echo "Error: terraform init failed for provider schema generation." >&2
    exit 7
fi
if ! (cd "${TEMP_DIR}" && terraform providers schema -json > "${PROVIDER_SCHEMA}" 2>/dev/null); then
    echo "Error: terraform providers schema failed." >&2
    exit 7
fi

# Verify the schema file is valid JSON and contains the expected provider key
if ! jq -e '.provider_schemas["registry.terraform.io/hashicorp/tfe"]' "${PROVIDER_SCHEMA}" >/dev/null 2>&1; then
    echo "Error: provider schema is missing or invalid. The provider may not have been found." >&2
    exit 7
fi

# Load no_description_required list from exceptions file
NO_DESCRIPTION_REQUIRED=()
if [ -f "${EXCEPTIONS_FILE}" ]; then
    # Extract the no_description_required array
    while IFS= read -r entry; do
        NO_DESCRIPTION_REQUIRED+=("${entry}")
    done < <(jq -r '.no_description_required[]? // empty' "${EXCEPTIONS_FILE}" 2>/dev/null)
fi

# Check if an attribute or block path is covered by any entry in no_description_required.
# An entry can match at three levels of granularity:
#   "schema_type/resource_name/attr.path"  - exact attribute or block path
#   "schema_type/resource_name/"           - all root-level attributes and blocks only
#                                            (trailing slash; does not match nested paths)
#   "schema_type/resource_name"            - entire component (all paths, any depth)
# 0 on true, 1 on false
is_description_not_required() {
    local resource_key="$1"   # schema_type/resource_name
    local attr_path="$2"      # dot-separated path within the block

    for entry in "${NO_DESCRIPTION_REQUIRED[@]}"; do
        # Exact match: schema_type/resource_name/attr.path
        if [ "${entry}" = "${resource_key}/${attr_path}" ]; then
            return 0
        fi
        # Root-level only: schema_type/resource_name/ (trailing slash)
        # Matches only when attr_path contains no dots (i.e. not a nested path)
        if [ "${entry}" = "${resource_key}/" ] && [[ "${attr_path}" != *.* ]]; then
            return 0
        fi
        # Entire component: schema_type/resource_name (no slash suffix)
        if [ "${entry}" = "${resource_key}" ]; then
            return 0
        fi
    done
    return 1
}

echo "Validating schema descriptions for provider components..."
echo ""

# Recurse through and find issues
# A description is only considered present if it is non-empty.
# Three kinds are emitted: MISSING_COMPONENT, MISSING_BLOCK, MISSING_ATTR
RAW_MISSING=$(jq -r '
  def walk_block(prefix):
    . as $b
    | (
        # Check block-level description (only for nested blocks, not the root block)
        if prefix != "" and (($b.description // "") == "")
        then { kind: "MISSING_BLOCK", path: prefix }
        else empty
        end
      ),
      (
        # Check each attribute for a non-empty description
        if $b | has("attributes") then
          $b.attributes | to_entries[] |
          select((.value.description // "") == "") |
          { kind: "MISSING_ATTR", path: (if prefix == "" then .key else "\(prefix).\(.key)" end) }
        else empty
        end
      ),
      (
        # Recurse into nested block_types
        if $b | has("block_types") then
          $b.block_types | to_entries[] |
          .key as $bt_name |
          .value.block | walk_block(if prefix == "" then $bt_name else "\(prefix).\($bt_name)" end)
        else empty
        end
      )
    ;

  .provider_schemas["registry.terraform.io/hashicorp/tfe"]
  | to_entries[]
  | select(.key != "resource_identity_schemas")
  | .key as $schema_type
  | .value
  | (
      # provider is a single block directly at .block; all others are maps of resources
      if $schema_type == "provider" then
        [{ key: "provider", value: . }]
      else
        to_entries
      end
    )
  | .[]
  | .key as $resource_name
  | .value.block as $root_block
  | (
      # Check component-level description on the root block
      if ($root_block.description // "") == "" then
        "\("MISSING_COMPONENT")\t\($schema_type)/\($resource_name)\t."
      else empty
      end
    ),
    (
      $root_block | walk_block("") |
      "\(.kind)\t\($schema_type)/\($resource_name)\t\(.path)"
    )
' "${PROVIDER_SCHEMA}")

# Apply exceptions to the raw missing list
MISSING_DESCRIPTIONS=()
UNEXPECTED_DESCRIPTIONS=()
while IFS=$'\t' read -r kind resource_key attr_path; do
    [ -z "${kind}" ] && continue
    if ! is_description_not_required "${resource_key}" "${attr_path}"; then
        MISSING_DESCRIPTIONS+=("${kind}"$'\t'"${resource_key}"$'\t'"${attr_path}")
    fi
done <<< "${RAW_MISSING}"

# Detect unused exceptions
for entry in "${NO_DESCRIPTION_REQUIRED[@]}"; do
    stale=true
    while IFS=$'\t' read -r _kind rk ap; do
        [ -z "${_kind}" ] && continue
        # Exact
        [ "${entry}" = "${rk}/${ap}" ] && stale=false && break
        # Root-level (trailing slash)
        [ "${entry}" = "${rk}/" ] && [[ "${ap}" != *.* ]] && stale=false && break
        # Whole component
        [ "${entry}" = "${rk}" ] && stale=false && break
    done <<< "${RAW_MISSING}"

    if [ "${stale}" = true ]; then
        UNEXPECTED_DESCRIPTIONS+=("${entry}: listed in no_description_required but description is now present")
    fi
done

# Report unused exceptions
# Separate so it's always present
if [ ${#UNEXPECTED_DESCRIPTIONS[@]} -gt 0 ]; then
    echo ""
    echo "The following entries in no_description_required now have descriptions:"
    echo ""
    for unexpected in "${UNEXPECTED_DESCRIPTIONS[@]}"; do
        echo "  - ${unexpected}"
    done
    echo ""
    echo "Consider removing these entries from no_description_required in error_exceptions.json"
    echo ""
fi

# Collapse into the standard exit codes and write JSON output should errors exist
if [ ${#MISSING_DESCRIPTIONS[@]} -gt 0 ]; then
    # Build a JSON object: { "schema_type/resource_name": { "attr.path": "MISSING_KIND", ... } }
    json_output="{}"
    for item in "${MISSING_DESCRIPTIONS[@]}"; do
        IFS=$'\t' read -r kind resource_key attr_path <<< "${item}"
        json_output=$(echo "${json_output}" | jq \
            --arg rk "${resource_key}" \
            --arg path "${attr_path}" \
            --arg kind "${kind}" \
            '.[$rk][$path] = $kind')
    done
    jq -S '.' <<< "${json_output}" > "${JSON_OUTPUT_NAME}"
    echo "Validation errors found. See ${JSON_OUTPUT_NAME} for details."
    exit 5
elif [ ${#UNEXPECTED_DESCRIPTIONS[@]} -gt 0 ]; then
    exit 3
fi

echo "All validations passed successfully."
exit 0
