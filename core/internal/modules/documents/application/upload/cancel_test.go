package upload

import (
	"context"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestCancelBoundReleasesOnlyImmediatePreEffectReservation(t *testing.T) {
	billing := &uploadBilling{}
	application := uploadApplication([]domain.ID{uploadReceipt}, billing)
	store := uploadStore{cancel: func(_ context.Context, id, receipt domain.ID,
		now time.Time,
	) (CancelResult, error) {
		if id != uploadHandle || receipt != uploadReceipt || now != uploadNow {
			t.Fatal("unexpected cancellation input")
		}
		return CancelResult{Snapshot: uploadSnapshot(), ReleaseReservation: true}, nil
	}}
	result, err := application.cancelBound(context.Background(), platformtx.TenantScope{},
		store, uploadHandle)
	if err != nil || result.ID != uploadHandle || billing.released != uploadReservation {
		t.Fatalf("result=%+v released=%s err=%v", result, billing.released.String(), err)
	}
}
