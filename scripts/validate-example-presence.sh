#!/usr/bin/env bash
# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0
#
# Usage:
#   ./scripts/validate-example-presence.sh
#
# Validates two categories of example presence in a single provider schema pass:
#
#   1. General examples: every resource, data source, action, and ephemeral
#      resource has at least one appropriately-prefixed *.tf example file.
#
#   2. Identity import examples: every resource with an identity schema has an
#      import-by-identity.tf file in its examples directory.
#
# The provider schema is generated once by building the provider binary and
# running `terraform providers schema -json`. Set SCHEMA_FILE to an existing
# JSON file to skip generation (used by tests).
#
# Exit codes:
#   0 - Success: All components have required examples
#   3 - Validation warning: Excepted components have unexpected examples
#   5 - Validation failed: One or more components are missing required examples
#   6 - Required commands (terraform, jq, go) not found
#   7 - Input files/directories not found or provider schema could not be generated
#   8 - Exceptions file exists but contains invalid JSON; or internal JSON output error
#   9 - Failure to build provider


# Crash on error
set -e

# Variables with defaults
SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]}")"
PROVIDER_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
EXAMPLES_DIR="${EXAMPLES_DIR:-${PROVIDER_DIR}/examples}"
EXCEPTIONS_FILE="${EXCEPTIONS_FILE:-${PROVIDER_DIR}/examples/error_exceptions.json}"
# SCHEMA_FILE may be set externally (e.g. by tests) to skip generation
SCHEMA_FILE="${SCHEMA_FILE:-}"

# Check dependencies
# These can erroneously pass if the command name exists, but don't refer to the real tool
if ! command -v jq >/dev/null 2>&1; then
    echo "Error: jq command not found. Please install jq for JSON processing." >&2
    exit 6
fi

# Only require terraform and go when we need to generate the schema
if [ -z "${SCHEMA_FILE}" ]; then
    if ! command -v terraform >/dev/null 2>&1; then
        echo "Error: terraform command not found. Please install Terraform." >&2
        exit 6
    fi
    if ! command -v go >/dev/null 2>&1; then
        echo "Error: go command not found. Please install Go." >&2
        exit 6
    fi
fi

# Exit if input folders are missing
if [ ! -d "${EXAMPLES_DIR}" ]; then
    echo "Error: examples directory not found at ${EXAMPLES_DIR}" >&2
    exit 7
fi

if [ ! -f "${EXCEPTIONS_FILE}" ]; then
    echo "Warning: exceptions file not found at ${EXCEPTIONS_FILE}" >&2
    echo "Proceeding without exceptions..." >&2
fi

# ---------------------------------------------------------------------------
# Schema generation
# ---------------------------------------------------------------------------

if [ -z "${SCHEMA_FILE}" ]; then
    echo "Generating provider schema..."
    TEMP_DIR=$(mktemp -d)
    trap 'rm -rf "${TEMP_DIR}"' EXIT INT TERM
    SCHEMA_FILE="${TEMP_DIR}/provider-schema.json"

    # Build provider binary
    GOOS="${GOOS:-$(go env GOOS)}"
    GOARCH="${GOARCH:-$(go env GOARCH)}"
    if [ -z "${GOOS}" ] || [ -z "${GOARCH}" ]; then
        echo "Error: could not determine GOOS/GOARCH from go env." >&2
        exit 9
    fi
    OS_ARCH="${GOOS}_${GOARCH}"
    PLUGIN_DIR="${TEMP_DIR}/plugins/registry.terraform.io/hashicorp/tfe/0.0.1/${OS_ARCH}" # tfe version is somewhat arbitrary for our particular usage of terraform init; this is the same as in tfplugindocs
    mkdir -p "${PLUGIN_DIR}"
    PROVIDER_BINARY="${PLUGIN_DIR}/terraform-provider-tfe"
    if ! (cd "${PROVIDER_DIR}" && go build -o "${PROVIDER_BINARY}" > /dev/null); then
        echo "Error: failed to build provider binary." >&2
        exit 9
    fi

    # Create minimal provider configuration
    cat > "${TEMP_DIR}/provider.tf" <<EOF
provider "tfe" {
}
EOF

    # Initialize and extract schema
    if ! (cd "${TEMP_DIR}" && terraform init -get=false -plugin-dir=./plugins > /dev/null); then
        echo "Error: terraform init failed for provider schema generation." >&2
        exit 7
    fi
    if ! (cd "${TEMP_DIR}" && terraform providers schema -json > "${SCHEMA_FILE}"); then
        echo "Error: terraform providers schema failed." >&2
        exit 7
    fi
fi

# Verify the schema file is valid JSON and contains the expected provider key
if ! jq -e '.provider_schemas["registry.terraform.io/hashicorp/tfe"]' "${SCHEMA_FILE}" >/dev/null 2>&1; then
    echo "Error: provider schema is missing or invalid. The provider may not have been found." >&2
    exit 7
fi

# ---------------------------------------------------------------------------
# Shared exception helpers
# ---------------------------------------------------------------------------

