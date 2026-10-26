package humantask

import (
	"crypto/sha256"
	"encoding/binary"

	humantaskapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/humantask"
)

func evidenceDigest(
	evidence humantaskapi.VerificationEvidence,
	candidateDigest [32]byte,
) [32]byte {
	hash := sha256.New()
	hash.Write([]byte("iam-workflow-human-task-evidence.v1\x00"))
	hash.Write(evidence.TenantID[:])
	hash.Write(evidence.ActorID[:])
	hash.Write([]byte(evidence.Candidate.Kind + "\x00"))
	hash.Write(evidence.Candidate.ID[:])
	hash.Write([]byte(evidence.Action + "\x00"))
	hash.Write(evidence.PermissionID[:])
	var scalar [8]byte
	binary.BigEndian.PutUint64(scalar[:], uint64(evidence.AuthorizationVersion))
	hash.Write(scalar[:])
	binary.BigEndian.PutUint64(scalar[:], uint64(evidence.VerifiedAt.UnixMicro()))
	hash.Write(scalar[:])
	hash.Write(candidateDigest[:])
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}
