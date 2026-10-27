package iam

import (
	"context"

	lifecycledeliveryv1 "github.com/Onurlulardan/harpy/contracts/lifecycledelivery/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
	"github.com/google/uuid"
)

func (application *lifecycleHTTPApplication) deliveryArtifact(
	ctx context.Context,
	tenantID, artifactID, tokenID uuid.UUID,
	purpose lifecycle.CredentialPurpose,
	target lifecycle.DeliveryTarget,
	issued *lifecycle.IssuedCredential,
) (lifecycle.DeliveryArtifact, error) {
	contractPurpose := lifecycledeliveryv1.Purpose(purpose)
	var envelope lifecycledeliveryv1.Envelope
	err := issued.RevealOnce(func(plaintext []byte) error {
		var sealErr error
		envelope, sealErr = application.delivery.SealLifecycleCredential(
			ctx, tenantID, artifactID, target.SubjectUserID, target.ResourceID,
			tokenID, contractPurpose, plaintext,
		)
		return sealErr
	})
	if err != nil || envelope.Validate(tenantID) != nil {
		return lifecycle.DeliveryArtifact{}, ErrLifecycleHTTPUnavailable
	}
	reference, err := lifecycledeliveryv1.NewReference(tenantID, artifactID)
	if err != nil {
		return lifecycle.DeliveryArtifact{}, ErrLifecycleHTTPUnavailable
	}
	return lifecycle.DeliveryArtifact{
		Reference:             reference.String(),
		RecipientIdentifierID: target.RecipientIdentifierID, Envelope: envelope,
	}, nil
}
