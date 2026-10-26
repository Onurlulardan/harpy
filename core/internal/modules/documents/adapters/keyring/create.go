package keyring

import (
	"context"
	"io"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/keywrap"
	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (ring *Ring) CreateDataKey(ctx context.Context, tenantID, documentID,
	versionID domain.ID,
) (uploadapp.DataKey, error) {
	if ring == nil || ctx == nil || ctx.Err() != nil || ring.randomness == nil ||
		!tenantID.Valid() || !documentID.UUIDv7() || !versionID.UUIDv7() {
		return uploadapp.DataKey{}, uploadapp.ErrUnavailable
	}
	result := uploadapp.DataKey{KeyID: ring.current}
	if _, err := io.ReadFull(ring.randomness, result.Plaintext[:]); err != nil {
		clear(result.Plaintext[:])
		return uploadapp.DataKey{}, err
	}
	kek := ring.keys[ring.current]
	wrapped, err := keywrap.Wrap(result.Plaintext[:], kek[:],
		wrapContext(tenantID, documentID, versionID, ring.current), ring.randomness)
	if err != nil {
		clear(result.Plaintext[:])
		return uploadapp.DataKey{}, err
	}
	result.Wrapped = wrapped
	return result, nil
}
