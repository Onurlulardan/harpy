package transaction

import (
	"fmt"

	"github.com/google/uuid"
)

const credentialAuthorityGroup = "iam_system_resources"

type CredentialTargetEvidence struct {
	OperationID              string
	TenantID                 uuid.UUID
	Purpose                  CredentialPurpose
	LocatorKind              CredentialLocatorKind
	SourceUserID             uuid.UUID
	SourceIdentifierID       uuid.UUID
	SourceIdentifierKind     CredentialIdentifierKind
	SourceTokenID            uuid.UUID
	SourceFamilyID           uuid.UUID
	SourceRequestID          uuid.UUID
	SourceGrantID            uuid.UUID
	SourceSelector           [16]byte
	LocatorBindingSHA256     [32]byte
	SourceRevision           int64
	GroupGeneration          int64
	IAMGeneration            int64
	SystemResourceGeneration int64
}

type VerifiedCredentialTarget struct {
	operationID          string
	tenantID             uuid.UUID
	purpose              CredentialPurpose
	locatorKind          CredentialLocatorKind
	sourceUserID         uuid.UUID
	sourceIdentifierID   uuid.UUID
	sourceIdentifierKind CredentialIdentifierKind
	sourceTokenID        uuid.UUID
	sourceFamilyID       uuid.UUID
	sourceRequestID      uuid.UUID
	sourceGrantID        uuid.UUID
	sourceSelector       [16]byte
	locatorBindingSHA256 [32]byte
	sourceRevision       int64
	operation            VerifiedOperation
	issued               bool
}

func issueCredentialTarget(
	evidence CredentialTargetEvidence,
) (VerifiedCredentialTarget, error) {
	if !validCredentialEvidence(evidence) {
		return VerifiedCredentialTarget{}, fmt.Errorf("credential target capability is invalid")
	}
	operation, err := NewAuthorityIssuer().Issue(
		evidence.OperationID, credentialAuthorityGroup, evidence.GroupGeneration,
		map[string]int64{
			"iam":                    evidence.IAMGeneration,
			"system_resource_layout": evidence.SystemResourceGeneration,
		},
	)
	if err != nil {
		return VerifiedCredentialTarget{}, fmt.Errorf("credential target capability is invalid")
	}
	return VerifiedCredentialTarget{
		operationID: evidence.OperationID, tenantID: evidence.TenantID,
		purpose: evidence.Purpose, locatorKind: evidence.LocatorKind,
		sourceUserID:         evidence.SourceUserID,
		sourceIdentifierID:   evidence.SourceIdentifierID,
		sourceIdentifierKind: evidence.SourceIdentifierKind,
		sourceTokenID:        evidence.SourceTokenID,
		sourceFamilyID:       evidence.SourceFamilyID,
		sourceRequestID:      evidence.SourceRequestID,
		sourceGrantID:        evidence.SourceGrantID,
		sourceSelector:       evidence.SourceSelector,
		locatorBindingSHA256: evidence.LocatorBindingSHA256,
		sourceRevision:       evidence.SourceRevision, operation: operation, issued: true,
	}, nil
}

func validCredentialEvidence(evidence CredentialTargetEvidence) bool {
	return operationIDPattern.MatchString(evidence.OperationID) &&
		evidence.TenantID != uuid.Nil && evidence.Purpose.valid() &&
		evidence.LocatorKind.validFor(evidence.Purpose) &&
		validCredentialSource(evidence) &&
		evidence.LocatorBindingSHA256 != [32]byte{} &&
		evidence.SourceRevision > 0 && evidence.GroupGeneration > 0 &&
		evidence.IAMGeneration > 0 && evidence.SystemResourceGeneration > 0
}

func (target VerifiedCredentialTarget) valid() bool {
	return target.issued && target.tenantID != uuid.Nil && target.purpose.valid() &&
		target.locatorKind.validFor(target.purpose) &&
		target.credentialSourceValid() &&
		target.locatorBindingSHA256 != [32]byte{} && target.sourceRevision > 0 &&
		target.operationID == target.operation.operationID && target.operation.valid()
}
