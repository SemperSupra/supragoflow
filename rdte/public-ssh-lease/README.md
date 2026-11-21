# Public ephemeral SSH certificate conformance

**Authority:** private IaC #595 / draft #597, Agent Dispatch #748 for actual signed effect admissions, personal-ops #97 for credential lifecycle. This directory is a **new, public-safe source-only test fixture**, not a reusable credential service, an issuer/CA, an SSH proxy, an RPC server, or an alternative dispatcher.

The fixture opens exactly one ephemeral `127.0.0.1:0` TCP listener inside each disposable public GHA job. It generates three fresh Ed25519 private keys entirely in memory (synthetic CA, synthetic SSH user and SSH host). No key or token is read from a provider secret, written to disk, sent to a network host, emitted in an artifact, or reused between runs. The client pins the exact generated host public key. The synthetic SSH user certificate is CA-signed with a 10-minute lifetime and a JSON KeyId binding a fixed synthetic workset/source hash, venue, target, action, random nonce and generation. A server-side gate checks signer, validity, exact user principal, TTL ceiling, certificate serial and signed claim equivalence, before admitting any action.

Only `canary-health` is an allowed `exec` request. A mutex-protected state machine consumes its nonce exactly once at the action boundary, then rejects replay, other commands, shells and post-revocation requests (including over a connection already authenticated). Competing commands cannot consume the same lease more than once. This demonstrates scoped, revocable, one-shot **synthetic action permission**; the program has no generic command handler, no privileged shell, no external SSH host and no actual cloud IAM.

Standard public GitHub-hosted Ubuntu 24.04, Windows 2025 and macOS 15 independently run real loopback SSH handshakes and adversarial unit tests. Linux also runs Go race detector. No private GitHub Actions minutes, paid/larger runner or self-hosted runner is required.

## Non-goals and gaps

- This does not implement real GitHub OIDC verification, Cloudflare Access identity, Colab/Kaggle notebook authentication, SSH CA signing service, persistent anti-replay ledger, durable lease revocation, Headscale admission or OCI federation.
- The in-memory generation and revocation gate cannot survive restart. Before protected use, a durable authority with signed generation fencing must exist under existing private #595/Agent Dispatch #748.
- This is not an accepted cloud-to-sovereign route; `127.0.0.1` is the only reachable destination. Native #594 host fingerprint and restricted OpenSSH policy remain absent.
- The certificate is public-test-only and has no real operational identity. The synthetic workset hash is not a GitHub repository SHA fetched from private sources.

Run in an approved disposable Go environment:

```sh
cd rdte/public-ssh-lease
go test -race -count=1 ./...
go vet ./...
go run ./cmd/ssh-canary
```

A passing receipt has `verdict=PASS`, `scope=loopback-only`, `provider_effects=NONE`, `persistent_keys=false`. The public workflow needs to actually pass at its current exact head before claiming qualification. Keep the draft PR unmerged until native tests and review.
