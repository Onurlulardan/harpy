package deploymentpolicy

import (
	"context"
	"errors"
	"testing"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
)

func TestCreateRejectsInvalidRequest(t *testing.T) {
	initializer := &policyInitializer{}
	if _, err := initializer.Create(
		context.Background(), tenantapi.InitializeDeploymentPolicyRequest{},
	); !errors.Is(err, errPolicyUnavailable) {
		t.Fatalf("Create() error = %v, want %v", err, errPolicyUnavailable)
	}
}
