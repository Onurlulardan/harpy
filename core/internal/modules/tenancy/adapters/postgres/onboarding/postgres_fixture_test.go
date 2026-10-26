package onboarding

import (
	"context"
	"strings"
	"testing"
	"time"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy/domain"
	"github.com/google/uuid"
)

type rowFixture func(...any) error

func (fixture rowFixture) Scan(destinations ...any) error { return fixture(destinations...) }

func postgresPlanFixture(t *testing.T) tenantapi.OnboardingPlan {
	t.Helper()
	tenantID := uuid.New()
	schema, _ := domain.SchemaNameFromUUID(tenantID.String())
	now := time.Now().UTC().Truncate(time.Microsecond)
	input := tenantapi.OnboardingPlanInput{OperationID: uuid.New(), EnvironmentID: uuid.New(),
		TenantID: tenantID, FirstUserID: uuid.New(), RootOrganizationID: uuid.New(),
		ProtectedAdminBindingID: uuid.New(), InstanceOwnerBindingID: uuid.New(),
		Kind:           tenantapi.OnboardingEnvironmentInitialize,
		DeploymentMode: tenantapi.DeploymentModeOnPremise,
		BindingPurpose: tenantapi.TenantBindingOnPremisePrimary,
		SchemaName:     schema.String(), IdempotencyKey: "initialize",
		RequestSHA256: strings.Repeat("1", 64), ManifestSHA256: strings.Repeat("2", 64),
		Offering: tenantapi.BaseOfferingPin{VersionID: uuid.New(), Revision: 1,
			SHA256: strings.Repeat("3", 64), EffectiveAt: now,
			Source: tenantapi.OfferingSourcePin{Sequence: 1,
				SHA256: strings.Repeat("4", 64), Present: true}},
		RecoverySigning: tenantapi.RecoverySigningPin{Generation: 1,
			TrustBundleSHA256: strings.Repeat("5", 64), Present: true},
		Authority: tenantapi.ProtectedAuthorityPlan{RegistryBindingID: uuid.New(),
			ScopeKind: tenantapi.AuthorityScopeEnvironment}, CreatedAt: now}
	input.Authority.ScopeID = input.EnvironmentID
	plan, err := tenantapi.NewOnboardingPlan(input)
	if err != nil {
		t.Fatal(err)
	}
	return plan
}

func transitionRow(operationID uuid.UUID, state tenantapi.OnboardingState, revision int64) Row {
	return rowFixture(func(destinations ...any) error {
		*destinations[0].(*uuid.UUID) = operationID
		*destinations[1].(*string) = string(state)
		*destinations[2].(*int64) = revision
		*destinations[3].(*time.Time) = time.Now().UTC().Truncate(time.Microsecond)
		return nil
	})
}

func noopExecute(context.Context, string, ...any) error { return nil }
