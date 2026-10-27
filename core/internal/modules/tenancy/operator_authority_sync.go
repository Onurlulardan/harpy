package tenancy

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	onboardingpostgres "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/adapters/postgres/onboarding"
	operatorbindingpostgres "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/adapters/postgres/operatorbinding"
	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

var errTenantAuthorityReconciliationIncomplete = errors.New(
	"tenant authority reconciliation is incomplete",
)

var tenantAuthorityDigestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

type TenantAuthorityReconciliation = onboardingpostgres.AuthorityReconciliation

type TenantAuthoritySynchronizer struct {
	synchronize func(context.Context, uuid.UUID, string) (
		TenantAuthorityReconciliation, error,
	)
}

func BindTenantAuthoritySynchronizer(
	pool operatorbindingpostgres.Pool,
) (*TenantAuthoritySynchronizer, error) {
	return operatorbindingpostgres.Bind(pool, func(
		database operatorbindingpostgres.Database,
	) (*TenantAuthoritySynchronizer, error) {
		bound, err := onboardingpostgres.NewAuthoritySynchronizer(func(
			ctx context.Context,
			statement string,
			arguments ...any,
		) onboardingpostgres.Row {
			return database.QueryRow(ctx, statement, arguments...)
		})
		if err != nil {
			return nil, fmt.Errorf("bind tenant authority reconciler: %w", err)
		}
		return newTenantAuthoritySynchronizer(bound.Synchronize)
	})
}

func newTenantAuthoritySynchronizer(
	synchronize func(context.Context, uuid.UUID, string) (
		TenantAuthorityReconciliation, error,
	),
) (*TenantAuthoritySynchronizer, error) {
	if synchronize == nil {
		return nil, tenantapi.ErrOnboardingUnavailable
	}
	return &TenantAuthoritySynchronizer{synchronize: synchronize}, nil
}

func (synchronizer *TenantAuthoritySynchronizer) Synchronize(
	ctx context.Context,
	operationID uuid.UUID,
	manifestSHA256 string,
) (TenantAuthorityReconciliation, error) {
	if synchronizer == nil || synchronizer.synchronize == nil ||
		ctx == nil || ctx.Err() != nil || operationID == uuid.Nil ||
		!tenantAuthorityDigestPattern.MatchString(manifestSHA256) {
		return TenantAuthorityReconciliation{}, tenantapi.ErrOnboardingUnavailable
	}
	result, err := synchronizer.synchronize(ctx, operationID, manifestSHA256)
	if err != nil {
		return TenantAuthorityReconciliation{},
			fmt.Errorf("reconcile current tenant authorities: %w", err)
	}
	if result.BillingGlobalReconciled < 0 || result.BillingGlobalRemaining != 0 ||
		result.BillingTenantReconciled < 0 || result.BillingTenantRemaining != 0 ||
		result.AnalyticsReconciled < 0 || result.AnalyticsRemaining != 0 {
		return TenantAuthorityReconciliation{},
			errTenantAuthorityReconciliationIncomplete
	}
	return result, nil
}
