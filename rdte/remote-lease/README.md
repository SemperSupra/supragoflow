# Public synthetic lease conformance (no provider access)

Source-only public-safe RDTE for the lifecycle contract in private infrastructure issue #595. These tests contain no secret material, host identifiers, private source, API operations, SSH connections, provider identity verification or production admission. This **does not implement a credential issuer**.

Bounded synthetic paths: `REQUESTED → ACTIVE → FENCED/EXPIRED/CLOSED`, denial of replay, unsupported venue/method, stale generation, late renewal and concurrent duplicate issuance. Real OIDC claims and native Colab/Kaggle operator authentication, provider reconciliation, crash durability, signed public SSH certificates, credential revocation and host isolation remain **NOT TESTED**. In particular the in-memory nonce and fence state is intentionally NOT restart-safe and MUST NOT be deployed.

The private owner must separately qualify all protected provider effects and never import this toy store into production. CI triggers only on public repo events and has read-only permissions. No fork secret grants.
