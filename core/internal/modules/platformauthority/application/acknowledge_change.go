package application

import (
	"context"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
)

func (service *Coordinator) AcknowledgeChange(
	ctx context.Context,
	operationID string,
) (api.AuthorityReceipt, error) {
	if strings.TrimSpace(operationID) == "" {
		return api.AuthorityReceipt{}, api.ErrConflict
	}
	operation, err := service.store.LoadOperation(ctx, operationID)
	if err != nil {
		return api.AuthorityReceipt{}, mapStoreError(err)
	}
	if !protectedChangePurpose(operation.Purpose) {
		return api.AuthorityReceipt{}, api.ErrConflict
	}
	return service.completeOperation(ctx, operation)
}

var _ api.ProtectedAuthority = (*Coordinator)(nil)
