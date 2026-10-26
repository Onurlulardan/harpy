package foundation

import (
	"context"
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

type authorityEvidenceKey struct{}

type AuthorityEvidence struct {
	ActorID, TenantID                                      domain.ID
	TokenVersion, AuthorizationVersion, AuthorizationClock int64
	Permission, Operation                                  string
	StepUpAt                                               time.Time
}

func NewAuthorityEvidence(actorID, tenantID domain.ID, tokenVersion,
	authorizationVersion, authorizationClock int64, permission, operation string,
	stepUpAt time.Time,
) (AuthorityEvidence, error) {
	evidence := AuthorityEvidence{ActorID: actorID, TenantID: tenantID,
		TokenVersion: tokenVersion, AuthorizationVersion: authorizationVersion,
		AuthorizationClock: authorizationClock, Permission: permission,
		Operation: operation, StepUpAt: stepUpAt.UTC()}
	if !actorID.Valid() || !tenantID.Valid() || tokenVersion < 1 ||
		authorizationVersion < 1 || authorizationClock < 1 || permission == "" ||
		operation == "" || (!stepUpAt.IsZero() && stepUpAt.Nanosecond() != 0) {
		return AuthorityEvidence{}, fmt.Errorf("Documents authority evidence is invalid")
	}
	return evidence, nil
}

func WithAuthorityEvidence(ctx context.Context, evidence AuthorityEvidence) (
	context.Context, error,
) {
	if ctx == nil || ctx.Err() != nil || !evidence.ActorID.Valid() ||
		!evidence.TenantID.Valid() || evidence.Permission == "" || evidence.Operation == "" {
		return nil, fmt.Errorf("Documents authority evidence is invalid")
	}
	return context.WithValue(ctx, authorityEvidenceKey{}, evidence), nil
}

func AuthorityEvidenceFromContext(ctx context.Context) (AuthorityEvidence, bool) {
	if ctx == nil {
		return AuthorityEvidence{}, false
	}
	evidence, found := ctx.Value(authorityEvidenceKey{}).(AuthorityEvidence)
	return evidence, found
}
