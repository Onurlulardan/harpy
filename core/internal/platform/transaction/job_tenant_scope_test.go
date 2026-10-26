package transaction

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
)

func TestJobTenantScopeCarriesPersistedSlugOnlyWhileActive(t *testing.T) {
	state := &jobScopeState{
		tenantID:    uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		tenantSlug:  "harpy",
		transaction: &postgres.RuntimeTransaction{},
	}
	state.active.Store(true)
	scope := JobTenantScope{state: state}
	if scope.TenantSlug() != "harpy" {
		t.Fatalf("active job tenant slug = %q", scope.TenantSlug())
	}
	state.active.Store(false)
	if scope.TenantSlug() != "" {
		t.Fatalf("closed job tenant slug = %q", scope.TenantSlug())
	}
}

func TestJobTenantScopeRejectsMissingTenantSlug(t *testing.T) {
	state := &jobScopeState{
		tenantID:    uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		transaction: &postgres.RuntimeTransaction{},
	}
	state.active.Store(true)
	if (JobTenantScope{state: state}).TenantID() != uuid.Nil {
		t.Fatal("job scope without persisted tenant slug remained valid")
	}
}
