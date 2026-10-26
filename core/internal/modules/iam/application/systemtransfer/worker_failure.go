package systemtransfer

import (
	"crypto/sha256"
	"errors"
	"fmt"

	tabulartransfer "github.com/Onurlulardan/harpy/core/internal/capabilities/tabulartransfer/api"
	"github.com/google/uuid"
)

type WorkerFailureClass uint8

const (
	WorkerFailurePermanent WorkerFailureClass = iota + 1
	WorkerFailureTransient
	WorkerFailureCancelled
)

type WorkerFailure struct {
	Class   WorkerFailureClass
	Code    string
	Digest  [sha256.Size]byte
	ActorID uuid.UUID
	cause   error
}

func (failure *WorkerFailure) Error() string {
	if failure == nil {
		return "IAM system import worker failed"
	}
	return fmt.Sprintf("IAM system import worker failed: %s", failure.Code)
}

func (failure *WorkerFailure) Unwrap() error {
	if failure == nil {
		return nil
	}
	return failure.cause
}

func newWorkerFailure(class WorkerFailureClass, code string, cause error) error {
	return &WorkerFailure{Class: class, Code: code,
		Digest: sha256.Sum256([]byte("iam-system-import-worker.v1\x00" + code)),
		cause:  cause}
}

func classifyWorkerError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ErrWorkerLeaseLost):
		return ErrWorkerLeaseLost
	case errors.Is(err, ErrAuthorizationCheckpoint):
		return newWorkerFailure(WorkerFailurePermanent, "IMPORT_AUTH_REVOKED", err)
	case errors.Is(err, ErrImportCatalogPinStale):
		return newWorkerFailure(WorkerFailurePermanent,
			"IMPORT_CATALOG_PIN_STALE", err)
	case errors.Is(err, tabulartransfer.ErrFormat),
		errors.Is(err, tabulartransfer.ErrRowLimit),
		errors.Is(err, tabulartransfer.ErrColumnLimit),
		errors.Is(err, tabulartransfer.ErrByteLimit),
		errors.Is(err, ErrInvalidMapping), errors.Is(err, ErrForbiddenTarget),
		errors.Is(err, ErrInvalidAttempt):
		return newWorkerFailure(WorkerFailurePermanent, "IMPORT_SOURCE_INVALID", err)
	case errors.Is(err, ErrReceiptConflict), errors.Is(err, ErrWorkerState):
		return newWorkerFailure(WorkerFailurePermanent, "IMPORT_STATE_INCONSISTENT", err)
	default:
		return newWorkerFailure(WorkerFailureTransient,
			"IMPORT_DEPENDENCY_UNAVAILABLE", err)
	}
}

func attachWorkerActor(err error, actorID uuid.UUID) error {
	var failure *WorkerFailure
	if errors.As(err, &failure) && failure.ActorID == uuid.Nil && actorID != uuid.Nil {
		copy := *failure
		copy.ActorID = actorID
		return &copy
	}
	return err
}

func ClassifyWorkerFailure(err error, actorID uuid.UUID) error {
	return attachWorkerActor(classifyWorkerError(err), actorID)
}
