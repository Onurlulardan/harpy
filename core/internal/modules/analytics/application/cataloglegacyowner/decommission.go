package cataloglegacyowner

import (
	"context"
	"fmt"

	contract "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v1"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (owner *Owner) decommission(
	ctx context.Context,
	scope platformtx.TenantScope,
	command contract.Command,
	receipt contract.Receipt,
) error {
	componentID, err := uuid.Parse(command.Component.ID)
	if err != nil || componentID.Version() != 7 ||
		receipt.OperationID != command.OperationID ||
		receipt.Component != command.Component || !receipt.Ready ||
		receipt.CompletedAt.IsZero() {
		return fmt.Errorf(
			"Analytics Catalog lifecycle receipt is invalid",
		)
	}
	return nil
}
