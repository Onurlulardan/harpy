package adoption

import (
	"context"
	"time"

	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/adoption"
)

func (storage *Store) Accept(ctx context.Context, authority adoptionapp.Authority,
	digest [32]byte, finalLSN string, now time.Time,
) (adoptionapp.Run, error) {
	if !storage.usable(ctx) || !storage.owns(authority) || digest == [32]byte{} ||
		finalLSN == "" || now.IsZero() {
		return adoptionapp.Run{}, adoptionapp.ErrInvalid
	}
	result, _, err := storage.scanRun(storage.queryRow(ctx, acceptSQL,
		authority.RunID.String(), int64(authority.CutoverGeneration),
		authority.ManifestSHA256[:], digest[:], finalLSN, now))
	return result, err
}
