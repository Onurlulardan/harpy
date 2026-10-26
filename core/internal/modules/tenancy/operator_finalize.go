package tenancy

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	onboardingapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/onboarding"
)

func (runtime *OperatorRuntime) FinalizeActivation(
	ctx context.Context,
	request OperatorFinalizationRequest,
	callback OperatorActivationCallback,
) error {
	if runtime == nil {
		return tenantapi.ErrOnboardingInvalid
	}
	return finalizeActivation(runtime.Onboarding, ctx, request, callback)
}

func finalizeActivation(
	onboarding *onboardingapp.Coordinator,
	ctx context.Context,
	request OperatorFinalizationRequest,
	callback OperatorActivationCallback,
) error {
	if onboarding == nil || callback == nil ||
		!request.Plan.IsValid() || request.ExpectedRevision < 1 {
		return tenantapi.ErrOnboardingInvalid
	}
	receiptID, receiptSHA256, evidenceSHA256 := completionEvidence(request.Plan)
	return onboarding.RunOperatorActivation(ctx, onboardingapp.ActivationRequest{
		Plan: request.Plan, OperationID: request.Plan.OperationID(),
		LeaseTokenSHA256: request.LeaseTokenSHA256,
		ExpectedRevision: request.ExpectedRevision,
	}, onboardingapp.ActivationCompletion{
		OperationID: request.Plan.OperationID(), ReceiptID: receiptID,
		ReceiptSHA256: receiptSHA256, EvidenceSHA256: evidenceSHA256,
	}, callback)
}

func completionEvidence(plan tenantapi.OnboardingPlan) (
	tenantapi.UUID, string, string,
) {
	operationID := plan.OperationID()
	receiptSeed := digestBytes("activation-receipt", operationID[:],
		[]byte(plan.StepPlanSHA256()))
	var receiptID tenantapi.UUID
	copy(receiptID[:], receiptSeed[:16])
	receiptID[6] = (receiptID[6] & 0x0f) | 0x70
	receiptID[8] = (receiptID[8] & 0x3f) | 0x80
	receiptSHA := digestBytes("activation-completed", receiptID[:],
		[]byte(plan.ManifestSHA256()), []byte(plan.StepPlanSHA256()))
	evidence := digestBytes("activation-evidence", operationID[:],
		[]byte(plan.RequestSHA256()), []byte(plan.Offering().SHA256))
	return receiptID, hex.EncodeToString(receiptSHA[:]), hex.EncodeToString(evidence[:])
}

func digestBytes(domain string, values ...[]byte) [32]byte {
	hash := sha256.New()
	_, _ = hash.Write([]byte(domain))
	for _, value := range values {
		_, _ = hash.Write([]byte{0})
		_, _ = hash.Write(value)
	}
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}
