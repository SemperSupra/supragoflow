package version

import "testing"

func TestInfo(t *testing.T) {
	i := Info()
	if i.SchemaVersion == "" {
		t.Fatal("expected SchemaVersion to be non-empty")
	}
	if i.Version == "" {
		t.Fatal("expected Version to be non-empty")
	}
}
