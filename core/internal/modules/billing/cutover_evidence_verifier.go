package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	billingv1adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/billingv1adoption"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/cutover"
)

type CutoverEvidenceClock func() time.Time

type cutoverEvidenceVerifier struct {
	artifacts config.CutoverAdoptionArtifacts
	now       CutoverEvidenceClock
}

func NewCutoverEvidenceVerifier(
	artifacts config.CutoverAdoptionArtifacts,
	now CutoverEvidenceClock,
) (billingv1adoptionapp.CutoverEvidenceVerifier, error) {
	if now == nil || artifacts.LeaseManifestPath().Value() == "" ||
		artifacts.SessionSealReceiptPath().Value() == "" {
		return nil, fmt.Errorf("billing cutover evidence verifier is not configured")
	}
	return &cutoverEvidenceVerifier{artifacts: artifacts, now: now}, nil
}

func (verifier *cutoverEvidenceVerifier) Verify(
	ctx context.Context,
	operation billingv1adoptionapp.Operation,
	plan billingv1adoptionapp.ScopePlan,
) (billingv1adoptionapp.CutoverEvidence, error) {
	if verifier == nil || ctx == nil || ctx.Err() != nil {
		return billingv1adoptionapp.CutoverEvidence{},
			fmt.Errorf("billing cutover evidence verification was cancelled")
	}
	now := verifier.now().UTC()
	evidence, err := cutover.LoadAdoptionEvidence(verifier.artifacts, now)
	if err != nil {
		return billingv1adoptionapp.CutoverEvidence{},
			fmt.Errorf("load verified cutover evidence: %w", err)
	}
	seal, err := cutover.LoadSessionSeal(
		verifier.artifacts.SessionSealReceiptPath().Value(), evidence, now,
	)
	if err != nil {
		return billingv1adoptionapp.CutoverEvidence{},
			fmt.Errorf("load verified cutover session seal: %w", err)
	}
	lease, err := decodeBillingCutoverLease(evidence.Lease.CanonicalDocument())
	if err != nil || !billingCutoverLeaseMatches(lease, evidence, operation, plan) {
		return billingv1adoptionapp.CutoverEvidence{},
			fmt.Errorf("verified cutover lease does not authorize Billing adoption")
	}
	payload := evidence.Batch.Payload()
	if payload.OperationID != evidence.Lease.OperationID() ||
		payload.RunID != evidence.Lease.RunID() || payload.ScopeKind != plan.ScopeKind ||
		payload.ScopeID != plan.ScopeID || payload.Position.BatchOrdinal < 1 {
		return billingv1adoptionapp.CutoverEvidence{},
			fmt.Errorf("verified cutover batch identity differs")
	}
	return billingv1adoptionapp.SealCutoverEvidence(
		billingv1adoptionapp.CutoverEvidence{
			OperationID: operation.OperationID.String(), RunID: operation.RunID.String(),
			LeaseID:            evidence.Lease.LeaseID(),
			EvidenceGeneration: int64(payload.Position.BatchOrdinal),
			ScopeKind:          plan.ScopeKind, ScopeID: plan.ScopeID,
			PlanSHA256:           operation.PlanSHA256,
			LeaseManifestSHA256:  evidence.Lease.ManifestSHA256(),
			BatchSHA256:          evidence.Batch.SHA256(),
			SessionSealReceiptID: seal.ReceiptID(),
			SessionSealSHA256:    seal.ReceiptSHA256(),
		},
	)
}

func decodeBillingCutoverLease(document []byte) (cutoverv1.CredentialLease, error) {
	var value cutoverv1.CredentialLease
	if err := json.Unmarshal(document, &value); err != nil {
		return cutoverv1.CredentialLease{}, fmt.Errorf("decode verified cutover lease: %w", err)
	}
	return value, nil
}

func billingCutoverLeaseMatches(
	lease cutoverv1.CredentialLease,
	evidence cutover.AdoptionEvidence,
	operation billingv1adoptionapp.Operation,
	plan billingv1adoptionapp.ScopePlan,
) bool {
	expectedGroup := "billing"
	expectedDomains := []string{"billing"}
	if plan.ScopeKind == "tenant" {
		expectedGroup = "billing_forms"
		expectedDomains = append(expectedDomains, "forms")
	}
	return evidence.Lease.OperationID() == operation.OperationID.String() &&
		evidence.Lease.RunID() == operation.RunID.String() &&
		evidence.Lease.PlanSHA256() == operation.PlanSHA256 &&
		evidence.Lease.ScopeKind() == plan.ScopeKind &&
		evidence.Lease.ScopeID() == plan.ScopeID && lease.GroupID == expectedGroup &&
		slices.Contains(lease.Domains, expectedDomains[0]) &&
		(len(expectedDomains) == 1 || slices.Contains(lease.Domains, expectedDomains[1]))
}
