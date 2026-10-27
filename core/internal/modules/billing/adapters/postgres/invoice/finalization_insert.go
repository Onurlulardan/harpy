package invoice

import (
	"context"
	"fmt"

	invoiceapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/invoice"
	"github.com/google/uuid"
)

func (storage *store) writeFinalization(
	ctx context.Context,
	value invoiceapp.Finalization,
) error {
	authorityID, err := storage.authorityID(ctx)
	if err != nil {
		return err
	}
	invoiceTotal := value.Totals.InvoiceTotal()
	settled := value.Totals.Settled()
	amountDue := value.Totals.AmountDue()
	creditBalance := value.Totals.CreditBalance()
	var invoiceID uuid.UUID
	err = storage.queryRow(ctx, writeFinalizationSQL, authorityID, storage.tenantID,
		value.InvoiceID, value.ExpectedRevision, value.ResultRevision,
		invoiceTotal.MinorUnits(), settled.MinorUnits(), amountDue.MinorUnits(),
		creditBalance.MinorUnits(), invoiceTotal.CurrencyRegistryVersion(),
		value.PolicySHA256[:], value.FinalizedAt, value.SourceOperationID,
	).Scan(&invoiceID)
	if err != nil {
		return fmt.Errorf("write billing invoice finalization: %w", err)
	}
	if invoiceID != value.InvoiceID {
		return invoiceapp.ErrSnapshot
	}
	return nil
}
