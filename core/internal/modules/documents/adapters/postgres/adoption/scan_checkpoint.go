package adoption

import (
	"strings"

	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/adoption"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *Store) scanCheckpoint(row Row) (adoptionapp.Checkpoint, bool, error) {
	var result adoptionapp.Checkpoint
	var runID, sourceItem, sourceVersion, targetItem, targetVersion, state string
	var sourceSHA, sourceMetadata, targetPlainSHA, targetCipherSHA, targetMetadata []byte
	var targetKey, targetVersionKey, targetKeyID, appliedLSN, issue dbText
	var targetGeneration, targetCipherBytes dbInt64
	var wrapped []byte
	var verified dbInstant
	var replay bool
	err := row.Scan(&runID, &sourceItem, &sourceVersion, &result.SourceObjectKey,
		&result.SourceRevision, &result.SourceBytes, &sourceSHA, &sourceMetadata,
		&result.SourceLSN, &targetItem, &targetVersion, &targetKey, &targetVersionKey, &targetGeneration,
		&targetCipherBytes, &targetPlainSHA, &targetCipherSHA, &targetKeyID, &wrapped,
		&targetMetadata, &appliedLSN, &state, &issue, &verified,
		&result.UpdatedAt, &replay)
	if err != nil || result.SourceBytes < 1 || len(sourceSHA) != 32 ||
		len(sourceMetadata) != 32 {
		return adoptionapp.Checkpoint{}, false, adoptionapp.ErrConflict
	}
	identifiers := []*domain.ID{&result.RunID, &result.SourceItemID,
		&result.SourceVersionID, &result.TargetItemID, &result.TargetVersionID}
	values := []string{runID, sourceItem, sourceVersion, targetItem, targetVersion}
	for index := range identifiers {
		parsed, parseErr := domain.ParseID(values[index])
		if parseErr != nil {
			return adoptionapp.Checkpoint{}, false, adoptionapp.ErrUnavailable
		}
		*identifiers[index] = parsed
	}
	copy(result.SourceSHA256[:], sourceSHA)
	copy(result.SourceMetadataSHA256[:], sourceMetadata)
	result.State = adoptionapp.CheckpointState(strings.ToLower(state))
	result.AppliedLSN, result.IssueCode, result.VerifiedAt = appliedLSN.Text,
		issue.Text, verified.Time
	if targetKey.Valid {
		result.TargetObject = domain.ObjectReceipt{Key: targetKey.Text,
			Version: targetVersionKey.Text, KeyID: targetKeyID.Text,
			Generation: uint64(targetGeneration.Int64), PlaintextBytes: result.SourceBytes,
			CiphertextBytes: targetCipherBytes.Int64,
			WrappedDEK:      append([]byte(nil), wrapped...)}
		copy(result.TargetObject.PlaintextSHA256[:], targetPlainSHA)
		copy(result.TargetObject.CiphertextSHA256[:], targetCipherSHA)
		copy(result.TargetMetadataSHA256[:], targetMetadata)
	}
	return result, replay, nil
}

const checkpointColumns = `run_id, source_document_id, source_version_id,
 source_object_key, source_object_revision, source_bytes, source_sha256,
 source_metadata_sha256, source_lsn::text, target_document_id, target_version_id, target_object_key,
 target_object_version, target_object_generation, target_ciphertext_bytes,
 target_plaintext_sha256, target_ciphertext_sha256, target_key_id,
 target_wrapped_dek, metadata_sha256, applied_lsn::text, lifecycle_state,
 issue_code, verified_at, updated_at`
