package recoveryorchestration

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	documentsapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

type restoredDocumentLocator struct {
	VersionID       string `json:"version_id"`
	ObjectKey       string `json:"object_key"`
	OldVersionID    string `json:"old_version_id"`
	NewVersionID    string `json:"new_version_id"`
	OldGeneration   int64  `json:"old_generation"`
	NewGeneration   int64  `json:"new_generation"`
	CipherSHA256    string `json:"ciphertext_sha256"`
	CiphertextBytes int64  `json:"ciphertext_bytes"`
}

func (executor *databaseRestoreExecutor) rehydrateDocuments(
	ctx context.Context,
	claim ports.RestoreExecutionClaim,
	inventory documentsRecoveryInventory,
	items []documentsapi.RecoveryInventoryItem,
	service documentsapi.Recovery,
) (string, error) {
	writer := executor.storage.RehydrateWriter(claim.OperationID)
	if writer == nil || service == nil || len(items) != len(inventory.Entries) {
		return "", terminalRestore(fmt.Errorf("documents rehydrate is unavailable"))
	}
	records := make([]restoredDocumentLocator, 0, len(items))
	for index, item := range items {
		entry := inventory.Entries[index]
		if entry.VersionID != item.VersionID ||
			item.ObjectGeneration == int64(^uint64(0)>>1) {
			return "", terminalRestore(fmt.Errorf("documents inventory differs"))
		}
		source := ports.ObjectIdentity{
			Key: entry.TargetObjectKey, VersionID: entry.TargetObjectVersionID,
			SHA256: entry.CiphertextSHA256, Bytes: entry.CiphertextBytes,
		}
		if err := executor.storage.RecoveryReader().DownloadVerified(
			ctx, source, io.Discard,
		); err != nil {
			return "", resumableRestore(err)
		}
		body, err := executor.storage.RecoveryReader().OpenExact(
			ctx, source.Key, source.VersionID,
		)
		if err != nil {
			return "", resumableRestore(err)
		}
		written, writeErr := writer.PutVerified(ctx, ports.WriteRequest{
			Key: item.ObjectKey, Body: body, ContentLength: item.CiphertextBytes,
			ExpectedSHA256: hex.EncodeToString(item.CiphertextSHA256[:]),
			ExpectedBytes:  item.CiphertextBytes,
		})
		closeErr := body.Close()
		if writeErr != nil || closeErr != nil {
			return "", resumableRestore(fmt.Errorf("documents object rehydrate failed"))
		}
		generation := item.ObjectGeneration + 1
		if err = service.ApplyLocator(ctx, documentsapi.RecoveryLocatorRemap{
			RestoreOperationID: claim.OperationID,
			RecoverySetID:      claim.RecoverySetID, Item: item,
			TargetObjectKey: item.ObjectKey, TargetObjectVersion: written.VersionID,
			TargetObjectGeneration: generation,
		}); err != nil {
			return "", terminalRestore(err)
		}
		records = append(records, restoredDocumentLocator{
			VersionID: item.VersionID, ObjectKey: item.ObjectKey,
			OldVersionID: item.ObjectVersion, NewVersionID: written.VersionID,
			OldGeneration: item.ObjectGeneration, NewGeneration: generation,
			CipherSHA256:    hex.EncodeToString(item.CiphertextSHA256[:]),
			CiphertextBytes: item.CiphertextBytes,
		})
	}
	raw, err := json.Marshal(struct {
		Schema  string                    `json:"schema"`
		Records []restoredDocumentLocator `json:"records"`
	}{Schema: "documents-locator-remap-evidence.v1", Records: records})
	if err != nil {
		return "", terminalRestore(err)
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(raw, 4*1024*1024)
	if err != nil {
		return "", terminalRestore(err)
	}
	digest := sha256.Sum256(canonical)
	return hex.EncodeToString(digest[:]), nil
}
