package billing

import invoiceapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/invoice"

type invoiceFinalizedPayload struct {
	InvoiceID       string `json:"invoice_id"`
	InvoiceRevision int64  `json:"invoice_revision"`
}

func encodeInvoiceFinalizedEvent(
	event invoiceapp.FinalizedEventIntent,
) ([]byte, string, error) {
	return canonicalEventPayload(invoiceFinalizedPayload{
		InvoiceID:       event.InvoiceID.String(),
		InvoiceRevision: event.InvoiceRevision,
	})
}
