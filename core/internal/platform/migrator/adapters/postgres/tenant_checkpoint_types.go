package postgres

import (
	"context"
	"database/sql"
	"time"
)

type TenantMigrationTarget struct {
	TenantID   string
	SchemaName string
}

type TenantSessionProof struct {
	BackendPID int64
	SearchPath string
	SchemaName string
}

type TenantCheckpointIdentity struct {
	TenantID     string
	Owner        string
	Version      int64
	Origin       string
	SelectedPath string
	SelectedHash string
	RunID        string
	Attempt      int
}

type TenantCheckpoint struct {
	TenantCheckpointIdentity
	State      string
	StartedAt  sql.NullTime
	FinishedAt sql.NullTime
	ErrorCode  sql.NullString
	UpdatedAt  time.Time
}

type tenantRowQuerier interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type tenantRowsQuerier interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}
