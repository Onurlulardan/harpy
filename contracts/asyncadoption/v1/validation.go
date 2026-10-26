package asyncadoptionv1

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	"github.com/google/uuid"
)

var (
	lowerSHA256 = regexp.MustCompile(`^[0-9a-f]{64}$`)
	domainName  = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)
	boundedName = regexp.MustCompile(`^[a-z][a-z0-9_.-]{2,127}$`)
	machineCode = regexp.MustCompile(`^[A-Z][A-Z0-9_]{1,63}$`)
)

func decodeStrict(document []byte, target any) error {
	if len(bytes.TrimSpace(document)) == 0 || len(document) > MaximumDocumentBytes || !utf8.Valid(document) {
		return fmt.Errorf("document size or UTF-8 is invalid")
	}
	if err := canonicaljsonv1.Validate(document); err != nil {
		return fmt.Errorf("document JSON is ambiguous: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("decode document: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("document has trailing content")
	}
	return nil
}

func canonicalDigest(value any) (string, error) {
	document, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, MaximumDocumentBytes)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(canonical)
	return hex.EncodeToString(digest[:]), nil
}

func digestLines(lines []string) string {
	copy := append([]string(nil), lines...)
	sort.Strings(copy)
	digest := sha256.Sum256([]byte(strings.Join(copy, "\n")))
	return hex.EncodeToString(digest[:])
}

func validUUID(value string) bool {
	parsed, err := uuid.Parse(value)
	return err == nil && parsed != uuid.Nil && parsed.String() == value
}

func validTimestamp(value string) bool {
	_, valid := parseTimestamp(value)
	return valid
}

func parseTimestamp(value string) (time.Time, bool) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	valid := err == nil && parsed.Location() == time.UTC &&
		parsed.Format(time.RFC3339Nano) == value
	return parsed, valid
}

func validText(value string, maximum int) bool {
	return value != "" && utf8.ValidString(value) && strings.TrimSpace(value) == value && len([]byte(value)) <= maximum
}

func cloneString(value *string) *string {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}
