# STATUS

Last updated: 2026-02-13

## Current State

- Branch: `main`
- Git status: clean (`main...origin/main`)
- Latest commit: `60bd1e6` (`Add securecomms package for TLS and SSH channel setup`)

## Completed Work

- Containerized Go toolchain/build flow fixed and stabilized.
- `./scripts/gg` lifecycle working for local dev and CI usage.
- Windows Wine smoke job added in CI.
- Security/tooling policy updates added (proven implementations + Windows/Wine compatibility requirement).
- New secure comms package added:
  - `internal/securecomms/tls.go`
  - `internal/securecomms/ssh.go`
  - tests for both TLS and SSH config builders

## Validation Snapshot

Verified locally (containerized):

- `./scripts/gg images` (build + dev images rebuild)
- `./scripts/gg fmt`
- `./scripts/gg vet`
- `./scripts/gg lint`
- `./scripts/gg vuln`
- `./scripts/gg test`
- `./scripts/gg build linux amd64`
- `./scripts/gg build windows amd64`
- `./scripts/gg package`

All passed in the SupraGoFlow environment.

## WineBot Integration Status

Result:
- `supragoflow.exe` runs correctly under Wine when executed directly as `winebot` user in WineBot container.
- WineBot launcher path (`scripts/run-app.sh` headless attached flow) is the problem area for CLI-style EXEs.

Artifacts and issue filed:
- WineBot issue: `https://github.com/mark-e-deyoung/WineBot/issues/2`
- Repro + logs + patch draft gist: `https://gist.github.com/mark-e-deyoung/c96406c7cfc7ba6c4d99eebe64e51048`
- Local artifact bundle: `/tmp/winebot-supragoflow-issue`

## Known External Dependency / Blocker

- Pending WineBot-side fix for launcher behavior:
  - direct CLI execution mode
  - optional explorer supervision disable for CLI workloads

SupraGoFlow core is not blocked for normal `gg` workflow, but end-to-end WineBot launcher compatibility is blocked pending that WineBot fix.

## Resume Plan (after WineBot check)

1. Pull/check WineBot issue updates and merged fix.
2. Re-run WineBot validation for SupraGoFlow EXE:
   - `./scripts/run-app.sh /apps/supragoflow.exe --mode headless --args "--version --json"`
3. If fixed, capture successful output and update this file.
4. Optionally add/strengthen SupraGoFlow CI path that validates WineBot launcher compatibility (not just bare Wine).

## Quick Resume Commands

From SupraGoFlow repo:

```bash
cd /home/mark/Projects/SupraGoFlow/workspace/supragoflow
git pull --ff-only
./scripts/gg test
./scripts/gg build windows amd64
```

From WineBot repo (current repro path):

```bash
cd /home/mark/Projects/WineBot
./scripts/run-app.sh /apps/supragoflow.exe --mode headless --args "--version --json"
```
