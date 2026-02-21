#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SCRIPT="${ROOT}/scripts/release-tag-immutability.sh"

fail() {
  echo "FAIL: $*" >&2
  exit 1
}

assert_contains() {
  local haystack="$1"
  local needle="$2"
  local message="$3"
  [[ "$haystack" == *"$needle"* ]] || fail "$message (missing '$needle')"
}

run_case() {
  local __out_var="$1"
  local __status_var="$2"
  shift 2

  set +e
  local captured
  local rc
  captured="$("$@" 2>&1)"
  rc=$?
  set -e

  printf -v "$__out_var" '%s' "$captured"
  printf -v "$__status_var" '%s' "$rc"
}

out=""
status=0
image="ghcr.io/example/supragoflow-build:v1.2.3"

# Missing existing digest should allow publish.
run_case out status "$SCRIPT" evaluate "$image" "" "sha256:candidate"
[[ "$status" -eq 0 ]] || fail "missing existing digest should be allowed"
assert_contains "$out" "does not exist yet" "new tag path message"

# Existing digest equals candidate digest should allow idempotent publish.
run_case out status "$SCRIPT" evaluate "$image" "sha256:same" "sha256:same"
[[ "$status" -eq 0 ]] || fail "idempotent digest should be allowed"
assert_contains "$out" "idempotent publish allowed" "idempotent path message"

# Existing digest mismatch should fail without override.
run_case out status "$SCRIPT" evaluate "$image" "sha256:old" "sha256:new"
[[ "$status" -eq 1 ]] || fail "digest mismatch should fail without override"
assert_contains "$out" "release tag conflict" "conflict message"
assert_contains "$out" "next steps" "guidance message"

# Existing digest mismatch should pass with explicit override.
run_case out status env SUPRAGOFLOW_ALLOW_TAG_OVERWRITE=true "$SCRIPT" evaluate "$image" "sha256:old" "sha256:new"
[[ "$status" -eq 0 ]] || fail "digest mismatch should pass with override"
assert_contains "$out" "WARNING override enabled" "override warning message"

echo "release-tag-immutability tests passed"
