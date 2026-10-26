package recoveryorchestration

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	documentsapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

const maximumDocumentsRecoveryInventory = 1024 * 1024

type documentsRecoveryInventory struct {
	Schema          string                  `json:"schema"`
	EnvironmentID   string                  `json:"environment_id"`
	RecoverySetID   string                  `json:"recovery_set_id"`
	TenantID        string                  `json:"tenant_id"`
	InventorySHA256 string                  `json:"inventory_sha256"`
	Entries         []documentsRecoveryCopy `json:"entries"`
}

type documentsRecoveryCopy struct {
	VersionID              string `json:"version_id"`
	SourceObjectKey        string `json:"source_object_key"`
	SourceObjectVersionID  string `json:"source_object_version_id"`
	SourceObjectGeneration int64  `json:"source_object_generation"`
	TargetObjectKey        string `json:"target_object_key"`
	TargetObjectVersionID  string `json:"target_object_version_id"`
	CiphertextSHA256       string `json:"ciphertext_sha256"`
	CiphertextBytes        int64  `json:"ciphertext_bytes"`
	KeyID                  string `json:"key_id"`
}

func buildDocumentsRecoveryInventory(
	ctx context.Context,
	service documentsapi.Recovery,
	source RecoveryObjectReader,
	target RecoveryObjectWriter,
	environmentID, tenantID, recoverySetID, snapshotID string,
) ([]byte, string, int64, error) {
	if service == nil ||
		source == nil || target == nil {
		return nil, "", 0, fmt.Errorf("documents recovery input is invalid")
	}
	items, inventoryDigest, err := service.ReadInventoryAtSnapshot(
		ctx, recoverySetID, snapshotID,
	)
	if err != nil {
		return nil, "", 0, err
	}
	entries := make([]documentsRecoveryCopy, 0, len(items))
	for _, item := range items {
		targetKey := documentsRecoveryObjectKey(
			environmentID, recoverySetID, tenantID, item.VersionID,
		)
		identity := ports.ObjectIdentity{
			Key: item.ObjectKey, VersionID: item.ObjectVersion,
			SHA256: hex.EncodeToString(item.CiphertextSHA256[:]),
			Bytes:  item.CiphertextBytes,
		}
		copyErr := source.DownloadVerified(ctx, identity, io.Discard)
		body, openErr := source.OpenExact(ctx, identity.Key, identity.VersionID)
		if copyErr == nil && openErr != nil {
			copyErr = openErr
		}
		var evidence ports.WriteEvidence
		if copyErr == nil {
			evidence, copyErr = target.PutVerified(
				ctx, ports.WriteRequest{
					Key: targetKey, Body: body,
					ContentLength:  identity.Bytes,
					ExpectedSHA256: identity.SHA256,
					ExpectedBytes:  identity.Bytes,
				},
			)
		}
		if body != nil {
			if closeErr := body.Close(); copyErr == nil {
				copyErr = closeErr
			}
		}
		if copyErr != nil ||
			service.RecordCopy(ctx, recoverySetID, item, evidence.VersionID) != nil {
			return nil, "", 0, fmt.Errorf("documents recovery copy failed")
		}
		entries = append(entries, documentsRecoveryCopy{
			VersionID:              item.VersionID,
			SourceObjectKey:        item.ObjectKey,
			SourceObjectVersionID:  item.ObjectVersion,
			SourceObjectGeneration: item.ObjectGeneration,
			TargetObjectKey:        evidence.Key,
			TargetObjectVersionID:  evidence.VersionID,
			CiphertextSHA256:       evidence.SHA256,
			CiphertextBytes:        evidence.Bytes, KeyID: item.KeyID,
		})
	}
	digest := hex.EncodeToString(inventoryDigest[:])
	value := documentsRecoveryInventory{
		Schema:        "documents-recovery-inventory.v1",
		EnvironmentID: environmentID, RecoverySetID: recoverySetID,
		TenantID: tenantID, InventorySHA256: digest, Entries: entries,
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, "", 0, err
	}
	document, err := canonicaljsonv1.CanonicalizeObject(
		raw, maximumDocumentsRecoveryInventory,
	)
	return document, digest, int64(len(entries)), err
}

func parseDocumentsRecoveryInventory(
	document []byte,
	environmentID, recoverySetID, tenantID string,
	service documentsapi.Recovery,
) (string, int64, error) {
	value, _, err := loadDocumentsRecoveryInventory(
		document, environmentID, recoverySetID, tenantID, service,
	)
	if err != nil {
		return "", 0, err
	}
	return value.InventorySHA256, int64(len(value.Entries)), nil
}

func documentsRecoveryObjectKey(
	environmentID, recoverySetID, tenantID, versionID string,
) string {
	return "environments/" + environmentID + "/recovery-sets/" +
		recoverySetID + "/documents/tenants/" + tenantID +
		"/versions/" + versionID + ".fcdoc1"
}
