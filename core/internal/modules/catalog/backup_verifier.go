package catalog

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strconv"

	backupstore "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/postgres/backup"
	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/platform/backupreceipt"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type BackupVerifier struct{ binder *backupstore.Binder }

func NewBackupVerifier() *BackupVerifier {
	return &BackupVerifier{binder: backupstore.NewBinder()}
}

func (verifier *BackupVerifier) Verify(
	ctx context.Context, scope platformtx.TenantScope, receiptRaw string,
	requirement catalogapp.BackupRequirement,
) (catalogapp.BackupEvidence, error) {
	receiptID, parsedRequirement, err := parseBackupRequirement(receiptRaw, requirement)
	if err != nil {
		return catalogapp.BackupEvidence{}, catalogapp.ErrInvalid
	}
	if verifier == nil || verifier.binder == nil {
		return catalogapp.BackupEvidence{}, catalogapp.ErrUnavailable
	}
	reader, err := verifier.binder.Bind(ctx, scope)
	if err != nil || reader == nil {
		return catalogapp.BackupEvidence{}, catalogapp.ErrUnavailable
	}
	raw, err := reader.Fetch(ctx, receiptID)
	if err != nil || raw.ID != receiptID {
		return catalogapp.BackupEvidence{}, catalogapp.ErrBackupBlocked
	}
	receipt, err := decodeBackupReceipt(raw)
	if err != nil || receipt.ValidFor(parsedRequirement) != nil {
		return catalogapp.BackupEvidence{}, catalogapp.ErrBackupBlocked
	}
	digest := backupEvidenceSHA256(receipt)
	if digest == "" {
		return catalogapp.BackupEvidence{}, catalogapp.ErrBackupBlocked
	}
	return catalogapp.BackupEvidence{
		ReceiptID: receipt.ID.String(), TenantID: receipt.TenantID.String(),
		PackageID:          receipt.PackageID.String(),
		CatalogOperationID: receipt.CatalogOperationID.String(),
		State:              "READY", SHA256: digest, VerifierVersion: receipt.VerifierVersion,
		RestoreVerifiedAt: receipt.RestoreVerifiedAt, ExpiresAt: receipt.ExpiresAt,
	}, nil
}

func parseBackupRequirement(
	receiptRaw string, input catalogapp.BackupRequirement,
) (uuid.UUID, backupreceipt.Requirement, error) {
	receiptID, err := uuid.Parse(receiptRaw)
	tenantID, tenantErr := uuid.Parse(input.TenantID)
	packageID, packageErr := uuid.Parse(input.PackageID)
	operationID, operationErr := uuid.Parse(input.CatalogOperationID)
	if err != nil || tenantErr != nil || packageErr != nil || operationErr != nil {
		return uuid.Nil, backupreceipt.Requirement{}, catalogapp.ErrInvalid
	}
	generations := make([]backupreceipt.OwnerGeneration, 0, len(input.OwnerGenerations))
	for _, item := range input.OwnerGenerations {
		generations = append(generations, backupreceipt.OwnerGeneration{
			Owner: item.Owner, Generation: item.Generation,
		})
	}
	return receiptID, backupreceipt.Requirement{TenantID: tenantID, PackageID: packageID,
		CatalogOperationID: operationID, OwnerGenerations: generations,
		Now: input.Now, MinimumExpiry: input.MinimumExpiry}, nil
}

func decodeBackupReceipt(raw backupstore.Receipt) (backupreceipt.Receipt, error) {
	value := backupreceipt.Receipt{ID: raw.ID, TenantID: raw.TenantID, PackageID: raw.PackageID,
		CatalogOperationID: raw.CatalogOperationID, OwnerManifestRoot: raw.OwnerManifestRoot,
		DatabaseSnapshotRoot: raw.DatabaseSnapshotRoot, ObjectInventoryRoot: raw.ObjectInventoryRoot,
		KeyRevision: raw.KeyRevision, CompletedAt: raw.CompletedAt,
		RestoreVerifiedAt: raw.RestoreVerifiedAt, ExpiresAt: raw.ExpiresAt,
		VerifierVersion: raw.VerifierVersion, LegalHold: raw.LegalHold}
	err := json.Unmarshal(raw.OwnerGenerationsJSON, &value.OwnerGenerations)
	return value, err
}

func backupEvidenceSHA256(value backupreceipt.Receipt) string {
	generations := append([]backupreceipt.OwnerGeneration(nil), value.OwnerGenerations...)
	sort.Slice(generations, func(i, j int) bool {
		return generations[i].Owner < generations[j].Owner
	})
	fields := []string{
		"catalog.backup-evidence.v2", value.ID.String(), value.TenantID.String(),
		value.PackageID.String(), value.CatalogOperationID.String(),
		value.OwnerManifestRoot, value.DatabaseSnapshotRoot, value.ObjectInventoryRoot,
		value.KeyRevision, value.CompletedAt.Format(timeLayout),
		value.RestoreVerifiedAt.Format(timeLayout), value.ExpiresAt.Format(timeLayout),
		value.VerifierVersion, strconv.FormatBool(value.LegalHold),
	}
	for _, generation := range generations {
		fields = append(fields, generation.Owner, strconv.FormatInt(generation.Generation, 10))
	}
	hash := sha256.New()
	var size [4]byte
	for _, field := range fields {
		binary.BigEndian.PutUint32(size[:], uint32(len(field)))
		_, _ = hash.Write(size[:])
		_, _ = hash.Write([]byte(field))
	}
	return hex.EncodeToString(hash.Sum(nil))
}

const timeLayout = "2006-01-02T15:04:05.000000Z07:00"

var _ catalogapp.BackupVerifier = (*BackupVerifier)(nil)
