package v1adoption

import (
	"errors"
	"strings"
)

const (
	IssueInventoryInvalid        = "V1_INVENTORY_INVALID"
	IssueOnPremTenantAmbiguous   = "V1_ONPREMISE_TENANT_INVENTORY_AMBIGUOUS"
	IssueControlTenantAmbiguous  = "V1_SAAS_CONTROL_TENANT_AMBIGUOUS"
	IssueControlModeMismatch     = "V1_CONTROL_TENANT_MODE_MISMATCH"
	IssueProtectedOwnerAmbiguous = "V1_PROTECTED_OWNER_AMBIGUOUS"
	IssueOwnerCredentialMissing  = "V1_PROTECTED_OWNER_CREDENTIAL_MISSING"
	IssueOwnerMappingMissing     = "V1_PROTECTED_OWNER_MAPPING_MISSING"
	IssueBaseEvidenceAmbiguous   = "V1_BASE_ASSIGNMENT_EVIDENCE_AMBIGUOUS"
	IssueLicenseExchangeRequired = "V1_LICENSE_EXCHANGE_REQUIRED"
)

var ErrFenced = errors.New("V1 adoption is fenced")

type Issue struct {
	Code     string `json:"code"`
	Tenant   string `json:"tenant_id,omitempty"`
	Evidence string `json:"evidence,omitempty"`
}

type FenceError struct{ Issues []Issue }

func (err *FenceError) Error() string {
	codes := make([]string, 0, len(err.Issues))
	for _, issue := range err.Issues {
		codes = append(codes, issue.Code)
	}
	return ErrFenced.Error() + ": " + strings.Join(codes, ",")
}

func (err *FenceError) Unwrap() error { return ErrFenced }
