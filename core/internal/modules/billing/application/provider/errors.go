package providerapp

import "errors"

var (
	ErrConfiguration         = errors.New("billing provider runtime is misconfigured")
	ErrCapability            = errors.New("billing provider capability is unavailable")
	ErrWebhookRequest        = errors.New("billing webhook request is invalid")
	ErrWebhookAuthentication = errors.New("billing webhook authentication failed")
	ErrWebhookReplay         = errors.New("billing webhook replay policy failed")
	ErrWebhookMalformed      = errors.New("billing webhook event is malformed")
	ErrWebhookMapping        = errors.New("billing webhook provider object mapping failed")
	ErrWebhookConflict       = errors.New("billing webhook event identity was reused")
	ErrWebhookDecision       = errors.New("billing webhook durable decision is invalid")
	ErrWebhookProcessing     = errors.New("billing webhook processing failed")
	ErrNormalizedEvent       = errors.New("billing normalized provider event is invalid")
	ErrEffectResult          = errors.New("billing provider effect result is invalid")
	ErrProviderDependency    = errors.New("billing provider dependency failed")
	ErrCommandRequest        = errors.New("billing provider command request is invalid")
	ErrCommandTarget         = errors.New("billing provider command target is invalid")
	ErrCommandConflict       = errors.New("billing provider command identity was reused")
	ErrCommandState          = errors.New("billing provider command state is invalid")
)
