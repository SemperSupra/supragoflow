# Operations Runbook

This runbook is the fast path for humans and agents operating SupraGoFlow.

## Repository Root

The Git repository root is:

`/home/mark/Projects/SupraGoFlow/workspace`

Run all `git` and `./scripts/gg` commands from `workspace/`.
Keep local scratch artifacts in the parent project directory so they are not committed.

## Standard Lifecycle

Use only `./scripts/gg` stages:

```bash
./scripts/gg bootstrap
./scripts/gg deps
./scripts/gg fmt
./scripts/gg vet
./scripts/gg lint
./scripts/gg self-test
./scripts/gg test
./scripts/gg build linux amd64
./scripts/gg package
```

## Contract and Compatibility Checks

- Print lifecycle contract: `./scripts/gg contract`
- Validate required config: `./scripts/gg validate-config`
- Validate writable roots: `./scripts/gg verify-writable`

Fail-fast compatibility env vars:

- `SUPRAGOFLOW_EXPECT_SCHEMA_VERSION`
- `SUPRAGOFLOW_EXPECT_LOG_SCHEMA_VERSION`
- `SUPRAGOFLOW_EXPECT_GG_INTERFACE_VERSION`

## Diagnostics and Failure Triage

Collect bounded diagnostics bundle:

```bash
./scripts/gg diagnose
```

Diagnostics are written under `dist/diagnostics/` (with fallback to `/tmp` when needed), and include redacted environment output.

Recommended first triage sequence:

```bash
./scripts/gg --log-level debug bootstrap
./scripts/gg validate-config
./scripts/gg verify-writable
./scripts/gg diagnose
```

## Cache and Permissions

Cache mode:

- `SUPRAGOFLOW_CACHE_STRATEGY=volume` (default)
- `SUPRAGOFLOW_CACHE_STRATEGY=host` (CI-friendly)

Prune host cache when bounded max is exceeded:

```bash
./scripts/gg cache-prune
```

If containerized stages fail with permission errors, run:

```bash
./scripts/gg verify-writable
```

Then ensure the current user owns/writes `dist/` and `.cache/` roots.

## CI/Policy Conformance

Run policy gate locally before PR:

```bash
./scripts/check-policy-conformance.sh
```

This checks required policy docs, pinned GitHub Action SHAs, and no `:latest` workflow tags.
