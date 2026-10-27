package actorassertionv1

import (
	"crypto/sha256"
	"encoding/binary"
	"regexp"

	"github.com/google/uuid"
)

const decisionDomain = "harpy:iam-v1-actor-decision:v1"

var catalogKeyPattern = regexp.MustCompile(
	`^[a-z][a-z0-9]*([-_][a-z0-9]+)*(\.[a-z][a-z0-9]*([-_][a-z0-9]+)*)*$`,
)

type DecisionEvidence struct {
	TenantID, SubjectID, PermissionID                      uuid.UUID
	TargetDomain                                           Domain
	ResourceKey, ActionKey                                 string
	TargetBindingSHA256                                    [32]byte
	IAMGroupGeneration, IAMGeneration, TargetGeneration    int64
	TokenVersion, AuthorizationVersion, AuthorizationClock int64
}

func (evidence DecisionEvidence) Validate() error {
	if evidence.TenantID == uuid.Nil || evidence.SubjectID == uuid.Nil ||
		evidence.PermissionID == uuid.Nil || evidence.TargetDomain.Validate() != nil ||
		!validCatalogKey(evidence.ResourceKey) || !validCatalogKey(evidence.ActionKey) ||
		evidence.TargetBindingSHA256 == [32]byte{} || evidence.IAMGroupGeneration <= 0 ||
		evidence.IAMGeneration <= 0 || evidence.TargetGeneration <= 0 ||
		evidence.TokenVersion <= 0 || evidence.AuthorizationVersion <= 0 ||
		evidence.AuthorizationClock < 0 {
		return ErrContract
	}
	return nil
}

func (evidence DecisionEvidence) Digest() ([32]byte, error) {
	if evidence.Validate() != nil {
		return [32]byte{}, ErrContract
	}
	parts := [][]byte{[]byte(decisionDomain), evidence.TenantID[:], evidence.SubjectID[:],
		evidence.PermissionID[:], []byte(evidence.TargetDomain), []byte(evidence.ResourceKey),
		[]byte(evidence.ActionKey), evidence.TargetBindingSHA256[:]}
	for _, value := range []int64{evidence.IAMGroupGeneration, evidence.IAMGeneration,
		evidence.TargetGeneration, evidence.TokenVersion, evidence.AuthorizationVersion,
		evidence.AuthorizationClock} {
		var encoded [8]byte
		binary.BigEndian.PutUint64(encoded[:], uint64(value))
		parts = append(parts, encoded[:])
	}
	hash := sha256.New()
	for _, part := range parts {
		var size [4]byte
		binary.BigEndian.PutUint32(size[:], uint32(len(part)))
		_, _ = hash.Write(size[:])
		_, _ = hash.Write(part)
	}
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result, nil
}

func validCatalogKey(value string) bool {
	return len(value) <= 100 && catalogKeyPattern.MatchString(value)
}
