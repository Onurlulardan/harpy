package deletion

import (
	"context"
	"time"

	deletionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/deletion"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) Confirm(ctx context.Context, task deletionapp.Task,
	receipt deletionapp.DeleteReceipt, billingReceiptID domain.ID, now time.Time,
) (deletionapp.Result, error) {
	if !storage.usable(ctx) || task.TenantID != storage.tenantID ||
		!task.ID.UUIDv7() || task.LeaseGeneration == 0 || receipt.ProviderRequestID == "" ||
		receipt.DeleteSHA256 == [32]byte{} || receipt.AbsenceSHA256 == [32]byte{} ||
		!billingReceiptID.UUIDv7() || now.IsZero() {
		return deletionapp.Result{}, deletionapp.ErrInvalid
	}
	var confirmedAt time.Time
	err := storage.queryRow(ctx, confirmSQL, task.ID.String(), task.LeaseGeneration,
		receipt.ProviderRequestID, receipt.DeleteSHA256[:], receipt.AbsenceSHA256[:],
		billingReceiptID.String(), now).Scan(&confirmedAt)
	if err != nil {
		if storage.notFound(err) {
			return deletionapp.Result{}, deletionapp.ErrConflict
		}
		return deletionapp.Result{}, deletionapp.ErrUnavailable
	}
	return deletionapp.Result{Task: task, BillingReceiptID: billingReceiptID,
		ConfirmedAt: confirmedAt.UTC()}, nil
}
