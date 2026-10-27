package application

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func TestSubmitRejectsUnboundTenantScope(t *testing.T) {
	application, err := New(fixtureBinder{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = application.Submit(context.Background(), platformtx.TenantScope{}, SubmitCommand{
		ProposedJobID: uuid.MustParse(fixtureJobID), Plan: fixturePlan(),
	})
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("unbound scope error = %v", err)
	}
}

func TestReceiptFromStoredAcceptsReplayJobIdentity(t *testing.T) {
	prepared, err := prepare(SubmitCommand{
		ProposedJobID: uuid.MustParse(fixtureJobID), Plan: fixturePlan(),
	})
	if err != nil {
		t.Fatal(err)
	}
	otherJob := "018f4f9e-7c21-7a3b-8c4d-123456789abf"
	receipt, err := receiptFromStored(StoredReceipt{
		JobID: otherJob, Generation: 3,
		PlanSHA256: prepared.PlanSHA256(), Replay: true,
	}, prepared.PlanSHA256())
	if err != nil || receipt.JobID.UUID().String() != otherJob || !receipt.Replay {
		t.Fatalf("replay receipt = %+v, %v", receipt, err)
	}
	_, err = receiptFromStored(StoredReceipt{
		JobID: otherJob, Generation: 3,
		PlanSHA256: "00" + prepared.PlanSHA256()[2:], Replay: true,
	}, prepared.PlanSHA256())
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("tampered digest error = %v", err)
	}
}

type fixtureBinder struct{}

func (fixtureBinder) Bind(
	context.Context,
	platformtx.TenantScope,
) (Store, error) {
	return nil, ErrUnavailable
}
