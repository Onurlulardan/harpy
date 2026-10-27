package installationactivation

import (
	"context"
	_ "embed"
	"errors"
	"time"

	"github.com/google/uuid"
)

//go:embed sql/platform/assign.sql
var assignSQL string

//go:embed sql/platform/open.sql
var openSQL string

//go:embed sql/platform/assert.sql
var assertSQL string

type Row interface{ Scan(...any) error }

type Database struct {
	QueryRow func(context.Context, string, ...any) Row
}

type Assignment struct {
	OperationID, EnvironmentID, TenantID uuid.UUID
	AuthorityID, AssignmentID            uuid.UUID
	DeploymentMode                       string
	BindingPurpose                       string
	OfferingVersionID                    uuid.UUID
	OfferingRevision                     int64
	OfferingSHA256                       string
	OfferingEffectiveAt                  time.Time
	SourceSequence                       any
	SourceSHA256                         any
	ManifestSHA256                       string
}

type Opening struct {
	OperationID, TenantID uuid.UUID
	ManifestSHA256        string
}

type Binding struct {
	Assert func(context.Context, uuid.UUID, uuid.UUID, string) (bool, error)
	Assign func(context.Context, Assignment) (bool, error)
	Open   func(context.Context, Opening) (int, bool, error)
}

func Bind(database Database) (Binding, error) {
	if database.QueryRow == nil {
		return Binding{}, errors.New("billing installation database is invalid")
	}
	return Binding{
		Assert: func(
			ctx context.Context, operationID, tenantID uuid.UUID, schemaName string,
		) (bool, error) {
			var valid bool
			err := database.QueryRow(ctx, assertSQL,
				operationID, tenantID, schemaName).Scan(&valid)
			return valid, err
		},
		Assign: func(ctx context.Context, input Assignment) (bool, error) {
			var replay bool
			err := database.QueryRow(ctx, assignSQL, input.OperationID,
				input.EnvironmentID, input.TenantID, input.AuthorityID,
				input.AssignmentID, input.DeploymentMode, input.OfferingVersionID,
				input.OfferingRevision, input.OfferingSHA256, input.OfferingEffectiveAt,
				input.SourceSequence, input.SourceSHA256, input.ManifestSHA256,
				input.BindingPurpose,
			).Scan(&replay)
			return replay, err
		},
		Open: func(ctx context.Context, input Opening) (int, bool, error) {
			var count int
			var replay bool
			err := database.QueryRow(ctx, openSQL, input.OperationID,
				input.TenantID, input.ManifestSHA256).Scan(&count, &replay)
			return count, replay, err
		},
	}, nil
}
