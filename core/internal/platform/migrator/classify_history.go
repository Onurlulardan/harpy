package migrator

import (
	"errors"
	"fmt"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	"github.com/google/uuid"
)

func classifyHistory(
	history map[int64]historyRow,
	selected []postgresv1.SelectedMigration,
	supersessions map[int64]historySupersession,
) (int64, []postgresv1.SelectedMigration, error) {
	expected := make(map[int64]postgresv1.SelectedMigration, len(selected))
	for _, migration := range selected {
		expected[migration.Version] = migration
	}
	for version, row := range history {
		migration, known := expected[version]
		if !known {
			return 0, nil, fmt.Errorf("history version %d is absent from release", version)
		}
		if err := verifyHistoryRow(row, migration, supersessions); err != nil {
			return 0, nil, err
		}
	}
	current := int64(0)
	missingSeen := false
	pending := make([]postgresv1.SelectedMigration, 0)
	for _, migration := range selected {
		_, applied := history[migration.Version]
		if !applied {
			missingSeen = true
			pending = append(pending, migration)
			continue
		}
		if missingSeen {
			return 0, nil, errors.Join(
				ErrOutOfOrder,
				fmt.Errorf("version %d is applied after a missing version", migration.Version),
			)
		}
		current = migration.Version
	}
	return current, pending, nil
}

func verifyHistoryRow(
	row historyRow,
	selected postgresv1.SelectedMigration,
	supersessions map[int64]historySupersession,
) error {
	parsedRunID, err := uuid.Parse(row.runID)
	if err != nil || parsedRunID.String() != row.runID || parsedRunID == uuid.Nil {
		return fmt.Errorf("history version %d run ID is invalid", row.version)
	}
	if row.recordKind != "migration" || !row.isApplied ||
		row.owner != selected.Owner || row.scope != selected.Scope ||
		row.origin != string(selected.Origin) ||
		row.targetDatabaseGeneration != selected.TargetDatabaseGeneration ||
		row.nextMinimumRuntimeGeneration != selected.NextMinimumRuntimeGeneration {
		return fmt.Errorf("history version %d differs from verified metadata", row.version)
	}
	if row.selectedPath == selected.Path && row.selectedSHA256 == selected.SHA256 {
		return nil
	}
	supersession, exists := supersessions[selected.Version]
	if !exists || !supersession.matches(selected, row) {
		return fmt.Errorf("history version %d differs from verified metadata", row.version)
	}
	return nil
}

func (value historySupersession) matches(
	selected postgresv1.SelectedMigration,
	row historyRow,
) bool {
	if !value.matchesCurrent(selected) {
		return false
	}
	for _, identity := range value.superseded {
		if row.selectedPath == identity.path &&
			row.selectedSHA256 == identity.sha256 {
			return true
		}
	}
	return false
}
