#!/usr/bin/env bash
# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0
#
# Usage:
#   ./scripts/validate-example-presence.sh
#
# Validates three categories of example presence in a single provider schema pass:
#
#   1. General examples: every resource, data source, action, and ephemeral
#      resource has at least one appropriately-prefixed *.tf example file,
#      and at least one such file contains a matching block for that same
#      component.
#
#   2. Identity import examples: every resource with an identity schema has an
#      import-by-identity.tf file in its examples directory, and that file
#      contains at least one import block for that same resource.
#
#   3. Action invoke examples: every action has an invoke.sh in its examples
#      directory containing at least one valid
#      `terraform apply -invoke=action.<name>.<label>` command.
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
#   7 - Provider schema could not be generated
#   8 - Exceptions file exists but contains invalid JSON; or internal JSON output error
#   9 - Failure to build provider


# Crash on error
set -e

# Variables with defaults
SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]}")"
PROVIDER_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
EXAMPLES_DIR="${EXAMPLES_DIR:-${PROVIDER_DIR}/examples}"
EXCEPTIONS_FILE="${EXCEPTIONS_FILE:-${PROVIDER_DIR}/examples/error_exceptions.json}"
# SCHEMA_FILE may be set externally (e.g. by tests) to skip schema generation
SCHEMA_FILE="${SCHEMA_FILE:-}"

# Dependency checks
# These can erroneously pass if the command name exists but don't refer to the real tool
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

# Missing examples directory is a validation failure, same as no examples present
if [ ! -d "${EXAMPLES_DIR}" ]; then
    echo "Error: examples directory not found at ${EXAMPLES_DIR}" >&2
    exit 5
fi

if [ ! -f "${EXCEPTIONS_FILE}" ]; then
    echo "Warning: exceptions file not found at ${EXCEPTIONS_FILE}" >&2
    echo "Proceeding without exceptions..." >&2
fi

# ---------------------------------------------------------------------------
# Schema generation
# ---------------------------------------------------------------------------

if [ -z "${SCHEMA_FILE}" ]; then
    echo "Building provider..."
    TEMP_DIR=$(mktemp -d)
    trap 'rm -rf "${TEMP_DIR}"' EXIT INT TERM
    SCHEMA_FILE="${TEMP_DIR}/provider-schema.json"

    GOOS="${GOOS:-$(go env GOOS)}"
    GOARCH="${GOARCH:-$(go env GOARCH)}"
    if [ -z "${GOOS}" ] || [ -z "${GOARCH}" ]; then
        echo "Error: could not determine GOOS/GOARCH from go env." >&2
        exit 9
    fi
    OS_ARCH="${GOOS}_${GOARCH}"
    # tfe version is somewhat arbitrary for our particular usage of terraform init;
    # this is the same version convention used by tfplugindocs internally.
    PLUGIN_DIR="${TEMP_DIR}/plugins/registry.terraform.io/hashicorp/tfe/0.0.1/${OS_ARCH}"
    mkdir -p "${PLUGIN_DIR}"
    PROVIDER_BINARY="${PLUGIN_DIR}/terraform-provider-tfe"
    if ! (cd "${PROVIDER_DIR}" && go build -o "${PROVIDER_BINARY}" > /dev/null); then
        echo "Error: failed to build provider binary." >&2
        exit 9
    fi

    cat > "${TEMP_DIR}/provider.tf" <<EOF
provider "tfe" {
}
EOF

    if ! (cd "${TEMP_DIR}" && terraform init -get=false -plugin-dir=./plugins > /dev/null); then
        echo "Error: terraform init failed for provider schema generation." >&2
        exit 7
    fi
    if ! (cd "${TEMP_DIR}" && terraform providers schema -json > "${SCHEMA_FILE}"); then
        echo "Error: terraform providers schema failed." >&2
        exit 7
    fi
    echo ""
fi

# Verify the schema file is valid JSON and contains the expected provider key
if ! jq -e '.provider_schemas["registry.terraform.io/hashicorp/tfe"]' "${SCHEMA_FILE}" >/dev/null 2>&1; then
    echo "Error: provider schema is missing or invalid. The provider may not have been found." >&2
    exit 7
