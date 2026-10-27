package keyring

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/keywrap"
	downloadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/download"
)

func (ring *Ring) Unwrap(ctx context.Context,
	descriptor downloadapp.Descriptor,
) ([32]byte, error) {
	if ring == nil || ctx == nil || ctx.Err() != nil || descriptor.Object.KeyID == "" {
		return [32]byte{}, downloadapp.ErrUnavailable
	}
	kek, found := ring.keys[descriptor.Object.KeyID]
	if !found {
		return [32]byte{}, downloadapp.ErrIntegrity
	}
	opened, err := keywrap.Unwrap(descriptor.Object.WrappedDEK, kek[:],
		wrapContext(descriptor.TenantID, descriptor.ItemID, descriptor.VersionID,
			descriptor.Object.KeyID))
	if err != nil || len(opened) != 32 {
		clear(opened)
		return [32]byte{}, downloadapp.ErrIntegrity
	}
	var result [32]byte
	copy(result[:], opened)
	clear(opened)
	return result, nil
}

var _ downloadapp.Keyring = (*Ring)(nil)
