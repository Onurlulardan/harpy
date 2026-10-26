package labels

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (service *Service) RestoreLabel(ctx context.Context, scope ports.Scope,
	meta ports.Mutation, boardID, labelID domain.ID) (ports.OperationReceipt, error) {
	return service.writeArchived(ctx, scope, meta, boardID, labelID, false)
}
