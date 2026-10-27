package cataloglegacyowner

import "testing"

func TestRetireRejectsMissingDependencies(t *testing.T) {
	if _, err := New(nil); err == nil {
		t.Fatal("retire accepted missing dependencies")
	}
}
