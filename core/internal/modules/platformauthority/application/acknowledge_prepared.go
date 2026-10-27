package application

import (
	"context"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
)

func (service *Coordinator) AcknowledgePrepared(
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
	if operation.Purpose != "initial_environment_authority" &&
		operation.Purpose != "initial_tenant_authority" {
		return api.AuthorityReceipt{}, api.ErrConflict
	}
	return service.completeOperation(ctx, operation)
}

var _ api.InitialAuthority = (*Coordinator)(nil)
