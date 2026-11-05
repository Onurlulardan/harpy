package onboarding

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

var activationSchemaPattern = regexp.MustCompile(`^tenant_[0-9a-f]{32}$`)

func (transaction *activationTransaction) bindTenantSchema(
	ctx context.Context,
	schema string,
) error {
	if !transaction.usable() || ctx == nil || ctx.Err() != nil {
		return errUnavailable
	}
	if !activationSchemaPattern.MatchString(schema) {
		shown := schema
		if shown == "" {
			shown = "<empty>"
		}
		return fmt.Errorf("%w: %s", platformpostgres.ErrTenantSchemaInvalid, shown)
	}
	transaction.mutex.Lock()
	defer transaction.mutex.Unlock()
	if transaction.schema != "" {
		if transaction.schema == schema {
			return nil
		}
		return errUnavailable
	}
	if transaction.bindSchema == nil {
		return errUnavailable
	}
	if err := transaction.bindSchema(ctx, schema); err != nil {
		if errors.Is(err, platformpostgres.ErrTenantSchemaInvalid) {
			return fmt.Errorf("%w: %s", platformpostgres.ErrTenantSchemaInvalid, schema)
		}
		return err
	}
	transaction.schema = schema
	return nil
}
