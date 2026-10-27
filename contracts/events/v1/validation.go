package eventsv1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	"github.com/google/uuid"
)

var (
	lowerSHA256 = regexp.MustCompile(`^[0-9a-f]{64}$`)
	domainName  = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)
	typeName    = regexp.MustCompile(`^[a-z][a-z0-9_.-]{2,127}$`)
	correlation = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,63}$`)
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

func validUUID(value string) bool {
	parsed, err := uuid.Parse(value)
	return err == nil && parsed != uuid.Nil && parsed.String() == value
}

func validText(value string, maximumBytes int) bool {
	return utf8.ValidString(value) && len(value) <= maximumBytes &&
		strings.TrimSpace(value) == value && value != ""
}

func validTimestamp(value string) (time.Time, bool) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	return parsed, err == nil && parsed.Location() == time.UTC && parsed.Format(time.RFC3339Nano) == value
}

func validCorrelationID(value string) bool {
	return correlation.MatchString(value)
}

func validDomain(value string) bool { return domainName.MatchString(value) }

func validType(value string) bool { return typeName.MatchString(value) }

func cloneOptional(value *string) *string {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}
