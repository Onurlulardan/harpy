package recoveryset

import (
	"testing"
	"time"
)

func TestValidateRestoreRejectsCrossModePortableAndAuthorityMutation(t *testing.T) {
	request := restoreFixture()
	if err := ValidateRestore(request); err != nil {
		t.Fatal(err)
	}
	cases := []func(*RestoreRequest){
		func(value *RestoreRequest) { value.Manifest.Portability = Portable },
		func(value *RestoreRequest) { value.Target.DeploymentMode = ModeSaaS },
		func(value *RestoreRequest) { value.Target.OwnerBindingID = uuidFixture(9) },
		func(value *RestoreRequest) { value.Target.TenantSetSHA256 = digestFixture("9") },
		func(value *RestoreRequest) { value.SeedRequested = true },
		func(value *RestoreRequest) { value.OwnerChange = true },
		func(value *RestoreRequest) { value.TenantReplacement = true },
	}
	for index, mutate := range cases {
		candidate := restoreFixture()
		mutate(&candidate)
		if ValidateRestore(candidate) == nil {
			t.Fatalf("invalid restore matrix case %d was accepted", index)
		}
	}
}

func TestValidateRestoreRequiresExactDispositionEvidence(t *testing.T) {
	empty := restoreFixture()
	empty.Disposition = TargetEmpty
	empty.Safety = SafetyEvidence{WaiverReason: "empty_no_database_or_policy",
		ProbeSHA256: digestFixture("8")}
	if err := ValidateRestore(empty); err != nil {
		t.Fatal(err)
	}
	unreadable := empty
	unreadable.Disposition = TargetUnreadable
	unreadable.Safety.WaiverReason = "unreadable_integrity_probe"
	unreadable.Safety.DestructiveConfirmed = true
	if err := ValidateRestore(unreadable); err != nil {
		t.Fatal(err)
	}
	unreadable.Safety.DestructiveConfirmed = false
	if ValidateRestore(unreadable) == nil {
		t.Fatal("unreadable target without destructive confirmation was accepted")
	}
}

func restoreFixture() RestoreRequest {
	identity := EnvironmentIdentity{EnvironmentID: uuidFixture(1), DeploymentMode: ModeOnPremise,
		PrimaryBindingID: uuidFixture(2), TenantSetSHA256: digestFixture("a"),
		OwnerBindingID: uuidFixture(3), OwnerBindingGeneration: 4,
		DatabaseLineageID: uuidFixture(4)}
	manifest := ManifestPayload{Schema: ManifestSchema, Portability: SameEnvironment,
		Identity: identity, DatabaseSystemID: uuidFixture(5),
		Components:            recoveryComponentsFixture(),
		DatabaseArchiveSHA256: digestFixture("b"), DocumentsMerkleSHA256: digestFixture("c"),
		KeyCustodySHA256: digestFixture("d"), OwnerAnchorSHA256: digestFixture("e"),
		ProtectedHeadAckSHA256: digestFixture("f"), RecoveryTrustBundleSHA256: digestFixture("1"),
		RecoverySigningGeneration: 1, CreatedAt: time.Date(2026, 7, 23, 2, 0, 0, 0, time.UTC)}
	return RestoreRequest{RecoverySetID: uuidFixture(7), Manifest: manifest, Target: identity,
		Disposition: TargetExistingReadable,
		Safety:      SafetyEvidence{RecoverySetID: uuidFixture(6), RecoverySetSHA256: digestFixture("2")}}
}

func recoveryComponentsFixture() RecoveryComponents {
	component := func(name, digest string) ComponentIdentity {
		return ComponentIdentity{ObjectKey: "objects/" + name,
			ObjectVersionID:  "version-" + name,
			CiphertextSHA256: digest, CiphertextBytes: 1}
	}
	return RecoveryComponents{
		DatabaseArchive:    component("database", digestFixture("b")),
		DocumentsInventory: component("documents", digestFixture("7")),
		KeyCustody:         component("custody", digestFixture("d")),
		OwnerAnchor:        component("anchor", digestFixture("e")),
		ProtectedAuthority: component("authority", digestFixture("f")),
	}
}

func uuidFixture(marker int) string {
	return "018f0000-0000-7000-8000-00000000000" + string(rune('0'+marker))
}

func digestFixture(marker string) string {
	result := ""
	for len(result) < 64 {
		result += marker
	}
	return result
}
