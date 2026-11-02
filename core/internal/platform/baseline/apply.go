package baseline

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

//go:embed sql/20261028000000_platform_baseline.sql
var platformScript string

//go:embed sql/20261028000000_tenant_baseline.sql
var tenantScript string

const tenantSchemaToken = "__tenant_schema__"

func ApplyPlatform(ctx context.Context, dsn string) error {
	return apply(ctx, dsn, platformScript)
}

func ApplyTenant(ctx context.Context, dsn, schema string) error {
	if !validSchema(schema) {
		return fmt.Errorf("tenant schema name is invalid")
	}
	script := strings.ReplaceAll(tenantScript, tenantSchemaToken, schema)
	return apply(ctx, dsn, script)
}

func apply(ctx context.Context, dsn, script string) error {
	body := stripGoose(script)
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return fmt.Errorf("open baseline database: %w", err)
	}
	defer conn.Close(ctx)
	if _, err := conn.Exec(ctx, body); err != nil {
		return fmt.Errorf("apply baseline: %w", err)
	}
	return nil
}

func stripGoose(script string) string {
	var b strings.Builder
	for _, line := range strings.Split(script, "\n") {
		switch strings.TrimSpace(line) {
		case "-- +goose Up", "-- +goose StatementBegin", "-- +goose StatementEnd":
			continue
		default:
			b.WriteString(line)
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func validSchema(schema string) bool {
	if schema == "" || len(schema) > 63 {
		return false
	}
	for index, char := range schema {
		ok := char == '_' || (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9' && index > 0)
		if !ok {
			return false
		}
	}
	return schema != tenantSchemaToken
}
