package authorization

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

const authorizationContextSchema = resolved.AuthorizationPolicySchemaV1

type authorizationContextDocument struct {
	Schema              string                      `json:"schema"`
	TenantID            string                      `json:"tenant_id"`
	ActorID             string                      `json:"actor_id"`
	QueryIdentitySHA256 string                      `json:"query_identity_sha256"`
	Owners              []authorizationOwnerContext `json:"owners"`
}

type authorizationOwnerContext struct {
	Owner                 string `json:"owner"`
	RequestBindingSHA256  string `json:"request_binding_sha256"`
	ContextRevisionSHA256 string `json:"context_revision_sha256"`
}

func authorizationContextSHA256(
	evaluations []policyEvaluation,
) ([sha256.Size]byte, error) {
	if len(evaluations) < 1 || len(evaluations) > maximumPolicyOwners {
		return [sha256.Size]byte{}, errPolicyEvaluationContract
	}
	first := evaluations[0].request
	if first.Validate() != nil {
		return [sha256.Size]byte{}, errPolicyEvaluationContract
	}
	document := authorizationContextDocument{
		Schema:   authorizationContextSchema,
		TenantID: first.tenantID.String(), ActorID: first.actorID.String(),
		QueryIdentitySHA256: hex.EncodeToString(first.queryIdentity[:]),
		Owners:              make([]authorizationOwnerContext, 0, len(evaluations)),
	}
	for index, evaluation := range evaluations {
		if !validContextEvaluation(first, evaluation, index, evaluations) {
			return [sha256.Size]byte{}, errPolicyEvaluationContract
		}
		document.Owners = append(document.Owners, authorizationOwnerContext{
			Owner: evaluation.owner,
			RequestBindingSHA256: hex.EncodeToString(
				evaluation.request.binding[:],
			),
			ContextRevisionSHA256: hex.EncodeToString(
				evaluation.decision.contextRevision[:],
			),
		})
	}
	canonical, err := attestation.MarshalCanonical(document)
	if err != nil {
		return [sha256.Size]byte{}, errPolicyEvaluationContract
	}
	return sha256.Sum256(canonical), nil
}

func validContextEvaluation(
	first PolicyRequest,
	evaluation policyEvaluation,
	index int,
	all []policyEvaluation,
) bool {
	request := evaluation.request
	if !evaluation.validate() ||
		request.tenantID != first.tenantID || request.actorID != first.actorID ||
		request.queryIdentity != first.queryIdentity {
		return false
	}
	return index == 0 || all[index-1].owner < evaluation.owner
}
