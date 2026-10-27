package migrator

import (
	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

type runIdentity struct {
	releaseID                string
	environmentID            string
	databaseName             string
	databaseOrigin           postgresv1.DatabaseOrigin
	databaseOwner            string
	databaseEncoding         string
	databaseLocaleProvider   string
	databaseCollation        string
	databaseCType            string
	dataChecksums            bool
	manifestSHA256           string
	extensionSHA256          string
	rolesSHA256              string
	privilegesSHA256         string
	ownershipSHA256          string
	authoritySHA256          string
	databaseSystemID         string
	databaseOID              uint32
	databaseGeneration       int64
	minimumRuntimeGeneration int64
}

type runRecord struct {
	runID       string
	command     string
	state       string
	failureCode string
}
