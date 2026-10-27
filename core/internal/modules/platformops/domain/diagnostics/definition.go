package diagnostics

import (
	"errors"
	"time"
)

type ID string
type Consumer string

const (
	DatabaseStats     ID = "database_stats_v1"
	SchemaMetadata    ID = "schema_metadata_v1"
	RecoveryReadiness ID = "recovery_readiness_v1"

	Interactive Consumer = "interactive"
	Backup      Consumer = "database_backup"
	Restore     Consumer = "database_restore"
	Admission   Consumer = "admission"
)

type Field struct {
	Name  string
	Value string
}

const (
	CheckTimeout = 10 * time.Second
	CacheTTL     = 15 * time.Minute
)

var (
	ErrUnknownID       = errors.New("diagnostic id is not registered")
	ErrInvalidConsumer = errors.New("diagnostic consumer is invalid")
	ErrUnavailable     = errors.New("diagnostic is unavailable")
)

type Definition struct {
	ID                 ID
	OutputSchemaSHA256 string
	MaximumRows        int
	MaximumBytes       int
}

func Resolve(id ID) (Definition, error) {
	definition, found := registry[id]
	if !found {
		return Definition{}, ErrUnknownID
	}
	return definition, nil
}

func ValidateConsumer(consumer Consumer) error {
	switch consumer {
	case Interactive, Backup, Restore, Admission:
		return nil
	default:
		return ErrInvalidConsumer
	}
}

func IsFailOpen(consumer Consumer) bool {
	return consumer == Interactive
}
