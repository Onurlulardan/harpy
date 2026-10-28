package migrationmanifest

import (
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

var logicalNamePattern = regexp.MustCompile(`^[a-z][a-z0-9]*(?:_[a-z0-9]+)*$`)

type documentIdentity struct {
	version     int64
	logicalName string
	origin      postgresv1.DatabaseOrigin
}

func parseDocumentIdentity(name string) (documentIdentity, error) {
	base := path.Base(name)
	var origin postgresv1.DatabaseOrigin
	var suffix string
	for _, candidate := range []postgresv1.DatabaseOrigin{
		postgresv1.DatabaseOriginFresh,
		postgresv1.DatabaseOriginV1Adopted,
	} {
		candidateSuffix := "." + string(candidate) + ".sql"
		if strings.HasSuffix(base, candidateSuffix) {
			origin, suffix = candidate, candidateSuffix
			break
		}
	}
	if !origin.Valid() || path.Base(path.Dir(name)) != string(origin) {
		return documentIdentity{}, fmt.Errorf("migration path does not bind origin directory")
	}
	logicalBase := strings.TrimSuffix(base, suffix)
	if len(logicalBase) < 16 || logicalBase[14] != '_' {
		return documentIdentity{}, fmt.Errorf("migration basename does not bind UTC version")
	}
	version, err := strconv.ParseInt(logicalBase[:14], 10, 64)
	if err != nil || !logicalNamePattern.MatchString(logicalBase[15:]) {
		return documentIdentity{}, fmt.Errorf("migration basename is not canonical")
	}
	return documentIdentity{
		version: version, logicalName: logicalBase[15:], origin: origin,
	}, nil
}
