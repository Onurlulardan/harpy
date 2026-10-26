package postgres

import (
	"context"
	_ "embed"
	"errors"
	"time"

	layoutapp "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/application"
	"github.com/google/uuid"
)

type InstallationRow interface{ Scan(...any) error }

type InstallationDatabase struct {
	QueryRow func(context.Context, string, ...any) InstallationRow
}

type InstallationRecordInput struct {
	OperationID, TenantID                                  uuid.UUID
	UserRevisionID, OrganizationRevisionID, RoleRevisionID uuid.UUID
	SchemaName                                             string
	InitializedAt                                          time.Time
}

type InstallationBinding struct {
	Store  layoutapp.Store
	Assert func(context.Context, uuid.UUID, uuid.UUID, string) (bool, error)
	Record func(context.Context, InstallationRecordInput) ([32]byte, error)
}

//go:embed sql/tenant/installation_record.sql
var installationRecordSQL string

//go:embed sql/tenant/installation_assert.sql
var installationAssertSQL string

func BindInstallationStore(
	database InstallationDatabase,
) (InstallationBinding, error) {
	if database.QueryRow == nil {
		return InstallationBinding{}, layoutUnavailable()
	}
	allowed := func(ctx context.Context) bool {
		return ctx != nil && ctx.Err() == nil
	}
	store := &layoutStore{
		queryRow: func(
			ctx context.Context, statement string, arguments ...any,
		) row {
			return database.QueryRow(ctx, statement, arguments...)
		},
		commandAllowed: allowed, queryAllowed: allowed,
	}
	return InstallationBinding{Store: store, Assert: func(
		ctx context.Context, operationID, tenantID uuid.UUID, schemaName string,
	) (bool, error) {
		var valid bool
		err := database.QueryRow(ctx, installationAssertSQL,
			operationID, tenantID, schemaName).Scan(&valid)
		return valid, err
	}, Record: func(
		ctx context.Context, input InstallationRecordInput,
	) ([32]byte, error) {
		if !allowed(ctx) || !validInstallationRecord(input) {
			return [32]byte{}, errors.New("system resource installation record is invalid")
		}
		var document []byte
		err := database.QueryRow(ctx, installationRecordSQL,
			input.OperationID, input.TenantID, input.UserRevisionID,
			input.OrganizationRevisionID, input.RoleRevisionID,
			input.InitializedAt, input.SchemaName).Scan(&document)
		if err != nil || len(document) != 32 {
			return [32]byte{}, layoutUnavailable()
		}
		var digest [32]byte
		copy(digest[:], document)
		return digest, nil
	}}, nil
}

func validInstallationRecord(input InstallationRecordInput) bool {
	return input.OperationID != uuid.Nil && input.TenantID != uuid.Nil &&
		input.UserRevisionID != uuid.Nil && input.OrganizationRevisionID != uuid.Nil &&
		input.RoleRevisionID != uuid.Nil && input.SchemaName != "" &&
		!input.InitializedAt.IsZero()
}
