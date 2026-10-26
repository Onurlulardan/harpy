package iam

import (
	"time"

	invitationaudit "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/invitationaudit"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

const invitationPayloadSchemaSHA256 = "3dbc2634af290b10dc25fad1db39d01f51d5acd5c07c920288be004c9d0a531b"

type InvitationClock func() time.Time

type InvitationChangeSink struct {
	audit    invitation.AuditStore
	catalogs platformasync.WriteCatalogs
	now      InvitationClock
}

func NewInvitationChangeSink(
	catalogs platformasync.WriteCatalogs,
	now InvitationClock,
) (*InvitationChangeSink, error) {
	if now == nil {
		return nil, invitation.ErrEffects
	}
	if _, _, err := catalogs.Snapshot(); err != nil {
		return nil, invitation.ErrEffects
	}
	audit, err := invitationaudit.NewStore()
	if err != nil {
		return nil, err
	}
	return &InvitationChangeSink{audit: audit, catalogs: catalogs, now: now}, nil
}

var _ invitation.Effects = (*InvitationChangeSink)(nil)
