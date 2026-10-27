package asyncmanifestv1

import jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"

const (
	WorkProjectionRepairJobKind      = "work.projection.repair"
	WorkPurgeExecuteJobKind          = "work.purge.execute"
	WorkRecurrenceMaterializeJobKind = "work.recurrence.materialize"
	WorkReminderDispatchJobKind      = "work.reminder.dispatch"
)

func WorkJobPolicies() []jobsv1.KindPolicy {
	kinds := []string{
		WorkProjectionRepairJobKind,
		WorkPurgeExecuteJobKind,
		WorkRecurrenceMaterializeJobKind,
		WorkReminderDispatchJobKind,
	}
	policies := make([]jobsv1.KindPolicy, 0, len(kinds))
	for _, kind := range kinds {
		policy, _ := jobsv1.NewKindPolicy(kind, 1, "work",
			jobsv1.EffectDatabaseTransactional, jobsv1.DefaultMaximumAttempts)
		policies = append(policies, policy)
	}
	return policies
}
