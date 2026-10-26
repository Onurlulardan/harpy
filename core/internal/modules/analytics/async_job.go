package analytics

import (
	"crypto/sha256"
	"fmt"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (effects *AsyncEffects) job(
	authority platformtx.OwnerAuthority, kind, reference, dedupe string,
	runAt time.Time,
) (jobsv1.Enqueue, error) {
	jobID, err := effects.ids.NewUUIDv7()
	if err != nil {
		return jobsv1.Enqueue{}, err
	}
	jobs, _, err := effects.catalogs.Snapshot()
	if err != nil {
		return jobsv1.Enqueue{}, err
	}
	policy, err := jobs.Resolve(kind, 1)
	if err != nil {
		return jobsv1.Enqueue{}, err
	}
	tenantID := authority.ScopeID().String()
	digest := sha256.Sum256([]byte(reference))
	return jobsv1.NewEnqueue(jobsv1.Enqueue{
		JobID: jobID.String(), Kind: kind, PayloadRevision: 1,
		TenantID: &tenantID, Authority: analyticsJobAuthority(authority),
		PayloadReference: reference, PayloadSHA256: fmt.Sprintf("%x", digest),
		RunAt: runAt.Format(time.RFC3339Nano), Priority: 100, DedupeKey: dedupe,
		MaximumAttempts: policy.MaximumAttempts,
	}, jobs)
}

func analyticsJobAuthority(authority platformtx.OwnerAuthority) jobsv1.Authority {
	return jobsv1.Authority{ScopeKind: jobsv1.ScopeTenant,
		ScopeID: authority.ScopeID().String(), Domain: authority.Domain(),
		CutoverGeneration: authority.Generation(), WriterOrigin: jobsv1.WriterV2,
		GroupID: authority.GroupID(), GroupGeneration: authority.GroupGeneration()}
}
