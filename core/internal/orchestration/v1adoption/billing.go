package v1adoption

import (
	"sort"
	"time"

	"github.com/google/uuid"
)

const (
	SaaSControlOfferingKey = "saas_platform_control_base_v1"
	SaaSFreeOfferingKey    = "saas_free_base_v1"
)

func baseAssignments(value Inventory, bindings []TenantBinding) (
	[]BaseAssignment, []Issue,
) {
	evidence := map[string][]BillingEvidence{}
	for _, item := range value.Billing {
		key := item.TenantID.String()
		evidence[key] = append(evidence[key], item)
	}
	assignments := make([]BaseAssignment, 0, len(bindings))
	issues := make([]Issue, 0)
	for _, binding := range bindings {
		items := evidence[binding.TenantID.String()]
		if len(items) != 1 {
			issues = append(issues, Issue{Code: IssueBaseEvidenceAmbiguous,
				Tenant: binding.TenantID.String()})
			continue
		}
		item := items[0]
		if code := validBaseEvidence(value.Mode, binding.Purpose, item); code != "" {
			issues = append(issues, Issue{Code: code, Tenant: binding.TenantID.String()})
			continue
		}
		assignments = append(assignments, BaseAssignment{
			TenantID: binding.TenantID, Kind: item.Kind, Offering: item.Offering,
			SubscriptionEvidenceSHA256: item.SubscriptionEvidenceSHA256,
		})
	}
	sort.Slice(assignments, func(i, j int) bool {
		return assignments[i].TenantID.String() < assignments[j].TenantID.String()
	})
	return assignments, issues
}

func validBaseEvidence(mode DeploymentMode, purpose string, item BillingEvidence) string {
	if item.TenantID == uuid.Nil || !validPin(item.Offering) {
		return IssueBaseEvidenceAmbiguous
	}
	if mode == ModeOnPremise {
		if item.Kind != BillingLicenseDerived || !item.VerifiedV2Publication ||
			item.Offering.SourceSequence < 1 || !validSHA(item.Offering.SourceSHA256) {
			return IssueLicenseExchangeRequired
		}
		return ""
	}
	switch purpose {
	case "saas_control":
		if item.Kind != BillingSaaSControl || item.Offering.Key != SaaSControlOfferingKey ||
			!item.VerifiedV2Publication || item.SubscriptionEvidenceSHA256 != "" ||
			!validSaaSSource(item.Offering) {
			return IssueBaseEvidenceAmbiguous
		}
	case "saas_customer":
		if item.Kind == BillingSaaSFree {
			if item.Offering.Key != SaaSFreeOfferingKey || !item.VerifiedV2Publication ||
				item.SubscriptionEvidenceSHA256 != "" || !validSaaSSource(item.Offering) {
				return IssueBaseEvidenceAmbiguous
			}
		} else if item.Kind != BillingSaaSPaid ||
			!item.VerifiedV2Publication || !validSHA(item.SubscriptionEvidenceSHA256) ||
			!validSaaSSource(item.Offering) {
			return IssueBaseEvidenceAmbiguous
		}
	default:
		return IssueBaseEvidenceAmbiguous
	}
	return ""
}

func validPin(pin OfferingPin) bool {
	return pin.VersionID != uuid.Nil && pin.Revision > 0 && validSHA(pin.SHA256) &&
		pin.Key != "" && !pin.EffectiveAt.IsZero() && pin.EffectiveAt.Location() == time.UTC
}

func validSaaSSource(pin OfferingPin) bool {
	return pin.SourceSequence == 0 && pin.SourceSHA256 == ""
}
