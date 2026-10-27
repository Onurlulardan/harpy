package iam

import (
	"context"

	deliveryartifactv1 "github.com/Onurlulardan/harpy/contracts/deliveryartifact/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/gcpkms"
	"github.com/google/uuid"
)

type gcpInvitationCredentialOpener struct{ opener *gcpkms.DeliveryOpener }

func (opener gcpInvitationCredentialOpener) Open(
	ctx context.Context,
	tenantID uuid.UUID,
	envelope deliveryartifactv1.Envelope,
) (invitationCredential, error) {
	return opener.opener.OpenInvitationCredential(ctx, tenantID, envelope)
}
