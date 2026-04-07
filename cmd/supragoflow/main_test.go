package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func runCmd(args []string, out *bytes.Buffer) error {
	cmd := newRootCmd()
	cmd.SetArgs(args)
	cmd.SetOut(out)
	cmd.SetErr(out)
	return cmd.Execute()
}

func TestRunCapabilitySetTable(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		wantSubstrs []string
	}{
		{
			name:        "usage",
			args:        nil,
			wantErr:     false,
			wantSubstrs: []string{"Usage:", "Available Commands:"},
		},
		{
			name:        "version subcommand",
			args:        []string{"version"},
			wantErr:     false,
			wantSubstrs: []string{"version=", "commit=", "date=", "builtBy="},
		},
		{
			name:        "legacy version flag",
			args:        []string{"--version"},
			wantErr:     false,
			wantSubstrs: []string{"version=", "commit=", "date=", "builtBy="},
		},
		{
			name:        "check tls",
			args:        []string{"check-tls"},
			wantErr:     false,
			wantSubstrs: []string{"Checking TLS configuration builder...", "OK: TLS client config builder initialized successfully."},
		},
		{
			name:        "check ssh",
			args:        []string{"check-ssh"},
			wantErr:     false,
			wantSubstrs: []string{"Checking SSH configuration builder...", "OK: SSH client config builder initialized successfully."},
		},
		{
			name:        "unknown subcommand",
			args:        []string{"nope"},
			wantErr:     true,
			wantSubstrs: []string{"unknown command \"nope\""},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			err := runCmd(tc.args, &out)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got err=%v", tc.wantErr, err)
			}

			gotText := out.String()
			for _, want := range tc.wantSubstrs {
				if tc.wantErr {
					if err == nil || !strings.Contains(err.Error(), want) {
						t.Fatalf("expected error containing %q, got %v", want, err)
					}
					continue
				}
				if !strings.Contains(gotText, want) {
					t.Fatalf("expected output to contain %q, got %q", want, gotText)
				}
			}
		})
	}
}

func TestRunDefaultOutput(t *testing.T) {
	var out bytes.Buffer
	if err := runCmd(nil, &out); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "Usage:") {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestRunVersionSubcommand(t *testing.T) {
	var out bytes.Buffer
	if err := runCmd([]string{"version"}, &out); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "version=") {
		t.Fatalf("expected version output, got %q", got)
	}
}

func TestRunCheckTLS(t *testing.T) {
	var out bytes.Buffer
	if err := runCmd([]string{"check-tls"}, &out); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "OK: TLS client config builder initialized successfully.") {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestRunCheckSSH(t *testing.T) {
	var out bytes.Buffer
	if err := runCmd([]string{"check-ssh"}, &out); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "OK: SSH client config builder initialized successfully.") {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestRunUnknownSubcommand(t *testing.T) {
	var out bytes.Buffer
	err := runCmd([]string{"unknown-subcommand"}, &out)
	if err == nil {
		t.Fatal("expected error for unknown subcommand")
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunVersionText(t *testing.T) {
	var out bytes.Buffer
	if err := runCmd([]string{"--version"}, &out); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	got := out.String()
	for _, want := range []string{"version=", "commit=", "date=", "builtBy="} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected output to contain %q, got %q", want, got)
		}
	}
}

func TestRunVersionJSON(t *testing.T) {
	var out bytes.Buffer
	if err := runCmd([]string{"--version", "--json"}, &out); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	var payload map[string]string
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid JSON output: %v\noutput: %q", err, out.String())
	}
	for _, key := range []string{"schemaVersion", "version", "commit", "date", "builtBy"} {
		if _, ok := payload[key]; !ok {
			t.Fatalf("missing key %q in JSON output: %v", key, payload)
		}
	}
	if len(payload) != 5 {
		t.Fatalf("unexpected key count in JSON output: %v", payload)
	}
	if payload["schemaVersion"] != "1" {
		t.Fatalf("unexpected schemaVersion %q", payload["schemaVersion"])
	}
}

func TestRunVersionJSONSchemaExpectationMatch(t *testing.T) {
	t.Setenv("SUPRAGOFLOW_EXPECT_SCHEMA_VERSION", "1")
	var out bytes.Buffer
	if err := runCmd([]string{"--version", "--json"}, &out); err != nil {
		t.Fatalf("run returned error with matching schema expectation: %v", err)
	}
}

func TestRunVersionJSONSchemaExpectationMismatch(t *testing.T) {
	t.Setenv("SUPRAGOFLOW_EXPECT_SCHEMA_VERSION", "999")
	var out bytes.Buffer
	err := runCmd([]string{"--version", "--json"}, &out)
	if err == nil {
		t.Fatal("expected error for schema mismatch")
	}
	if !strings.Contains(err.Error(), "schema compatibility mismatch") {
		t.Fatalf("unexpected error: %v", err)
	}
}
