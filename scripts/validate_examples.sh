#!/usr/bin/env bash
# Copyright IBM Corp. 2018, 2026
# SPDX-License-Identifier: MPL-2.0
#
# USAGE:
#   ./scripts/validate_examples.sh [-t <target directory; ./examples>] [-o <json holding failure data; no default] [-e <json holding errors to except; ./examples/error_exceptions>] [--help]


# Crash on error
set -e

# Variables with defaults
SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]}")"
JSON_OUTPUT_NAME="failure_info.json"
EXCEPTIONS_FILE=""
TEMP_BREAKING_INFO=""
CLEANUP_DONE=false
HAS_ERRORS=false
HAS_WARNINGS=false

# Normalize paths by resolving .. components
# Note that this makes a (strong) assumption about file structure
TARGET_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)/examples" # this default is set to ./examples
TEST_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)/temp-example-validation" # this is set to ./temp-example-validation

# Constants
JQ_ERROR_PROCESSING='{"diagnostics":[{"severity":"error","summary":"Processing error","detail":"jq processing failed"}],"error_count":1,"warning_count":0}'

# Arg parsing
while [[ $# -gt 0 ]]; do
    case $1 in
        -t|--target)
            TARGET_DIR="$2"
            shift 2
            ;;
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
            echo "Validates Terraform files in the target directory."
            echo ""
            echo "Options:"
            echo "  -t, --target DIR            Target directory containing .tf files (default: ./examples)"
            echo "  -o, --output FILE           Output JSON file for failures (default: failure_info.json)"
            echo "  -e, --exceptions FILE       JSON file with error/warning exceptions to ignore"
            echo "  -h, --help                  Show this help message"
            echo ""
            echo "Exit Codes:"
            echo "  0 - Complete success"
            echo "  3 - Warnings found in examples, no errors"
            echo "  4 - Errors found in examples"
            echo "  5 - Input files/directories do not exist"
            echo "  6 - Test directory already exists"
            echo "  7 - Required commands (terraform, jq) not found"
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
if ! command -v terraform >/dev/null 2>&1; then
    echo "Error: terraform command not found. Please install Terraform." >&2
    exit 7
fi

if ! command -v jq >/dev/null 2>&1; then
    echo "Error: jq command not found. Please install jq for JSON processing." >&2
    exit 7
fi

# Check if TARGET_DIR exists and make it absolute
if [ ! -d "${TARGET_DIR}" ]; then
    echo "Error: Target directory does not exist: ${TARGET_DIR}"
    exit 5
fi
TARGET_DIR="$(cd "${TARGET_DIR}" && pwd)"

# Exit if TEST_DIR already exists
if [ -d "${TEST_DIR}" ]; then
    echo "Error: Test directory already exists: ${TEST_DIR}"
    echo "Use of this script will destroy that directory. Please delete or rename this directory to use this script"
    exit 6
fi

# Check if EXCEPTIONS_FILE exists and make it absolute, if provided
if [ -n "${EXCEPTIONS_FILE}" ]; then
    if [ ! -f "${EXCEPTIONS_FILE}" ]; then
        echo "Error: Exceptions file does not exist: ${EXCEPTIONS_FILE}"
        exit 5
    fi
    # Convert to absolute path before we cd into TEST_DIR
    EXCEPTIONS_FILE="$(cd "$(dirname "${EXCEPTIONS_FILE}")" && pwd)/$(basename "${EXCEPTIONS_FILE}")"
fi

# Cleanup function and exception hook
cleanup() {
    # Remove temporary files
    if [ -n "${TEMP_BREAKING_INFO}" ] && [ -f "${TEMP_BREAKING_INFO}" ]; then
        rm -f "${TEMP_BREAKING_INFO}" "${TEMP_BREAKING_INFO}.single" "${TEMP_BREAKING_INFO}.new"
    fi
    if [ -d "${TEST_DIR}" ]; then
        rm -f "${TEST_DIR}/terraform.rc" "${TEST_DIR}/main.tf"
        rmdir "${TEST_DIR}" 2>/dev/null || true
    fi
}
trap cleanup EXIT INT TERM

# Temporary file to hold (potentially large) JSON
TEMP_BREAKING_INFO=$(mktemp)
echo "{}" > "${TEMP_BREAKING_INFO}"

# Create TEST_DIR and place terraform.rc
mkdir "${TEST_DIR}"
cat > "${TEST_DIR}/terraform.rc" << 'EOF'
provider_installation {
  dev_overrides {
    "registry.terraform.io/hashicorp/tfe" = "../"
  }
  direct {}
}
EOF

# Setup to run validate tests
cd "${TEST_DIR}"
export TF_CLI_CONFIG_FILE="$(pwd)/terraform.rc"

# Recurse across the target directory
while IFS= read -r -d '' path; do
    [ ! -f "${path}" ] && continue
    
    # Get path relative to target
    relative_path="${path#${TARGET_DIR}/}"
    
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
            }] |
            .warning_count = (.warning_count // 0) + 1
        ' 2>&1); then
            # jq failed, swap to jq error
            validate_output="${JQ_ERROR_PROCESSING}"
        fi
    fi
    
    # Always ignore "Provider development overrides are in effect" warning
    # this error comes from how we formulate .terraform.rc
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
        # Pull on the specific relative path
        has_exceptions=$(jq --arg path "${relative_path}" 'has($path)' "${EXCEPTIONS_FILE}" 2>/dev/null)
        
        # Then destroy those errors/warnings
        if [ "${has_exceptions}" = "true" ]; then
            if ! validate_output=$(echo "${validate_output}" | jq --arg path "${relative_path}" --slurpfile exceptions "${EXCEPTIONS_FILE}" '
                ($exceptions[0][$path] // []) as $exception_list |
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

# Return
cd "${TEST_DIR}/.."

# Collapse into the 'standard' error codes and make sorted json output should errors exist
if [ "${HAS_ERRORS}" = "true" ]; then
    jq -S '.' "${TEMP_BREAKING_INFO}" > "${JSON_OUTPUT_NAME}"
    echo "Validation errors found. See ${JSON_OUTPUT_NAME} for details."
    exit 4
elif [ "${HAS_WARNINGS}" = "true" ]; then
    jq -S '.' "${TEMP_BREAKING_INFO}" > "${JSON_OUTPUT_NAME}"
    echo "Warnings found. See ${JSON_OUTPUT_NAME} for details."
    exit 3
fi

echo "All validations passed successfully."
exit 0