# Load both exception lists from exceptions file in a single guarded pass
NO_EXAMPLE_REQUIRED=()
NO_IDENTITY_EXAMPLE_REQUIRED=()
if [ -f "${EXCEPTIONS_FILE}" ]; then
    if ! jq -e '.' "${EXCEPTIONS_FILE}" >/dev/null 2>&1; then
        echo "Error: exceptions file is not valid JSON: ${EXCEPTIONS_FILE}" >&2
        exit 8
    fi
    while IFS= read -r component; do
        NO_EXAMPLE_REQUIRED+=("${component}")
    done < <(jq -r '.no_example_required[]? // empty' "${EXCEPTIONS_FILE}")
    while IFS= read -r component; do
        NO_IDENTITY_EXAMPLE_REQUIRED+=("${component}")
    done < <(jq -r '.no_identity_example_required[]? // empty' "${EXCEPTIONS_FILE}")
fi

# is_example_not_required <component_path> — 0 if excepted, 1 otherwise
is_example_not_required() {
    local component_path="$1"
    for excluded in "${NO_EXAMPLE_REQUIRED[@]}"; do
        [ "${excluded}" = "${component_path}" ] && return 0
    done
    return 1
}

# is_identity_example_not_required <component_path> — 0 if excepted, 1 otherwise
is_identity_example_not_required() {
    local component_path="$1"
    for excluded in "${NO_IDENTITY_EXAMPLE_REQUIRED[@]}"; do
        [ "${excluded}" = "${component_path}" ] && return 0
    done
    return 1
}

# ---------------------------------------------------------------------------
# Check 1: general example presence
# ---------------------------------------------------------------------------

MISSING_EXAMPLES=()
UNEXPECTED_EXAMPLES=()
TOTAL_COMPONENTS=0

# check_examples <component_type> <component_name>
check_examples() {
    local component_type="$1"  # e.g., "resources", "data-sources", "actions", "ephemeral-resources"
    local component_name="$2"  # e.g., "tfe_workspace"
    local component_path="${component_type}/${component_name}"

    TOTAL_COMPONENTS=$((TOTAL_COMPONENTS + 1))

    local example_dir="${EXAMPLES_DIR}/${component_path}"
    local has_examples=false

    # Determine required filename prefix based on component type
    local required_prefix=""
    case "${component_type}" in
        "resources")          required_prefix="resource"          ;;
        "data-sources")       required_prefix="data-source"       ;;
        "actions")            required_prefix="action"            ;;
        "ephemeral-resources") required_prefix="ephemeral-resource" ;;
    esac

    # Check if examples exist with the correct prefix (excludes import files and other non-examples)
    if [ -d "${example_dir}" ] && [ -n "${required_prefix}" ] && \
       find "${example_dir}" -maxdepth 1 -name "${required_prefix}*.tf" -type f | grep -q .; then
        has_examples=true
    fi

    if is_example_not_required "${component_path}"; then
        if [ "${has_examples}" = true ]; then
            UNEXPECTED_EXAMPLES+=("${component_path}: marked as no_example_required but examples exist")
        fi
        return 0
    fi

    if [ "${has_examples}" = false ]; then
        if [ ! -d "${example_dir}" ]; then
            MISSING_EXAMPLES+=("${component_path}: directory does not exist")
        else
            MISSING_EXAMPLES+=("${component_path}: directory exists but contains no example .tf files with the required prefix '${required_prefix}'")
        fi
    fi
}

echo "Validating example presence for provider components..."
echo ""

echo "Checking resources..."
if ! RESOURCES=$(jq -r '.provider_schemas["registry.terraform.io/hashicorp/tfe"].resource_schemas | keys? | .[]?' "${SCHEMA_FILE}" 2>/dev/null); then
    echo "Error: failed to read resource_schemas from provider schema." >&2
    exit 7
fi
if [ -n "${RESOURCES}" ]; then
    while IFS= read -r resource; do
        check_examples "resources" "${resource}" || true
    done <<< "${RESOURCES}"
fi

echo "Checking data sources..."
if ! DATA_SOURCES=$(jq -r '.provider_schemas["registry.terraform.io/hashicorp/tfe"].data_source_schemas | keys? | .[]?' "${SCHEMA_FILE}" 2>/dev/null); then
    echo "Error: failed to read data_source_schemas from provider schema." >&2
    exit 7
fi
if [ -n "${DATA_SOURCES}" ]; then
    while IFS= read -r data_source; do
        check_examples "data-sources" "${data_source}" || true
    done <<< "${DATA_SOURCES}"
fi

echo "Checking actions..."
if ! ACTIONS=$(jq -r '.provider_schemas["registry.terraform.io/hashicorp/tfe"].action_schemas | keys? | .[]?' "${SCHEMA_FILE}" 2>/dev/null); then
    echo "Error: failed to read action_schemas from provider schema." >&2
    exit 7
