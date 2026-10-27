package iam

import (
	"context"

	lifecycledeliveryv1 "github.com/Onurlulardan/harpy/contracts/lifecycledelivery/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/gcpkms"
	"github.com/google/uuid"
)

type gcpLifecycleCredentialOpener struct{ opener *gcpkms.DeliveryOpener }

func (opener gcpLifecycleCredentialOpener) Open(ctx context.Context,
	tenantID uuid.UUID, envelope lifecycledeliveryv1.Envelope,
) (lifecycleCredential, error) {
	return opener.opener.OpenLifecycleCredential(ctx, tenantID, envelope)
}
