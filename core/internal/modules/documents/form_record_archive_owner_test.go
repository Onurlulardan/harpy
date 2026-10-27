package documents

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type formRecordArchiveGenerations struct{}

func (formRecordArchiveGenerations) ResolveTenantBillingGenerations(
	context.Context, uuid.UUID,
) (int64, int64, int64, error) {
	return 1, 1, 1, nil
}

func (formRecordArchiveGenerations) ResolveTenantDocumentsGenerations(
	context.Context, uuid.UUID,
) (int64, int64, error) {
	return 1, 1, nil
}

type formRecordArchiveScopes struct{}

func (formRecordArchiveScopes) InTenant(context.Context, platformtx.ScopeRequest,
	func(context.Context, platformtx.TenantScope) error,
) error {
	return nil
}

type formRecordArchiveAccess struct{}

func (formRecordArchiveAccess) RefreshDocumentsArchive(
	context.Context, string, string, uint64, []string,
) error {
	return nil
}

type formRecordArchiveJobs struct{}

func (formRecordArchiveJobs) Enqueue(context.Context, platformtx.TenantScope,
	foundation.ID, foundation.ID, []byte,
) error {
	return nil
}

func TestNewFormRecordArchiveOwnerBuildsConcreteDocumentsAdapters(t *testing.T) {
	owner, err := NewFormRecordArchiveOwner(FormRecordArchiveOwnerDependencies{
		Generations: formRecordArchiveGenerations{}, Scopes: formRecordArchiveScopes{},
		Access: formRecordArchiveAccess{}, Jobs: formRecordArchiveJobs{},
		IDs: platformIDs{}, Clock: platformClock{},
	})
	if err != nil || owner == nil {
		t.Fatalf("build form-record archive owner: owner=%T err=%v", owner, err)
	}
}
