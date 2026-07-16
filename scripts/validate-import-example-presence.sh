#!/usr/bin/env bash
# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0
#
# Usage:
#   ./scripts/validate-import-example-presence.sh
#
# Validates that every resource which defines CLI import support has an
# associated import.sh in examples/resources/ containing at least one
# valid `terraform import` command of the form:
#   terraform import <resource_type>.<name> <id>
#
# The canonical resource name list is derived from two sources:
#  - internal/provider/provider.go  ResourcesMap keys  (SDK v2 resources)
#  - internal/provider/resource_*.go Metadata methods  (framework resources)
#
# CLI import support is detected by the presence of either:
#  - `Importer:`               (SDK v2 schema.ResourceImporter)
#  - `ResourceWithImportState` (plugin framework interface assertion)
#
# Exit codes:
#  0 - Success: All CLI-importable resources have import.sh examples
#  3 - Validation warning: Unexpected import.sh examples found
#  5 - Validation failed: One or more resources are missing a valid import.sh
#  7 - Provider source directory or provider.go not found
#  8 - Exceptions file exists but contains invalid JSON


# Crash on error
set -e

# Variables with defaults
SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]}")"
PROVIDER_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
EXAMPLES_DIR="${EXAMPLES_DIR:-${PROVIDER_DIR}/examples}"
SOURCE_DIR="${SOURCE_DIR:-${PROVIDER_DIR}/internal/provider}"
PROVIDER_GO="${PROVIDER_GO:-${SOURCE_DIR}/provider.go}"
EXCEPTIONS_FILE="${EXCEPTIONS_FILE:-${PROVIDER_DIR}/examples/error_exceptions.json}"

# Verify source files exist
if [ ! -d "${SOURCE_DIR}" ]; then
    echo "Error: provider source directory not found at ${SOURCE_DIR}" >&2
    exit 7
fi

if [ ! -f "${PROVIDER_GO}" ]; then
    echo "Error: provider.go not found at ${PROVIDER_GO}" >&2
    exit 7
fi

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

# Load no_import_example_required list from exceptions file
NO_IMPORT_EXAMPLE_REQUIRED=()
if [ -f "${EXCEPTIONS_FILE}" ]; then
    if ! jq -e '.' "${EXCEPTIONS_FILE}" >/dev/null 2>&1; then
        echo "Error: exceptions file is not valid JSON: ${EXCEPTIONS_FILE}" >&2
        exit 8
    fi
    # Extract the no_import_example_required array
    while IFS= read -r component; do
        NO_IMPORT_EXAMPLE_REQUIRED+=("${component}")
    done < <(jq -r '.no_import_example_required[]? // empty' "${EXCEPTIONS_FILE}")
fi

# Check if a component is in the no_import_example_required list
# 0 on true, 1 on false
is_import_example_not_required() {
    local component_path="$1"
    for excluded in "${NO_IMPORT_EXAMPLE_REQUIRED[@]}"; do
        if [ "${excluded}" = "${component_path}" ]; then
            return 0
        fi
    done
    return 1
}

# Build the full list of canonical resource names from provider.go and
# framework resource Metadata methods.
#
# Three patterns cover all resources:
#
#   1. SDK v2 ResourcesMap keys in provider.go:
#        "tfe_foo":   resourceTFEFoo(),
#
#   2. Framework resources using req.ProviderTypeName concatenation:
#        resp.TypeName = req.ProviderTypeName + "_foo"
#        -> tfe_foo
#
#   3. Framework resources using a literal type name assignment:
#        res.TypeName = "tfe_foo"  /  resp.TypeName = "tfe_foo"
#
collect_resource_names() {
    local sdk_names
    # Pattern 1: SDK v2 ResourcesMap — scoped to that block only.
    # The block runs from "ResourcesMap:" up to "ConfigureContextFunc:".
    sdk_names=$(awk '/ResourcesMap:/,/ConfigureContextFunc:/' "${PROVIDER_GO}" \
        | grep '"tfe_' \
        | awk 'match($0, /"tfe_[^"]+"/) { print substr($0, RSTART+1, RLENGTH-2) }')
    if [ -z "${sdk_names}" ]; then
        echo "Error: collect_resource_names found no SDK v2 resources in ${PROVIDER_GO}." \
            "The ResourcesMap block may have been renamed or restructured." >&2
        exit 7
    fi
    echo "${sdk_names}"

    # Patterns 2 & 3: framework resource files
    for src in "${SOURCE_DIR}"/resource_*.go; do
        [[ "${src}" == *_test.go ]] && continue
        # Pattern 2: resp.TypeName = req.ProviderTypeName + "_suffix"
        grep -h 'ProviderTypeName + "_' "${src}" 2>/dev/null \
            | awk '/ProviderTypeName \+ "_/ { split($0,a,"\""); for(i in a) if(a[i]~/^_/){ print "tfe" a[i]; break } }'
        # Pattern 3: literal TypeName = "tfe_..."
        grep -h 'TypeName = "tfe_' "${src}" 2>/dev/null \
            | awk 'match($0, /"tfe_[^"]+"/) { print substr($0, RSTART+1, RLENGTH-2) }'
    done
}

