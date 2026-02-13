package securecomms

import (
	"errors"
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// NewSSHClientConfig builds a strict SSH client config using known_hosts validation.
func NewSSHClientConfig(user string, privateKeyPEM []byte, knownHostsData []byte) (*ssh.ClientConfig, error) {
	if user == "" {
		return nil, errors.New("user is required")
	}
	if len(privateKeyPEM) == 0 {
		return nil, errors.New("private key is required")
	}
	if len(knownHostsData) == 0 {
		return nil, errors.New("known_hosts data is required")
	}

	signer, err := ssh.ParsePrivateKey(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

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
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: hostKeyCallback,
		Timeout:         10 * time.Second,
	}, nil
}
