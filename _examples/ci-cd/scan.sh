#!/usr/bin/env bash
# scan.sh — Shell script wrapper for kspec scanning.
#
# This script provides a convenient wrapper for CI/CD systems or local use.
# It installs kspec (if needed), runs a scan, and checks the result.
#
# Usage:
#   ./scan.sh <provider> <asset-args...> [--policy-dir ./policies]
#
# Examples:
#   ./scan.sh network example.com
#   ./scan.sh network example.com --policy-dir ./my-policies
#   GITHUB_TOKEN=ghp_xxx ./scan.sh github org my-org

set -euo pipefail

# --- Configuration ---
POLICY_DIR="${POLICY_DIR:-./policies}"
EXPORT_PATH="${EXPORT_PATH:-report.json}"
SCORE_THRESHOLD="${SCORE_THRESHOLD:-0}"
KSPEC_FLAGS="${KSPEC_FLAGS:---no-ui}"

# --- Colors ---
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# --- Functions ---
info()  { echo -e "${GREEN}[INFO]${NC}  $*"; }
warn()  { echo -e "${YELLOW}[WARN]${NC}  $*"; }
error() { echo -e "${RED}[ERROR]${NC} $*" >&2; }

# --- Check prerequisites ---
if ! command -v kspec &> /dev/null; then
    info "kspec not found, installing..."
    go install github.com/kopexa-grc/kspec/cmd/kspec@latest
fi

if [ $# -lt 2 ]; then
    error "Usage: $0 <provider> <asset-args...>"
    error "Example: $0 network example.com"
    exit 1
fi

# --- Parse arguments ---
PROVIDER="$1"
shift

# Collect asset args (everything before any -- flags)
ASSET_ARGS=()
EXTRA_FLAGS=()
parsing_assets=true

for arg in "$@"; do
    if [[ "$arg" == --* ]]; then
        parsing_assets=false
    fi
    if $parsing_assets && [[ "$arg" != --* ]]; then
        ASSET_ARGS+=("$arg")
    else
        EXTRA_FLAGS+=("$arg")
    fi
done

# Check for --policy-dir override in extra flags
for i in "${!EXTRA_FLAGS[@]}"; do
    if [[ "${EXTRA_FLAGS[$i]}" == "--policy-dir" ]]; then
        POLICY_DIR="${EXTRA_FLAGS[$((i+1))]}"
    fi
done

# --- Validate ---
if [ ! -d "$POLICY_DIR" ]; then
    error "Policy directory not found: $POLICY_DIR"
    error "Create it or set POLICY_DIR environment variable."
    exit 1
fi

# --- Run scan ---
info "Running kspec scan: provider=$PROVIDER target=${ASSET_ARGS[*]}"
info "Policy directory: $POLICY_DIR"

# shellcheck disable=SC2086
kspec scan "$PROVIDER" "${ASSET_ARGS[@]}" \
    -d "$POLICY_DIR" \
    -o "$EXPORT_PATH" \
    $KSPEC_FLAGS || true

# --- Check results ---
if [ -f "$EXPORT_PATH" ] && command -v jq &> /dev/null; then
    SCORE=$(jq '.metadata.score' "$EXPORT_PATH")
    GRADE=$(jq -r '.metadata.grade' "$EXPORT_PATH")
    FAILED=$(jq '.metadata.failed_checks' "$EXPORT_PATH")

    info "Score: $SCORE/100 (Grade: $GRADE)"
    info "Failed checks: $FAILED"
    info "Report: $EXPORT_PATH"

    if [ "$SCORE" -lt "$SCORE_THRESHOLD" ]; then
        error "Score $SCORE is below threshold $SCORE_THRESHOLD"
        exit 1
    fi
else
    warn "Report not found or jq not available. Skipping score check."
fi

info "Done."
