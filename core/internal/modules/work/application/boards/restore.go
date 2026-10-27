package boards

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (service *Service) RestoreBoard(ctx context.Context, scope ports.Scope,
	meta ports.Mutation, boardID domain.ID) (ports.OperationReceipt, error) {
	return service.writeLifecycle(ctx, scope, meta, boardID, domain.LifecycleActive,
		"work.board.restore")
}
