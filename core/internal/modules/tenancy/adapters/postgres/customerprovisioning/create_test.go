package customerprovisioning

import (
	"context"
	"errors"
	"testing"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	provisioning "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerprovisioning"
)

func TestCreateRejectsInvalidPlan(t *testing.T) {
	store := &Store{}
	if _, err := store.Create(context.Background(), tenantapi.OnboardingPlan{}, provisioning.AcceptedRequest{}); !errors.Is(err, provisioning.ErrInvalid) {
		t.Fatalf("Create() error = %v, want %v", err, provisioning.ErrInvalid)
	}
}
