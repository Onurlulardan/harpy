package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
	"github.com/google/uuid"
)

func (store *Store) SaveBinding(
	ctx context.Context,
	record ports.BindingRecord,
	proof *artifact.VerifiedBindingProof,
) error {
	if !validBindingRecord(record, proof) {
		return ports.ErrConflict
	}
	now := store.clock.Now()
	if !validDatabaseInstant(now) {
		return ports.ErrConflict
	}
	var proofSHA any
	var proofBytes any
	if proof != nil {
		proofSHA, proofBytes = proof.SHA256, proof.Bytes
	}
	err := store.withBound(ctx, transactionCommand, func(bound boundStore) error {
		_, err := bound.database.Execute(ctx, bindSQL,
			record.BindingID, record.EnvironmentID, string(record.DeploymentMode),
			string(record.RegistryKind), optionalText(record.LicenseBindingID),
			optionalText(record.FinalizationReceiptSHA256),
			optionalText(record.RecoveryTrustBundleSHA256), record.SignerGeneration,
			record.SignerKID, record.SignerPublicKeySHA256,
			proofSHA, proofBytes, now)
		return store.mapDatabaseError(err)
	})
	if err != nil {
		return fmt.Errorf("create platform authority binding: %w", err)
	}
	return nil
}

func validBindingRecord(
	record ports.BindingRecord,
	proof *artifact.VerifiedBindingProof,
) bool {
	if _, err := uuid.Parse(record.BindingID); err != nil {
		return false
	}
	if _, err := uuid.Parse(record.EnvironmentID); err != nil ||
		record.SignerGeneration < 1 || record.SignerKID == "" ||
		len(record.SignerPublicKeySHA256) != 64 {
		return false
	}
	if record.DeploymentMode == artifact.ModeSaaS {
		return record.RegistryKind == artifact.RegistrySaaS && proof == nil &&
			record.LicenseBindingID == nil && record.FinalizationReceiptSHA256 == nil &&
			record.RecoveryTrustBundleSHA256 == nil
	}
	return validOnPremiseBinding(record, proof)
}

func validDatabaseInstant(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond() == 0
}

func optionalText(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}
