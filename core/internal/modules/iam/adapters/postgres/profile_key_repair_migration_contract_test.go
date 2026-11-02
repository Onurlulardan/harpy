package postgres

import (
	"strings"
	"testing"
)

func TestIAMInstallationBindingUsesStableProfileKeySuccessor(t *testing.T) {
	if !strings.Contains(installationInitializeSQL, "iam_installation_initialize_v4") ||
		strings.Contains(installationInitializeSQL, "iam_installation_initialize_v2(") {
		t.Fatal("installation binding does not use stable profile key successor")
	}
}
