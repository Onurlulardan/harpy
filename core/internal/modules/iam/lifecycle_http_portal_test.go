package iam

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
)

func TestCanonicalPortalApprovalDecisionMapsPublicActions(t *testing.T) {
	for input, expected := range map[string]lifecycle.PortalApprovalDecision{
		"approve": lifecycle.PortalApprovalApproved,
		"reject":  lifecycle.PortalApprovalRejected,
	} {
		actual, valid := canonicalPortalApprovalDecision(input)
		if !valid || actual != expected {
			t.Fatalf("decision %q = %q/%t, want %q/true",
				input, actual, valid, expected)
		}
	}
	for _, invalid := range []string{"", "approved", "rejected", "revision_required"} {
		if actual, valid := canonicalPortalApprovalDecision(invalid); valid || actual != "" {
			t.Fatalf("invalid decision %q = %q/%t", invalid, actual, valid)
		}
	}
}
