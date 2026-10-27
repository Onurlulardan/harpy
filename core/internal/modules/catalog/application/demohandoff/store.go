package demohandoff

import (
	"context"
	"errors"

	api "github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/demohandoff"
	"github.com/google/uuid"
)

type Row interface {
	Scan(...any) error
}

type QueryRow func(context.Context, string, ...any) Row
type ScopeGuard func(context.Context) bool

type BoundStore struct {
	query    string
	queryRow QueryRow
	allowed  ScopeGuard
}

func NewBoundStore(
	query string,
	queryRow QueryRow,
	allowed ScopeGuard,
) (*BoundStore, error) {
	if query == "" || queryRow == nil || allowed == nil {
		return nil, api.ErrInvalid
	}
	return &BoundStore{query: query, queryRow: queryRow, allowed: allowed}, nil
}

func (store *BoundStore) Persist(
	ctx context.Context,
	projection Projection,
) (api.Receipt, error) {
	if store == nil || store.query == "" || store.queryRow == nil ||
		store.allowed == nil || !store.allowed(ctx) ||
		projection.ProposedHandoffID.Version() != 7 ||
		projection.OwnerGeneration < 1 || !requestValid(projection.Request) {
		return api.Receipt{}, api.ErrInvalid
	}
	request := projection.Request
	var receipt api.Receipt
	var handoffID, provisionID, componentID, packageID, releaseID, runID uuid.UUID
	var requestSHA, readbackSHA []byte
	err := store.queryRow(ctx, store.query,
		projection.ProposedHandoffID, projection.OwnerGeneration,
		request.ProvisionReceiptID, request.ComponentID,
		request.PackageID, request.ReleaseID, request.FormsRunID,
		request.RequestSHA256[:],
		request.FormCount, request.RowCount, request.ReadbackSHA256[:],
		request.CompletedAt,
	).Scan(&handoffID, &provisionID, &componentID, &packageID, &releaseID, &runID,
		&requestSHA, &receipt.FormCount, &receipt.RowCount, &readbackSHA,
		&receipt.CompletedAt)
	if err != nil {
		return api.Receipt{}, mapStoreError(err)
	}
	if len(requestSHA) != 32 || len(readbackSHA) != 32 {
		return api.Receipt{}, api.ErrUnavailable
	}
	copy(receipt.RequestSHA256[:], requestSHA)
	copy(receipt.ReadbackSHA256[:], readbackSHA)
	receipt.HandoffID = [16]byte(handoffID)
	receipt.ProvisionReceiptID = [16]byte(provisionID)
	receipt.ComponentID = [16]byte(componentID)
	receipt.PackageID = [16]byte(packageID)
	receipt.ReleaseID = [16]byte(releaseID)
	receipt.FormsRunID = [16]byte(runID)
	receipt.CompletedAt = receipt.CompletedAt.UTC()
	return receipt, nil
}

func mapStoreError(err error) error {
	var state interface{ SQLState() string }
	if errors.As(err, &state) {
		switch state.SQLState() {
		case "22023":
			return api.ErrInvalid
		case "23505", "40001":
			return api.ErrConflict
		}
	}
	return api.ErrUnavailable
}

var _ Store = (*BoundStore)(nil)
