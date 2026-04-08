# Dependency Analysis & Strategy

## Executive Summary

The project's application dependencies are largely up-to-date with minimal actionable updates required for stability. The primary recommendation is to improve build reproducibility by pinning development tools in the `Dockerfile.dev`.

## Application Dependencies (`go.mod`)

| Dependency | Current Version | Recommended Action | Tradeoffs/Implications |
| :--- | :--- | :--- | :--- |
| `golang.org/x/crypto` | `v0.48.0` | **Keep as is** | Latest version compatible with Go 1.25. No updates available. |
| `golang.org/x/net` | `v0.49.0` (indirect) | **Keep as is** | Version `v0.50.0` is available, but the project does not directly depend on `x/net`. The dependency is transitively pulled by `x/crypto`, which pins it to `v0.49.0`. Forcing an update via `go get` is reverted by `go mod tidy` unless explicitly imported. Staying on `v0.49.0` maintains standard dependency management. |
| `golang.org/x/sys` | `v0.41.0` (indirect) | **Keep as is** | Latest version. No updates available. |
| `golang.org/x/term` | `v0.40.0` (indirect) | **Keep as is** | Latest version. No updates available. |
| `golang.org/x/text` | `v0.34.0` (indirect) | **Keep as is** | Latest version. No updates available. |

**Implication:** The application dependency graph is healthy and secure. No manual intervention is required for `go.mod`.

## Tooling Dependencies (`Dockerfile.dev`)

| Tool | Previous State | Recommended Action | Tradeoffs/Implications |
| :--- | :--- | :--- | :--- |
| `govulncheck` | `@latest` | **Pinned to `v1.1.4`** | **Tradeoff:** Requires manual updates in the future. **Implication:** Ensures reproducible builds and prevents CI breakage if a new version introduces breaking changes or bugs. |
| `gotestsum` | `@latest` | **Pinned to `v1.13.0`** | **Tradeoff:** Requires manual updates. **Implication:** Ensures consistent test output formatting and behavior across environments. |
| `golangci-lint` | `v2.9.0` | **Monitor** | Current version is pinned in `.versions`. Monitor for future releases to enable new linters. |

## Upgrade Strategy

1.  **Pin Dependencies:** Move away from `@latest` for critical build and test tools to ensure deterministic environments.
2.  **Automated Checks:** Continue using `dependabot` or similar tools (if available) to propose updates for pinned versions.
3.  **Review Process:** evaluate dependency updates quarterly or when security vulnerabilities are disclosed.
4.  **Testing:** Run the full test suite (`./scripts/gg test`) after any dependency update to verify compatibility.

## Alternatives Considered

*   **Force Update `x/net`:** We considered forcing `golang.org/x/net` to `v0.50.0` by adding a direct dependency.
    *   *Pros:* Access to latest potential fixes.
    *   *Cons:* Adds unnecessary direct dependency to `go.mod`, creating noise. `go mod tidy` fights this unless a dummy import is added.
    *   *Decision:* Rejected. Stick to standard `go mod tidy` behavior.
