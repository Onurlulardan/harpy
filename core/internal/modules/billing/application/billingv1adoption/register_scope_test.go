package billingv1adoptionapp

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
)

type canonicalEvidenceFixtureVerifier struct{ evidence CutoverEvidence }

func (value canonicalEvidenceFixtureVerifier) Verify(
	context.Context, Operation, ScopePlan,
) (CutoverEvidence, error) {
	return value.evidence, nil
}

func canonicalEvidenceFixture(operation Operation, plan ScopePlan) CutoverEvidence {
	digest := strings.Repeat("ab", 32)
	value, err := SealCutoverEvidence(CutoverEvidence{
		OperationID: operation.OperationID.String(), RunID: operation.RunID.String(),
		LeaseID: uuid.NewString(), EvidenceGeneration: 1,
		ScopeKind: plan.ScopeKind, ScopeID: plan.ScopeID, PlanSHA256: operation.PlanSHA256,
		LeaseManifestSHA256: digest, BatchSHA256: digest,
		SessionSealReceiptID: uuid.NewString(), SessionSealSHA256: digest,
	})
	if err != nil {
		panic(err)
	}
	return value
}

func TestRegisterScopePinsCanonicalPlan(t *testing.T) {
	plan := emptyEncodedPlan()
	operation := scopeOperation(plan)
	store := &fakeScopeStore{state: "planned"}
	verifier := canonicalEvidenceFixtureVerifier{canonicalEvidenceFixture(operation, plan)}
	state, err := RegisterScope(context.Background(), store, verifier, operation, plan)
	if err != nil || state != "planned" || len(store.document) == 0 {
		t.Fatalf("register result state=%q err=%v", state, err)
	}
	operation.PlanSHA256 = "00" + operation.PlanSHA256[2:]
	if _, err = RegisterScope(context.Background(), store, verifier, operation, plan); err == nil {
		t.Fatal("mismatched cutover plan hash was accepted")
	}
}
