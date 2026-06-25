#!/usr/bin/env bash
# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0
#
# Usage:
#   ./scripts/validate-example-presence.sh
#
# Exit codes:
#   0 - Success: All components have examples
#   1 - Error: Missing dependencies or required files
#   3 - Validation failed: One or more components are missing examples
#   4 - Validation warning: Components marked as no_example_required have examples


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
    exit 1
fi

if ! command -v terraform >/dev/null 2>&1; then
    echo "Error: terraform command not found. Please install Terraform." >&2
    exit 1
fi

if ! command -v go >/dev/null 2>&1; then
    echo "Error: go command not found. Please install Go." >&2
    exit 1
fi

# Generate provider schema to temporary file
echo "Generating provider schema..."
TEMP_DIR=$(mktemp -d)
TEMP_SCHEMA="${TEMP_DIR}/provider-schema.json"
trap "rm -rf ${TEMP_DIR}" EXIT INT TERM

# Build provider binary
OS_ARCH="$(go env GOOS)_$(go env GOARCH)"
PLUGIN_DIR="${TEMP_DIR}/plugins/registry.terraform.io/hashicorp/tfe/0.0.1/${OS_ARCH}"
mkdir -p "${PLUGIN_DIR}"
PROVIDER_BINARY="${PLUGIN_DIR}/terraform-provider-tfe"
(cd "${PROVIDER_DIR}" && go build -o "${PROVIDER_BINARY}") >/dev/null 2>&1

# Create minimal provider configuration
cat > "${TEMP_DIR}/provider.tf" <<EOF
provider "tfe" {
}
EOF

# Initialize and extract schema
(cd "${TEMP_DIR}" && terraform init -get=false -plugin-dir=./plugins >/dev/null 2>&1)
(cd "${TEMP_DIR}" && terraform providers schema -json > "${TEMP_SCHEMA}" 2>/dev/null)

PROVIDER_SCHEMA="${TEMP_SCHEMA}"

# Exit if input folders are missing
if [ ! -d "${EXAMPLES_DIR}" ]; then
    echo "Error: examples directory not found at ${EXAMPLES_DIR}" >&2
    exit 1
fi

if [ ! -f "${EXCEPTIONS_FILE}" ]; then
    echo "Warning: error_exceptions.json not found at ${EXCEPTIONS_FILE}" >&2
    echo "Proceeding without exceptions..." >&2
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

# Check if examples exist for a component
# 0 on true, 1 on false
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
            MISSING_EXAMPLES+=("${component_path}: directory exists but contains no .tf files")
        fi
        return 1
    fi
    
    return 0
}

echo "Validating example presence for provider components..."
echo ""

# Extract and check resources
echo "Checking resources..."
RESOURCES=$(jq -r '.provider_schemas["registry.terraform.io/hashicorp/tfe"].resource_schemas | keys[]' "${PROVIDER_SCHEMA}")
while IFS= read -r resource; do
    check_examples "resources" "${resource}" || true
done <<< "${RESOURCES}"

# Extract and check data sources
echo "Checking data sources..."
DATA_SOURCES=$(jq -r '.provider_schemas["registry.terraform.io/hashicorp/tfe"].data_source_schemas | keys[]' "${PROVIDER_SCHEMA}")
while IFS= read -r data_source; do
    check_examples "data-sources" "${data_source}" || true
done <<< "${DATA_SOURCES}"

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

echo ""
echo "=========================================="
echo "Total components checked: ${TOTAL_COMPONENTS}"
echo "Missing examples: ${#MISSING_EXAMPLES[@]}"
echo "Unexpected examples: ${#UNEXPECTED_EXAMPLES[@]}"
echo ""

# Check for unexpected examples first (warning)
if [ ${#UNEXPECTED_EXAMPLES[@]} -gt 0 ]; then
    echo "Components marked as no_example_required but have examples:"
    echo ""
    for unexpected in "${UNEXPECTED_EXAMPLES[@]}"; do
        echo "  - ${unexpected}"
    done
    echo ""
    echo "Consider either removing these components from no_example_requireed in the error exceptions json"
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
    exit 3
fi

# Exit with warning code if there are unexpected examples
if [ ${#UNEXPECTED_EXAMPLES[@]} -gt 0 ]; then
    exit 4
fi

echo "All components have at least one example file"
exit 0

# Made with Bob
