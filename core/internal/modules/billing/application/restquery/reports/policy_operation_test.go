package reports

import "testing"

func TestSubscriptionPolicyAllowsOnlyListAndTransitionReadback(t *testing.T) {
	t.Parallel()
	spec := subscriptionSpec()
	for _, operation := range []string{
		SubscriptionOperation, SubscriptionTransitionOperation,
	} {
		if !allowedReportOperation(spec, operation) {
			t.Fatalf("operation %q was rejected", operation)
		}
	}
	for _, operation := range []string{
		AuditOperation, RevenueOperation, "billing.report.subscription.delete",
	} {
		if allowedReportOperation(spec, operation) {
			t.Fatalf("operation %q was widened", operation)
		}
	}
	if allowedReportOperation(auditSpec(), SubscriptionTransitionOperation) {
		t.Fatal("subscription transition widened audit report")
	}
}
