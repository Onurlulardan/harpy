package invoice

import (
	"context"
	"fmt"
	"strings"

	invoiceapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/invoice"
	invoicedomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/invoice"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/money"
	"github.com/google/uuid"
)

func (storage *store) findFinalization(
	ctx context.Context,
	sourceOperationID uuid.UUID,
) (invoiceapp.Finalization, bool, error) {
	authorityID, err := storage.authorityID(ctx)
	if err != nil {
		return invoiceapp.Finalization{}, false, err
	}
	result := invoiceapp.Finalization{}
	var invoiceTotal, settled, amountDue, creditBalance int64
	var currency string
	var registryVersion int64
	var policySHA256 []byte
	err = storage.queryRow(ctx, findFinalizationSQL, authorityID, storage.tenantID,
		sourceOperationID,
	).Scan(&result.InvoiceID, &result.SourceOperationID, &result.ExpectedRevision,
		&result.ResultRevision, &invoiceTotal, &settled, &amountDue, &creditBalance,
		&currency, &registryVersion, &policySHA256, &result.FinalizedAt)
	if storage.notFound(err) {
		return invoiceapp.Finalization{}, false, nil
	}
	if err != nil {
		return invoiceapp.Finalization{}, false,
			fmt.Errorf("find billing invoice finalization: %w", err)
	}
	currency = strings.TrimSpace(currency)
	values := make([]money.Money, 4)
	for index, units := range []int64{invoiceTotal, settled, amountDue, creditBalance} {
		values[index], err = money.New(units, currency, registryVersion)
		if err != nil {
			return invoiceapp.Finalization{}, false, invoiceapp.ErrSnapshot
		}
	}
	result.Totals, err = invoicedomain.NewVerifiedTotals(
		values[0], values[1], values[2], values[3],
	)
	if err != nil || len(policySHA256) != len(result.PolicySHA256) {
		return invoiceapp.Finalization{}, false, invoiceapp.ErrSnapshot
	}
	copy(result.PolicySHA256[:], policySHA256)
	return result, true, nil
}
