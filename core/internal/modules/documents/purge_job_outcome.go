package documents

import (
	"time"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/api/jobcontract"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	purgeapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/purge"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/google/uuid"
)

const maximumPurgeChildren = 15

func (command *purgeDatabaseJobCommand) outcome(parsed jobcontract.PurgeReference,
	execution purgeapp.Execution,
) (platformasync.DatabaseJobOutcome, error) {
	if execution.Result.Purged() {
		return platformasync.CompleteDatabaseJob(), nil
	}
	now := command.clock.Now().UTC().Truncate(time.Microsecond)
	children := make([]platformasync.DatabaseJobChildInput, 0, len(execution.Deletions))
	for _, deletion := range execution.Deletions {
		reference, digest, err := jobcontract.Deletion(parsed.TenantID,
			jobcontract.ID(deletion.DeletionID))
		if err != nil || deletion.JobID != deletion.DeletionID {
			return platformasync.DatabaseJobOutcome{}, purgeapp.ErrInvalid
		}
		children = append(children, platformasync.DatabaseJobChildInput{
			JobID: uuid.UUID(deletion.JobID), Kind: asyncmanifestv1.DocumentsBlobDeleteJobKind,
			PayloadRevision: 1, RunAt: now, PayloadReference: reference,
			PayloadSHA256: digest, DedupeKey: "documents:blob-delete:" + deletion.DeletionID.String(),
			EffectIdempotencyKey: "documents:blob-delete:" + deletion.DeletionID.String(),
			Priority:             100,
		})
	}
	after := foundation.ID{}
	delay := 5 * time.Second
	if execution.More && len(execution.Deletions) > 0 {
		after = execution.Deletions[len(execution.Deletions)-1].DeletionID
		delay = time.Millisecond
	}
	reference, digest, err := jobcontract.PurgePage(parsed.TenantID,
		parsed.ResourceID, jobcontract.ID(after))
	if err != nil {
		return platformasync.DatabaseJobOutcome{}, err
	}
	nextID, err := command.ids.NewUUIDv7()
	if err != nil || !nextID.UUIDv7() {
		return platformasync.DatabaseJobOutcome{}, purgeapp.ErrUnavailable
	}
	successor := platformasync.DatabaseJobSuccessorInput{JobID: uuid.UUID(nextID),
		RunAt: now.Add(delay), PayloadReference: reference, PayloadSHA256: digest,
		DedupeKey: "documents:purge:" + parsed.ResourceID.String() + ":" + nextID.String(),
		Priority:  100}
	return platformasync.CompleteDatabaseJobWithEffects(platformasync.DatabaseJobEffects{
		Successor: &successor, Children: children,
	})
}
