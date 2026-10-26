package iam

import (
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/google/uuid"
)

type identityPublicEvidence struct {
	tenantID             uuid.UUID
	actorID              uuid.UUID
	tokenVersion         int64
	authorizationVersion int64
}

func newIdentityPublicEvidence(
	evidence publicresource.ActorEvidence,
) (identityPublicEvidence, error) {
	tenantID, tenantErr := evidence.TenantID()
	actorID, actorErr := evidence.ActorID()
	tokenVersion, authorizationVersion, versionsErr := evidence.Versions()
	if tenantErr != nil || actorErr != nil || versionsErr != nil {
		return identityPublicEvidence{}, publicresource.ErrContract
	}
	return identityPublicEvidence{
		tenantID: uuid.UUID(tenantID), actorID: uuid.UUID(actorID), tokenVersion: tokenVersion,
		authorizationVersion: authorizationVersion,
	}, nil
}

func (evidence identityPublicEvidence) Subject() uuid.UUID { return evidence.actorID }
func (evidence identityPublicEvidence) Tenant() uuid.UUID  { return evidence.tenantID }
func (evidence identityPublicEvidence) TokenVersion() int64 {
	return evidence.tokenVersion
}
func (evidence identityPublicEvidence) AuthorizationVersion() int64 {
	return evidence.authorizationVersion
}
