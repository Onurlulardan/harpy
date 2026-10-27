package tasks

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (service *Service) RestoreTask(ctx context.Context, scope ports.Scope,
	meta ports.Mutation, taskID domain.ID) (ports.OperationReceipt, error) {
	return service.writeArchived(ctx, scope, meta, taskID, false, "work.task.restore")
}
