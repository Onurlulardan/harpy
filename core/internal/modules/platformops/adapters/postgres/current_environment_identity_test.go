package postgres

import (
	"context"
	"strings"
	"testing"
)

func TestCurrentIdentityUsesOwnedProjection(t *testing.T) {
	if !strings.Contains(currentEnvironmentIdentitySQL,
		"platformops_current_environment_identity_v1") {
		t.Fatal("current identity bypasses the PlatformOps owner routine")
	}
	if _, err := (*Store)(nil).CurrentIdentity(
		context.Background(), "invalid",
	); err == nil {
		t.Fatal("invalid current identity unexpectedly succeeded")
	}
}
