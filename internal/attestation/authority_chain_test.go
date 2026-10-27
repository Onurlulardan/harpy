package attestation

import (
	"strings"
	"testing"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
)

func TestAuthorityReceiptChainAcceptsGenerationOneThroughThree(t *testing.T) {
	first := authorityInitialReceiptFixture()
	secondOperation, secondReceipt, release, authority := authorityChainUpgradeFixture(first, 2)
	if err := bindAuthorityReceipt(secondReceipt, secondOperation, release, authority); err != nil {
		t.Fatalf("generation two rejected: %v", err)
	}
	second := newVerifiedAuthorityAdoptionReceipt(
		secondReceipt, strings.Repeat("a", 64), []byte("two"),
	)
	thirdOperation, thirdReceipt, release, authority := authorityChainUpgradeFixture(second, 3)
	if err := bindAuthorityReceipt(thirdReceipt, thirdOperation, release, authority); err != nil {
		t.Fatalf("generation three rejected: %v", err)
	}
}

func TestAuthorityReceiptChainRejectsOrphanAndWrongPrevious(t *testing.T) {
	first := authorityInitialReceiptFixture()
	operation, receipt, release, authority := authorityChainUpgradeFixture(first, 2)
	receipt.PreviousCatalogCheckpoint = nil
	if err := bindAuthorityReceipt(receipt, operation, release, authority); err == nil {
		t.Fatal("orphan non-initial receipt accepted")
	}
	_, receipt, _, _ = authorityChainUpgradeFixture(first, 2)
	receipt.PreviousCatalogCheckpoint.ReceiptSHA256 = strings.Repeat("f", 64)
	if err := bindAuthorityReceipt(receipt, operation, release, authority); err == nil {
		t.Fatal("wrong prior receipt accepted")
	}
}

func TestAuthorityReceiptMustBeInsideOperationWindow(t *testing.T) {
	first := authorityInitialReceiptFixture()
	operation, receipt, release, authority := authorityChainUpgradeFixture(first, 2)
	receipt.AdoptedAt = "2026-07-14T10:10:00Z"
	if err := bindAuthorityReceipt(receipt, operation, release, authority); err == nil {
		t.Fatal("receipt at exclusive operation expiry accepted")
	}
}

func TestAuthorityReceiptChainRejectsRetainedOIDAndHashTamper(t *testing.T) {
	first := authorityInitialReceiptFixture()
	operation, receipt, release, authority := authorityChainUpgradeFixture(first, 2)
	receipt.Objects[0].ObjectOID++
	if err := bindAuthorityReceipt(receipt, operation, release, authority); err == nil {
		t.Fatal("retained object OID tamper accepted")
	}
	_, receipt, _, _ = authorityChainUpgradeFixture(first, 2)
	receipt.Objects[0].DefinitionSHA256 = strings.Repeat("e", 64)
	if err := bindAuthorityReceipt(receipt, operation, release, authority); err == nil {
		t.Fatal("retained object hash tamper accepted")
	}
}

func TestAuthorityReceiptRejectsSelfConsistentDependencyDrift(t *testing.T) {
	first := authorityInitialReceiptFixture()
	operation, receipt, release, authority := authorityChainUpgradeFixture(first, 2)
	receipt.Dependencies[0].DefinitionSHA256 = strings.Repeat("d", 64)
	receipt.DependencyContractSHA256 =
		cutoverv1.AuthorityReceiptDependencyContractSHA256(receipt.Dependencies)
	receipt.DependencyInstanceSHA256 =
		cutoverv1.AuthorityReceiptDependencyInstanceSHA256(receipt.Dependencies)
	if err := bindAuthorityReceipt(receipt, operation, release, authority); err == nil {
		t.Fatal("receipt dependency catalog drift from signed operation accepted")
	}
}

