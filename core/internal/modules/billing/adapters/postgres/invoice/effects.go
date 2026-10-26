package invoice

import (
	"context"

	invoiceapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/invoice"
)

type Effects interface {
	PublishFinalized(
		context.Context,
		invoiceapp.FinalizedEventIntent,
		invoiceapp.Finalization,
	) error
}
