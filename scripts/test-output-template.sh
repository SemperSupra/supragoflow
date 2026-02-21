#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GG="$ROOT/scripts/gg"

# Keep tests deterministic even if caller has local overrides set.
unset SUPRAGOFLOW_OUTPUT_TEMPLATE
unset SUPRAGOFLOW_IMAGE_TAG
unset SUPRAGOFLOW_WINE_RUNNER_IMAGE

fail() {
  echo "FAIL: $*" >&2
  exit 1
}

assert_eq() {
  local expected="$1"
  local actual="$2"
  local message="$3"
  [[ "$actual" == "$expected" ]] || fail "$message (expected '$expected', got '$actual')"
}

assert_contains() {
  local haystack="$1"
  local needle="$2"
  local message="$3"
  [[ "$haystack" == *"$needle"* ]] || fail "$message (missing '$needle')"
}

run_and_capture() {
  local __out_var="$1"
  local __status_var="$2"
  shift 2

  set +e
  local captured
  local rc
  captured="$($@ 2>&1)"
  rc=$?
  set -e

  printf -v "$__out_var" '%s' "$captured"
  printf -v "$__status_var" '%s' "$rc"
}

out=""
status=0

out="$(SUPRAGOFLOW_BUILD_VERSION=v1.2.3 "$GG" build linux amd64 --print-output-name)"
assert_eq "supragoflow" "$out" "default template output name"

out="$(SUPRAGOFLOW_BUILD_VERSION=v1.2.3 "$GG" build windows amd64 --output-template '{name}-{version}-{os}-{arch}{ext}' --print-output-name)"
assert_eq "supragoflow-v1.2.3-windows-amd64.exe" "$out" "templated output name"

out="$(SUPRAGOFLOW_BUILD_VERSION=v9.9.9 "$GG" build linux arm64 --output-template '{name}-{os}-{arch}' --print-output-name)"
assert_eq "supragoflow-linux-arm64" "$out" "custom template without ext"

run_and_capture out status env SUPRAGOFLOW_BUILD_VERSION=v1.2.3 "$GG" build linux amd64 --output-template '{name}-{bad}' --print-output-name
[[ "$status" -eq 2 ]] || fail "unknown token should exit with status 2"
assert_contains "$out" "unknown token" "unknown token validation message"

run_and_capture out status env SUPRAGOFLOW_BUILD_VERSION=v1.2.3 "$GG" build linux amd64 --output-template '{name}/{os}' --print-output-name
[[ "$status" -eq 2 ]] || fail "path separators should exit with status 2"
assert_contains "$out" "must not contain path separators" "path separator validation message"

run_and_capture out status env SUPRAGOFLOW_BUILD_VERSION=v1.2.3 "$GG" build linux amd64 --output-template '{name}' --unknown-option --print-output-name
[[ "$status" -eq 2 ]] || fail "unknown build option should exit with status 2"
assert_contains "$out" "unknown build option" "unknown option validation message"

run_and_capture out status env SUPRAGOFLOW_BUILD_VERSION=v1.2.3 "$GG" --log-format json --log-level debug build linux amd64 --print-output-name
[[ "$status" -eq 0 ]] || fail "json logging self-test run should succeed"
assert_contains "$out" "\"logSchemaVersion\":\"1\"" "json log schema version"
assert_contains "$out" "\"run_id\":\"" "json log run id field"

echo "output-template tests passed"
