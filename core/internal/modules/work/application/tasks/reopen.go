package tasks

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (service *Service) ReopenTask(ctx context.Context, scope ports.Scope,
	meta ports.Mutation, taskID domain.ID) (ports.OperationReceipt, error) {
	return service.writeTransition(ctx, scope, meta, taskID, domain.TransitionReopen)
}
