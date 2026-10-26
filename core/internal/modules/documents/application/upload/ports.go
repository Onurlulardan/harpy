package upload

import (
	"context"
	"io"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Access uint8

const (
	Command Access = iota + 1
	SnapshotAccess
)

type Store interface {
	FindCreate(context.Context, string, [32]byte) (Snapshot, bool, error)
	Create(context.Context, CreateMutation) (Snapshot, error)
	Get(context.Context, domain.ID) (Snapshot, error)
	Begin(context.Context, domain.ID, domain.ID, domain.ID) (AttemptPlan, error)
	PersistEffect(context.Context, EffectReceipt) error
	PersistFailure(context.Context, EffectFailure, domain.ID) error
	PrepareFinalize(context.Context, EffectReceipt) (FinalizeDecision, error)
	PersistFinalization(context.Context, EffectReceipt, domain.ID, domain.ID) (ReadyResult, error)
	Cancel(context.Context, domain.ID, domain.ID, time.Time) (CancelResult, error)
}

type StoreBinder interface {
	Bind(context.Context, platformtx.TenantScope, Access) (Store, foundation.Identity, error)
}

type Billing interface {
	Reserve(context.Context, platformtx.TenantScope, ReservationRequest) (Reservation, error)
	StartEffect(context.Context, platformtx.TenantScope, domain.ID, domain.ID) error
	Commit(context.Context, platformtx.TenantScope, domain.ID, domain.ID, int64) error
	Release(context.Context, platformtx.TenantScope, domain.ID, domain.ID) error
}

type Jobs interface {
	EnqueueDeletion(context.Context, platformtx.TenantScope,
		domain.ID, domain.ID, []byte) error
	EnqueueReconciliation(context.Context, platformtx.TenantScope,
		domain.ID, domain.ID, uint64, time.Time, []byte) error
}

type ReservationRequest struct {
	SourceOperationID domain.ID
	Amount            int64
	ExpiresAt         time.Time
}

type Reservation struct {
	ID      domain.ID
	Allowed bool
	Replay  bool
}

type EffectRunner interface {
	Execute(context.Context, AttemptPlan, io.Reader) (EffectReceipt, *EffectFailure)
}

type DataKey struct {
	Plaintext [32]byte
	Wrapped   []byte
	KeyID     string
}

type Keyring interface {
	CreateDataKey(context.Context, domain.ID, domain.ID, domain.ID) (DataKey, error)
}

type CipherMetadata struct {
	TenantID, DocumentID, VersionID domain.ID
	Generation                      uint64
	PlaintextBytes                  int64
}

type CipherReceipt struct {
	PlaintextSHA256, CiphertextSHA256 [32]byte
	CiphertextBytes                   int64
}

type Cipher interface {
	Encrypt(context.Context, io.Reader, io.Writer, [32]byte, CipherMetadata) (CipherReceipt, error)
}

type ScanSession interface {
	io.Writer
	Finish(context.Context) (ScanResult, error)
}

type ScanResult struct {
	Clean                           bool
	DetectedKind                    domain.FileKind
	DetectedMIME, SignatureRevision string
	ScannedAt                       time.Time
}

type Scanner interface {
	Open(context.Context, string, string, int64) (ScanSession, error)
}

type ObjectPutRequest struct {
	Key             string
	Generation      uint64
	CiphertextBytes int64
	Body            io.Reader
}

type ObjectPutResult struct {
	Version    string
	Generation uint64
}

type ObjectExpected struct {
	Key, Version     string
	Generation       uint64
	CiphertextBytes  int64
	CiphertextSHA256 [32]byte
}

type ObjectStore interface {
	Put(context.Context, ObjectPutRequest) (ObjectPutResult, error)
	Reconcile(context.Context, ObjectExpected) (ObjectPutResult, error)
	Verify(context.Context, ObjectExpected) error
}

type ScopeRunner interface {
	InCommand(context.Context, string,
		func(context.Context, platformtx.TenantScope) error) error
}
