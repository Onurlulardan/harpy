package dashboardexport

import (
	"crypto/sha256"
	"encoding/binary"
	"strconv"

	"github.com/google/uuid"
)

func renderAttemptIdentity(attempt RenderDispatchAttempt) RenderAttemptIdentity {
	return RenderAttemptIdentity{TenantID: attempt.Candidate.TenantID,
		ExportID: attempt.Candidate.ExportID, JobID: attempt.Candidate.JobID,
		AttemptID: attempt.AttemptID, LeaseOwner: attempt.Claim.LeaseOwner,
		Generation: attempt.Candidate.AttemptGeneration}
}

func validateClaimedAttempt(attempt RenderDispatchAttempt, spool EncryptedSpool) error {
	if validateRenderDispatch(attempt.Candidate, spool.Identity(), attempt, spool) != nil ||
		attempt.Claim.Disposition != ClaimExecute {
		return ErrInvalid
	}
	return nil
}

func renderUploadRequestSHA(draft ArtifactDraft) [32]byte {
	hash := sha256.New()
	hash.Write([]byte("analytics.dashboard-export.upload.v1\x00"))
	writeHashField(hash, []byte(draft.ObjectKey))
	writeHashField(hash, draft.HeaderSHA256[:])
	writeHashField(hash, draft.CiphertextSHA256[:])
	writeHashField(hash, []byte(strconv.FormatInt(draft.CiphertextBytes, 10)))
	writeHashField(hash, []byte(draft.ContentType))
	writeHashField(hash, []byte(draft.Filename))
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}

func renderUploadResultSHA(receipt UploadReceipt) [32]byte {
	hash := sha256.New()
	hash.Write([]byte("analytics.dashboard-export.upload-result.v1\x00"))
	writeHashField(hash, []byte(receipt.ObjectVersion))
	writeHashField(hash, []byte(receipt.ProviderRequestID))
	writeHashField(hash, []byte(receipt.UploadedAt.UTC().Format("2006-01-02T15:04:05.999999999Z07:00")))
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}

func writeHashField(hash interface{ Write([]byte) (int, error) }, value []byte) {
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(value)))
	_, _ = hash.Write(length[:])
	_, _ = hash.Write(value)
}

func failureSHA(code string) [32]byte { return sha256.Sum256([]byte(code)) }

func validRenderAttemptIdentity(value RenderAttemptIdentity) bool {
	return value.TenantID != uuid.Nil && validUUIDv7(value.ExportID) &&
		validUUIDv7(value.JobID) && validUUIDv7(value.AttemptID) &&
		validUUIDv7(value.LeaseOwner) && value.Generation >= 1 &&
		value.Generation <= MaximumAttempts
}

var _ RenderAttemptExecutor = (*Worker)(nil)
