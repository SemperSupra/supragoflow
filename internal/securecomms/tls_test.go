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

	// Test with trustSystemCAs = false
	cfg, err := NewTLSClientConfig(caPEM, "example.com", nil, nil, false)
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
	if len(cfg.CipherSuites) == 0 {
		t.Fatal("expected CipherSuites to be set")
	}
	// Verify strict cipher suites
	for _, cs := range cfg.CipherSuites {
		found := false
		for _, allowed := range secureCipherSuites {
			if cs == allowed {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("CipherSuite %d is not in the allowed list", cs)
		}
	}

	// Test with trustSystemCAs = true (should not error, but harder to verify impact without actual system roots)
	cfg2, err := NewTLSClientConfig(caPEM, "example.com", nil, nil, true)
	if err != nil {
		t.Fatalf("NewTLSClientConfig(trustSystemCAs=true) returned error: %v", err)
	}
	if cfg2.RootCAs == nil {
		t.Fatal("expected RootCAs to be set")
	}
}

func TestNewTLSClientConfigInvalidInput(t *testing.T) {
	if _, err := NewTLSClientConfig([]byte("bad"), "example.com", nil, nil, false); err == nil {
		t.Fatal("expected error for invalid CA PEM")
	}
	if _, err := NewTLSClientConfig(nil, "", nil, nil, false); err == nil {
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
	if len(cfg.CipherSuites) == 0 {
		t.Fatal("expected CipherSuites to be set")
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

func TestTLSHandshakeBehavior(t *testing.T) {
	caPEM, caCert, caKey, err := generateCA(t)
	if err != nil {
		t.Fatalf("generateCA: %v", err)
	}

	serverCertPEM, serverKeyPEM, err := generateLeafCert(t, caCert, caKey, "server.example")
	if err != nil {
		t.Fatalf("generateLeafCert: %v", err)
	}
	serverCfg, err := NewTLSServerConfig(serverCertPEM, serverKeyPEM, nil, false)
	if err != nil {
		t.Fatalf("NewTLSServerConfig: %v", err)
	}

	ln, err := tls.Listen("tcp", "127.0.0.1:0", serverCfg)
	if err != nil {
		t.Fatalf("tls.Listen: %v", err)
	}
	defer func() { _ = ln.Close() }()

	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		tlsConn := conn.(*tls.Conn)
		_ = tlsConn.SetDeadline(time.Now().Add(3 * time.Second))
		_ = tlsConn.Handshake()
		_, _ = tlsConn.Write([]byte("ok"))
		_ = conn.Close()
	}()

	clientCfg, err := NewTLSClientConfig(caPEM, "server.example", nil, nil, false)
	if err != nil {
		t.Fatalf("NewTLSClientConfig: %v", err)
	}
	clientConn, err := tls.Dial("tcp", ln.Addr().String(), clientCfg)
	if err != nil {
		t.Fatalf("tls.Dial should succeed: %v", err)
	}
	_ = clientConn.Close()
	<-done

	ln2, err := tls.Listen("tcp", "127.0.0.1:0", serverCfg)
	if err != nil {
		t.Fatalf("tls.Listen second: %v", err)
	}
	defer func() { _ = ln2.Close() }()
	go func() {
		conn, err := ln2.Accept()
		if err != nil {
			return
		}
		tlsConn := conn.(*tls.Conn)
		_ = tlsConn.SetDeadline(time.Now().Add(3 * time.Second))
		_ = tlsConn.Handshake()
		_ = tlsConn.Close()
	}()

	badClientCfg, err := NewTLSClientConfig(caPEM, "wrong.example", nil, nil, false)
	if err != nil {
		t.Fatalf("NewTLSClientConfig bad host: %v", err)
	}
	badConn, err := tls.Dial("tcp", ln2.Addr().String(), badClientCfg)
	if err == nil {
		_ = badConn.Close()
		t.Fatal("expected tls.Dial to fail for wrong ServerName")
	}
}

func TestTLSMutualTLSHandshakeBehavior(t *testing.T) {
	caPEM, caCert, caKey, err := generateCA(t)
	if err != nil {
		t.Fatalf("generateCA: %v", err)
	}
	serverCertPEM, serverKeyPEM, err := generateLeafCert(t, caCert, caKey, "mtls.example")
	if err != nil {
		t.Fatalf("generateLeafCert server: %v", err)
	}
	clientCertPEM, clientKeyPEM, err := generateLeafCert(t, caCert, caKey, "client.example")
	if err != nil {
		t.Fatalf("generateLeafCert client: %v", err)
	}

	serverCfg, err := NewTLSServerConfig(serverCertPEM, serverKeyPEM, caPEM, true)
	if err != nil {
		t.Fatalf("NewTLSServerConfig: %v", err)
	}
	ln, err := tls.Listen("tcp", "127.0.0.1:0", serverCfg)
	if err != nil {
		t.Fatalf("tls.Listen: %v", err)
	}
	defer func() { _ = ln.Close() }()

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		tlsConn := conn.(*tls.Conn)
		_ = tlsConn.SetDeadline(time.Now().Add(3 * time.Second))
		_ = tlsConn.Handshake()
		_ = tlsConn.Close()
	}()

	clientCfg, err := NewTLSClientConfig(caPEM, "mtls.example", clientCertPEM, clientKeyPEM, false)
	if err != nil {
		t.Fatalf("NewTLSClientConfig: %v", err)
	}
	conn, err := tls.Dial("tcp", ln.Addr().String(), clientCfg)
	if err != nil {
		t.Fatalf("expected mTLS handshake success: %v", err)
	}
	_ = conn.Close()

	// New listener for missing client cert negative path.
	ln2, err := tls.Listen("tcp", "127.0.0.1:0", serverCfg)
	if err != nil {
		t.Fatalf("tls.Listen second: %v", err)
	}
	defer func() { _ = ln2.Close() }()
	handshakeErrCh := make(chan error, 1)
	go func() {
		conn, err := ln2.Accept()
		if err != nil {
			handshakeErrCh <- err
			return
		}
		tlsConn := conn.(*tls.Conn)
		_ = tlsConn.SetDeadline(time.Now().Add(3 * time.Second))
		handshakeErrCh <- tlsConn.Handshake()
		_ = tlsConn.Close()
	}()

	noClientCertCfg, err := NewTLSClientConfig(caPEM, "mtls.example", nil, nil, false)
	if err != nil {
		t.Fatalf("NewTLSClientConfig no cert: %v", err)
	}
	conn2, err := tls.Dial("tcp", ln2.Addr().String(), noClientCertCfg)
	if err == nil {
		_ = conn2.Close()
	}
	serverErr := <-handshakeErrCh
	if serverErr == nil {
		t.Fatal("expected server-side mTLS handshake failure when client cert missing")
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
