package asyncmanifestv1

import jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"

const NotificationsRetentionJobKind = "notifications.retention.purge"

func NotificationsJobPolicies() []jobsv1.KindPolicy {
	definitions := []struct {
		kind   string
		effect jobsv1.EffectClass
	}{
		{"notifications.email.delivery", jobsv1.EffectExternal},
		{NotificationsRetentionJobKind, jobsv1.EffectDatabaseTransactional},
		{"notifications.telegram.delivery", jobsv1.EffectExternal},
	}
	result := make([]jobsv1.KindPolicy, 0, len(definitions))
	for _, definition := range definitions {
		policy, _ := jobsv1.NewKindPolicy(
			definition.kind, 1, "notifications", definition.effect, 8)
		result = append(result, policy)
	}
	return result
}
