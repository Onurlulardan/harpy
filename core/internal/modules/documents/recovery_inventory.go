package documents

import (
	"fmt"

	documentsapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api"
	recoveryapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/recovery"
	"github.com/google/uuid"
)

func (facade *recoveryFacade) ValidateInventory(
	items []documentsapi.RecoveryInventoryItem,
	expected [32]byte,
) ([]documentsapi.RecoveryInventoryItem, error) {
	internal := make([]recoveryapp.InventoryItem, len(items))
	for index, item := range items {
		value, err := importRecoveryInventoryItem(item)
		if err != nil {
			return nil, err
		}
		internal[index] = value
	}
	normalized, err := recoveryapp.ValidateInventory(internal, expected)
	return exportRecoveryInventory(normalized), err
}

func exportRecoveryInventory(
	items []recoveryapp.InventoryItem,
) []documentsapi.RecoveryInventoryItem {
	result := make([]documentsapi.RecoveryInventoryItem, len(items))
	for index, item := range items {
		result[index] = documentsapi.RecoveryInventoryItem{
			VersionID: item.VersionID.String(), ObjectKey: item.ObjectKey,
			ObjectVersion:    item.ObjectVersion,
			ObjectGeneration: item.ObjectGeneration,
			CiphertextSHA256: item.CiphertextSHA256,
			CiphertextBytes:  item.CiphertextBytes, KeyID: item.KeyID,
		}
	}
	return result
}

func importRecoveryInventoryItem(
	item documentsapi.RecoveryInventoryItem,
) (recoveryapp.InventoryItem, error) {
	versionID, err := uuid.Parse(item.VersionID)
	if err != nil || versionID == uuid.Nil || versionID.String() != item.VersionID {
		return recoveryapp.InventoryItem{},
			fmt.Errorf("documents recovery inventory identity is invalid")
	}
	return recoveryapp.InventoryItem{
		VersionID: versionID, ObjectKey: item.ObjectKey,
		ObjectVersion:    item.ObjectVersion,
		ObjectGeneration: item.ObjectGeneration,
		CiphertextSHA256: item.CiphertextSHA256,
		CiphertextBytes:  item.CiphertextBytes, KeyID: item.KeyID,
	}, nil
}
