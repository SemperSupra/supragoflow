# Agent Prompt: Implement SupraGoFlow (SemperSupra)

You are an implementation agent working in a GitHub repo under the SemperSupra organization. Implement **SupraGoFlow**:
a git-native, containerized lifecycle for lightweight Go projects (CLI tool, service, bot, telemetry agent) supporting
humans, CLI agents, CI/CD, and VS Code devcontainers. Targets **Linux + Windows** (no mac). No GUIs, no heavy CGO.

**Canonical releases:** Prefer GHCR images tagged to GitHub Releases. Local builds are for development only.

**Container policy:** Build/push containers **only on GitHub Release**. No publishing on PR/branch builds.

**Contribution policy:** Repo is public, but Issues/PRs are invite-only by automation.
Fork PRs are never accepted unless the author is a designated fork contributor. The standard allowlist does not grant fork
privileges. Agents must be allowlisted by GitHub username (bot identities).

Follow the project skeleton and acceptance criteria in this repository.
