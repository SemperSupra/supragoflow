# SupraGoFlow

SupraGoFlow is a git-native, containerized lifecycle for lightweight Go projects (CLI tools, services, bots, telemetry agents).
It provides a **single lifecycle entrypoint** (`./scripts/gg`) that works for:

- Humans on the command line
- CLI agents
- CI/CD pipelines
- VS Code Dev Containers

Targets: **Linux + Windows** (cross-compile). No GUI and no heavy CGO.

## Canonical releases (GHCR)

**Prefer GHCR-published images tagged to GitHub Releases as the canonical toolchain.**
Local builds (`*:local`) are for development only and are not authoritative.

Images published on release:

- `ghcr.io/<org>/supragoflow-build:<tag>`
- `ghcr.io/<org>/supragoflow-dev:<tag>`

> Containers are built and pushed **only on GitHub Releases**.

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
```

Outputs land in `dist/<goos>-<goarch>/`.

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
