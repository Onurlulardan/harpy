package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type databaseColumn struct {
	schema, table, column, dataType string
}

const inventoryColumnsSQL = `
SELECT namespace.nspname,relation.relname,attribute.attname,type_row.typname
FROM pg_catalog.pg_attribute AS attribute
JOIN pg_catalog.pg_class AS relation ON relation.oid=attribute.attrelid
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=relation.relnamespace
JOIN pg_catalog.pg_type AS type_row ON type_row.oid=attribute.atttypid
WHERE attribute.attnum>0 AND NOT attribute.attisdropped
  AND relation.relkind IN ('r','p')
  AND namespace.nspname NOT IN ('pg_catalog','information_schema')
  AND namespace.nspname NOT LIKE 'pg_toast%'
  AND type_row.typname IN ('text','varchar','bpchar','json','jsonb','bytea','citext')
ORDER BY namespace.nspname,relation.relname,attribute.attnum`

func scanPostgres(ctx context.Context, secret []byte, dsnFile string) error {
	dsn, err := readDSNFile(dsnFile)
	if err != nil {
		return err
	}
	defer clear(dsn)
	configuration, err := pgxpool.ParseConfig(string(dsn))
	if err != nil {
		return fmt.Errorf("PostgreSQL inventory configuration is invalid")
	}
	configuration.MaxConns = 2
	pool, err := pgxpool.NewWithConfig(ctx, configuration)
	if err != nil {
		return fmt.Errorf("PostgreSQL inventory connection failed")
	}
	defer pool.Close()
	columns, err := loadDatabaseColumns(ctx, pool)
	if err != nil {
		return err
	}
	for _, column := range columns {
		found, queryErr := databaseColumnContains(ctx, pool, column, secret)
		if queryErr != nil {
			return fmt.Errorf("scan PostgreSQL column %s.%s.%s: %w",
				column.schema, column.table, column.column, queryErr)
		}
		if found {
			return fmt.Errorf("plaintext secret found in PostgreSQL column %s.%s.%s",
				column.schema, column.table, column.column)
		}
	}
	return nil
}

func loadDatabaseColumns(ctx context.Context, pool *pgxpool.Pool) ([]databaseColumn, error) {
	rows, err := pool.Query(ctx, inventoryColumnsSQL)
	if err != nil {
		return nil, fmt.Errorf("PostgreSQL inventory catalog is unavailable")
	}
	defer rows.Close()
	columns := make([]databaseColumn, 0, 256)
	for rows.Next() {
		var column databaseColumn
		if err = rows.Scan(&column.schema, &column.table, &column.column, &column.dataType); err != nil {
			return nil, fmt.Errorf("PostgreSQL inventory catalog is invalid")
		}
		columns = append(columns, column)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("PostgreSQL inventory catalog read failed")
	}
	return columns, nil
}

func databaseColumnContains(
	ctx context.Context,
	pool *pgxpool.Pool,
	column databaseColumn,
	secret []byte,
) (bool, error) {
	query := databaseColumnQuery(column)
	var found bool
	err := pool.QueryRow(ctx, query, secret).Scan(&found)
	return found, err
}

func databaseColumnQuery(column databaseColumn) string {
	relation := pgx.Identifier{column.schema, column.table}.Sanitize()
	field := pgx.Identifier{column.column}.Sanitize()
	value := field
	if column.dataType != "bytea" {
		value = "pg_catalog.convert_to(" + field + "::text,'UTF8')"
	}
	return "SELECT EXISTS (SELECT 1 FROM " + relation + " WHERE " + field +
		" IS NOT NULL AND position($1::bytea IN " + value + ")>0)"
}

func readDSNFile(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 ||
		info.Size() < 1 || info.Size() > 4096 ||
		runtime.GOOS != "windows" && info.Mode().Perm()&0o077 != 0 {
		return nil, fmt.Errorf("PostgreSQL inventory DSN is invalid")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("PostgreSQL inventory DSN is unavailable")
	}
	defer file.Close()
	value, err := io.ReadAll(io.LimitReader(file, 4097))
	value = bytes.TrimSpace(value)
	if err != nil || len(value) == 0 || len(value) > 4096 || bytes.IndexByte(value, 0) >= 0 {
		clear(value)
		return nil, fmt.Errorf("PostgreSQL inventory DSN is invalid")
	}
	return value, nil
}
