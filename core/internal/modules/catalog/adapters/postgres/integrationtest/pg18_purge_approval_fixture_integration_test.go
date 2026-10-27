//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"testing"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	catalogcomponentv2 "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v2"
	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
	"github.com/google/uuid"
)

const approvalPackageID = "e3000000-0000-0000-0000-000000000001"
const approvalComponentID = "019fa100-0000-7000-8000-000000000002"
const approvalInstallationID = "019fa100-0000-7000-8000-000000000003"
const approvalOperationID = "019fa100-0000-7000-8000-000000000004"
const approvalReceiptID = "019fa100-0000-7000-8000-000000000005"
const approvalBackupID = "019fa100-0000-7000-8000-000000000006"
const approvalRequesterID = "019fa100-0000-7000-8000-000000000007"

func catalogApprovalReceipt(t *testing.T, approvedAt time.Time) catalogapp.PurgeApprovalReceipt {
	t.Helper()
	value := catalogapp.PurgeApprovalReceipt{
		ID: approvalReceiptID, TenantID: catalogTenantA,
		InstallationID: approvalInstallationID, PackageID: approvalPackageID,
		OperationID: approvalOperationID, RequestedBy: approvalRequesterID,
		ApprovedBy: catalogActorID, BackupReceiptID: approvalBackupID,
		BackupSHA256: catalogManifestDigest, ConfirmationSlug: "approval-proof",
		Reason: "verified destructive cleanup", PurgeIdempotencyKey: "purge-proof-0001",
		ApprovalIdempotencyKey: "approval-proof-0001",
		StepUpAuthenticatedAt:  approvedAt.Add(-time.Minute),
		ApprovedAt:             approvedAt, ExpiresAt: approvedAt.Add(30 * time.Minute),
		State: "AVAILABLE",
	}
	document := struct {
		TenantID, InstallationID, PackageID, OperationID, BackupReceiptID string
		RequestedBy, Reason, ConfirmationSlug, PurgeIdempotencyKey        string
	}{
		value.TenantID, value.InstallationID, value.PackageID, value.OperationID,
		value.BackupReceiptID, value.RequestedBy, value.Reason,
		value.ConfirmationSlug, value.PurgeIdempotencyKey,
	}
	raw, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(raw, 16384)
	if err != nil {
		t.Fatal(err)
	}
	requestDigest := sha256.Sum256(canonical)
	value.RequestSHA256 = hex.EncodeToString(requestDigest[:])
	value.ApprovalSHA256 = approvalDigest(
		"catalog.purge-approval-receipt.v2", value.ID, value.TenantID,
		value.InstallationID, value.PackageID, value.OperationID,
		value.RequestedBy, value.ApprovedBy, value.BackupReceiptID,
		value.BackupSHA256, value.ConfirmationSlug, value.Reason,
		value.PurgeIdempotencyKey, value.RequestSHA256,
		value.ApprovalIdempotencyKey, approvalTime(value.StepUpAuthenticatedAt),
		approvalTime(value.ApprovedAt), approvalTime(value.ExpiresAt),
	)
	if !value.Valid() {
		t.Fatal("purge approval fixture is invalid")
	}
	return value
}

