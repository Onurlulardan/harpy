package policy

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
	"github.com/google/uuid"
)

type ActorVersions struct {
	actorID              uuid.UUID
	tokenVersion         int64
	authorizationVersion int64
	validated            bool
}

func NewActorVersions(
	actorID uuid.UUID,
	tokenVersion int64,
	authorizationVersion int64,
) (ActorVersions, error) {
	versions := ActorVersions{
		actorID: actorID, tokenVersion: tokenVersion,
		authorizationVersion: authorizationVersion, validated: true,
	}
	if versions.Validate() != nil {
		return ActorVersions{}, catalog.ErrContract
	}
	return versions, nil
}

func (versions ActorVersions) Validate() error {
	if !versions.validated || versions.actorID == uuid.Nil ||
		versions.tokenVersion < 1 || versions.authorizationVersion < 1 {
		return catalog.ErrContract
	}
	return nil
}

func (versions ActorVersions) ActorID() uuid.UUID  { return versions.actorID }
func (versions ActorVersions) TokenVersion() int64 { return versions.tokenVersion }
func (versions ActorVersions) AuthorizationVersion() int64 {
	return versions.authorizationVersion
}
