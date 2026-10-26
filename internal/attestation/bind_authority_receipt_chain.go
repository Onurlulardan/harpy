package attestation

import (
	"fmt"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
)

func bindAuthorityReceiptCheckpoint(
	receipt cutoverv1.AuthorityAdoptionReceipt,
	operation cutoverv1.AuthorityAdoptionOperation,
) error {
	checkpoint := receipt.PreviousCatalogCheckpoint
	if checkpoint == nil || operation.PreviousAuthorityReceiptSHA256 == nil ||
		receipt.PreviousAuthorityReceiptSHA256 == nil ||
		checkpoint.ReceiptSHA256 != *operation.PreviousAuthorityReceiptSHA256 ||
		checkpoint.ReceiptSHA256 != *receipt.PreviousAuthorityReceiptSHA256 {
		return fmt.Errorf("authority adoption signed prior checkpoint is required")
	}
	priorContract := cutoverv1.AuthorityReceiptCatalogContractSHA256(checkpoint.Objects)
	if checkpoint.EnvironmentID != operation.EnvironmentID || checkpoint.Target != operation.Target ||
		checkpoint.CatalogGeneration != operation.PreviousCatalogGeneration ||
		checkpoint.CatalogContractSHA256 != priorContract ||
		operation.PreCatalogContractSHA256 != priorContract ||
		receipt.PreCatalogContractSHA256 != priorContract ||
		receipt.PreAdoptionOwnedObjectCount != len(checkpoint.Objects) {
		return fmt.Errorf("authority adoption prior catalog differs")
	}
	if !adoptedAfter(receipt.AdoptedAt, checkpoint.AdoptedAt) {
		return fmt.Errorf("authority adoption receipt chain time differs")
	}
	if operation.Mode == "reaffirmation" && len(receipt.Objects) != len(checkpoint.Objects) {
		return fmt.Errorf("authority reaffirmation changed catalog size")
	}
	if operation.Mode == "catalog_upgrade" && len(receipt.Objects) <= len(checkpoint.Objects) {
		return fmt.Errorf("authority catalog upgrade did not append objects")
	}
	for index, priorObject := range checkpoint.Objects {
		if index >= len(receipt.Objects) || receipt.Objects[index] != priorObject ||
			!retainedOperationMatches(operation.Objects[index], priorObject) {
			return fmt.Errorf("authority adoption retained catalog differs")
		}
	}
	return nil
}

func retainedOperationMatches(
	operation cutoverv1.AuthorityAdoptionObject,
	prior cutoverv1.AuthorityAdoptionReceiptObject,
) bool {
	return operation.Kind == prior.Kind && operation.SchemaName == prior.SchemaName &&
		operation.ObjectName == prior.ObjectName &&
		operation.IdentityArguments == prior.IdentityArguments &&
		operation.IntroducedGeneration == prior.IntroducedGeneration &&
		operation.Action == "retain" &&
		operation.PreAdoptionOwner == prior.Owner && operation.AdoptedOwner == prior.Owner &&
		operation.PreDefinitionSHA256 == prior.DefinitionSHA256 &&
		operation.AdoptedDefinitionSHA256 == prior.DefinitionSHA256 &&
		operation.PreAdoptionACLSHA256 == prior.ACLSHA256 &&
		operation.AdoptedACLSHA256 == prior.ACLSHA256
}

func adoptedAfter(current, previous string) bool {
	currentTime, currentErr := time.Parse(time.RFC3339Nano, current)
	previousTime, previousErr := time.Parse(time.RFC3339Nano, previous)
	return currentErr == nil && previousErr == nil && currentTime.After(previousTime)
}
