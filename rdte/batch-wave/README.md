# Offline batch-wave validation for weak/local actors

**Purpose:** validate a bounded plan for independent work before humans, Qwen/edge agents, or existing Agent Dispatch runners execute anything. This is a **pure read-only compiler**: it cannot submit jobs, create infrastructure, read GitHub issues, provide identity/authorization, or act as a scheduler. All run admission, issue ownership, branch writes, verification, and reconciliation stay in the owning GitHub DLE. The private portfolio authority is infrastructure issue #603; this public directory contains only fresh synthetic examples and source.

`batch-wave/v1` requires each card to specify a distinct branch, exact 40-hex source commit, unique workset/actor, venue, confidentiality class, effect class, at most six small concrete steps, max thirty-minute scope, explicit negative oracle and rollback, durable receipt target, read paths, write paths, resource reservations, and actual dependencies. Paths remain Git-relative. Shared branch, overlapping file/read/write or physical resources are rejected unless one card depends on the other. Dependency cycles and ambiguous JSON/unknown fields fail closed.

For smaller models: a **single bounded baton** is a complete input. Give the actor only that card, the named issue, the exact source ref and required local AGENTS.md. The actor must not infer scope expansion from a chat history or change another workcell. After each unit of work, return the exact observed result and source SHA to the designated issue. The stronger planner can independently verify the receipt and rerun the compiler for the next wave; blocked cards do not block independent siblings.

Inside an approved isolated Go workcell:

```sh
cd rdte/batch-wave
GOTOOLCHAIN=local go test -race -count=1 ./...
go run ./cmd/batch-wave plan examples/public-wave.json
go run ./cmd/batch-wave baton examples/public-wave.json b-validate
```

Public workflow .github/workflows/public-batch-wave.yml independently exercises Linux, Windows and macOS hosted runners and builds Linux amd64/arm64, Windows amd64, and macOS arm64 **test utilities**. It does not execute the described tasks. The private portfolio manifest is never passed to public CI.

**Trust caveats:** an effect field or admission string is not a real signed Agent Dispatch grant; the parser only checks shape and rejects obvious dangerous placement, not source/secret classification itself. Public safety requires human/code-review publication screening. No live target settings, OCI/GHA/Colab/Kaggle login, SSH/Headscale or secret values belong in a public manifest. Validating a manifest is not permission to execute an effect.

**Acceptance:** current private source and issue/PR claims remain the only work authority. No automatic merge or local installer release is implied by public green tests.
