# SupraGoFlow Policies

## Single entrypoint

All automation (humans, agents, CI, devcontainer) must use:

- `./scripts/gg <stage>`

CI workflows should not invent bespoke command sequences beyond calling `gg`.

## Toolchain pinning

All toolchain versions are pinned in the `.versions` file at the root of the repository.

- `DEBIAN_BASE`: Base image for all containers.
- `GO_VERSION`: Go version used for building and development.
- `GOLANGCI_LINT_VERSION`: Version of `golangci-lint`.
- `GOVULNCHECK_VERSION`: Version of `govulncheck`.
- `GOTESTSUM_VERSION`: Version of `gotestsum`.

To upgrade any tool, update the corresponding version in `.versions` and submit a PR.
The `scripts/gg` tool automatically uses these versions when building images.

## Dependency discipline

- `gg deps` uses `go mod download` and **must not** modify `go.mod`/`go.sum`.
- If imports change, run `gg tidy` to update `go.mod`/`go.sum`.
- CI may enforce "tidy produces no diff" (optional).

## Gates (dev image)

Typical gates for incremental development:

- `gg fmt` (format)
- `gg vet`
- `gg lint`
- `gg vuln`
- `gg self-test` (script-level harness checks)
- `gg test`

## Logging and diagnostics

- `scripts/gg` supports tunable logging with `--log-level` (`debug|info|warn|error|none`) and `--log-format` (`text|json`).
- Each `gg` invocation includes a run correlation id (`--run-id` or auto-generated) in log messages.
- JSON log output includes `logSchemaVersion` for compatibility-aware consumers.
- Informational diagnostics should be bounded; avoid repeating non-actionable messages across stages.
- Timeouts and pull retry/backoff are configurable to bound long-running operations (`SUPRAGOFLOW_TIMEOUT_*`, `SUPRAGOFLOW_PULL_RETRIES`, `SUPRAGOFLOW_PULL_BACKOFF_SEC`).
- Long-running operations provide liveness/progress feedback with configurable heartbeat interval (`SUPRAGOFLOW_HEARTBEAT_SEC`).
- Cache strategy is configurable (`SUPRAGOFLOW_CACHE_STRATEGY=volume|host`); CI should prefer `host` to enable cache restore/save across runs.
- Host cache should be size-bounded in automation (`SUPRAGOFLOW_HOST_CACHE_MAX_MB`) and pruned with `gg cache-prune`.
- `gg diagnose` should produce bounded diagnostics bundles suitable for corrective maintenance and CI triage.

## Builds (build image)

`gg build <goos> <goarch>` produces deterministic artifacts into `dist/<goos>-<goarch>/` using:

- `-trimpath`
- `CGO_ENABLED=0` by default
- Optional semantic naming via `--output-template` with tokens `{name}`, `{version}`, `{os}`, `{arch}`, `{ext}`.
- Reproducible build metadata is supported via `SUPRAGOFLOW_BUILD_DATE` (fixed RFC3339 UTC timestamp).

## Machine-readable output compatibility

- `supragoflow --version --json` includes `schemaVersion`.
- Consumers should treat `schemaVersion` changes as contract changes requiring compatibility review.
- `gg contract` publishes the machine-readable lifecycle interface contract (`contracts/gg-interface.json`).
- Strict fail-fast compatibility checks are supported via `SUPRAGOFLOW_EXPECT_SCHEMA_VERSION`, `SUPRAGOFLOW_EXPECT_LOG_SCHEMA_VERSION`, and `SUPRAGOFLOW_EXPECT_GG_INTERFACE_VERSION`.

## Security library/framework selection

- For security-sensitive features (transport security, authentication, key management, crypto protocols), prefer proven and widely adopted implementations.
- Prefer standard library primitives first when they satisfy requirements.
- For third-party dependencies, prefer projects with a strong maintenance record, broad production usage, and clear security posture.
- Avoid introducing niche or experimental security frameworks without explicit maintainer approval in PR discussion.

## Windows compatibility requirement

- Windows-targeted binaries must work on both native Windows and Wine in CI smoke checks.
- PRs that change runtime behavior for Windows builds should include/update coverage for both environments.
- If a dependency is known to break on Wine or native Windows, it is not an acceptable default dependency for Windows paths.

## Canonical releases

- Container images are built and pushed to GHCR **only** on GitHub Release (`release.published`).
- Release workflows publish explicit release tags only; `:latest` is not part of the canonical path.
- Release GHCR tags are immutable by default: publish fails on existing-tag digest mismatch.
- Users/agents should prefer GHCR release tags over local images.
- `scripts/gg` supports pull-first image reuse via explicit refs:
  - Preferred: `SUPRAGOFLOW_BUILD_IMAGE_REF` and `SUPRAGOFLOW_DEV_IMAGE_REF` (digest pinning supported).
  - Fallback: `SUPRAGOFLOW_IMAGE_TAG`.
- Emergency override is explicit-only via `SUPRAGOFLOW_ALLOW_TAG_OVERWRITE=true` in release workflow context.
- CI may set repository variable `SUPRAGOFLOW_IMAGE_TAG` to a canonical release tag to enable pull-first image reuse before local fallback build.
- `gg smoke-windows` requires an explicit runner image via `SUPRAGOFLOW_WINE_RUNNER_IMAGE`.

## Container execution identity

- Container images must default to non-root execution identity (`USER` must not be root).
- `scripts/gg` must run containerized stages with invoker UID:GID mapping to preserve host artifact ownership.
- Writable roots must be validated before/after containerized stages with `gg verify-writable`.
- Current exception: short-lived volume-ownership preparation may run as root to align named-volume ownership before stage execution.
- Any additional root-execution exception must be explicitly documented in-policy and time-bounded.

## CI tiering

- Pull requests run a fast gate by default (format/vet/lint/test/linux build).
- Full gate (including vulnerability scan and Wine-based Windows smoke) runs on `main` pushes.
- CI includes repository secret scanning.
- All GitHub Actions must be pinned to immutable commit SHAs.
- CI includes policy conformance checks (`scripts/check-policy-conformance.sh`).
- CI should validate stored configuration and writable artifact roots before/after containerized stages (`gg validate-config`, `gg verify-writable`).

## Release security

- Release publish job uses protected `release` environment controls.
- Release workflow publishes SBOM assets for release images and binaries.
## Service discoverability

- Local service discovery (Bonjour/mDNS) is not applicable to the current CLI-only architecture.
- Internet/runtime service discovery is not applicable to the current CLI-only architecture.
- If a network service mode is introduced in future, any discovery mechanism must be explicit opt-in and undergo security review before enablement.

## Contribution policy (Option C)

Repo is public but enforced as invite-only:

- Issues and PRs opened by non-allowed actors are commented on and closed.
- Fork PRs are rejected unless the author is in `designated_fork_contributors`.
- The standard allowlist (humans/agents) does **not** grant fork PR privileges.

See `.github/allowlist.yml` and the enforcement workflow.
