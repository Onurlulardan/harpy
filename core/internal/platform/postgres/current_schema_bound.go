package postgres

import "context"

func (executor BoundExecutor) CurrentSchema(ctx context.Context) (string, error) {
	var schema string
	if err := executor.QueryRow(ctx, currentBoundSchemaSQL).Scan(&schema); err != nil {
		return "", err
	}
	if !schemaNamePattern.MatchString(schema) {
		return "", ErrTenantSchemaInvalid
	}
	return schema, nil
}
