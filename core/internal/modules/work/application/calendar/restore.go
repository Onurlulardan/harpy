package calendar

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (service *Service) RestoreEvent(ctx context.Context, scope ports.Scope,
	meta ports.Mutation, eventID domain.ID) (ports.OperationReceipt, error) {
	return service.writeArchived(ctx, scope, meta, eventID, false)
}
