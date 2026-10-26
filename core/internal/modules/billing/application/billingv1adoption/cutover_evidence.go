package billingv1adoptionapp

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	"github.com/google/uuid"
)

const CutoverEvidenceContractVersion = "harpy/billing-v1-cutover-evidence/v1"

type CutoverEvidence struct {
	ContractVersion      string `json:"contract_version"`
	OperationID          string `json:"operation_id"`
	RunID                string `json:"run_id"`
	LeaseID              string `json:"lease_id"`
	EvidenceGeneration   int64  `json:"evidence_generation"`
	ScopeKind            string `json:"scope_kind"`
	ScopeID              string `json:"scope_id"`
	PlanSHA256           string `json:"plan_sha256"`
	LeaseManifestSHA256  string `json:"lease_manifest_sha256"`
	BatchSHA256          string `json:"batch_sha256"`
	SessionSealReceiptID string `json:"session_seal_receipt_id"`
	SessionSealSHA256    string `json:"session_seal_sha256"`
	CanonicalDocument    []byte `json:"-"`
	CanonicalEnvelopeSHA string `json:"-"`
}

func SealCutoverEvidence(value CutoverEvidence) (CutoverEvidence, error) {
	value.ContractVersion = CutoverEvidenceContractVersion
	value.CanonicalDocument = nil
	value.CanonicalEnvelopeSHA = ""
	document, err := json.Marshal(value)
	if err != nil {
		return CutoverEvidence{}, fmt.Errorf("marshal cutover evidence: %w", err)
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, 16*1024)
	if err != nil {
		return CutoverEvidence{}, fmt.Errorf("canonicalize cutover evidence: %w", err)
	}
	digest := sha256.Sum256(canonical)
	value.CanonicalDocument = canonical
	value.CanonicalEnvelopeSHA = hex.EncodeToString(digest[:])
	return value, nil
}

func validCutoverEvidence(operation Operation, plan ScopePlan, value CutoverEvidence) bool {
	sealed, err := SealCutoverEvidence(value)
	return err == nil && value.OperationID == operation.OperationID.String() &&
		value.RunID == operation.RunID.String() && value.PlanSHA256 == operation.PlanSHA256 &&
		value.ScopeKind == plan.ScopeKind && value.ScopeID == plan.ScopeID &&
		uuid.Validate(value.LeaseID) == nil && uuid.Validate(value.SessionSealReceiptID) == nil &&
		value.EvidenceGeneration > 0 && validDigest(value.LeaseManifestSHA256) &&
		validDigest(value.BatchSHA256) && validDigest(value.SessionSealSHA256) &&
		value.CanonicalEnvelopeSHA == sealed.CanonicalEnvelopeSHA &&
		bytes.Equal(value.CanonicalDocument, sealed.CanonicalDocument)
}
