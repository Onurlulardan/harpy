package cataloglegacyowner

import "testing"

func TestPurgeRejectsMissingDependencies(t *testing.T) {
	if _, err := New(nil, nil); err == nil {
		t.Fatal("purge accepted missing dependencies")
	}
}
