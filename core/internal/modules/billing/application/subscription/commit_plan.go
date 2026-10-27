package subscriptionapp

type CommitPlan struct {
	BeforeRevision int64
	After          SubscriptionSnapshot
	Receipt        OperationReceipt
	Audit          AuditIntent
	Outbox         []OutboxIntent
	Jobs           []JobIntent
	CancelJobs     []JobCancellation
	CompleteJob    *JobCompletion
	Change         *SubscriptionChange
	ResolveChange  *ChangeResolution
	TrialClaim     *TrialClaim
	Replay         bool
}

// Valid reports whether every planner-owned invariant is satisfied. Storage
// adapters still have to verify that the plan belongs to their bound execution.
func (plan CommitPlan) Valid() bool { return plan.valid() }

func (plan CommitPlan) valid() bool {
	if plan.Replay {
		return plan.After.valid() && plan.Receipt.valid() &&
			plan.After.Revision == plan.BeforeRevision && plan.emptyEffects()
	}
	if plan.BeforeRevision < 1 || !plan.After.valid() ||
		plan.After.Revision != plan.BeforeRevision+1 || !plan.Receipt.valid() ||
		plan.Receipt.ResultSHA256 != snapshotDigest(plan.After) || !plan.Audit.valid() {
		return false
	}
	for _, event := range plan.Outbox {
		if !event.valid() || event.AggregateRevision != plan.After.Revision {
			return false
		}
	}
	for _, job := range plan.Jobs {
		if !job.valid() {
			return false
		}
	}
	for _, job := range plan.CancelJobs {
		if !job.valid() {
			return false
		}
	}
	return (plan.CompleteJob == nil || plan.CompleteJob.valid()) &&
		(plan.Change == nil || plan.Change.valid()) &&
		(plan.ResolveChange == nil || plan.ResolveChange.valid()) &&
		(plan.TrialClaim == nil || plan.TrialClaim.valid())
}

func (plan CommitPlan) emptyEffects() bool {
	return plan.Audit == (AuditIntent{}) && len(plan.Outbox) == 0 &&
		len(plan.Jobs) == 0 && len(plan.CancelJobs) == 0 &&
		plan.CompleteJob == nil && plan.Change == nil &&
		plan.ResolveChange == nil && plan.TrialClaim == nil
}

type Result struct {
	Subscription SubscriptionSnapshot
	Replay       bool
}
