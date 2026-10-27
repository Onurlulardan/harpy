package providerapp

import (
	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

const (
	WebhookProcessorJobKind       = asyncmanifestv1.BillingWebhookProcessorJobKind
	ProviderCommandJobKind        = asyncmanifestv1.BillingProviderCommandJobKind
	ProviderFinalizeJobKind       = asyncmanifestv1.BillingProviderFinalizeJobKind
	ProviderReconciliationJobKind = asyncmanifestv1.BillingProviderReconcileJobKind
)

func WebhookProcessorPolicy() jobsv1.KindPolicy {
	return asyncmanifestv1.BillingWebhookProcessorJobPolicy()
}

func ProviderCommandPolicy() jobsv1.KindPolicy {
	return asyncmanifestv1.BillingProviderCommandJobPolicy()
}

func ProviderFinalizePolicy() jobsv1.KindPolicy {
	return asyncmanifestv1.BillingProviderFinalizeJobPolicy()
}

func ProviderReconciliationPolicy() jobsv1.KindPolicy {
	return asyncmanifestv1.BillingProviderReconcileJobPolicy()
}

func JobPolicies() []jobsv1.KindPolicy {
	return asyncmanifestv1.BillingProviderJobPolicies()
}
