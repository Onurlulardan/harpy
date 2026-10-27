package authorization

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/Onurlulardan/harpy/internal/attestation"
)

const globalPolicyRequestIdentityVersion = "restquery_global_policy_request_v1"

type globalPolicyRequestIdentityDocument struct {
	Version              string   `json:"version"`
	EnvironmentID        string   `json:"environment_id"`
	ActorID              string   `json:"actor_id,omitempty"`
	Audience             string   `json:"audience"`
	OperationID          string   `json:"operation_id"`
	SnapshotAt           string   `json:"snapshot_at"`
	QueryIdentitySHA256  string   `json:"query_identity_sha256"`
	Owner                string   `json:"owner"`
	RequirementSHA256Set []string `json:"requirement_sha256_set"`
}

func (request GlobalPolicyRequest) bindingSHA256() ([sha256.Size]byte, error) {
	if !validAuthorizationUUID(request.environmentID) || request.owner == "" ||
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
	document := globalPolicyRequestIdentityDocument{
		Version:       globalPolicyRequestIdentityVersion,
		EnvironmentID: request.environmentID.String(),
		Audience:      "authenticated", OperationID: request.operationID,
		SnapshotAt:          request.snapshotAt.Format(time.RFC3339Nano),
		QueryIdentitySHA256: hex.EncodeToString(request.queryIdentity[:]),
		Owner:               request.owner, RequirementSHA256Set: identities,
	}
	if request.public {
		document.Audience = "public"
	} else {
		document.ActorID = request.actorID.String()
	}
	canonical, err := attestation.MarshalCanonical(document)
	if err != nil {
		return [sha256.Size]byte{}, errInvalidPolicyRequestIdentity
	}
	return sha256.Sum256(canonical), nil
}
