package resolved

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/internal/jsonscan"
	"github.com/Onurlulardan/harpy/internal/attestation"
	"github.com/google/uuid"
	"golang.org/x/text/unicode/norm"
)

func (value Value) Validate() error {
	if !value.validated || !value.kind.Valid() {
		return fmt.Errorf("resolved REST Query value is invalid")
	}
	valid := false
	switch value.state {
	case valueStateScalar:
		valid = value.scalarValid() && !value.boolean && value.json == nil
	case valueStateBoolean:
		valid = value.kind == api.LogicalTypeBoolean && value.scalar == "" && value.json == nil
	case valueStateJSON:
		valid = value.kind == api.LogicalTypeJSON && value.scalar == "" &&
			!value.boolean && value.jsonValid()
	}
	if !valid {
		return fmt.Errorf("resolved REST Query value is invalid")
	}
	return nil
}

func (value Value) scalarValid() bool {
	if !utf8.ValidString(value.scalar) ||
		strings.ContainsRune(value.scalar, '\x00') || !norm.NFC.IsNormalString(value.scalar) {
		return false
	}
	switch value.kind {
	case api.LogicalTypeText:
		return true
	case api.LogicalTypeEnum:
		return value.scalar != ""
	case api.LogicalTypeInteger:
		return validInteger(value.scalar)
	case api.LogicalTypeDecimal:
		return validDecimal(value.scalar)
	case api.LogicalTypeDate:
		return validDate(value.scalar)
	case api.LogicalTypeTimestamp:
		return validTimestamp(value.scalar)
	case api.LogicalTypeUUID, api.LogicalTypeRelation:
		parsed, err := uuid.Parse(value.scalar)
		return err == nil && parsed != uuid.Nil && parsed.String() == value.scalar &&
			parsed.Variant() == uuid.RFC4122
	default:
		return false
	}
}

func (value Value) jsonValid() bool {
	if len(value.json) == 0 || !utf8.Valid(value.json) || !json.Valid(value.json) ||
		bytes.Equal(value.json, []byte("null")) || !validJSONStrings(value.json) {
		return false
	}
	scan, scanErr := jsonscan.Scan(value.json, jsonscan.MaximumNesting)
	if scanErr != nil || len(scan.DuplicatePointers()) != 0 {
		return false
	}
	canonical, err := attestation.Canonicalize(value.json)
	return err == nil && bytes.Equal(canonical, value.json)
}

func validInteger(value string) bool {
	if value == "" || value == "-0" || value[0] == '+' || len(value) > 1 && value[0] == '0' ||
		len(value) > 1 && value[0] == '-' && value[1] == '0' {
		return false
	}
	_, err := strconv.ParseInt(value, 10, 64)
	return err == nil
}

func validDecimal(value string) bool {
	if value == "" || value[0] == '+' || strings.ContainsAny(value, "eE") {
		return false
	}
	unsigned := strings.TrimPrefix(value, "-")
	parts := strings.Split(unsigned, ".")
	if len(parts) > 2 || parts[0] == "" || len(parts) == 2 && parts[1] == "" ||
		len(parts[0]) > 1 && parts[0][0] == '0' {
		return false
	}
	for _, part := range parts {
		for _, character := range part {
			if character < '0' || character > '9' {
				return false
			}
		}
	}
	return value[0] != '-' || strings.ContainsAny(unsigned, "123456789")
}

func validDate(value string) bool {
	parsed, err := time.Parse(time.DateOnly, value)
	return err == nil && parsed.Format(time.DateOnly) == value
}

func validTimestamp(value string) bool {
	const layout = "2006-01-02T15:04:05.000000Z"
	parsed, err := time.Parse(layout, value)
	return err == nil && parsed.UTC().Format(layout) == value
}
