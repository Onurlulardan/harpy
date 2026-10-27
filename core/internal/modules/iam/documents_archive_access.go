package iam

import (
	"context"
	"errors"
	"fmt"

	authorizationpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/authorization"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	"github.com/Onurlulardan/harpy/core/internal/platform/actorcontext"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

var ErrDocumentsArchiveAccess = errors.New("IAM Documents archive access is denied")

type DocumentsArchiveGenerationResolver interface {
	ResolveTenantDocumentsGenerations(context.Context, uuid.UUID) (int64, int64, error)
}

type DocumentsArchiveAccessRuntime struct {
	generations DocumentsArchiveGenerationResolver
	scopes      *platformtx.Coordinator
	authorizer  authorization.Authorizer
}

func NewDocumentsArchiveAccessRuntime(generations DocumentsArchiveGenerationResolver,
	scopes *platformtx.Coordinator,
) (*DocumentsArchiveAccessRuntime, error) {
	if generations == nil || scopes == nil {
		return nil, fmt.Errorf("IAM Documents archive access dependencies are required")
	}
	port, err := authorizationpostgres.NewDocumentsAuthorizer()
	if err != nil {
		return nil, fmt.Errorf("build IAM Documents archive evidence: %w", err)
	}
	authorizer, err := authorization.NewAuthorizer(port)
	if err != nil {
		return nil, fmt.Errorf("build IAM Documents archive authorizer: %w", err)
	}
	return &DocumentsArchiveAccessRuntime{generations: generations,
		scopes: scopes, authorizer: authorizer}, nil
}

func (runtime *DocumentsArchiveAccessRuntime) Refresh(ctx context.Context,
	tenantID, actorID uuid.UUID, pinnedRevision uint64, itemIDs []uuid.UUID,
) error {
	if runtime == nil || ctx == nil || ctx.Err() != nil || tenantID == uuid.Nil ||
		actorID == uuid.Nil || pinnedRevision == 0 || !validArchiveItems(itemIDs) ||
		!platformtx.NetworkAllowed(ctx) {
		return ErrDocumentsArchiveAccess
	}
	_, generation, err := runtime.generations.
		ResolveTenantDocumentsGenerations(ctx, tenantID)
	if err != nil || generation < 1 {
		return ErrDocumentsArchiveAccess
	}
	actor, err := actorcontext.Issue(actorID, tenantID)
	if err != nil {
		return ErrDocumentsArchiveAccess
	}
	operation, err := platformtx.NewSnapshotAuthorityIssuer().Issue(
		"documents.archive.build", map[string]int64{"documents": generation})
	if err != nil {
		return ErrDocumentsArchiveAccess
	}
	request := platformtx.ScopeRequest{Actor: actor, Operation: operation,
		Mode: platformtx.SnapshotQuery}
	return runtime.scopes.InTenant(ctx, request,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			probe, probeErr := authorization.NewCurrentActorRequest(
				"documents.item", "download", actorID)
			if probeErr != nil {
				return ErrDocumentsArchiveAccess
			}
			decision, decisionErr := runtime.authorizer.
				CurrentActorDecision(scoped, scope, probe)
			if decisionErr != nil || decision.Validate() != nil || !decision.Allowed() ||
				uint64(decision.AuthorizationVersion()) != pinnedRevision {
				return ErrDocumentsArchiveAccess
			}
			return nil
		})
}

func validArchiveItems(values []uuid.UUID) bool {
	if len(values) == 0 || len(values) > 10_000 {
		return false
	}
	for _, value := range values {
		if value == uuid.Nil || value.Version() != 7 {
			return false
		}
	}
	return true
}
