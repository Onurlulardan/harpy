package onboarding

import (
	"context"
	"regexp"
)

var activationSchemaPattern = regexp.MustCompile(`^t_[a-z2-7]{26}$`)

func (transaction *activationTransaction) bindTenantSchema(
	ctx context.Context,
	schema string,
) error {
	if !transaction.usable() || ctx == nil || ctx.Err() != nil ||
		!activationSchemaPattern.MatchString(schema) {
		return errUnavailable
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
		return err
	}
	transaction.schema = schema
	return nil
}
