package providerapp

import (
	"regexp"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

var lowerSHA256Pattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

type AsyncJob struct {
	policy      jobsv1.KindPolicy
	payloadRef  string
	payloadHash string
	runAt       time.Time
	dedupeKey   string
	effectKey   string
}

func newAsyncJob(
	policy jobsv1.KindPolicy,
	payloadRef string,
	payloadHash string,
	runAt time.Time,
	dedupeKey string,
	effectKey string,
) (AsyncJob, error) {
	canonicalRunAt, err := canonicalInstant(runAt)
	if err != nil || policy.Validate() != nil || payloadRef == "" ||
		len(payloadRef) > jobsv1.MaximumReferenceBytes ||
		!lowerSHA256Pattern.MatchString(payloadHash) || dedupeKey == "" ||
		len(dedupeKey) > jobsv1.MaximumKeyBytes ||
		(policy.EffectClass == jobsv1.EffectExternal) != (effectKey != "") ||
		len(effectKey) > jobsv1.MaximumKeyBytes {
		return AsyncJob{}, ErrWebhookDecision
	}
	return AsyncJob{
		policy: policy, payloadRef: payloadRef, payloadHash: payloadHash,
		runAt: canonicalRunAt, dedupeKey: dedupeKey, effectKey: effectKey,
	}, nil
}

func (job AsyncJob) Policy() jobsv1.KindPolicy { return job.policy }
func (job AsyncJob) PayloadReference() string  { return job.payloadRef }
func (job AsyncJob) PayloadSHA256() string     { return job.payloadHash }
func (job AsyncJob) RunAt() time.Time          { return job.runAt }
func (job AsyncJob) DedupeKey() string         { return job.dedupeKey }
func (job AsyncJob) EffectKey() string         { return job.effectKey }

func (job AsyncJob) valid() bool {
	validated, err := newAsyncJob(
		job.policy, job.payloadRef, job.payloadHash, job.runAt,
		job.dedupeKey, job.effectKey,
	)
	return err == nil && validated == job
}
