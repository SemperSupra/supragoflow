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

tmpdir="$(mktemp -d)"
trap 'rm -rf "${tmpdir}"' EXIT

cat > "${tmpdir}/docker" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

echo "$*" >> "${FAKE_DOCKER_LOG}"

cmd="${1:-}"
shift || true

case "${cmd}" in
  image)
    if [[ "${1:-}" == "inspect" ]]; then
      exit 1
    fi
    ;;
  pull)
    count=0
    if [[ -f "${FAKE_PULL_COUNT_FILE}" ]]; then
      count="$(cat "${FAKE_PULL_COUNT_FILE}")"
    fi
    count=$((count + 1))
    echo "${count}" > "${FAKE_PULL_COUNT_FILE}"
    sleep "${FAKE_PULL_SLEEP_SEC:-0}"
    if [[ "${count}" -lt "${FAKE_PULL_SUCCEED_ON:-999}" ]]; then
      exit 1
    fi
    exit 0
    ;;
  build|tag|run|system)
    exit 0
    ;;
esac

exit 0
EOF
chmod +x "${tmpdir}/docker"

out=""
status=0

run_and_capture out status env \
  PATH="${tmpdir}:${PATH}" \
  FAKE_DOCKER_LOG="${tmpdir}/docker.log" \
  FAKE_PULL_COUNT_FILE="${tmpdir}/pull.count" \
  FAKE_PULL_SUCCEED_ON=3 \
  FAKE_PULL_SLEEP_SEC=0 \
  SUPRAGOFLOW_BUILD_IMAGE_REF="ghcr.io/test/supragoflow-build@sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" \
  SUPRAGOFLOW_PULL_RETRIES=3 \
  SUPRAGOFLOW_PULL_BACKOFF_SEC=0 \
  "${GG}" deps

[[ "${status}" -eq 0 ]] || fail "retry path should eventually succeed"
assert_contains "${out}" "docker pull attempt 1/3 failed" "retry attempt 1 message"
assert_contains "${out}" "docker pull attempt 2/3 failed" "retry attempt 2 message"

run_and_capture out status env \
  PATH="${tmpdir}:${PATH}" \
  FAKE_DOCKER_LOG="${tmpdir}/docker-timeout.log" \
  FAKE_PULL_COUNT_FILE="${tmpdir}/pull-timeout.count" \
  FAKE_PULL_SUCCEED_ON=999 \
  FAKE_PULL_SLEEP_SEC=2 \
  SUPRAGOFLOW_BUILD_IMAGE_REF="ghcr.io/test/supragoflow-build@sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" \
  SUPRAGOFLOW_PULL_RETRIES=1 \
  SUPRAGOFLOW_PULL_BACKOFF_SEC=0 \
  SUPRAGOFLOW_TIMEOUT_PULL_SEC=1 \
  "${GG}" deps

[[ "${status}" -eq 0 ]] || fail "timeout fallback path should complete via local image build"
assert_contains "${out}" "timeout: docker pull" "timeout message should be present"

echo "gg fault-injection tests passed"
