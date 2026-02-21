package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRunDefaultOutput(t *testing.T) {
	var out bytes.Buffer
	if err := run(nil, &out); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "usage: supragoflow") {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestRunVersionSubcommand(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"version"}, &out); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "version=") {
		t.Fatalf("expected version output, got %q", got)
	}
}

func TestRunCheckTLS(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"check-tls"}, &out); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "OK: TLS client config builder initialized successfully.") {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestRunCheckSSH(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"check-ssh"}, &out); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "OK: SSH client config builder initialized successfully.") {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestRunUnknownSubcommand(t *testing.T) {
	var out bytes.Buffer
	err := run([]string{"unknown-subcommand"}, &out)
	if err == nil {
		t.Fatal("expected error for unknown subcommand")
	}
	if !strings.Contains(err.Error(), "unknown subcommand") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunVersionText(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"--version"}, &out); err != nil {
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
	if err := run([]string{"--version", "--json"}, &out); err != nil {
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
	if payload["schemaVersion"] != "1" {
		t.Fatalf("unexpected schemaVersion %q", payload["schemaVersion"])
	}
}
