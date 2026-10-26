package api

import (
	"context"
	"time"
)

type DiagnosticID string
type DiagnosticConsumer string

const (
	DiagnosticDatabaseStats     DiagnosticID = "database_stats_v1"
	DiagnosticSchemaMetadata    DiagnosticID = "schema_metadata_v1"
	DiagnosticRecoveryReadiness DiagnosticID = "recovery_readiness_v1"

	ConsumerInteractive DiagnosticConsumer = "interactive"
	ConsumerBackup      DiagnosticConsumer = "database_backup"
	ConsumerRestore     DiagnosticConsumer = "database_restore"
	ConsumerAdmission   DiagnosticConsumer = "admission"
)

type DiagnosticRequest struct {
	ID       DiagnosticID
	Consumer DiagnosticConsumer
}

type DiagnosticField struct {
	Name  string
	Value string
}

type DiagnosticSnapshot struct {
	ID                 DiagnosticID
	OutputSchemaSHA256 string
	ObservedAt         time.Time
	Fields             []DiagnosticField
	Cached             bool
	Stale              bool
}

type Diagnostics interface {
	ReadDiagnostic(context.Context, DiagnosticRequest) (DiagnosticSnapshot, error)
}
