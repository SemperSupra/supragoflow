package securecomms

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

func TestNewSSHClientConfig(t *testing.T) {
	clientKeyPEM, err := generateRSAPrivateKeyPEM()
	if err != nil {
		t.Fatalf("generateRSAPrivateKeyPEM: %v", err)
	}

	hostKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey: %v", err)
	}
	hostSigner, err := ssh.NewSignerFromKey(hostKey)
	if err != nil {
		t.Fatalf("ssh.NewSignerFromKey: %v", err)
	}

	knownHostsLine := knownhosts.Line([]string{"example.com"}, hostSigner.PublicKey())

	cfg, err := NewSSHClientConfig("alice", clientKeyPEM, []byte(knownHostsLine+"\n"))
	if err != nil {
		t.Fatalf("NewSSHClientConfig returned error: %v", err)
	}
	if cfg.User != "alice" {
		t.Fatalf("unexpected user: %q", cfg.User)
	}
	if len(cfg.Auth) == 0 {
		t.Fatal("expected auth methods")
	}

	// Verify strict crypto settings
	if len(cfg.Ciphers) == 0 {
		t.Fatal("expected Ciphers to be set")
	}
	for _, c := range cfg.Ciphers {
		found := false
		for _, allowed := range secureSSHCiphers {
			if c == allowed {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Cipher %q is not in the allowed list", c)
		}
	}

	if len(cfg.KeyExchanges) == 0 {
		t.Fatal("expected KeyExchanges to be set")
	}
	for _, k := range cfg.KeyExchanges {
		found := false
		for _, allowed := range secureSSHKeyExchanges {
			if k == allowed {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("KeyExchange %q is not in the allowed list", k)
		}
	}

	if len(cfg.MACs) == 0 {
		t.Fatal("expected MACs to be set")
	}
	for _, m := range cfg.MACs {
		found := false
		for _, allowed := range secureSSHMACs {
			if m == allowed {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("MAC %q is not in the allowed list", m)
		}
	}

	// Verify HostKeyCallback works
	if err := cfg.HostKeyCallback("example.com:22", &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 22}, hostSigner.PublicKey()); err != nil {
		t.Fatalf("expected host key callback success, got: %v", err)
	}
}

func TestNewSSHClientConfigWithTimeout(t *testing.T) {
	clientKeyPEM, err := generateRSAPrivateKeyPEM()
	if err != nil {
		t.Fatalf("generateRSAPrivateKeyPEM: %v", err)
	}
	hostKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey: %v", err)
	}
	hostSigner, err := ssh.NewSignerFromKey(hostKey)
	if err != nil {
		t.Fatalf("ssh.NewSignerFromKey: %v", err)
	}
	knownHostsLine := knownhosts.Line([]string{"example.com"}, hostSigner.PublicKey())

	cfg, err := NewSSHClientConfigWithTimeout("alice", clientKeyPEM, []byte(knownHostsLine+"\n"), 3*time.Second)
	if err != nil {
		t.Fatalf("NewSSHClientConfigWithTimeout returned error: %v", err)
	}
	if cfg.Timeout != 3*time.Second {
		t.Fatalf("unexpected timeout: %v", cfg.Timeout)
	}
	if _, err := NewSSHClientConfigWithTimeout("alice", clientKeyPEM, []byte(knownHostsLine+"\n"), 0); err == nil {
		t.Fatal("expected error for non-positive timeout")
	}
}

func TestNewSSHClientConfigInvalidInput(t *testing.T) {
	if _, err := NewSSHClientConfig("", nil, nil); err == nil {
		t.Fatal("expected error for missing user/key/known_hosts")
	}
	keyPEM, err := generateRSAPrivateKeyPEM()
	if err != nil {
		t.Fatalf("generateRSAPrivateKeyPEM: %v", err)
	}

	// Verify that invalid known_hosts data (e.g. junk string) causes a parser error.
	if _, err := NewSSHClientConfig("alice", keyPEM, []byte("not-known-hosts")); err == nil {
		t.Fatal("expected error for invalid known_hosts data")
	}
}

func TestSSHHostKeyCallbackRejectsMismatchedHostKey(t *testing.T) {
	clientKeyPEM, err := generateRSAPrivateKeyPEM()
	if err != nil {
		t.Fatalf("generateRSAPrivateKeyPEM: %v", err)
	}
	knownHostKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey known host: %v", err)
	}
	knownSigner, err := ssh.NewSignerFromKey(knownHostKey)
	if err != nil {
		t.Fatalf("ssh.NewSignerFromKey known host: %v", err)
	}
	knownHostsLine := knownhosts.Line([]string{"example.com"}, knownSigner.PublicKey())

	cfg, err := NewSSHClientConfig("alice", clientKeyPEM, []byte(knownHostsLine+"\n"))
	if err != nil {
		t.Fatalf("NewSSHClientConfig: %v", err)
	}

	otherHostKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey other host: %v", err)
	}
	otherSigner, err := ssh.NewSignerFromKey(otherHostKey)
	if err != nil {
		t.Fatalf("ssh.NewSignerFromKey other host: %v", err)
	}

	err = cfg.HostKeyCallback("example.com:22", &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 22}, otherSigner.PublicKey())
	if err == nil {
		t.Fatal("expected host key callback failure for mismatched host key")
	}
}

func generateRSAPrivateKeyPEM() ([]byte, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}), nil
}
