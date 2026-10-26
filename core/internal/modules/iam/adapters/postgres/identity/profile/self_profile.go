package profilepostgres

import (
	"context"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
	"github.com/google/uuid"
)

type SelfWriter interface {
	PrepareSelfPatch(
		context.Context, tenantScope, uuid.UUID, []byte,
	) (SelfPreparedWrite, error)
	SourceForSelf(
		context.Context, tenantScope, uuid.UUID,
	) (profileapp.Schema, profileapp.Document, error)
}

func NewSelfWriter() (SelfWriter, error) {
	return newStore()
}

func (storage *store) PrepareSelfPatch(
	ctx context.Context,
	scope tenantScope,
	actorID uuid.UUID,
	patch []byte,
) (SelfPreparedWrite, error) {
	if scope.ActorID() != actorID {
		return SelfPreparedWrite{}, identitydomain.ErrInvalidCommand
	}
	prepared, err := storage.preparePatchWrite(
		ctx, scope, identitydomain.ProfileUser,
		identitydomain.IdentityID(actorID), patch,
	)
	if err != nil {
		return SelfPreparedWrite{}, err
	}
	return SelfPreparedWrite{
		ExpectedSourceSchemaID: prepared.expectedSourceSchemaID,
		SourceSchema:           prepared.sourceSchema, Target: prepared.target,
	}, nil
}

func (storage *store) SourceForSelf(
	ctx context.Context,
	scope tenantScope,
	actorID uuid.UUID,
) (profileapp.Schema, profileapp.Document, error) {
	if !profileLookupScopeUsable(ctx, scope) || actorID == uuid.Nil ||
		scope.ActorID() != actorID {
		return profileapp.Schema{}, profileapp.Document{}, identitydomain.ErrInvalidQuery
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return profileapp.Schema{}, profileapp.Document{}, identitydomain.ErrStorage
	}
	contract, err := storage.loadSelfContract(ctx, scope, bound)
	if err != nil {
		return profileapp.Schema{}, profileapp.Document{}, identitydomain.ErrStorage
	}
	current, err := storage.loadSelfResource(ctx, scope, bound, contract, actorID)
	if err != nil {
		return profileapp.Schema{}, profileapp.Document{}, err
	}
	source, _, err := profileSourceRead(contract, current)
	if err != nil {
		return profileapp.Schema{}, profileapp.Document{}, identitydomain.ErrStorage
	}
	return contract.source, source, nil
}

func (storage *store) loadSelfContract(
	ctx context.Context, scope tenantScope, bound *boundStore,
) (profileWriteContract, error) {
	if scopeUsable(ctx, scope, commandMode) {
		return storage.loadWriteContract(ctx, bound, identitydomain.ProfileUser)
	}
	return storage.loadReadContract(ctx, bound, identitydomain.ProfileUser)
}

func (storage *store) loadSelfResource(
	ctx context.Context,
	scope tenantScope,
	bound *boundStore,
	contract profileWriteContract,
	actorID uuid.UUID,
) (profileWriteResource, error) {
	if scopeUsable(ctx, scope, commandMode) {
		return storage.loadWriteResource(
			ctx, bound, contract, identitydomain.ProfileUser, actorID,
		)
	}
	return storage.loadReadResource(
		ctx, bound, contract, identitydomain.ProfileUser, actorID,
	)
}
