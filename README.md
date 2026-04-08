# SupraGoFlow

SupraGoFlow is a git-native, containerized lifecycle for lightweight Go projects (CLI tools, services, bots, telemetry agents).
It provides a **single lifecycle entrypoint** (`./scripts/gg`) that works for:

- Humans on the command line
- CLI agents
- CI/CD pipelines
- VS Code Dev Containers

Targets: **Linux + Windows** (cross-compile). No GUI and no heavy CGO.

## Repository location note

This project lives in the repository root directory.
Keep local scratch artifacts outside the project directory so they do not get committed.

## Secure comms package

This repo now includes `internal/securecomms` with proven primitives for secure channel development:

- TLS client/server config builders (including optional mutual TLS)
- SSH client config builder with strict `known_hosts` validation

See [docs/securecomms.md](docs/securecomms.md) for detailed documentation and examples.

The implementation prefers Go standard library crypto/TLS and `golang.org/x/crypto/ssh`.

## Canonical releases (GHCR)

**Prefer GHCR-published images tagged to GitHub Releases as the canonical toolchain.**
Local builds (`*:local`) are for development only and are not authoritative.

Images published on release:

- `ghcr.io/<org>/supragoflow-build:<tag>`
- `ghcr.io/<org>/supragoflow-dev:<tag>`

> Containers are built and pushed **only on GitHub Releases** and only as explicit release tags (no implicit `:latest` flow).
> Release tags are immutable by default; conflicting digest republishes fail unless explicit emergency override is set in release workflow policy.
> To pull release images in `gg`, set `SUPRAGOFLOW_IMAGE_TAG=<release-tag>`.

## Two images

- **Build image** (minimal): deterministic builds for mature/proven code.
- **Dev image** (robust but slim): adds linter + vuln tooling for incremental development.

## Quickstart (local with Docker)

```bash
./scripts/gg bootstrap
./scripts/gg deps
./scripts/gg lint
./scripts/gg test
./scripts/gg build linux amd64
./scripts/gg build windows amd64
./scripts/gg build windows 386
```

Outputs land in `dist/<goos>-<goarch>/`.

## WBAB Compatibility

