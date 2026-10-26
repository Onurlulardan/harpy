package systemresourcelayout

import (
	"context"
	"time"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

type InstallationActivationRow interface{ Scan(...any) error }

type InstallationActivationExecutor struct {
	QueryRow func(context.Context, string, ...any) InstallationActivationRow
}

type InstallationBaselineInput struct {
	OperationID, TenantID [16]byte
	ActivatedAt           time.Time
	SchemaName            string
	UserCatalog           layoutapi.PresentationFieldCatalog
	OrganizationCatalog   layoutapi.PresentationFieldCatalog
	RoleCatalog           layoutapi.PresentationFieldCatalog
}

func validInstallationBaseline(input InstallationBaselineInput) bool {
	return input.OperationID != [16]byte{} && input.TenantID != [16]byte{} &&
		!input.ActivatedAt.IsZero() && input.ActivatedAt.Location() == time.UTC &&
		len(input.SchemaName) >= 3 && len(input.SchemaName) <= 63 &&
		input.UserCatalog.Validate() == nil &&
		input.OrganizationCatalog.Validate() == nil &&
		input.RoleCatalog.Validate() == nil
}
