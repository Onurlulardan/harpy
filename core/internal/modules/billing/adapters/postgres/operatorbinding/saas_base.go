package operatorbinding

import (
	"context"
	_ "embed"
	"time"

	"github.com/google/uuid"
)

//go:embed sql/platform/saas_base.sql
var saasBaseSQL string

type SaaSBase struct {
	VersionID   uuid.UUID
	Revision    int64
	SHA256      string
	EffectiveAt time.Time
}

type SaaSBaseDatabase interface {
	FetchSaaSBase(context.Context) (SaaSBase, error)
}

func (value saasBaseDatabase) FetchSaaSBase(ctx context.Context) (SaaSBase, error) {
	var result SaaSBase
	err := value.bound.QueryRow(ctx, saasBaseSQL).Scan(
		&result.VersionID, &result.Revision, &result.SHA256, &result.EffectiveAt,
	)
	return result, err
}
