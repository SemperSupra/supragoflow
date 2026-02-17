package securecomms

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
)

// secureCipherSuites is a list of strict, AEAD-based cipher suites for TLS 1.2.
// TLS 1.3 cipher suites are not configurable and are secure by default.
var secureCipherSuites = []uint16{
	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
	tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
}

// NewTLSClientConfig builds a secure client TLS config with optional mTLS certs.
// trustSystemCAs determines whether to include the system's root CAs in the trust pool.
// If set to false, only the provided caPEM will be trusted.
func NewTLSClientConfig(caPEM []byte, serverName string, clientCertPEM []byte, clientKeyPEM []byte, trustSystemCAs bool) (*tls.Config, error) {
	if serverName == "" {
		return nil, errors.New("serverName is required")
	}

	var rootCAs *x509.CertPool
	var err error

	if trustSystemCAs {
		rootCAs, err = x509.SystemCertPool()
		if err != nil || rootCAs == nil {
			// Fallback if system pool fails or is unavailable (e.g. windows container without certs)
			rootCAs = x509.NewCertPool()
		}
	} else {
		rootCAs = x509.NewCertPool()
	}

	if len(caPEM) > 0 {
		if ok := rootCAs.AppendCertsFromPEM(caPEM); !ok {
			return nil, errors.New("failed to append CA PEM")
		}
	}

	cfg := &tls.Config{
		MinVersion:   tls.VersionTLS12,
		CipherSuites: secureCipherSuites,
		RootCAs:      rootCAs,
		ServerName:   serverName,
	}

	if len(clientCertPEM) > 0 || len(clientKeyPEM) > 0 {
		if len(clientCertPEM) == 0 || len(clientKeyPEM) == 0 {
			return nil, errors.New("client cert and key must both be provided")
		}
		cert, err := tls.X509KeyPair(clientCertPEM, clientKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("failed to parse client keypair: %w", err)
		}
		cfg.Certificates = []tls.Certificate{cert}
	}

	return cfg, nil
}

// NewTLSServerConfig builds a secure server TLS config with optional client cert enforcement.
func NewTLSServerConfig(serverCertPEM []byte, serverKeyPEM []byte, clientCAPEM []byte, requireClientCert bool) (*tls.Config, error) {
	if len(serverCertPEM) == 0 || len(serverKeyPEM) == 0 {
		return nil, errors.New("server cert and key are required")
	}
	cert, err := tls.X509KeyPair(serverCertPEM, serverKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse server keypair: %w", err)
	}

	cfg := &tls.Config{
		MinVersion:   tls.VersionTLS12,
		CipherSuites: secureCipherSuites,
		Certificates: []tls.Certificate{cert},
	}

	if requireClientCert {
		if len(clientCAPEM) == 0 {
			return nil, errors.New("client CA is required when requireClientCert is true")
		}
		clientCAs := x509.NewCertPool()
		if ok := clientCAs.AppendCertsFromPEM(clientCAPEM); !ok {
			return nil, errors.New("failed to append client CA PEM")
		}
		cfg.ClientCAs = clientCAs
		cfg.ClientAuth = tls.RequireAndVerifyClientCert
	}

	return cfg, nil
}
