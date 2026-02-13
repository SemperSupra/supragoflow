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
	if !strings.Contains(got, "supragoflow: hello") {
		t.Fatalf("unexpected output: %q", got)
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
	for _, key := range []string{"version", "commit", "date", "builtBy"} {
		if _, ok := payload[key]; !ok {
			t.Fatalf("missing key %q in JSON output: %v", key, payload)
		}
	}
}
