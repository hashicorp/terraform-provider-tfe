#!/usr/bin/env bash
# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0
#
# Usage:
#   ./scripts/validate-examples.sh
#
# Environment variables:
#   EXCEPTIONS_FILE    JSON file with error/warning exceptions to ignore
#                      (default: examples/error_exceptions.json — soft warning if absent)
#                      Set to an explicit path to hard-fail if missing.
#                      Set to empty string ("") to disable exceptions entirely.
#
# Exit codes:
#  0 - Complete success
#  3 - Warnings found in examples, no errors
#  4 - Warning that unused exceptions were found in error_exceptions.json
#  5 - Errors found in examples or a missing examples directory
#  6 - Required commands (terraform, jq, go) not found
#  7 - Exceptions file does not exist
#  8 - Internal data merge error
#  9 - Failure to build provider


# Crash on error
set -e

# Variables with defaults
SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]}")"
HAS_ERRORS=false
HAS_WARNINGS=false
UNUSED_EXCEPTIONS=()
WARNINGS=()
FAILURES=()

# Normalize paths
PROVIDER_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
TARGET_DIR="${PROVIDER_DIR}/examples"

# EXCEPTIONS_FILE: if explicitly set to a non-empty path, treat as hard requirement
# (exit 7 if missing). If unset, fall back to the default with a soft warning if
# absent. If explicitly set to empty string, disable exceptions entirely.
EXCEPTIONS_FILE_EXPLICIT=false
if [ -n "${EXCEPTIONS_FILE+x}" ]; then
    # Variable is set (possibly empty)
    if [ -n "${EXCEPTIONS_FILE}" ]; then
        EXCEPTIONS_FILE_EXPLICIT=true
        # Resolve to an absolute path now so that later file-existence checks and jq calls are CWD-independent.  
        # Only resolve when the parent directory exists; if it doesn't, the existence check below will emit the correct error and exit 7.
        if [ "${EXCEPTIONS_FILE:0:1}" != "/" ] && [ -d "$(dirname "${EXCEPTIONS_FILE}")" ]; then
            EXCEPTIONS_FILE="$(cd "$(dirname "${EXCEPTIONS_FILE}")" && pwd)/$(basename "${EXCEPTIONS_FILE}")"
        fi
    fi
    # else: explicitly set to "", leave EXCEPTIONS_FILE="" — no exceptions
else
    # Variable is unset — use default
    EXCEPTIONS_FILE="${PROVIDER_DIR}/examples/error_exceptions.json"
fi

# Constants
JQ_ERROR_PROCESSING='{"diagnostics":[{"severity":"error","summary":"Processing error","detail":"jq processing failed"}],"error_count":1,"warning_count":0}'

# Dependency checks
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

# Missing examples directory is a validation failure, same as no examples present.
if [ ! -d "${TARGET_DIR}" ]; then
    echo "Error: Examples directory does not exist: ${TARGET_DIR}" >&2
    exit 5
fi

if [ "${EXCEPTIONS_FILE_EXPLICIT}" = true ]; then
    if [ ! -f "${EXCEPTIONS_FILE}" ]; then
        echo "Error: Exceptions file does not exist: ${EXCEPTIONS_FILE}" >&2
        exit 7
    fi
fi
if [ -n "${EXCEPTIONS_FILE}" ] && [ -f "${EXCEPTIONS_FILE}" ]; then
    EXCEPTIONS_FILE="$(cd "$(dirname "${EXCEPTIONS_FILE}")" && pwd)/$(basename "${EXCEPTIONS_FILE}")"
    if ! jq -e '.' "${EXCEPTIONS_FILE}" >/dev/null 2>&1; then
        echo "Error: exceptions file is not valid JSON: ${EXCEPTIONS_FILE}" >&2
        exit 8
    fi
elif [ -n "${EXCEPTIONS_FILE}" ]; then
    # Default path was set but file doesn't exist — soft warning
    echo "Warning: error_exceptions.json not found at ${EXCEPTIONS_FILE}" >&2
    echo "Proceeding without exceptions..." >&2
    EXCEPTIONS_FILE=""
