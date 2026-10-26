package ports

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

type ProviderOutcome string

const (
	ProviderAccepted  ProviderOutcome = "accepted"
	ProviderTransient ProviderOutcome = "transient"
	ProviderPermanent ProviderOutcome = "permanent"
	ProviderUnknown   ProviderOutcome = "unknown"
)

type ProviderRequest struct {
	DeliveryID   domain.ID
	Target       []byte
	Subject      string
	Body         string
	TemplateKind domain.TemplateKind
	Metadata     ProviderMetadata
	Credential   []byte
	Requested    time.Time
}

type ProviderResult struct {
	Outcome                 ProviderOutcome
	Code, ResponseClass     string
	ProviderReferenceDigest [32]byte
	StatusCode              string
	CompletedAt             time.Time
}

func (value ProviderResult) Definitive() bool {
	return value.Outcome == ProviderAccepted || value.Outcome == ProviderTransient ||
		value.Outcome == ProviderPermanent
}

type ProviderSender interface {
	Send(context.Context, ProviderRequest) ProviderResult
}

type ProviderRegistry interface {
	Resolve(ProviderKind) (ProviderSender, error)
}

type DeliveryTargetResolver interface {
	ResolveDeliveryTarget(context.Context, domain.ID, domain.ID,
		domain.Channel) ([]byte, error)
}
