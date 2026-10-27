package catalogcomponentv2

import (
	"testing"
	"time"

	catalogcomponentv1 "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v1"
)

func TestPurgeAuthorityRejectsApprovalBackupAndCoolingDrift(t *testing.T) {
	authority := authorityFixture()
	if !authority.Valid() {
		t.Fatal("valid authority rejected")
	}
	for name, mutate := range map[string]func(*PurgeAuthority){
		"same approver": func(value *PurgeAuthority) { value.ApprovedBy = value.RequestedBy },
		"backup state":  func(value *PurgeAuthority) { value.BackupState = "COMPLETE" },
		"confirmation":  func(value *PurgeAuthority) { value.ConfirmationSlug = "other" },
		"nil tenant":    func(value *PurgeAuthority) { value.TenantID = "00000000-0000-0000-0000-000000000000" },
		"short cooling": func(value *PurgeAuthority) {
			value.CoolingUntil = value.RequestedAt.Add(24*time.Hour - time.Microsecond)
		},
	} {
		t.Run(name, func(t *testing.T) {
			tampered := authority
			mutate(&tampered)
			tampered.AuthoritySHA256 = AuthoritySHA256(tampered)
			if tampered.Valid() {
				t.Fatal("tampered authority accepted")
			}
		})
	}
}

func TestLifecycleAndPurgeAcceptOpaqueStableTenantAndPackageIDs(t *testing.T) {
	authority := authorityFixture()
	authority.TenantID = "76000000-0000-4000-8000-000000000071"
	authority.PackageID = "e3000000-0000-0000-0000-000000000001"
	authority.AuthoritySHA256 = AuthoritySHA256(authority)
	if !authority.Valid() {
		t.Fatal("opaque stable tenant and package IDs rejected")
	}
	command := LifecycleCommand{
		TenantID: authority.TenantID, PackageID: authority.PackageID,
		PackageSlug: authority.PackageSlug, OperationID: authority.OperationID,
		ActorID: authority.ActorID, Action: "retire", Reason: authority.Reason,
		ReferenceEpoch: authority.ReferenceEpoch, IdempotencyKey: "retire-owner-1",
		Component: catalogcomponentv1.Component{
			ID: "018f3000-0000-7000-8000-000000000008", Owner: "forms",
			Generation: 7, Contract: "forms.catalog.v2",
			ContentSHA256: "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
		},
	}
	if !command.Valid() {
		t.Fatal("opaque stable lifecycle tenant and package IDs rejected")
	}
}

func TestStableUUIDRejectsNilMalformedAndNonCanonicalValues(t *testing.T) {
	for _, value := range []string{
		"00000000-0000-0000-0000-000000000000",
		"not-a-uuid",
		"E3000000-0000-0000-0000-000000000001",
	} {
		if validStableUUID(value) {
			t.Fatalf("invalid stable UUID accepted: %q", value)
		}
	}
}

func TestOwnerResultBindsExactAuthorityAndTarget(t *testing.T) {
	command := PurgeCommand{
		Authority: authorityFixture(),
		Component: catalogcomponentv1.Component{
			ID: "018f3000-0000-7000-8000-000000000008", Owner: "forms",
			Generation: 7, Contract: "forms.catalog.v2",
			ContentSHA256: "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
		},
		IdempotencyKey: "purge-owner-1",
	}
	result := OwnerResult{
		ReceiptID: "018f3000-0000-7000-8000-000000000009",
		TenantID:  command.Authority.TenantID, PackageID: command.Authority.PackageID,
		Component: command.Component, OperationID: command.Authority.OperationID,
		ResourceID:     "018f3000-0000-7000-8000-000000000008",
		VersionID:      "018f3000-0000-7000-8000-000000000010",
		HeadID:         "018f3000-0000-7000-8000-000000000011",
		ReferenceEpoch: command.Authority.ReferenceEpoch, State: StatePurged,
		AuthoritySHA256: command.Authority.AuthoritySHA256, Ready: true,
		CompletedAt: command.Authority.CoolingUntil,
	}
	result.ResultSHA256 = ResultSHA256(result)
	if !result.ValidFor(command) {
		t.Fatal("exact owner result rejected")
	}
	result.HeadID = "018f3000-0000-7000-8000-000000000012"
	if result.ValidFor(command) {
		t.Fatal("forged publication binding accepted")
	}
}

