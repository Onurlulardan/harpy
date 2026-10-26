package humantask

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	humantaskapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/humantask"
	humantaskapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/humantask"
	"github.com/google/uuid"
)

const maximumCandidateEvidenceBytes = 16 << 10

type candidateDocument struct {
	ActorID              uuid.UUID       `json:"actor_id"`
	AuthorizationVersion int64           `json:"authorization_version"`
	CandidateKind        string          `json:"candidate_kind"`
	CandidateID          uuid.UUID       `json:"candidate_id"`
	Facts                []candidateFact `json:"facts"`
}

type candidateFact struct {
	Kind     string    `json:"kind"`
	ID       uuid.UUID `json:"id"`
	Revision int64     `json:"revision"`
}

func decodeCandidate(
	document []byte,
	request humantaskapi.VerificationRequest,
) (humantaskapp.CandidateEvidence, error) {
	canonical, err := canonicaljsonv1.CanonicalizeObject(
		document, maximumCandidateEvidenceBytes,
	)
	if err != nil {
		return humantaskapp.CandidateEvidence{}, errCandidateRead
	}
	decoder := json.NewDecoder(bytes.NewReader(canonical))
	decoder.DisallowUnknownFields()
	var value candidateDocument
	if err := decoder.Decode(&value); err != nil || !documentValid(value, request) {
		return humantaskapp.CandidateEvidence{}, errCandidateRead
	}
	return humantaskapp.CandidateEvidence{
		ActorID: [16]byte(value.ActorID), AuthorizationVersion: value.AuthorizationVersion,
		Candidate: request.Candidate, EvidenceSHA256: sha256.Sum256(canonical),
	}, nil
}

func documentValid(value candidateDocument, request humantaskapi.VerificationRequest) bool {
	if [16]byte(value.ActorID) != request.ActorID || value.AuthorizationVersion < 1 ||
		value.CandidateKind != request.Candidate.Kind ||
		[16]byte(value.CandidateID) != request.Candidate.ID ||
		len(value.Facts) < 1 || len(value.Facts) > 4 {
		return false
	}
	seen := map[uuid.UUID]bool{}
	for _, fact := range value.Facts {
		if fact.ID == uuid.Nil || fact.ID.Variant() != uuid.RFC4122 || fact.Revision < 1 ||
			!validFactKind(fact.Kind) || seen[fact.ID] {
			return false
		}
		seen[fact.ID] = true
	}
	return true
}

func validFactKind(value string) bool {
	return value == "user" || value == "role" || value == "role_assignment" ||
		value == "organization" || value == "organization_membership"
}