func approvalDigest(values ...string) string {
	hash := sha256.New()
	var size [4]byte
	for _, value := range values {
		binary.BigEndian.PutUint32(size[:], uint32(len(value)))
		_, _ = hash.Write(size[:])
		_, _ = hash.Write([]byte(value))
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func approvalTime(value time.Time) string {
	return value.UTC().Format("2006-01-02T15:04:05.000000Z07:00")
}

func catalogPurgeMutation(
	t *testing.T, receipt catalogapp.PurgeApprovalReceipt, backupID, operationID string,
	at time.Time, mutate bool,
) catalogapp.StartMutation {
	t.Helper()
	packageID, _ := packagecontract.ParsePackageID(approvalPackageID)
	componentID, _ := packagecontract.ParseComponentID(approvalComponentID)
	version, _ := packagecontract.ParseVersion("1.0.0")
	plan, err := catalogdomain.NewSealedPlan(
		packageID, version, "approval-proof", catalogManifestDigest, nil,
		[]catalogdomain.ComponentPin{{
			ComponentID: componentID, Owner: "catalog", OwnerGeneration: 7,
			Contract: "catalog-approval.v2", ContentSHA256: catalogManifestDigest,
		}}, nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	stableID, _ := packagecontract.ParseStableID(approvalInstallationID)
	installation, _ := catalogdomain.NewInstalling(stableID, packageID)
	installation, _ = installation.Activate(version)
	installation, _ = installation.Retire()
	retired := installation.State
	installation, _ = installation.BeginPurge()
	operationStable, _ := packagecontract.ParseStableID(operationID)
	operation, err := catalogdomain.NewOperation(
		operationStable, catalogdomain.OperationPurge, plan, at,
	)
	if err != nil {
		t.Fatal(err)
	}
	authority := catalogapp.PurgeAuthority{
		TenantID: catalogTenantA, PackageID: approvalPackageID,
		PackageSlug: "approval-proof", OperationID: operationID,
		ActorID: approvalRequesterID, RequestedBy: approvalRequesterID,
		ApprovedBy: receipt.ApprovedBy, ApprovalReceiptID: receipt.ID,
		ApprovalSHA256: receipt.ApprovalSHA256, BackupReceiptID: backupID,
		BackupState: catalogcomponentv2.BackupReady, BackupSHA256: catalogManifestDigest,
		ConfirmationSlug: "approval-proof", Reason: receipt.Reason,
		IdempotencyKey: "purge-proof-0001", ReferenceEpoch: installation.ReferenceEpoch,
		AuthorizationTokenVersion: 3, AuthorizationVersion: 5,
		AuthorizationClock:          7,
		AuthorizationPermissionID:   "019f3000-0000-7000-8000-000000000015",
		AuthorizationEvidenceSHA256: catalogManifestDigest,
		StepUpPurposeSHA256:         catalogManifestDigest,
		StepUpAuthenticatedAt:       at.Add(-time.Minute),
		ApprovalStepUpAt:            receipt.StepUpAuthenticatedAt,
		ApprovedAt:                  receipt.ApprovedAt, ApprovalExpiresAt: receipt.ExpiresAt,
		RequestedAt: at, CoolingUntil: at.Add(24 * time.Hour),
	}
	authority.AuthoritySHA256 = catalogcomponentv2.AuthoritySHA256(authority)
	return catalogapp.StartMutation{
		Installation: installation, Operation: operation, PreviousState: &retired,
		ActorID: uuid.MustParse(approvalRequesterID), IdempotencyKey: authority.IdempotencyKey,
		RequestSHA256: authority.AuthoritySHA256, PurgeAuthority: &authority,
		OccurredAt: at, MutateInstallation: mutate,
	}
}

func catalogappAuthoritySHA(value catalogapp.PurgeAuthority) string {
	return catalogcomponentv2.AuthoritySHA256(value)
}

func assertApprovalState(t *testing.T, ctx context.Context,
	database catalogPG18, installationState, approvalState string,
) {
	t.Helper()
	var gotInstallation, gotApproval string
	if err := database.owner.QueryRowContext(ctx,
		`SELECT state FROM catalog_a.catalog_installations WHERE installation_id=$1`,
		approvalInstallationID,
	).Scan(&gotInstallation); err != nil {
		t.Fatal(err)
	}
	if err := database.owner.QueryRowContext(ctx,
		`SELECT state FROM catalog_a.catalog_purge_approval_receipts_v2 WHERE receipt_id=$1`,
		approvalReceiptID,
	).Scan(&gotApproval); err != nil {
		t.Fatal(err)
	}
	if gotInstallation != installationState || gotApproval != approvalState {
		t.Fatalf("installation/approval state=%s/%s, want %s/%s",
			gotInstallation, gotApproval, installationState, approvalState)
	}
}

func assertApprovalReceiptImmutable(
	t *testing.T, ctx context.Context, database catalogPG18,
) {
	t.Helper()
	if _, err := database.owner.ExecContext(ctx,
		`UPDATE catalog_a.catalog_purge_approval_receipts_v2
SET reason=reason WHERE receipt_id=$1`, approvalReceiptID,
	); err == nil {
		t.Fatal("purge approval update was accepted")
	}
	if _, err := database.owner.ExecContext(ctx,
		`DELETE FROM catalog_a.catalog_purge_approval_receipts_v2 WHERE receipt_id=$1`,
		approvalReceiptID,
	); err == nil {
		t.Fatal("purge approval delete was accepted")
	}
}
