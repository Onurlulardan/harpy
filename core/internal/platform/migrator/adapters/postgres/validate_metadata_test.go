package postgres

import (
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestValidateMigrationMetadataRejectsNonCanonicalSelection(t *testing.T) {
	tests := []struct {
		name   string
		mutate func([]postgresv1.SelectedMigration)
	}{
		{"mixed origin", func(value []postgresv1.SelectedMigration) {
			value[1].Origin = postgresv1.DatabaseOriginV1Adopted
		}},
		{"unordered version", func(value []postgresv1.SelectedMigration) {
			value[0], value[1] = value[1], value[0]
		}},
		{"invalid owner", func(value []postgresv1.SelectedMigration) {
			value[0].Owner = "Platform"
		}},
		{"invalid scope", func(value []postgresv1.SelectedMigration) {
			value[0].Scope = "global"
		}},
		{"invalid transaction mode", func(value []postgresv1.SelectedMigration) {
			value[0].TransactionMode = "implicit"
		}},
		{"invalid phase", func(value []postgresv1.SelectedMigration) {
			value[0].Phase = "down"
		}},
		{"generation regression", func(value []postgresv1.SelectedMigration) {
			value[1].TargetDatabaseGeneration = 0
		}},
		{"path traversal", func(value []postgresv1.SelectedMigration) {
			value[0].Path = "migrations/../20260714000100_foundation.fresh.sql"
		}},
		{"wrong path origin", func(value []postgresv1.SelectedMigration) {
			value[0].Path = "migrations/platform/20260714000100_foundation.v1_adopted.sql"
		}},
		{"invalid digest", func(value []postgresv1.SelectedMigration) {
			value[0].SHA256 = strings.Repeat("A", 64)
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			selected := validSelectedMigrations()
			test.mutate(selected)
			if _, err := validateMigrationMetadata(testRunID, selected); err == nil {
				t.Fatal("unsafe migration metadata accepted")
			}
		})
	}
}
