package securecomms

import (
	"errors"
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// secureSSHCiphers lists preferred AEAD ciphers and strong CTR modes.
// It avoids CBC modes which are vulnerable to padding oracle attacks.
var secureSSHCiphers = []string{
	"chacha20-poly1305@openssh.com",
	"aes128-gcm@openssh.com",
	"aes256-gcm@openssh.com",
	"aes128-ctr",
	"aes192-ctr",
	"aes256-ctr",
}

// secureSSHKeyExchanges lists preferred key exchange algorithms, prioritizing
// modern elliptic curves (Curve25519) and strong NIST curves.
var secureSSHKeyExchanges = []string{
	"curve25519-sha256",
	"curve25519-sha256@libssh.org",
	"ecdh-sha2-nistp256",
	"ecdh-sha2-nistp384",
	"ecdh-sha2-nistp521",
	"diffie-hellman-group14-sha256",
}

// secureSSHMACs lists preferred message authentication code algorithms,
// prioritizing Encrypt-then-MAC (EtM) modes and SHA-2.
var secureSSHMACs = []string{
	"hmac-sha2-256-etm@openssh.com",
	"hmac-sha2-512-etm@openssh.com",
	"hmac-sha2-256",
	"hmac-sha2-512",
}

// NewSSHClientConfig builds a strict SSH client config using known_hosts validation.
func NewSSHClientConfig(user string, privateKeyPEM []byte, knownHostsData []byte) (*ssh.ClientConfig, error) {
	return NewSSHClientConfigWithTimeout(user, privateKeyPEM, knownHostsData, 10*time.Second)
}

// NewSSHClientConfigWithTimeout builds a strict SSH client config using
// known_hosts validation and a caller-controlled connection timeout.
func NewSSHClientConfigWithTimeout(user string, privateKeyPEM []byte, knownHostsData []byte, timeout time.Duration) (*ssh.ClientConfig, error) {
	if user == "" {
		return nil, errors.New("user is required")
	}
	if len(privateKeyPEM) == 0 {
		return nil, errors.New("private key is required")
	}
	if len(knownHostsData) == 0 {
		return nil, errors.New("known_hosts data is required")
	}
	if timeout <= 0 {
		return nil, errors.New("timeout must be > 0")
	}

	signer, err := ssh.ParsePrivateKey(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// We use a temporary file for known_hosts because the proven knownhosts package
	// from golang.org/x/crypto/ssh/knownhosts only accepts file paths, not io.Reader.
	// While suboptimal, using the battle-hardened parser is safer than implementing
	// a custom one. We ensure the file is cleaned up.
	tmpFile, err := os.CreateTemp("", "known_hosts_*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp known_hosts file: %w", err)
	}
	tmpPath := tmpFile.Name()
	if _, err := tmpFile.Write(knownHostsData); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
		return nil, fmt.Errorf("failed to write known_hosts data: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return nil, fmt.Errorf("failed to close temp known_hosts file: %w", err)
	}
	defer func() {
		_ = os.Remove(tmpPath)
	}()

	hostKeyCallback, err := knownhosts.New(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse known_hosts data: %w", err)
	}

	return &ssh.ClientConfig{
		Config: ssh.Config{
			Ciphers:      secureSSHCiphers,
			KeyExchanges: secureSSHKeyExchanges,
			MACs:         secureSSHMACs,
		},
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: hostKeyCallback,
		Timeout:         timeout,
	}, nil
}
