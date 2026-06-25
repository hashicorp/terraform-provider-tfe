#!/usr/bin/env bash
# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0
#
# Usage:
#   ./scripts/validate-example-presence.sh
#
# Exit codes:
#   0 - Success: All components have examples
#   3 - Validation warning: Components marked as no_example_required have examples
#   5 - Validation failed: One or more components are missing examples in examples/
#   6 - Required commands (terraform, jq, go) not found
#   7 - Provider directory not found or its schema could not be generated


# Crash on error
set -e

# Variables with defaults
SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]}")"
PROVIDER_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
EXAMPLES_DIR="${PROVIDER_DIR}/examples"
EXCEPTIONS_FILE="${PROVIDER_DIR}/examples/error_exceptions.json"

# Check dependencies
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

# Exit if input folders are missing
if [ ! -d "${EXAMPLES_DIR}" ]; then
    echo "Error: examples directory not found at ${EXAMPLES_DIR}" >&2
    exit 5
fi

if [ ! -f "${EXCEPTIONS_FILE}" ]; then
    echo "Warning: error_exceptions.json not found at ${EXCEPTIONS_FILE}" >&2
    echo "Proceeding without exceptions..." >&2
fi

# Generate provider schema to temporary file
echo "Generating provider schema..."
TEMP_DIR=$(mktemp -d)
PROVIDER_SCHEMA="${TEMP_DIR}/provider-schema.json"
trap "rm -rf ${TEMP_DIR}" EXIT INT TERM

# Build provider binary
OS_ARCH="$(go env GOOS)_$(go env GOARCH)"
PLUGIN_DIR="${TEMP_DIR}/plugins/registry.terraform.io/hashicorp/tfe/0.0.1/${OS_ARCH}"
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

# Track missing examples and unexpected examples
MISSING_EXAMPLES=()
UNEXPECTED_EXAMPLES=()
TOTAL_COMPONENTS=0

# Load no_example_required list from exceptions file
NO_EXAMPLE_REQUIRED=()
if [ -f "${EXCEPTIONS_FILE}" ]; then
    # Extract the no_example_required array
    while IFS= read -r component; do
        NO_EXAMPLE_REQUIRED+=("${component}")
    done < <(jq -r '.no_example_required[]? // empty' "${EXCEPTIONS_FILE}" 2>/dev/null)
fi

# Check if a component is in the no_example_required list
# 0 on true, 1 on false
is_example_not_required() {
    local component_path="$1"
    for excluded in "${NO_EXAMPLE_REQUIRED[@]}"; do
        if [ "${excluded}" = "${component_path}" ]; then
            return 0
        fi
    done
    return 1
}

# Check if examples exist for a component; appends to MISSING_EXAMPLES or UNEXPECTED_EXAMPLES as appropriate
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
        "resources")
            required_prefix="resource"
            ;;
        "data-sources")
            required_prefix="data-source"
            ;;
        "actions")
            required_prefix="action"
            ;;
        "ephemeral-resources")
            required_prefix="ephemeral-resource"
            ;;
    esac
    
    # Check if examples exist with the correct prefix (excludes import files and other non-examples)
    if [ -d "${example_dir}" ] && [ -n "${required_prefix}" ] && \
       find "${example_dir}" -maxdepth 1 -name "${required_prefix}*.tf" -type f | grep -q .; then
        has_examples=true
    fi
    
    # Check if component is in no_example_required list
    if is_example_not_required "${component_path}"; then
        if [ "${has_examples}" = true ]; then
            UNEXPECTED_EXAMPLES+=("${component_path}: marked as no_example_required but examples exist")
        fi
        return 0
    fi
    
    # Component requires examples but doesn't have them
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

# Extract and check resources
echo "Checking resources..."
RESOURCES=$(jq -r '.provider_schemas["registry.terraform.io/hashicorp/tfe"].resource_schemas | keys[]' "${PROVIDER_SCHEMA}" 2>/dev/null || echo "")
if [ -n "${RESOURCES}" ]; then
    while IFS= read -r resource; do
        check_examples "resources" "${resource}" || true
    done <<< "${RESOURCES}"
fi

# Extract and check data sources
echo "Checking data sources..."
DATA_SOURCES=$(jq -r '.provider_schemas["registry.terraform.io/hashicorp/tfe"].data_source_schemas | keys[]' "${PROVIDER_SCHEMA}" 2>/dev/null || echo "")
if [ -n "${DATA_SOURCES}" ]; then
    while IFS= read -r data_source; do
        check_examples "data-sources" "${data_source}" || true
    done <<< "${DATA_SOURCES}"
fi

# Extract and check actions
echo "Checking actions..."
ACTIONS=$(jq -r '.provider_schemas["registry.terraform.io/hashicorp/tfe"].action_schemas | keys[]' "${PROVIDER_SCHEMA}" 2>/dev/null || echo "")
if [ -n "${ACTIONS}" ]; then
    while IFS= read -r action; do
        check_examples "actions" "${action}" || true
    done <<< "${ACTIONS}"
fi

# Extract and check ephemeral resources
echo "Checking ephemeral resources..."
EPHEMERAL_RESOURCES=$(jq -r '.provider_schemas["registry.terraform.io/hashicorp/tfe"].ephemeral_resource_schemas | keys[]' "${PROVIDER_SCHEMA}" 2>/dev/null || echo "")
if [ -n "${EPHEMERAL_RESOURCES}" ]; then
    while IFS= read -r ephemeral_resource; do
        check_examples "ephemeral-resources" "${ephemeral_resource}" || true
    done <<< "${EPHEMERAL_RESOURCES}"
fi

# Check for unexpected examples first (warning)
if [ ${#UNEXPECTED_EXAMPLES[@]} -gt 0 ]; then
    echo "Components marked as no_example_required but have examples:"
    echo ""
    for unexpected in "${UNEXPECTED_EXAMPLES[@]}"; do
        echo "  - ${unexpected}"
    done
    echo ""
    echo "Consider either removing these components from no_example_required in the error exceptions json"
    echo ""
fi

# Check for missing examples (error)
if [ ${#MISSING_EXAMPLES[@]} -gt 0 ]; then
    echo "Components missing examples:"
    echo ""
    for missing in "${MISSING_EXAMPLES[@]}"; do
        echo "  - ${missing}"
    done
    echo ""
    echo "Checked ${TOTAL_COMPONENTS} components total."
    exit 5
fi

# Exit with warning code if there are unexpected examples
if [ ${#UNEXPECTED_EXAMPLES[@]} -gt 0 ]; then
    exit 3
fi

echo "All ${TOTAL_COMPONENTS} components have at least one example file, or are excepted"
exit 0

