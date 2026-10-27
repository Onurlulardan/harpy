package iam

import (
	"context"
	"time"

	deliveryartifactv1 "github.com/Onurlulardan/harpy/contracts/deliveryartifact/v1"
	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/smtp"
	invitationdeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitationdelivery"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
	"github.com/google/uuid"
)

type invitationCredential interface {
	RevealOnce(func([]byte) error) error
	Destroy()
}

type invitationCredentialOpener interface {
	Open(context.Context, uuid.UUID, deliveryartifactv1.Envelope) (invitationCredential, error)
}

type invitationSMTPSender interface {
	StableMessageID(string) (string, error)
	Send(context.Context, smtp.Message) smtp.Result
}

type invitationDeliveryState interface {
	Claim(context.Context, invitationdeliveryapp.ClaimCommand) (
		invitationdeliveryapp.ClaimDecision, string, error)
	Release(context.Context, invitationdeliveryapp.EffectReference) error
	MarkUnknown(context.Context, invitationdeliveryapp.EffectReference) error
	Finalize(context.Context, invitationdeliveryapp.FinalizeCommand) (invitationdeliveryapp.StoredResult, error)
}

type invitationDeliveryProvider struct {
	identity effectexecution.Identity
	state    invitationDeliveryState
	opener   invitationCredentialOpener
	smtp     invitationSMTPSender
	now      func() time.Time
}

func (provider *invitationDeliveryProvider) Identity() effectexecution.Identity {
	if provider == nil {
		return effectexecution.Identity{}
	}
	return provider.identity
}

var _ effectexecution.Provider = (*invitationDeliveryProvider)(nil)

type invitationRequest struct {
	request     effectsv1.Request
	tenantID    uuid.UUID
	requestedAt time.Time
	messageID   string
	reference   invitationdeliveryapp.EffectReference
}
