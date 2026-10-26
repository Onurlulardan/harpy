package iam

import (
	invitationpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/invitation"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identifier"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

// InvitationRuntime is IAM's transaction-bound invitation aggregate.
// Application services remain private so composition cannot bypass IAM effects.
type InvitationRuntime struct {
	service *invitation.Service
}

func NewInvitationRuntime(
	catalogs platformasync.WriteCatalogs,
	now InvitationClock,
) (*InvitationRuntime, error) {
	store, err := invitationpostgres.NewInvitationStore()
	if err != nil {
		return nil, err
	}
	sink, err := NewInvitationChangeSink(catalogs, now)
	if err != nil {
		return nil, err
	}
	service, err := invitation.NewService(store, invitationNormalizer{}, sink)
	if err != nil {
		return nil, err
	}
	return &InvitationRuntime{service: service}, nil
}

func (runtime *InvitationRuntime) Valid() bool {
	return runtime != nil && runtime.service != nil
}

type invitationNormalizer struct{}

func (invitationNormalizer) NormalizeEmail(
	raw string,
) (invitation.EmailEvidence, error) {
	value, err := identifier.NormalizeEmailV1(raw)
	if err != nil {
		return invitation.EmailEvidence{}, err
	}
	return invitation.EmailEvidence{
		NormalizedValue: value.Normalized(), SHA256: value.NormalizedSHA256(),
	}, nil
}
