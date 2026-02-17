# Cryptography Guidelines

This project prioritizes the use of proven, battle-hardened, and open-source cryptographic implementations. We avoid implementing custom cryptographic algorithms ("rolling your own crypto") and rely on the Go standard library and the extended `golang.org/x/crypto` packages.

## Recommended Libraries

For all cryptographic operations, use the following libraries:

### Transport Layer Security (TLS)

- **Library**: `crypto/tls` (Standard Library)
- **Usage**: Secure communication between services and clients.
- **Configuration**:
  - Enforce `MinVersion: tls.VersionTLS12` (TLS 1.2 or higher).
  - Use restricted `CipherSuites` prioritizing AEAD ciphers (GCM, ChaCha20-Poly1305).
  - Avoid CBC mode ciphers.
  - Use strict certificate verification.

### Secure Shell (SSH)

- **Library**: `golang.org/x/crypto/ssh`
- **Usage**: Secure remote access and command execution.
- **Configuration**:
  - Use strict `HostKeyCallback` (e.g., `knownhosts` package).
  - Restrict `KeyExchanges` to modern curves (Curve25519, NIST P-256/384/521).
  - Restrict `Ciphers` to AEAD modes (ChaCha20-Poly1305, AES-GCM).
  - Restrict `MACs` to HMAC-SHA2 (SHA-256/512) or EtM modes.

### Hashing

- **Library**: `crypto/sha256` or `crypto/sha512` (Standard Library)
- **Usage**: Data integrity, checksums, identifiers.
- **Note**: Do not use MD5 or SHA-1 for security purposes.

### Password Hashing

- **Library**: `golang.org/x/crypto/argon2`
- **Usage**: Storing user passwords.
- **Configuration**: Use recommended parameters for memory and time cost (e.g., Argon2id).

### Symmetric Encryption

- **Library**: `crypto/aes` with `crypto/cipher` (GCM mode) or `golang.org/x/crypto/chacha20poly1305`
- **Usage**: Encrypting data at rest or specific fields.
- **Note**: Always use authenticated encryption (AEAD).

### Randomness

- **Library**: `crypto/rand` (Standard Library)
- **Usage**: Generating keys, nonces, salts, and tokens.
- **Note**: Do not use `math/rand` for security-sensitive values.

## Key Management

- Keys should never be hardcoded in the source code.
- Use environment variables or a dedicated secrets manager (e.g., HashiCorp Vault, AWS Secrets Manager) to inject keys at runtime.
- In `internal/securecomms`, keys are accepted as PEM-encoded byte slices, allowing flexibility in how they are loaded.

## Implementation Details

The `internal/securecomms` package provides helper functions that adhere to these guidelines, enforcing strict defaults for TLS and SSH configurations.
