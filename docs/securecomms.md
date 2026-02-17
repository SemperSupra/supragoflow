# Secure Communications Package

The `internal/securecomms` package provides helper functions to create secure TLS and SSH configurations with best practices and strict defaults.

## TLS Configuration

The package simplifies creating `tls.Config` objects for both clients and servers, enforcing TLS 1.2+ and proper certificate handling.
It enforces a strict list of AEAD-based cipher suites for TLS 1.2 to ensure strong security.

### Client Configuration

Use `NewTLSClientConfig` to create a `tls.Config` for a client. It supports mutual TLS (mTLS) if client certificates are provided.
It also allows controlling whether to trust the system's root CAs via the `trustSystemCAs` parameter.

```go
package main

import (
	"crypto/tls"
	"log"
	"os"

	"github.com/SemperSupra/supragoflow/internal/securecomms"
)

func main() {
	// Load CA certificate
	caPEM, err := os.ReadFile("ca.crt")
	if err != nil {
		log.Fatal(err)
	}

	// Load client certificate and key for mTLS (optional)
	clientCertPEM, _ := os.ReadFile("client.crt")
	clientKeyPEM, _ := os.ReadFile("client.key")

	// Create TLS config
	tlsConfig, err := securecomms.NewTLSClientConfig(
		caPEM,
		"server.example.com", // Expected ServerName
		clientCertPEM,        // Optional: nil or empty if not using mTLS
		clientKeyPEM,         // Optional: nil or empty if not using mTLS
		false,                // trustSystemCAs: false (only trust provided CA)
	)
	if err != nil {
		log.Fatalf("Failed to create TLS config: %v", err)
	}

	// Use tlsConfig in tls.Dial or http.Transport
	conn, err := tls.Dial("tcp", "server.example.com:443", tlsConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
}
```

### Server Configuration

Use `NewTLSServerConfig` to create a `tls.Config` for a server. It supports enforcing client authentication (mTLS).

```go
package main

import (
	"crypto/tls"
	"log"
	"os"

	"github.com/SemperSupra/supragoflow/internal/securecomms"
)

func main() {
	// Load server certificate and key
	serverCertPEM, err := os.ReadFile("server.crt")
	if err != nil {
		log.Fatal(err)
	}
	serverKeyPEM, err := os.ReadFile("server.key")
	if err != nil {
		log.Fatal(err)
	}

	// Load Client CA if requiring client certs (mTLS)
	clientCAPEM, _ := os.ReadFile("client_ca.crt")

	// Create TLS config
	tlsConfig, err := securecomms.NewTLSServerConfig(
		serverCertPEM,
		serverKeyPEM,
		clientCAPEM, // Optional: nil or empty if not requiring client certs
		true,        // requireClientCert: true to enforce mTLS
	)
	if err != nil {
		log.Fatalf("Failed to create TLS config: %v", err)
	}

	// Use tlsConfig in tls.Listen or http.Server
	ln, err := tls.Listen("tcp", ":8443", tlsConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
}
```

## SSH Configuration

The package provides `NewSSHClientConfig` to create a strict `ssh.ClientConfig` that validates host keys against a provided `known_hosts` data.
It enforces a strict set of Ciphers, Key Exchanges, and MACs, prioritizing AEAD and modern elliptic curves.

### Client Configuration

```go
package main

import (
	"log"
	"os"

	"golang.org/x/crypto/ssh"
	"github.com/SemperSupra/supragoflow/internal/securecomms"
)

func main() {
	// Load private key
	privateKeyPEM, err := os.ReadFile("id_rsa")
	if err != nil {
		log.Fatal(err)
	}

	// Load known_hosts data
	knownHostsData, err := os.ReadFile("known_hosts")
	if err != nil {
		log.Fatal(err)
	}

	// Create SSH client config
	sshConfig, err := securecomms.NewSSHClientConfig(
		"myuser",
		privateKeyPEM,
		knownHostsData,
	)
	if err != nil {
		log.Fatalf("Failed to create SSH config: %v", err)
	}

	// Connect to SSH server
	client, err := ssh.Dial("tcp", "example.com:22", sshConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
}
```
