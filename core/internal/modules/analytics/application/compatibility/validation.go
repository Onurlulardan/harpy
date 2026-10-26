package compatibility

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	reasonPattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]{0,63}$`)
	slugPattern   = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._~-]{0,99}$`)
	schemaPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)
)

func validUUID(value uuid.UUID) bool {
	return value != uuid.Nil && value.Variant() == uuid.RFC4122 &&
		(value.Version() == 4 || value.Version() == 7)
}

func validAuthority(value AdoptionAuthority) bool {
	return validV7(value.AuthorityID) && value.TenantID != uuid.Nil &&
		schemaPattern.MatchString(value.SchemaName)
}

func validV7(value uuid.UUID) bool {
	return value != uuid.Nil && value.Variant() == uuid.RFC4122 && value.Version() == 7
}

func validTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond()%1000 == 0
}

func validKey(value LegacyKey) bool {
	return len(value.TenantIdentity) >= 1 && len(value.TenantIdentity) <= 200 &&
		(value.Method == "GET" || value.Method == "POST") &&
		strings.HasPrefix(value.Path, "/") && len(value.Path) <= 512 &&
		slugPattern.MatchString(value.Slug)
}

func validMapping(value LegacyMapping) bool {
	return validKey(value.Key) && value.TenantID != uuid.Nil &&
		validUUID(value.LegacyEndpointID) && validV7(value.ReleaseID) &&
		value.PublicID != [16]byte{} && (value.State == "active" || value.State == "retired") &&
		validTime(value.RemoveAfter)
}