# Check if an import.sh example exists for a resource; appends to
# MISSING_EXAMPLES or UNEXPECTED_EXAMPLES as appropriate
check_import_example() {
    local resource_name="$1"  # e.g., "tfe_workspace"
    local component_path="resources/${resource_name}"

    # Derive the expected Go source filename for this resource
    local src="${SOURCE_DIR}/resource_${resource_name}.go"
    if [ ! -f "${src}" ]; then
        return 0
    fi

    # Skip if this resource does not define CLI import support.
    # Two indicators are checked:
    #   - SDK v2:    Importer: field on schema.Resource
    #   - Framework: func (...) ImportState( method definition
    if ! grep -qE 'Importer:|func \(.*\) ImportState\(' "${src}" 2>/dev/null; then
        return 0
    fi

    TOTAL_COMPONENTS=$((TOTAL_COMPONENTS + 1))

    local import_sh="${EXAMPLES_DIR}/${component_path}/import.sh"
    local has_example=false
    # A valid import line is an uncommented line of the form:
    #   terraform import <resource_type>.<name> <id_with_possible_slashes>
    if [ -f "${import_sh}" ] && grep -qE "^terraform import ${resource_name}\.[^ ]+ [^ ]+" "${import_sh}" 2>/dev/null; then
        has_example=true
    fi

    # Check if component is in no_import_example_required list
    if is_import_example_not_required "${resource_name}"; then
        if [ "${has_example}" = true ]; then
            UNEXPECTED_EXAMPLES+=("${component_path}: marked as no_import_example_required but import.sh exists")
        fi
        return 0
    fi

    # Component requires an example but doesn't have one
    if [ "${has_example}" = false ]; then
        if [ ! -f "${import_sh}" ]; then
            MISSING_EXAMPLES+=("${component_path}: missing examples/${component_path}/import.sh")
        else
            MISSING_EXAMPLES+=("${component_path}: import.sh exists but contains no valid 'terraform import ${resource_name}.<name> <id>' command")
        fi
    fi
}

echo "Validating import.sh example presence for CLI-importable resources..."
echo ""

# Collect all canonical resource names once, used both for checking and for
# detecting orphaned import.sh files.
# collect_resource_names exits 7 inside the subshell when the ResourcesMap is
# empty; we detect that by checking whether the output array is empty after the
# process substitution completes.
RESOURCE_NAMES=()
_collect_names_output=$(collect_resource_names 2>/dev/null)
_collect_names_exit=$?
if [ ${_collect_names_exit} -ne 0 ] || [ -z "${_collect_names_output}" ]; then
    echo "Error: collect_resource_names found no SDK v2 resources in ${PROVIDER_GO}." \
        "The ResourcesMap block may have been renamed or restructured." >&2
    exit 7
fi
while IFS= read -r name; do
    RESOURCE_NAMES+=("${name}")
done < <(echo "${_collect_names_output}" | sort -u)

# Extract and check resources
echo "Checking resources..."
for resource in "${RESOURCE_NAMES[@]}"; do
    check_import_example "${resource}" || true
done

# Warn about orphaned import.sh files — present in examples/ but the resource
# is either not known to the provider or not detected as CLI-importable
echo "Checking for orphaned import.sh files..."
while IFS= read -r import_sh; do
    resource=$(basename "$(dirname "${import_sh}")")
    # Check if this resource is in the canonical list
    found=false
    for known in "${RESOURCE_NAMES[@]}"; do
        if [ "${known}" = "${resource}" ]; then
            found=true
            break
        fi
    done
    if [ "${found}" = false ]; then
        UNEXPECTED_EXAMPLES+=("resources/${resource}: import.sh exists but resource is not known to the provider")
        continue
    fi
    # Check if the resource is actually CLI-importable
    src="${SOURCE_DIR}/resource_${resource}.go"
    if [ -f "${src}" ] && ! grep -qE 'Importer:|func \(.*\) ImportState\(' "${src}" 2>/dev/null; then
        UNEXPECTED_EXAMPLES+=("resources/${resource}: import.sh exists but resource does not define CLI import support")
    fi
done < <(find "${EXAMPLES_DIR}/resources" -maxdepth 2 -name "import.sh" -type f | sort)

# Check for unexpected examples first (warning)
if [ ${#UNEXPECTED_EXAMPLES[@]} -gt 0 ]; then
    echo "Unexpected import.sh examples found:"
    echo ""
    for unexpected in "${UNEXPECTED_EXAMPLES[@]}"; do
        echo "  - ${unexpected}"
    done
    echo ""
    echo "Consider removing these components from no_import_example_required in ${EXCEPTIONS_FILE}"
    echo ""
fi

# Check for missing examples (error)
if [ ${#MISSING_EXAMPLES[@]} -gt 0 ]; then
    echo "Resources missing import.sh examples:"
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

echo "All ${TOTAL_COMPONENTS} components have at least one import.sh example, or are excepted"
exit 0
