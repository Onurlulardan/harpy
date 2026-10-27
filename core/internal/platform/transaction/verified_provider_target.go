package transaction

import (
	"fmt"

	"github.com/google/uuid"
)

const providerAuthorityGroup = "billing_forms"

type ProviderTargetEvidence struct {
	TenantID            uuid.UUID
	MappingID           uuid.UUID
	ProviderAccountID   uuid.UUID
	SubscriptionID      uuid.UUID
	InvoiceID           uuid.UUID
	MappingRevision     int64
	GroupGeneration     int64
	BillingGeneration   int64
	LookupBindingSHA256 [32]byte
}

type verifiedProviderTarget struct {
	tenantID          uuid.UUID
	mappingID         uuid.UUID
	providerAccountID uuid.UUID
	subscriptionID    uuid.UUID
	invoiceID         uuid.UUID
	mappingRevision   int64
	lookupBinding     [32]byte
	operation         VerifiedOperation
	issued            bool
}

func issueProviderTarget(
	lookup ProviderLookup,
	evidence ProviderTargetEvidence,
) (verifiedProviderTarget, error) {
	if !lookup.valid() || !validProviderEvidence(evidence) ||
		!providerEvidenceMatchesLookup(lookup, evidence) ||
		evidence.LookupBindingSHA256 != providerLookupBinding(lookup) {
		return verifiedProviderTarget{}, fmt.Errorf("provider target capability is invalid")
	}
	operation, err := NewAuthorityIssuer().Issue(
		lookup.OperationID(), providerAuthorityGroup, evidence.GroupGeneration,
		map[string]int64{"billing": evidence.BillingGeneration},
	)
	if err != nil {
		return verifiedProviderTarget{}, fmt.Errorf("provider target capability is invalid")
	}
	return verifiedProviderTarget{
		tenantID: evidence.TenantID, mappingID: evidence.MappingID,
		providerAccountID: evidence.ProviderAccountID,
		subscriptionID:    evidence.SubscriptionID, invoiceID: evidence.InvoiceID,
		mappingRevision: evidence.MappingRevision,
		lookupBinding:   evidence.LookupBindingSHA256, operation: operation, issued: true,
	}, nil
}

func validProviderEvidence(value ProviderTargetEvidence) bool {
	return value.TenantID != uuid.Nil && value.MappingID != uuid.Nil &&
		value.ProviderAccountID != uuid.Nil && value.MappingRevision > 0 &&
		value.GroupGeneration > 0 && value.BillingGeneration > 0 &&
		value.LookupBindingSHA256 != [32]byte{}
}

func providerEvidenceMatchesLookup(
	lookup ProviderLookup,
	evidence ProviderTargetEvidence,
) bool {
	switch lookup.ObjectKind() {
	case "customer":
		return evidence.SubscriptionID == uuid.Nil && evidence.InvoiceID == uuid.Nil
	case "subscription":
		return evidence.SubscriptionID != uuid.Nil
	case "invoice", "payment":
		return evidence.InvoiceID != uuid.Nil
	default:
		return false
	}
}

func (target verifiedProviderTarget) valid() bool {
	return target.issued && target.tenantID != uuid.Nil && target.mappingID != uuid.Nil &&
		target.providerAccountID != uuid.Nil && target.mappingRevision > 0 &&
		target.lookupBinding != [32]byte{} && target.operation.valid()
}
