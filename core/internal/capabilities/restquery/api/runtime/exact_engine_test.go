package runtime

import "testing"

func TestNewExactRejectsMissingOwnerDependencies(t *testing.T) {
	if _, err := NewExact("forms", nil, nil, CursorKeys{}); err == nil {
		t.Fatal("missing exact engine dependencies were accepted")
	}
}
