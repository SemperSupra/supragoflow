# Public synthetic stateless-workcell RDTE

This directory is an **entirely public-safe, zero-credential, synthetic** conformance component for the existing SupraGoFlow cross-platform Go lifecycle. Nothing was copied from private IaC or sovereign-host files. It contains no IP addresses, endpoint names, host identities, SSH keys, OAuth clients, provider tokens, environment variables, privileged operations, network probes or deployable remote-access configuration.

## Scope of the public runner qualification

- Go standard-library deterministic observe/plan/apply/verify/rollback/destroy of a **single synthetic marker payload** in a newly created child of the OS temporary directory
- Strict synthetic-only ownership marker, exact plan hash, refusal to adopt preexisting paths, symlink/file escape rejection, staged atomic file publication and journal
- Negative tests for wrong plan/digest, stale state, before/after-write interruption, manual edits, unknown files, wrong roots, duplicate apply and rollback
- Native read-only Windows PowerShell 5.1 check of Hyper-V module / `vmms` / VM and switch inventory permissions and **presence only** of optional tools; no hostnames, IPs, credentials, account names or sensitive file paths logged
- Linux native Go race+vet, Windows native Go and PS5.1, macOS native Go, and SHA256-checksummed cross-binaries for linux/amd64, linux/arm64, windows/amd64 and darwin/arm64
- Public unprivileged GitHub-hosted runners (no personal secrets, OIDC, self-hosted runner, elevated runner identity or privileged nested hypervisor), disposable by provider design; read-only GitHub content permission; 15-minute job timeouts

## Run locally without installing build tools

For read-only Windows capability discovery, use the reviewed PowerShell 5.1 source file that ships in the PR:

```powershell
# from a clean checkout of this public repository
& .\rdte\public-workcell\scripts\Invoke-PublicWorkcellPreflight.ps1 -Stage Observe
```

If PowerShell execution policy disallows it, report BLOCKED; do not change workstation policy just to run the census.

The downloadable GitHub Actions multiarch archive (when the public workflow has actually succeeded) contains a **synthetic** test binary with no ability to configure real SSH, Headscale, Cloudflare, OCI, or Hyper-V. To exercise it:

```powershell
# after verifying the artifact checksum and approved source revision
& .\workcell-rdte-windows-amd64.exe exercise
```

Successful output must include `mode: SYNTHETIC_FILESYSTEM_ONLY`, `destroy_verified: true`, and `provider_effects: NONE`. The only filesystem effect is in a unique, owned OS-temp directory, destroyed after rollback.

For developers in a **separately admitted disposable Go workcell only**:

```sh
cd rdte/public-workcell
GOTOOLCHAIN=local go test -count=1 -race ./...
GOTOOLCHAIN=local go vet ./...
go run ./cmd/workcell-rdte exercise
```

## What this DOES NOT qualify

Public GitHub runners cannot prove the owner's physical Windows Hyper-V delegation, TrueNAS, BigMac, OCI VPS, live SSH/Cloudflare/Headscale settings, native notebook login/egress, work browser authentication, issuer signing keys, P2P/direct-only routing, serial guest login, or an actual service deployment on persistent hosts. The synthetic exact-plan hash is **not an owner identity signature**. The temp marker is **not a secure hypervisor admission**. No live issuer, credentials, host configuration or static deployment image is built by this project.

The actual local installer/workcell deployment remains gated on authorized native host census, qualified trusted guest image, Hyper-V VM management permissions, protected remote credential issuer, current workset effect admission, and independent positive/negative/rollback tests. Never interpret this project's green checks as local deployment-ready status.

## Release and provenance

The public workflow writes `dist/SHA256SUMS` using deterministic cross-build flags, then publishes the binaries as a short-retention PR artifact. This is **not** a trusted signed release until independent review and release approval. Do not use a synthetic artifact as a remote-access installer.

Runtimes are meant to be replaceable. Protected durable state is always external and is not simulated as safely disposable simply because a runner VM is transient. 

## Optional native Hyper-V canary on a public Windows runner

The isolated workflow job `windows-hyperv-native-experiment` uses a **separate
short-lived standard hosted Windows runner** to attempt creating a synthetic
1 GiB dynamic parent VHDX, differencing child, and an unbooted Generation 2
Hyper-V VM **without a vSwitch or network**, then verify native API ownership
and remove the exact created VM/disk directory in a finally block. The test
strictly requires public GITHUB_ACTIONS + RUNNER_TEMP and refuses workstation
execution. If nested Hyper-V is unsupported, the typed result is
`BLOCKED_HYPERV_CAPABILITY`, not PASS. Any unverified cleanup fails the job.
GitHub documentation does not officially support nested virtualization on
standard hosted runners; native proof here cannot replace owner-host HIL.

A successful `PASS_DISK_AND_VM_CONTROL_ONLY` still does **not** prove boot,
cloud-init, guest service deployment, privileged remote reachability,
Windows owner permissions, or network isolation for a running guest.
