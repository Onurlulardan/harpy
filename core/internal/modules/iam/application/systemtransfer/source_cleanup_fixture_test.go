package systemtransfer

import (
	"context"
	"crypto/sha256"
	"sync"
	"time"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/google/uuid"
)

type importCleanupStoreStub struct {
	mutex       sync.Mutex
	scans       []SourceCleanupScan
	claim       SourceCleanupClaim
	completion  SourceCleanupCompletion
	retry       SourceCleanupRetry
	completeHit chan struct{}
	retryHit    chan struct{}
	claimHit    chan struct{}
	finalErr    error
	retryErr    error
}

func (store *importCleanupStoreStub) AcquireSourceCleanup(
	_ context.Context, claim SourceCleanupClaim,
) (SourceCleanupScan, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	store.claim = claim
	if store.claimHit != nil {
		select {
		case store.claimHit <- struct{}{}:
		default:
		}
	}
	if len(store.scans) == 0 {
		return SourceCleanupScan{}, nil
	}
	scan := store.scans[0]
	store.scans = store.scans[1:]
	return scan, nil
}

func (store *importCleanupStoreStub) PersistSourceCleanupCompletion(
	_ context.Context, completion SourceCleanupCompletion,
) (bool, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	store.completion = completion
	if store.completeHit != nil {
		select {
		case store.completeHit <- struct{}{}:
		default:
		}
	}
	return false, store.finalErr
}

func (store *importCleanupStoreStub) ScheduleSourceCleanupRetry(
	_ context.Context, retry SourceCleanupRetry,
) (bool, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	store.retry = retry
	if store.retryHit != nil {
		select {
		case store.retryHit <- struct{}{}:
		default:
		}
	}
	return false, store.retryErr
}

type importCleanupObjectStub struct {
	key, version string
	result       SourceDeleteResult
	err          error
}

func (objects *importCleanupObjectStub) DeleteExactImportSource(
	_ context.Context, key, version string,
) (SourceDeleteResult, error) {
	objects.key, objects.version = key, version
	return objects.result, objects.err
}

type importCleanupValuesStub struct {
	now time.Time
	id  uuid.UUID
}

func (values importCleanupValuesStub) Now() time.Time { return values.now }
func (values importCleanupValuesStub) NewV7(time.Time) (uuid.UUID, error) {
	return values.id, nil
}

type importCleanupJitterStub struct{ value uint64 }

func (jitter importCleanupJitterStub) Uint64N(bound uint64) uint64 {
	if jitter.value >= bound {
		return bound - 1
	}
	return jitter.value
}

func importCleanupFixture() (SourceCleanupTarget,
	SourceDeleteResult, time.Time,
) {
	now := time.Date(2026, 7, 18, 9, 10, 11, 123000000, time.UTC)
	sourceID := uuid.MustParse("019be200-0000-7000-8000-000000000001")
	target := SourceCleanupTarget{
		TenantID: uuid.MustParse("019be200-0000-7000-8000-000000000002"),
		ActorID:  uuid.MustParse("019be200-0000-7000-8000-000000000003"),
		SourceID: sourceID, Resource: publicresource.Users,
		ObjectKey: importSourceObjectKey(sourceID), ObjectVersion: "exact-version-1",
		LeaseOwnerID: uuid.MustParse("019be200-0000-7000-8000-000000000004"),
		Attempt:      2, LeaseGeneration: 3,
	}
	result := SourceDeleteResult{ProviderRequestID: "delete-request-1",
		DeleteResponseSHA256: sha256.Sum256([]byte("delete")),
		AbsenceProbeSHA256:   sha256.Sum256([]byte("absent")),
		VerifiedAbsentAt:     now}
	return target, result, now
}
