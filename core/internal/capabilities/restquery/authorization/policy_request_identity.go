package authorization

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/Onurlulardan/harpy/internal/attestation"
)

const policyRequestIdentityVersion = "restquery_policy_request_v1"

type policyRequestIdentityDocument struct {
	Version              string   `json:"version"`
	TenantID             string   `json:"tenant_id"`
	ActorID              string   `json:"actor_id"`
	QueryIdentitySHA256  string   `json:"query_identity_sha256"`
	Owner                string   `json:"owner"`
	RequirementSHA256Set []string `json:"requirement_sha256_set"`
}

func (request PolicyRequest) bindingSHA256() ([sha256.Size]byte, error) {
	if !validAuthorizationUUID(request.tenantID) ||
		!validAuthorizationUUID(request.actorID) || request.owner == "" ||
		request.queryIdentity == [sha256.Size]byte{} || request.requirements == nil {
		return [sha256.Size]byte{}, errInvalidPolicyRequestIdentity
	}
	identities := make([]string, len(request.requirements))
	for index, requirement := range request.requirements {
		if requirement.Validate() != nil {
			return [sha256.Size]byte{}, errInvalidPolicyRequestIdentity
		}
		identities[index] = hex.EncodeToString(requirement.id[:])
	}
	document := policyRequestIdentityDocument{
		Version:  policyRequestIdentityVersion,
		TenantID: request.tenantID.String(), ActorID: request.actorID.String(),
		QueryIdentitySHA256: hex.EncodeToString(request.queryIdentity[:]),
		Owner:               request.owner, RequirementSHA256Set: identities,
	}
	canonical, err := attestation.MarshalCanonical(document)
	if err != nil {
		return [sha256.Size]byte{}, errInvalidPolicyRequestIdentity
	}
	return sha256.Sum256(canonical), nil
}
