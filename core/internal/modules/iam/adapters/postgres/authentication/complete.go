package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
)

func (store *store) CompleteReauthentication(
	ctx context.Context,
	scope authentication.CredentialScope,
	completion authentication.ReauthenticationCompletion,
) (authentication.ReauthenticationCompleted, error) {
	bound, err := store.bindReauthentication(ctx, scope)
	if err != nil {
		return authentication.ReauthenticationCompleted{}, err
	}
	row, err := bound.performReauthenticationCompletion(ctx, scope, completion)
	if err != nil {
		return authentication.ReauthenticationCompleted{}, err
	}
	defer row.clear()
	if err := bound.advanceLocator(ctx, scope, row.advance); err != nil {
		return authentication.ReauthenticationCompleted{}, err
	}
	if err := bound.createLocator(
		ctx, scope, completion.SuccessorID, completion.FamilyID, row.successor,
	); err != nil {
		return authentication.ReauthenticationCompleted{}, err
	}
	return row.completed, nil
}

type reauthenticationCompletionRow struct {
	completed authentication.ReauthenticationCompleted
	advance   locatorAdvance
	successor locatorCreate
}

func (row reauthenticationCompletionRow) valid() bool {
	return !row.completed.IssuedAt.IsZero() && row.advance.valid() &&
		row.successor.valid()
}

func (row *reauthenticationCompletionRow) clear() {
	if row == nil {
		return
	}
	clear(row.advance.sourceSHA256)
	clear(row.successor.sourceSHA256)
}
