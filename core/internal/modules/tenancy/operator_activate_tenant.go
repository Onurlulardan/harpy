package tenancy

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	onboardingpostgres "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/adapters/postgres/onboarding"
	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

var operatorTenantSlugPattern = regexp.MustCompile(
	`^[a-z0-9](?:[a-z0-9-]{0,98}[a-z0-9])?$`,
)

type ActivateTenantInput struct {
	OperationID, EnvironmentID, TenantID, OwnerUserID tenantapi.UUID
	Name, Slug, SchemaName, Kind, BindingPurpose      string
	ManifestSHA256                                    string
}

func ActivateTenant(
	ctx context.Context,
	executor OperatorActivationExecutor,
	input ActivateTenantInput,
) error {
	if ctx == nil || ctx.Err() != nil || executor.QueryRow == nil ||
		!validActivateTenantInput(input) {
		return tenantapi.ErrOnboardingInvalid
	}
	binding, err := onboardingpostgres.BindInstallation(onboardingpostgres.InstallationDatabase{
		QueryRow: func(ctx context.Context, statement string, arguments ...any) onboardingpostgres.InstallationRow {
			return executor.QueryRow(ctx, statement, arguments...)
		},
	})
	if err != nil {
		return fmt.Errorf("bind tenancy installation owner: %w", err)
	}
	_, err = binding.ActivateTenant(ctx, onboardingpostgres.InstallationTenant{
		OperationID: uuid.UUID(input.OperationID), EnvironmentID: uuid.UUID(input.EnvironmentID),
		TenantID: uuid.UUID(input.TenantID), OwnerUserID: uuid.UUID(input.OwnerUserID),
		Name: input.Name, Slug: input.Slug, SchemaName: input.SchemaName,
		Kind: input.Kind, BindingPurpose: input.BindingPurpose,
		ManifestSHA256: input.ManifestSHA256,
	})
	if err != nil {
		return fmt.Errorf("initialize tenant activation: %w", err)
	}
	return nil
}

func validActivateTenantInput(input ActivateTenantInput) bool {
	ids := []tenantapi.UUID{input.OperationID, input.EnvironmentID,
		input.TenantID, input.OwnerUserID}
	for _, id := range ids {
		if id == tenantapi.NilUUID {
			return false
		}
	}
	if input.Name == "" || len(input.Name) > 255 ||
		strings.TrimSpace(input.Name) != input.Name ||
		!operatorTenantSlugPattern.MatchString(input.Slug) ||
		len(input.SchemaName) < 1 || len(input.SchemaName) > 63 ||
		!strings.HasPrefix(input.SchemaName, "t_") ||
		!validOperatorDigest(input.ManifestSHA256) {
		return false
	}
	return (input.Kind == "environment_initialize" &&
		(input.BindingPurpose == "onpremise_primary" ||
			input.BindingPurpose == "saas_control")) ||
		(input.Kind == "saas_customer_tenant_provision" &&
			input.BindingPurpose == "saas_customer")
}

func validOperatorDigest(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, character := range value {
		if !strings.ContainsRune("0123456789abcdef", character) {
			return false
		}
	}
	return true
}
