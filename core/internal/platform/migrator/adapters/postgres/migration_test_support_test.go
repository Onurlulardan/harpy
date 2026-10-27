package postgres

import (
	"strings"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

const testRunID = "11111111-1111-4111-8111-111111111111"

func validSelectedMigrations() []postgresv1.SelectedMigration {
	digest := strings.Repeat("a", 64)
	return []postgresv1.SelectedMigration{
		{
			Owner: "platform_migrator", Scope: "platform", Version: 20260714000100,
			TransactionMode: "transactional", Phase: "foundation",
			TargetDatabaseGeneration: 1, NextMinimumRuntimeGeneration: 1,
			Origin: postgresv1.DatabaseOriginFresh,
			Path:   "migrations/platform/20260714000100_foundation.fresh.sql",
			SHA256: digest,
		},
		{
			Owner: "platform_migrator", Scope: "platform", Version: 20260714000200,
			TransactionMode: "transactional", Phase: "expand",
			TargetDatabaseGeneration: 2, NextMinimumRuntimeGeneration: 1,
			Origin: postgresv1.DatabaseOriginFresh,
			Path:   "migrations/platform/20260714000200_expand.fresh.sql",
			SHA256: digest,
		},
	}
}
