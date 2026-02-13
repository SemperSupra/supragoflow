package version

import "testing"

func TestInfo(t *testing.T) {
	i := Info()
	if i.Version == "" {
		t.Fatal("expected Version to be non-empty")
	}
}
