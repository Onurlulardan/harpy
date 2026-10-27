package postgres

import postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"

type migrationMetadata struct {
	owner                        string
	scope                        string
	version                      int64
	origin                       postgresv1.DatabaseOrigin
	selectedPath                 string
	selectedSHA256               string
	runID                        string
	targetDatabaseGeneration     int64
	nextMinimumRuntimeGeneration int64
}
