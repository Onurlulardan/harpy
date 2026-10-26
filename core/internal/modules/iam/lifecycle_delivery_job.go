package iam

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	lifecycledeliveryv1 "github.com/Onurlulardan/harpy/contracts/lifecycledelivery/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
	"github.com/google/uuid"
)

const LifecycleDeliveryJobKind = "iam.lifecycle.delivery"

func (sink *LifecycleChangeSink) deliveryJob(change lifecycle.Change,
	tenantID uuid.UUID, cutoverGeneration, groupGeneration int64,
) (jobsv1.Enqueue, bool, error) {
	if change.DeliveryArtifactRef == "" {
		return jobsv1.Enqueue{}, false, nil
	}
	reference, err := lifecycledeliveryv1.ParseReference(change.DeliveryArtifactRef)
	if err != nil || reference.TenantID() != tenantID {
		return jobsv1.Enqueue{}, false, lifecycle.ErrEffects
	}
	jobs, _, err := sink.catalogs.Snapshot()
	if err != nil {
		return jobsv1.Enqueue{}, false, err
	}
	tenant := tenantID.String()
	digest := sha256.Sum256([]byte(
		"iam.lifecycle.delivery.v1\x00" + change.DeliveryArtifactRef,
	))
	effectKey := "iam.lifecycle.delivery:" + change.DeliveryArtifactRef
	value := jobsv1.Enqueue{
		JobID: change.OperationID.String(), Kind: LifecycleDeliveryJobKind,
		PayloadRevision: 1, TenantID: &tenant,
		Authority: jobsv1.Authority{
			ScopeKind: jobsv1.ScopeTenant, ScopeID: tenant, Domain: "iam",
			CutoverGeneration: cutoverGeneration, WriterOrigin: jobsv1.WriterV2,
			GroupID: "iam_system_resources", GroupGeneration: groupGeneration,
		},
		PayloadReference: change.DeliveryArtifactRef,
		PayloadSHA256:    hex.EncodeToString(digest[:]),
		RunAt:            sink.now().UTC().Format(time.RFC3339Nano), Priority: 100,
		DedupeKey: change.OperationID.String(), EffectIdempotencyKey: &effectKey,
		MaximumAttempts: jobsv1.DefaultMaximumAttempts,
	}
	job, err := jobsv1.NewEnqueue(value, jobs)
	return job, true, err
}
