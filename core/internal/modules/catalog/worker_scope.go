package catalog

import (
	"context"

	lifecyclepostgres "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/postgres/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (processor *workerProcessor) scopeRequest(
	ctx context.Context,
	tenantID uuid.UUID,
	snapshot lifecyclepostgres.JobSnapshot,
) (platformtx.ScopeRequest, error) {
	groups, domains, err := resolveResumeGenerations(ctx, processor.generations, tenantID)
	if err != nil {
		return platformtx.ScopeRequest{}, asyncUnavailable()
	}
	request, err := platformtx.AdmitMultiGroupCommandScope(
		snapshot.ActorID, tenantID, "catalog.operation.resume", groups, domains,
		[]string{"analytics", "catalog", "documents", "forms", "iam", "workflow"},
	)
	if err != nil {
		return platformtx.ScopeRequest{}, asyncUnavailable()
	}
	return request, nil
}