fi

# ---------------------------------------------------------------------------
# Shared exception helpers
# ---------------------------------------------------------------------------

# Load all exception lists from exceptions file in a single guarded pass
NO_EXAMPLE_REQUIRED=()
NO_IDENTITY_EXAMPLE_REQUIRED=()
NO_INVOKE_EXAMPLE_REQUIRED=()
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
    while IFS= read -r component; do
        NO_INVOKE_EXAMPLE_REQUIRED+=("${component}")
    done < <(jq -r '.no_invoke_example_required[]? // empty' "${EXCEPTIONS_FILE}")
fi

# is_example_not_required <component_path> — returns 0 if excepted, 1 otherwise
is_example_not_required() {
    local component_path="$1"
    for excluded in "${NO_EXAMPLE_REQUIRED[@]}"; do
        [ "${excluded}" = "${component_path}" ] && return 0
    done
    return 1
}

# is_identity_example_not_required <component_path> — returns 0 if excepted, 1 otherwise
is_identity_example_not_required() {
    local component_path="$1"
    for excluded in "${NO_IDENTITY_EXAMPLE_REQUIRED[@]}"; do
        [ "${excluded}" = "${component_path}" ] && return 0
    done
    return 1
}

# is_invoke_example_not_required <component_path> — returns 0 if excepted, 1 otherwise
is_invoke_example_not_required() {
    local component_path="$1"
    for excluded in "${NO_INVOKE_EXAMPLE_REQUIRED[@]}"; do
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
#   component_type: e.g. "resources", "data-sources", "actions", "ephemeral-resources"
#   component_name: e.g. "tfe_workspace"
check_examples() {
    local component_type="$1"
    local component_name="$2"
    local component_path="${component_type}/${component_name}"

    TOTAL_COMPONENTS=$((TOTAL_COMPONENTS + 1))

    local example_dir="${EXAMPLES_DIR}/${component_path}"
    local has_examples=false
    local has_matching_example=false

    # Determine required filename prefix and HCL block keyword based on component type
    local required_prefix=""
    local block_keyword=""
    case "${component_type}" in
        "resources")           required_prefix="resource";           block_keyword="resource"  ;;
        "data-sources")        required_prefix="data-source";        block_keyword="data"      ;;
        "actions")             required_prefix="action";             block_keyword="action"    ;;
        "ephemeral-resources") required_prefix="ephemeral-resource"; block_keyword="ephemeral" ;;
    esac

    # Check if examples exist with the correct prefix (excludes import files and other non-examples)
    if [ -d "${example_dir}" ] && [ -n "${required_prefix}" ] && \
       find "${example_dir}" -maxdepth 1 -name "${required_prefix}*.tf" -type f | grep -q .; then
        has_examples=true
        while IFS= read -r example_file; do
            if grep -qE "^\\s*${block_keyword}\\s+\"${component_name}\"\\s+\"[A-Za-z0-9_-]+\"" "${example_file}" 2>/dev/null; then
                has_matching_example=true
                break
            fi
        done < <(find "${example_dir}" -maxdepth 1 -name "${required_prefix}*.tf" -type f | sort)
    fi

    echo "Validating: ${component_path}"

    if is_example_not_required "${component_path}"; then
        if [ "${has_examples}" = true ]; then
            echo "  warning"
            echo "    \"marked as no_example_required but examples exist\""
            UNEXPECTED_EXAMPLES+=("${component_path}"$'\t'"marked as no_example_required but examples exist")
        else
            echo "  pass (excepted)"
        fi
        return 0
    fi

    if [ "${has_examples}" = false ]; then
        if [ ! -d "${example_dir}" ]; then
            echo "  fail"
            echo "    \"directory does not exist\""
            MISSING_EXAMPLES+=("${component_path}"$'\t'"directory does not exist")
        else
            echo "  fail"
            echo "    \"no example .tf files with required prefix '${required_prefix}'\""
            MISSING_EXAMPLES+=("${component_path}"$'\t'"no example .tf files with required prefix '${required_prefix}'")
        fi
    elif [ "${has_matching_example}" = false ]; then
        echo "  fail"
        echo "    \"no ${block_keyword} block targeting ${component_name} found in example files\""
        MISSING_EXAMPLES+=("${component_path}"$'\t'"no ${block_keyword} block targeting ${component_name} found in example files")
    else
        echo "  pass"
    fi
}

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

