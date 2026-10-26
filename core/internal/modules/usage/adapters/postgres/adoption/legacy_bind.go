package adoption

import (
	"context"
	"fmt"
	"regexp"

	app "github.com/Onurlulardan/harpy/core/internal/modules/usage/application/adoption"
	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
	"github.com/google/uuid"
)

var legacyRelationPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

type Rows interface {
	Next() bool
	Scan(...any) error
	Err() error
	Close()
}

type LegacyDatabase interface {
	QueryRow(context.Context, string, ...any) Row
	Query(context.Context, string, ...any) (Rows, error)
}

type LegacyTransactionRunner func(
	context.Context, string, func(LegacyDatabase) error,
) error

type LegacySource struct {
	run                          LegacyTransactionRunner
	schema, relation             string
	environmentID, defaultTenant uuid.UUID
	manifestSQL, pageSQL         string
}

func NewLegacySourceForProfile(run LegacyTransactionRunner, schema, relation string,
	environmentID, defaultTenant uuid.UUID,
) (*LegacySource, error) {
	if run == nil || !legacyRelationPattern.MatchString(schema) ||
		relation != "trace_logs" || environmentID == uuid.Nil ||
		defaultTenant == uuid.Nil {
		return nil, fmt.Errorf("usage legacy source profile is invalid")
	}
	return &LegacySource{run: run, schema: schema, relation: relation,
		environmentID: environmentID, defaultTenant: defaultTenant,
		manifestSQL: legacyManifestSQL, pageSQL: legacyPageSQL}, nil
}

func (source *LegacySource) Inspect(ctx context.Context) (app.Manifest, error) {
	if source == nil || source.run == nil || ctx == nil || ctx.Err() != nil {
		return app.Manifest{}, domain.ErrUnavailable
	}
	var manifest app.Manifest
	err := source.run(ctx, source.schema, func(database LegacyDatabase) error {
		var err error
		manifest, err = source.inspect(ctx, database)
		return err
	})
	if err != nil {
		return app.Manifest{}, domain.ErrUnavailable
	}
	return manifest, nil
}

var _ app.Source = (*LegacySource)(nil)
