package adoption

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
	"github.com/google/uuid"
)

type Zero struct{ manifest Manifest }

func NewZeroSource(snapshotID string) (*Zero, error) {
	manifest := Manifest{Kind: ZeroSource, Schema: "none", Relation: "none",
		SnapshotID: snapshotID,
		RootSHA256: sha256.Sum256([]byte("harpy.usage.zero-source.v1:" + snapshotID))}
	if manifest.Validate() != nil {
		return nil, fmt.Errorf("usage zero source is invalid")
	}
	return &Zero{manifest: manifest}, nil
}

func (source *Zero) Inspect(ctx context.Context) (Manifest, error) {
	if source == nil || ctx == nil || ctx.Err() != nil || source.manifest.Validate() != nil {
		return Manifest{}, domain.ErrUnavailable
	}
	return source.manifest, nil
}

func (*Zero) Page(context.Context, string, *uuid.UUID, int) (SourcePage, error) {
	return SourcePage{}, domain.ErrSourceDrift
}

var _ Source = (*Zero)(nil)
