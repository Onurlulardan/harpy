package postgres

import (
	"fmt"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func NewStore(runID string, selected []postgresv1.SelectedMigration) (*Store, error) {
	metadata, err := validateMigrationMetadata(runID, selected)
	if err != nil {
		return nil, fmt.Errorf("create PostgreSQL migration history store: %w", err)
	}
	return &Store{metadataByVersion: metadata}, nil
}
