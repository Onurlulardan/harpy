package invitationaudit

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
)

//go:embed sql/tenant/persist_audit.sql
var persistAuditSQL string

func (store *store) PersistTenant(
	ctx context.Context,
	scope tenantScope,
	change invitation.AuditChange,
) error {
	if change.Validate() != nil || !tenantUsable(ctx, scope) ||
		scope.ActorID() != change.ActorID {
		return errUnavailable
	}
	bound, err := store.bindTenant(scope)
	if err != nil {
		return err
	}
	return persist(ctx, bound, change)
}

func (store *store) PersistCredential(
	ctx context.Context,
	scope credentialScope,
	change invitation.AuditChange,
) error {
	if change.Validate() != nil || !credentialUsable(ctx, scope) ||
		change.CommandKey != "iam.invitation.accept" ||
		scope.SourceUserID() != change.ActorID {
		return errUnavailable
	}
	bound, err := store.bindCredential(scope)
	if err != nil {
		return err
	}
	return persist(ctx, bound, change)
}

func persist(
	ctx context.Context,
	bound *boundStore,
	change invitation.AuditChange,
) error {
	var artifactRef any
	if change.DeliveryArtifactRef != "" {
		artifactRef = change.DeliveryArtifactRef
	}
	var authorizationClock any
	if change.AuthorizationClock > 0 {
		authorizationClock = change.AuthorizationClock
	}
	var written bool
	err := bound.database.QueryRow(
		ctx, persistAuditSQL, change.OperationID, change.ActorID,
		change.InvitationID, change.RequestSHA256[:], change.CommandKey,
		string(change.ResultState), change.ResultRevision,
		artifactRef, authorizationClock,
	).Scan(&written)
	if err != nil || !written {
		return errUnavailable
	}
	return nil
}