func TestLifecycleResultBindsActionEpochAndHead(t *testing.T) {
	authority := authorityFixture()
	command := LifecycleCommand{
		TenantID: authority.TenantID, PackageID: authority.PackageID,
		PackageSlug: authority.PackageSlug, OperationID: authority.OperationID,
		ActorID: authority.ActorID, Action: "retire", Reason: authority.Reason,
		ReferenceEpoch: authority.ReferenceEpoch, IdempotencyKey: "retire-owner-1",
		Component: catalogcomponentv1.Component{
			ID: "018f3000-0000-7000-8000-000000000008", Owner: "forms",
			Generation: 7, Contract: "forms.catalog.v2",
			ContentSHA256: "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
		},
	}
	result := OwnerResult{
		ReceiptID: "018f3000-0000-7000-8000-000000000009",
		TenantID:  command.TenantID, PackageID: command.PackageID,
		Component: command.Component, OperationID: command.OperationID,
		ResourceID:     command.Component.ID,
		VersionID:      "018f3000-0000-7000-8000-000000000010",
		HeadID:         "018f3000-0000-7000-8000-000000000011",
		ReferenceEpoch: command.ReferenceEpoch, State: StateRetired, Ready: true,
		CompletedAt: authority.RequestedAt,
	}
	result.ResultSHA256 = ResultSHA256(result)
	if !result.ValidForLifecycle(command) {
		t.Fatal("exact lifecycle result rejected")
	}
	result.ReferenceEpoch++
	if result.ValidForLifecycle(command) {
		t.Fatal("stale lifecycle epoch accepted")
	}
}

func TestGenericOwnerResultDoesNotInventResourceHeadIDs(t *testing.T) {
	authority := authorityFixture()
	command := PurgeCommand{
		Authority: authority, IdempotencyKey: "catalog-owner-1",
		Component: catalogcomponentv1.Component{
			ID: "018f3000-0000-7000-8000-000000000008", Owner: "catalog",
			Generation: 2, Contract: "catalog.component.v2",
			ContentSHA256: "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
		},
	}
	result := OwnerResult{
		ReceiptID: "018f3000-0000-7000-8000-000000000009",
		TenantID:  authority.TenantID, PackageID: authority.PackageID,
		Component: command.Component, OperationID: authority.OperationID,
		ReferenceEpoch: authority.ReferenceEpoch, State: StatePurgePending,
		AuthoritySHA256: authority.AuthoritySHA256, Ready: true,
		CompletedAt: authority.CoolingUntil,
	}
	result.ResultSHA256 = ResultSHA256(result)
	if !result.ValidBeginFor(command) || result.HasResourceBinding() {
		t.Fatal("generic authority-bound owner result was not represented exactly")
	}
}

func authorityFixture() PurgeAuthority {
	requested := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	value := PurgeAuthority{
		TenantID:  "018f3000-0000-7000-8000-000000000001",
		PackageID: "018f3000-0000-7000-8000-000000000002", PackageSlug: "hr",
		OperationID:       "018f3000-0000-7000-8000-000000000003",
		ActorID:           "018f3000-0000-7000-8000-000000000004",
		RequestedBy:       "018f3000-0000-7000-8000-000000000004",
		ApprovedBy:        "018f3000-0000-7000-8000-000000000005",
		ApprovalReceiptID: "018f3000-0000-7000-8000-000000000007",
		ApprovalSHA256:    "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		BackupReceiptID:   "018f3000-0000-7000-8000-000000000006",
		BackupState:       BackupReady,
		BackupSHA256:      "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		ConfirmationSlug:  "hr", Reason: "tenant requested removal",
		IdempotencyKey:            "catalog-purge-1",
		ReferenceEpoch:            3,
		AuthorizationTokenVersion: 5, AuthorizationVersion: 7,
		AuthorizationClock:          11,
		AuthorizationPermissionID:   "018f3000-0000-7000-8000-000000000013",
		AuthorizationEvidenceSHA256: "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
		StepUpPurposeSHA256:         "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
		StepUpAuthenticatedAt:       requested.Add(-time.Minute),
		ApprovalStepUpAt:            requested.Add(-3 * time.Minute),
		ApprovedAt:                  requested.Add(-2 * time.Minute),
		ApprovalExpiresAt:           requested.Add(28 * time.Minute),
		RequestedAt:                 requested, CoolingUntil: requested.Add(24 * time.Hour),
	}
	value.AuthoritySHA256 = AuthoritySHA256(value)
	return value
}
