package systemtransfer

import (
	"context"
	"crypto/sha256"
	"testing"
	"time"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

var fixtureExportCatalogID = uuid.MustParse("019be000-0000-7000-8000-000000000010")
var fixtureExportReceiptID = uuid.MustParse("019be000-0000-7000-8000-000000000012")

type exportClockFixture struct{ now time.Time }

func (clock exportClockFixture) Now() time.Time { return clock.now }

type exportIdentityFixture struct{ identity uuid.UUID }

func (source exportIdentityFixture) NewExportIdentity(time.Time) (uuid.UUID, error) {
	return source.identity, nil
}

type exportLoaderFixture struct{ attempt ExportAttempt }

func (loader exportLoaderFixture) LoadExportAttempt(
	context.Context,
	ExportJobClaim,
) (ExportAttempt, error) {
	return loader.attempt, nil
}

type exportLeaseFixture struct{ calls int }

func (lease *exportLeaseFixture) CheckpointExportLease(
	context.Context,
	ExportJobClaim,
) error {
	lease.calls++
	return nil
}

type exportScopesFixture struct{}

func (exportScopesFixture) InFreshExportCheckpoint(
	ctx context.Context,
	attempt ExportAttempt,
	command func(context.Context, ExportScope) error,
) error {
	return command(ctx, fixtureBatchScope{attempt.TenantID, attempt.ActorID})
}

type exportAuthorizerFixture struct{}

func (exportAuthorizerFixture) AuthorizeExportCheckpoint(
	context.Context,
	platformtx.TenantScope,
	ExportAttempt,
) (publicresource.ActorEvidence, error) {
	return publicresource.NewActorEvidence(publicresource.ID(fixtureTenantID),
		publicresource.ID(fixtureActorID), 3, 7)
}

type exportStoreFixture struct {
	spool       *exportSpoolFixture
	completed   int
	completeErr error
	starts      []ExportArtifactEffectStart
	unknown     []ExportOutcomeUnknownRecord
	finished    []ExportFailure
	completion  ExportCompletion
}

func (store *exportStoreFixture) AcquireOutbound(
	context.Context,
	ExportAttempt,
) (ExportClaimResult, error) {
	return ExportClaimResult{State: ExportRunning}, nil
}

func (*exportStoreFixture) CheckpointExport(
	context.Context,
	ExportAttempt,
	ExportProgress,
) error {
	return nil
}

func (store *exportStoreFixture) BeginArtifactEffect(
	_ context.Context,
	start ExportArtifactEffectStart,
) (bool, error) {
	store.starts = append(store.starts, start)
	return false, nil
}

func (store *exportStoreFixture) PersistOutboundSuccess(
	_ context.Context,
	completion ExportCompletion,
) (bool, error) {
	if store.spool == nil || !store.spool.discarded {
		return false, ErrExportStateConflict
	}
	if store.completeErr != nil {
		return false, store.completeErr
	}
	store.completed++
	store.completion = completion
	return false, nil
}

func (store *exportStoreFixture) PersistOutboundFailure(
	_ context.Context,
	failure ExportFailure,
) error {
	store.finished = append(store.finished, failure)
	return nil
}

func (store *exportStoreFixture) PersistOutcomeUnknown(
	_ context.Context,
	record ExportOutcomeUnknownRecord,
) error {
	store.unknown = append(store.unknown, record)
	return nil
}

type exportStoreBinderFixture struct{ store *exportStoreFixture }

func (binder exportStoreBinderFixture) BindExportWorker(
	context.Context,
	platformtx.TenantScope,
) (ExportWorkerStore, error) {
	return binder.store, nil
}

func fixtureExportAttempt(t *testing.T, now time.Time) ExportAttempt {
	t.Helper()
	query := []byte(`{"select":["id"],"sort":[{"direction":"asc","field":"id"}]}`)
	attempt := ExportAttempt{TenantID: fixtureTenantID, ActorID: fixtureActorID,
		JobID: fixtureJobID, ExportID: fixtureTransferID,
		CatalogVersionID: fixtureExportCatalogID, Resource: publicresource.Users,
		Format: "jsonl", State: ExportWaiting, TokenVersion: 3,
		AuthorizationVersion: 7, CatalogSHA256: sha256.Sum256([]byte("catalog-v1")),
		QuerySHA256: sha256.Sum256(query), QueryDocument: query}
	if attempt.CheckOwner() != nil || !exportArtifactTime(now) {
		t.Fatal("export fixture is invalid")
	}
	return attempt
}

func fixtureExportClaim() ExportJobClaim {
	return ExportJobClaim{TenantID: fixtureTenantID, JobID: fixtureJobID,
		ExportID: fixtureTransferID, AttemptID: fixtureAttemptID,
		LeaseOwnerID: fixtureLeaseID, Resource: publicresource.Users,
		AttemptNumber: 1, MaximumAttempts: SystemExportMaximumAttempts,
		LeaseGeneration: 1}
}
