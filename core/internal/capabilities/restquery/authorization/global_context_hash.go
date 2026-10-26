package authorization

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/Onurlulardan/harpy/internal/attestation"
)

type globalAuthorizationContextDocument struct {
	Schema                string `json:"schema"`
	EnvironmentID         string `json:"environment_id"`
	ActorID               string `json:"actor_id,omitempty"`
	Audience              string `json:"audience"`
	OperationID           string `json:"operation_id"`
	SnapshotAt            string `json:"snapshot_at"`
	QueryIdentitySHA256   string `json:"query_identity_sha256"`
	Owner                 string `json:"owner"`
	RequestBindingSHA256  string `json:"request_binding_sha256"`
	ContextRevisionSHA256 string `json:"context_revision_sha256"`
}

func globalAuthorizationContextSHA256(
	request GlobalPolicyRequest,
	decision validatedDecision,
) ([sha256.Size]byte, error) {
	if request.Validate() != nil || !sameDigest(request.binding, decision.requestBinding) ||
		decision.contextRevision == [sha256.Size]byte{} {
		return [sha256.Size]byte{}, errPolicyEvaluationContract
	}
	document := globalAuthorizationContextDocument{
		Schema:        "restquery_global_authorization_v1",
		EnvironmentID: request.environmentID.String(), Audience: "authenticated",
		OperationID:           request.operationID,
		SnapshotAt:            request.snapshotAt.Format(time.RFC3339Nano),
		QueryIdentitySHA256:   hex.EncodeToString(request.queryIdentity[:]),
		Owner:                 request.owner,
		RequestBindingSHA256:  hex.EncodeToString(request.binding[:]),
		ContextRevisionSHA256: hex.EncodeToString(decision.contextRevision[:]),
	}
	if request.public {
		document.Audience = "public"
	} else {
		document.ActorID = request.actorID.String()
	}
	canonical, err := attestation.MarshalCanonical(document)
	if err != nil {
		return [sha256.Size]byte{}, errPolicyEvaluationContract
	}
	return sha256.Sum256(canonical), nil
}
