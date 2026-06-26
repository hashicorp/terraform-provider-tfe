#!/usr/bin/env bash
# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0
#
# Usage:
#   ./scripts/validate-examples.sh [-o <json holding failure data; ./examples/failure_info.json] [-e <json holding errors to except; no default] [--help]
#
# Exit codes:
#  0 - Complete success
#  1 - Generic errors (including command line args)
#  3 - Warnings found in examples, no errors
#  4 - Warning that unused exceptions were found in error_exceptions.json
#  5 - Errors found in examples
#  6 - Required commands (terraform, jq) not found
#  7 - Input files/directories do not exist
#  8 - Internal data merge error


# Crash on error
set -e

# Variables with defaults
SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]}")"
JSON_OUTPUT_NAME="failure_info.json"
EXCEPTIONS_FILE=""
TEMP_BREAKING_INFO=""
HAS_ERRORS=false
HAS_WARNINGS=false
UNUSED_EXCEPTIONS=()

# Normalize paths by resolving .. components
# Note that this makes a (strong) assumption about file structure
PROVIDER_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
TARGET_DIR="${PROVIDER_DIR}/examples"

# Constants
JQ_ERROR_PROCESSING='{"diagnostics":[{"severity":"error","summary":"Processing error","detail":"jq processing failed"}],"error_count":1,"warning_count":0}'

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
            echo "Validates Terraform files in the examples directory."
            echo ""
            echo "Options:"
            echo "  -o, --output FILE           Output JSON file for failures (default: ./failure_info.json)"
            echo "  -e, --exceptions FILE       JSON file with error/warning exceptions to ignore"
            echo "  -h, --help                  Show this help message"
            echo ""
            echo "Exit Codes:"
            echo "  0 - Complete success"
            echo "  3 - Warnings found in examples, no errors"
            echo "  4 - Unused exceptions found in error_exceptions.json"
            echo "  5 - Errors found in examples"
            echo "  6 - Required commands (terraform, jq) not found"
            echo "  7 - Input files/directories do not exist"
            echo "  8 - Internal data merge error"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use -h or --help for usage information"
            exit 1
            ;;
    esac
done

# Terraform and jq dependencies
# These can erroneously pass if the command name exists, but don't refer to the real tool
if ! command -v terraform >/dev/null 2>&1; then
    echo "Error: terraform command not found. Please install Terraform." >&2
    exit 6
fi

if ! command -v jq >/dev/null 2>&1; then
    echo "Error: jq command not found. Please install jq for JSON processing." >&2
    exit 6
fi

# Check if TARGET_DIR exists
if [ ! -d "${TARGET_DIR}" ]; then
    echo "Error: Examples directory does not exist: ${TARGET_DIR}"
    exit 7
fi

# Check if EXCEPTIONS_FILE exists and make it absolute, if provided
if [ -n "${EXCEPTIONS_FILE}" ]; then
    if [ ! -f "${EXCEPTIONS_FILE}" ]; then
        echo "Error: Exceptions file does not exist: ${EXCEPTIONS_FILE}"
        exit 7
    fi
    EXCEPTIONS_FILE="$(cd "$(dirname "${EXCEPTIONS_FILE}")" && pwd)/$(basename "${EXCEPTIONS_FILE}")"
fi

# Create temp working directory and register cleanup
TEST_DIR=$(mktemp -d)
TEMP_BREAKING_INFO=""

cleanup() {
    if [ -n "${TEMP_BREAKING_INFO}" ] && [ -f "${TEMP_BREAKING_INFO}" ]; then
        rm -f "${TEMP_BREAKING_INFO}" "${TEMP_BREAKING_INFO}.single" "${TEMP_BREAKING_INFO}.new"
    fi
    [ -n "${TEST_DIR}" ] && rm -rf "${TEST_DIR}"
}
trap cleanup EXIT INT TERM

# Temporary file to hold JSON
TEMP_BREAKING_INFO=$(mktemp)
echo "{}" > "${TEMP_BREAKING_INFO}"

# Place terraform.rc using PROVIDER_DIR as an absolute path so the dev
# override resolves correctly regardless of where TEST_DIR is located.
cat > "${TEST_DIR}/terraform.rc" << EOF
provider_installation {
  dev_overrides {
    "registry.terraform.io/hashicorp/tfe" = "${PROVIDER_DIR}"
  }
  direct {}
}
EOF

# Setup to run validate tests
cd "${TEST_DIR}"
export TF_CLI_CONFIG_FILE="$(pwd)/terraform.rc"

