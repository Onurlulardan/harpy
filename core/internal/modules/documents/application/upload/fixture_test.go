package upload

import (
	"context"
	"io"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

var (
	uploadTenant      = uploadID("11111111-1111-4111-8111-111111111111")
	uploadActor       = uploadID("22222222-2222-4222-8222-222222222222")
	uploadFolder      = uploadID("018f0000-0000-7000-8000-000000000601")
	uploadHandle      = uploadID("018f0000-0000-7000-8000-000000000602")
	uploadDocument    = uploadID("018f0000-0000-7000-8000-000000000603")
	uploadVersion     = uploadID("018f0000-0000-7000-8000-000000000604")
	uploadOperation   = uploadID("018f0000-0000-7000-8000-000000000605")
	uploadAudit       = uploadID("018f0000-0000-7000-8000-000000000606")
	uploadReservation = uploadID("018f0000-0000-7000-8000-000000000607")
	uploadReceipt     = uploadID("018f0000-0000-7000-8000-000000000608")
	uploadNow         = time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
)

func uploadID(value string) domain.ID {
	result, _ := domain.ParseID(value)
	return result
}

type uploadIDs struct{ values []domain.ID }

func (source *uploadIDs) NewUUIDv7() (domain.ID, error) {
	value := source.values[0]
	source.values = source.values[1:]
	return value, nil
}

type uploadClock struct{}

func (uploadClock) Now() time.Time { return uploadNow }

type uploadBilling struct {
	released domain.ID
}

func (billing *uploadBilling) Reserve(context.Context, platformtx.TenantScope,
	ReservationRequest,
) (Reservation, error) {
	return Reservation{ID: uploadReservation, Allowed: true}, nil
}
func (*uploadBilling) StartEffect(context.Context, platformtx.TenantScope, domain.ID, domain.ID) error {
	return nil
}
func (*uploadBilling) Commit(context.Context, platformtx.TenantScope, domain.ID, domain.ID,
	int64,
) error {
	return nil
}
func (billing *uploadBilling) Release(_ context.Context, _ platformtx.TenantScope,
	reservationID, _ domain.ID,
) error {
	billing.released = reservationID
	return nil
}

type uploadEffects struct{}

func (uploadEffects) Execute(context.Context, AttemptPlan, io.Reader) (EffectReceipt, *EffectFailure) {
	return EffectReceipt{}, &EffectFailure{Kind: FailureTransient}
}

type uploadJobs struct{}

func (uploadJobs) EnqueueDeletion(context.Context, platformtx.TenantScope,
	domain.ID, domain.ID, []byte,
) error {
	return nil
}

func (uploadJobs) EnqueueReconciliation(context.Context, platformtx.TenantScope,
	domain.ID, domain.ID, uint64, time.Time, []byte,
) error {
	return nil
}

type uploadStore struct {
	create func(context.Context, CreateMutation) (Snapshot, error)
	get    func(context.Context, domain.ID) (Snapshot, error)
	cancel func(context.Context, domain.ID, domain.ID, time.Time) (CancelResult, error)
}

func (store uploadStore) FindCreate(context.Context, string, [32]byte) (Snapshot, bool, error) {
	return Snapshot{}, false, nil
}
func (store uploadStore) Create(ctx context.Context, mutation CreateMutation) (Snapshot, error) {
	return store.create(ctx, mutation)
}
func (store uploadStore) Get(ctx context.Context, id domain.ID) (Snapshot, error) {
	return store.get(ctx, id)
}
func (uploadStore) Begin(context.Context, domain.ID, domain.ID, domain.ID) (AttemptPlan, error) {
	return AttemptPlan{}, ErrUnavailable
}
func (uploadStore) PersistEffect(context.Context, EffectReceipt) error { return ErrUnavailable }
func (uploadStore) PersistFailure(context.Context, EffectFailure, domain.ID) error {
	return ErrUnavailable
}
func (uploadStore) PrepareFinalize(context.Context, EffectReceipt) (FinalizeDecision, error) {
	return FinalizeDecision{}, ErrUnavailable
}
func (uploadStore) PersistFinalization(context.Context, EffectReceipt, domain.ID, domain.ID) (ReadyResult, error) {
	return ReadyResult{}, ErrUnavailable
}
func (store uploadStore) Cancel(ctx context.Context, id, receipt domain.ID,
	now time.Time,
) (CancelResult, error) {
	return store.cancel(ctx, id, receipt, now)
}

func uploadApplication(ids []domain.ID, billing *uploadBilling) *Application {
	return &Application{ids: &uploadIDs{values: ids}, clock: uploadClock{},
		billing: billing, jobs: uploadJobs{}, effects: uploadEffects{}}
}

func uploadIdentity() foundation.Identity {
	return foundation.Identity{ActorID: uploadActor, TenantID: uploadTenant}
}

func uploadSnapshot() Snapshot {
	return Snapshot{ID: uploadHandle, TenantID: uploadTenant, ActorID: uploadActor,
		DocumentID: uploadDocument, CandidateVersionID: uploadVersion, FolderID: uploadFolder,
		BillingReservationID: uploadReservation, ExpectedBytes: 7,
		State: domain.UploadCreated, ExpiresAt: uploadNow.Add(UploadKeep),
		CreatedAt: uploadNow, UpdatedAt: uploadNow}
}
