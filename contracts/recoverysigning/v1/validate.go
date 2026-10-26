package recoverysigningv1

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

func requireJSONEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("trailing JSON value")
	}
	return nil
}

func validPurpose(value Purpose) bool {
	return value == PurposeRecoveryManifest || value == PurposeRecoveryControl ||
		value == PurposeRestoreReceipt
}

func microsecondUTC(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond()%1_000 == 0
}

func parseTime(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil || parsed.Format(time.RFC3339Nano) != value || !microsecondUTC(parsed) {
		return time.Time{}, fmt.Errorf("timestamp is invalid")
	}
	return parsed, nil
}

func validUUID(value string) bool {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' ||
		value[18] != '-' || value[23] != '-' ||
		value == "00000000-0000-0000-0000-000000000000" {
		return false
	}
	for index, character := range value {
		if index == 8 || index == 13 || index == 18 || index == 23 {
			continue
		}
		if !strings.ContainsRune("0123456789abcdef", character) {
			return false
		}
	}
	return true
}
