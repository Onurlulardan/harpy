package asyncmanifestv1

import jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"

const (
	WorkflowInstanceExecuteJobKind = "workflow.instance.execute"
	WorkflowScheduleFireJobKind    = "workflow.schedule.fire"
	WorkflowTimerResumeJobKind     = "workflow.timer.resume"
)

func WorkflowInstanceExecuteJobPolicy() jobsv1.KindPolicy {
	return workflowDatabaseJobPolicy(WorkflowInstanceExecuteJobKind)
}

func WorkflowScheduleFireJobPolicy() jobsv1.KindPolicy {
	return workflowDatabaseJobPolicy(WorkflowScheduleFireJobKind)
}

func WorkflowTimerResumeJobPolicy() jobsv1.KindPolicy {
	return workflowDatabaseJobPolicy(WorkflowTimerResumeJobKind)
}

func WorkflowJobPolicies() []jobsv1.KindPolicy {
	return []jobsv1.KindPolicy{
		WorkflowInstanceExecuteJobPolicy(),
		WorkflowScheduleFireJobPolicy(),
		WorkflowTimerResumeJobPolicy(),
	}
}

func workflowDatabaseJobPolicy(kind string) jobsv1.KindPolicy {
	policy, _ := jobsv1.NewKindPolicy(
		kind, 1, "workflow", jobsv1.EffectDatabaseTransactional,
		jobsv1.DefaultMaximumAttempts,
	)
	return policy
}
