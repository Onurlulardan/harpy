package authorization

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestAuthorizationContextHashIsCanonicalAndRevisionBound(t *testing.T) {
	evaluation := policyEvaluationForRequest(t, mustPolicyRequest(t))
	first, err := authorizationContextSHA256([]policyEvaluation{evaluation})
	second, secondErr := authorizationContextSHA256([]policyEvaluation{evaluation})
	if err != nil || secondErr != nil || first != second {
		t.Fatalf("context hashes differ: %x %x errors=%v,%v", first, second, err, secondErr)
	}
	const want = "8021214a0b723d5116a79ee40cf583c73c1278b6c59ff80b020c2865f011595a"
	if hex.EncodeToString(first[:]) != want {
		t.Fatalf("context hash = %s", hex.EncodeToString(first[:]))
	}
	changed := evaluation
	changed.decision.contextRevision = sha256.Sum256([]byte("changed-revision"))
	changedHash, err := authorizationContextSHA256([]policyEvaluation{changed})
	if err != nil || changedHash == first {
		t.Fatal("owner context revision was not bound")
	}
}

func TestAuthorizationContextHashRequiresOrderedExactEvaluations(t *testing.T) {
	forms := policyEvaluationForRequest(t, policyRequestForOwner(
		t, "forms", mustPolicyRequest(t).tenantID, mustPolicyRequest(t).actorID,
		sha256.Sum256([]byte("multi-owner-query")),
	))
	workflow := policyEvaluationForRequest(t, policyRequestForOwner(
		t, "workflow", forms.request.tenantID, forms.request.actorID,
		forms.request.queryIdentity,
	))
	if _, err := authorizationContextSHA256(
		[]policyEvaluation{forms, workflow},
	); err != nil {
		t.Fatalf("ordered evaluations rejected: %v", err)
	}
	cases := [][]policyEvaluation{
		nil,
		{workflow, forms},
		{forms, forms},
		{{owner: "forms", request: forms.request}},
	}
	for index, evaluations := range cases {
		if _, err := authorizationContextSHA256(evaluations); err == nil {
			t.Fatalf("invalid evaluation set %d was accepted", index)
		}
	}
}

func policyEvaluationForRequest(
	t *testing.T,
	request PolicyRequest,
) policyEvaluation {
	t.Helper()
	decision := exactDecision(t, request, nil)
	validated, err := validatePolicyDecision(request, decision)
	if err != nil {
		t.Fatal(err)
	}
	return policyEvaluation{
		owner: request.owner, request: request, decision: validated,
	}
}
