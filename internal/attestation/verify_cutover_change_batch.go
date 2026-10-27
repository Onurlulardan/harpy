package attestation

import (
	"bytes"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
)

func VerifyCutoverChangeBatch(
	document []byte,
	lease VerifiedCutoverCredentialLease,
) (VerifiedCutoverChangeBatch, error) {
	if !lease.verified {
		return VerifiedCutoverChangeBatch{}, fmt.Errorf("cutover credential lease was not verified")
	}
	var batch cutoverv1.ChangeBatch
	if err := decodeStrictBounded(document, &batch, cutoverv1.MaximumChangeBatchBytes); err != nil {
		return VerifiedCutoverChangeBatch{}, fmt.Errorf("decode cutover change batch: %w", err)
	}
	if err := cutoverv1.ValidateChangeBatch(batch); err != nil {
		return VerifiedCutoverChangeBatch{}, err
	}
	canonical, err := marshalCutoverCanonical(batch)
	if err != nil || !bytes.Equal(document, canonical) {
		return VerifiedCutoverChangeBatch{}, fmt.Errorf("cutover change batch is not canonical")
	}
	payload, err := marshalCutoverCanonical(batch.Payload)
	if err != nil {
		return VerifiedCutoverChangeBatch{}, err
	}
	digest := sha256.Sum256(payload)
	digestHex := hex.EncodeToString(digest[:])
	if subtle.ConstantTimeCompare([]byte(digestHex), []byte(batch.BatchSHA256)) != 1 {
		return VerifiedCutoverChangeBatch{}, fmt.Errorf("cutover change batch digest differs")
	}
	if err := bindBatchToLease(batch.Payload, lease); err != nil {
		return VerifiedCutoverChangeBatch{}, err
	}
	return VerifiedCutoverChangeBatch{
		batch: batch, payloadJSON: payload, verified: true,
	}, nil
}

func marshalCutoverCanonical(value any) ([]byte, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode cutover change batch JSON: %w", err)
	}
	return canonicalizeBounded(encoded, cutoverv1.MaximumChangeBatchBytes)
}

func bindBatchToLease(
	payload cutoverv1.ChangeBatchPayload,
	lease VerifiedCutoverCredentialLease,
) error {
	if payload.OperationID != lease.OperationID() || payload.RunID != lease.RunID() ||
		payload.ScopeKind != lease.ScopeKind() || payload.ScopeID != lease.ScopeID() {
		return fmt.Errorf("cutover change batch authority differs")
	}
	relation, found := lease.Relation(payload.RelationOrdinal)
	if !found {
		return fmt.Errorf("cutover change batch relation is not allowlisted")
	}
	if !checkpointKeyAllowed(relation.ExpectedCheckpointKeys, payload.CheckpointKey) {
		return fmt.Errorf("cutover change batch checkpoint is not allowlisted")
	}
	return nil
}

func checkpointKeyAllowed(expected []string, candidate string) bool {
	index := sort.SearchStrings(expected, candidate)
	return index < len(expected) && expected[index] == candidate
}
