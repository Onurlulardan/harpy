package goose

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"

	"github.com/Onurlulardan/harpy/core/internal/platform/foundation"
	pressly "github.com/pressly/goose/v3"
)

func UpPlatform(ctx context.Context, cfg foundation.Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	dsn, err := gooseDSN(cfg)
	if err != nil {
		return err
	}
	source, err := fs.Sub(platformRoot, "sql/platform")
	if err != nil {
		return fmt.Errorf("open platform goose source: %w", err)
	}
	return up(ctx, dsn, source, foundation.RoleOwner, foundation.PlatformSchema)
}

func up(ctx context.Context, dsn string, source fs.FS, role, schema string) error {
	database, err := openGoose(ctx, dsn, schema)
	if err != nil {
		return err
	}
	defer database.Close()
	if role != "" {
		if err = setRole(ctx, database, role); err != nil {
			return err
		}
	}
	if err = applySearchPath(ctx, database, schema); err != nil {
		return err
	}
	provider, err := pressly.NewProvider(pressly.DialectPostgres, database, source)
	if err != nil {
		return err
	}
	_, err = provider.Up(ctx)
	return err
}

func setRole(ctx context.Context, database *sql.DB, role string) error {
	_, err := database.ExecContext(ctx, "SET ROLE "+pgIdent(role))
	if err != nil {
		return fmt.Errorf("set goose role: %w", err)
	}
	return nil
}

func pgIdent(value string) string {
	return `"` + replaceQuotes(value) + `"`
}

func replaceQuotes(value string) string {
	out := make([]byte, 0, len(value))
	for i := 0; i < len(value); i++ {
		if value[i] == '"' {
			out = append(out, '"', '"')
			continue
		}
		out = append(out, value[i])
	}
	return string(out)
}
