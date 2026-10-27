package tenantsettingspostgres

import (
	"context"

	settings "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/tenantsettings"
)

func nullableName(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func validMutation(value settings.RegistryMutation) bool {
	return value.ExpectedRevision > 0 && value.IdempotencyKey != "" &&
		value.RequestSHA256 != [32]byte{}
}

func contextUsable(ctx context.Context) bool {
	return ctx != nil && ctx.Err() == nil
}