echo ""
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

echo ""
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

echo ""
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

check_identity_example() {
    local component_name="$1"
    local component_path="resources/${component_name}"

    TOTAL_IDENTITY=$((TOTAL_IDENTITY + 1))

    local example_dir="${EXAMPLES_DIR}/${component_path}"
    local example_file="${example_dir}/import-by-identity.tf"
    local has_example=false
    local has_matching_import=false
    if [ -f "${example_file}" ]; then
        has_example=true
        if grep -qE "(^|[[:space:]])to[[:space:]]*=[[:space:]]*${component_name}\.[[:alnum:]_-]+([[:space:]]|$)" "${example_file}" 2>/dev/null; then
            has_matching_import=true
        fi
    fi

    echo "Validating: ${component_path} (identity import)"

    if is_identity_example_not_required "${component_path}"; then
        if [ "${has_example}" = true ]; then
            echo "  warning"
            echo "    \"marked as no_identity_example_required but import-by-identity.tf exists\""
            UNEXPECTED_IDENTITY+=("${component_path}"$'\t'"marked as no_identity_example_required but import-by-identity.tf exists")
        else
            echo "  pass (excepted)"
        fi
        return 0
    fi

    if [ "${has_example}" = false ]; then
        if [ ! -d "${example_dir}" ]; then
            echo "  fail"
            echo "    \"directory does not exist\""
            MISSING_IDENTITY+=("${component_path}"$'\t'"directory does not exist")
        else
            echo "  fail"
            echo "    \"no import-by-identity.tf file\""
            MISSING_IDENTITY+=("${component_path}"$'\t'"no import-by-identity.tf file")
        fi
    elif [ "${has_matching_import}" = false ]; then
        echo "  fail"
        echo "    \"import-by-identity.tf contains no import block targeting ${component_name}.<name>\""
        MISSING_IDENTITY+=("${component_path}"$'\t'"import-by-identity.tf contains no import block targeting ${component_name}.<name>")
    else
        echo "  pass"
    fi
}

echo ""
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
# Check 3: action invoke.sh presence
# ---------------------------------------------------------------------------

MISSING_INVOKE=()
UNEXPECTED_INVOKE=()
TOTAL_INVOKE=0

check_invoke_example() {
    local action_name="$1"
    local component_path="actions/${action_name}"

    TOTAL_INVOKE=$((TOTAL_INVOKE + 1))

    local invoke_sh="${EXAMPLES_DIR}/${component_path}/invoke.sh"
    local has_example=false
    if [ -f "${invoke_sh}" ] && \
       grep -qE "^terraform apply .*-invoke=action\.${action_name}\.[^ ]+" "${invoke_sh}" 2>/dev/null; then
        has_example=true
    fi

    echo "Validating: ${component_path} (invoke)"

    if is_invoke_example_not_required "${component_path}"; then
        if [ "${has_example}" = true ]; then
            echo "  warning"
            echo "    \"marked as no_invoke_example_required but invoke.sh exists\""
            UNEXPECTED_INVOKE+=("${component_path}"$'\t'"marked as no_invoke_example_required but invoke.sh exists")
        else
            echo "  pass (excepted)"
        fi
        return 0
    fi

    if [ "${has_example}" = false ]; then
        if [ ! -f "${invoke_sh}" ]; then
            echo "  fail"
            echo "    \"missing invoke.sh\""
            MISSING_INVOKE+=("${component_path}"$'\t'"missing invoke.sh")
        else
            echo "  fail"
            echo "    \"invoke.sh contains no valid 'terraform apply -invoke=action.${action_name}.<label>' command\""
            MISSING_INVOKE+=("${component_path}"$'\t'"invoke.sh contains no valid 'terraform apply -invoke=action.${action_name}.<label>' command")
        fi
    else
        echo "  pass"
    fi
}

