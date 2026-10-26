package iam

import (
	"context"
	"fmt"

	deliveryartifactv1 "github.com/Onurlulardan/harpy/contracts/deliveryartifact/v1"
	lifecycledeliveryv1 "github.com/Onurlulardan/harpy/contracts/lifecycledelivery/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/gcpkms"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/google/uuid"
)

type DeliverySealer struct{ inner *gcpkms.DeliverySealer }

func OpenDeliverySealer(
	ctx context.Context, settings config.IAMAPI,
) (*DeliverySealer, error) {
	key, err := settings.DeliveryKMSKeyID()
	if err != nil {
		return nil, fmt.Errorf("load IAM delivery KMS key: %w", err)
	}
	return openDeliverySealer(ctx, key)
}

func OpenWorkerDeliverySealer(
	ctx context.Context, settings config.IAMWorker,
) (*DeliverySealer, error) {
	key, err := settings.DeliveryKMSKeyID()
	if err != nil {
		return nil, fmt.Errorf("load IAM worker delivery KMS key: %w", err)
	}
	return openDeliverySealer(ctx, key)
}

func openDeliverySealer(
	ctx context.Context, key config.IAMKMSKeyID,
) (*DeliverySealer, error) {
	keyID, err := key.Value()
	if err != nil {
		return nil, fmt.Errorf("load IAM delivery KMS key value: %w", err)
	}
	inner, err := gcpkms.OpenDeliverySealer(ctx, keyID)
	if err != nil {
		return nil, fmt.Errorf("open IAM delivery sealer: %w", err)
	}
	return &DeliverySealer{inner: inner}, nil
}

func (sealer *DeliverySealer) SealLifecycleCredential(
	ctx context.Context,
	tenantID, artifactID, subjectUserID, resourceID, tokenID uuid.UUID,
	purpose lifecycledeliveryv1.Purpose,
	plaintext []byte,
) (lifecycledeliveryv1.Envelope, error) {
	if sealer == nil || sealer.inner == nil {
		return lifecycledeliveryv1.Envelope{}, ErrInvitationHTTPUnavailable
	}
	return sealer.inner.SealLifecycleCredential(ctx, tenantID, artifactID,
		subjectUserID, resourceID, tokenID, purpose, plaintext)
}

func (sealer *DeliverySealer) SealInvitationCredential(
	ctx context.Context,
	tenantID, artifactID, invitationID, tokenID uuid.UUID,
	plaintext []byte,
) (deliveryartifactv1.Envelope, error) {
	if sealer == nil || sealer.inner == nil {
		return deliveryartifactv1.Envelope{}, ErrInvitationHTTPUnavailable
	}
	return sealer.inner.SealInvitationCredential(
		ctx, tenantID, artifactID, invitationID, tokenID, plaintext,
	)
}

func (sealer *DeliverySealer) Close() error {
	if sealer == nil || sealer.inner == nil {
		return nil
	}
	return sealer.inner.Close()
}

var _ InvitationDeliverySealer = (*DeliverySealer)(nil)
