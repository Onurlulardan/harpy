package iam

import (
	"context"

	postrestorepg "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/postrestore"
	postrestoreapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/postrestore"
)

type RestoredIdentityCommand = postrestoreapp.ReconcileCommand
type RestoredIdentityResult = postrestoreapp.ReconcileResult

const RestoredIdentityStateReactivationRequired = postrestoreapp.StateReactivationRequired

type RestoredIdentityReconciler interface {
	Reconcile(context.Context, RestoredIdentityCommand) (RestoredIdentityResult, error)
}

func NewRestoredIdentityReconciler(
	pool postrestorepg.RuntimePool,
) (RestoredIdentityReconciler, error) {
	store, err := postrestorepg.NewStore(pool)
	if err != nil {
		return nil, err
	}
	return postrestoreapp.NewReconciler(store)
}
