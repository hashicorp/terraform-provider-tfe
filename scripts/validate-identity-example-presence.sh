#!/usr/bin/env bash
# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0
#
# Usage:
#   ./scripts/validate-identity-example-presence.sh
#
# Validates that every resource with an identity schema in the provider schema
# has an associated import-by-identity.tf example file in examples/resources/.
#
# Exit codes:
#   0 - Success: All resources with identity schemas have import-by-identity examples
#   3 - Validation warning: Resources marked as no_identity_example_required have import-by-identity.tf
#   5 - Validation failed: One or more resources are missing import-by-identity.tf
#   6 - Required commands (jq) not found
#   7 - Schema file not found or invalid


# Crash on error
set -e

# Variables with defaults
SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]}")"
PROVIDER_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
EXAMPLES_DIR="${EXAMPLES_DIR:-${PROVIDER_DIR}/examples}"
SCHEMA_FILE="${SCHEMA_FILE:-${SCRIPT_DIR}/real-schema.json}"
EXCEPTIONS_FILE="${EXCEPTIONS_FILE:-${PROVIDER_DIR}/examples/error_exceptions.json}"

# Check dependencies
# These can erroneously pass if the command name exists, but don't refer to the real tool
if ! command -v jq >/dev/null 2>&1; then
    echo "Error: jq command not found. Please install jq for JSON processing." >&2
    exit 6
fi

# Verify schema file exists and is valid
if [ ! -f "${SCHEMA_FILE}" ]; then
    echo "Error: schema file not found at ${SCHEMA_FILE}" >&2
    exit 7
fi

if ! jq -e '.provider_schemas["registry.terraform.io/hashicorp/tfe"].resource_identity_schemas' "${SCHEMA_FILE}" >/dev/null 2>&1; then
    echo "Error: schema file is missing resource_identity_schemas or is invalid JSON." >&2
    exit 7
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

# Track missing examples and unexpected examples
MISSING_EXAMPLES=()
UNEXPECTED_EXAMPLES=()
TOTAL_COMPONENTS=0

# Load no_identity_example_required list from exceptions file
NO_IDENTITY_EXAMPLE_REQUIRED=()
if [ -f "${EXCEPTIONS_FILE}" ]; then
    # Extract the no_identity_example_required array
    while IFS= read -r component; do
        NO_IDENTITY_EXAMPLE_REQUIRED+=("${component}")
    done < <(jq -r '.no_identity_example_required[]? // empty' "${EXCEPTIONS_FILE}" 2>/dev/null)
fi

# Check if a component is in the no_identity_example_required list
# 0 on true, 1 on false
is_identity_example_not_required() {
    local component_path="$1"
    for excluded in "${NO_IDENTITY_EXAMPLE_REQUIRED[@]}"; do
        if [ "${excluded}" = "${component_path}" ]; then
            return 0
        fi
    done
    return 1
}

# Check if an import-by-identity.tf example exists for a resource; appends to
# MISSING_EXAMPLES or UNEXPECTED_EXAMPLES as appropriate
check_identity_example() {
    local component_name="$1"  # e.g., "tfe_workspace"
    local component_path="resources/${component_name}"

    TOTAL_COMPONENTS=$((TOTAL_COMPONENTS + 1))

    local example_dir="${EXAMPLES_DIR}/${component_path}"
    local has_example=false
    if [ -f "${example_dir}/import-by-identity.tf" ]; then
        has_example=true
    fi

    # Check if component is in no_identity_example_required list
    if is_identity_example_not_required "${component_path}"; then
        if [ "${has_example}" = true ]; then
            UNEXPECTED_EXAMPLES+=("${component_path}: marked as no_identity_example_required but import-by-identity.tf exists")
        fi
        return 0
    fi

    # Component requires an import-by-identity.tf but doesn't have one
    if [ "${has_example}" = false ]; then
        if [ ! -d "${example_dir}" ]; then
            MISSING_EXAMPLES+=("${component_path}: directory does not exist")
        else
            MISSING_EXAMPLES+=("${component_path}: directory exists but contains no import-by-identity.tf file")
        fi
    fi
}

echo "Validating import-by-identity example presence for provider resources..."
echo ""

# Extract and check resources with identity schemas
echo "Checking resources..."
IDENTITY_RESOURCES=$(jq -r '.provider_schemas["registry.terraform.io/hashicorp/tfe"].resource_identity_schemas | keys[]' "${SCHEMA_FILE}" 2>/dev/null || echo "")
if [ -n "${IDENTITY_RESOURCES}" ]; then
    while IFS= read -r resource; do
        check_identity_example "${resource}" || true
    done <<< "${IDENTITY_RESOURCES}"
fi

# Check for unexpected examples first (warning)
if [ ${#UNEXPECTED_EXAMPLES[@]} -gt 0 ]; then
    echo "Resources marked as no_identity_example_required but have import-by-identity.tf:"
    echo ""
    for unexpected in "${UNEXPECTED_EXAMPLES[@]}"; do
        echo "  - ${unexpected}"
    done
    echo ""
    echo "Consider either removing these components from no_identity_example_required in the error exceptions json"
    echo ""
fi

# Check for missing import-by-identity.tf files (error)
if [ ${#MISSING_EXAMPLES[@]} -gt 0 ]; then
    echo "Resources missing import-by-identity.tf:"
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

echo "All ${TOTAL_COMPONENTS} components have an import-by-identity.tf, or are excepted"
exit 0
