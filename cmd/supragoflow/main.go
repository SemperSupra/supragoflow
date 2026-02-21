package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/SemperSupra/supragoflow/internal/securecomms"
	"github.com/SemperSupra/supragoflow/internal/version"
	"golang.org/x/crypto/ssh"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

func run(args []string, out io.Writer) error {
	if len(args) == 0 {
		return printUsage(out)
	}

	// Handle legacy flags for backward compatibility and smoke tests
	if args[0] == "--version" || args[0] == "-version" {
		return handleVersion(args, out)
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "version":
		return handleVersion(subArgs, out)
	case "check-tls":
		return handleCheckTLS(subArgs, out)
	case "check-ssh":
		return handleCheckSSH(subArgs, out)
	default:
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}
}

func printUsage(out io.Writer) error {
	_, err := fmt.Fprintln(out, "usage: supragoflow <subcommand> [flags]\n\nsubcommands:\n  version    Print version info\n  check-tls  Validate TLS configuration builder\n  check-ssh  Validate SSH configuration builder")
	return err
}

func handleVersion(args []string, out io.Writer) error {
	var asJSON bool
	fs := flag.NewFlagSet("version", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.BoolVar(&asJSON, "json", false, "print version info as JSON")
	// If called with --version (legacy), we need to skip the first arg if it's "version"
	if len(args) > 0 && (args[0] == "version" || args[0] == "--version" || args[0] == "-version") {
		_ = fs.Parse(args[1:])
	} else {
		_ = fs.Parse(args)
	}

	if asJSON {
		if expected := os.Getenv("SUPRAGOFLOW_EXPECT_SCHEMA_VERSION"); expected != "" && expected != version.SchemaVersion {
			return fmt.Errorf(
				"schema compatibility mismatch: expected schemaVersion=%s actual=%s; update consumer expectations or binary version",
				expected,
				version.SchemaVersion,
			)
		}
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(version.Info())
	}
	_, err := fmt.Fprintf(out, "version=%s commit=%s date=%s builtBy=%s\n",
		version.Version, version.Commit, version.Date, version.BuiltBy)
	return err
}

func handleCheckTLS(args []string, out io.Writer) error {
	_, _ = fmt.Fprintln(out, "Checking TLS configuration builder...")
	// Minimal check: can we build a config without crashing/failing on provider init?
	// We use dummy data that is syntactically valid PEM where possible.
	_, err := securecomms.NewTLSClientConfig(nil, "example.com", nil, nil, true)
	if err != nil {
		return fmt.Errorf("TLS client config failed: %w", err)
	}
	_, _ = fmt.Fprintln(out, "OK: TLS client config builder initialized successfully.")
	return nil
}

func handleCheckSSH(args []string, out io.Writer) error {
	_, _ = fmt.Fprintln(out, "Checking SSH configuration builder...")
	privateKeyPEM, knownHostsData, err := generateSmokeSSHMaterial()
	if err != nil {
		return fmt.Errorf("failed to generate SSH smoke material: %w", err)
	}
	_, err = securecomms.NewSSHClientConfig("user", privateKeyPEM, knownHostsData)
	if err != nil {
		return fmt.Errorf("SSH client config failed: %w", err)
	}
	_, _ = fmt.Fprintln(out, "OK: SSH client config builder initialized successfully.")
	return nil
}

func generateSmokeSSHMaterial() ([]byte, []byte, error) {
	userKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("generate user key: %w", err)
	}
	userKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(userKey),
	})

	hostKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("generate host key: %w", err)
	}
	hostSigner, err := ssh.NewSignerFromKey(hostKey)
	if err != nil {
		return nil, nil, fmt.Errorf("host signer: %w", err)
	}
	knownHosts := []byte("example.com " + string(ssh.MarshalAuthorizedKey(hostSigner.PublicKey())))

	return userKeyPEM, knownHosts, nil
}
