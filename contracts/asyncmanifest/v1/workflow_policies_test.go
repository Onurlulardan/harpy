package asyncmanifestv1

import (
	"testing"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

func TestWorkflowJobPoliciesAreClosedAndDatabaseTransactional(t *testing.T) {
	t.Parallel()
	want := []string{
		WorkflowInstanceExecuteJobKind,
		WorkflowScheduleFireJobKind,
		WorkflowTimerResumeJobKind,
	}
	policies := WorkflowJobPolicies()
	if len(policies) != len(want) {
		t.Fatalf("policy count = %d", len(policies))
	}
	for index, policy := range policies {
		if policy.Kind != want[index] || policy.PayloadRevision != 1 ||
			policy.OwnerDomain != "workflow" ||
			policy.EffectClass != jobsv1.EffectDatabaseTransactional ||
			policy.MaximumAttempts != jobsv1.DefaultMaximumAttempts ||
			policy.Validate() != nil {
			t.Fatalf("policy %d = %+v", index, policy)
		}
	}
}

func TestReleaseManifestContainsWorkflowJobs(t *testing.T) {
	t.Parallel()
	manifest, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{
		WorkflowInstanceExecuteJobKind,
		WorkflowScheduleFireJobKind,
		WorkflowTimerResumeJobKind,
	} {
		worker, workerErr := manifest.Jobs().Resolve(kind, 1)
		writer, writerErr := manifest.WritableJobs().Resolve(kind, 1)
		if workerErr != nil || writerErr != nil || worker != writer {
			t.Fatalf("job %s = %+v/%+v (%v/%v)", kind, worker, writer, workerErr, writerErr)
		}
	}
}
