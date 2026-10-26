package devjournal

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func (registry *Registry) EnsureBinding(
	ctx context.Context,
	proof artifact.VerifiedBindingProof,
) error {
	return registry.ensureVerifiedBinding(ctx, proof.Bytes)
}

func (registry *Registry) EnsureProviderBinding(
	context.Context,
	ports.ProviderBinding,
) error {
	return fmt.Errorf("SaaS provider binding is forbidden in development journal")
}

func (registry *Registry) ensureVerifiedBinding(
	ctx context.Context,
	proofBytes []byte,
) error {
	return registry.withLock(ctx, func() error {
		value, err := decodeBindingProof(proofBytes)
		if err != nil {
			return err
		}
		key, err := registry.resolver.ResolveProposalKey(ctx, value.ControlKID)
		if err != nil {
			return fmt.Errorf("development authority proposal key is unavailable")
		}
		verified, err := verifyBinding(proofBytes, key)
		if err != nil {
			return err
		}
		state, err := registry.loadState(ctx)
		if err != nil {
			return err
		}
		if existing, exists := state.bindings[verified.Value.LicenseBindingID]; exists {
			if existing.proof.SHA256 == verified.SHA256 {
				return nil
			}
			return fmt.Errorf("development authority binding conflicts")
		}
		if state.environments[verified.Value.EnvironmentID] != "" {
			return fmt.Errorf("development authority environment cannot be rebound")
		}
		environmentID := verified.Value.EnvironmentID
		return appendRecord(registry.path, journalRecord{
			Kind: "binding", RegistryBindingID: verified.Value.LicenseBindingID,
			EnvironmentID: &environmentID,
			Artifact:      base64.RawURLEncoding.EncodeToString(verified.Bytes),
		})
	})
}
