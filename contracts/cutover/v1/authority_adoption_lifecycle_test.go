package cutoverv1

import (
	"fmt"
	"strings"
	"testing"
)

func TestAuthorityCatalogUpgradeRetainsEveryPriorGeneration(t *testing.T) {
	initialOperation := validAuthorityAdoptionOperation()
	initialReceipt := validAuthorityAdoptionReceipt(initialOperation)
	secondOperation := authorityUpgradeOperation(initialReceipt, 2)
	secondReceipt := lifecycleReceipt(secondOperation, &initialReceipt)
	thirdOperation := authorityUpgradeOperation(secondReceipt, 3)
	thirdReceipt := lifecycleReceipt(thirdOperation, &secondReceipt)
	for _, value := range []interface{ Validate() error }{
		initialOperation, initialReceipt, secondOperation, secondReceipt,
		thirdOperation, thirdReceipt,
	} {
		if err := value.Validate(); err != nil {
			t.Fatalf("valid authority lifecycle rejected: %v", err)
		}
	}
	assertAuthoritySchema(t, "authority-adoption-operation.schema.json", thirdOperation)
	assertAuthoritySchema(t, "authority-adoption-receipt.schema.json", thirdReceipt)
	if thirdOperation.Objects[len(initialReceipt.Objects)].Action != "retain" {
		t.Fatal("generation two addition was not retained by generation three")
	}
	tampered := thirdOperation
	tampered.Objects = append([]AuthorityAdoptionObject(nil), thirdOperation.Objects...)
	tampered.Objects[len(initialReceipt.Objects)].Action = "add"
	tampered.PostCatalogContractSHA256 = AuthorityOperationCatalogContractSHA256(tampered.Objects)
	if err := tampered.Validate(); err == nil {
		t.Fatal("generation three re-add of generation two object accepted")
	}
}

func TestAuthorityCatalogDigestsRejectOrderedInventoryTamper(t *testing.T) {
	operation := validAuthorityAdoptionOperation()
	operation.Objects[0].AdoptedDefinitionSHA256 = strings.Repeat("c", 64)
	if err := operation.Validate(); err == nil {
		t.Fatal("operation catalog digest drift accepted")
	}
	receipt := validAuthorityAdoptionReceipt(validAuthorityAdoptionOperation())
	receipt.Objects[0].ObjectOID++
	if err := receipt.Validate(); err == nil {
		t.Fatal("receipt catalog instance digest drift accepted")
	}
}

func TestAuthorityUpgradeRejectsScopeDuplicateOrderAndSuffixDrift(t *testing.T) {
	initial := validAuthorityAdoptionReceipt(validAuthorityAdoptionOperation())
	valid := authorityUpgradeOperation(initial, 2)
	tests := map[string]func(*AuthorityAdoptionOperation){
		"scope": func(value *AuthorityAdoptionOperation) {
			value.Objects[len(value.Objects)-1].SchemaName = "platform"
		},
		"suffix": func(value *AuthorityAdoptionOperation) {
			value.Objects[len(value.Objects)-1].ObjectName = "catalog_probe_v3"
		},
		"empty function identity": func(value *AuthorityAdoptionOperation) {
			value.Objects[len(value.Objects)-1].IdentityArguments = ""
		},
		"duplicate": func(value *AuthorityAdoptionOperation) {
			value.Objects[len(value.Objects)-1] = value.Objects[0]
		},
		"order": func(value *AuthorityAdoptionOperation) {
			value.Objects[0], value.Objects[1] = value.Objects[1], value.Objects[0]
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			operation := valid
			operation.Objects = append([]AuthorityAdoptionObject(nil), valid.Objects...)
			mutate(&operation)
			operation.PostCatalogContractSHA256 = AuthorityOperationCatalogContractSHA256(operation.Objects)
			if err := operation.Validate(); err == nil {
				t.Fatal("authority upgrade object drift accepted")
			}
		})
	}
}

