package identity

import (
	"context"

	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
)

type Profile[Scope any] interface {
	Load(
		context.Context, Scope,
		identitydomain.ProfileResource, identitydomain.IdentityID,
	) (identitydomain.ProfileDocument, error)
	PrepareNew(
		context.Context, Scope,
		identitydomain.ProfileResource, []byte,
	) (identitydomain.ProfileDocument, error)
	PreparePatch(
		context.Context, Scope,
		identitydomain.ProfileResource, identitydomain.IdentityID, []byte,
	) (identitydomain.ProfileDocument, error)
}

type ProfileAliases[Scope any] interface {
	Profile[Scope]
	FieldIDs(
		context.Context, Scope, identitydomain.ProfileResource, []string,
	) (map[string]identitydomain.IdentityID, error)
	FieldIDsForSchema(
		context.Context, Scope, identitydomain.ProfileResource,
		identitydomain.IdentityID, []string,
	) (map[string]identitydomain.IdentityID, error)
	PublicFieldNamesForSchema(
		context.Context, Scope, identitydomain.ProfileResource,
		identitydomain.IdentityID,
	) (map[identitydomain.IdentityID]string, error)
}