# Recurse across the examples directory
while IFS= read -r -d '' path; do
    relative_path="${path#${TARGET_DIR}/}"
    echo "Validating: ${relative_path}"
    
    # Copy file to TEST_DIR
    cp "${path}" "${TEST_DIR}/main.tf"
    
    # Run terraform validate --json and capture
    validate_output=$(terraform validate --json 2>&1 || true)
    
    # Inject formatting violation warning if fmt check fails
    if ! terraform fmt -check >/dev/null 2>&1; then
        if ! validate_output=$(echo "${validate_output}" | jq '
            .diagnostics += [{
                "severity": "warning",
                "summary": "Formatting violation",
                "detail": "File does not conform to terraform fmt standards"
            }]
        ' 2>&1); then
            # jq failed, swap to jq error
            validate_output="${JQ_ERROR_PROCESSING}"
        fi
    fi
    
    # Always ignore "Provider development overrides are in effect" warning
    # this warning comes from how we formulate terraform.rc
    if ! validate_output=$(echo "${validate_output}" | jq '
        if .diagnostics then
            .diagnostics = [.diagnostics[] | select(.summary != "Provider development overrides are in effect")]
        else . end |
        if .diagnostics then
            .warning_count = ([.diagnostics[] | select(.severity == "warning")] | length) |
            .error_count = ([.diagnostics[] | select(.severity == "error")] | length)
        else . end
    ' 2>&1); then
        # jq failed, swap to jq error
        validate_output="${JQ_ERROR_PROCESSING}"
    fi
    
    # Apply file-specific exceptions, if provided
    if [ -n "${EXCEPTIONS_FILE}" ]; then
        # Check whether this file has any exceptions defined
        has_exceptions=$(jq --arg path "${relative_path}" '.file_exceptions | has($path)' "${EXCEPTIONS_FILE}" 2>/dev/null)
        
        # Then destroy those errors/warnings
        if [ "${has_exceptions}" = "true" ]; then
            # Find which individual exception summaries did not match any diagnostic
            unmatched=$(echo "${validate_output}" | jq -r --arg path "${relative_path}" --slurpfile exceptions "${EXCEPTIONS_FILE}" '
                ($exceptions[0].file_exceptions[$path] // []) as $exception_list |
                ($exception_list - ([.diagnostics[]?.summary] | unique)) |
                .[]
            ' 2>/dev/null || echo "")
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
                # jq failed during exception filtering, swap to jq error
                validate_output="${JQ_ERROR_PROCESSING}"
            fi
        fi
    fi
    
    # Check if there are any warnings or errors remaining
    warning_count=$(echo "${validate_output}" | jq -r '.warning_count // 0')
    error_count=$(echo "${validate_output}" | jq -r '.error_count // 0')
    
    if [ "${warning_count}" -gt 0 ] || [ "${error_count}" -gt 0 ]; then
        # Flag if any error or warning exists
        [ "${error_count}" -gt 0 ] && HAS_ERRORS=true
        [ "${warning_count}" -gt 0 ] && HAS_WARNINGS=true
        
        # Store result in JSON
        echo "${validate_output}" > "${TEMP_BREAKING_INFO}.single"
        if jq --arg path "${relative_path}" --slurpfile single "${TEMP_BREAKING_INFO}.single" \
            '. + {($path): $single[0]}' "${TEMP_BREAKING_INFO}" > "${TEMP_BREAKING_INFO}.new"; then
            mv "${TEMP_BREAKING_INFO}.new" "${TEMP_BREAKING_INFO}"
        else
            # In the case where this jq merge fails, we include it here
            # Since this failure is almost certainly a broad failure and
            # invalidating of other work, we exit fatally
            echo "Failed to merge validation results for ${relative_path}" >&2
            exit 8
        fi
        rm -f "${TEMP_BREAKING_INFO}.single"
    fi
done < <(find "${TARGET_DIR}" -name "*.tf" -type f -print0) # Find next .tf file


cd "${PROVIDER_DIR}"

# Flag exception entries whose file path does not exist under TARGET_DIR
if [ -n "${EXCEPTIONS_FILE}" ]; then
    while IFS= read -r key; do
        if [ ! -f "${TARGET_DIR}/${key}" ]; then
            while IFS= read -r summary; do
                UNUSED_EXCEPTIONS+=("${key}: \"${summary}\"")
            done < <(jq -r --arg p "${key}" '.file_exceptions[$p][]' "${EXCEPTIONS_FILE}" 2>/dev/null)
        fi
    done < <(jq -r '.file_exceptions | keys[]' "${EXCEPTIONS_FILE}" 2>/dev/null)
fi

# Report unused exceptions
# This happens separately so that we always get this output
if [ ${#UNUSED_EXCEPTIONS[@]} -gt 0 ]; then
    echo ""
    echo "The following exception entries did not match any diagnostic:"
    echo ""
    for unused in "${UNUSED_EXCEPTIONS[@]}"; do
        echo "  - ${unused}"
    done
    echo ""
    echo "Consider removing these entries from file_exceptions in error_exceptions.json"
    echo ""
fi

# Collapse into the 'standard' exit codes and make sorted json output should errors exist
if [ "${HAS_ERRORS}" = "true" ]; then
    jq -S '.' "${TEMP_BREAKING_INFO}" > "${JSON_OUTPUT_NAME}"
    echo "Validation errors found. See ${JSON_OUTPUT_NAME} for details."
    exit 5
elif [ "${HAS_WARNINGS}" = "true" ]; then
    jq -S '.' "${TEMP_BREAKING_INFO}" > "${JSON_OUTPUT_NAME}"
    echo "Warnings found. See ${JSON_OUTPUT_NAME} for details."
    exit 3
elif [ ${#UNUSED_EXCEPTIONS[@]} -gt 0 ]; then
    exit 4
fi

echo "All validations passed successfully."
exit 0
