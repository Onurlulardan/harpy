package sessionview

import (
	"context"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/sessionview"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/sessioncapability"
	"github.com/google/uuid"
)

func (store *store) Read(
	ctx context.Context,
	scope tenantScope,
	userID uuid.UUID,
	sessionID uuid.UUID,
) (app.View, error) {
	bound, err := store.bind(ctx, scope)
	if err != nil || scope.ActorID() != userID {
		return app.View{}, errRead
	}
	var view app.View
	var client, mode string
	var portalBindingID *uuid.UUID
	var platformSubjectID *uuid.UUID
	var deploymentMode, tenantPurpose string
	var securityEpoch []byte
	err = bound.database.QueryRow(
		ctx, readSQL, userID, sessionID,
	).Scan(
		&view.UserID, &view.SessionID, &client, &mode,
		&portalBindingID, &view.TokenVersion,
		&view.AuthorizationVersion, &deploymentMode, &tenantPurpose,
		&platformSubjectID, &view.PlatformAuthorizationClock,
		&securityEpoch, &view.TenantCapabilities,
		&view.PasswordCredentialRevision,
		&view.Email, &view.DisplayName,
	)
	if err != nil || len(securityEpoch) != len(view.SecurityEpoch) {
		clear(securityEpoch)
		if rejectedRead(err) {
			return app.View{}, app.ErrRejected
		}
		return app.View{}, errRead
	}
	copy(view.SecurityEpoch[:], securityEpoch)
	clear(securityEpoch)
	view.ClientID, view.AccessMode = app.ClientID(client), app.AccessMode(mode)
	view.DeploymentMode = sessioncapability.DeploymentMode(deploymentMode)
	view.TenantPurpose = sessioncapability.TenantPurpose(tenantPurpose)
	if platformSubjectID != nil {
		view.PlatformSubjectID = *platformSubjectID
	}
	if portalBindingID != nil {
		view.PortalBindingID = *portalBindingID
	}
	return view, nil
}
