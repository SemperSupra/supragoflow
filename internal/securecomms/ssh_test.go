package securecomms

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net"
	"testing"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

func TestNewSSHClientConfig(t *testing.T) {
	clientKeyPEM, err := generateRSAPrivateKeyPEM()
	if err != nil {
		t.Fatalf("generateRSAPrivateKeyPEM: %v", err)
	}
	hostSigner, err := generateSSHSigner()
	if err != nil {
		t.Fatalf("generateSSHSigner: %v", err)
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
	if err := cfg.HostKeyCallback("example.com:22", &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 22}, hostSigner.PublicKey()); err != nil {
		t.Fatalf("expected host key callback success, got: %v", err)
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
	if _, err := NewSSHClientConfig("alice", keyPEM, []byte("not-known-hosts")); err == nil {
		t.Fatal("expected error for invalid known_hosts data")
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

func generateSSHSigner() (ssh.Signer, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	return ssh.NewSignerFromKey(key)
}
