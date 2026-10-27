package lifecycle

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func TestPurgeRejectsMissingEvidence(t *testing.T) {
	_, err := (&Application{}).Purge(context.Background(), platformtx.TenantScope{}, PurgeRequest{})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("purge error = %v", err)
	}
}

func TestPurgeAcceptsStableFirstPartyPackageIDAtBoundary(t *testing.T) {
	request := PurgeRequest{
		OperationID:       uuid.MustParse("019fa000-0000-7000-8000-000000000001"),
		PackageID:         uuid.MustParse("e3000000-0000-0000-0000-000000000001"),
		BackupReceiptID:   uuid.MustParse("019fa000-0000-7000-8000-000000000002"),
		ApprovalReceiptID: uuid.MustParse("019fa000-0000-7000-8000-000000000003"),
		RequestedBy:       uuid.MustParse("019fa000-0000-7000-8000-000000000004"),
		StepUpAtUnix:      1,
		Reason:            "verified tenant removal",
		ConfirmationSlug:  "human-resources",
		IdempotencyKey:    "catalog-purge-stable-1",
	}
	_, err := (&Application{}).Purge(
		context.Background(), platformtx.TenantScope{}, request,
	)
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("stable first-party package boundary error = %v", err)
	}
	request.PackageID = uuid.Nil
	if _, err = (&Application{}).Purge(
		context.Background(), platformtx.TenantScope{}, request,
	); !errors.Is(err, ErrInvalid) {
		t.Fatalf("nil package boundary error = %v", err)
	}
}
