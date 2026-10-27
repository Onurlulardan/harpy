package application

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"sync"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

type fakeRegistry struct {
	mutex      sync.Mutex
	privateKey ed25519.PrivateKey
	kid        string
	kind       artifact.RegistryKind
	sequence   int64
	acks       map[string]artifact.VerifiedAck
	fail       bool
	exchanges  int
}

func (registry *fakeRegistry) EnsureBinding(
	context.Context, artifact.VerifiedBindingProof,
) error {
	return nil
}

func (registry *fakeRegistry) EnsureProviderBinding(
	context.Context, ports.ProviderBinding,
) error {
	return nil
}

func (registry *fakeRegistry) ExchangeProposal(
	_ context.Context, proposal artifact.SignedProposal,
) ([]byte, error) {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()
	if registry.fail {
		return nil, fmt.Errorf("registry unavailable")
	}
	if ack, exists := registry.acks[proposal.Value.OperationID]; exists {
		if ack.Value.ProposalSHA256 != proposal.SHA256 {
			return nil, fmt.Errorf("operation replay conflicts")
		}
		return append([]byte(nil), ack.Bytes...), nil
	}
	registry.sequence++
	registry.exchanges++
	ack, err := artifact.SignAck(artifact.Ack{
		Schema: artifact.AckSchema, RegistryKind: registry.kind,
		RegistryBindingID: proposal.Value.RegistryBindingID,
		ScopeKind:         proposal.Value.ScopeKind, ScopeID: proposal.Value.ScopeID,
		RegistrySequence: registry.sequence, OperationID: proposal.Value.OperationID,
		ProposalSHA256: proposal.SHA256, Generation: proposal.Value.Generation,
		PreviousHeadSHA256: proposal.Value.PreviousHeadSHA256,
		StateSHA256:        proposal.Value.StateSHA256, IssuedAt: testNow.Format(time.RFC3339),
		RegistryKID: registry.kid,
	}, proposal, registry.privateKey)
	if err != nil {
		return nil, err
	}
	registry.acks[proposal.Value.OperationID] = ack
	return append([]byte(nil), ack.Bytes...), nil
}

func (registry *fakeRegistry) LatestHead(
	context.Context, ports.HeadKey,
) (ports.ExternalHead, error) {
	return ports.ExternalHead{}, ports.ErrNotFound
}
