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
- `gg test`

## Builds (build image)

`gg build <goos> <goarch>` produces deterministic artifacts into `dist/<goos>-<goarch>/` using:

- `-trimpath`
- `CGO_ENABLED=0` by default

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
- The system follows a "pull first" policy: builds attempt to pull images from GHCR before building locally, unless developing local changes that require a rebuild.
- Users/agents should prefer GHCR release tags over local images.

## Contribution policy (Option C)

Repo is public but enforced as invite-only:

- Issues and PRs opened by non-allowed actors are commented on and closed.
- Fork PRs are rejected unless the author is in `designated_fork_contributors`.
- The standard allowlist (humans/agents) does **not** grant fork PR privileges.

See `.github/allowlist.yml` and the enforcement workflow.
