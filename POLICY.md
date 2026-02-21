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

## Builds (build image)

`gg build <goos> <goarch>` produces deterministic artifacts into `dist/<goos>-<goarch>/` using:

- `-trimpath`
- `CGO_ENABLED=0` by default
- Optional semantic naming via `--output-template` with tokens `{name}`, `{version}`, `{os}`, `{arch}`, `{ext}`.
- Reproducible build metadata is supported via `SUPRAGOFLOW_BUILD_DATE` (fixed RFC3339 UTC timestamp).

## Machine-readable output compatibility

- `supragoflow --version --json` includes `schemaVersion`.
- Consumers should treat `schemaVersion` changes as contract changes requiring compatibility review.

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
- Users/agents should prefer GHCR release tags over local images.
- `scripts/gg` only attempts remote pulls when `SUPRAGOFLOW_IMAGE_TAG` is set.
- CI may set repository variable `SUPRAGOFLOW_IMAGE_TAG` to a canonical release tag to enable pull-first image reuse before local fallback build.
- `gg smoke-windows` requires an explicit runner image via `SUPRAGOFLOW_WINE_RUNNER_IMAGE`.

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