echo ""
echo "Checking action invoke examples..."
if [ -n "${ACTIONS}" ]; then
    while IFS= read -r action_name; do
        check_invoke_example "${action_name}" || true
    done <<< "${ACTIONS}"
fi

# ---------------------------------------------------------------------------
# Check 4: orphan example directories (no matching schema component)
# ---------------------------------------------------------------------------

ORPHAN_DIRS=()

# check_orphan_dirs <component_type> <schema_list_var>
#   Scans every immediate subdirectory of examples/<component_type>/ and warns
#   if the directory name does not appear in the provider schema.
check_orphan_dirs() {
    local component_type="$1"
    local schema_names="$2"
    local type_dir="${EXAMPLES_DIR}/${component_type}"

    [ -d "${type_dir}" ] || return 0

    while IFS= read -r dir; do
        local name
        name=$(basename "${dir}")
        if ! echo "${schema_names}" | grep -qx "${name}"; then
            echo "Orphan: ${component_type}/${name}"
            echo "  warning"
            echo "    \"directory has no matching provider component\""
            ORPHAN_DIRS+=("${component_type}/${name}"$'\t'"directory has no matching provider component")
        fi
    done < <(find "${type_dir}" -mindepth 1 -maxdepth 1 -type d | sort)
}

echo ""
echo "Checking for orphan example directories..."
check_orphan_dirs "resources"            "${RESOURCES}"
check_orphan_dirs "data-sources"         "${DATA_SOURCES}"
check_orphan_dirs "actions"              "${ACTIONS}"
check_orphan_dirs "ephemeral-resources"  "${EPHEMERAL_RESOURCES}"

# ---------------------------------------------------------------------------
# Check 5: wrong-prefix example files
# ---------------------------------------------------------------------------

WRONG_PREFIX_FILES=()

# check_wrong_prefix_files <component_type> <correct_prefix>
#   Warns about any *.tf file in a component dir that does not carry the
#   expected prefix (e.g. "resource.tf" living under data-sources/).
#   import-by-identity.tf and invoke.sh are always allowed regardless of type.
check_wrong_prefix_files() {
    local component_type="$1"
    local correct_prefix="$2"
    local type_dir="${EXAMPLES_DIR}/${component_type}"

    [ -d "${type_dir}" ] || return 0

    while IFS= read -r tf_file; do
        local filename dir_name component_path
        filename=$(basename "${tf_file}")
        dir_name=$(basename "$(dirname "${tf_file}")")
        component_path="${component_type}/${dir_name}"

        # import-by-identity.tf is a well-known non-example file — always skip
        [ "${filename}" = "import-by-identity.tf" ] && continue

        if [[ "${filename}" != ${correct_prefix}* ]]; then
            echo "Wrong prefix: ${component_path}/${filename}"
            echo "  warning"
            echo "    \"file prefix '${filename%%_*}' does not match expected '${correct_prefix}' for ${component_type}\""
            WRONG_PREFIX_FILES+=("${component_path}/${filename}"$'\t'"file prefix does not match expected '${correct_prefix}' for ${component_type}")
        fi
    done < <(find "${type_dir}" -mindepth 2 -maxdepth 2 -name "*.tf" -type f | sort)
}

echo ""
echo "Checking for wrong-prefix example files..."
check_wrong_prefix_files "resources"            "resource"
check_wrong_prefix_files "data-sources"         "data-source"
check_wrong_prefix_files "actions"              "action"
check_wrong_prefix_files "ephemeral-resources"  "ephemeral-resource"

# ---------------------------------------------------------------------------
# Check 6: example files missing a block for their own component
# ---------------------------------------------------------------------------

MISSING_INTERNAL_COMPONENT=()

