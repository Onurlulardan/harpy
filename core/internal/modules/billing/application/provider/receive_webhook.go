package providerapp

import (
	"context"
	"crypto/sha256"
)

type Ingress struct {
	registry  Registry
	clock     Clock
	ids       IDSource
	protector BodyProtector
	store     WebhookIngressStore
}

func NewIngress(
	registry Registry,
	clock Clock,
	ids IDSource,
	protector BodyProtector,
	store WebhookIngressStore,
) (*Ingress, error) {
	if registry.adapters == nil || nilDependency(clock) || nilDependency(ids) ||
		nilDependency(protector) || nilDependency(store) {
		return nil, ErrConfiguration
	}
	return &Ingress{
		registry: registry, clock: clock, ids: ids, protector: protector, store: store,
	}, nil
}

func (ingress *Ingress) ReceiveWebhook(
	ctx context.Context,
	command ReceiveCommand,
) (ReceiveResult, error) {
	if ingress == nil || ctx == nil || ctx.Err() != nil || !command.valid() {
		return ReceiveResult{}, ErrWebhookRequest
	}
	adapter, err := ingress.registry.Webhook(command.Provider)
	if err != nil {
		return ReceiveResult{}, err
	}
	receivedAt, err := canonicalInstant(ingress.clock.Now())
	if err != nil {
		return ReceiveResult{}, ErrProviderDependency
	}
	body := append([]byte(nil), command.Body...)
	defer clear(body)
	verification := newWebhookVerificationRequest(command.Headers, body, receivedAt)
	verified, err := adapter.VerifyWebhook(ctx, verification)
	if err != nil {
		return ReceiveResult{}, err
	}
	if !verified.valid() {
		return ReceiveResult{}, ErrWebhookMalformed
	}
	receiptID := ingress.ids.New()
	if receiptID == [16]byte{} {
		return ReceiveResult{}, ErrProviderDependency
	}
	associatedData := webhookAssociatedData(receiptID, command.Provider, verified)
	protected, err := ingress.protector.Protect(ctx, associatedData, body)
	if err != nil {
		return ReceiveResult{}, ErrProviderDependency
	}
	digest := sha256.Sum256(body)
	if !protected.valid() || protected.BodySHA256() != digest {
		return ReceiveResult{}, ErrProviderDependency
	}
	processor, err := newWebhookProcessorJob(receiptID, digest, receivedAt)
	if err != nil {
		return ReceiveResult{}, err
	}
	acceptance, err := newWebhookAcceptance(
		receiptID, command.Provider, verified, protected, receivedAt, processor,
	)
	if err != nil {
		return ReceiveResult{}, err
	}
	decision, err := ingress.store.AcceptWebhook(ctx, acceptance)
	if err != nil {
		return ReceiveResult{}, err
	}
	if !decision.valid(receiptID) {
		return ReceiveResult{}, ErrWebhookDecision
	}
	return ReceiveResult{ReceiptID: decision.ReceiptID, Created: decision.Created}, nil
}
