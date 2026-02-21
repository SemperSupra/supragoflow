#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GG="${ROOT}/scripts/gg"

fail() {
  echo "FAIL: $*" >&2
  exit 1
}

assert_contains() {
  local haystack="$1"
  local needle="$2"
  local message="$3"
  [[ "${haystack}" == *"${needle}"* ]] || fail "${message} (missing '${needle}')"
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

  printf -v "${__out_var}" '%s' "${captured}"
  printf -v "${__status_var}" '%s' "${rc}"
}

out=""
status=0

run_and_capture out status "${GG}" contract
[[ "${status}" -eq 0 ]] || fail "gg contract should succeed"
assert_contains "${out}" "\"interfaceVersion\": \"1\"" "contract should include interface version"
assert_contains "${out}" "\"logSchemaVersion\": \"1\"" "contract should include log schema version"
assert_contains "${out}" "\"versionSchemaVersion\": \"1\"" "contract should include version schema version"

run_and_capture out status env SUPRAGOFLOW_EXPECT_GG_INTERFACE_VERSION=999 "${GG}" contract
[[ "${status}" -eq 2 ]] || fail "gg interface mismatch should exit 2"
assert_contains "${out}" "gg interface compatibility mismatch" "gg interface mismatch message"

run_and_capture out status env SUPRAGOFLOW_EXPECT_LOG_SCHEMA_VERSION=999 "${GG}" contract
[[ "${status}" -eq 2 ]] || fail "log schema mismatch should exit 2"
assert_contains "${out}" "log schema compatibility mismatch" "log schema mismatch message"

echo "gg contract tests passed"
