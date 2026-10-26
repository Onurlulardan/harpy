package asyncmanifestv1

import (
	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

const (
	BillingSubscriptionBoundaryJobKind = "billing.subscription.boundary"
	BillingWebhookProcessorJobKind     = "billing.webhook.process"
	BillingProviderCommandJobKind      = "billing.provider.command"
	BillingProviderFinalizeJobKind     = "billing.provider.finalize"
	BillingProviderReconcileJobKind    = "billing.provider.reconcile"
	BillingRetentionJobKind            = "billing.retention.purge"
)

func BillingWebhookProcessorJobPolicy() jobsv1.KindPolicy {
	return billingJobPolicy(
		BillingWebhookProcessorJobKind,
		jobsv1.EffectDatabaseTransactional,
	)
}

func BillingProviderCommandJobPolicy() jobsv1.KindPolicy {
	return billingJobPolicy(BillingProviderCommandJobKind, jobsv1.EffectExternal)
}

func BillingProviderFinalizeJobPolicy() jobsv1.KindPolicy {
	return billingJobPolicy(
		BillingProviderFinalizeJobKind,
		jobsv1.EffectDatabaseTransactional,
	)
}

func BillingProviderReconcileJobPolicy() jobsv1.KindPolicy {
	return billingJobPolicy(BillingProviderReconcileJobKind, jobsv1.EffectExternal)
}

func BillingRetentionJobPolicy() jobsv1.KindPolicy {
	return billingJobPolicy(BillingRetentionJobKind, jobsv1.EffectDatabaseTransactional)
}

func BillingProviderJobPolicies() []jobsv1.KindPolicy {
	return []jobsv1.KindPolicy{
		BillingWebhookProcessorJobPolicy(),
		BillingProviderCommandJobPolicy(),
		BillingProviderFinalizeJobPolicy(),
		BillingProviderReconcileJobPolicy(),
	}
}

func BillingJobPolicies() []jobsv1.KindPolicy {
	policies := []jobsv1.KindPolicy{billingJobPolicy(
		BillingSubscriptionBoundaryJobKind,
		jobsv1.EffectDatabaseTransactional,
	)}
	policies = append(policies, BillingProviderJobPolicies()...)
	return append(policies, BillingRetentionJobPolicy())
}

func BillingProducerPolicies() []eventsv1.ProducerPolicy {
	definitions := [][2]string{
		{"billing.entitlements.changed", "ae7a1c1cb43f0d58d7cf292f5d92e099ba3534eac709fa54da6c4f6bf6e2403e"},
		{"billing.invoice.finalized", "b55f45f8a7321fd9a94a5436782eb8a116b75f5cbfb72d9dbc4d735c51eb5264"},
		{"billing.payment.changed", "3ce299b97ab5b34969096bd0ebc9e2631bb4ed10ce88bffd56fdc5336edcbb00"},
		{"billing.subscription.changed", "fc511ef78cedb23785e362d889271374bc30f9446e221b4fdb742130ecba9dc3"},
		{"billing.usage.threshold_reached", "69e78f61bc3db12a3c72883fab952d6d87a29f9ceafec54bb3772fa03647d591"},
	}
	policies := make([]eventsv1.ProducerPolicy, 0, len(definitions))
	for _, definition := range definitions {
		policy, _ := eventsv1.NewProducerPolicy(
			definition[0], 1, "billing", definition[1],
		)
		policies = append(policies, policy)
	}
	return policies
}

func billingJobPolicy(kind string, effect jobsv1.EffectClass) jobsv1.KindPolicy {
	policy, _ := jobsv1.NewKindPolicy(
		kind, 1, "billing", effect, jobsv1.DefaultMaximumAttempts,
	)
	return policy
}
