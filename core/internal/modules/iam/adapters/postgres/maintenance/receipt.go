package maintenance

import (
	"context"
	"fmt"

	maintenanceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/maintenance"
)

func (store *store) readReceipt(
	ctx context.Context,
	claim maintenanceapp.Claim,
) (maintenanceapp.Outcome, bool, error) {
	values, err := claimValues(claim)
	if err != nil {
		return maintenanceapp.Outcome{}, false, err
	}
	outcome, found, err := scanReceiptOutcome(store.database.QueryRow(
		ctx, readReceiptSQL, values[0], values[1], string(claim.Kind),
		values[3], values[4], values[5],
	), claim.Kind)
	if err != nil {
		return maintenanceapp.Outcome{}, false,
			fmt.Errorf("read IAM maintenance receipt: %w", err)
	}
	return outcome, found, nil
}
