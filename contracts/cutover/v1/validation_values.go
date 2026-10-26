package cutoverv1

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	digestPattern         = regexp.MustCompile(`^[0-9a-f]{64}$`)
	keyIDPattern          = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,63}$`)
	sqlNamePattern        = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)
	transformIDPattern    = regexp.MustCompile(`^[a-z][a-z0-9_.-]{0,127}$`)
	applyProcedurePattern = regexp.MustCompile(
		`^[a-z][a-z0-9_]{0,62}\.[a-z][a-z0-9_]{0,62}\(uuid,uuid,text,uuid,smallint,text,pg_lsn,pg_lsn,pg_lsn,xid8,integer,jsonb,text\)$`,
	)
)

func validUUID(value string) bool {
	parsed, err := uuid.Parse(value)
	return err == nil && parsed != uuid.Nil && parsed.String() == value &&
		parsed.Variant() == uuid.RFC4122 && parsed.Version() >= 1 && parsed.Version() <= 8
}

func validDigest(value string) bool { return digestPattern.MatchString(value) }

func parseUTC(value string) (time.Time, error) {
	if !strings.HasSuffix(value, "Z") && !strings.HasSuffix(value, "+00:00") {
		return time.Time{}, fmt.Errorf("cutover timestamp is not UTC")
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("cutover timestamp is invalid")
	}
	_, offset := parsed.Zone()
	if offset != 0 {
		return time.Time{}, fmt.Errorf("cutover timestamp is not UTC")
	}
	return parsed.UTC(), nil
}

func validSystemIdentifier(value string) bool {
	if value == "" || value[0] == '0' || len(value) > 20 {
		return false
	}
	parsed, err := strconv.ParseUint(value, 10, 64)
	return err == nil && parsed > 0
}

func validateSignature(signature Signature) error {
	if signature.Algorithm != "Ed25519" || !keyIDPattern.MatchString(signature.KeyID) {
		return fmt.Errorf("cutover signature identity is invalid")
	}
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(signature.Value)
	if err != nil || len(decoded) != 64 ||
		base64.RawURLEncoding.EncodeToString(decoded) != signature.Value {
		return fmt.Errorf("cutover signature value is invalid")
	}
	return nil
}

func allUnique(values ...string) bool {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			return false
		}
		seen[value] = struct{}{}
	}
	return true
}
