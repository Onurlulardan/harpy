package humantask

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	humantaskapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/humantask"
	"github.com/google/uuid"
)

type assignmentDocument struct {
	Kind     string          `json:"kind"`
	ID       uuid.UUID       `json:"id"`
	Revision int64           `json:"revision"`
	Facts    []candidateFact `json:"facts"`
}

func decodeAssignment(
	document []byte,
	request humantaskapi.AssignmentRequest,
) (humantaskapi.AssignmentEvidence, error) {
	canonical, err := canonicaljsonv1.CanonicalizeObject(
		document, maximumCandidateEvidenceBytes,
	)
	if err != nil {
		return humantaskapi.AssignmentEvidence{}, errCandidateRead
	}
	decoder := json.NewDecoder(bytes.NewReader(canonical))
	decoder.DisallowUnknownFields()
	var value assignmentDocument
	if err := decoder.Decode(&value); err != nil || !assignmentDocumentValid(value, request) {
		return humantaskapi.AssignmentEvidence{}, errCandidateRead
	}
	return humantaskapi.AssignmentEvidence{
		TenantID: request.TenantID, Assignment: request.Assignment,
		Revision: value.Revision, EvidenceSHA256: sha256.Sum256(canonical),
		VerifiedAt: request.VerifyAt,
	}, nil
}

func assignmentDocumentValid(
	value assignmentDocument,
	request humantaskapi.AssignmentRequest,
) bool {
	if value.Kind != request.Assignment.Kind || [16]byte(value.ID) != request.Assignment.ID ||
		value.Revision < 1 || len(value.Facts) < 1 || len(value.Facts) > 2 {
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
