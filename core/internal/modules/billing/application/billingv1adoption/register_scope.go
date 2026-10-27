package billingv1adoptionapp

import (
	"context"
	"fmt"
)

func RegisterScope(
	ctx context.Context,
	store ScopeStore,
	verifier CutoverEvidenceVerifier,
	operation Operation,
	plan ScopePlan,
) (string, error) {
	if ctx == nil || ctx.Err() != nil || store == nil || verifier == nil ||
		!validOperation(operation) {
		return "", fmt.Errorf("Billing V1 adoption registration is not configured")
	}
	document, digest, err := EncodeScopePlan(plan)
	if err != nil || digest != operation.PlanSHA256 {
		return "", fmt.Errorf("Billing V1 adoption plan does not match cutover operation")
	}
	evidence, err := verifier.Verify(ctx, operation, plan)
	if err != nil || !validCutoverEvidence(operation, plan, evidence) {
		return "", fmt.Errorf("Billing V1 adoption cutover evidence is invalid")
	}
	state, err := store.Register(ctx, operation, plan, document, evidence)
	if err != nil {
		return "", fmt.Errorf("register Billing V1 adoption plan: %w", err)
	}
	if state != "planned" && state != "blocked" {
		return "", fmt.Errorf("Billing V1 adoption registration returned invalid state")
	}
	return state, nil
}