fi
if [ -n "${ACTIONS}" ]; then
    while IFS= read -r action; do
        check_examples "actions" "${action}" || true
    done <<< "${ACTIONS}"
fi

echo "Checking ephemeral resources..."
if ! EPHEMERAL_RESOURCES=$(jq -r '.provider_schemas["registry.terraform.io/hashicorp/tfe"].ephemeral_resource_schemas | keys? | .[]?' "${SCHEMA_FILE}" 2>/dev/null); then
    echo "Error: failed to read ephemeral_resource_schemas from provider schema." >&2
    exit 7
fi
if [ -n "${EPHEMERAL_RESOURCES}" ]; then
    while IFS= read -r ephemeral_resource; do
        check_examples "ephemeral-resources" "${ephemeral_resource}" || true
    done <<< "${EPHEMERAL_RESOURCES}"
fi

# ---------------------------------------------------------------------------
# Check 2: identity import example presence
# ---------------------------------------------------------------------------

MISSING_IDENTITY=()
UNEXPECTED_IDENTITY=()
TOTAL_IDENTITY=0

# check_identity_example <component_name>
check_identity_example() {
    local component_name="$1"  # e.g., "tfe_workspace"
    local component_path="resources/${component_name}"

    TOTAL_IDENTITY=$((TOTAL_IDENTITY + 1))

    local example_dir="${EXAMPLES_DIR}/${component_path}"
    local has_example=false
    if [ -f "${example_dir}/import-by-identity.tf" ]; then
        has_example=true
    fi

    if is_identity_example_not_required "${component_path}"; then
        if [ "${has_example}" = true ]; then
            UNEXPECTED_IDENTITY+=("${component_path}: marked as no_identity_example_required but import-by-identity.tf exists")
        fi
        return 0
    fi

    if [ "${has_example}" = false ]; then
        if [ ! -d "${example_dir}" ]; then
            MISSING_IDENTITY+=("${component_path}: directory does not exist")
        else
            MISSING_IDENTITY+=("${component_path}: directory exists but contains no import-by-identity.tf file")
        fi
    fi
}

echo "Checking identity import examples..."
if ! IDENTITY_RESOURCES=$(jq -r '.provider_schemas["registry.terraform.io/hashicorp/tfe"].resource_identity_schemas | keys? | .[]?' "${SCHEMA_FILE}" 2>/dev/null); then
    echo "Error: failed to read resource_identity_schemas from provider schema." >&2
    exit 7
fi
if [ -n "${IDENTITY_RESOURCES}" ]; then
    while IFS= read -r resource; do
        check_identity_example "${resource}" || true
    done <<< "${IDENTITY_RESOURCES}"
fi

# ---------------------------------------------------------------------------
# Report
# ---------------------------------------------------------------------------

echo ""

# Unexpected general examples (warning)
if [ ${#UNEXPECTED_EXAMPLES[@]} -gt 0 ]; then
    echo "Components marked as no_example_required but have examples:"
    echo ""
    for unexpected in "${UNEXPECTED_EXAMPLES[@]}"; do
        echo "  - ${unexpected}"
    done
    echo ""
    echo "Consider removing these components from no_example_required in ${EXCEPTIONS_FILE}"
    echo ""
fi

# Unexpected identity examples (warning)
if [ ${#UNEXPECTED_IDENTITY[@]} -gt 0 ]; then
    echo "Resources marked as no_identity_example_required but have import-by-identity.tf:"
    echo ""
    for unexpected in "${UNEXPECTED_IDENTITY[@]}"; do
        echo "  - ${unexpected}"
    done
    echo ""
    echo "Consider removing these components from no_identity_example_required in ${EXCEPTIONS_FILE}"
    echo ""
fi

# Missing general examples (error)
if [ ${#MISSING_EXAMPLES[@]} -gt 0 ]; then
    echo "Components missing examples:"
    echo ""
    for missing in "${MISSING_EXAMPLES[@]}"; do
        echo "  - ${missing}"
    done
    echo ""
    echo "Checked ${TOTAL_COMPONENTS} components total."
    echo ""
fi

# Missing identity examples (error)
if [ ${#MISSING_IDENTITY[@]} -gt 0 ]; then
    echo "Resources missing import-by-identity.tf:"
    echo ""
    for missing in "${MISSING_IDENTITY[@]}"; do
        echo "  - ${missing}"
    done
    echo ""
    echo "Checked ${TOTAL_IDENTITY} identity resources total."
    echo ""
fi

# Exit codes: error beats warning
if [ ${#MISSING_EXAMPLES[@]} -gt 0 ] || [ ${#MISSING_IDENTITY[@]} -gt 0 ]; then
    exit 5
fi

if [ ${#UNEXPECTED_EXAMPLES[@]} -gt 0 ] || [ ${#UNEXPECTED_IDENTITY[@]} -gt 0 ]; then
    exit 3
fi

echo "All ${TOTAL_COMPONENTS} components have at least one example file, or are excepted"
echo "All ${TOTAL_IDENTITY} identity resources have an import-by-identity.tf, or are excepted"
exit 0
