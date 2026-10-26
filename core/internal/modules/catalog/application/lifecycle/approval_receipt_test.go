package lifecycle

import (
	"testing"
	"time"
)

func TestApprovalReceiptRejectsIdentityTimeAndConsumptionDrift(t *testing.T) {
	receipt := approvalReceiptFixture()
	if !receipt.Valid() {
		t.Fatal("valid approval receipt rejected")
	}
	for name, mutate := range map[string]func(*PurgeApprovalReceipt){
		"same actor": func(value *PurgeApprovalReceipt) {
			value.ApprovedBy = value.RequestedBy
		},
		"foreign backup": func(value *PurgeApprovalReceipt) {
			value.BackupReceiptID = "019fa000-0000-7000-8000-000000000099"
		},
		"stale step-up": func(value *PurgeApprovalReceipt) {
			value.StepUpAuthenticatedAt = value.ApprovedAt.Add(-10*time.Minute - time.Microsecond)
		},
	} {
		t.Run(name, func(t *testing.T) {
			tampered := receipt
			mutate(&tampered)
			tampered.ApprovalSHA256 = approvalReceiptSHA256(tampered)
			if tampered.Valid() {
				t.Fatal("tampered approval receipt accepted")
			}
		})
	}
	consumed := receipt
	consumed.State = purgeApprovalConsumed
	consumed.ConsumedOperationID = receipt.OperationID
	consumed.ConsumedAuthoritySHA256 =
		"cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	consumed.ConsumedAt = receipt.ApprovedAt.Add(time.Minute)
	if !consumed.Valid() {
		t.Fatal("exact consumed receipt rejected")
	}
	consumed.ConsumedOperationID = "019fa000-0000-7000-8000-000000000098"
	if !consumed.Valid() {
		t.Fatal("receipt shape should remain valid before binding comparison")
	}
}

func TestApprovalReceiptAcceptsStableTenantAndFirstPartyPackageIDs(t *testing.T) {
	receipt := approvalReceiptFixture()
	if !receipt.Valid() {
		t.Fatal("stable tenant and first-party package IDs were rejected")
	}
	for name, mutate := range map[string]func(*PurgeApprovalReceipt){
		"nil tenant": func(value *PurgeApprovalReceipt) {
			value.TenantID = "00000000-0000-0000-0000-000000000000"
		},
		"nil package": func(value *PurgeApprovalReceipt) {
			value.PackageID = "00000000-0000-0000-0000-000000000000"
		},
		"non-canonical tenant": func(value *PurgeApprovalReceipt) {
			value.TenantID = "A6000000-0000-4000-8000-000000000071"
		},
		"non-canonical package": func(value *PurgeApprovalReceipt) {
			value.PackageID = "E3000000-0000-0000-0000-000000000001"
		},
	} {
		t.Run(name, func(t *testing.T) {
			tampered := receipt
			refreshApprovalDigests(t, &tampered)
			mutate(&tampered)
			refreshApprovalDigests(t, &tampered)
			if tampered.Valid() {
				t.Fatal("invalid stable identity was accepted")
			}
		})
	}
}

func approvalReceiptFixture() PurgeApprovalReceipt {
	approvedAt := time.Date(2026, 7, 31, 12, 0, 0, 123000, time.UTC)
	value := PurgeApprovalReceipt{
		ID:               "019fa000-0000-7000-8000-000000000001",
		TenantID:         "76000000-0000-4000-8000-000000000071",
		InstallationID:   "019fa000-0000-7000-8000-000000000003",
		PackageID:        "e3000000-0000-0000-0000-000000000001",
		OperationID:      "019fa000-0000-7000-8000-000000000005",
		RequestedBy:      "019fa000-0000-7000-8000-000000000006",
		ApprovedBy:       "019fa000-0000-7000-8000-000000000007",
		BackupReceiptID:  "019fa000-0000-7000-8000-000000000008",
		BackupSHA256:     "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		ConfirmationSlug: "human-resources", Reason: "verified tenant removal",
		PurgeIdempotencyKey:    "catalog-purge-key-0001",
		ApprovalIdempotencyKey: "catalog-approval-key-0001",
		StepUpAuthenticatedAt:  approvedAt.Add(-time.Minute),
		ApprovedAt:             approvedAt, ExpiresAt: approvedAt.Add(30 * time.Minute),
		State: purgeApprovalAvailable,
	}
	value.RequestSHA256, _ = approvalRequestDocumentSHA(purgeApprovalRequestDocument{
		TenantID: value.TenantID, InstallationID: value.InstallationID,
		PackageID: value.PackageID, OperationID: value.OperationID,
		BackupReceiptID: value.BackupReceiptID, RequestedBy: value.RequestedBy,
		Reason: value.Reason, ConfirmationSlug: value.ConfirmationSlug,
		PurgeIdempotencyKey: value.PurgeIdempotencyKey,
	})
	value.ApprovalSHA256 = approvalReceiptSHA256(value)
	return value
}

func refreshApprovalDigests(t *testing.T, value *PurgeApprovalReceipt) {
	t.Helper()
	var err error
	value.RequestSHA256, err = approvalRequestDocumentSHA(purgeApprovalRequestDocument{
		TenantID: value.TenantID, InstallationID: value.InstallationID,
		PackageID: value.PackageID, OperationID: value.OperationID,
		BackupReceiptID: value.BackupReceiptID, RequestedBy: value.RequestedBy,
		Reason: value.Reason, ConfirmationSlug: value.ConfirmationSlug,
		PurgeIdempotencyKey: value.PurgeIdempotencyKey,
	})
	if err != nil {
		t.Fatal(err)
	}
	value.ApprovalSHA256 = approvalReceiptSHA256(*value)
}