This project is fully compatible with [WineBotAppBuilder (WBAB)](https://github.com/SemperSupra/WineBotAppBuilder) workflows.
It produces artifacts in the standard `dist/` directory, making it suitable for inclusion in WBAB pipelines for packaging and distribution.
The `scripts/gg` toolchain adheres to containerized build principles, ensuring deterministic outputs across environments.

## Development Commands

The `./scripts/gg` entrypoint supports the following stages:

| Command | Description |
| :--- | :--- |
| `bootstrap` | Verify environment tools and go version |
| `deps` | Download dependencies (`go mod download`) |
| `tidy` | Tidy dependencies (`go mod tidy`) |
| `fmt` | Format code (`go fmt`) |
| `vet` | Run static analysis (`go vet`) |
| `lint` | Run linter (`golangci-lint`) |
| `vuln` | Run vulnerability check (`govulncheck`) |
| `self-test` | Run local shell-level harness checks (non-Docker) |
| `test` | Run tests |
| `contract` | Print machine-readable `gg` interface contract manifest |
| `validate-config` | Validate required stored configuration inputs |
| `verify-writable [path...]` | Verify writable/owned output paths for current user |
| `targets` | List supported build targets |
| `build <goos> <goarch> [--output-template <template>]` | Build binary for target OS/Arch |
| `package` | Package artifacts (create checksums) |
| `images` | Build local Docker images |
| `cache-prune` | Prune host cache when above configured limit |
| `diagnose [out_dir]` | Generate bounded diagnostics bundle and compressed archive |
| `smoke-windows [goarch]` | Run Windows binary in Wine (default: amd64) |

Global `gg` logging options (place before stage): `--log-level <debug|info|warn|error|none>`, `--log-format <text|json>`, `--run-id <id>`.
When `--log-format json` is used, logs include `logSchemaVersion` and `run_id` fields.
Optional strict compatibility checks are supported with:
- `SUPRAGOFLOW_EXPECT_LOG_SCHEMA_VERSION`
- `SUPRAGOFLOW_EXPECT_GG_INTERFACE_VERSION`

Runtime bounds are configurable with env vars, including:
`SUPRAGOFLOW_TIMEOUT_STAGE_SEC`, `SUPRAGOFLOW_TIMEOUT_PULL_SEC`, `SUPRAGOFLOW_TIMEOUT_IMAGE_BUILD_SEC`, `SUPRAGOFLOW_TIMEOUT_SMOKE_SEC`, `SUPRAGOFLOW_PULL_RETRIES`, `SUPRAGOFLOW_PULL_BACKOFF_SEC`.
Liveness heartbeat interval for long operations is configurable via `SUPRAGOFLOW_HEARTBEAT_SEC` (set `0` to disable periodic in-progress logs).
Resource telemetry in heartbeat logs is enabled by default via `SUPRAGOFLOW_HEARTBEAT_RESOURCE=1` and includes process CPU/memory plus host memory/load/disk usage; set `SUPRAGOFLOW_HEARTBEAT_RESOURCE=0` to disable resource fields.
Cache behavior is configurable via `SUPRAGOFLOW_CACHE_STRATEGY`:
- `volume` (default): Docker named volumes for local iterative runs
- `host`: bind mounts under `SUPRAGOFLOW_HOST_CACHE_ROOT` (default `.cache/supragoflow`) for CI cache restore/save compatibility
Host cache can be bounded with `SUPRAGOFLOW_HOST_CACHE_MAX_MB` and pruned via `./scripts/gg cache-prune`.
For failure analysis, `./scripts/gg diagnose` writes a bounded diagnostics bundle under `dist/diagnostics/`.

`smoke-windows` requires `SUPRAGOFLOW_WINE_RUNNER_IMAGE` to be set explicitly.
To enable pull-first CI image reuse from GHCR release images, set one of:
- `SUPRAGOFLOW_BUILD_IMAGE_REF` / `SUPRAGOFLOW_DEV_IMAGE_REF` (preferred; supports digest-pinned refs)
- `SUPRAGOFLOW_IMAGE_TAG` (tag-based fallback)

`build` supports semantic artifact naming templates with tokens: `{name}`, `{version}`, `{os}`, `{arch}`, `{ext}`.
For reproducible metadata, `SUPRAGOFLOW_BUILD_DATE` can be set to a fixed RFC3339 UTC timestamp used for embedded build date.
Example:

```bash
SUPRAGOFLOW_BUILD_VERSION=v1.2.3 ./scripts/gg build windows amd64 --output-template '{name}-{version}-{os}-{arch}{ext}'
```

`supragoflow --version --json` emits machine-readable build metadata with `schemaVersion`.
Optional strict compatibility check is supported with `SUPRAGOFLOW_EXPECT_SCHEMA_VERSION`.

## VS Code Dev Container

Open the repo in VS Code and choose **Dev Containers: Reopen in Container**.
Inside the container:

```bash
./scripts/gg deps
./scripts/gg test
```

`./scripts/gg` detects it is already inside a container and will not try to run nested Docker.

## Local incremental container bring-up

1. Build the minimal build image first and validate it:
   ```bash
   docker build -f docker/Dockerfile.build -t supragoflow-build:local .
   docker run --rm -t supragoflow-build:local bash -lc "go version"
   ```

2. Build the dev image incrementally:
   ```bash
   docker build -f docker/Dockerfile.dev -t supragoflow-dev:local .
   docker run --rm -t supragoflow-dev:local bash -lc "golangci-lint version && govulncheck -h"
   ```

3. Use `./scripts/gg` for everything (humans, agents, CI).

## Contribution policy (invite-only behavior)

This repo is public, but Issues/PRs are **invite-only by behavior**:
- Non-allowed Issues/PRs are automatically commented on and closed.
- PRs from forks are rejected unless opened by a **designated fork contributor**.
- Agents are supported via a whitelist of GitHub usernames (bot identities).

See `CONTRIBUTING.md` and `.github/allowlist.yml`.

## Security

- Security contact and disclosure process: see [SECURITY.md](SECURITY.md).
- CI includes repository secret scanning and pinned GitHub Actions references.
- Container images default to non-root `USER`, and `./scripts/gg` runs containerized stages with invoker UID:GID mapping to preserve host ownership.
- Release workflow publishes SBOM assets for release images and binaries.

## Documentation Index

- [CONTRIBUTING.md](CONTRIBUTING.md) - Contribution guidelines.
- [POLICY.md](POLICY.md) - Project policies (tooling, security, releases).
- [docs/operations-runbook.md](docs/operations-runbook.md) - Operational runbook for humans and agents.
- [docs/securecomms.md](docs/securecomms.md) - Documentation for the secure communications package.
- [docs/work-log.md](docs/work-log.md) - Detailed work logs and integration status.
- [STATUS.md](STATUS.md) - High-level project status.
