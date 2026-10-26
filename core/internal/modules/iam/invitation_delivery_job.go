package iam

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const InvitationDeliveryJobKind = "iam.invitation.delivery"

func (sink *InvitationChangeSink) deliveryJob(
	change invitation.AuditChange,
	scope platformtx.TenantScope,
	authority platformtx.OwnerAuthority,
) (jobsv1.Enqueue, bool, error) {
	if change.CommandKey != "iam.invitation.create" {
		return jobsv1.Enqueue{}, false, nil
	}
	jobs, _, err := sink.catalogs.Snapshot()
	if err != nil {
		return jobsv1.Enqueue{}, false, err
	}
	tenantID := scope.TenantID().String()
	digest := sha256.Sum256([]byte(
		"iam.invitation.delivery.v1\x00" + change.DeliveryArtifactRef,
	))
	effectKey := "iam.invitation.delivery:" + change.DeliveryArtifactRef
	value := jobsv1.Enqueue{
		JobID: change.OperationID.String(), Kind: InvitationDeliveryJobKind,
		PayloadRevision: 1, TenantID: &tenantID,
		Authority: jobsv1.Authority{
			ScopeKind: jobsv1.ScopeTenant, ScopeID: authority.ScopeID().String(),
			Domain: authority.Domain(), CutoverGeneration: authority.Generation(),
			WriterOrigin: jobsv1.WriterV2, GroupID: authority.GroupID(),
			GroupGeneration: authority.GroupGeneration(),
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
