package recoveryanchor

import (
	"context"
	"fmt"

	anchorports "github.com/Onurlulardan/harpy/core/internal/modules/iam/ports/recoveryanchor"
)

type AuthorityFinalizer struct {
	store anchorports.AuthorityFinalizerStore
}

func NewAuthorityFinalizer(
	store anchorports.AuthorityFinalizerStore,
) (*AuthorityFinalizer, error) {
	if store == nil {
		return nil, fmt.Errorf("recovery anchor authority finalizer is unavailable")
	}
	return &AuthorityFinalizer{store: store}, nil
}

func (finalizer *AuthorityFinalizer) Finalize(
	ctx context.Context,
	operationID string,
) error {
	if finalizer == nil || finalizer.store == nil ||
		ctx == nil || operationID == "" {
		return fmt.Errorf("recovery anchor authority finalizer is unavailable")
	}
	return finalizer.store.FinalizeAuthority(ctx, operationID)
}
