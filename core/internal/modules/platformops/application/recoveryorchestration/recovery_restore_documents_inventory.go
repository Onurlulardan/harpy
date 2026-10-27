package recoveryorchestration

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	documentsapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api"
	"github.com/google/uuid"
)

func loadDocumentsRecoveryInventory(
	document []byte,
	environmentID, recoverySetID, tenantID string,
	service documentsapi.Recovery,
) (documentsRecoveryInventory, []documentsapi.RecoveryInventoryItem, error) {
	var value documentsRecoveryInventory
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&value) != nil || decoder.Decode(new(any)) != io.EOF ||
		value.Schema != "documents-recovery-inventory.v1" ||
		value.EnvironmentID != environmentID ||
		value.RecoverySetID != recoverySetID || value.TenantID != tenantID {
		return value, nil, fmt.Errorf("documents recovery inventory is invalid")
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(
		document, maximumDocumentsRecoveryInventory,
	)
	expected, digestErr := decodeDocumentsDigest(value.InventorySHA256)
	if err != nil || !bytes.Equal(canonical, document) || digestErr != nil {
		return value, nil, fmt.Errorf("documents recovery inventory is invalid")
	}
	if service == nil {
		return value, nil, fmt.Errorf("documents recovery inventory is invalid")
	}
	items := make([]documentsapi.RecoveryInventoryItem, len(value.Entries))
	for index, entry := range value.Entries {
		item, itemErr := documentsInventoryItem(
			entry, environmentID, recoverySetID, tenantID,
		)
		if itemErr != nil || index > 0 &&
			value.Entries[index-1].VersionID >= entry.VersionID {
			return value, nil, fmt.Errorf("documents recovery inventory is invalid")
		}
		items[index] = item
	}
	normalized, err := service.ValidateInventory(items, expected)
	if err != nil || len(normalized) != len(items) {
		return value, nil, fmt.Errorf("documents recovery inventory is invalid")
	}
	return value, normalized, nil
}

func documentsInventoryItem(
	entry documentsRecoveryCopy,
	environmentID, recoverySetID, tenantID string,
) (documentsapi.RecoveryInventoryItem, error) {
	versionID, versionErr := uuid.Parse(entry.VersionID)
	digest, digestErr := decodeDocumentsDigest(entry.CiphertextSHA256)
	expectedTarget := documentsRecoveryObjectKey(
		environmentID, recoverySetID, tenantID, entry.VersionID,
	)
	if versionErr != nil || versionID.String() != entry.VersionID ||
		entry.SourceObjectKey == "" || entry.SourceObjectVersionID == "" ||
		entry.SourceObjectGeneration < 1 || entry.TargetObjectKey != expectedTarget ||
		entry.TargetObjectVersionID == "" || entry.CiphertextBytes < 1 ||
		entry.KeyID == "" || digestErr != nil {
		return documentsapi.RecoveryInventoryItem{},
			fmt.Errorf("documents recovery inventory item is invalid")
	}
	return documentsapi.RecoveryInventoryItem{
		VersionID: versionID.String(), ObjectKey: entry.SourceObjectKey,
		ObjectVersion:    entry.SourceObjectVersionID,
		ObjectGeneration: entry.SourceObjectGeneration,
		CiphertextSHA256: digest, CiphertextBytes: entry.CiphertextBytes,
		KeyID: entry.KeyID,
	}, nil
}

func decodeDocumentsDigest(value string) ([32]byte, error) {
	var result [32]byte
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != len(result) {
		return result, fmt.Errorf("documents digest is invalid")
	}
	copy(result[:], decoded)
	return result, nil
}
