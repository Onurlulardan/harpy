package tenancy

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestTenantAuthoritySynchronizerAcceptsCompleteReplay(t *testing.T) {
	synchronizer, err := newTenantAuthoritySynchronizer(func(
		context.Context, uuid.UUID, string,
	) (TenantAuthorityReconciliation, error) {
		return TenantAuthorityReconciliation{
			BillingGlobalReconciled: 1, BillingTenantReconciled: 2,
			WorkflowReconciled: 3, AnalyticsReconciled: 4,
		}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := synchronizer.Synchronize(
		context.Background(), authorityOperationID, authorityManifestSHA256,
	)
	if err != nil || result.BillingGlobalReconciled != 1 ||
		result.BillingTenantReconciled != 2 ||
		result.WorkflowReconciled != 3 || result.AnalyticsReconciled != 4 {
		t.Fatalf("reconciliation = %#v, %v", result, err)
	}
}

func TestTenantAuthoritySynchronizerRejectsIncompleteAndStorageFailure(t *testing.T) {
	incomplete, _ := newTenantAuthoritySynchronizer(func(
		context.Context, uuid.UUID, string,
	) (TenantAuthorityReconciliation, error) {
		return TenantAuthorityReconciliation{WorkflowRemaining: 1}, nil
	})
	if _, err := incomplete.Synchronize(
		context.Background(), authorityOperationID, authorityManifestSHA256,
	); !errors.Is(
		err, errTenantAuthorityReconciliationIncomplete,
	) {
		t.Fatalf("incomplete error = %v", err)
	}
	billingIncomplete, _ := newTenantAuthoritySynchronizer(func(
		context.Context, uuid.UUID, string,
	) (TenantAuthorityReconciliation, error) {
		return TenantAuthorityReconciliation{BillingTenantRemaining: 1}, nil
	})
	if _, err := billingIncomplete.Synchronize(
		context.Background(), authorityOperationID, authorityManifestSHA256,
	); !errors.Is(
		err, errTenantAuthorityReconciliationIncomplete,
	) {
		t.Fatalf("Billing incomplete error = %v", err)
	}
	storageError := errors.New("database failed")
	failed, _ := newTenantAuthoritySynchronizer(func(
		context.Context, uuid.UUID, string,
	) (TenantAuthorityReconciliation, error) {
		return TenantAuthorityReconciliation{}, storageError
	})
	if _, err := failed.Synchronize(
		context.Background(), authorityOperationID, authorityManifestSHA256,
	); !errors.Is(
		err, storageError,
	) {
		t.Fatalf("storage error = %v", err)
	}
}

var (
	authorityOperationID    = uuid.MustParse("019b0208-0b4d-7c25-9a61-8021b57b3d25")
	authorityManifestSHA256 = strings.Repeat("a", 64)
)
