package postgresv1

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	PostgresImageDigest    = "sha256:c8a919765f2ef63681329fa21021b830cd4d79d1165bdca730dd016014e4da84"
	PostgresEngineVersion  = "18.4"
	MaximumOwnershipChunks = 512
	maximumSafeInteger     = int64(9007199254740991)
)

type DatabaseOrigin string

const (
	DatabaseOriginFresh     DatabaseOrigin = "fresh"
	DatabaseOriginV1Adopted DatabaseOrigin = "v1_adopted"
)

var (
	digestPattern     = regexp.MustCompile(`^[0-9a-f]{64}$`)
	databasePattern   = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)
	modulePattern     = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,63}$`)
	uuidPattern       = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	versionPattern    = regexp.MustCompile(`^[0-9]+\.[0-9]+(?:\.[0-9]+)?$`)
	artifactPathChars = regexp.MustCompile(`^[a-z0-9][a-z0-9_./-]{0,511}$`)
)

func (origin DatabaseOrigin) Valid() bool {
	return origin == DatabaseOriginFresh || origin == DatabaseOriginV1Adopted
}

func validateDigest(value, field string) error {
	if !digestPattern.MatchString(value) {
		return fmt.Errorf("%s is invalid", field)
	}
	return nil
}

func validateDatabaseName(value, field string) error {
	if !databasePattern.MatchString(value) {
		return fmt.Errorf("%s is invalid", field)
	}
	return nil
}

func validateIdentifier(value, field string) error {
	return validateDatabaseName(value, field)
}

func validateModule(value, field string) error {
	if !modulePattern.MatchString(value) {
		return fmt.Errorf("%s is invalid", field)
	}
	return nil
}

func validateArtifactPath(value, field string) error {
	if !artifactPathChars.MatchString(value) || strings.Contains(value, "//") {
		return fmt.Errorf("%s is invalid", field)
	}
	for _, part := range strings.Split(value, "/") {
		if part == "." || part == ".." {
			return fmt.Errorf("%s is invalid", field)
		}
	}
	return nil
}

func validateOrderedUnique(keys []string, field string) error {
	for index, key := range keys {
		if key == "" {
			return fmt.Errorf("%s contains an empty key", field)
		}
		if index > 0 && keys[index-1] >= key {
			return fmt.Errorf("%s must be strictly ordered and unique", field)
		}
	}
	return nil
}
