package iam

import (
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	"github.com/google/uuid"
)

func systemTransferResource(value SystemTransferResource) (publicresource.Resource, error) {
	resource, ok := publicresource.ParseResource(string(value))
	if !ok {
		return "", ErrSystemTransferInvalid
	}
	return resource, nil
}

func systemTransferEvidence(
	access SystemTransferAccess,
	resource publicresource.Resource,
) systemtransfer.ControlEvidence {
	return systemtransfer.ControlEvidence{
		TenantID: access.TenantID(), ActorID: access.PrincipalID(),
		TokenVersion:         access.TokenVersion(),
		AuthorizationVersion: access.AuthorizationVersion(),
		AuthorizationClock:   access.AuthorizationClock(), Resource: resource,
	}
}

func systemTransferActor(
	access SystemTransferAccess,
) (publicresource.ActorEvidence, error) {
	tenantID, err := publicresource.ParseID(access.TenantID().String())
	if err != nil {
		return publicresource.ActorEvidence{}, ErrSystemTransferInvalid
	}
	actorID, err := publicresource.ParseID(access.PrincipalID().String())
	if err != nil {
		return publicresource.ActorEvidence{}, ErrSystemTransferInvalid
	}
	return publicresource.NewActorEvidence(tenantID, actorID,
		access.TokenVersion(), access.AuthorizationVersion())
}

func systemTransferRead(access SystemTransferAccess,
	resource publicresource.Resource, transferID uuid.UUID) systemtransfer.ControlRead {
	return systemtransfer.ControlRead{
		Evidence: systemTransferEvidence(access, resource), TransferID: transferID,
	}
}
