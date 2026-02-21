#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GG="${ROOT}/scripts/gg"

fail() {
  echo "FAIL: $*" >&2
  exit 1
}

assert_file() {
  local path="$1"
  [[ -f "${path}" ]] || fail "expected file: ${path}"
}

tmpdir="$(mktemp -d)"
trap 'rm -rf "${tmpdir}"' EXIT

out_dir="${tmpdir}/diag"
env \
  SUPRAGOFLOW_SECRET_TOKEN="topsecret" \
  SUPRAGOFLOW_API_KEY="abc123" \
  "${GG}" diagnose "${out_dir}" >/dev/null

assert_file "${out_dir}/summary.txt"
assert_file "${out_dir}/env.txt"
assert_file "${out_dir}/git-status.txt"
assert_file "${out_dir}/git-log.txt"
assert_file "${out_dir}/toolchain-versions.txt"
assert_file "${out_dir}/sizes.txt"
assert_file "${out_dir}/self-test.txt"
assert_file "${out_dir}.tar.gz"

grep -q 'SUPRAGOFLOW_SECRET_TOKEN=\[REDACTED\]' "${out_dir}/env.txt" || fail "secret token should be redacted"
grep -q 'SUPRAGOFLOW_API_KEY=\[REDACTED\]' "${out_dir}/env.txt" || fail "api key should be redacted"
if grep -q 'topsecret\|abc123' "${out_dir}/env.txt"; then
  fail "raw secret values must not appear in env diagnostics"
fi

bytes="$(wc -c < "${out_dir}.tar.gz")"
[[ "${bytes}" -lt 5242880 ]] || fail "diagnostics bundle unexpectedly large (${bytes} bytes)"

echo "gg diagnose tests passed"
