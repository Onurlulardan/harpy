package policypostgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
	policyapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/policy"
	"github.com/google/uuid"
)

//go:embed sql/tenant/read_actor_versions.sql
var readActorVersionsSQL string

func (ports *ports) Versions(
	ctx context.Context,
	scope tenantScope,
	actorID uuid.UUID,
) (policyapp.ActorVersions, error) {
	if ports == nil || !scopeUsable(ctx, scope) || actorID == uuid.Nil ||
		actorID != scope.ActorID() {
		return policyapp.ActorVersions{}, catalog.ErrContract
	}
	bound, err := ports.bindVersions(scope)
	if err != nil {
		return policyapp.ActorVersions{}, catalog.ErrContract
	}
	var tokenVersion, authorizationVersion int64
	err = bound.database.QueryRow(ctx, readActorVersionsSQL, actorID).Scan(
		&tokenVersion, &authorizationVersion,
	)
	if err != nil {
		return policyapp.ActorVersions{}, catalog.ErrContract
	}
	return policyapp.NewActorVersions(actorID, tokenVersion, authorizationVersion)
}
