package iam

import (
	"context"
	"time"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	lifecycledeliveryv1 "github.com/Onurlulardan/harpy/contracts/lifecycledelivery/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/smtp"
	lifecycledeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycledelivery"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
	"github.com/google/uuid"
)

type lifecycleCredential interface {
	RevealOnce(func([]byte) error) error
	Destroy()
}

type lifecycleCredentialOpener interface {
	Open(context.Context, uuid.UUID, lifecycledeliveryv1.Envelope) (lifecycleCredential, error)
}

type lifecycleSMTPSender interface {
	StableMessageID(string) (string, error)
	Send(context.Context, lifecycledeliveryv1.Purpose, smtp.Message) smtp.Result
}

type lifecycleDeliveryState interface {
	Claim(context.Context, lifecycledeliveryapp.ClaimCommand) (
		lifecycledeliveryapp.ClaimDecision, string, error)
	Release(context.Context, lifecycledeliveryapp.EffectReference) error
	MarkUnknown(context.Context, lifecycledeliveryapp.EffectReference) error
	Finalize(context.Context, lifecycledeliveryapp.FinalizeCommand) (lifecycledeliveryapp.StoredResult, error)
}

type lifecycleDeliveryProvider struct {
	identity effectexecution.Identity
	state    lifecycleDeliveryState
	opener   lifecycleCredentialOpener
	smtp     lifecycleSMTPSender
	now      func() time.Time
}

func (provider *lifecycleDeliveryProvider) Identity() effectexecution.Identity {
	if provider == nil {
		return effectexecution.Identity{}
	}
	return provider.identity
}

var _ effectexecution.Provider = (*lifecycleDeliveryProvider)(nil)

type lifecycleDeliveryRequest struct {
	request     effectsv1.Request
	tenantID    uuid.UUID
	requestedAt time.Time
	messageID   string
	reference   lifecycledeliveryapp.EffectReference
}
