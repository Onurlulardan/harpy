package dashboardexport

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
)

type workerStoreFixture struct {
	claim         RenderClaim
	deleteClaim   DeleteClaim
	expiryControl ExpiryControl
	failedCode    string
	unknown       bool
	effectStarted bool
	deleted       bool
	effectWrite   EffectStart
	readyWrite    ReadyWrite
	failureWrite  RenderFailure
	unknownWrite  RenderUnknown
	markEffectErr error
	expiryErr     error
}

func (store *workerStoreFixture) MarkEffectStarted(_ context.Context, write EffectStart) error {
	store.effectStarted = true
	store.effectWrite = write
	return store.markEffectErr
}
func (store *workerStoreFixture) CompleteReady(_ context.Context, write ReadyWrite) (State, error) {
	store.readyWrite = write
	return StateReady, nil
}
func (store *workerStoreFixture) MarkOutcomeUnknown(
	_ context.Context, write RenderUnknown,
) error {
	store.unknown = true
	store.unknownWrite = write
	return nil
}
func (store *workerStoreFixture) FinishRender(
	_ context.Context, write RenderFailure,
) (State, error) {
	store.failedCode = write.Code
	store.failureWrite = write
	if write.FailureClass == "cancelled" {
		return StateCancelled, nil
	}
	return StateFailed, nil
}
func (store *workerStoreFixture) ControlExpiry(
	context.Context, uuid.UUID, uuid.UUID, uuid.UUID, time.Time,
) (ExpiryControl, error) {
	return store.expiryControl, store.expiryErr
}
func (store *workerStoreFixture) ClaimDelete(
	context.Context, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, time.Time,
) (DeleteClaim, error) {
	return store.deleteClaim, nil
}
func (store *workerStoreFixture) CompleteDeleted(
	context.Context, DeleteClaim, DeleteReceipt, time.Time,
) error {
	store.deleted = true
	return nil
}

type workerAuthorizerFixture struct{ calls, failAt int }

func (authorizer *workerAuthorizerFixture) Authorize(
	context.Context, AuthorizationRequest,
) error {
	authorizer.calls++
	if authorizer.calls == authorizer.failAt {
		return ErrForbidden
	}
	return nil
}

type workerRendererFixture struct{}

func (workerRendererFixture) Render(context.Context, Pin, io.Writer) (RenderStats, error) {
	return RenderStats{Rows: 1}, nil
}

type pendingFixture struct {
	uploaded bool
	err      error
	receipt  UploadReceipt
	upload   func(context.Context) (UploadReceipt, error)
}

func (pending *pendingFixture) Draft() ArtifactDraft { return workerDraft() }
func (pending *pendingFixture) Upload(ctx context.Context) (UploadReceipt, error) {
	pending.uploaded = true
	if pending.upload != nil {
		return pending.upload(ctx)
	}
	return pending.receipt, pending.err
}
func (*pendingFixture) Discard() error { return nil }

type artifactRuntimeFixture struct {
	pending         *pendingFixture
	reconciliation  Reconciliation
	reconcileErr    error
	deleteCalled    bool
	deleteVersion   string
	deleteReceipt   DeleteReceipt
	deleteErr       error
	prepareReserved bool
}

func (runtime *artifactRuntimeFixture) PrepareReserved(
	context.Context, Pin, int64, uuid.UUID, EncryptedSpool, Renderer,
) (PendingArtifact, error) {
	runtime.prepareReserved = true
	return runtime.pending, nil
}
func (runtime *artifactRuntimeFixture) Reconcile(
	context.Context, ArtifactDraft,
) (Reconciliation, error) {
	return runtime.reconciliation, runtime.reconcileErr
}
func (runtime *artifactRuntimeFixture) Delete(
	_ context.Context, _ ArtifactDraft, version string,
) (DeleteReceipt, error) {
	runtime.deleteCalled = true
	runtime.deleteVersion = version
	return runtime.deleteReceipt, runtime.deleteErr
}
func (*artifactRuntimeFixture) Open(context.Context, StoredContent) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("artifact")), nil
}

type workerClockFixture struct{ now time.Time }

func (clock workerClockFixture) Now() time.Time { return clock.now }

type workerIDFixture struct{ next []uuid.UUID }

func (source *workerIDFixture) NewUUIDv7() (uuid.UUID, error) {
	if len(source.next) == 0 {
		return uuid.Nil, ErrUnavailable
	}
	value := source.next[0]
	source.next = source.next[1:]
	return value, nil
}

func workerIDs() *workerIDFixture {
	return &workerIDFixture{next: []uuid.UUID{
		uuid.MustParse("019b1000-0000-7000-8000-00000000000d"),
		uuid.MustParse("019b1000-0000-7000-8000-00000000000e"),
	}}
}

func workerDraft() ArtifactDraft {
	return ArtifactDraft{AttemptGeneration: 1, ObjectKey: "analytics/export/object",
		CiphertextBytes: 64, PlaintextBytes: 16, Rows: 1,
		ContentType: FormatCSV.ContentType(), Filename: "dashboard.csv"}
}
