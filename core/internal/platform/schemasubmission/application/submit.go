package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	submissiondomain "github.com/Onurlulardan/harpy/core/internal/platform/schemasubmission/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type Application struct{ binder StoreBinder }

func New(binder StoreBinder) (*Application, error) {
	if binder == nil {
		return nil, ErrInvalid
	}
	return &Application{binder: binder}, nil
}

func (application *Application) Submit(
	ctx context.Context,
	scope platformtx.TenantScope,
	command SubmitCommand,
) (submissiondomain.Receipt, error) {
	if application == nil || application.binder == nil || ctx == nil || ctx.Err() != nil ||
		platformtx.NetworkAllowed(ctx) || !platformtx.ContextOwnsScope(ctx, scope) ||
		platformtx.RequireActive(scope, platformtx.Command) != nil ||
		platformtx.RequireOwner(scope, "forms") != nil {
		return submissiondomain.Receipt{}, ErrUnavailable
	}
	prepared, err := prepare(command)
	if err != nil || prepared.TenantID() != scope.TenantID() {
		return submissiondomain.Receipt{}, ErrInvalid
	}
	store, err := application.binder.Bind(ctx, scope)
	if err != nil || store == nil {
		return submissiondomain.Receipt{}, ErrUnavailable
	}
	stored, err := store.Submit(ctx, prepared)
	if err != nil {
		return submissiondomain.Receipt{}, fmt.Errorf("submit schema plan: %w", err)
	}
	return receiptFromStored(stored, prepared.PlanSHA256())
}

func receiptFromStored(
	stored StoredReceipt,
	expectedDigest string,
) (submissiondomain.Receipt, error) {
	jobUUID, err := uuid.Parse(stored.JobID)
	if err != nil || jobUUID.String() != stored.JobID {
		return submissiondomain.Receipt{}, ErrUnavailable
	}
	jobID, err := submissiondomain.NewJobID(jobUUID)
	if err != nil || stored.PlanSHA256 != expectedDigest {
		return submissiondomain.Receipt{}, ErrUnavailable
	}
	digestBytes, err := hex.DecodeString(stored.PlanSHA256)
	if err != nil || len(digestBytes) != sha256.Size {
		return submissiondomain.Receipt{}, ErrUnavailable
	}
	var digest [sha256.Size]byte
	copy(digest[:], digestBytes)
	receipt, ok := submissiondomain.NewReceipt(
		jobID, stored.Generation, digest, stored.Replay,
	)
	if !ok {
		return submissiondomain.Receipt{}, ErrUnavailable
	}
	return receipt, nil
}
