package documents

import (
	"context"

	itemapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/item"
	knowledgeapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/knowledgefile"
	"github.com/Onurlulardan/harpy/core/internal/platform/actorcontext"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type knowledgeFileScopes struct {
	generations HTTPGenerationResolver
	coordinator HTTPScopeCoordinator
}

func (runtime knowledgeFileScopes) InCommand(ctx context.Context,
	access knowledgeapp.Access, operation string,
	action func(context.Context, platformtx.TenantScope) error,
) error {
	if action == nil || !knowledgeFileAccess(access) {
		return ErrDocumentsRuntime
	}
	request, err := runtime.commandRequest(ctx, access, operation)
	if err != nil {
		return err
	}
	return runtime.coordinator.InTenant(ctx, request,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			return action(context.WithValue(scoped, knowledgeFileEvidenceKey{}, access), scope)
		})
}

func (runtime knowledgeFileScopes) InSnapshot(ctx context.Context,
	access knowledgeapp.Access, operation string,
	action func(context.Context, platformtx.TenantScope) error,
) error {
	if action == nil || !knowledgeFileAccess(access) {
		return ErrDocumentsRuntime
	}
	_, generation, err := runtime.generations.ResolveTenantDocumentsGenerations(
		ctx, uuid.UUID(access.TenantID))
	actor, actorErr := actorcontext.Issue(uuid.UUID(access.ActorID), uuid.UUID(access.TenantID))
	verified, operationErr := platformtx.NewSnapshotAuthorityIssuer().Issue(operation,
		map[string]int64{"documents": generation})
	if err != nil || actorErr != nil || operationErr != nil || operation != "documents.item.version.download" {
		return ErrDocumentsRuntime
	}
	request := platformtx.ScopeRequest{Actor: actor, Operation: verified,
		Mode: platformtx.SnapshotQuery}
	return runtime.coordinator.InTenant(ctx, request,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			return action(context.WithValue(scoped, knowledgeFileEvidenceKey{}, access), scope)
		})
}

func (runtime knowledgeFileScopes) commandRequest(ctx context.Context,
	access knowledgeapp.Access, operation string,
) (platformtx.ScopeRequest, error) {
	tenantID, actorID := uuid.UUID(access.TenantID), uuid.UUID(access.ActorID)
	if operation == "documents.upload.create" || operation == "documents.upload.put" {
		billingGroup, billing, err := runtime.generations.
			ResolveTenantBillingGenerations(ctx, tenantID)
		documentsGroup, documentsGeneration, documentsErr := runtime.generations.
			ResolveTenantDocumentsGenerations(ctx, tenantID)
		if err != nil || documentsErr != nil || !positiveGenerations([]int64{
			billingGroup, billing, documentsGroup, documentsGeneration}) {
			return platformtx.ScopeRequest{}, ErrDocumentsRuntime
		}
		return platformtx.AdmitMultiGroupCommandScope(actorID, tenantID, operation,
			[]platformtx.CommandAuthorityGroup{{ID: "billing", Generation: billingGroup},
				{ID: "documents", Generation: documentsGroup}},
			[]platformtx.CommandAuthorityDomain{{Name: "billing", Generation: billing},
				{Name: "documents", Generation: documentsGeneration}}, []string{"billing", "documents"})
	}
	group, generation, err := runtime.generations.ResolveTenantDocumentsGenerations(ctx, tenantID)
	if err != nil || group < 1 || generation < 1 ||
		operation != "documents.folder.create" && operation != itemapp.OperationDelete {
		return platformtx.ScopeRequest{}, ErrDocumentsRuntime
	}
	return platformtx.AdmitCommandScope(actorID, tenantID, operation,
		"documents", group, map[string]int64{"documents": generation})
}

func knowledgeFileAccess(value knowledgeapp.Access) bool {
	return value.TenantID.Valid() && value.ActorID.Valid() && value.AuthorizationVersion > 0
}

var _ knowledgeapp.Scopes = knowledgeFileScopes{}
