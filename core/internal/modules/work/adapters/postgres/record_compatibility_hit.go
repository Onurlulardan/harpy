package postgres

import (
	"context"
	_ "embed"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

//go:embed sql/tenant/record_compatibility_hit.sql
var recordCompatibilityHitSQL string

type CompatibilityRecorder struct {
	access *ownerAccess
}

func NewCompatibilityRecorder() (*CompatibilityRecorder, error) {
	access, err := newOwnerAccess()
	if err != nil || recordCompatibilityHitSQL == "" {
		return nil, application.ErrUnavailable
	}
	return &CompatibilityRecorder{access: access}, nil
}

func (recorder *CompatibilityRecorder) Record(ctx context.Context, local ports.Scope,
	route, consumer, build string,
) error {
	bridge, ok := local.(*scopeBridge)
	if recorder == nil || !ok || !scopeUsable(ctx, bridge) || bridge.SnapshotQuery() ||
		!compatibilityText(route, 200) || !compatibilityText(consumer, 200) ||
		!compatibilityText(build, 200) {
		return application.ErrInvalid
	}
	database, err := recorder.access.bind(ctx, bridge)
	if err != nil {
		return application.ErrUnavailable
	}
	result, err := database.run(ctx, recordCompatibilityHitSQL,
		route, consumer, build)
	if err != nil || result.RowsAffected() != 1 {
		return application.ErrUnavailable
	}
	return nil
}

func compatibilityText(value string, maximum int) bool {
	return value != "" && len(value) <= maximum && strings.TrimSpace(value) == value
}
