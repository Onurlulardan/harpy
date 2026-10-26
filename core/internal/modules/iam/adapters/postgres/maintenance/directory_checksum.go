package maintenance

import (
	"context"
	"crypto/sha256"
	"fmt"

	maintenanceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/maintenance"
	"github.com/google/uuid"
)

func (store *store) executeDirectoryChecksum(
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
	subjects, err := store.loadDirectoryChecksumSubjects(
		ctx, claim.TenantID, values,
	)
	if err != nil {
		return maintenanceapp.Outcome{}, err
	}
	defer func() {
		for index := range subjects {
			subjects[index].wipe()
		}
	}()
	digest := sha256.New()
	counts := make(map[string]int64, len(keys))
	var scanned, drift int64
	for index := range subjects {
		scanned++
		if evaluateDirectoryChecksumSubject(subjects[index], keys, counts, digest) {
			drift++
		}
	}
	evidence, err := marshalDirectoryCountEvidence(keys, counts)
	if err != nil {
		return maintenanceapp.Outcome{}, fmt.Errorf(
			"encode directory checksum evidence: %w", err,
		)
	}
	resultSHA := digest.Sum(nil)
	finish := append(values, scanned, drift, resultSHA, evidence, nextID)
	outcome, err := scanOutcome(
		store.database.QueryRow(ctx, finishDirectoryChecksumSQL, finish...), claim.Kind,
	)
	clear(resultSHA)
	if err != nil {
		return maintenanceapp.Outcome{}, fmt.Errorf("finish directory checksum: %w", err)
	}
	return outcome, nil
}
