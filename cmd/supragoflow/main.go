package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/SemperSupra/supragoflow/internal/securecomms"
	"github.com/SemperSupra/supragoflow/internal/version"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

var (
	jsonOutput bool
	// We handle --version legacy flag via the root command's PreRun or similar, or just a custom flag.
	rootVersionFlag bool
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(2)
	}
}

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "supragoflow",
		Short: "SupraGoFlow CLI",
		RunE: func(cmd *cobra.Command, args []string) error {
			if rootVersionFlag {
				return handleVersion(cmd, args)
			}
			return cmd.Help()
		},
	}
	rootCmd.Flags().BoolVarP(&rootVersionFlag, "version", "v", false, "print version info")
	// The original flag could be --json, but it was on the 'version' subcommand.
	// For --version --json we can add a persistent or local json flag.
	rootCmd.Flags().BoolVar(&jsonOutput, "json", false, "print version info as JSON")

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version info",
		RunE:  handleVersion,
	}
	versionCmd.Flags().BoolVar(&jsonOutput, "json", false, "print version info as JSON")

	checkTLSCmd := &cobra.Command{
		Use:   "check-tls",
		Short: "Validate TLS configuration builder",
		RunE:  handleCheckTLS,
	}

	checkSSHCmd := &cobra.Command{
		Use:   "check-ssh",
		Short: "Validate SSH configuration builder",
		RunE:  handleCheckSSH,
	}

	rootCmd.AddCommand(versionCmd, checkTLSCmd, checkSSHCmd)
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)
	return rootCmd
}

func handleVersion(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()
	if jsonOutput {
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

func handleCheckTLS(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()
	_, _ = fmt.Fprintln(out, "Checking TLS configuration builder...")
	_, err := securecomms.NewTLSClientConfig(nil, "example.com", nil, nil, true)
	if err != nil {
		return fmt.Errorf("TLS client config failed: %w", err)
	}
	_, _ = fmt.Fprintln(out, "OK: TLS client config builder initialized successfully.")
	return nil
}

func handleCheckSSH(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()
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
