package asyncmanifestv1

import (
	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

// Release policy lists are closed. A domain is added only when its executable
// worker and API producer ship in this same release artifact.
func jobPolicies() []jobsv1.KindPolicy {
	kinds := []string{
		"iam.credential.bcrypt-retire",
		"iam.directory.checksum",
		"iam.directory.repair",
		"iam.directory.tombstone-purge",
		"iam.invitation.delivery",
		"iam.lifecycle.delivery",
		"iam.identifier.retire",
		"iam.profile.backfill",
		"iam.security-artifact-purge",
	}
	policies := BillingJobPolicies()
	policies = append(policies, AnalyticsJobPolicies()...)
	policies = append(policies, CatalogJobPolicies()...)
	policies = append(policies, WorkJobPolicies()...)
	policies = append(policies, DocumentsJobPolicies()...)
	policies = append(policies, NotificationsJobPolicies()...)
	policies = append(policies, IAMSystemTransferJobPolicies()...)
	for _, kind := range kinds {
		effect := jobsv1.EffectDatabaseTransactional
		if kind == "iam.invitation.delivery" || kind == "iam.lifecycle.delivery" {
			effect = jobsv1.EffectExternal
		}
		policy, _ := jobsv1.NewKindPolicy(
			kind, 1, "iam", effect, jobsv1.DefaultMaximumAttempts,
		)
		policies = append(policies, policy)
	}
	return policies
}

func producerPolicies() []eventsv1.ProducerPolicy {
	definitions := [][2]string{
		{"installation.activated", "4a44e2f3f9e2669280d713512812010d50bf9f077f4ee1a0709e80c3633dcb5e"},
		{"iam.identity.changed", "e3c637083d0a126cf82ecd547e147bffa4bda5c9ae19ea7b89135710b1b71e6b"},
		{"iam.invitation.changed", "3dbc2634af290b10dc25fad1db39d01f51d5acd5c07c920288be004c9d0a531b"},
		{"iam.lifecycle.changed", "f5100cde78ae1d5057c538d22ee199765816f6648650a19f7f11526a46ff8230"},
		{"iam.permission_request.decided", "f42bbb6b9274fc239afed496f6553340cc514a5bd8eb5c1549d002fe16e4997d"},
		{"iam.permission_request.transitioned", "65c815b2163d6f3ad5ee04f6615b1cf92b4d2161bee3796f1e49fd3872c5a399"},
		{"iam.rbac.changed", "ec25ad29fe3139a6db7b0ee837d3b9a28371867a671c88322a1bfa14dca64657"},
	}
	policies := BillingProducerPolicies()
	policies = append(policies, AnalyticsProducerPolicies()...)
	policies = append(policies, CatalogProducerPolicies()...)
	policies = append(policies, WorkProducerPolicies()...)
	for _, definition := range definitions {
		policy, _ := eventsv1.NewProducerPolicy(definition[0], 1, "iam", definition[1])
		policies = append(policies, policy)
	}
	return policies
}

func consumerPolicies() []eventsv1.ConsumerPolicy { return nil }
