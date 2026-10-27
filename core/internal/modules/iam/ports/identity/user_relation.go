package identity

import (
	"context"

	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
)

type UserRelationStore[Scope any] interface {
	ReplaceUserRelations(
		context.Context,
		Scope,
		identitydomain.ReplaceUserRelationsCommand,
	) (identitydomain.MutationResult, error)
}
