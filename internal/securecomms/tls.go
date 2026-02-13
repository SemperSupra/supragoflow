package securecomms

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
)

// NewTLSClientConfig builds a secure client TLS config with optional mTLS certs.
func NewTLSClientConfig(caPEM []byte, serverName string, clientCertPEM []byte, clientKeyPEM []byte) (*tls.Config, error) {
	if serverName == "" {
		return nil, errors.New("serverName is required")
	}

	rootCAs, err := x509.SystemCertPool()
	if err != nil || rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}
	if len(caPEM) > 0 {
		if ok := rootCAs.AppendCertsFromPEM(caPEM); !ok {
			return nil, errors.New("failed to append CA PEM")
		}
	}

	cfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
		RootCAs:    rootCAs,
		ServerName: serverName,
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
