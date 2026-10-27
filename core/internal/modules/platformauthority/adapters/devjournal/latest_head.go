package devjournal

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func (registry *Registry) LatestHead(
	ctx context.Context,
	key ports.HeadKey,
) (ports.ExternalHead, error) {
	var result ports.ExternalHead
	err := registry.withLock(ctx, func() error {
		state, err := registry.loadState(ctx)
		if err != nil {
			return err
		}
		head, exists := state.heads[headMapKey(
			key.RegistryBindingID, key.ScopeKind, key.ScopeID,
		)]
		if !exists {
			return ports.ErrNotFound
		}
		result = head
		return nil
	})
	if err != nil {
		return ports.ExternalHead{}, fmt.Errorf("read development authority head: %w", err)
	}
	return result, nil
}
