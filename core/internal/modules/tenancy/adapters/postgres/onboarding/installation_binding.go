package onboarding

import (
	"context"
	_ "embed"
	"errors"

	"github.com/google/uuid"
)

//go:embed sql/platform/activation_initialize.sql
var activationInitializeSQL string

type InstallationRow interface{ Scan(...any) error }

type InstallationDatabase struct {
	QueryRow func(context.Context, string, ...any) InstallationRow
}

type InstallationTenant struct {
	OperationID, EnvironmentID, TenantID, OwnerUserID uuid.UUID
	Name, Slug, SchemaName, Kind, BindingPurpose      string
	ManifestSHA256                                    string
}

type InstallationBinding struct {
	ActivateTenant func(context.Context, InstallationTenant) (bool, error)
}

func BindInstallation(database InstallationDatabase) (InstallationBinding, error) {
	if database.QueryRow == nil {
		return InstallationBinding{}, errors.New("tenancy installation database is invalid")
	}
	return InstallationBinding{ActivateTenant: func(
		ctx context.Context, input InstallationTenant,
	) (bool, error) {
		var replay bool
		err := database.QueryRow(ctx, activationInitializeSQL,
			input.OperationID, input.EnvironmentID, input.TenantID, input.Name,
			input.Slug, input.OwnerUserID, input.SchemaName, input.BindingPurpose,
			input.ManifestSHA256, input.Kind).Scan(&replay)
		return replay, err
	}}, nil
}
