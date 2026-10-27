package maintenance

import (
	"context"
	"crypto/sha256"
	"fmt"
	"hash"

	maintenanceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/maintenance"
	"github.com/google/uuid"
)

func (store *store) executeDirectoryRepair(
	ctx context.Context,
	claim maintenanceapp.Claim,
	nextID uuid.UUID,
) (maintenanceapp.Outcome, error) {
	if ctx == nil || ctx.Err() != nil || store == nil || !store.runtime.usable() {
		return maintenanceapp.Outcome{}, ErrRuntimeInvalid
	}
	if receipt, found, err := store.readReceipt(ctx, claim); found || err != nil {
		return receipt, err
	}
	keys, err := store.runtime.snapshot()
	if err != nil {
		return maintenanceapp.Outcome{}, err
	}
	defer wipeDirectoryKeys(keys)
	values, err := claimValues(claim)
	if err != nil {
		return maintenanceapp.Outcome{}, err
	}
	sources, err := store.readDirectoryRepairPage(ctx, values)
	if err != nil {
		return maintenanceapp.Outcome{}, err
	}
	defer func() {
		for index := range sources {
			sources[index].wipe()
		}
	}()
	digest := sha256.New()
	var cursor any
	var scanned, affected int64
	for index := range sources {
		source := &sources[index]
		changed, applyErr := store.applyDirectoryRepair(ctx, values, *source, keys)
		writeDirectoryRepairDigest(digest, *source, keys)
		if applyErr != nil {
			return maintenanceapp.Outcome{}, applyErr
		}
		scanned++
		if changed {
			affected++
		}
		cursor = source.id
	}
	resultSHA := digest.Sum(nil)
	finish := append(values, cursor, scanned, affected, resultSHA, nextID)
	outcome, err := scanOutcome(
		store.database.QueryRow(ctx, finishDirectoryRepairSQL, finish...), claim.Kind,
	)
	clear(resultSHA)
	if err != nil {
		return maintenanceapp.Outcome{}, fmt.Errorf("finish directory repair: %w", err)
	}
	return outcome, nil
}

func (store *store) applyDirectoryRepair(
	ctx context.Context,
	claimValues []any,
	source directorySourceRow,
	keys []DirectoryKey,
) (bool, error) {
	evidence, err := marshalDirectoryRepairEvidence(
		keys, source.normalized, source.state == "active",
	)
	if err != nil {
		return false, fmt.Errorf("encode directory repair evidence: %w", err)
	}
	arguments := append(append([]any(nil), claimValues...),
		source.id, source.userID, source.kind, source.state,
		source.sourceRevision, source.normalizedSHA, evidence,
	)
	var validatedRevision int64
	if err := store.database.QueryRow(ctx, applyDirectoryRepairSQL, arguments...).Scan(
		&validatedRevision,
	); err != nil {
		return false, fmt.Errorf("validate directory repair source: %w", err)
	}
	if validatedRevision != source.sourceRevision {
		return false, ErrRuntimeInvalid
	}
	var changed int64
	if err := store.database.QueryRow(
		ctx, applyDirectoryProjectionSQL, claimValues[0], source.id,
		source.userID, source.kind, source.state, source.sourceRevision,
		source.normalizedSHA, evidence,
	).Scan(&changed); err != nil {
		return false, fmt.Errorf("apply directory projection: %w", err)
	}
	return changed > 0, nil
}

func writeDirectoryRepairDigest(
	digest hash.Hash,
	source directorySourceRow,
	keys []DirectoryKey,
) {
	writeDirectoryDigest(digest, []byte(source.id.String()), source.normalizedSHA,
		[]byte(source.kind), []byte(source.state))
	if source.state != "active" {
		return
	}
	for _, key := range keys {
		value := directoryLookupHMAC(key, source.normalized)
		writeDirectoryDigest(digest, []byte(key.ID), value[:])
		clear(value[:])
	}
}
