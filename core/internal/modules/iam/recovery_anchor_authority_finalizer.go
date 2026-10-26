package iam

import (
	recoveryanchorpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/recoveryanchor"
	anchorapplication "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/recoveryanchor"
)

type RecoveryAnchorAuthorityFinalizer = anchorapplication.AuthorityFinalizer

func NewRecoveryAnchorAuthorityFinalizer(
	pool *recoveryanchorpostgres.AuthorityFinalizerPool,
) (*RecoveryAnchorAuthorityFinalizer, error) {
	store, err := recoveryanchorpostgres.NewAuthorityFinalizerStore(pool)
	if err != nil {
		return nil, err
	}
	return anchorapplication.NewAuthorityFinalizer(store)
}