func TestAuthorityReceiptChainRejectsPriorGenerationOmission(t *testing.T) {
	first := authorityInitialReceiptFixture()
	secondOperation, secondReceipt, release, authority := authorityChainUpgradeFixture(first, 2)
	if err := bindAuthorityReceipt(secondReceipt, secondOperation, release, authority); err != nil {
		t.Fatal(err)
	}
	second := newVerifiedAuthorityAdoptionReceipt(secondReceipt, strings.Repeat("a", 64), nil)
	thirdOperation, thirdReceipt, release, authority := authorityChainUpgradeFixture(second, 3)
	operation := thirdOperation.Value()
	operation.Objects = append(operation.Objects[:1:1], operation.Objects[2:]...)
	operation.AuthorityOwnerRole.OwnedObjectCount = len(operation.Objects)
	operation.PostCatalogContractSHA256 = strings.Repeat("d", 64)
	thirdOperation = newVerifiedAuthorityAdoptionOperation(
		operation, thirdOperation.ManifestSHA256(), nil,
	)
	thirdReceipt.Objects = append(thirdReceipt.Objects[:1:1], thirdReceipt.Objects[2:]...)
	thirdReceipt.AuthorityOwnerRole.OwnedObjectCount = len(thirdReceipt.Objects)
	thirdReceipt.PostCatalogContractSHA256 = operation.PostCatalogContractSHA256
	thirdReceipt.CatalogInstanceSHA256 = strings.Repeat("c", 64)
	thirdReceipt.PreAdoptionOwnedObjectCount = 1
	if err := bindAuthorityReceipt(
		thirdReceipt, thirdOperation, release, authority,
	); err == nil {
		t.Fatal("prior generation omission accepted")
	}
}

func TestVerifiedAuthorityValuesAreMutationIsolated(t *testing.T) {
	previous, artifact := strings.Repeat("1", 64), strings.Repeat("2", 64)
	operation := authorityChainUpgradeOperationValue(previous, artifact)
	verifiedOperation := newVerifiedAuthorityAdoptionOperation(operation, "digest", []byte("op"))
	operation.Objects[0].ObjectName = "source_mutation"
	operation.QuiescencePolicy.Roles[0].Principal = "source_mutation"
	*operation.PreviousAuthorityReceiptSHA256 = strings.Repeat("3", 64)
	firstOperation := verifiedOperation.Value()
	firstOperation.Objects[0].ObjectName = "returned_mutation"
	firstOperation.QuiescencePolicy.Roles[0].Principal = "returned_mutation"
	*firstOperation.UpgradeArtifactSHA256 = strings.Repeat("4", 64)
	if verifiedOperation.Value().Objects[0].ObjectName != "catalog_v1" ||
		verifiedOperation.Value().QuiescencePolicy.Roles[0].Principal != "harpy_migrator" ||
		*verifiedOperation.Value().PreviousAuthorityReceiptSHA256 != previous ||
		*verifiedOperation.Value().UpgradeArtifactSHA256 != artifact {
		t.Fatal("verified authority operation leaked mutable state")
	}
	receipt := authorityInitialReceiptFixture().Value()
	verifiedReceipt := newVerifiedAuthorityAdoptionReceipt(receipt, "digest", []byte("receipt"))
	receipt.Objects[0].ObjectOID++
	receipt.QuiescenceEvidence.Roles[0].Principal = "source_mutation"
	returned := verifiedReceipt.Value()
	returned.Objects[0].ObjectOID++
	returned.QuiescenceEvidence.Roles[0].Principal = "returned_mutation"
	if verifiedReceipt.Value().Objects[0].ObjectOID != 17001 ||
		verifiedReceipt.Value().QuiescenceEvidence.Roles[0].Principal != "harpy_migrator" {
		t.Fatal("verified authority receipt leaked mutable state")
	}
	initial := authorityInitialReceiptFixture()
	_, upgrade, _, _ := authorityChainUpgradeFixture(initial, 2)
	verifiedUpgrade := newVerifiedAuthorityAdoptionReceipt(upgrade, "digest", nil)
	upgrade.PreviousCatalogCheckpoint.Objects[0].ObjectOID++
	returnedUpgrade := verifiedUpgrade.Value()
	returnedUpgrade.PreviousCatalogCheckpoint.Objects[0].ObjectOID++
	if verifiedUpgrade.Value().PreviousCatalogCheckpoint.Objects[0].ObjectOID != 17001 {
		t.Fatal("verified authority checkpoint leaked mutable state")
	}
}

func authorityChainUpgradeOperationValue(previous, artifact string) cutoverv1.AuthorityAdoptionOperation {
	operation := cutoverv1.AuthorityAdoptionOperation{
		PreviousAuthorityReceiptSHA256: &previous, UpgradeArtifactSHA256: &artifact,
		Objects: []cutoverv1.AuthorityAdoptionObject{{ObjectName: "catalog_v1"}},
	}
	operation.QuiescencePolicy = authorityQuiescencePolicyFixture(operation.Objects)
	return operation
}
