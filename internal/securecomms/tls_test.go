package securecomms

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"
)

func TestNewTLSClientConfig(t *testing.T) {
	caPEM, _, _, err := generateCA(t)
	if err != nil {
		t.Fatalf("generateCA: %v", err)
	}

	cfg, err := NewTLSClientConfig(caPEM, "example.com", nil, nil)
	if err != nil {
		t.Fatalf("NewTLSClientConfig returned error: %v", err)
	}
	if cfg.MinVersion != tls.VersionTLS12 {
		t.Fatalf("unexpected MinVersion: %d", cfg.MinVersion)
	}
	if cfg.ServerName != "example.com" {
		t.Fatalf("unexpected ServerName: %q", cfg.ServerName)
	}
	if cfg.RootCAs == nil {
		t.Fatal("expected RootCAs to be set")
	}
}

func TestNewTLSClientConfigInvalidInput(t *testing.T) {
	if _, err := NewTLSClientConfig([]byte("bad"), "example.com", nil, nil); err == nil {
		t.Fatal("expected error for invalid CA PEM")
	}
	if _, err := NewTLSClientConfig(nil, "", nil, nil); err == nil {
		t.Fatal("expected error for empty serverName")
	}
}

func TestNewTLSServerConfigMutualTLS(t *testing.T) {
	caPEM, caCert, caKey, err := generateCA(t)
	if err != nil {
		t.Fatalf("generateCA: %v", err)
	}
	serverCertPEM, serverKeyPEM, err := generateLeafCert(t, caCert, caKey, "server.example")
	if err != nil {
		t.Fatalf("generateLeafCert: %v", err)
	}

	cfg, err := NewTLSServerConfig(serverCertPEM, serverKeyPEM, caPEM, true)
	if err != nil {
		t.Fatalf("NewTLSServerConfig returned error: %v", err)
	}
	if cfg.ClientAuth != tls.RequireAndVerifyClientCert {
		t.Fatalf("unexpected ClientAuth: %v", cfg.ClientAuth)
	}
	if cfg.ClientCAs == nil {
		t.Fatal("expected ClientCAs to be set")
	}
}

func TestNewTLSServerConfigInvalidInput(t *testing.T) {
	if _, err := NewTLSServerConfig(nil, nil, nil, false); err == nil {
		t.Fatal("expected error for missing server cert/key")
	}
	if _, err := NewTLSServerConfig([]byte("bad"), []byte("bad"), nil, false); err == nil {
		t.Fatal("expected parse error for invalid server cert/key")
	}
}

func generateCA(t *testing.T) ([]byte, *x509.Certificate, *rsa.PrivateKey, error) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, nil, err
	}
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "test-ca"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            1,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return nil, nil, nil, err
	}
	caCert, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, nil, nil, err
	}
	caPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	return caPEM, caCert, key, nil
}

func generateLeafCert(t *testing.T, caCert *x509.Certificate, caKey *rsa.PrivateKey, cn string) ([]byte, []byte, error) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: cn},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		DNSNames:     []string{cn},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, caCert, &key.PublicKey, caKey)
	if err != nil {
		return nil, nil, err
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	return certPEM, keyPEM, nil
}
