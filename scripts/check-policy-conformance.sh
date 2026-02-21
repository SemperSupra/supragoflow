#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

fail() {
  echo "policy-conformance: $*" >&2
  exit 1
}

[[ -f "POLICY.md" ]] || fail "missing POLICY.md"
[[ -f "README.md" ]] || fail "missing README.md"
[[ -f "SECURITY.md" ]] || fail "missing SECURITY.md"

workflows=(.github/workflows/*.yml)
[[ ${#workflows[@]} -gt 0 ]] || fail "no workflow files found"

bad_uses=0
while IFS= read -r line; do
  if [[ "${line}" =~ uses:[[:space:]]+([^[:space:]]+)@([^[:space:]#]+) ]]; then
    ref="${BASH_REMATCH[2]}"
    if ! [[ "${ref}" =~ ^[0-9a-f]{40}$ ]]; then
      echo "policy-conformance: non-pinned action reference: ${line}" >&2
      bad_uses=1
    fi
  fi
done < <(grep -RhoE 'uses:[[:space:]]+[^[:space:]]+@[^\s]+' .github/workflows || true)
[[ "${bad_uses}" -eq 0 ]] || fail "workflow action references must be pinned to immutable commit SHAs"

if grep -RIn -- ':latest' .github/workflows >/dev/null 2>&1; then
  grep -RIn -- ':latest' .github/workflows >&2 || true
  fail "workflow files must not use implicit ':latest' tags"
fi

grep -q "All GitHub Actions must be pinned to immutable commit SHAs" POLICY.md || fail "POLICY.md missing action pinning policy statement"
grep -q "Release workflows publish explicit release tags only" POLICY.md || fail "POLICY.md missing explicit release tag policy statement"
grep -q "Container images must default to non-root execution identity" POLICY.md || fail "POLICY.md missing non-root container execution identity policy statement"
grep -Eq 'scripts/gg.*invoker UID:GID mapping' POLICY.md || fail "POLICY.md missing invoker UID:GID mapping policy statement"

for dockerfile in docker/Dockerfile.build docker/Dockerfile.dev; do
  [[ -f "${dockerfile}" ]] || fail "missing ${dockerfile}"
  grep -Eq '^USER[[:space:]]+' "${dockerfile}" || fail "${dockerfile} missing USER directive"
  if grep -Eq '^USER[[:space:]]+(root|0(:0)?)$' "${dockerfile}"; then
    fail "${dockerfile} must not use root USER"
  fi
done

grep -q -- '--user "${uid}:${gid}"' scripts/gg || fail "scripts/gg missing invoker UID:GID user mapping for container runs"

echo "policy-conformance: passed"