# check_internal_component <component_type> <block_keyword> <correct_prefix>
#   Fails any example .tf file (with the correct prefix) that does not contain
#   at least one block whose type label matches the containing folder name.
#   import-by-identity.tf is excluded (it uses import blocks, not component blocks).
check_internal_component() {
    local component_type="$1"
    local block_keyword="$2"
    local correct_prefix="$3"
    local type_dir="${EXAMPLES_DIR}/${component_type}"

    [ -d "${type_dir}" ] || return 0

    while IFS= read -r tf_file; do
        local filename dir_name component_name component_path
        filename=$(basename "${tf_file}")
        dir_name=$(basename "$(dirname "${tf_file}")")
        component_name="${dir_name}"
        component_path="${component_type}/${dir_name}"

        # import-by-identity.tf uses import blocks — not subject to this check
        [ "${filename}" = "import-by-identity.tf" ] && continue

        # Only check files with the correct prefix for this component type
        [[ "${filename}" != ${correct_prefix}* ]] && continue

        if ! grep -qE "^[[:space:]]*${block_keyword}[[:space:]]+\"${component_name}\"[[:space:]]+\"[A-Za-z0-9_-]+\"" "${tf_file}" 2>/dev/null; then
            echo "Missing component: ${component_path}/${filename}"
            echo "  fail"
            echo "    \"no ${block_keyword} block targeting ${component_name} found in file\""
            MISSING_INTERNAL_COMPONENT+=("${component_path}/${filename}"$'\t'"no ${block_keyword} block targeting ${component_name} found in file")
        fi
    done < <(find "${type_dir}" -mindepth 2 -maxdepth 2 -name "*.tf" -type f | sort)
}

echo ""
echo "Checking for example files missing their own component block..."
check_internal_component "resources"            "resource"          "resource"
check_internal_component "data-sources"         "data"              "data-source"
check_internal_component "actions"              "action"            "action"
check_internal_component "ephemeral-resources"  "ephemeral"         "ephemeral-resource"

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------

WARNINGS=()
FAILURES=()

for w in "${UNEXPECTED_EXAMPLES[@]}";  do WARNINGS+=("${w}"); done
for w in "${UNEXPECTED_IDENTITY[@]}";  do WARNINGS+=("${w}"); done
for w in "${UNEXPECTED_INVOKE[@]}";    do WARNINGS+=("${w}"); done
for w in "${ORPHAN_DIRS[@]}";          do WARNINGS+=("${w}"); done
for w in "${WRONG_PREFIX_FILES[@]}";   do WARNINGS+=("${w}"); done
for f in "${MISSING_EXAMPLES[@]}";            do FAILURES+=("${f}"); done
for f in "${MISSING_IDENTITY[@]}";            do FAILURES+=("${f}"); done
for f in "${MISSING_INVOKE[@]}";              do FAILURES+=("${f}"); done
for f in "${MISSING_INTERNAL_COMPONENT[@]}";  do FAILURES+=("${f}"); done

echo ""
echo "========================================"
echo " Summary"
echo "========================================"

if [ ${#WARNINGS[@]} -gt 0 ]; then
    echo ""
    echo "Warnings (${#WARNINGS[@]}):"
    for w in "${WARNINGS[@]}"; do
        IFS=$'\t' read -r path msg <<< "${w}"
        echo "  - ${path}"
        echo "      \"${msg}\""
    done
fi

if [ ${#FAILURES[@]} -gt 0 ]; then
    echo ""
    echo "Failures (${#FAILURES[@]}):"
    for f in "${FAILURES[@]}"; do
        IFS=$'\t' read -r path msg <<< "${f}"
        echo "  - ${path}"
        echo "      \"${msg}\""
    done
fi

echo ""
if [ ${#FAILURES[@]} -gt 0 ]; then
    echo "Result: FAILED"
    echo ""
    echo "Add example .tf files under examples/resources/<name>/ or"
    echo "examples/data-sources/<name>/. To skip a resource, add it to"
    echo "examples/error_exceptions.json under 'no_example_required'."
    echo "To run locally: ./scripts/validate-example-presence.sh"
    exit 5
elif [ ${#WARNINGS[@]} -gt 0 ]; then
    echo "Result: WARNINGS"
    exit 3
else
    echo "Result: PASSED"
fi

exit 0
