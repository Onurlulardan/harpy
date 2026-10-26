package postgres

import (
	"context"
	"fmt"
	"sort"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func (database *AdoptionDatabase) Apply(
	ctx context.Context,
	lease attestation.VerifiedCutoverCredentialLease,
	batch attestation.VerifiedCutoverChangeBatch,
) (cutoverv1.ApplyResult, error) {
	if database == nil || !database.sealVerified || len(database.target) < 1 {
		return cutoverv1.ApplyResult{}, fmt.Errorf("cutover session seal is not live-verified")
	}
	payload := batch.Payload()
	relation, found := lease.Relation(payload.RelationOrdinal)
	if !found || payload.OperationID != lease.OperationID() || payload.RunID != lease.RunID() {
		return cutoverv1.ApplyResult{}, fmt.Errorf("cutover batch authority differs")
	}
	checkpointIndex := sort.SearchStrings(relation.ExpectedCheckpointKeys, payload.CheckpointKey)
	if checkpointIndex >= len(relation.ExpectedCheckpointKeys) ||
		relation.ExpectedCheckpointKeys[checkpointIndex] != payload.CheckpointKey {
		return cutoverv1.ApplyResult{}, fmt.Errorf("cutover batch checkpoint authority differs")
	}
	input := targetApplyInput{
		payload: payload, relation: relation, payloadJSON: batch.PayloadJSON(),
		batchSHA256: batch.SHA256(), leaseManifestSHA256: lease.ManifestSHA256(),
	}
	return applyTargetTransaction(
		ctx, database.target[0], database.cleanupTimeout, database.applyRegistry, input,
	)
}
