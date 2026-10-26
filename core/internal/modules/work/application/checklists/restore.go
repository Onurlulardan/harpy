package checklists

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (service *Service) RestoreChecklistItem(ctx context.Context, scope ports.Scope,
	meta ports.Mutation, taskID, itemID domain.ID, taskVersion int64,
) (ports.OperationReceipt, error) {
	return service.writeArchived(ctx, scope, meta, taskID, itemID, taskVersion, false)
}
