package aiplatformowner

import (
	"context"

	"github.com/google/uuid"
)

func (reader *boundReader) readSession(
	ctx context.Context, input Input,
) (Evidence, error) {
	if reader == nil || ctx == nil || ctx.Err() != nil {
		return Evidence{}, ErrUnavailable
	}
	var evidence Evidence
	var portal, subject uuid.NullUUID
	var epoch []byte
	err := reader.database.QueryRow(
		ctx, currentSessionSQL, input.ActorID, input.SessionID,
		input.ExpectedTokenVersion, input.ExpectedAuthorizationVersion,
	).Scan(
		&evidence.UserID, &evidence.SessionID, &evidence.ClientID,
		&evidence.AccessMode, &portal, &evidence.TokenVersion,
		&evidence.AuthorizationVersion, &evidence.DeploymentMode,
		&evidence.TenantPurpose, &subject, &evidence.PlatformClock, &epoch,
	)
	if err != nil {
		if rowNotFound(err) {
			return Evidence{}, ErrRejected
		}
		return Evidence{}, ErrUnavailable
	}
	if len(epoch) != len(evidence.SecurityEpoch) {
		clear(epoch)
		return Evidence{}, ErrRejected
	}
	copy(evidence.SecurityEpoch[:], epoch)
	clear(epoch)
	if portal.Valid {
		evidence.PortalBindingID = portal.UUID
	}
	if subject.Valid {
		evidence.PlatformSubjectID = subject.UUID
	}
	return evidence, nil
}

func (reader *boundReader) readOwner(
	ctx context.Context, input Input, evidence *Evidence,
) error {
	if reader == nil || evidence == nil || ctx == nil || ctx.Err() != nil {
		return ErrUnavailable
	}
	err := reader.database.QueryRow(
		ctx, currentOwnerSQL,
		input.TenantID, input.ActorID, input.ExpectedPlatformClock,
	).Scan(
		&evidence.EnvironmentID, &evidence.OwnerDeploymentMode,
		&evidence.OwnerSubjectID, &evidence.OwnerPlatformClock,
	)
	if err != nil {
		if rowNotFound(err) {
			return ErrRejected
		}
		return ErrUnavailable
	}
	return nil
}