func authorityUpgradeOperation(
	previous AuthorityAdoptionReceipt, generation int,
) AuthorityAdoptionOperation {
	digest := strings.Repeat("d", 64)
	previousDigest := strings.Repeat(fmt.Sprintf("%x", generation), 64)[:64]
	objects := make([]AuthorityAdoptionObject, 0, len(previous.Objects)+1)
	for _, object := range previous.Objects {
		objects = append(objects, AuthorityAdoptionObject{
			Kind: object.Kind, SchemaName: object.SchemaName, ObjectName: object.ObjectName,
			IdentityArguments: object.IdentityArguments, Action: "retain",
			IntroducedGeneration: object.IntroducedGeneration,
			PreAdoptionOwner:     object.Owner, AdoptedOwner: object.Owner,
			PreDefinitionSHA256:     object.DefinitionSHA256,
			AdoptedDefinitionSHA256: object.DefinitionSHA256,
			PreAdoptionACLSHA256:    object.ACLSHA256, AdoptedACLSHA256: object.ACLSHA256,
		})
	}
	objects = append(objects, AuthorityAdoptionObject{
		Kind: "function", SchemaName: "cutover_authority",
		ObjectName: fmt.Sprintf("catalog_probe_v%d", generation), IdentityArguments: "uuid",
		Action: "add", IntroducedGeneration: generation,
		PreAdoptionOwner: authorityOwner, AdoptedOwner: authorityOwner,
		PreDefinitionSHA256: digest, AdoptedDefinitionSHA256: digest,
		PreAdoptionACLSHA256: digest, AdoptedACLSHA256: digest,
	})
	artifact := strings.Repeat("e", 64)
	operation := validAuthorityAdoptionOperation()
	operation.Mode = "catalog_upgrade"
	operation.OwnershipMode = "superuser_set_local_role_append_only"
	operation.PreviousCatalogGeneration = generation - 1
	operation.CatalogGeneration = generation
	operation.PreCatalogContractSHA256 = previous.PostCatalogContractSHA256
	operation.PreviousAuthorityReceiptSHA256 = &previousDigest
	operation.UpgradeArtifactSHA256 = &artifact
	operation.AuthorityOwnerRole.OwnedObjectCount = len(objects)
	operation.Objects = objects
	operation.PostCatalogContractSHA256 = AuthorityOperationCatalogContractSHA256(objects)
	operation.QuiescencePolicy = validAuthorityQuiescencePolicy(objects)
	return operation
}

func lifecycleReceipt(
	operation AuthorityAdoptionOperation, previous *AuthorityAdoptionReceipt,
) AuthorityAdoptionReceipt {
	receipt := validAuthorityAdoptionReceipt(operation)
	receipt.AdoptedAt = fmt.Sprintf("2026-07-14T10:0%d:00Z", operation.CatalogGeneration)
	receipt.PreviousAuthorityReceiptSHA256 = cloneDigest(operation.PreviousAuthorityReceiptSHA256)
	receipt.UpgradeArtifactSHA256 = cloneDigest(operation.UpgradeArtifactSHA256)
	receipt.PreAdoptionOwnedObjectCount = len(operation.Objects) - 1
	receipt.AuthorityOwnerRole.OwnedObjectCount = len(operation.Objects)
	for index := range receipt.Objects {
		receipt.Objects[index].IntroducedGeneration = operation.Objects[index].IntroducedGeneration
	}
	receipt.PostCatalogContractSHA256 = AuthorityReceiptCatalogContractSHA256(receipt.Objects)
	receipt.CatalogInstanceSHA256 = AuthorityReceiptCatalogInstanceSHA256(receipt.Objects)
	if previous != nil {
		receipt.PreviousCatalogCheckpoint = authorityCheckpoint(
			*previous, *operation.PreviousAuthorityReceiptSHA256,
		)
	}
	return receipt
}

func authorityCheckpoint(
	previous AuthorityAdoptionReceipt, receiptSHA256 string,
) *AuthorityCatalogCheckpoint {
	return &AuthorityCatalogCheckpoint{
		ReceiptSHA256: receiptSHA256, ReceiptID: previous.ReceiptID,
		EnvironmentID: previous.EnvironmentID, Target: previous.Target,
		CatalogGeneration: previous.CatalogGeneration, AdoptedAt: previous.AdoptedAt,
		CatalogContractSHA256: previous.PostCatalogContractSHA256,
		CatalogInstanceSHA256: previous.CatalogInstanceSHA256,
		Objects:               append([]AuthorityAdoptionReceiptObject(nil), previous.Objects...),
	}
}

func cloneDigest(value *string) *string {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}