fi
# else: EXCEPTIONS_FILE="" (explicitly disabled) — proceed silently

# Create temp working directory and register cleanup
TEST_DIR=$(mktemp -d)
cleanup() {
    [ -n "${TEST_DIR}" ] && rm -rf "${TEST_DIR}"
}
trap cleanup EXIT INT TERM

# Build provider binary
echo "Building provider..."
PLUGIN_DIR="${TEST_DIR}/provider-bin"
mkdir -p "${PLUGIN_DIR}"
PROVIDER_BINARY="${PLUGIN_DIR}/terraform-provider-tfe"
if ! (cd "${PROVIDER_DIR}" && go build -o "${PROVIDER_BINARY}" > /dev/null); then
    echo "Error: failed to build provider binary." >&2
    exit 9
fi
echo ""

cat > "${TEST_DIR}/terraform.rc" << EOF
provider_installation {
  dev_overrides {
    "hashicorp/tfe" = "${PLUGIN_DIR}"
  }
  direct {}
}
EOF

cd "${TEST_DIR}"
export TF_CLI_CONFIG_FILE="$(pwd)/terraform.rc"

# Recurse across the examples directory
while IFS= read -r -d '' path; do
    relative_path="${path#${TARGET_DIR}/}"
    echo "Validating: ${relative_path}"

    # Copy file to TEST_DIR for validation
    cp "${path}" "${TEST_DIR}/main.tf"

    # Run terraform validate --json and capture output
    validate_output=$(terraform validate --json 2>&1 || true)

    # Inject formatting violation error if fmt check fails
    if ! terraform fmt -check >/dev/null 2>&1; then
        if ! validate_output=$(echo "${validate_output}" | jq '
            .diagnostics += [{
                "severity": "error",
                "summary": "Formatting violation",
                "detail": "File does not conform to terraform fmt standards"
            }]
        ' 2>&1); then
            # jq failed — swap to generic processing error
            validate_output="${JQ_ERROR_PROCESSING}"
        fi
    fi

    # Strip "Provider development overrides are in effect" — always present due to terraform.rc
    if ! validate_output=$(echo "${validate_output}" | jq '
        if .diagnostics then
            .diagnostics = [.diagnostics[] | select(.summary != "Provider development overrides are in effect")]
        else . end |
        if .diagnostics then
            .warning_count = ([.diagnostics[] | select(.severity == "warning")] | length) |
            .error_count = ([.diagnostics[] | select(.severity == "error")] | length)
        else . end
    ' 2>&1); then
        # jq failed — swap to generic processing error
        validate_output="${JQ_ERROR_PROCESSING}"
    fi

    # Apply file-specific exceptions
    if [ -n "${EXCEPTIONS_FILE}" ]; then
        if ! has_exceptions=$(jq --arg path "${relative_path}" '.file_exceptions | has($path)' "${EXCEPTIONS_FILE}"); then
            echo "Error: failed to read file_exceptions from ${EXCEPTIONS_FILE}" >&2
            exit 8
        fi

        if [ "${has_exceptions}" = "true" ]; then
            unmatched=$(echo "${validate_output}" | jq -r --arg path "${relative_path}" --slurpfile exceptions "${EXCEPTIONS_FILE}" '
                ($exceptions[0].file_exceptions[$path] // []) as $exception_list |
                ($exception_list - ([.diagnostics[]?.summary] | unique)) |
                .[]
            ' || echo "")
            if [ -n "${unmatched}" ]; then
                while IFS= read -r summary; do
                    UNUSED_EXCEPTIONS+=("${relative_path}: \"${summary}\"")
                done <<< "${unmatched}"
            fi

            if ! validate_output=$(echo "${validate_output}" | jq --arg path "${relative_path}" --slurpfile exceptions "${EXCEPTIONS_FILE}" '
                ($exceptions[0].file_exceptions[$path] // []) as $exception_list |
                if .diagnostics then
                    .diagnostics = [.diagnostics[] | select(.summary as $sum | ($exception_list | index($sum)) == null)]
                else . end |
                if .diagnostics then
                    .warning_count = ([.diagnostics[] | select(.severity == "warning")] | length) |
                    .error_count = ([.diagnostics[] | select(.severity == "error")] | length)
                else . end
            ' 2>&1); then
                # jq failed during exception filtering — swap to generic processing error
                validate_output="${JQ_ERROR_PROCESSING}"
            fi
        fi
    fi

    warning_count=$(echo "${validate_output}" | jq -r '.warning_count // 0')
    error_count=$(echo "${validate_output}" | jq -r '.error_count // 0')

    if [ "${error_count}" -gt 0 ]; then
        HAS_ERRORS=true
        while IFS=$'\t' read -r summary detail; do
            echo "  fail"
            echo "    ${summary}"
            [ -n "${detail}" ] && echo "    \"${detail}\""
            FAILURES+=("${relative_path}"$'\t'"${summary}"$'\t'"${detail}")
        done < <(echo "${validate_output}" | jq -r '.diagnostics[] | select(.severity=="error") | "\(.summary)\t\(.detail // "" | gsub("\n";" "))"')
    elif [ "${warning_count}" -gt 0 ]; then
        HAS_WARNINGS=true
        while IFS=$'\t' read -r summary detail; do
            echo "  warning"
            echo "    ${summary}"
            [ -n "${detail}" ] && echo "    \"${detail}\""
            WARNINGS+=("${relative_path}"$'\t'"${summary}"$'\t'"${detail}")
        done < <(echo "${validate_output}" | jq -r '.diagnostics[] | select(.severity=="warning") | "\(.summary)\t\(.detail // "" | gsub("\n";" "))"')
    else
        echo "  pass"
    fi

done < <(find "${TARGET_DIR}" -name "*.tf" -type f -print0)

cd "${PROVIDER_DIR}"

# Flag exception entries whose file path no longer exists under TARGET_DIR
if [ -n "${EXCEPTIONS_FILE}" ]; then
    while IFS= read -r key; do
        if [ ! -f "${TARGET_DIR}/${key}" ]; then
            while IFS= read -r summary; do
                UNUSED_EXCEPTIONS+=("${key}: \"${summary}\"")
            done < <(jq -r --arg p "${key}" '.file_exceptions[$p][]?' "${EXCEPTIONS_FILE}")
        fi
    done < <(jq -r '.file_exceptions | keys? | .[]?' "${EXCEPTIONS_FILE}")
fi

if [ ${#UNUSED_EXCEPTIONS[@]} -gt 0 ]; then
    for unused in "${UNUSED_EXCEPTIONS[@]}"; do
        WARNINGS+=("stale exception"$'\t'"file_exceptions: ${unused}"$'\t'"")
    done
fi

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------

echo ""
echo "========================================"
echo " Summary"
echo "========================================"

if [ ${#WARNINGS[@]} -gt 0 ]; then
    echo ""
    echo "Warnings (${#WARNINGS[@]}):"
    for w in "${WARNINGS[@]}"; do
        IFS=$'\t' read -r file summary detail <<< "${w}"
        echo "  - ${file}"
        echo "      ${summary}"
        [ -n "${detail}" ] && echo "      \"${detail}\""
    done
fi

if [ ${#FAILURES[@]} -gt 0 ]; then
    echo ""
    echo "Failures (${#FAILURES[@]}):"
    for f in "${FAILURES[@]}"; do
        IFS=$'\t' read -r file summary detail <<< "${f}"
        echo "  - ${file}"
        echo "      ${summary}"
        [ -n "${detail}" ] && echo "      \"${detail}\""
    done
fi

echo ""
if [ "${HAS_ERRORS}" = "true" ]; then
    echo "Result: FAILED"
    echo ""
    echo "To suppress known errors, add them to examples/error_exceptions.json"
    echo "under 'file_exceptions'. To run locally: ./scripts/validate-examples.sh"
    exit 5
elif [ "${HAS_WARNINGS}" = "true" ]; then
    echo "Result: WARNINGS"
    exit 3
elif [ ${#UNUSED_EXCEPTIONS[@]} -gt 0 ]; then
    echo "Result: WARNINGS"
    exit 4
else
    echo "Result: PASSED"
fi

exit 0
