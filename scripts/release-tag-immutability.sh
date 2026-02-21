#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
usage:
  release-tag-immutability.sh resolve-existing <image_ref>
  release-tag-immutability.sh evaluate <image_ref> <existing_digest_or_empty> <candidate_digest>

environment:
  SUPRAGOFLOW_ALLOW_TAG_OVERWRITE=true|false  (default: false)
EOF
}

log() {
  printf '%s\n' "$*" >&2
}

resolve_existing() {
  local image_ref="${1:-}"
  [[ -n "${image_ref}" ]] || { usage; return 2; }

  local out
  set +e
  out="$(docker buildx imagetools inspect "${image_ref}" --format '{{json .Manifest.Digest}}' 2>&1)"
  local rc=$?
  set -e

  if [[ "${rc}" -eq 0 ]]; then
    printf '%s\n' "${out}" | tr -d '"' | tail -n1
    return 0
  fi

  # Not found is non-fatal: caller treats empty digest as "tag missing".
  if printf '%s\n' "${out}" | grep -Eiq 'not found|no such manifest|manifest unknown|name unknown'; then
    printf '\n'
    return 0
  fi

  log "immutability-check: failed to resolve existing digest for ${image_ref}"
  log "${out}"
  return 2
}

evaluate() {
  local image_ref="${1:-}"
  local existing_digest="${2:-}"
  local candidate_digest="${3:-}"
  local allow_overwrite="${SUPRAGOFLOW_ALLOW_TAG_OVERWRITE:-false}"

  [[ -n "${image_ref}" ]] || { usage; return 2; }
  [[ -n "${candidate_digest}" ]] || {
    log "immutability-check: candidate digest missing for ${image_ref}"
    return 2
  }

  if [[ -z "${existing_digest}" ]]; then
    log "immutability-check: ${image_ref} does not exist yet; publishing new immutable tag"
    return 0
  fi

  if [[ "${existing_digest}" == "${candidate_digest}" ]]; then
    log "immutability-check: ${image_ref} already points to digest ${candidate_digest}; idempotent publish allowed"
    return 0
  fi

  if [[ "${allow_overwrite}" == "true" ]]; then
    log "immutability-check: WARNING override enabled; updating ${image_ref} from ${existing_digest} to ${candidate_digest}"
    return 0
  fi

  log "immutability-check: release tag conflict for ${image_ref}"
  log "existing digest:  ${existing_digest}"
  log "candidate digest: ${candidate_digest}"
  log "next steps:"
  log "  1) publish a new release tag (recommended), or"
  log "  2) re-run with SUPRAGOFLOW_ALLOW_TAG_OVERWRITE=true only when overwrite is explicitly approved"
  return 1
}

main() {
  local cmd="${1:-}"
  case "${cmd}" in
    resolve-existing)
      shift
      resolve_existing "$@"
      ;;
    evaluate)
      shift
      evaluate "$@"
      ;;
    *)
      usage
      return 2
      ;;
  esac
}

main "$@"
