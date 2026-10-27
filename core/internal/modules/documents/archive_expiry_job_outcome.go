package documents

import (
	"time"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/api/jobcontract"
	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/google/uuid"
)

func archiveExpiryOutcome(tenantIDString string, execution archiveapp.ExpiryExecution,
	now time.Time,
) (platformasync.DatabaseJobOutcome, error) {
	if !execution.Created {
		return platformasync.CompleteDatabaseJob(), nil
	}
	tenantID, err := parseDocumentID(tenantIDString)
	if err != nil || !execution.DeletionID.UUIDv7() {
		return platformasync.DatabaseJobOutcome{}, archiveapp.ErrInvalid
	}
	reference, digest, err := jobcontract.Deletion(jobcontract.ID(tenantID),
		jobcontract.ID(execution.DeletionID))
	if err != nil {
		return platformasync.DatabaseJobOutcome{}, err
	}
	key := "documents:blob-delete:" + execution.DeletionID.String()
	child := platformasync.DatabaseJobChildInput{
		JobID: uuid.UUID(execution.DeletionID), Kind: asyncmanifestv1.DocumentsBlobDeleteJobKind,
		PayloadRevision: 1, RunAt: now.UTC().Truncate(time.Microsecond),
		PayloadReference: reference, PayloadSHA256: digest, DedupeKey: key,
		EffectIdempotencyKey: key, Priority: 100,
	}
	return platformasync.CompleteDatabaseJobWithEffects(
		platformasync.DatabaseJobEffects{Children: []platformasync.DatabaseJobChildInput{child}})
}
