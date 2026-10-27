package onboarding

import (
	"context"
	_ "embed"
	"regexp"

	"github.com/google/uuid"
)

//go:embed sql/platform/reconcile_current_authorities.sql
var reconcileCurrentAuthoritiesSQL string

type AuthorityReconciliation struct {
	BillingGlobalReconciled int
	BillingGlobalRemaining  int
	BillingTenantReconciled int
	BillingTenantRemaining  int
	AnalyticsReconciled     int
	AnalyticsRemaining      int
}

type AuthoritySynchronizer struct{ queryRow QueryRow }

var authorityDigestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

func NewAuthoritySynchronizer(queryRow QueryRow) (*AuthoritySynchronizer, error) {
	if queryRow == nil {
		return nil, errUnavailable
	}
	return &AuthoritySynchronizer{queryRow: queryRow}, nil
}

func (synchronizer *AuthoritySynchronizer) Synchronize(
	ctx context.Context,
	operationID uuid.UUID,
	manifestSHA256 string,
) (AuthorityReconciliation, error) {
	if synchronizer == nil || synchronizer.queryRow == nil ||
		ctx == nil || ctx.Err() != nil || operationID == uuid.Nil ||
		!authorityDigestPattern.MatchString(manifestSHA256) {
		return AuthorityReconciliation{}, errUnavailable
	}
	var result AuthorityReconciliation
	err := synchronizer.queryRow(
		ctx, reconcileCurrentAuthoritiesSQL, operationID, manifestSHA256,
	).Scan(
		&result.BillingGlobalReconciled,
		&result.BillingGlobalRemaining,
		&result.BillingTenantReconciled,
		&result.BillingTenantRemaining,
		&result.AnalyticsReconciled,
		&result.AnalyticsRemaining,
	)
	return result, err
}
